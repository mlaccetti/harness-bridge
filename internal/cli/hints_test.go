package cli

import (
	"testing"

	"github.com/mlaccetti/harness-bridge/internal/session"
)

func TestTaskHint_stripsCursorMetadata(t *testing.T) {
	sess := session.Session{
		Turns: []session.Turn{{
			Role: session.RoleUser,
			Text: "<timestamp>Mon</timestamp>\n<user_query>\nFix the flaky test\n</user_query>",
		}},
	}
	hint := TaskHint(sess)
	if hint != "Fix the flaky test" {
		t.Fatalf("got %q", hint)
	}
}

func TestTaskHint_skipsBlankUserTurns(t *testing.T) {
	sess := session.Session{
		Turns: []session.Turn{
			{Role: session.RoleUser, Text: "   "},
			{Role: session.RoleAssistant, Text: "ok"},
			{Role: session.RoleUser, Text: "follow up please"},
		},
	}
	hint := TaskHint(sess)
	if hint != "follow up please" {
		t.Fatalf("got %q", hint)
	}
}
