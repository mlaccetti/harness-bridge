package parser

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/mlaccetti/harness-bridge/internal/session"
)

type claudeParser struct {
	cli string
}

func NewClaudeParser() Parser { return &claudeParser{cli: "claude"} }

func (p *claudeParser) CLI() string { return p.cli }

func (p *claudeParser) Parse(path string) (session.Session, error) {
	return parseClaudeJSONL(path, p.cli)
}

func parseClaudeJSONL(path, sourceCLI string) (session.Session, error) {
	f, err := os.Open(path)
	if err != nil {
		return session.Session{}, err
	}
	defer func() { _ = f.Close() }()

	var turns []session.Turn
	var compactions []session.CompactionEvent
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
		if sessionID == "" {
			sessionID = stringField(rec, "sessionId")
		}
		if cwd == "" {
			cwd = stringField(rec, "cwd")
		}
		if startedAt == "" {
			startedAt = stringField(rec, "timestamp")
		}

		rtype := stringField(rec, "type")
		if rtype == "system" && stringField(rec, "subtype") == "compact_boundary" {
			meta, _ := rec["compactMetadata"].(map[string]interface{})
			ev := session.CompactionEvent{
				TurnIndex: len(turns),
				Timestamp: stringField(rec, "timestamp"),
				Trigger:   stringField(meta, "trigger"),
			}
			if v, ok := meta["preTokens"].(float64); ok {
				ev.PreTokens = int(v)
			}
			if v, ok := meta["preTokens"].(json.Number); ok {
				if i, err := v.Int64(); err == nil {
					ev.PreTokens = int(i)
				}
			}
			if v, ok := meta["postTokens"].(float64); ok {
				ev.PostTokens = int(v)
			}
			if v, ok := meta["postTokens"].(json.Number); ok {
				if i, err := v.Int64(); err == nil {
					ev.PostTokens = int(i)
				}
			}
			compactions = append(compactions, ev)
			continue
		}

		role := ""
		var content interface{}
		if rtype == "user" || rtype == "assistant" {
			role = rtype
			msg, _ := rec["message"].(map[string]interface{})
			if msg != nil {
				if r := stringField(msg, "role"); r == "user" || r == "assistant" {
					role = r
				}
				content = msg["content"]
			}
		} else if r := stringField(rec, "role"); r == "user" || r == "assistant" {
			role = r
			msg, _ := rec["message"].(map[string]interface{})
			if msg != nil {
				content = msg["content"]
			}
		} else {
			continue
		}

		text, summaries := stringifyClaudeContent(content)
		if role == "user" && text == "" && toolResultOnly(content) {
			continue
		}
		if role == "user" && isNoiseUserTurn(text) && len(summaries) == 0 {
			continue
		}
		if text == "" && len(summaries) == 0 {
			continue
		}
		coalesceTurn(&turns, session.Turn{
			Role:          session.Role(role),
			Text:          text,
			ToolSummaries: summaries,
			Timestamp:     stringField(rec, "timestamp"),
		})
	}

	base := strings.TrimSuffix(filepath.Base(path), ".jsonl")
	return session.Session{
		SourceCLI:       sourceCLI,
		SourceSessionID: orDefault(sessionID, base),
		CWD:             cwd,
		StartedAt:       startedAt,
		Turns:           turns,
		Compactions:     compactions,
	}, sc.Err()
}

func stringifyClaudeContent(content interface{}) (string, []string) {
	if content == nil {
		return "", nil
	}
	if s, ok := content.(string); ok {
		return strings.TrimSpace(s), nil
	}
	items, ok := content.([]interface{})
	if !ok {
		return "", nil
	}
	var textParts []string
	var summaries []string
	for _, item := range items {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		t := stringField(m, "type")
		switch t {
		case "text":
			txt := strings.TrimSpace(stringField(m, "text"))
			if txt != "" {
				textParts = append(textParts, txt)
			}
		case "tool_use":
			name := stringField(m, "name", "Tool")
			inp, _ := m["input"].(map[string]interface{})
			if inp == nil {
				inp = map[string]interface{}{}
			}
			summaries = append(summaries, SummarizeClaudeTool(name, inp))
		case "tool_result":
			inner := m["content"]
			if list, ok := inner.([]interface{}); ok {
				for _, sub := range list {
					sm, ok := sub.(map[string]interface{})
					if !ok {
						continue
					}
					if stringField(sm, "type") == "text" {
						body := strings.TrimSpace(stringField(sm, "text"))
						if len(body) > 0 && strings.Contains(strings.ToLower(body[:min(80, len(body))]), "error") {
							summaries = append(summaries, "  ↳ tool error: "+truncate(body, 120))
						}
					}
				}
			}
		}
	}
	return strings.Join(textParts, "\n\n"), summaries
}

func toolResultOnly(content interface{}) bool {
	items, ok := content.([]interface{})
	if !ok || len(items) == 0 {
		return false
	}
	for _, item := range items {
		m, ok := item.(map[string]interface{})
		if !ok || stringField(m, "type") != "tool_result" {
			return false
		}
	}
	return true
}

func coalesceTurn(turns *[]session.Turn, t session.Turn) {
	if len(*turns) > 0 && (*turns)[len(*turns)-1].Role == t.Role {
		last := &(*turns)[len(*turns)-1]
		if t.Text != "" {
			if last.Text != "" {
				last.Text = strings.TrimSpace(last.Text + "\n\n" + t.Text)
			} else {
				last.Text = t.Text
			}
		}
		last.ToolSummaries = append(last.ToolSummaries, t.ToolSummaries...)
		return
	}
	*turns = append(*turns, t)
}

func stringField(m map[string]interface{}, key string, def ...string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	if len(def) > 0 {
		return def[0]
	}
	return ""
}

func orDefault(s, def string) string {
	if s != "" {
		return s
	}
	return def
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
