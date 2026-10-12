package paths

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestClaudeEncodeCWD(t *testing.T) {
	got := ClaudeEncodeCWD("C:\\Users\\foo bar\\proj")
	want := "C--Users-foo-bar-proj"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestCursorEncodeWorkspace(t *testing.T) {
	got := CursorEncodeWorkspace("/Users/me/workspace/harness-bridge")
	want := "Users-me-workspace-harness-bridge"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestGeminiProjectFolderNameFromIndex(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	geminiDir := filepath.Join(home, ".gemini")
	if err := os.MkdirAll(geminiDir, 0o755); err != nil {
		t.Fatal(err)
	}
	idx := filepath.Join(geminiDir, "projects.json")
	if err := os.WriteFile(idx, []byte(`{"projects":{"/tmp/my-repo":"friendly-name"}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := GeminiProjectFolderName("/tmp/my-repo"); got != "friendly-name" {
		t.Fatalf("got %q", got)
	}
}

func TestClaudeSessionsForCWD_sortedNewest(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cwd := "/tmp/sample-project"
	encoded := ClaudeEncodeCWD(cwd)
	projectDir := filepath.Join(home, ".claude", "projects", encoded)
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatal(err)
	}
	old := filepath.Join(projectDir, "old.jsonl")
	newer := filepath.Join(projectDir, "new.jsonl")
	if err := os.WriteFile(old, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(newer, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	oldTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	newTime := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	_ = os.Chtimes(old, oldTime, oldTime)
	_ = os.Chtimes(newer, newTime, newTime)

	sessions := ClaudeSessionsForCWD(cwd)
	if len(sessions) != 2 {
		t.Fatalf("sessions: %d", len(sessions))
	}
	if filepath.Base(sessions[0]) != "new.jsonl" {
		t.Fatalf("expected newest first, got %v", sessions)
	}
}
