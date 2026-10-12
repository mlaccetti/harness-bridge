package integration

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mlaccetti/harness-bridge/internal/cli"
	"github.com/mlaccetti/harness-bridge/internal/testfixture"
)

// In-process CLI tests (Cobra commands, no subprocess). Complements cmd/harness-bridge
// subprocess tests that exercise Fang + the compiled binary.

func runRoot(t *testing.T, home string, args ...string) (stdout string, err error) {
	t.Helper()
	t.Setenv("HOME", home)

	root := cli.NewRoot()
	var out bytes.Buffer
	cli.WireOutput(root, &out)
	root.SetArgs(args)
	err = root.Execute()
	return out.String(), err
}

func TestCLI_transferMarkdown_inProcess(t *testing.T) {
	home := t.TempDir()
	work := filepath.Join(home, "work")
	if err := os.MkdirAll(work, 0o755); err != nil {
		t.Fatal(err)
	}

	modRoot := moduleRoot(t)
	sample, err := os.ReadFile(filepath.Join(modRoot, "testdata", "claude_sample.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	if err := testfixture.SeedClaudeSession(home, "/tmp/project", "sess-abc.jsonl", sample); err != nil {
		t.Fatal(err)
	}

	out, err := runRoot(t, home,
		"transfer", "--from", "claude", "--to", "markdown",
		"--session", "sess-abc", "--cwd", work,
	)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "HANDOFF.md") {
		t.Fatalf("output: %s", out)
	}
	handoff := filepath.Join(work, "HANDOFF.md")
	if _, err := os.Stat(handoff); err != nil {
		t.Fatal(err)
	}
}

func TestCLI_listEnhancedFields_inProcess(t *testing.T) {
	home := t.TempDir()
	modRoot := moduleRoot(t)
	sample, err := os.ReadFile(filepath.Join(modRoot, "testdata", "claude_sample.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	if err := testfixture.SeedClaudeSession(home, "/tmp/project", "sess-abc.jsonl", sample); err != nil {
		t.Fatal(err)
	}

	out, err := runRoot(t, home, "list", "--from", "claude", "--limit", "1")
	if err != nil {
		t.Fatal(err)
	}
	for _, needle := range []string{"turns:", "hint:", "Hello from Claude"} {
		if !strings.Contains(out, needle) {
			t.Fatalf("missing %q in:\n%s", needle, out)
		}
	}
}

func moduleRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	dir := wd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found")
		}
		dir = parent
	}
}
