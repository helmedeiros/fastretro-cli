package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/helmedeiros/fastretro-cli/internal/styles"
)

func (m Model) viewReview() string {
	if m.state == nil {
		return ""
	}

	var b strings.Builder

	// --- Actions from this retrospective ---
	accent := lipgloss.NewStyle().Foreground(styles.Accent).Bold(true)
	muted := lipgloss.NewStyle().Foreground(styles.Muted)

	sessionLabel := "retrospective"
	if m.state.Meta.Type == "check" {
		sessionLabel = "check"
	}
	b.WriteString(accent.Render("Actions from this " + sessionLabel))
	b.WriteString("\n")
	b.WriteString(muted.Render("────────────────────────────────────────"))
	b.WriteString("\n\n")

	actions := m.actionNotes()
	if len(actions) == 0 {
		b.WriteString(muted.Render("  No action items recorded"))
		b.WriteString("\n")
	} else {
		for i, a := range actions {
			cursor := "  "
			if i == m.cursor {
				cursor = "> "
			}

			owner := m.ownerForAction(a.id)
			ownerStr := muted.Render("  unassigned")
			if owner != "" {
				ownerStr = muted.Render("  " + m.participantName(owner))
			}

			checkmark := accent.Render("✓")
			line := fmt.Sprintf("%s%s %s", cursor, checkmark, a.text)

			if i == m.cursor {
				b.WriteString(styles.Selected.Render(line))
				b.WriteString("\n")
				b.WriteString("    " + ownerStr)
			} else {
				b.WriteString(line)
				b.WriteString("\n")
				b.WriteString("    " + ownerStr)
			}
			b.WriteString("\n\n")
		}
	}

	// --- Owner assignment input ---
	if m.inputMode {
		b.WriteString(fmt.Sprintf("  Assign owner: %s▌\n", m.inputText))
		b.WriteString(muted.Render("  Participants: "))
		var names []string
		for _, p := range m.state.Participants {
			names = append(names, p.Name)
		}
		b.WriteString(muted.Render(strings.Join(names, ", ")))
		b.WriteString("\n")
		b.WriteString(muted.Render("  [Enter] save  [Esc] cancel"))
		b.WriteString("\n")
	}

	// --- Board overview (retro only) ---
	if m.state.Meta.Type != "check" {
		b.WriteString("\n")
		b.WriteString(accent.Render("Board overview"))
		b.WriteString("\n")
		b.WriteString(muted.Render("────────────────────────────────────────"))
		b.WriteString("\n\n")

		columns := m.getColumns()
		grouped := m.groupedCardIDs()

		var boardContents []string
		var boardStyles []lipgloss.Style
		for _, col := range columns {
			var lines []string
			lines = append(lines, accent.Render(col.title))
			lines = append(lines, "")

			for _, g := range m.groupsForColumn(col.id) {
				lines = append(lines, accent.Render(g.Name))
				for _, cid := range g.CardIDs {
					if card, ok := m.cardByID(cid); ok {
						lines = append(lines, "  "+muted.Render(card.Text))
					}
				}
				lines = append(lines, "")
			}

			for _, c := range m.cardsForColumn(col.id) {
				if !grouped[c.ID] {
					lines = append(lines, muted.Render(c.Text))
				}
			}

			boardContents = append(boardContents, strings.Join(lines, "\n"))
			boardStyles = append(boardStyles, styles.Column)
		}

		if len(boardContents) > 0 {
			b.WriteString(joinColumnsEqualHeight(boardContents, boardStyles))
			b.WriteString("\n")
		}
	}

	// Help
	if !m.inputMode {
		b.WriteString("\n")
		b.WriteString(muted.Render("[↑↓] navigate  [o] assign owner  [q] quit"))
	}

	return b.String()
}

func (m Model) handleReviewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.inputMode {
		return m.handleReviewInput(msg)
	}

	actions := m.actionNotes()

	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(actions)-1 {
			m.cursor++
		}
	case "o":
		if m.cursor < len(actions) {
			m.inputMode = true
			m.inputText = ""
		}
	}
	return m, nil
}

func (m Model) handleReviewInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		name := strings.TrimSpace(m.inputText)
		if name != "" && m.state != nil {
			actions := m.actionNotes()
			if m.cursor < len(actions) {
				if m.state.ActionOwners == nil {
					m.state.ActionOwners = make(map[string]string)
				}
				// Match to participant ID if possible
				ownerID := name
				for _, p := range m.state.Participants {
					if strings.EqualFold(p.Name, name) {
						ownerID = p.ID
						break
					}
				}
				m.state.ActionOwners[actions[m.cursor].id] = ownerID
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
		if len(msg.Runes) > 0 {
			m.inputText += string(msg.Runes)
		}
	}
	return m, nil
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
