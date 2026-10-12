package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/mlaccetti/harness-bridge/internal/discover"
	"github.com/mlaccetti/harness-bridge/internal/parser"
	"github.com/mlaccetti/harness-bridge/internal/writer"
)

func newTransferCmd() *cobra.Command {
	var fromCLI string
	var toCLI string
	var sessionID string
	var cwd string
	var dryRun bool
	var plan bool
	var slice sliceFlags

	cmd := &cobra.Command{
		Use:   "transfer",
		Short: "Bridge a session from one CLI into another",
		Example: `  harness-bridge transfer --from claude --to codex
  harness-bridge transfer --from claude --to cursor --dry-run --since-compaction
  harness-bridge transfer --from cursor --to markdown --last-turns 50 --plan`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			from, err := discover.NormalizeCLI(fromCLI)
			if err != nil {
				return err
			}
			to := toCLI
			if to != "markdown" {
				to, err = discover.NormalizeCLI(toCLI)
				if err != nil {
					return err
				}
			}

			targetCWD := cwd
			if targetCWD == "" {
				targetCWD, err = os.Getwd()
				if err != nil {
					return err
				}
			}

			srcPath, err := discover.FindSession(from, sessionID, targetCWD)
			if err != nil {
				return err
			}
			if srcPath == "" {
				msg := fmt.Sprintf("No source session found for %s", from)
				if sessionID != "" {
					msg += fmt.Sprintf(" matching %q", sessionID)
				}
				writef(cmd.ErrOrStderr(), "%s\n", msg)
				writef(cmd.ErrOrStderr(), "Hint: run `harness-bridge list --from %s` to see what's available.\n", from)
				return fmt.Errorf("%s", msg)
			}

			out := cmd.OutOrStdout()
			sess, err := parser.Parse(from, srcPath)
			if err != nil {
				return err
			}

			isDry := dryRun || plan
			if isDry {
				writeln(out, "Session stats")
				writeln(out, "--------------")
				writeStats(out, srcPath, sess, slice, true)
				writeTransferPlan(out, to, targetCWD, sess, slice, true)
				return nil
			}

			writef(out, "Reading source: %s\n", srcPath)
			sliced, report := sess.ApplySlice(slice.opts())
			writef(out, "  parsed %d turns; bridging %d (%s); cwd in source: %s\n",
				len(sess.Turns), len(sliced.Turns), report.Method, orUnknown(sess.CWD))

			outPath, newID, err := writer.Write(to, sliced, targetCWD)
			if err != nil {
				return err
			}
			writef(out, "\nBridged into %s:\n", to)
			writef(out, "  %s\n", outPath)
			writef(out, "  session id: %s\n", newID)
			if hint := ResumeHint(to, newID); hint != "" {
				writeln(out, hint)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&fromCLI, "from", "", "source CLI (required)")
	cmd.Flags().StringVar(&toCLI, "to", "", "target CLI or markdown (required)")
	cmd.Flags().StringVar(&sessionID, "session", "", "source session id or substring (default: latest)")
	cmd.Flags().StringVar(&cwd, "cwd", "", "target cwd (default: current directory)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "show plan and stats without writing")
	cmd.Flags().BoolVar(&plan, "plan", false, "alias for --dry-run")
	cmd.Flags().IntVar(&slice.lastTurns, "last-turns", 0, "bridge only the last N conversational turns")
	cmd.Flags().BoolVar(&slice.sinceCompaction, "since-compaction", false, "bridge turns after the last Claude compaction boundary")
	_ = cmd.MarkFlagRequired("from")
	_ = cmd.MarkFlagRequired("to")
	return cmd
}
