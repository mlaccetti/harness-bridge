package writer

import (
	"path/filepath"

	"github.com/mlaccetti/harness-bridge/internal/paths"
)

// DescribeTarget returns where a bridged session would be written (placeholder ids).
func DescribeTarget(cli, cwd string) string {
	switch cli {
	case "markdown":
		return filepath.Join(cwd, "HANDOFF.md")
	case "claude":
		return filepath.Join(paths.ClaudeProjectDir(cwd), "<new-session-uuid>.jsonl")
	case "codex":
		return filepath.Join(paths.CodexSessionsRoot(), "YYYY", "MM", "DD", "rollout-<ts>-<new-uuid>.jsonl")
	case "gemini":
		name := paths.GeminiProjectFolderName(cwd)
		return filepath.Join(paths.GeminiRoot(), "tmp", name, "chats", "session-<ts>-<id8>.json")
	case "cursor":
		dir := paths.CursorProjectDir(cwd)
		return filepath.Join(dir, "agent-transcripts", "<new-uuid>", "<new-uuid>.jsonl")
	case "antigravity":
		return filepath.Join(paths.AntigravityProjectDir(cwd), "<new-session-uuid>.jsonl")
	default:
		return "(unknown target)"
	}
}
