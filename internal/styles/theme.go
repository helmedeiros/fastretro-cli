package styles

import "github.com/charmbracelet/lipgloss"

var (
	// Colors matching fastRetro dark theme
	Accent    = lipgloss.Color("#5ec4c8")
	Muted     = lipgloss.Color("#8899aa")
	Surface   = lipgloss.Color("#1e2a32")
	Border    = lipgloss.Color("#2a3a44")
	Danger    = lipgloss.Color("#e06060")
	Success   = lipgloss.Color("#6ec76e")

	// Styles
	Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(Accent).
		MarginBottom(1)

	Subtitle = lipgloss.NewStyle().
		Foreground(Muted).
		MarginBottom(1)

	Card = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Border).
		Padding(0, 1)

	ActiveCard = Card.
		BorderForeground(Accent)

	StatusBar = lipgloss.NewStyle().
		Foreground(Muted).
		MarginTop(1)

	Selected = lipgloss.NewStyle().
		Foreground(Accent).
		Bold(true)

	Taken = lipgloss.NewStyle().
		Foreground(Muted).
		Strikethrough(true)

	Column = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Border).
		Padding(1).
		Width(52)

	VoteBadge = lipgloss.NewStyle().
		Background(Accent).
		Foreground(lipgloss.Color("#1a1a2e")).
		Padding(0, 1).
		Bold(true)

	HistoryColumn = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Border).
		Padding(1).
		Width(60)
)
