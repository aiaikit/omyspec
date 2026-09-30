package version

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aiaikit/speckit/internal/dlsec"
)

func TestParseGitHubRelease(t *testing.T) {
	raw := `{"tag_name": "v1.2.3", "name": "Release 1.2.3"}`
	var rel GitHubRelease
	if err := json.Unmarshal([]byte(raw), &rel); err != nil {
		t.Fatalf("err: %v", err)
	}
	if rel.TagName != "v1.2.3" {
		t.Errorf("TagName = %q, want v1.2.3", rel.TagName)
	}
}

func TestFetchLatestRelease_OK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(GitHubRelease{TagName: "v2.0.0"})
	}))
	defer srv.Close()

	// httptest.NewServer uses HTTP on loopback; dlsec.ValidateURL accepts
	// http://127.0.0.1 for tests.
	client := dlsec.NewClient()
	rel, err := FetchLatestRelease(context.Background(), client, srv.URL)
	if err != nil {
		t.Fatalf("FetchLatestRelease: %v", err)
	}
	if rel.TagName != "v2.0.0" {
		t.Errorf("TagName = %q, want v2.0.0", rel.TagName)
	}
}

func TestFetchLatestRelease_HTTPNotAllowed(t *testing.T) {
	// Use a non-localhost HTTP URL to confirm ValidateURL rejects.
	client := dlsec.NewClient()
	_, err := FetchLatestRelease(context.Background(), client, "http://example.com/releases/latest")
	if err == nil {
		t.Error("expected error on http://example.com, got nil")
	}
}

func TestFetchLatestRelease_ReadsResponseBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"tag_name": "v9.9.9"}`)
	}))
	defer srv.Close()

	client := dlsec.NewClient()
	rel, err := FetchLatestRelease(context.Background(), client, srv.URL)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if rel.TagName != "v9.9.9" {
		t.Errorf("TagName = %q, want v9.9.9", rel.TagName)
	}
}
