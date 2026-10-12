package paths

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
)

func Home() string {
	return os.Getenv("HOME")
}

// ── Claude Code ──────────────────────────────────────────────────────────

func ClaudeRoot() string {
	return filepath.Join(Home(), ".claude")
}

func ClaudeProjectsRoot() string {
	return filepath.Join(ClaudeRoot(), "projects")
}

var claudeEncodeRe = regexp.MustCompile(`[\\/:\s]`)

func ClaudeEncodeCWD(cwd string) string {
	encoded := claudeEncodeRe.ReplaceAllString(cwd, "-")
	return strings.TrimRight(encoded, "-")
}

func ClaudeProjectDir(cwd string) string {
	encoded := ClaudeEncodeCWD(cwd)
	root := ClaudeProjectsRoot()
	entries, err := os.ReadDir(root)
	if err != nil {
		return filepath.Join(root, encoded)
	}
	encodedLower := strings.ToLower(encoded)
	for _, e := range entries {
		if e.IsDir() && strings.EqualFold(e.Name(), encodedLower) {
			return filepath.Join(root, e.Name())
		}
	}
	return filepath.Join(root, encoded)
}

func ClaudeSessionsForCWD(cwd string) []string {
	return sortedGlobNewest(filepath.Join(ClaudeProjectDir(cwd), "*.jsonl"))
}

// ── Codex CLI ────────────────────────────────────────────────────────────

func CodexRoot() string {
	return filepath.Join(Home(), ".codex")
}

func CodexSessionsRoot() string {
	return filepath.Join(CodexRoot(), "sessions")
}

func CodexSessionPathForNow(sessionUUID string) string {
	now := time.Now()
	folder := filepath.Join(CodexSessionsRoot(),
		now.Format("2006"), now.Format("01"), now.Format("02"))
	ts := now.Format("2006-01-02T15-04-05")
	return filepath.Join(folder, "rollout-"+ts+"-"+sessionUUID+".jsonl")
}

func CodexAllSessions() []string {
	return sortedRglobNewest(CodexSessionsRoot(), "rollout-*.jsonl")
}

// ── Gemini CLI ───────────────────────────────────────────────────────────

func GeminiRoot() string {
	return filepath.Join(Home(), ".gemini")
}

func GeminiProjectsIndexPath() string {
	return filepath.Join(GeminiRoot(), "projects.json")
}

func GeminiProjectFolderName(cwd string) string {
	cwdKey := cwd
	if isWindows() {
		cwdKey = strings.ToLower(strings.ReplaceAll(cwd, "/", "\\"))
	}
	idx := GeminiProjectsIndexPath()
	data, err := os.ReadFile(idx)
	if err == nil {
		var parsed struct {
			Projects map[string]string `json:"projects"`
		}
		if json.Unmarshal(data, &parsed) == nil {
			if name, ok := parsed.Projects[cwdKey]; ok {
				return name
			}
			for k, v := range parsed.Projects {
				if strings.EqualFold(k, cwdKey) {
					return v
				}
			}
		}
	}
	return filepath.Base(cwd)
}

func GeminiRegisterProject(cwd, friendlyName string) error {
	idx := GeminiProjectsIndexPath()
	var data map[string]interface{}
	if raw, err := os.ReadFile(idx); err == nil {
		_ = json.Unmarshal(raw, &data)
	}
	if data == nil {
		data = map[string]interface{}{}
	}
	projects, _ := data["projects"].(map[string]interface{})
	if projects == nil {
		projects = map[string]interface{}{}
	}
	raw := strings.TrimSpace(cwd)
	cwdKey := raw
	if isWindows() {
		cwdKey = strings.ToLower(strings.ReplaceAll(raw, "/", "\\"))
	}
	projects[cwdKey] = friendlyName
	data["projects"] = projects
	if err := os.MkdirAll(filepath.Dir(idx), 0o755); err != nil {
		return err
	}
	out, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(idx, out, 0o644)
}

func GeminiChatsDir(cwd string) string {
	return filepath.Join(GeminiRoot(), "tmp", GeminiProjectFolderName(cwd), "chats")
}

func GeminiHistoryDir(cwd string) string {
	return filepath.Join(GeminiRoot(), "history", GeminiProjectFolderName(cwd))
}

func GeminiSessionsForCWD(cwd string) []string {
	return sortedGlobNewest(filepath.Join(GeminiChatsDir(cwd), "session-*.json"))
}

func GeminiAllSessions() []string {
	return sortedGlobNewest(filepath.Join(GeminiRoot(), "tmp", "*", "chats", "session-*.json"))
}

// ── Cursor Agent ─────────────────────────────────────────────────────────

func CursorRoot() string {
	return filepath.Join(Home(), ".cursor")
}

func CursorProjectsRoot() string {
	return filepath.Join(CursorRoot(), "projects")
}

func CursorEncodeWorkspace(cwd string) string {
	s := strings.TrimPrefix(filepath.Clean(cwd), string(filepath.Separator))
	s = strings.ReplaceAll(s, string(filepath.Separator), "-")
	return s
}

func CursorProjectDir(cwd string) string {
	encoded := CursorEncodeWorkspace(cwd)
	root := CursorProjectsRoot()
	entries, err := os.ReadDir(root)
	if err != nil {
		return filepath.Join(root, encoded)
	}
	for _, e := range entries {
		if e.IsDir() && strings.EqualFold(e.Name(), encoded) {
			return filepath.Join(root, e.Name())
		}
	}
	return filepath.Join(root, encoded)
}

func cursorTranscriptDirs(projectDir string) []string {
	base := filepath.Join(projectDir, "agent-transcripts")
	entries, err := os.ReadDir(base)
	if err != nil {
		return nil
	}
	var dirs []string
	for _, e := range entries {
		if e.IsDir() {
			dirs = append(dirs, filepath.Join(base, e.Name()))
		}
	}
	return dirs
}

func CursorSessionsForCWD(cwd string) []string {
	return cursorSessionFiles(CursorProjectDir(cwd))
}

func CursorAllSessions() []string {
	root := CursorProjectsRoot()
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil
	}
	var files []string
	for _, e := range entries {
		if e.IsDir() {
			files = append(files, cursorSessionFiles(filepath.Join(root, e.Name()))...)
		}
	}
	return sortByMtimeDesc(files)
}

func cursorSessionFiles(projectDir string) []string {
	var files []string
	for _, dir := range cursorTranscriptDirs(projectDir) {
		id := filepath.Base(dir)
		candidate := filepath.Join(dir, id+".jsonl")
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			files = append(files, candidate)
		}
	}
	return sortByMtimeDesc(files)
}

// ── Antigravity (agy) ────────────────────────────────────────────────────
// JSONL layout matches Claude Code; roots vary by platform.

func AntigravityRoot() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		p := filepath.Join(xdg, "antigravity")
		if pathExists(p) {
			return p
		}
	}
	candidates := []string{
		filepath.Join(Home(), ".config", "antigravity"),
		filepath.Join(Home(), ".antigravity"),
	}
	if isDarwin() {
		candidates = append(candidates, filepath.Join(Home(), "Library", "Application Support", "Antigravity"))
	}
	for _, c := range candidates {
		if pathExists(c) {
			return c
		}
	}
	return candidates[0]
}

func AntigravityProjectsRoot() string {
	return filepath.Join(AntigravityRoot(), "projects")
}

func AntigravityEncodeCWD(cwd string) string {
	return ClaudeEncodeCWD(cwd)
}

func AntigravityProjectDir(cwd string) string {
	encoded := AntigravityEncodeCWD(cwd)
	root := AntigravityProjectsRoot()
	entries, err := os.ReadDir(root)
	if err != nil {
		return filepath.Join(root, encoded)
	}
	for _, e := range entries {
		if e.IsDir() && strings.EqualFold(e.Name(), encoded) {
			return filepath.Join(root, e.Name())
		}
	}
	return filepath.Join(root, encoded)
}

func AntigravitySessionsForCWD(cwd string) []string {
	return sortedGlobNewest(filepath.Join(AntigravityProjectDir(cwd), "*.jsonl"))
}

func AntigravityAllSessions() []string {
	return sortedGlobNewest(filepath.Join(AntigravityProjectsRoot(), "*", "*.jsonl"))
}

func isWindows() bool {
	return os.PathSeparator == '\\'
}

func isDarwin() bool {
	return runtime.GOOS == "darwin"
}

func pathExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

func sortedGlobNewest(pattern string) []string {
	matches, _ := filepath.Glob(pattern)
	return sortByMtimeDesc(matches)
}

func sortedRglobNewest(root, pattern string) []string {
	var matches []string
	_ = walkFiles(root, func(path string, _ os.DirEntry) error {
		if matchGlob(filepath.Base(pattern), filepath.Base(path)) {
			matches = append(matches, path)
		}
		return nil
	})
	return sortByMtimeDesc(matches)
}

func walkFiles(root string, fn func(string, os.DirEntry) error) error {
	entries, err := os.ReadDir(root)
	if err != nil {
		return err
	}
	for _, e := range entries {
		path := filepath.Join(root, e.Name())
		if e.IsDir() {
			_ = walkFiles(path, fn)
			continue
		}
		_ = fn(path, e)
	}
	return nil
}

func matchGlob(pattern, name string) bool {
	ok, _ := filepath.Match(pattern, name)
	return ok
}

func sortByMtimeDesc(paths []string) []string {
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
