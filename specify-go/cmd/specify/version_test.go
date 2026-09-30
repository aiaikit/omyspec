package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aiaikit/speckit/internal/dlsec"
	"github.com/aiaikit/speckit/internal/version"
	"github.com/spf13/cobra"
)

func TestVersionCmd_JSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(version.GitHubRelease{TagName: "v99.0.0"})
	}))
	defer srv.Close()

	origVersion := version.Version
	version.Version = "v1.0.0"
	defer func() { version.Version = origVersion }()

	cmd := versionCmdTestable(srv.URL)
	var outStr strings.Builder
	cmd.SetOut(&outStr)
	cmd.SetErr(&nopWriter{})

	cmd.SetArgs([]string{"--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute error: %v", err)
	}

	var out map[string]any
	if err := json.Unmarshal([]byte(outStr.String()), &out); err != nil {
		t.Fatalf("output is not valid JSON: %s", outStr.String())
	}
	if out["version"] != "v1.0.0" {
		t.Errorf("version = %v, want v1.0.0", out["version"])
	}
	if out["installMethod"] == "" {
		t.Error("installMethod should not be empty")
	}
	if out["latestRelease"] != "v99.0.0" {
		t.Errorf("latestRelease = %v, want v99.0.0", out["latestRelease"])
	}
}

func versionCmdTestable(releasesURL string) *cobra.Command {
	var jsonFlag bool
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print version and exit",
		RunE: func(cmd *cobra.Command, args []string) error {
			v := version.Version
			method := version.DetectInstallMethod()
			client := dlsec.NewClient()

			if jsonFlag {
				type versionOutput struct {
					Version       string `json:"version"`
					InstallMethod string `json:"installMethod"`
					LatestRelease string `json:"latestRelease,omitempty"`
				}
				out := versionOutput{Version: v, InstallMethod: method.String()}
				rel, err := version.FetchLatestRelease(context.Background(), client, releasesURL)
				if err == nil {
					out.LatestRelease = rel.TagName
				}
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", " ")
				return enc.Encode(out)
			}
			cmd.OutOrStdout().Write([]byte("text mode not tested here\n"))
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonFlag, "json", false, "Output as JSON")
	return cmd
}

type nopWriter struct{}

func (w *nopWriter) Write(p []byte) (int, error) { return len(p), nil }
