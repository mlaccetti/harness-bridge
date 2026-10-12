package writer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mlaccetti/harness-bridge/internal/parser"
	"github.com/mlaccetti/harness-bridge/internal/session"
)

func TestClaudeRoundTrip(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cwd := filepath.Join(home, "project")
	if err := os.MkdirAll(cwd, 0o755); err != nil {
		t.Fatal(err)
	}

	src := session.Session{
		SourceCLI:       "codex",
		SourceSessionID: "old",
		Turns: []session.Turn{
			{Role: session.RoleUser, Text: "hello"},
			{Role: session.RoleAssistant, Text: "world"},
		},
	}
	outPath, newID, err := Write("claude", src, cwd)
	if err != nil {
		t.Fatal(err)
	}
	if newID == "" {
		t.Fatal("empty session id")
	}
	parsed, err := parser.Parse("claude", outPath)
	if err != nil {
		t.Fatal(err)
	}
	// banner user turn + 2 bridged turns
	if len(parsed.Turns) < 2 {
		t.Fatalf("turns %d", len(parsed.Turns))
	}
}
