package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/aiaikit/speckit/internal/dlsec"
	"github.com/aiaikit/speckit/internal/version"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

func SelfCheckCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check",
		Short: "Check for newer specify releases",
		RunE: func(cmd *cobra.Command, args []string) error {
			v := version.Version
			method := version.DetectInstallMethod()

			yellow := lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
			green := lipgloss.NewStyle().Foreground(lipgloss.Color("2"))

			fmt.Fprintf(cmd.OutOrStdout(), "Installed: %s\n", v)
			fmt.Fprintf(cmd.OutOrStdout(), "Via: %s\n", method.String())

			client := dlsec.NewClient()
			rel, err := version.FetchLatestRelease(context.Background(), client,
				"https://github.com/aiaikit/speckit/releases/latest")
			if err != nil {
				fmt.Fprintf(cmd.OutOrStdout(), "%s\n", yellow.Render("Could not check latest release"))
				return nil
			}

			current := strings.TrimPrefix(v, "v")
			latest := strings.TrimPrefix(rel.TagName, "v")
			if latest != current {
				fmt.Fprintf(cmd.OutOrStdout(), "Latest release: %s (%s)\n", rel.TagName, green.Render("Update available"))
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "Latest release: %s (%s)\n", rel.TagName, green.Render("Up to date"))
			}
			return nil
		},
	}
	return cmd
}
