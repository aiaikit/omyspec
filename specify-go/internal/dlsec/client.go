package dlsec

import (
	"context"
	"net/http"
	"time"
)

// Client is a size-capped, HTTPS-only HTTP client.
// Phase 1 implements Get with the redirect-stripping rule from Task 13;
// Task 13 wires the redirect handler.
type Client struct {
	httpClient *http.Client
	timeout    time.Duration
}

// Option configures Client construction.
type Option func(*Client)

// WithTimeout sets the per-request timeout.
func WithTimeout(d time.Duration) Option {
	return func(c *Client) { c.timeout = d }
}

// DefaultTimeout is used when WithTimeout is not specified.
const DefaultTimeout = 30 * time.Second

// NewClient constructs a Client with the given options.
func NewClient(opts ...Option) *Client {
	c := &Client{timeout: DefaultTimeout}
	for _, opt := range opts {
		opt(c)
	}
	c.httpClient = &http.Client{Timeout: c.timeout}
	return c
}

// Get performs an HTTPS-only GET with size cap.
// Phase 1 returns the raw response; Phase 2 (Task 13) wires redirect
// stripping and response-body size limit.
func (c *Client) Get(ctx context.Context, rawURL string) (*http.Response, error) {
	if err := ValidateURL(rawURL); err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	return c.httpClient.Do(req)
}
