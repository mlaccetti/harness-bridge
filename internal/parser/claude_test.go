package parser

import (
	"path/filepath"
	"testing"
)

func TestParseClaudeSkipsTaskNotification(t *testing.T) {
	path := filepath.Join("..", "..", "testdata", "claude_noise.jsonl")
	sess, err := Parse("claude", path)
	if err != nil {
		t.Fatal(err)
	}
	if len(sess.Turns) != 2 {
		t.Fatalf("turns %d", len(sess.Turns))
	}
}

func TestParseClaudeGolden(t *testing.T) {
	path := filepath.Join("..", "..", "testdata", "claude_sample.jsonl")
	sess, err := Parse("claude", path)
	if err != nil {
		t.Fatal(err)
	}
	if sess.SourceSessionID != "sess-abc" {
		t.Fatalf("session id: %q", sess.SourceSessionID)
	}
	if sess.CWD != "/tmp/project" {
		t.Fatalf("cwd: %q", sess.CWD)
	}
	if len(sess.Turns) != 2 {
		t.Fatalf("turns: %d", len(sess.Turns))
	}
	if len(sess.Turns[1].ToolSummaries) != 1 {
		t.Fatalf("tool summaries: %v", sess.Turns[1].ToolSummaries)
	}
}
