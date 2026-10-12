package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mlaccetti/harness-bridge/internal/parser"
	"github.com/mlaccetti/harness-bridge/internal/testfixture"
)

func TestTransferDryRun_claudeSinceCompaction(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	modRoot := moduleRoot(t)
	sample, err := os.ReadFile(filepath.Join(modRoot, "testdata", "claude_compact.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	if err := testfixture.SeedClaudeSession(home, "/tmp/p", "compact-sess.jsonl", sample); err != nil {
		t.Fatal(err)
	}

	root := NewRoot()
	var out bytes.Buffer
	WireOutput(root, &out)
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{
		"transfer", "--from", "claude", "--to", "markdown",
		"--session", "compact-sess", "--cwd", t.TempDir(),
		"--dry-run", "--since-compaction",
	})

	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	body := out.String()
	if !strings.Contains(body, "dry-run") || !strings.Contains(body, "since-compaction") {
		t.Fatalf("output:\n%s", body)
	}
	if !strings.Contains(body, "2 turns") && !strings.Contains(body, "→ 2 turns") {
		t.Fatalf("expected 2 turns in plan:\n%s", body)
	}
}

func TestTransferDryRun_allSources(t *testing.T) {
	modRoot := moduleRoot(t)
	cases := []struct {
		from string
		file string
	}{
		{"claude", "claude_sample.jsonl"},
		{"codex", "codex_sample.jsonl"},
		{"cursor", "cursor_sample.jsonl"},
		{"antigravity", "antigravity_sample.jsonl"},
	}
	for _, tc := range cases {
		path := filepath.Join(modRoot, "testdata", tc.file)
		sess, err := parser.Parse(tc.from, path)
		if err != nil {
			t.Fatalf("%s parse: %v", tc.from, err)
		}
		if len(sess.Turns) == 0 {
			t.Fatalf("%s: no turns", tc.from)
		}
		sliced, _ := sess.ApplySlice(sliceFlags{lastTurns: 1}.opts())
		if len(sliced.Turns) != 1 {
			t.Fatalf("%s slice: %d turns", tc.from, len(sliced.Turns))
		}
		var buf bytes.Buffer
		writeTransferPlan(&buf, "markdown", t.TempDir(), sess, sliceFlags{lastTurns: 1}, true)
		if !strings.Contains(buf.String(), "HANDOFF.md") {
			t.Fatalf("%s plan missing target: %s", tc.from, buf.String())
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
