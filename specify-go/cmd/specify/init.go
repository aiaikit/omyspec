package main

import (
	"fmt"

	"github.com/aiaikit/speckit/internal/integration"
	"github.com/aiaikit/speckit/internal/projectstate"
	"github.com/aiaikit/speckit/internal/ui"
	"github.com/spf13/cobra"
)

// InitCmd returns the `specify init` command.
func InitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a Spec Kit project",
		Long:  "Initialize a new Spec Kit project in the current directory or a named path.",
		Args:  cobra.RangeArgs(0, 1),
		RunE:  runInit,
	}
	cmd.Flags().Bool("here", false, "Initialize in the current directory")
	cmd.Flags().Bool("force", false, "Force init even if already initialized")
	cmd.Flags().Bool("non-interactive", false, "Run without prompting")
	cmd.Flags().Bool("script", false, "Output the init script instead of running it")
	cmd.Flags().String("integration", "", "Integration key to configure (claude, copilot, codex, generic)")
	cmd.Flags().String("integration-options", "", "JSON options for the integration")
	cmd.Flags().String("preset", "", "Preset name to apply")
	cmd.Flags().StringSlice("extension", nil, "Extension names to install")
	cmd.Flags().Bool("ignore-agent-tools", false, "Skip agent tool checks")
	return cmd
}

func runInit(cmd *cobra.Command, args []string) error {
	flags := parseInitFlags(cmd, args)

	_, err := projectstate.ResolveProjectRoot()
	if err == nil && !flags.force {
		return fmt.Errorf("project already initialized (use --force to re-init)")
	}

	banner := ui.Show()
	fmt.Fprintln(cmd.OutOrStdout(), banner)

	tr := ui.NewTracker("Init Steps")
	tr.Add("project", "Initialize project")
	tr.Mark("project", ui.Done, "done")

	fmt.Fprintln(cmd.OutOrStdout(), tr.Render())

	if flags.integration != "" {
		_, err := integration.GetIntegration(flags.integration)
		if err != nil {
			return fmt.Errorf("unknown integration key %q: %w", flags.integration, err)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "integration: %s\n", flags.integration)
	}

	if len(flags.extensions) > 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "extensions: %v\n", flags.extensions)
	}

	if flags.script {
		fmt.Fprintln(cmd.OutOrStdout(), "(script output not yet implemented)")
	}

	_ = flags // ponytail: use parsed flags in follow-up tasks
	return nil
}

type initFlags struct {
	here              bool
	force             bool
	nonInteractive    bool
	script            bool
	integration       string
	integrationOpts   string
	preset            string
	extensions        []string
	ignoreAgentTools  bool
}

func parseInitFlags(cmd *cobra.Command, args []string) initFlags {
	flags := initFlags{}
	flags.here, _ = cmd.Flags().GetBool("here")
	flags.force, _ = cmd.Flags().GetBool("force")
	flags.nonInteractive, _ = cmd.Flags().GetBool("non-interactive")
	flags.script, _ = cmd.Flags().GetBool("script")
	flags.integration, _ = cmd.Flags().GetString("integration")
	flags.integrationOpts, _ = cmd.Flags().GetString("integration-options")
	flags.preset, _ = cmd.Flags().GetString("preset")
	flags.extensions, _ = cmd.Flags().GetStringSlice("extension")
	flags.ignoreAgentTools, _ = cmd.Flags().GetBool("ignore-agent-tools")
	return flags
}
