package main

import (
	"fmt"

	"github.com/aiaikit/speckit/internal/integration"
	"github.com/aiaikit/speckit/internal/projectstate"
	"github.com/aiaikit/speckit/internal/ui"
	"github.com/spf13/cobra"
)

func newCheckCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "check",
		Short: "Check that all required tools are installed",
		RunE: func(cmd *cobra.Command, args []string) error {
			projectRoot, err := projectstate.ResolveProjectRoot()
			if err != nil {
				if err == projectstate.ErrNotAProject {
					fmt.Fprintln(cmd.OutOrStdout(), "Run 'specify init' first")
					return nil
				}
				return err
			}

			c := ui.NewConsole(cmd.OutOrStdout(), cmd.ErrOrStderr())
			c.Print(ui.Show())

			tr := ui.NewTracker("Check Available Tools")
			if err := integration.CheckTools(projectRoot, tr); err != nil {
				return err
			}

			c.Print(tr.Render())
			return nil
		},
	}
}
