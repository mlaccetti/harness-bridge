package writer

import (
	"fmt"

	"github.com/mlaccetti/harness-bridge/internal/session"
)

type Writer interface {
	CLI() string
	Write(sess session.Session, cwd string) (outPath, newSessionID string, err error)
}

var writers = map[string]Writer{
	"claude":      NewClaudeWriter(),
	"codex":       NewCodexWriter(),
	"gemini":      NewGeminiWriter(),
	"cursor":      NewCursorWriter(),
	"antigravity": NewAntigravityWriter(),
	"markdown":    NewMarkdownWriter(),
}

func Get(cli string) (Writer, error) {
	w, ok := writers[cli]
	if !ok {
		return nil, fmt.Errorf("unknown target CLI: %q", cli)
	}
	return w, nil
}

func Write(cli string, sess session.Session, cwd string) (string, string, error) {
	w, err := Get(cli)
	if err != nil {
		return "", "", err
	}
	return w.Write(sess, cwd)
}
