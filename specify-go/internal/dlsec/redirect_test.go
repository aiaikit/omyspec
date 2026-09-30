package dlsec

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRedirectStripAuthOnCrossHost(t *testing.T) {
	target := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "" {
			t.Error("expected Authorization on first hop")
		}
		w.Header().Set("Location", "https://other-host.example/foo")
		w.WriteHeader(http.StatusFound)
	}))
	defer target.Close()

	// We're going to redirect to target.URL but rewrite Host header.
	// Easier: just verify the stripAuthOnRedirect function directly.
	if !stripAuthOnRedirect("https://first.example", "https://other.example/path") {
		t.Error("expected strip on cross-host https->https")
	}
}

func TestRedirectStripAuthOnHTTPDowngrade(t *testing.T) {
	if !stripAuthOnRedirect("https://example.com", "http://example.com/foo") {
		t.Error("expected strip on https->http downgrade")
	}
}

func TestRedirectKeepAuthSameHost(t *testing.T) {
	if stripAuthOnRedirect("https://example.com/a", "https://example.com/b") {
		t.Error("expected KEEP on same-host https->https")
	}
}
