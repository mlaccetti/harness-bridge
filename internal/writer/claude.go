package writer

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"

	"github.com/mlaccetti/harness-bridge/internal/paths"
	"github.com/mlaccetti/harness-bridge/internal/render"
	"github.com/mlaccetti/harness-bridge/internal/session"
)

type claudeWriter struct {
	cli string
}

func NewClaudeWriter() Writer { return &claudeWriter{cli: "claude"} }

func (w *claudeWriter) CLI() string { return w.cli }

func (w *claudeWriter) Write(sess session.Session, cwd string) (string, string, error) {
	return writeClaudeJSONL(sess, cwd, w.cli, paths.ClaudeProjectDir)
}

func writeClaudeJSONL(sess session.Session, cwd, sourceCLI string, projectDir func(string) string) (string, string, error) {
	newSessionID := uuid.New().String()
	dir := projectDir(cwd)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", "", err
	}
	outPath := filepath.Join(dir, newSessionID+".jsonl")
	nowISO := time.Now().UTC().Format(time.RFC3339)

	var records []map[string]interface{}
	var parent *string

	push := func(role string, text string) {
		recUUID := uuid.New().String()
		rec := map[string]interface{}{
			"parentUuid":  parent,
			"isSidechain": false,
			"type":        role,
			"timestamp":   nowISO,
			"sessionId":   newSessionID,
			"cwd":         cwd,
			"version":     "harness-bridge/0.1.0",
			"uuid":        recUUID,
			"message": map[string]interface{}{
				"role": role,
				"content": []map[string]string{
					{"type": "text", "text": text},
				},
			},
		}
		records = append(records, rec)
		parent = &recUUID
	}

	push("user", render.BridgeBanner)
	for _, t := range sess.Turns {
		body := render.RenderTurnText(t)
		if body == "" {
			continue
		}
		role := "user"
		if t.Role == session.RoleAssistant {
			role = "assistant"
		}
		push(role, body)
	}

	f, err := os.Create(outPath)
	if err != nil {
		return "", "", err
	}
	enc := json.NewEncoder(f)
	enc.SetEscapeHTML(false)
	for _, r := range records {
		if err := enc.Encode(r); err != nil {
			_ = f.Close()
			return "", "", err
		}
	}
	if err := f.Close(); err != nil {
		return "", "", err
	}
	_ = sourceCLI
	return outPath, newSessionID, nil
}
