package version

import (
	"strings"
	"testing"
)

func TestInfoLine_withCommit(t *testing.T) {
	line := Info{Version: "0.2.0", Commit: "abc123", Date: "2026-01-01T00:00:00Z"}.Line()
	if !strings.Contains(line, "0.2.0 (abc123,") {
		t.Fatalf("unexpected line: %q", line)
	}
}

func TestInfoLine_versionOnly(t *testing.T) {
	line := Info{Version: "0.2.0", Commit: "none", Date: "unknown"}.Line()
	if strings.TrimSpace(line) != "0.2.0" {
		t.Fatalf("unexpected line: %q", line)
	}
}

func TestToolLabel_usesResolvedVersion(t *testing.T) {
	// Default test binary has version=dev unless ldflags are set at compile time.
	label := ToolLabel()
	if !strings.HasPrefix(label, "harness-bridge/") {
		t.Fatalf("unexpected label: %q", label)
	}
}
