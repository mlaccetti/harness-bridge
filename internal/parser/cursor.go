package parser

import "github.com/mlaccetti/harness-bridge/internal/session"

type cursorParser struct{}

func NewCursorParser() Parser {
	return &cursorParser{}
}

func (p *cursorParser) CLI() string { return "cursor" }

func (p *cursorParser) Parse(path string) (session.Session, error) {
	return parseClaudeJSONL(path, "cursor")
}
