package cli

import (
	"github.com/spf13/cobra"
)

func NewRoot() *cobra.Command {
	root := &cobra.Command{
		Use:     "harness-bridge",
		Aliases: []string{"hb", "session-bridge", "sb"},
		Short:   "Bridge coding-agent sessions between CLIs",
		Long: `harness-bridge transfers conversation context between coding-agent CLIs
(Claude Code, Codex, Cursor, Gemini, Antigravity) so you can resume elsewhere
without re-explaining the task.`,
		Example: `  harness-bridge list --from claude
  harness-bridge show --from codex --session abc123
  harness-bridge transfer --from claude --to cursor --cwd .`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.AddCommand(newListCmd(), newShowCmd(), newTransferCmd())
	setupCompletions(root)
	return root
}
