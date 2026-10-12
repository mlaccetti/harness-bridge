package upgrade

import (
	"context"
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

func TestMaybeSuggestUpgrade_skipsDevBuild(t *testing.T) {
	var buf bytes.Buffer
	MaybeSuggestUpgrade(context.Background(), "dev", []string{"harness-bridge", "list"}, &buf)
	if buf.Len() != 0 {
		t.Fatalf("expected no output for dev build, got %q", buf.String())
	}
}

func TestMaybeSuggestUpgrade_skipsWhenCurrent(t *testing.T) {
	SetAssumeHomebrewInstall(true)
	defer SetAssumeHomebrewInstall(false)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"tag_name":"v0.1.0"}`))
	}))
	defer srv.Close()

	client = redirectClient(srv.URL)
	once = sync.Once{}

	var buf bytes.Buffer
	MaybeSuggestUpgrade(context.Background(), "0.1.0", []string{"harness-bridge", "list"}, &buf)
	if buf.Len() != 0 {
		t.Fatalf("expected no output when up to date, got %q", buf.String())
	}
}

func TestMaybeSuggestUpgrade_printsCommentWhenBehind(t *testing.T) {
	SetAssumeHomebrewInstall(true)
	defer SetAssumeHomebrewInstall(false)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"tag_name":"v0.2.0"}`))
	}))
	defer srv.Close()

	client = redirectClient(srv.URL)
	once = sync.Once{}

	var buf bytes.Buffer
	MaybeSuggestUpgrade(context.Background(), "0.1.0", []string{"harness-bridge", "list"}, &buf)
	out := buf.String()
	if !strings.HasPrefix(out, "# harness-bridge ") {
		t.Fatalf("expected standalone comment, got %q", out)
	}
	if !strings.Contains(out, "brew update && brew upgrade harness-bridge") {
		t.Fatalf("missing brew upgrade hint: %q", out)
	}
}

func Test_compareSemver(t *testing.T) {
	if compareSemver("0.1.0", "0.2.0") >= 0 {
		t.Fatal("0.1.0 should be less than 0.2.0")
	}
	if compareSemver("1.0.0", "0.9.9") <= 0 {
		t.Fatal("1.0.0 should be greater than 0.9.9")
	}
}

func redirectClient(srvURL string) *http.Client {
	return &http.Client{
		Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			req.URL.Scheme = "http"
			req.URL.Host = strings.TrimPrefix(srvURL, "http://")
			return http.DefaultTransport.RoundTrip(req)
		}),
		Timeout: checkTimeout,
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
