package testfixture

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSeedClaudeSession(t *testing.T) {
	home := t.TempDir()
	if err := SeedClaudeSession(home, "/tmp/proj", "abc.jsonl", []byte("{}\n")); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(home, ".claude", "projects", "-tmp-proj", "abc.jsonl")
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
}
