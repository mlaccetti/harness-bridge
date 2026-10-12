package parser

import (
	"fmt"

	"github.com/mlaccetti/harness-bridge/internal/session"
)

type Parser interface {
	CLI() string
	Parse(path string) (session.Session, error)
}

var parsers = map[string]Parser{
	"claude":      NewClaudeParser(),
	"codex":       NewCodexParser(),
	"gemini":      NewGeminiParser(),
	"cursor":      NewCursorParser(),
	"antigravity": NewAntigravityParser(),
}

func Get(cli string) (Parser, error) {
	p, ok := parsers[cli]
	if !ok {
		return nil, fmt.Errorf("unknown source CLI: %q", cli)
	}
	return p, nil
}

func Parse(cli, path string) (session.Session, error) {
	p, err := Get(cli)
	if err != nil {
		return session.Session{}, err
	}
	return p.Parse(path)
}
