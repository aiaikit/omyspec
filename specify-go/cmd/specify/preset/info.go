package preset

import (
	"fmt"

	"github.com/aiaikit/speckit/internal/presetcatalog"
	"github.com/spf13/cobra"
)

func PresetInfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info [id]",
		Short: "Show details for a preset",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			p, err := presetcatalog.GetPreset(id)
			if err != nil {
				if err == presetcatalog.ErrNotFound {
					return fmt.Errorf("Preset not found: %s", id)
				}
				return err
			}
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Name: %s\n", p.Name)
			fmt.Fprintf(out, "ID: %s\n", p.ID)
			fmt.Fprintf(out, "Version: %s\n", p.Version)
			fmt.Fprintf(out, "Description: %s\n", p.Description)
			fmt.Fprintf(out, "Bundled: %t\n", p.Bundled)
			return nil
		},
	}
}
