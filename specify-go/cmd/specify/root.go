package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewRootCmd constructs the root cobra command for the specify CLI.
// Phase 1 wires only --version, version, and check. The 9 sub-groups
// (init, integration, extension, preset, workflow, event, bundle,
// artifact, self) are placeholder commands whose Execute returns
// "not yet implemented" — they will be filled in by Phase 2+.
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "specify",
		Short: "Setup tool for Specify spec-driven development projects",
		Long:  "specify — bootstrap and manage Spec Kit projects. Phase 1 ships only the skeleton.",
	}
	cmd.AddCommand(newVersionCmd(), newCheckCmd())
	cmd.AddCommand(placeholderCmd("init"))
	cmd.AddCommand(placeholderCmd("integration"))
	cmd.AddCommand(placeholderCmd("extension"))
	cmd.AddCommand(placeholderCmd("preset"))
	cmd.AddCommand(placeholderCmd("workflow"))
	cmd.AddCommand(placeholderCmd("event"))
	cmd.AddCommand(placeholderCmd("bundle"))
	cmd.AddCommand(placeholderCmd("artifact"))
	cmd.AddCommand(placeholderCmd("self"))
	return cmd
}

// newVersionCmd is added in Task 3; placeholder for now.
func newVersionCmd() *cobra.Command {
	return &cobra.Command{Use: "version", Short: "Print version"}
}

// newCheckCmd is added in Task 14; placeholder for now.
func newCheckCmd() *cobra.Command {
	return &cobra.Command{Use: "check", Short: "Check tool availability"}
}

func placeholderCmd(name string) *cobra.Command {
	return &cobra.Command{
		Use:   name,
		Short: name + " (not yet implemented in Phase 1)",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintf(cmd.ErrOrStderr(), "%s: not yet implemented in Phase 1\n", name)
			return fmt.Errorf("%s: not yet implemented", name)
		},
	}
}
