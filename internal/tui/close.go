package tui

import (
	"fmt"
	"strings"

	"github.com/helmedeiros/fastretro-cli/internal/styles"
)

func (m Model) viewClose() string {
	if m.state == nil {
		return ""
	}

	var b strings.Builder

	b.WriteString(styles.Subtitle.Render("Retrospective Summary"))
	b.WriteString("\n\n")

	// Meta info
	meta := m.state.Meta
	if meta.Name != "" {
		b.WriteString(fmt.Sprintf("  Name: %s\n", meta.Name))
	}
	if meta.Date != "" {
		b.WriteString(fmt.Sprintf("  Date: %s\n", meta.Date))
	}
	b.WriteString("\n")

	// Stats
	b.WriteString(styles.Subtitle.Render("Stats"))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("  Participants: %d\n", len(m.state.Participants)))
	b.WriteString(fmt.Sprintf("  Cards: %d\n", len(m.state.Cards)))
	b.WriteString(fmt.Sprintf("  Groups: %d\n", len(m.state.Groups)))
	b.WriteString(fmt.Sprintf("  Votes cast: %d\n", len(m.state.Votes)))

	actions := m.actionNotes()
	b.WriteString(fmt.Sprintf("  Action items: %d\n", len(actions)))

	// Action items
	if len(actions) > 0 {
		b.WriteString("\n")
		b.WriteString(styles.Subtitle.Render("Action Items"))
		b.WriteString("\n")
		for i, a := range actions {
			owner := m.ownerForAction(a.id)
			ownerStr := ""
			if owner != "" {
				ownerStr = fmt.Sprintf(" → %s", owner)
			}
			b.WriteString(fmt.Sprintf("  %d. %s%s\n", i+1, a.text, ownerStr))
		}
	}

	// Board overview
	b.WriteString("\n")
	b.WriteString(styles.Subtitle.Render("Board"))
	b.WriteString("\n")
	columns := m.getColumns()
	for _, col := range columns {
		cards := m.cardsForColumn(col.id)
		b.WriteString(fmt.Sprintf("  %s: %d cards\n", col.title, len(cards)))
	}

	b.WriteString("\n")
	b.WriteString(styles.StatusBar.Render("Retro complete! [q] quit"))

	return b.String()
}
