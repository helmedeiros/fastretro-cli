package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/helmedeiros/fastretro-cli/internal/protocol"
	"github.com/helmedeiros/fastretro-cli/internal/styles"
)

func (m Model) viewDiscuss() string {
	if m.state == nil || m.state.Discuss == nil {
		return styles.Subtitle.Render("Waiting for discussion to start...")
	}

	var b strings.Builder

	discuss := m.state.Discuss
	order := discuss.Order

	if discuss.CurrentIndex >= len(order) {
		b.WriteString(styles.Subtitle.Render("Discussion complete"))
		return b.String()
	}

	currentID := order[discuss.CurrentIndex]
	label := m.labelForItem(currentID)

	// Carousel bar: show all items, highlight current
	b.WriteString(styles.Subtitle.Render(
		fmt.Sprintf("Discussing %d of %d", discuss.CurrentIndex+1, len(order))))
	b.WriteString("\n\n")

	for i, id := range order {
		itemLabel := truncate(m.labelForItem(id), 20)
		if i == discuss.CurrentIndex {
			b.WriteString(styles.VoteBadge.Render(fmt.Sprintf(" %s ", itemLabel)))
		} else {
			b.WriteString(styles.Subtitle.Render(fmt.Sprintf(" %s ", itemLabel)))
		}
		b.WriteString(" ")
	}
	b.WriteString("\n\n")

	// Current item
	b.WriteString(styles.ActiveCard.Render(fmt.Sprintf("  %s  ", label)))
	b.WriteString("\n\n")

	// Segment tabs
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

	// Notes for current item + segment
	notes := m.notesForItem(currentID, segment)
	if len(notes) == 0 {
		b.WriteString(styles.Subtitle.Render("  No notes yet"))
	} else {
		for i, n := range notes {
			cursor := "  "
			if i == m.cursor {
				cursor = "> "
			}
			line := fmt.Sprintf("%s• %s", cursor, truncate(n.Text, 50))
			if i == m.cursor {
				b.WriteString(styles.Selected.Render(line))
			} else {
				b.WriteString(line)
			}
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")

	// Input mode
	if m.inputMode {
		b.WriteString(fmt.Sprintf("  Add note: %s▌\n", m.inputText))
		b.WriteString(styles.StatusBar.Render("[Enter] save  [Esc] cancel"))
	} else {
		// Timer
		if m.state.Timer != nil && m.state.Timer.Status == "running" {
			remaining := m.state.Timer.RemainingMs / 1000
			b.WriteString(styles.StatusBar.Render(
				fmt.Sprintf("Timer: %d:%02d", remaining/60, remaining%60)))
			b.WriteString("\n")
		}
		b.WriteString(styles.StatusBar.Render("[↑↓] notes  [←→] segment  [a] add note  [q] quit"))
	}

	return b.String()
}

func (m Model) handleDiscussKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.inputMode {
		return m.handleDiscussInput(msg)
	}

	if m.state == nil || m.state.Discuss == nil {
		return m, nil
	}

	discuss := m.state.Discuss
	currentID := ""
	if discuss.CurrentIndex < len(discuss.Order) {
		currentID = discuss.Order[discuss.CurrentIndex]
	}

	segment := discuss.Segment
	if segment == "" {
		segment = "context"
	}

	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		notes := m.notesForItem(currentID, segment)
		if m.cursor < len(notes)-1 {
			m.cursor++
		}
	case "left", "h":
		if segment == "actions" {
			m.state.Discuss.Segment = "context"
			m.cursor = 0
		}
	case "right", "l":
		if segment == "context" {
			m.state.Discuss.Segment = "actions"
			m.cursor = 0
		}
	case "a":
		m.inputMode = true
		m.inputText = ""
	}
	return m, nil
}

func (m Model) handleDiscussInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		text := strings.TrimSpace(m.inputText)
		if text != "" && m.state != nil && m.state.Discuss != nil {
			discuss := m.state.Discuss
			if discuss.CurrentIndex < len(discuss.Order) {
				segment := discuss.Segment
				if segment == "" {
					segment = "context"
				}
				note := protocol.DiscussNote{
					ID:           fmt.Sprintf("cli-%s-%d", m.participantID, len(m.state.DiscussNotes)),
					ParentCardID: discuss.Order[discuss.CurrentIndex],
					Lane:         segment,
					Text:         text,
				}
				m.state.DiscussNotes = append(m.state.DiscussNotes, note)
				m.broadcastState()
			}
		}
		m.inputMode = false
		m.inputText = ""
	case "esc":
		m.inputMode = false
		m.inputText = ""
	case "backspace":
		if len(m.inputText) > 0 {
			m.inputText = m.inputText[:len(m.inputText)-1]
		}
	default:
		if len(msg.String()) == 1 {
			m.inputText += msg.String()
		}
	}
	return m, nil
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
