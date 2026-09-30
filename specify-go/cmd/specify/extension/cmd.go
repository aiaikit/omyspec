package extension

import "github.com/spf13/cobra"

func ExtensionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "extension",
		Short: "Manage extensions",
	}
	cmd.AddCommand(ExtensionListCmd(), ExtensionInfoCmd())
	return cmd
}
