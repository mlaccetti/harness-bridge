package cli

import (
	"os"
	"path/filepath"

	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"

	"github.com/mlaccetti/harness-bridge/internal/discover"
)

func setupCompletions(root *cobra.Command) {
	carapace.Gen(root).Standalone()

	carapace.Gen(root).PositionalAnyCompletion(
		carapace.ActionValues(),
	)

	completeFrom := carapace.ActionValues("claude", "codex", "gemini", "cursor", "antigravity", "agy").Tag("source CLI")
	completeTo := carapace.ActionValues("claude", "codex", "gemini", "cursor", "antigravity", "agy", "markdown").Tag("target CLI")

	sessionCompletion := func(ctx carapace.Context) carapace.Action {
		from := ""
		for i, arg := range ctx.Args {
			if arg == "--from" && i+1 < len(ctx.Args) {
				from = ctx.Args[i+1]
			}
		}
		if from == "" {
			return carapace.ActionMessage("pass --from first")
		}
		norm, err := discover.NormalizeCLI(from)
		if err != nil {
			return carapace.ActionMessage(err.Error())
		}
		files, err := discover.ListSessionPaths(norm, "")
		if err != nil {
			return carapace.ActionMessage(err.Error())
		}
		const cap = 30
		vals := make([]string, 0, cap)
		for i, p := range files {
			if i >= cap {
				break
			}
			vals = append(vals, filepath.Base(p))
		}
		return carapace.ActionValues(vals...).Tag("session")
	}

	cwdCompletion := carapace.ActionCallback(func(c carapace.Context) carapace.Action {
		pwd, _ := os.Getwd()
		return carapace.ActionValues(pwd).Tag("cwd")
	})

	for _, name := range []string{"list", "show", "transfer"} {
		sub, _, _ := root.Find([]string{name})
		if sub == nil {
			continue
		}
		carapace.Gen(sub).FlagCompletion(carapace.ActionMap{
			"from":    completeFrom,
			"to":      completeTo,
			"session": carapace.ActionCallback(sessionCompletion),
			"cwd":     cwdCompletion,
		})
	}
}
