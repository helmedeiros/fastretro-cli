package tui

import (
	"fmt"
	"strings"

	"github.com/helmedeiros/fastretro-cli/internal/styles"
)

func (m Model) viewDiscuss() string {
	if m.state == nil || m.state.Discuss == nil {
		return styles.Subtitle.Render("Waiting for discussion to start...")
	}

	var b strings.Builder

	discuss := m.state.Discuss
	order := discuss.Order
	idx := discuss.CurrentIndex

	if idx >= len(order) {
		b.WriteString(styles.Subtitle.Render("Discussion complete"))
		return b.String()
	}

	currentID := order[idx]
	label := m.labelForItem(currentID)

	b.WriteString(styles.Subtitle.Render(
		fmt.Sprintf("Discussing %d of %d", idx+1, len(order))))
	b.WriteString("\n\n")

	b.WriteString(styles.ActiveCard.Render(fmt.Sprintf("  %s  ", label)))
	b.WriteString("\n\n")

	// Segment indicator
	segment := discuss.Segment
	if segment == "" {
		segment = "context"
	}
	segments := []string{"context", "actions"}
	for _, s := range segments {
		if s == segment {
			b.WriteString(styles.Selected.Render(fmt.Sprintf(" [%s] ", s)))
		} else {
			b.WriteString(fmt.Sprintf("  %s  ", s))
		}
	}
	b.WriteString("\n\n")

	// Notes for current item
	notes := m.notesForItem(currentID, segment)
	if len(notes) == 0 {
		b.WriteString(styles.Subtitle.Render("  No notes yet"))
	}
	for _, n := range notes {
		b.WriteString(fmt.Sprintf("  • %s\n", n.Text))
	}

	b.WriteString("\n")

	// Timer
	if m.state.Timer != nil && m.state.Timer.Status == "running" {
		remaining := m.state.Timer.RemainingMs / 1000
		b.WriteString(styles.StatusBar.Render(
			fmt.Sprintf("Timer: %d:%02d", remaining/60, remaining%60)))
		b.WriteString("\n")
	}

	b.WriteString(styles.StatusBar.Render("View only — use web app to add notes"))

	return b.String()
}

func (m Model) labelForItem(itemID string) string {
	if m.state == nil {
		return itemID
	}
	for _, g := range m.state.Groups {
		if g.ID == itemID {
			return g.Name
		}
	}
	for _, c := range m.state.Cards {
		if c.ID == itemID {
			return c.Text
		}
	}
	return itemID
}

func (m Model) notesForItem(itemID, lane string) []noteEntry {
	if m.state == nil {
		return nil
	}
	var result []noteEntry
	for _, n := range m.state.DiscussNotes {
		if n.ParentCardID == itemID && n.Lane == lane {
			result = append(result, noteEntry{Text: n.Text})
		}
	}
	return result
}

type noteEntry struct {
	Text string
}
