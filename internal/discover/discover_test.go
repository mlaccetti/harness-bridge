package discover

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mlaccetti/harness-bridge/internal/paths"
)

func TestNormalizeCLI_agyAlias(t *testing.T) {
	got, err := NormalizeCLI("agy")
	if err != nil {
		t.Fatal(err)
	}
	if got != "antigravity" {
		t.Fatalf("got %q", got)
	}
}

func TestNormalizeCLI_unknown(t *testing.T) {
	_, err := NormalizeCLI("copilot")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestFindSession_byIDAcrossCWD(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cwd := "/tmp/bridge-test"
	projectDir := filepath.Join(home, ".claude", "projects", paths.ClaudeEncodeCWD(cwd))
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(projectDir, "abc123-session.jsonl")
	if err := os.WriteFile(target, []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}

	found, err := FindSession("claude", "abc123", "/other/cwd")
	if err != nil {
		t.Fatal(err)
	}
	if found != target {
		t.Fatalf("got %q want %q", found, target)
	}
}

func TestFindSession_mostRecentInCWD(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cwd := "/tmp/bridge-cwd"
	projectDir := filepath.Join(home, ".claude", "projects", paths.ClaudeEncodeCWD(cwd))
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatal(err)
	}
	older := filepath.Join(projectDir, "older.jsonl")
	newer := filepath.Join(projectDir, "newer.jsonl")
	_ = os.WriteFile(older, []byte("{}"), 0o644)
	_ = os.WriteFile(newer, []byte("{}"), 0o644)
	oldTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	newTime := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	_ = os.Chtimes(older, oldTime, oldTime)
	_ = os.Chtimes(newer, newTime, newTime)

	found, err := FindSession("claude", "", cwd)
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(found) != "newer.jsonl" {
		t.Fatalf("got %q", found)
	}
}
