package dlsec

import (
	"fmt"
	"net/url"
	"strings"
)

// ValidateURL accepts only https://, http://localhost, http://127.0.0.1,
// or http://[::1]. Mirrors Python's `_validate_https_url` with the
// localhost exception.
func ValidateURL(rawURL string) error {
	if rawURL == "" {
		return fmt.Errorf("dlsec: empty URL")
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("dlsec: parse %q: %w", rawURL, err)
	}
	scheme := strings.ToLower(u.Scheme)
	host := strings.ToLower(u.Hostname())
	switch scheme {
	case "https":
		return nil
	case "http":
		if host == "localhost" || host == "127.0.0.1" || host == "::1" {
			return nil
		}
		return fmt.Errorf("dlsec: http only allowed for localhost, got %q", host)
	default:
		return fmt.Errorf("dlsec: scheme %q not allowed", scheme)
	}
}
