package render

import (
	"strings"
	"testing"

	"github.com/mlaccetti/harness-bridge/internal/session"
)

func TestRenderTurnText_appendsToolSummaries(t *testing.T) {
	out := RenderTurnText(session.Turn{
		Role:          session.RoleAssistant,
		Text:          "Done.",
		ToolSummaries: []string{"ran `go test`", "read main.go"},
	})
	if !strings.Contains(out, "Done.") {
		t.Fatalf("missing body: %q", out)
	}
	if !strings.Contains(out, "ran `go test`") {
		t.Fatalf("missing summary: %q", out)
	}
	if !strings.Contains(out, "actions taken in the previous session") {
		t.Fatalf("missing banner: %q", out)
	}
}

func TestRenderTurnText_plainText(t *testing.T) {
	out := RenderTurnText(session.Turn{Role: session.RoleUser, Text: "hello"})
	if out != "hello" {
		t.Fatalf("got %q", out)
	}
}
