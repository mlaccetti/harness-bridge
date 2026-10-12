package writer

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/google/uuid"

	"github.com/mlaccetti/harness-bridge/internal/paths"
	"github.com/mlaccetti/harness-bridge/internal/render"
	"github.com/mlaccetti/harness-bridge/internal/session"
)

type cursorWriter struct{}

func NewCursorWriter() Writer { return &cursorWriter{} }

func (w *cursorWriter) CLI() string { return "cursor" }

func (w *cursorWriter) Write(sess session.Session, cwd string) (string, string, error) {
	newSessionID := uuid.New().String()
	projectDir := paths.CursorProjectDir(cwd)
	transcriptDir := filepath.Join(projectDir, "agent-transcripts", newSessionID)
	if err := os.MkdirAll(transcriptDir, 0o755); err != nil {
		return "", "", err
	}
	outPath := filepath.Join(transcriptDir, newSessionID+".jsonl")

	f, err := os.Create(outPath)
	if err != nil {
		return "", "", err
	}
	enc := json.NewEncoder(f)
	enc.SetEscapeHTML(false)

	writeLine := func(role, text string) error {
		rec := map[string]interface{}{
			"role": role,
			"message": map[string]interface{}{
				"content": []map[string]string{{"type": "text", "text": text}},
			},
		}
		return enc.Encode(rec)
	}

	if err := writeLine("user", render.BridgeBanner); err != nil {
		_ = f.Close()
		return "", "", err
	}
	for _, t := range sess.Turns {
		body := render.RenderTurnText(t)
		if body == "" {
			continue
		}
		role := "user"
		if t.Role == session.RoleAssistant {
			role = "assistant"
		}
		if err := writeLine(role, body); err != nil {
			_ = f.Close()
			return "", "", err
		}
	}
	if err := enc.Encode(map[string]interface{}{"type": "turn_ended", "status": "success"}); err != nil {
		_ = f.Close()
		return "", "", err
	}
	if err := f.Close(); err != nil {
		return "", "", err
	}
	return outPath, newSessionID, nil
}
