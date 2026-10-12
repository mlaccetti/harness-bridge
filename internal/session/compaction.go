package session

// CompactionEvent marks where Claude Code compacted context in the JSONL timeline.
// TurnIndex is the first neutral turn to keep when bridging with --since-compaction.
type CompactionEvent struct {
	TurnIndex  int
	Timestamp  string
	Trigger    string
	PreTokens  int
	PostTokens int
}
