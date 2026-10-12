package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/mlaccetti/harness-bridge/internal/discover"
	"github.com/mlaccetti/harness-bridge/internal/parser"
)

func newListCmd() *cobra.Command {
	var fromCLI string
	var cwd string
	var limit int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List recent sessions for one or all supported CLIs",
		Example: `  harness-bridge list
  harness-bridge list --from cursor --cwd /path/to/project --limit 5`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			targets := SourceCLIs
			if fromCLI != "" {
				norm, err := discover.NormalizeCLI(fromCLI)
				if err != nil {
					return err
				}
				targets = []string{norm}
			}
			for _, cli := range targets {
				files, err := discover.ListSessionPaths(cli, cwd)
				if err != nil {
					writef(cmd.ErrOrStderr(), "[%s] error: %v\n", cli, err)
					continue
				}
				writef(cmd.OutOrStdout(), "\n== %s (%d sessions) ==\n", cli, len(files))
				n := limit
				if n > len(files) {
					n = len(files)
				}
				for i := 0; i < n; i++ {
					p := files[i]
					info, err := os.Stat(p)
					if err != nil {
						continue
					}
					sess, _ := parser.Parse(cli, p)
					hint := TaskHint(sess)
					cwdKnown := sess.CWD
					if cwdKnown == "" {
						cwdKnown = "(unknown)"
					}
					out := cmd.OutOrStdout()
					writef(out, "  %10s  %s  (%s bytes)\n", fmtAge(info.ModTime()), info.Name(), formatBytes(info.Size()))
					writef(out, "             %s\n", p)
					writef(out, "             cwd: %s  turns: %d", cwdKnown, len(sess.Turns))
					if hint != "" {
						writef(out, "  hint: %s", hint)
					}
					writeln(out)
				}
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&fromCLI, "from", "", "show only this CLI (claude, codex, gemini, cursor, antigravity; agy → antigravity)")
	cmd.Flags().StringVar(&cwd, "cwd", "", "filter to sessions for this working directory (where supported)")
	cmd.Flags().IntVar(&limit, "limit", 10, "max sessions per CLI")
	return cmd
}

func fmtAge(t time.Time) string {
	delta := time.Since(t)
	switch {
	case delta < time.Minute:
		return fmt.Sprintf("%ds ago", int(delta.Seconds()))
	case delta < time.Hour:
		return fmt.Sprintf("%dm ago", int(delta.Minutes()))
	case delta < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(delta.Hours()))
	default:
		return fmt.Sprintf("%dd ago", int(delta.Hours()/24))
	}
}

func formatBytes(n int64) string {
	return fmt.Sprintf("%d", n)
}
