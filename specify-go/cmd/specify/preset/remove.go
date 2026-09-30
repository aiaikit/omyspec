package preset

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/aiaikit/speckit/internal/presetcatalog"
	"github.com/spf13/cobra"
)

func PresetRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove [id]",
		Short: "Remove a local preset",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			presetDir := filepath.Join(".specify", "presets", id)

			// Check if directory exists
			if _, err := os.Stat(presetDir); os.IsNotExist(err) {
				return fmt.Errorf("Preset not found: %s", id)
			}

			// Check if bundled
			p, err := presetcatalog.GetPreset(id)
			if err == nil && p.Bundled {
				return fmt.Errorf("Bundled presets cannot be removed")
			}

			// Remove the preset directory
			if err := os.RemoveAll(presetDir); err != nil {
				return fmt.Errorf("Failed to remove preset %s: %w", id, err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Removed preset: %s\n", id)
			return nil
		},
	}
}
