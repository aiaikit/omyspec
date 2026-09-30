// Package dlsec provides a hardened HTTP client for fetching remote
// resources safely (HTTPS-only, size caps, redirect policy).
//
// stub — see Task 12. Task 11 (version.FetchLatestRelease) depends on
// Client + ValidateURL + Get + MaxJSONCatalogBytes. The full implementation
// replaces this stub.
package dlsec

import (
	"context"
	"errors"
	"net/http"
	"net/url"
)

// MaxDownloadBytes is the hard ceiling for any single body returned by Get.
const MaxDownloadBytes = 50 << 20

// MaxJSONCatalogBytes is the ceiling for parsed JSON catalogs (e.g. release
// manifests). Smaller than MaxDownloadBytes because JSON is parsed into
// memory.
const MaxJSONCatalogBytes = 5 << 20

// Client wraps http.Client with HTTPS-only enforcement and response-size
// limits. Task 12 expands this with a real configuration struct.
type Client struct {
	HTTP *http.Client
}

// NewClient returns a Client with default settings.
func NewClient() *Client {
	return &Client{HTTP: http.DefaultClient}
}

// ValidateURL rejects URLs that are not https:// except for localhost
// loopback addresses (http://localhost, http://127.0.0.1, http://[::1])
// which are allowed for tests.
func ValidateURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return err
	}
	switch u.Scheme {
	case "https":
		return nil
	case "http":
		host := u.Hostname()
		if host == "localhost" || host == "127.0.0.1" || host == "::1" {
			return nil
		}
		return errors.New("http scheme only allowed for loopback addresses")
	default:
		return errors.New("unsupported scheme: " + u.Scheme)
	}
}

// Get performs an HTTP GET through the configured client. Body limit and
// redirect policy are not yet enforced in this stub.
func (c *Client) Get(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	return c.HTTP.Do(req)
}
