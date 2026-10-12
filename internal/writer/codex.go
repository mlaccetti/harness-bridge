package writer

import (
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/mlaccetti/harness-bridge/internal/paths"
	"github.com/mlaccetti/harness-bridge/internal/render"
	"github.com/mlaccetti/harness-bridge/internal/session"
	"github.com/mlaccetti/harness-bridge/internal/version"

	_ "modernc.org/sqlite"
)

type codexWriter struct{}

func NewCodexWriter() Writer { return &codexWriter{} }

func (w *codexWriter) CLI() string { return "codex" }

func (w *codexWriter) Write(sess session.Session, cwd string) (string, string, error) {
	newSessionID := uuid.New().String()
	outPath := paths.CodexSessionPathForNow(newSessionID)
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return "", "", err
	}

	now := time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
	var records []map[string]interface{}

	records = append(records, map[string]interface{}{
		"timestamp": now,
		"type":      "session_meta",
		"payload": map[string]interface{}{
			"id":          newSessionID,
			"timestamp":   now,
			"cwd":         cwd,
			"originator":  "harness_bridge",
			"cli_version": version.ToolLabel(),
			"source":      "harness_bridge",
		},
	})

	msg := func(role, text, contentType string) map[string]interface{} {
		ts := time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
		return map[string]interface{}{
			"timestamp": ts,
			"type":      "response_item",
			"payload": map[string]interface{}{
				"type":    "message",
				"role":    role,
				"content": []map[string]string{{"type": contentType, "text": text}},
			},
		}
	}

	records = append(records, msg("user", render.BridgeBanner, "input_text"))
	for _, t := range sess.Turns {
		body := render.RenderTurnText(t)
		if body == "" {
			continue
		}
		if t.Role == session.RoleAssistant {
			records = append(records, msg("assistant", body, "output_text"))
		} else {
			records = append(records, msg("user", body, "input_text"))
		}
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

	title := bestCodexTitle(sess)
	appendSessionIndex(newSessionID, now, title)
	insertThreadRow(newSessionID, outPath, cwd, title)

	return outPath, newSessionID, nil
}

func bestCodexTitle(sess session.Session) string {
	reTags := []*regexp.Regexp{
		regexp.MustCompile(`<ide_selection>.*?</ide_selection>`),
		regexp.MustCompile(`<ide_opened_file>.*?</ide_opened_file>`),
		regexp.MustCompile(`<environment_context>.*?</environment_context>`),
		regexp.MustCompile(`<system-reminder>.*?</system-reminder>`),
	}
	for _, t := range sess.Turns {
		if t.Role != session.RoleUser || t.Text == "" {
			continue
		}
		if strings.Contains(t.Text, "[harness-bridge]") || strings.Contains(t.Text[:min(30, len(t.Text))], "session-bridge") {
			continue
		}
		stripped := t.Text
		for _, re := range reTags {
			stripped = re.ReplaceAllString(stripped, "")
		}
		for _, line := range strings.Split(stripped, "\n") {
			line = strings.TrimSpace(line)
			if line != "" && !strings.HasPrefix(line, "<") && !strings.Contains(line, "[harness-bridge]") {
				if len(line) > 60 {
					return line[:60]
				}
				return line
			}
		}
	}
	return "Bridged from " + sess.SourceCLI
}

func appendSessionIndex(sessionID, timestamp, title string) {
	idxPath := filepath.Join(paths.CodexRoot(), "session_index.jsonl")
	_ = os.MkdirAll(filepath.Dir(idxPath), 0o755)
	entry := map[string]string{
		"id":          sessionID,
		"thread_name": title,
		"updated_at":  timestamp,
	}
	f, err := os.OpenFile(idxPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	enc := json.NewEncoder(f)
	_ = enc.Encode(entry)
	_ = f.Close()
}

func insertThreadRow(sessionID, rolloutPath, cwd, title string) {
	dbPath := filepath.Join(paths.CodexRoot(), "state_5.sqlite")
	if _, err := os.Stat(dbPath); err != nil {
		return
	}
	nowS := time.Now().Unix()
	nowMS := time.Now().UnixMilli()
	cwdNorm := strings.ReplaceAll(cwd, "/", "\\")
	if !strings.HasPrefix(cwdNorm, "\\\\?\\") {
		cwdNorm = "\\\\?\\" + cwdNorm
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return
	}
	defer func() { _ = db.Close() }()

	var existingID string
	err = db.QueryRow("SELECT id FROM threads WHERE id = ?", sessionID).Scan(&existingID)
	exists := err == nil

	toolVersion := version.ToolLabel()
	sandbox := `{"type":"workspace-write","writable_roots":[],"network_access":false,"exclude_tmpdir_env_var":false,"exclude_slash_tmp":false}`
	if exists {
		_, _ = db.Exec(`
			UPDATE threads SET title = ?, first_user_message = ?, has_user_event = 1,
			updated_at = ?, updated_at_ms = ?, rollout_path = ?,
			cli_version = COALESCE(NULLIF(cli_version, ''), ?)
			WHERE id = ?`, title, title, nowS, nowMS, rolloutPath, toolVersion, sessionID)
		return
	}
	_, _ = db.Exec(`
		INSERT INTO threads (
			id, rollout_path, created_at, updated_at, source,
			model_provider, cwd, title, sandbox_policy,
			approval_mode, tokens_used, has_user_event, archived,
			cli_version, first_user_message, memory_mode,
			created_at_ms, updated_at_ms
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0, 1, 0, ?, ?, 'enabled', ?, ?)`,
		sessionID, rolloutPath, nowS, nowS, "harness_bridge",
		"openai", cwdNorm, title, sandbox, "on-request",
		toolVersion, title, nowMS, nowMS)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
