package discover

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mlaccetti/harness-bridge/internal/paths"
)

type SessionFile struct {
	Path      string
	CWD       string
	TurnCount int
	TaskHint  string
}

func NormalizeCLI(cli string) (string, error) {
	cli = strings.ToLower(strings.TrimSpace(cli))
	if cli == "agy" {
		cli = "antigravity"
	}
	switch cli {
	case "claude", "codex", "gemini", "cursor", "antigravity":
		return cli, nil
	default:
		return "", fmt.Errorf("unknown CLI: %q", cli)
	}
}

func ListSessionPaths(cli, cwd string) ([]string, error) {
	norm, err := NormalizeCLI(cli)
	if err != nil {
		return nil, err
	}
	switch norm {
	case "claude":
		if cwd != "" {
			return paths.ClaudeSessionsForCWD(cwd), nil
		}
		return sortedAll(filepath.Join(paths.ClaudeProjectsRoot(), "*", "*.jsonl"))
	case "codex":
		return paths.CodexAllSessions(), nil
	case "gemini":
		if cwd != "" {
			res := paths.GeminiSessionsForCWD(cwd)
			if len(res) > 0 {
				return res, nil
			}
		}
		return paths.GeminiAllSessions(), nil
	case "cursor":
		if cwd != "" {
			return paths.CursorSessionsForCWD(cwd), nil
		}
		return paths.CursorAllSessions(), nil
	case "antigravity":
		if cwd != "" {
			return paths.AntigravitySessionsForCWD(cwd), nil
		}
		return paths.AntigravityAllSessions(), nil
	default:
		return nil, fmt.Errorf("unknown CLI: %q", cli)
	}
}

func FindSession(cli, sessionIDOrPrefix, cwd string) (string, error) {
	norm, err := NormalizeCLI(cli)
	if err != nil {
		return "", err
	}
	if sessionIDOrPrefix != "" {
		candidates, err := ListSessionPaths(norm, "")
		if err != nil {
			return "", err
		}
		needle := strings.ToLower(sessionIDOrPrefix)
		for _, p := range candidates {
			if strings.Contains(strings.ToLower(filepath.Base(p)), needle) {
				return p, nil
			}
		}
		return "", nil
	}
	candidates, err := ListSessionPaths(norm, cwd)
	if err != nil {
		return "", err
	}
	if len(candidates) > 0 {
		return candidates[0], nil
	}
	candidates, err = ListSessionPaths(norm, "")
	if err != nil {
		return "", err
	}
	if len(candidates) > 0 {
		return candidates[0], nil
	}
	return "", nil
}

func sortedAll(pattern string) ([]string, error) {
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}
	return sortByMtime(matches), nil
}

func sortByMtime(paths []string) []string {
	type pair struct {
		path string
		mt   int64
	}
	var pairs []pair
	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			continue
		}
		pairs = append(pairs, pair{p, info.ModTime().UnixNano()})
	}
	for i := 0; i < len(pairs); i++ {
		for j := i + 1; j < len(pairs); j++ {
			if pairs[j].mt > pairs[i].mt {
				pairs[i], pairs[j] = pairs[j], pairs[i]
			}
		}
	}
	out := make([]string, len(pairs))
	for i, p := range pairs {
		out[i] = p.path
	}
	return out
}
