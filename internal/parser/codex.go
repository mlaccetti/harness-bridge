package parser

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/mlaccetti/harness-bridge/internal/session"
)

type codexParser struct{}

func NewCodexParser() Parser { return &codexParser{} }

func (p *codexParser) CLI() string { return "codex" }

func (p *codexParser) Parse(path string) (session.Session, error) {
	f, err := os.Open(path)
	if err != nil {
		return session.Session{}, err
	}
	defer func() { _ = f.Close() }()

	var turns []session.Turn
	var sessionID, cwd, startedAt string

	sc := newJSONLScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var rec map[string]interface{}
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			continue
		}
		rtype := stringField(rec, "type")
		payload, _ := rec["payload"].(map[string]interface{})
		if payload == nil {
			payload = map[string]interface{}{}
		}

		switch rtype {
		case "session_meta":
			if sessionID == "" {
				sessionID = stringField(payload, "id")
			}
			if cwd == "" {
				cwd = stringField(payload, "cwd")
			}
			if startedAt == "" {
				startedAt = orDefault(stringField(payload, "timestamp"), stringField(rec, "timestamp"))
			}
			continue
		case "response_item":
			if stringField(payload, "type") != "message" {
				continue
			}
			role := stringField(payload, "role")
			if role != "user" && role != "assistant" {
				continue
			}
			content, _ := payload["content"].([]interface{})
			var textParts []string
			for _, item := range content {
				m, ok := item.(map[string]interface{})
				if !ok {
					continue
				}
				ct := stringField(m, "type")
				if ct == "input_text" || ct == "output_text" || ct == "text" {
					txt := strings.TrimSpace(stringField(m, "text"))
					if txt != "" {
						textParts = append(textParts, txt)
					}
				}
			}
			text := strings.TrimSpace(strings.Join(textParts, "\n\n"))
			if text == "" {
				continue
			}
			coalesceTurn(&turns, session.Turn{
				Role:      session.Role(role),
				Text:      text,
				Timestamp: stringField(rec, "timestamp"),
			})
		case "event_msg":
			if stringField(payload, "type") != "function_call" {
				continue
			}
			args := map[string]interface{}{}
			if raw := payload["arguments"]; raw != nil {
				if s, ok := raw.(string); ok {
					_ = json.Unmarshal([]byte(s), &args)
				} else if m, ok := raw.(map[string]interface{}); ok {
					args = m
				}
			}
			summary := SummarizeCodexFunction(stringField(payload, "name", "tool"), args)
			if len(turns) > 0 && turns[len(turns)-1].Role == session.RoleAssistant {
				turns[len(turns)-1].ToolSummaries = append(turns[len(turns)-1].ToolSummaries, summary)
			} else {
				turns = append(turns, session.Turn{
					Role:          session.RoleAssistant,
					ToolSummaries: []string{summary},
				})
			}
		}
	}

	base := strings.TrimSuffix(filepath.Base(path), ".jsonl")
	return session.Session{
		SourceCLI:       "codex",
		SourceSessionID: orDefault(sessionID, base),
		CWD:             cwd,
		StartedAt:       startedAt,
		Turns:           turns,
	}, sc.Err()
}
