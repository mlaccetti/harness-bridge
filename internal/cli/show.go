package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mlaccetti/harness-bridge/internal/discover"
	"github.com/mlaccetti/harness-bridge/internal/parser"
	"github.com/mlaccetti/harness-bridge/internal/session"
)

func newShowCmd() *cobra.Command {
	var fromCLI string
	var sessionID string
	var cwd string
	var stats bool
	var showAll bool
	var previewLimit int
	var slice sliceFlags

	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show a parsed summary and turn preview for a session",
		Example: `  harness-bridge show --from claude --session 3fa85f64
  harness-bridge show --from claude --stats
  harness-bridge show --from claude --all`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			norm, err := discover.NormalizeCLI(fromCLI)
			if err != nil {
				return err
			}
			path, err := discover.FindSession(norm, sessionID, cwd)
			if err != nil {
				return err
			}
			if path == "" {
				msg := fmt.Sprintf("No session found for %s", norm)
				if sessionID != "" {
					msg += fmt.Sprintf(" matching %q", sessionID)
				}
				return fmt.Errorf("%s", msg)
			}
			sess, err := parser.Parse(norm, path)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()

			if stats {
				writeln(out, "Session stats")
				writeln(out, "--------------")
				writeStats(out, path, sess, slice, false)
				writeln(out, "")
			}

			display, _ := sess.ApplySlice(slice.opts())
			afterSlice := len(display.Turns)
			preview := display.Turns
			if !showAll {
				limit := previewLimit
				if limit <= 0 {
					limit = session.DefaultShowPreviewTurns
				}
				preview = session.TailTurns(display.Turns, limit)
			}

			writef(out, "Source: %s  id=%s\n", sess.SourceCLI, sess.SourceSessionID)
			writef(out, "CWD:    %s\n", orUnknown(sess.CWD))
			writef(out, "Turns:  %d parsed", len(sess.Turns))
			if afterSlice != len(sess.Turns) {
				writef(out, ", %d after slice", afterSlice)
			}
			if len(preview) < afterSlice {
				writef(out, ", showing last %d", len(preview))
			}
			writeln(out, "")
			if len(preview) < afterSlice {
				writef(out, "(use --all for full preview, or --preview-limit N)\n\n")
			}

			baseIndex := afterSlice - len(preview)
			for i, t := range preview {
				writef(out, "--- [%d] %s ---\n", baseIndex+i, t.Role)
				previewLine := "(no text)"
				if t.Text != "" {
					line := strings.Split(t.Text, "\n")[0]
					if len(line) > 160 {
						line = line[:160]
					}
					previewLine = line
				}
				writeln(out, previewLine)
				if len(t.ToolSummaries) > 0 {
					writef(out, "  +%d tool calls\n", len(t.ToolSummaries))
				}
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&fromCLI, "from", "", "source CLI (required)")
	cmd.Flags().StringVar(&sessionID, "session", "", "session id or filename substring (default: most recent)")
	cmd.Flags().StringVar(&cwd, "cwd", "", "restrict search to this cwd")
	cmd.Flags().BoolVar(&stats, "stats", false, "print size estimates, compaction info, and slice suggestions")
	cmd.Flags().BoolVar(&showAll, "all", false, "show every turn in the preview (default: last 25)")
	cmd.Flags().IntVar(&previewLimit, "preview-limit", session.DefaultShowPreviewTurns, "max turns to print in preview (with --all: ignored)")
	cmd.Flags().IntVar(&slice.lastTurns, "last-turns", 0, "slice session to last N turns before preview")
	cmd.Flags().BoolVar(&slice.sinceCompaction, "since-compaction", false, "slice to turns after last Claude compaction")
	_ = cmd.MarkFlagRequired("from")
	return cmd
}
