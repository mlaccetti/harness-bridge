package writer

import (
	"github.com/mlaccetti/harness-bridge/internal/paths"
	"github.com/mlaccetti/harness-bridge/internal/session"
)

type antigravityWriter struct{}

func NewAntigravityWriter() Writer { return &antigravityWriter{} }

func (w *antigravityWriter) CLI() string { return "antigravity" }

func (w *antigravityWriter) Write(sess session.Session, cwd string) (string, string, error) {
	return writeClaudeJSONL(sess, cwd, "antigravity", paths.AntigravityProjectDir)
}
