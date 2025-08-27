package tui

import (
	"fmt"
	"strings"

	"github.com/helmedeiros/fastretro-cli/internal/styles"
)

func (m Model) viewReview() string {
	if m.state == nil {
		return ""
	}

	var b strings.Builder

	b.WriteString(styles.Subtitle.Render("Action Items"))
	b.WriteString("\n\n")

	actions := m.actionNotes()
	if len(actions) == 0 {
		b.WriteString("  No action items recorded\n")
	}
	for i, a := range actions {
		owner := m.ownerForAction(a.id)
		ownerStr := ""
		if owner != "" {
			ownerStr = fmt.Sprintf(" → %s", owner)
		}
		b.WriteString(fmt.Sprintf("  %d. %s%s\n", i+1, a.text, ownerStr))
	}

	b.WriteString("\n")
	b.WriteString(styles.Subtitle.Render("Board Overview"))
	b.WriteString("\n\n")

	columns := m.getColumns()
	for _, col := range columns {
		cards := m.cardsForColumn(col.id)
		b.WriteString(fmt.Sprintf("  %s (%d cards)\n", col.title, len(cards)))
	}

	if len(m.state.Groups) > 0 {
		b.WriteString(fmt.Sprintf("\n  %d groups created\n", len(m.state.Groups)))
	}

	b.WriteString("\n")
	b.WriteString(styles.StatusBar.Render("View only — use web app to assign owners"))

	return b.String()
}

type actionEntry struct {
	id   string
	text string
}

func (m Model) actionNotes() []actionEntry {
	if m.state == nil {
		return nil
	}
	var result []actionEntry
	for _, n := range m.state.DiscussNotes {
		if n.Lane == "actions" {
			result = append(result, actionEntry{id: n.ID, text: n.Text})
		}
	}
	return result
}

func (m Model) ownerForAction(noteID string) string {
	if m.state == nil || m.state.ActionOwners == nil {
		return ""
	}
	return m.state.ActionOwners[noteID]
}
