package session

const DefaultShowPreviewTurns = 25

// TailTurns returns the last n turns (or all if n <= 0 or n >= len).
func TailTurns(turns []Turn, n int) []Turn {
	if n <= 0 || len(turns) <= n {
		return turns
	}
	return turns[len(turns)-n:]
}
