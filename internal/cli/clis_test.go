package cli

import (
	"strings"
	"testing"
)

func TestResumeHint_codex(t *testing.T) {
	hint := ResumeHint("codex", "uuid-1")
	if hint == "" || !strings.Contains(hint, "uuid-1") {
		t.Fatalf("got %q", hint)
	}
}

func TestResumeHint_markdown(t *testing.T) {
	hint := ResumeHint("markdown", "")
	if hint == "" || !strings.Contains(hint, "HANDOFF.md") {
		t.Fatalf("got %q", hint)
	}
}
