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

type geminiWriter struct{}

func NewGeminiWriter() Writer { return &geminiWriter{} }

func (w *geminiWriter) CLI() string { return "gemini" }

func (w *geminiWriter) Write(sess session.Session, cwd string) (string, string, error) {
	newSessionID := uuid.New().String()
	friendly := paths.GeminiProjectFolderName(cwd)
	_ = paths.GeminiRegisterProject(cwd, friendly)

	projectDir := filepath.Join(paths.GeminiRoot(), "tmp", friendly)
	chatsDir := filepath.Join(projectDir, "chats")
	if err := os.MkdirAll(chatsDir, 0o755); err != nil {
		return "", "", err
	}

	rootMarker := filepath.Join(projectDir, ".project_root")
	if _, err := os.Stat(rootMarker); err != nil {
		_ = os.WriteFile(rootMarker, []byte(cwd), 0o644)
	}

	historyDir := paths.GeminiHistoryDir(cwd)
	_ = os.MkdirAll(historyDir, 0o755)
	histMarker := filepath.Join(historyDir, ".project_root")
	if _, err := os.Stat(histMarker); err != nil {
		_ = os.WriteFile(histMarker, []byte(cwd), 0o644)
	}

	now := time.Now().UTC()
	fname := "session-" + now.Format("2006-01-02T15-04") + "-" + newSessionID[:8] + ".json"
	outPath := filepath.Join(chatsDir, fname)
	nowISO := now.Format("2006-01-02T15:04:05.000Z")

	messages := []map[string]interface{}{
		{
			"id":        uuid.New().String(),
			"timestamp": nowISO,
			"type":      "user",
			"content":   render.BridgeBanner,
		},
	}

	for _, t := range sess.Turns {
		body := render.RenderTurnText(t)
		if body == "" {
			continue
		}
		geminiType := "user"
		if t.Role == session.RoleAssistant {
			geminiType = "gemini"
		}
		messages = append(messages, map[string]interface{}{
			"id":        uuid.New().String(),
			"timestamp": time.Now().UTC().Format("2006-01-02T15:04:05.000Z"),
			"type":      geminiType,
			"content":   body,
		})
	}

	data := map[string]interface{}{
		"sessionId":   newSessionID,
		"projectHash": friendly,
		"startTime":   nowISO,
		"lastUpdated": nowISO,
		"messages":    messages,
	}

	raw, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", "", err
	}
	if err := os.WriteFile(outPath, raw, 0o644); err != nil {
		return "", "", err
	}
	return outPath, newSessionID, nil
}
