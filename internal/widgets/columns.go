package widgets

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// JoinColumnsEqualHeight renders multiple styled columns at the same height,
// matching the tallest column, then joins them horizontally.
func JoinColumnsEqualHeight(contents []string, colStyles []lipgloss.Style) string {
	maxH := 0
	for i, content := range contents {
		r := colStyles[i].Render(content)
		h := strings.Count(r, "\n") + 1
		if h > maxH {
			maxH = h
		}
	}
	var rendered []string
	for i, content := range contents {
		r := colStyles[i].Height(maxH).Render(content)
		rendered = append(rendered, r)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, rendered...)
}
