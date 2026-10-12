package session

import (
	"strings"
	"testing"
	"time"
)

func TestNowISO_utcRFC3339(t *testing.T) {
	got := NowISO()
	if !strings.HasSuffix(got, "Z") && !strings.Contains(got, "+") {
		t.Fatalf("not RFC3339-ish: %q", got)
	}
	_, err := time.Parse(time.RFC3339, got)
	if err != nil {
		t.Fatal(err)
	}
}
