package session

type SliceOptions struct {
	LastTurns       int
	SinceCompaction bool
}

type SliceReport struct {
	OriginalTurns int
	SelectedTurns int
	StartIndex    int
	Method        string
	Note          string
}

func (s Session) ApplySlice(opts SliceOptions) (Session, SliceReport) {
	report := SliceReport{
		OriginalTurns: len(s.Turns),
		StartIndex:    0,
		Method:        "all",
	}
	if len(s.Turns) == 0 {
		report.SelectedTurns = 0
		return s, report
	}

	start := 0
	if opts.SinceCompaction && len(s.Compactions) > 0 {
		last := s.Compactions[len(s.Compactions)-1]
		start = last.TurnIndex
		if start > len(s.Turns) {
			start = len(s.Turns)
		}
		report.Method = "since-compaction"
		report.Note = CompactionEventNote(last)
	} else if opts.SinceCompaction {
		report.Method = "since-compaction"
		report.Note = "no compaction markers in source (Claude only); kept all turns"
	}

	if opts.LastTurns > 0 {
		tailStart := len(s.Turns) - opts.LastTurns
		if tailStart < 0 {
			tailStart = 0
		}
		if tailStart > start {
			start = tailStart
			report.Method = "last-turns"
			report.Note = ""
		} else if report.Method == "since-compaction" && opts.LastTurns > 0 {
			report.Method = "since-compaction+last-turns"
		}
	}

	if start > 0 {
		trimmed := s
		trimmed.Turns = s.Turns[start:]
		report.StartIndex = start
		report.SelectedTurns = len(trimmed.Turns)
		return trimmed, report
	}

	report.SelectedTurns = len(s.Turns)
	return s, report
}

func CompactionEventNote(ev CompactionEvent) string {
	note := "last Claude compaction"
	if ev.Timestamp != "" {
		note += " at " + ev.Timestamp
	}
	if ev.Trigger != "" {
		note += " (" + ev.Trigger + ")"
	}
	if ev.PreTokens > 0 && ev.PostTokens > 0 {
		note += "; context " + itoa(ev.PreTokens) + "→" + itoa(ev.PostTokens) + " tokens"
	}
	return note
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	if neg {
		b = append([]byte{'-'}, b...)
	}
	return string(b)
}

type Suggestion struct {
	Flag  string
	Turns int
	Chars int
	Note  string
}

func SuggestSlices(sess Session) []Suggestion {
	var out []Suggestion
	if len(sess.Compactions) > 0 {
		last := sess.Compactions[len(sess.Compactions)-1]
		sliced, _ := sess.ApplySlice(SliceOptions{SinceCompaction: true})
		st := ComputeStats(sliced)
		out = append(out, Suggestion{
			Flag:  "--since-compaction",
			Turns: st.Turns,
			Chars: st.Chars,
			Note:  CompactionEventNote(last),
		})
	}
	for _, n := range []int{30, 50, 100, 200} {
		if n >= len(sess.Turns) {
			continue
		}
		sliced, _ := sess.ApplySlice(SliceOptions{LastTurns: n})
		st := ComputeStats(sliced)
		out = append(out, Suggestion{
			Flag:  "--last-turns " + itoa(n),
			Turns: st.Turns,
			Chars: st.Chars,
		})
	}
	return out
}
