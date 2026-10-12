package main_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/mlaccetti/harness-bridge/internal/testfixture"
)

var (
	binMu     sync.Mutex
	cachedBin string
)

func cliBin(t *testing.T) string {
	t.Helper()
	binMu.Lock()
	defer binMu.Unlock()
	if cachedBin != "" {
		return cachedBin
	}
	modRoot := goModDir(t)
	out := filepath.Join(os.TempDir(), "harness-bridge-e2e-test-bin")
	cmd := exec.Command("go", "build", "-o", out, "./cmd/harness-bridge")
	cmd.Dir = modRoot
	cmd.Env = goBuildEnv()
	if err := cmd.Run(); err != nil {
		t.Fatalf("build harness-bridge: %v", err)
	}
	cachedBin = out
	return cachedBin
}

func withHome(environ []string, home string) []string {
	var out []string
	for _, e := range environ {
		if strings.HasPrefix(e, "HOME=") {
			continue
		}
		out = append(out, e)
	}
	return append(out, "HOME="+home)
}

func goBuildEnv() []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return os.Environ()
	}
	gomodcache := os.Getenv("GOMODCACHE")
	if gomodcache == "" {
		gomodcache = filepath.Join(home, "go", "pkg", "mod")
	}
	gocache := os.Getenv("GOCACHE")
	if gocache == "" {
		gocache = filepath.Join(home, "Library", "Caches", "go-build")
	}
	var env []string
	for _, e := range os.Environ() {
		if strings.HasPrefix(e, "HOME=") ||
			strings.HasPrefix(e, "GOMODCACHE=") ||
			strings.HasPrefix(e, "GOCACHE=") {
			continue
		}
		env = append(env, e)
	}
	env = append(env, "HOME="+home, "GOMODCACHE="+gomodcache, "GOCACHE="+gocache)
	return env
}

func goModDir(t *testing.T) string {
	t.Helper()
	cmd := exec.Command("go", "env", "GOMOD")
	cmd.Env = goBuildEnv()
	out, err := cmd.Output()
	if err != nil {
		t.Fatal(err)
	}
	mod := strings.TrimSpace(string(out))
	if mod == "" {
		t.Fatal("GOMOD empty")
	}
	return filepath.Dir(mod)
}

func runCLI(t *testing.T, home string, args ...string) (stdout, stderr string, exitCode int) {
	t.Helper()
	bin := cliBin(t)
	cmd := exec.Command(bin, args...)
	cmd.Env = withHome(os.Environ(), home)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err := cmd.Run()
	code := 0
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			code = ee.ExitCode()
		} else {
			t.Fatalf("run %v: %v", args, err)
		}
	}
	return outBuf.String(), errBuf.String(), code
}

func readTestdata(t *testing.T, name string) []byte {
	t.Helper()
	path := filepath.Join(goModDir(t), "testdata", name)
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestE2E_version(t *testing.T) {
	home := t.TempDir()
	out, _, code := runCLI(t, home, "--version")
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	if strings.TrimSpace(out) == "" {
		t.Fatalf("empty version output")
	}
}

func TestE2E_listShowTransfer_markdown(t *testing.T) {
	home := t.TempDir()
	work := filepath.Join(home, "work")
	if err := os.MkdirAll(work, 0o755); err != nil {
		t.Fatal(err)
	}

	sample := readTestdata(t, "claude_sample.jsonl")
	if err := testfixture.SeedClaudeSession(home, "/tmp/project", "sess-abc.jsonl", sample); err != nil {
		t.Fatal(err)
	}

	listOut, listErr, code := runCLI(t, home, "list", "--from", "claude", "--limit", "5")
	if code != 0 {
		t.Fatalf("list exit %d stderr=%q", code, listErr)
	}
	if !strings.Contains(listOut, "sess-abc") {
		t.Fatalf("list missing session:\n%s", listOut)
	}

	showOut, showErr, code := runCLI(t, home, "show", "--from", "claude", "--session", "sess-abc")
	if code != 0 {
		t.Fatalf("show exit %d stderr=%q", code, showErr)
	}
	if !strings.Contains(showOut, "Hello from Claude") {
		t.Fatalf("show output:\n%s", showOut)
	}

	_, transferErr, code := runCLI(t, home, "transfer", "--from", "claude", "--to", "markdown", "--session", "sess-abc", "--cwd", work)
	if code != 0 {
		t.Fatalf("transfer exit %d stderr=%q", code, transferErr)
	}
	handoff := filepath.Join(work, "HANDOFF.md")
	raw, err := os.ReadFile(handoff)
	if err != nil {
		t.Fatal(err)
	}
	body := string(raw)
	if !strings.Contains(body, "Hello from Claude") || !strings.Contains(body, "Hi there") {
		t.Fatalf("HANDOFF.md:\n%s", body)
	}
}

func TestE2E_transfer_claudeToClaude_noExternalHarness(t *testing.T) {
	home := t.TempDir()
	targetCWD := filepath.Join(home, "target-project")
	if err := os.MkdirAll(targetCWD, 0o755); err != nil {
		t.Fatal(err)
	}

	sample := readTestdata(t, "claude_sample.jsonl")
	if err := testfixture.SeedClaudeSession(home, "/tmp/project", "bridge-src.jsonl", sample); err != nil {
		t.Fatal(err)
	}

	_, stderr, code := runCLI(t, home,
		"transfer", "--from", "claude", "--to", "claude",
		"--session", "bridge-src", "--cwd", targetCWD,
	)
	if code != 0 {
		t.Fatalf("transfer exit %d stderr=%q", code, stderr)
	}

	entries, err := os.ReadDir(filepath.Join(home, ".claude", "projects"))
	if err != nil {
		t.Fatal(err)
	}
	var wrote bool
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		files, _ := filepath.Glob(filepath.Join(home, ".claude", "projects", e.Name(), "*.jsonl"))
		for _, f := range files {
			if strings.Contains(f, "bridge-src") {
				continue
			}
			info, _ := os.Stat(f)
			if info != nil && info.Size() > 0 {
				wrote = true
			}
		}
	}
	if !wrote {
		t.Fatal("expected new claude session jsonl under target project dir")
	}
}

func TestE2E_transfer_unknownTarget_fails(t *testing.T) {
	home := t.TempDir()
	_, stderr, code := runCLI(t, home, "transfer", "--from", "claude", "--to", "copilot")
	if code == 0 {
		t.Fatalf("expected failure, stderr=%q", stderr)
	}
}

func TestE2E_codexListFromFixture(t *testing.T) {
	home := t.TempDir()
	sample := readTestdata(t, "codex_sample.jsonl")
	if err := testfixture.SeedCodexRollout(home, "rollout-e2e-codex-1.jsonl", sample); err != nil {
		t.Fatal(err)
	}
	out, _, code := runCLI(t, home, "list", "--from", "codex", "--limit", "3")
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	if !strings.Contains(out, "rollout-e2e-codex-1") {
		t.Fatalf("list:\n%s", out)
	}
}
