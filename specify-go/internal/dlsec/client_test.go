package dlsec

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewClient_DefaultTimeout(t *testing.T) {
	c := NewClient()
	if c.timeout != DefaultTimeout {
		t.Errorf("timeout = %v, want %v", c.timeout, DefaultTimeout)
	}
	if c.httpClient == nil {
		t.Error("httpClient is nil")
	}
	if c.httpClient.Timeout != DefaultTimeout {
		t.Errorf("httpClient.Timeout = %v, want %v", c.httpClient.Timeout, DefaultTimeout)
	}
}

func TestNewClient_WithTimeout(t *testing.T) {
	c := NewClient(WithTimeout(5 * 1_000_000_000))
	if c.timeout != 5*1_000_000_000 {
		t.Errorf("timeout = %v, want 5s", c.timeout)
	}
	if c.httpClient.Timeout != 5*1_000_000_000 {
		t.Errorf("httpClient.Timeout = %v, want 5s", c.httpClient.Timeout)
	}
}

func TestClient_Get_RejectsBadURL(t *testing.T) {
	c := NewClient()
	_, err := c.Get(context.Background(), "http://example.com/foo")
	if err == nil {
		t.Error("expected error for http://example.com, got nil")
	}
}

func TestClient_Get_RejectsEmptyURL(t *testing.T) {
	c := NewClient()
	_, err := c.Get(context.Background(), "")
	if err == nil {
		t.Error("expected error for empty URL, got nil")
	}
}

func TestClient_Get_OK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "pong")
	}))
	defer srv.Close()

	c := NewClient()
	resp, err := c.Get(context.Background(), srv.URL)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if string(body) != "pong" {
		t.Errorf("body = %q, want pong", string(body))
	}
}

func TestClient_Get_ContextCancelled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "pong")
	}))
	defer srv.Close()

	c := NewClient()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := c.Get(ctx, srv.URL)
	if err == nil {
		t.Error("expected error from cancelled context, got nil")
	}
	if !strings.Contains(err.Error(), "context canceled") {
		t.Logf("err: %v (expected context-canceled wording to be implementation-dependent)", err)
	}
}
