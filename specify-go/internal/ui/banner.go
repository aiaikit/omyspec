package ui

import "github.com/charmbracelet/lipgloss"

// Show returns the rendered banner string. Phase 1 uses ASCII only —
// no font embedding. The style is a simple cyan border via lipgloss.
func Show() string {
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("6")).
		Padding(0, 2).
		Width(50)
	return box.Render("Spec Kit — Specify CLI bootstrap tool")
}