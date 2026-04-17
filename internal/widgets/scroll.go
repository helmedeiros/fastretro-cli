package widgets

// ScrollWindow returns start/end indices for a visible window around the cursor.
// Keeps the cursor centered when possible, clamped to boundaries.
func ScrollWindow(total, cursor, maxVisible int) (int, int) {
	if total <= maxVisible {
		return 0, total
	}
	half := maxVisible / 2
	start := cursor - half
	if start < 0 {
		start = 0
	}
	end := start + maxVisible
	if end > total {
		end = total
		start = end - maxVisible
	}
	return start, end
}
