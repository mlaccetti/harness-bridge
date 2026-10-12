package parser

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/mlaccetti/harness-bridge/internal/session"
)

type geminiParser struct{}

func NewGeminiParser() Parser { return &geminiParser{} }

func (p *geminiParser) CLI() string { return "gemini" }

func (p *geminiParser) Parse(path string) (session.Session, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return session.Session{}, err
	}
	var data map[string]interface{}
	if err := json.Unmarshal(raw, &data); err != nil {
		return session.Session{}, err
	}

	roleMap := map[string]session.Role{
		"user":      session.RoleUser,
		"gemini":    session.RoleAssistant,
		"model":     session.RoleAssistant,
		"assistant": session.RoleAssistant,
	}

	var turns []session.Turn
	messages, _ := data["messages"].([]interface{})
	for _, msg := range messages {
		m, ok := msg.(map[string]interface{})
		if !ok {
			continue
		}
		mtype := stringField(m, "type")
		role, ok := roleMap[mtype]
		if !ok {
			continue
		}
		content := m["content"]
		var text string
		switch v := content.(type) {
		case string:
			text = strings.TrimSpace(v)
		default:
			if content != nil {
				b, _ := json.Marshal(content)
				text = strings.TrimSpace(string(b))
			}
		}
		if text == "" {
			continue
		}
		coalesceTurn(&turns, session.Turn{
			Role:      role,
			Text:      text,
			Timestamp: stringField(m, "timestamp"),
		})
	}

	base := strings.TrimSuffix(filepath.Base(path), ".json")
	return session.Session{
		SourceCLI:       "gemini",
		SourceSessionID: orDefault(stringField(data, "sessionId"), base),
		StartedAt:       stringField(data, "startTime"),
		Turns:           turns,
	}, nil
}
