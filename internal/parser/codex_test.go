package parser

import (
	"path/filepath"
	"testing"
)

func TestParseCodexGolden(t *testing.T) {
	path := filepath.Join("..", "..", "testdata", "codex_sample.jsonl")
	sess, err := Parse("codex", path)
	if err != nil {
		t.Fatal(err)
	}
	if sess.SourceSessionID != "codex-1" {
		t.Fatalf("id %q", sess.SourceSessionID)
	}
	if len(sess.Turns) != 2 {
		t.Fatalf("turns %d", len(sess.Turns))
	}
}
