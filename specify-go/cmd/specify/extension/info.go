package extension

import (
	"fmt"
	"strings"

	"github.com/aiaikit/speckit/internal/extcatalog"
	"github.com/spf13/cobra"
)

func ExtensionInfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info [id]",
		Short: "Show details for an extension",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			ext, err := extcatalog.GetExtension(id)
			if err != nil {
				if err == extcatalog.ErrNotFound {
					return fmt.Errorf("Extension not found: %s", id)
				}
				return err
			}
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Name: %s\n", ext.Name)
			fmt.Fprintf(out, "ID: %s\n", ext.ID)
			fmt.Fprintf(out, "Version: %s\n", ext.Version)
			fmt.Fprintf(out, "Description: %s\n", ext.Description)
			fmt.Fprintf(out, "Author: %s\n", ext.Author)
			fmt.Fprintf(out, "Tags: %s\n", strings.Join(ext.Tags, ", "))
			fmt.Fprintf(out, "Repository: %s\n", ext.Repository)
			fmt.Fprintf(out, "Bundled: %t\n", ext.Bundled)
			return nil
		},
	}
}
