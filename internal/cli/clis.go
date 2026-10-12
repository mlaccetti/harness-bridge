package cli

var SourceCLIs = []string{"claude", "codex", "gemini", "cursor", "antigravity"}

var TargetCLIs = []string{"claude", "codex", "gemini", "cursor", "antigravity", "markdown"}

func ResumeHint(target, sessionID string) string {
	switch target {
	case "claude":
		return "Start Claude in this folder; use /resume to pick the session."
	case "codex":
		return "cd into the cwd and run `codex resume " + sessionID + "` (or `codex resume --all`)."
	case "gemini":
		return "cd into the cwd and run `gemini --resume " + sessionID + "`."
	case "cursor":
		return "Run `cursor-agent --resume " + sessionID + "`."
	case "antigravity":
		return "Run `agy --conversation " + sessionID + "` or `agy --continue`."
	case "markdown":
		return "Open the new CLI in this cwd and say: read HANDOFF.md and continue from where the user last spoke."
	default:
		return ""
	}
}
