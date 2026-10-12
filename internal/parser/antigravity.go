package parser

import "github.com/mlaccetti/harness-bridge/internal/session"

type antigravityParser struct{}

func NewAntigravityParser() Parser { return &antigravityParser{} }

func (p *antigravityParser) CLI() string { return "antigravity" }

func (p *antigravityParser) Parse(path string) (session.Session, error) {
	return parseClaudeJSONL(path, "antigravity")
}
