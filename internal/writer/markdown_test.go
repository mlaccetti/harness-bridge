package writer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mlaccetti/harness-bridge/internal/session"
)

func TestWriteMarkdown(t *testing.T) {
	dir := t.TempDir()
	sess := session.Session{
		SourceCLI:       "claude",
		SourceSessionID: "src-1",
		CWD:             "/tmp/old",
		Turns: []session.Turn{
			{Role: session.RoleUser, Text: "Continue here"},
			{Role: session.RoleAssistant, Text: "Will do"},
		},
	}
	outPath, _, err := Write("markdown", sess, dir)
	if err != nil {
		t.Fatal(err)
	}
	if outPath != filepath.Join(dir, "HANDOFF.md") {
		t.Fatalf("path %q", outPath)
	}
	raw, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	body := string(raw)
	if !strings.Contains(body, "Continue here") || !strings.Contains(body, "Session handoff") {
		t.Fatalf("unexpected content: %s", body)
	}
}
