package extension

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/aiaikit/speckit/internal/extcatalog"
	"github.com/spf13/cobra"
)

func ExtensionRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove [id]",
		Short: "Remove a local extension",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			extDir := filepath.Join(".specify", "extensions", id)

			// Check if directory exists
			if _, err := os.Stat(extDir); os.IsNotExist(err) {
				return fmt.Errorf("Extension not found: %s", id)
			}

			// Check if bundled (in catalog)
			ext, err := extcatalog.GetExtension(id)
			if err == nil && ext.Bundled {
				return fmt.Errorf("Bundled extensions cannot be removed")
			}

			// Remove the extension directory
			if err := os.RemoveAll(extDir); err != nil {
				return fmt.Errorf("Failed to remove extension %s: %w", id, err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Removed extension: %s\n", id)
			return nil
		},
	}
}
