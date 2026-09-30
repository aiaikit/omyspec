package preset

import "github.com/spf13/cobra"

func PresetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "preset",
		Short: "Manage presets",
	}
	cmd.AddCommand(PresetListCmd(), PresetInfoCmd(), PresetRemoveCmd())
	return cmd
}
