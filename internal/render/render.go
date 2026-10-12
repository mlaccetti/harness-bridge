package render

import (
	"strings"

	"github.com/mlaccetti/harness-bridge/internal/session"
)

const BridgeBanner = "[harness-bridge] This conversation was bridged from another CLI. " +
	"Tool calls from the previous CLI are described in prose below. " +
	"Continue from where the user last spoke; do not try to re-run those tool calls."

func RenderTurnText(t session.Turn) string {
	body := t.Text
	if len(t.ToolSummaries) > 0 {
		var bullets []string
		for _, s := range t.ToolSummaries {
			bullets = append(bullets, "  - "+s)
		}
		body = strings.TrimSpace(body + "\n\n_(actions taken in the previous session:)_\n" + strings.Join(bullets, "\n"))
	}
	return body
}
