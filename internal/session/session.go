package session

import "time"

type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleSystem    Role = "system"
)

type Turn struct {
	Role          Role
	Text          string
	ToolSummaries []string
	Timestamp     string
}

type Session struct {
	SourceCLI       string
	SourceSessionID string
	CWD             string
	StartedAt       string
	Turns           []Turn
	Compactions     []CompactionEvent
}

func NowISO() string {
	return time.Now().UTC().Format(time.RFC3339)
}
