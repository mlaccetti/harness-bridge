// Package version exposes the CLI release identity (ldflags at release build, module version for go install).
package version

import (
	"fmt"
	"runtime/debug"
	"strings"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// Info is the resolved version, commit, and build date shown by --version.
type Info struct {
	Version string
	Commit  string
	Date    string
}

// Get returns build metadata from release ldflags, falling back to debug.BuildInfo.
func Get() Info {
	v, c, d := version, commit, date
	if bi, ok := debug.ReadBuildInfo(); ok {
		if v == "" || v == "dev" {
			if mv := moduleVersion(bi); mv != "" {
				v = strings.TrimPrefix(mv, "v")
			}
		}
		if c == "" || c == "none" {
			c = buildSetting(bi, "vcs.revision")
		}
		if d == "" || d == "unknown" {
			d = buildSetting(bi, "vcs.time")
		}
	}
	if v == "" {
		v = "dev"
	}
	if c == "" {
		c = "none"
	}
	if d == "" {
		d = "unknown"
	}
	return Info{Version: v, Commit: c, Date: d}
}

func moduleVersion(bi *debug.BuildInfo) string {
	if bi.Main.Version != "" && bi.Main.Version != "(devel)" {
		return bi.Main.Version
	}
	const module = "github.com/mlaccetti/harness-bridge"
	for _, dep := range bi.Deps {
		if dep.Path == module && dep.Version != "" {
			return dep.Version
		}
	}
	return ""
}

func buildSetting(bi *debug.BuildInfo, key string) string {
	for _, s := range bi.Settings {
		if s.Key == key {
			return s.Value
		}
	}
	return ""
}

// Line formats output for Cobra --version (includes trailing newline).
func (i Info) Line() string {
	if i.Commit != "" && i.Commit != "none" {
		return fmt.Sprintf("%s (%s, %s)\n", i.Version, i.Commit, i.Date)
	}
	return i.Version + "\n"
}

// ToolLabel is stamped into transferred session metadata (e.g. harness-bridge/0.1.0).
func ToolLabel() string {
	return "harness-bridge/" + Get().Version
}
