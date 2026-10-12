package parser

import (
	"regexp"
	"strings"
)

var reNoiseTag = regexp.MustCompile(`^<[a-z0-9-]+`)

// isNoiseUserTurn drops Claude/Cursor hook and task status user lines from transcripts.
func isNoiseUserTurn(text string) bool {
	s := strings.TrimSpace(text)
	if s == "" {
		return false
	}
	if strings.HasPrefix(s, "<task-notification") {
		return true
	}
	if strings.HasPrefix(s, "<local-command-caveat") {
		return true
	}
	tag := reNoiseTag.FindString(s)
	if tag == "" {
		return false
	}
	switch tag {
	case "<task-notification", "<local-command-caveat", "<command-name", "<command-message",
		"<command-args", "<command-stdout", "<command-stderr", "<command-finished",
		"<command-status", "<command-duration", "<command-cwd", "<system-reminder":
		return true
	default:
		return strings.HasPrefix(tag, "<command-")
	}
}
