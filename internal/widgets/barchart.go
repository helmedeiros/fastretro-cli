package widgets

import (
	"fmt"
	"strings"
)

// BarChart renders a horizontal bar chart block with labels, filled bars, and values.
func BarChart(labels []string, values []float64, maxValue, barWidth int) string {
	if len(labels) != len(values) || len(labels) == 0 {
		return ""
	}

	// Find longest label for alignment
	maxLabelLen := 0
	for _, l := range labels {
		if len(l) > maxLabelLen {
			maxLabelLen = len(l)
		}
	}

	var b strings.Builder
	for i, label := range labels {
		padded := fmt.Sprintf("%-*s", maxLabelLen, label)
		filled := 0
		if maxValue > 0 && values[i] > 0 {
			filled = int(values[i] / float64(maxValue) * float64(barWidth))
			if filled > barWidth {
				filled = barWidth
			}
		}
		empty := barWidth - filled
		bar := strings.Repeat("█", filled) + strings.Repeat("░", empty)
		score := "  —"
		if values[i] > 0 {
			score = fmt.Sprintf("%.1f", values[i])
		}
		b.WriteString(fmt.Sprintf("  %s  %s  %s\n", padded, bar, score))
	}
	return b.String()
}
