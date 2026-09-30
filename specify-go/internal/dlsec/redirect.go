package dlsec

import (
	"net/url"
	"strings"
)

// stripAuthOnRedirect returns true if the Authorization header should be
// stripped when redirecting from prevURL to newURL.
//
// Rules (mirror Python's `_StripAuthOnRedirect` in authentication/http.py):
//  1. Cross-host: ALWAYS strip.
//  2. Same-host, scheme downgrade https->http: strip.
//  3. Same-host, same scheme: keep.
func stripAuthOnRedirect(prevURL, newURL string) bool {
	pu, err := url.Parse(prevURL)
	if err != nil {
		return true // fail closed: strip
	}
	nu, err := url.Parse(newURL)
	if err != nil {
		return true
	}
	if !strings.EqualFold(pu.Hostname(), nu.Hostname()) {
		return true
	}
	if pu.Scheme == "https" && nu.Scheme == "http" {
		return true
	}
	return false
}
