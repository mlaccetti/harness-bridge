package parser

import (
	"path/filepath"
	"testing"

	"github.com/mlaccetti/harness-bridge/internal/session"
)

func TestParseClaudeCompactBoundary(t *testing.T) {
	path := filepath.Join("..", "..", "testdata", "claude_compact.jsonl")
	sess, err := Parse("claude", path)
	if err != nil {
		t.Fatal(err)
	}
	if len(sess.Turns) != 4 {
		t.Fatalf("turns %d", len(sess.Turns))
	}
	if len(sess.Compactions) != 1 {
		t.Fatalf("compactions %d", len(sess.Compactions))
	}
	if sess.Compactions[0].TurnIndex != 2 {
		t.Fatalf("index %d", sess.Compactions[0].TurnIndex)
	}
	sliced, _ := sess.ApplySlice(session.SliceOptions{SinceCompaction: true})
	if len(sliced.Turns) != 2 {
		t.Fatalf("sliced turns %d", len(sliced.Turns))
	}
	if sliced.Turns[0].Text != "new topic after compact" {
		t.Fatalf("text %q", sliced.Turns[0].Text)
	}
}
