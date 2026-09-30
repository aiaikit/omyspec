package main

import (
	"fmt"
	"strings"

	"github.com/aiaikit/speckit/internal/executil"
	"github.com/aiaikit/speckit/internal/version"
	"github.com/spf13/cobra"
)

func SelfUpgradeCmd() *cobra.Command {
	var dryRun bool
	var tag string

	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade the specify CLI",
		RunE: func(cmd *cobra.Command, args []string) error {
			method := version.DetectInstallMethod()
			var argv []string

			switch method {
			case version.InstallUVTool:
				argv = []string{"uv", "tool", "upgrade", "specify"}
			case version.InstallPipx:
				argv = []string{"pipx", "upgrade", "specify-cli"}
			case version.InstallNPM:
				pkg := "specify-cli"
				if tag != "" {
					pkg = "specify-cli@" + tag // e.g. specify-cli@v1.2.3
				}
				argv = []string{"npm", "install", "-g", pkg}
			case version.InstallUVXEphemeral:
				if dryRun {
					fmt.Fprintln(cmd.OutOrStdout(), "[dry run] would print: Upgrade uvx-installed specify via: uvx run specify")
					return nil
				}
				fmt.Fprintln(cmd.OutOrStdout(), "Upgrade uvx-installed specify via: uvx run specify")
				return nil
			case version.InstallSourceCheckout:
				if dryRun {
					fmt.Fprintln(cmd.OutOrStdout(), "[dry run] would print: Upgrade source checkout via: git pull origin main")
					return nil
				}
				fmt.Fprintln(cmd.OutOrStdout(), "Upgrade source checkout via: git pull origin main")
				return nil
			default:
				if dryRun {
					fmt.Fprintln(cmd.OutOrStdout(), "[dry run] would print: Could not detect install method. Upgrade manually.")
					return nil
				}
				fmt.Fprintln(cmd.OutOrStdout(), "Could not detect install method. Upgrade manually.")
				return nil
			}

			if tag != "" {
				if method == version.InstallUVTool {
					fmt.Fprintln(cmd.OutOrStdout(), "Note: --tag is ignored for uv-tool installs (uv manages versions automatically)")
				} else if method == version.InstallPipx {
					fmt.Fprintln(cmd.OutOrStdout(), "Note: --tag is ignored for pipx installs (pipx manages versions automatically)")
				}
			}

			if dryRun {
				fmt.Fprintf(cmd.OutOrStdout(), "[dry run] would run: %s\n", strings.Join(argv, " "))
				return nil
			}

			result, err := executil.Run(argv[0], argv[1:], executil.RunOpts{
				Stdout: cmd.OutOrStdout(),
				Stderr: cmd.OutOrStderr(),
			})
			if err != nil {
				return err
			}
			if result.ExitCode != 0 {
				return fmt.Errorf("upgrade failed with exit code %d", result.ExitCode)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview upgrade command without running it")
	cmd.Flags().StringVar(&tag, "tag", "", "Pin target version (e.g. v1.2.3)")
	return cmd
}
