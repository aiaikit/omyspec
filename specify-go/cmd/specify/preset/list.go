package preset

import (
	"fmt"

	"github.com/aiaikit/speckit/internal/presetcatalog"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/spf13/cobra"
)

func PresetListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List available presets",
		RunE: func(cmd *cobra.Command, args []string) error {
			presets := presetcatalog.ListBundledPresets()

			headers := []string{"Name", "ID", "Version", "Description"}
			rows := make([][]string, 0, len(presets))
			for _, p := range presets {
				rows = append(rows, []string{p.Name, p.ID, p.Version, p.Description})
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
