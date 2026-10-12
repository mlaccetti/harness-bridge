package session

type Stats struct {
	Turns     int
	Chars     int
	EstTokens int
	ToolCalls int
}

func renderedLen(t Turn) int {
	n := len(t.Text)
	for _, s := range t.ToolSummaries {
		n += len(s) + 32
	}
	return n
}

func ComputeStats(sess Session) Stats {
	var chars, tools int
	for _, t := range sess.Turns {
		chars += renderedLen(t)
		tools += len(t.ToolSummaries)
	}
	return Stats{
		Turns:     len(sess.Turns),
		Chars:     chars,
		EstTokens: chars / 4,
		ToolCalls: tools,
	}
}
