package writer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mlaccetti/harness-bridge/internal/render"
	"github.com/mlaccetti/harness-bridge/internal/session"
)

type markdownWriter struct{}

func NewMarkdownWriter() Writer { return &markdownWriter{} }

func (w *markdownWriter) CLI() string { return "markdown" }

func (w *markdownWriter) Write(sess session.Session, cwd string) (string, string, error) {
	outPath := filepath.Join(cwd, "HANDOFF.md")
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return "", "", err
	}

	nowISO := time.Now().UTC().Format(time.RFC3339)
	var lines []string
	lines = append(lines, "# Session handoff", "")
	lines = append(lines, fmt.Sprintf(
		"_Bridged from **%s** session `%s` at %s._",
		sess.SourceCLI, sess.SourceSessionID, nowISO,
	))
	lines = append(lines, "", "> Read this file end-to-end. The user was working with another CLI "+
		"and switched mid-task. Tool calls from the previous CLI are described "+
		"in prose; do not try to replay them. Continue from where the user "+
		"last spoke.", "")
	if sess.CWD != "" {
		lines = append(lines, fmt.Sprintf("Original cwd: `%s`", sess.CWD), "")
	}
	lines = append(lines, fmt.Sprintf("Total turns: **%d**", len(sess.Turns)), "", "---", "")

	for i, t := range sess.Turns {
		body := render.RenderTurnText(t)
		if body == "" {
			continue
		}
		roleLabel := "User"
		if t.Role == session.RoleAssistant {
			roleLabel = "Assistant"
		}
		lines = append(lines, fmt.Sprintf("## [%d] %s", i, roleLabel), "", body, "")
	}

	if err := os.WriteFile(outPath, []byte(strings.Join(lines, "\n")), 0o644); err != nil {
		return "", "", err
	}
	return outPath, outPath, nil
}
