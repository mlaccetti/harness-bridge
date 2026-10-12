package main_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mlaccetti/harness-bridge/internal/testfixture"
)

func TestE2E_dryRun_codexCursorAntigravity(t *testing.T) {
	modRoot := goModDir(t)
	home := t.TempDir()
	work := t.TempDir()

	codexSample, _ := os.ReadFile(filepath.Join(modRoot, "testdata", "codex_sample.jsonl"))
	_ = testfixture.SeedCodexRollout(home, "rollout-e2e-dry.jsonl", codexSample)

	_, stderr, code := runCLI(t, home,
		"transfer", "--from", "codex", "--to", "markdown",
		"--session", "e2e-dry", "--cwd", work,
		"--dry-run", "--last-turns", "1",
	)
	if code != 0 {
		t.Fatalf("codex dry-run exit %d: %s", code, stderr)
	}

	agySample, _ := os.ReadFile(filepath.Join(modRoot, "testdata", "antigravity_sample.jsonl"))
	encoded := "-tmp-agy"
	agyDir := filepath.Join(home, ".config", "antigravity", "projects", encoded)
	_ = os.MkdirAll(agyDir, 0o755)
	_ = os.WriteFile(filepath.Join(agyDir, "agy-1.jsonl"), agySample, 0o644)

	_, stderr, code = runCLI(t, home,
		"transfer", "--from", "antigravity", "--to", "markdown",
		"--session", "agy-1", "--cwd", work,
		"--plan",
	)
	if code != 0 {
		t.Fatalf("antigravity plan exit %d: %s", code, stderr)
	}

	cursorSample, _ := os.ReadFile(filepath.Join(modRoot, "testdata", "cursor_sample.jsonl"))
	// cursor layout: ~/.cursor/projects/<encoded>/agent-transcripts/<id>/<id>.jsonl
	proj := filepath.Join(home, ".cursor", "projects", "Users-me-work")
	id := "e2e-cursor-id"
	transcript := filepath.Join(proj, "agent-transcripts", id)
	_ = os.MkdirAll(transcript, 0o755)
	_ = os.WriteFile(filepath.Join(transcript, id+".jsonl"), cursorSample, 0o644)

	out, stderr, code := runCLI(t, home,
		"transfer", "--from", "cursor", "--to", "markdown",
		"--session", id, "--cwd", work,
		"--dry-run",
	)
	if code != 0 {
		t.Fatalf("cursor dry-run exit %d: %s", code, stderr)
	}
	if !strings.Contains(out, "dry-run") {
		t.Fatalf("cursor output: %s", out)
	}
}
