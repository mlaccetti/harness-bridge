// Package upgrade checks GitHub for a newer release and prints a one-line Homebrew hint.
//
// We only ping the API once per process, only for release builds installed via Homebrew,
// and only when stderr is a TTY. Set HARNESS_BRIDGE_NO_UPGRADE_NOTICE=1 to disable.
package upgrade

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	githubLatestRelease = "https://api.github.com/repos/mlaccetti/harness-bridge/releases/latest"
	checkTimeout        = 3 * time.Second
)

var (
	once           sync.Once
	client         = &http.Client{Timeout: checkTimeout}
	assumeHomebrew bool
)

// SetAssumeHomebrewInstall forces the Homebrew install check (tests only).
func SetAssumeHomebrewInstall(v bool) {
	assumeHomebrew = v
}

// MaybeSuggestUpgrade compares current with the latest GitHub release and prints a hint to w.
func MaybeSuggestUpgrade(ctx context.Context, current string, args []string, w io.Writer) {
	if !shouldCheck(current, args, w) {
		return
	}
	once.Do(func() {
		suggest(ctx, current, w)
	})
}

func shouldCheck(current string, args []string, w io.Writer) bool {
	if current == "" || current == "dev" {
		return false
	}
	if os.Getenv("HARNESS_BRIDGE_NO_UPGRADE_NOTICE") == "1" {
		return false
	}
	if !installedViaHomebrew() {
		return false
	}
	if f, ok := w.(*os.File); ok && !isTerminal(f) {
		return false
	}
	return !argsSkipUpgradeCheck(args)
}

func argsSkipUpgradeCheck(args []string) bool {
	for i := 1; i < len(args); i++ {
		a := args[i]
		switch a {
		case "--version", "-v", "--help", "-h":
			return true
		case "--":
			return false
		}
		if strings.HasPrefix(a, "-") && strings.Contains(a, "version") {
			return true
		}
		if a == "completion" || strings.HasPrefix(a, "__complete") {
			return true
		}
	}
	return false
}

func installedViaHomebrew() bool {
	if assumeHomebrew {
		return true
	}
	exe, err := os.Executable()
	if err != nil {
		return false
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return false
	}
	exe = filepath.Clean(exe)
	return strings.Contains(exe, filepath.Join("Cellar", "harness-bridge"+string(filepath.Separator)))
}

func isTerminal(f *os.File) bool {
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return (info.Mode() & os.ModeCharDevice) != 0
}

func suggest(ctx context.Context, current string, w io.Writer) {
	latest, err := fetchLatestTag(ctx)
	if err != nil || latest == "" {
		return
	}
	if !versionLess(current, latest) {
		return
	}
	// Standalone shell comment so the upgrade line can be copied from scripts or docs.
	_, _ = fmt.Fprintf(w, "# harness-bridge %s is available (you have %s). Run: brew update && brew upgrade harness-bridge\n",
		stripV(latest), stripV(current))
}

func fetchLatestTag(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, githubLatestRelease, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "harness-bridge")

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = res.Body.Close() }()
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("github releases: %s", res.Status)
	}

	var body struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		return "", err
	}
	return body.TagName, nil
}

func stripV(v string) string {
	return strings.TrimPrefix(strings.TrimSpace(v), "v")
}

func versionLess(current, latest string) bool {
	return compareSemver(stripV(current), stripV(latest)) < 0
}

func compareSemver(a, b string) int {
	pa := parseParts(a)
	pb := parseParts(b)
	for i := 0; i < 3; i++ {
		if pa[i] != pb[i] {
			return cmpInt(pa[i], pb[i])
		}
	}
	return 0
}

func parseParts(v string) [3]int {
	var out [3]int
	parts := strings.SplitN(v, "-", 2)[0]
	seg := strings.Split(parts, ".")
	for i := 0; i < 3 && i < len(seg); i++ {
		out[i] = atoi(seg[i])
	}
	return out
}

func atoi(s string) int {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			break
		}
		n = n*10 + int(c-'0')
	}
	return n
}

func cmpInt(a, b int) int {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}

// SetHTTPClient replaces the default client (tests only).
func SetHTTPClient(c *http.Client) {
	client = c
}
