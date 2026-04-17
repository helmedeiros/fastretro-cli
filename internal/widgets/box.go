// Package widgets provides reusable TUI components.
package widgets

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// BoxConfig holds the styling options for a titled box.
type BoxConfig struct {
	ActiveColor   lipgloss.Color
	InactiveColor lipgloss.Color
}

// DefaultBoxConfig returns the default box styling.
func DefaultBoxConfig(activeColor, inactiveColor lipgloss.Color) BoxConfig {
	return BoxConfig{
		ActiveColor:   activeColor,
		InactiveColor: inactiveColor,
	}
}

// TitledBox renders content inside a bordered box with a title in the top
// border and an optional label positioned at the last 1/4 of the bottom border.
//
//	╭─ TITLE ──────────────────────────╮
//	│ content                          │
//	╰──────────────────── 4 total ─────╯
func TitledBox(cfg BoxConfig, title, content, bottomLabel string, width, minHeight int, active bool) string {
	borderColor := cfg.InactiveColor
	if active {
		borderColor = cfg.ActiveColor
	}
	bc := lipgloss.NewStyle().Foreground(borderColor)

	innerWidth := width - 2 // minus left+right border chars

	// Top border with title
	titleStr := " " + title + " "
	titleLen := lipgloss.Width(titleStr)
	rightDash := innerWidth - 1 - titleLen
	if rightDash < 0 {
		rightDash = 0
	}
	top := bc.Render("╭─") + titleStr + bc.Render(strings.Repeat("─", rightDash)+"╮")

	// Content lines with side borders, padded to minHeight
	contentLines := strings.Split(content, "\n")
	for len(contentLines) < minHeight {
		contentLines = append(contentLines, "")
	}
	var body strings.Builder
	for _, line := range contentLines {
		lineWidth := lipgloss.Width(line)
		pad := innerWidth - lineWidth
		if pad < 0 {
			pad = 0
		}
		body.WriteString(bc.Render("│") + " " + line + strings.Repeat(" ", pad-1) + bc.Render("│") + "\n")
	}

	// Bottom border with optional label at the last 1/4
	var bottom string
	if bottomLabel != "" {
		labelStr := " " + bottomLabel + " "
		labelLen := lipgloss.Width(labelStr)
		leftDashes := innerWidth*3/4 - 1
		if leftDashes < 1 {
			leftDashes = 1
		}
		rightDashes := innerWidth - leftDashes - labelLen
		if rightDashes < 0 {
			rightDashes = 0
		}
		bottom = bc.Render("╰"+strings.Repeat("─", leftDashes)) + labelStr + bc.Render(strings.Repeat("─", rightDashes)+"╯")
	} else {
		bottom = bc.Render("╰" + strings.Repeat("─", innerWidth) + "╯")
	}

	return top + "\n" + body.String() + bottom
}

// ContentHeight counts the number of lines in a content string.
func ContentHeight(content string) int {
	return strings.Count(content, "\n") + 1
}
