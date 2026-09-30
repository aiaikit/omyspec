package version

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/aiaikit/speckit/internal/dlsec"
)

// GitHubRelease is a minimal subset of the GitHub Releases API response.
// Mirrors Python's `(tag_name, None)` tuple.
type GitHubRelease struct {
	TagName string `json:"tag_name"`
	Name    string `json:"name"`
}

// FetchLatestRelease hits the GitHub Releases API and returns the parsed
// tag. Caller passes a dlsec.Client (already configured with HTTPS-only
// rules). The URL is validated through dlsec.ValidateURL before fetch.
func FetchLatestRelease(ctx context.Context, client *dlsec.Client, url string) (*GitHubRelease, error) {
	if err := dlsec.ValidateURL(url); err != nil {
		return nil, err
	}
	resp, err := client.Get(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("fetch %s: HTTP %d", url, resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, dlsec.MaxJSONCatalogBytes+1))
	if err != nil {
		return nil, err
	}
	if len(body) > dlsec.MaxJSONCatalogBytes {
		return nil, fmt.Errorf("fetch %s: response body exceeds %d bytes", url, dlsec.MaxJSONCatalogBytes)
	}
	var rel GitHubRelease
	if err := json.Unmarshal(body, &rel); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}
	return &rel, nil
}
