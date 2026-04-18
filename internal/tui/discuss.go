package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/helmedeiros/fastretro-cli/internal/protocol"
	"github.com/helmedeiros/fastretro-cli/internal/styles"
	"github.com/helmedeiros/fastretro-cli/internal/widgets"
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

	// Dot indicators (inline, no margins)
	activeDot := lipgloss.NewStyle().Foreground(styles.Accent).Render("●")
	inactiveDot := lipgloss.NewStyle().Foreground(styles.Muted).Render("○")
	b.WriteString("  ")
	for i := range order {
		if i > 0 {
			b.WriteString(" ")
		}
		if i == discuss.CurrentIndex {
			b.WriteString(activeDot)
		} else {
			b.WriteString(inactiveDot)
		}
	}
	b.WriteString("\n\n")

	isCheck := m.state.Meta.Type == "check"

	// Carousel: show items with vote counts or medians, current enlarged
	var carouselCards []string
	for i, id := range order {
		label := m.labelForItem(id)

		if i == discuss.CurrentIndex {
			var lines []string
			if isCheck {
				median := m.medianForItem(id)
				if median == 0 {
					lines = append(lines, styles.VoteBadge.Render("—"))
				} else {
					lines = append(lines, styles.VoteBadge.Render(fmt.Sprintf("%.1f", median)))
				}
			}
			lines = append(lines, styles.Selected.Render(label))
			if !isCheck {
				subcards := m.subcardsForItem(id)
				for _, sc := range subcards {
					lines = append(lines, styles.Subtitle.Render("  "+sc))
				}
			}
			if isCheck {
				// Show question description for current item
				tmpl := protocol.GetCheckTemplate(m.state.Meta.TemplateID)
				for _, q := range tmpl.Questions {
					if q.ID == id {
						lines = append(lines, styles.Subtitle.Render(q.Description))
						break
					}
				}
			} else {
				votes := m.votesForItem(id)
				lines = append(lines, styles.VoteBadge.Render(fmt.Sprintf("Votes: %d", votes)))
			}
			carouselCards = append(carouselCards, styles.ActiveCard.Render(strings.Join(lines, "\n")))
		} else {
			content := label
			if isCheck {
				median := m.medianForItem(id)
				if median > 0 {
					content += fmt.Sprintf("  %.1f", median)
				}
			} else {
				votes := m.votesForItem(id)
				if votes > 0 {
					content += fmt.Sprintf("  +%d", votes)
				}
			}
			carouselCards = append(carouselCards, styles.Card.Render(content))
		}
	}
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, carouselCards...))
	b.WriteString("\n\n")

	// Prev / Next bar
	muted := lipgloss.NewStyle().Foreground(styles.Muted)
	accent := lipgloss.NewStyle().Foreground(styles.Accent).Bold(true)
	prevLabel := "← Prev"
	nextLabel := "Next →"
	if discuss.CurrentIndex == 0 {
		prevLabel = muted.Render(prevLabel)
	} else {
		prevLabel = accent.Render(prevLabel)
	}
	if discuss.CurrentIndex >= len(order)-1 {
		nextLabel = muted.Render(nextLabel)
	} else {
		nextLabel = accent.Render(nextLabel)
	}
	b.WriteString(fmt.Sprintf("  %s       %s", prevLabel, nextLabel))
	b.WriteString("\n\n")

	// Context & Actions side by side
	segment := discuss.Segment
	if segment == "" {
		segment = "context"
	}

	contextNotes := m.notesForItem(currentID, "context")
	actionNotes := m.notesForItem(currentID, "actions")

	contextCol := m.renderNoteLane("CONTEXT", contextNotes, segment == "context")
	actionsCol := m.renderNoteLane("ACTIONS", actionNotes, segment == "actions")

	colStyle := styles.Column
	activeColStyle := colStyle.BorderForeground(styles.Accent)

	var laneContents []string
	var laneStyles []lipgloss.Style
	if segment == "context" {
		laneContents = []string{contextCol, actionsCol}
		laneStyles = []lipgloss.Style{activeColStyle, colStyle}
	} else {
		laneContents = []string{contextCol, actionsCol}
		laneStyles = []lipgloss.Style{colStyle, activeColStyle}
	}

	b.WriteString(widgets.JoinColumnsEqualHeight(laneContents, laneStyles))
	b.WriteString("\n")

	// Input mode
	if m.inputMode {
		laneLabel := segment
		b.WriteString(fmt.Sprintf("\n  Add %s: %s▌\n", laneLabel, m.inputText))
		b.WriteString(styles.StatusBar.Render("[Enter] save  [Esc] cancel"))
	} else {
		// Timer
		if m.state.Timer != nil && m.state.Timer.Status == "running" {
			remaining := m.state.Timer.RemainingMs / 1000
			b.WriteString(styles.StatusBar.Render(
				fmt.Sprintf("Timer: %d:%02d", remaining/60, remaining%60)))
			b.WriteString("\n")
		}
		b.WriteString(styles.StatusBar.Render("[j/k] notes  [h/l] lane  [n/p] item  [a] add  [q] back"))
	}

	return b.String()
}

func (m Model) renderNoteLane(title string, notes []noteEntry, active bool) string {
	var lines []string
	maxTextWidth := styles.Column.GetWidth() - 4 // padding + cursor prefix

	var header string
	if active {
		header = styles.Selected.Render(title)
	} else {
		header = styles.Subtitle.Render(title)
	}
	lines = append(lines, header)
	lines = append(lines, "")

	if len(notes) == 0 {
		lines = append(lines, styles.Subtitle.Render("(empty)"))
	} else {
		for i, n := range notes {
			cursor := "  "
			if active && i == m.cursor {
				cursor = "> "
			}
			wrapped := wrapText(n.Text, maxTextWidth)
			for j, wline := range wrapped {
				prefix := cursor
				if j > 0 {
					prefix = "  " // continuation lines get plain indent
				}
				line := prefix + wline
				if active && i == m.cursor {
					lines = append(lines, styles.Selected.Render(line))
				} else {
					lines = append(lines, line)
				}
			}
		}
	}

	return strings.Join(lines, "\n")
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
	case "tab":
		if segment == "context" {
			m.state.Discuss.Segment = "actions"
		} else {
			m.state.Discuss.Segment = "context"
		}
		m.cursor = 0
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
	case "n":
		if discuss.CurrentIndex < len(discuss.Order)-1 {
			m.state.Discuss.CurrentIndex++
			m.cursor = 0
			m.broadcastState()
		}
	case "p":
		if discuss.CurrentIndex > 0 {
			m.state.Discuss.CurrentIndex--
			m.cursor = 0
			m.broadcastState()
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
		if len(msg.Runes) > 0 {
			m.inputText += string(msg.Runes)
		}
	}
	return m, nil
}

func (m Model) labelForItem(itemID string) string {
	if m.state == nil {
		return itemID
	}
	// For checks, resolve from template questions
	if m.state.Meta.Type == "check" {
		tmpl := protocol.GetCheckTemplate(m.state.Meta.TemplateID)
		for _, q := range tmpl.Questions {
			if q.ID == itemID {
				return q.Title
			}
		}
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

func (m Model) medianForItem(itemID string) float64 {
	if m.state == nil {
		return 0
	}
	var ratings []int
	for _, r := range m.state.SurveyResponses {
		if r.QuestionID == itemID {
			ratings = append(ratings, r.Rating)
		}
	}
	return widgets.MedianInt(ratings)
}

func (m Model) subcardsForItem(itemID string) []string {
	if m.state == nil {
		return nil
	}
	for _, g := range m.state.Groups {
		if g.ID == itemID {
			var texts []string
			for _, cid := range g.CardIDs {
				if card, ok := m.cardByID(cid); ok {
					texts = append(texts, card.Text)
				}
			}
			return texts
		}
	}
	return nil
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

func wrapText(text string, maxWidth int) []string {
	return widgets.WrapText(text, maxWidth)
}
