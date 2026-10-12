package parser

import (
	"path/filepath"
	"testing"
)

func TestParseCursorGolden(t *testing.T) {
	path := filepath.Join("..", "..", "testdata", "cursor_sample.jsonl")
	sess, err := Parse("cursor", path)
	if err != nil {
		t.Fatal(err)
	}
	if sess.SourceCLI != "cursor" {
		t.Fatalf("cli %q", sess.SourceCLI)
	}
	if len(sess.Turns) != 2 {
		t.Fatalf("turns %d", len(sess.Turns))
	}
	if len(sess.Turns[1].ToolSummaries) != 1 {
		t.Fatalf("summaries %v", sess.Turns[1].ToolSummaries)
	}
}
