package parser

import (
	"path/filepath"
	"testing"
)

func TestParseGeminiGolden(t *testing.T) {
	path := filepath.Join("..", "..", "testdata", "gemini_sample.json")
	sess, err := Parse("gemini", path)
	if err != nil {
		t.Fatal(err)
	}
	if sess.SourceSessionID != "gem-sess-1" {
		t.Fatalf("id %q", sess.SourceSessionID)
	}
	if len(sess.Turns) != 2 {
		t.Fatalf("turns %d", len(sess.Turns))
	}
}
