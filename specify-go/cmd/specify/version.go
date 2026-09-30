package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aiaikit/speckit/internal/dlsec"
	"github.com/aiaikit/speckit/internal/version"
	"github.com/spf13/cobra"
)

func VersionCmd() *cobra.Command {
	var jsonFlag bool

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print version and check for updates",
		RunE: func(cmd *cobra.Command, args []string) error {
			v := version.Version
			method := version.DetectInstallMethod()
			installMethodStr := method.String()

			if jsonFlag {
				type versionOutput struct {
					Version       string `json:"version"`
					InstallMethod string `json:"installMethod"`
					LatestRelease string `json:"latestRelease,omitempty"`
				}
				out := versionOutput{
					Version:       v,
					InstallMethod: installMethodStr,
				}
				// Try to fetch latest release; network failures are silently skipped.
				client := dlsec.NewClient()
				rel, err := version.FetchLatestRelease(context.Background(), client,
					"https://github.com/aiaikit/speckit/releases/latest")
				if err == nil {
					out.LatestRelease = rel.TagName
				}
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", " ")
				return enc.Encode(out)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "specify version %s\n", v)
			fmt.Fprintf(cmd.OutOrStdout(), "Installed via: %s\n", installMethodStr)

			client := dlsec.NewClient()
			rel, err := version.FetchLatestRelease(context.Background(), client,
				"https://github.com/aiaikit/speckit/releases/latest")
			if err != nil {
				// Network failure — skip upgrade notice silently.
				return nil
			}
			current := strings.TrimPrefix(v, "v")
			latest := strings.TrimPrefix(rel.TagName, "v")
			status := "up-to-date"
			if latest != current {
				status = "upgrade available"
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Latest release: %s (%s)\n", rel.TagName, status)
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonFlag, "json", false, "Output as JSON")
	return cmd
}
