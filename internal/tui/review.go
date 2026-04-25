package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/helmedeiros/fastretro-cli/internal/styles"
	"github.com/helmedeiros/fastretro-cli/internal/widgets"
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

	// --- Participant picker ---
	if m.reviewPickMode {
		b.WriteString(accent.Render("  Select owner:"))
		b.WriteString("\n")
		for i, p := range m.state.Participants {
			cursor := "    "
			if i == m.reviewPickCursor {
				cursor = "  > "
			}
			line := cursor + p.Name
			if i == m.reviewPickCursor {
				b.WriteString(styles.Selected.Render(line))
			} else {
				b.WriteString(line)
			}
			b.WriteString("\n")
		}
		// Option to type a new name
		newCursor := "    "
		if m.reviewPickCursor == len(m.state.Participants) {
			newCursor = "  > "
		}
		newLine := newCursor + "(type new name...)"
		if m.reviewPickCursor == len(m.state.Participants) {
			b.WriteString(styles.Selected.Render(newLine))
		} else {
			b.WriteString(muted.Render(newLine))
		}
		b.WriteString("\n")
		b.WriteString(muted.Render("  [j/k] select  [Enter] confirm  [Esc] cancel"))
		b.WriteString("\n")
	}

	// --- Manual name input ---
	if m.inputMode {
		b.WriteString(fmt.Sprintf("  Owner name: %s▌\n", m.inputText))
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
			b.WriteString(widgets.JoinColumnsEqualHeight(boardContents, boardStyles))
			b.WriteString("\n")
		}
	}

	// Help
	if !m.inputMode && !m.reviewPickMode {
		b.WriteString("\n")
		b.WriteString(muted.Render("[j/k] navigate  [a] assign owner  [q] back"))
	}

	return b.String()
}

func (m Model) handleReviewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.inputMode {
		return m.handleReviewInput(msg)
	}
	if m.reviewPickMode {
		return m.handleReviewPick(msg)
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
	case "a":
		if m.cursor < len(actions) && m.state != nil {
			m.reviewPickMode = true
			m.reviewPickCursor = 0
		}
	}
	return m, nil
}

func (m Model) handleReviewPick(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.state == nil {
		return m, nil
	}
	maxIdx := len(m.state.Participants) // last entry is "type new name"

	switch msg.String() {
	case "up", "k":
		if m.reviewPickCursor > 0 {
			m.reviewPickCursor--
		}
	case "down", "j":
		if m.reviewPickCursor < maxIdx {
			m.reviewPickCursor++
		}
	case "enter":
		if m.reviewPickCursor < len(m.state.Participants) {
			// Selected an existing participant
			p := m.state.Participants[m.reviewPickCursor]
			actions := m.actionNotes()
			if m.cursor < len(actions) {
				if m.state.ActionOwners == nil {
					m.state.ActionOwners = make(map[string]string)
				}
				m.state.ActionOwners[actions[m.cursor].id] = p.ID
				m.broadcastState()
			}
			m.reviewPickMode = false
		} else {
			// "Type new name" selected — switch to text input
			m.reviewPickMode = false
			m.inputMode = true
			m.inputText = ""
		}
	case "esc":
		m.reviewPickMode = false
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
