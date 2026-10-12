package testfixture

import (
	"os"
	"path/filepath"

	"github.com/mlaccetti/harness-bridge/internal/paths"
)

// SeedClaudeSession writes a Claude Code JSONL file under a fake HOME tree.
func SeedClaudeSession(home, cwd, fileName string, content []byte) error {
	encoded := paths.ClaudeEncodeCWD(cwd)
	dir := filepath.Join(home, ".claude", "projects", encoded)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, fileName), content, 0o644)
}

// SeedCodexRollout writes a Codex rollout JSONL under fake HOME.
func SeedCodexRollout(home, fileName string, content []byte) error {
	dir := filepath.Join(home, ".codex", "sessions", "2026", "01", "15")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, fileName), content, 0o644)
}
