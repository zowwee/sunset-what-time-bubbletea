package main

import "github.com/charmbracelet/lipgloss"

var (
	// Soft peach for titles
	TitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("217"))

	// Pastel pink for selection
	SelectedStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("218"))

	// Soft lavender-gray for help text
	HelpStyle = lipgloss.NewStyle().
		Faint(true).
		Foreground(lipgloss.Color("249"))

	// Muted dusty rose for errors (still readable, not alarming)
	ErrorStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("210"))

	// Pastel peach border
	BoxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("223")).
		Padding(1, 2)
)
