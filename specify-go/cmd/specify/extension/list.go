package extension

import (
	"fmt"

	"github.com/aiaikit/speckit/internal/extcatalog"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/spf13/cobra"
)

func ExtensionListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List installed extensions",
		RunE: func(cmd *cobra.Command, args []string) error {
			bundled := extcatalog.ListBundledExtensions()
			local := extcatalog.ListLocalExtensions(".") // ponytail: use project root from flag

			headers := []string{"Name", "ID", "Version", "Source"}
			rows := make([][]string, 0, len(bundled)+len(local))
			for _, ext := range bundled {
				rows = append(rows, []string{ext.Name, ext.ID, ext.Version, "bundled"})
			}
			for _, ext := range local {
				rows = append(rows, []string{ext.Name, ext.ID, ext.Version, "local"})
			}

			border := lipgloss.Color("6")
			t := table.New().
				Border(lipgloss.RoundedBorder()).
				BorderStyle(lipgloss.NewStyle().Foreground(border)).
				Headers(headers...).
				Rows(rows...)

			fmt.Fprintln(cmd.OutOrStdout(), t)
			return nil
		},
	}
}
