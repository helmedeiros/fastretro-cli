package widgets

// WrapText breaks a string into lines of at most maxWidth characters,
// splitting at word boundaries when possible.
func WrapText(text string, maxWidth int) []string {
	if maxWidth <= 0 || len(text) == 0 {
		return []string{text}
	}
	if len(text) <= maxWidth {
		return []string{text}
	}

	var lines []string
	for len(text) > 0 {
		if len(text) <= maxWidth {
			lines = append(lines, text)
			break
		}
		// Find last space within maxWidth
		cut := maxWidth
		for cut > 0 && text[cut] != ' ' {
			cut--
		}
		if cut == 0 {
			// No space found — hard break
			cut = maxWidth
		}
		lines = append(lines, text[:cut])
		text = text[cut:]
		// Skip leading space on next line
		if len(text) > 0 && text[0] == ' ' {
			text = text[1:]
		}
	}
	return lines
}
