package cli

import (
	"fmt"
	"io"

	"github.com/mlaccetti/harness-bridge/internal/session"
	"github.com/mlaccetti/harness-bridge/internal/writer"
)

type sliceFlags struct {
	lastTurns       int
	sinceCompaction bool
}

func (f sliceFlags) opts() session.SliceOptions {
	return session.SliceOptions{
		LastTurns:       f.lastTurns,
		SinceCompaction: f.sinceCompaction,
	}
}

func writeStats(w io.Writer, srcPath string, sess session.Session, slice sliceFlags, dryRun bool) {
	full := session.ComputeStats(sess)
	writef(w, "Source file: %s\n", srcPath)
	writef(w, "Session:     %s  id=%s\n", sess.SourceCLI, sess.SourceSessionID)
	writef(w, "CWD:         %s\n", orUnknown(sess.CWD))
	writef(w, "Parsed:      %d turns, ~%s chars (~%s tokens est.), %d tool summaries\n",
		full.Turns, formatSize(full.Chars), formatSize(full.EstTokens), full.ToolCalls)

	if len(sess.Compactions) > 0 {
		writef(w, "Compactions: %d Claude compact_boundary event(s)\n", len(sess.Compactions))
		last := sess.Compactions[len(sess.Compactions)-1]
		writef(w, "  latest:    %s\n", session.CompactionEventNote(last))
	} else if sess.SourceCLI == "claude" {
		writeln(w, "Compactions: none found in JSONL")
	}

	writeln(w, "")
	writeln(w, "Suggested slices:")
	for _, s := range session.SuggestSlices(sess) {
		writef(w, "  %s  → %d turns (~%s chars)\n", s.Flag, s.Turns, formatSize(s.Chars))
		if s.Note != "" {
			writef(w, "             %s\n", s.Note)
		}
	}

	sliced, report := sess.ApplySlice(slice.opts())
	if slice.lastTurns > 0 || slice.sinceCompaction {
		sel := session.ComputeStats(sliced)
		writeln(w, "")
		writef(w, "Active slice: %s (turns %d..%d → %d turns, ~%s chars)\n",
			report.Method, report.StartIndex, report.OriginalTurns-1, sel.Turns, formatSize(sel.Chars))
		if report.Note != "" {
			writef(w, "  %s\n", report.Note)
		}
	}

	if dryRun {
		writeln(w, "")
		writeln(w, "(dry-run: no files written)")
	}
}

func writeTransferPlan(w io.Writer, to, targetCWD string, sess session.Session, slice sliceFlags, dryRun bool) {
	sliced, report := sess.ApplySlice(slice.opts())
	sel := session.ComputeStats(sliced)
	writeln(w, "")
	if dryRun {
		writeln(w, "Transfer plan (dry-run)")
	} else {
		writeln(w, "Transfer")
	}
	writef(w, "  Target CLI:  %s\n", to)
	writef(w, "  Target cwd:  %s\n", targetCWD)
	writef(w, "  Output:      %s\n", writer.DescribeTarget(to, targetCWD))
	writef(w, "  Turns out:   %d (from %d parsed, slice=%s)\n", sel.Turns, len(sess.Turns), report.Method)
	if !dryRun {
		writeln(w, "")
		writeln(w, "Resume hint will be printed after a successful write.")
	}
}

func formatSize(n int) string {
	if n < 1024 {
		return fmt.Sprintf("%d", n)
	}
	if n < 1024*1024 {
		return fmt.Sprintf("%.1fK", float64(n)/1024)
	}
	return fmt.Sprintf("%.1fM", float64(n)/(1024*1024))
}
