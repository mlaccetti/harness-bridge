package cli

import (
	"regexp"
	"strings"

	"github.com/mlaccetti/harness-bridge/internal/session"
)

var (
	reUserQuery  = regexp.MustCompile(`<user_query>\s*([\s\S]*?)\s*</user_query>`)
	reTimestamp  = regexp.MustCompile(`<timestamp>[\s\S]*?</timestamp>`)
	reWhitespace = regexp.MustCompile(`\s+`)
)

func TaskHint(sess session.Session) string {
	for _, t := range sess.Turns {
		if t.Role != session.RoleUser || strings.TrimSpace(t.Text) == "" {
			continue
		}
		text := stripMeta(t.Text)
		line := firstLine(text)
		if line != "" {
			return truncate(line, 80)
		}
	}
	for i := len(sess.Turns) - 1; i >= 0; i-- {
		t := sess.Turns[i]
		if t.Role == session.RoleUser && strings.TrimSpace(t.Text) != "" {
			line := firstLine(stripMeta(t.Text))
			if line != "" {
				return "last: " + truncate(line, 72)
			}
		}
	}
	return ""
}

func stripMeta(text string) string {
	text = reTimestamp.ReplaceAllString(text, "")
	if m := reUserQuery.FindStringSubmatch(text); len(m) > 1 {
		text = m[1]
	}
	return strings.TrimSpace(reWhitespace.ReplaceAllString(text, " "))
}

func firstLine(s string) string {
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			return line
		}
	}
	return ""
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}
