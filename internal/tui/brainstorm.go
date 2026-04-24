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

// brainstormItem is a flat entry for cursor navigation within a column.
type brainstormItem struct {
	kind   string // "card", "group-header", "group-card"
	cardID string
	mine   bool
}

// columnBrainstormItems builds a navigable list for one column.
func (m Model) columnBrainstormItems(colID string) []brainstormItem {
	if m.state == nil {
		return nil
	}
	var items []brainstormItem
	grouped := m.groupedCardIDs()

	for _, g := range m.groupsForColumn(colID) {
		items = append(items, brainstormItem{kind: "group-header"})
		for _, cid := range g.CardIDs {
			items = append(items, brainstormItem{
				kind:   "group-card",
				cardID: cid,
				mine:   m.isMyCard(cid),
			})
		}
	}

	for _, c := range m.cardsForColumn(colID) {
		if !grouped[c.ID] {
			items = append(items, brainstormItem{
				kind:   "card",
				cardID: c.ID,
				mine:   m.isMyCard(c.ID),
			})
		}
	}
	return items
}

func (m Model) isMyCard(cardID string) bool {
	return strings.HasPrefix(cardID, "cli-"+m.participantID)
}

func (m Model) viewBrainstorm() string {
	if m.state == nil {
		return ""
	}

	muted := lipgloss.NewStyle().Foreground(styles.Muted)
	columns := m.getColumns()
	if len(columns) == 0 {
		return muted.Render("No columns defined")
	}

	var contents []string
	var colStyles []lipgloss.Style
	for ci, col := range columns {
		items := m.columnBrainstormItems(col.id)
		isActive := ci == m.activeCol

		var lines []string
		gi := -1 // track which group we're in for headers
		for idx, item := range items {
			cursor := "  "
			if isActive && idx == m.cursor {
				cursor = "> "
			}

			switch item.kind {
			case "group-header":
				gi++
				groups := m.groupsForColumn(col.id)
				if gi < len(groups) {
					label := fmt.Sprintf("┌ %s", groups[gi].Name)
					if isActive && idx == m.cursor {
						lines = append(lines, styles.Selected.Render(cursor+label))
					} else {
						lines = append(lines, styles.Selected.Render("  "+label))
					}
				}
			case "group-card":
				card, ok := m.cardByID(item.cardID)
				if !ok {
					continue
				}
				text := card.Text
				line := fmt.Sprintf("%s│  %s", cursor, text)
				if isActive && idx == m.cursor {
					lines = append(lines, styles.Selected.Render(line))
				} else {
					lines = append(lines, fmt.Sprintf("  │  %s", text))
				}
			case "card":
				card, ok := m.cardByID(item.cardID)
				if !ok {
					continue
				}
				text := card.Text
				line := fmt.Sprintf("%s• %s", cursor, text)
				if isActive && idx == m.cursor {
					lines = append(lines, styles.Selected.Render(line))
				} else {
					lines = append(lines, fmt.Sprintf("  • %s", text))
				}
			}
		}

		header := col.title
		if isActive {
			header = styles.Selected.Render("▶ " + header)
		}
		if col.description != "" {
			header += "\n" + muted.Render(col.description) + "\n"
		}

		body := strings.Join(lines, "\n")
		if len(lines) == 0 {
			body = muted.Render("  (empty)")
		}

		if m.inputMode && isActive {
			body += fmt.Sprintf("\n\n  Add: %s▌", m.inputText)
		}

		content := header + "\n" + body
		contents = append(contents, content)
		style := styles.Column
		if isActive {
			style = style.BorderForeground(styles.Accent)
		}
		colStyles = append(colStyles, style)
	}

	// Show sliding window of columns centered on active
	maxVisibleCols := 3
	colStart, colEnd := widgets.ScrollWindow(len(contents), m.activeCol, maxVisibleCols)
	board := widgets.JoinColumnsEqualHeight(contents[colStart:colEnd], colStyles[colStart:colEnd])

	// Column position indicator
	if len(columns) > maxVisibleCols {
		indicator := muted.Render(fmt.Sprintf("  column %d of %d", m.activeCol+1, len(columns)))
		board += "\n" + indicator
	}

	help := "[j/k] navigate  [h/l] column  [a] add  [d] delete  [q] back"
	if m.inputMode {
		help = "[Enter] submit  [Esc] cancel"
	}

	return board + "\n\n" + muted.Render(help)
}

func (m Model) handleBrainstormKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.inputMode {
		return m.handleBrainstormInput(msg)
	}

	columns := m.getColumns()
	if len(columns) == 0 {
		return m, nil
	}

	items := m.columnBrainstormItems(columns[m.activeCol].id)

	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(items)-1 {
			m.cursor++
		}
	case "tab", "right", "l":
		if len(columns) > 0 {
			m.activeCol = (m.activeCol + 1) % len(columns)
			m.cursor = 0
		}
	case "shift+tab", "left", "h":
		if len(columns) > 0 {
			m.activeCol = (m.activeCol - 1 + len(columns)) % len(columns)
			m.cursor = 0
		}
	case "a":
		m.inputMode = true
		m.inputText = ""
	case "d":
		if m.cursor < len(items) {
			item := items[m.cursor]
			if (item.kind == "card" || item.kind == "group-card") && item.mine {
				m.removeCard(item.cardID)
			}
		}
	}
	return m, nil
}

func (m *Model) removeCard(cardID string) {
	if m.state == nil {
		return
	}
	// Remove from cards
	for i, c := range m.state.Cards {
		if c.ID == cardID {
			m.state.Cards = append(m.state.Cards[:i], m.state.Cards[i+1:]...)
			break
		}
	}
	// Remove from any group
	for i, g := range m.state.Groups {
		for j, cid := range g.CardIDs {
			if cid == cardID {
				m.state.Groups[i].CardIDs = append(g.CardIDs[:j], g.CardIDs[j+1:]...)
				if len(m.state.Groups[i].CardIDs) < 2 {
					m.state.Groups = append(m.state.Groups[:i], m.state.Groups[i+1:]...)
				}
				break
			}
		}
	}
	m.broadcastState()
}

func (m Model) handleBrainstormInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		text := strings.TrimSpace(m.inputText)
		if text != "" && m.state != nil {
			columns := m.getColumns()
			if m.activeCol < len(columns) {
				col := columns[m.activeCol]
				card := protocol.Card{
					ID:       fmt.Sprintf("cli-%s-%d", m.participantID, len(m.state.Cards)),
					ColumnID: col.id,
					Text:     text,
				}
				m.state.Cards = append(m.state.Cards, card)
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
		if len(msg.Runes) > 0 && len(m.inputText) < 140 {
			m.inputText += string(msg.Runes)
		}
	}
	return m, nil
}

type columnInfo struct {
	id          string
	title       string
	description string
}

func (m Model) getColumns() []columnInfo {
	if m.state == nil {
		return nil
	}

	templateID := m.state.Meta.TemplateID
	tmpl := protocol.GetTemplate(templateID)

	// If template has columns, use that order and metadata
	if len(tmpl.Columns) > 0 {
		var cols []columnInfo
		for _, ct := range tmpl.Columns {
			cols = append(cols, columnInfo{id: ct.ID, title: ct.Title, description: ct.Description})
		}
		return cols
	}

	// Fallback: derive from cards
	seen := make(map[string]bool)
	var cols []columnInfo
	for _, c := range m.state.Cards {
		if !seen[c.ColumnID] {
			seen[c.ColumnID] = true
			ct, ok := protocol.GetColumnTemplate(templateID, c.ColumnID)
			if ok {
				cols = append(cols, columnInfo{id: ct.ID, title: ct.Title, description: ct.Description})
			} else {
				cols = append(cols, columnInfo{id: c.ColumnID, title: c.ColumnID})
			}
		}
	}
	if len(cols) == 0 {
		cols = []columnInfo{
			{id: "stop", title: "Stop", description: "What factors are slowing us down or holding us back?"},
			{id: "start", title: "Start", description: "What factors are driving us forward and enabling our success?"},
		}
	}
	return cols
}

func (m Model) cardsForColumn(colID string) []protocol.Card {
	if m.state == nil {
		return nil
	}
	var result []protocol.Card
	for _, c := range m.state.Cards {
		if c.ColumnID == colID {
			result = append(result, c)
		}
	}
	return result
}

func (m Model) groupsForColumn(colID string) []protocol.Group {
	if m.state == nil {
		return nil
	}
	var result []protocol.Group
	for _, g := range m.state.Groups {
		if g.ColumnID == colID {
			result = append(result, g)
		}
	}
	return result
}

func (m Model) groupedCardIDs() map[string]bool {
	ids := make(map[string]bool)
	if m.state == nil {
		return ids
	}
	for _, g := range m.state.Groups {
		for _, cid := range g.CardIDs {
			ids[cid] = true
		}
	}
	return ids
}

func (m Model) cardByID(id string) (protocol.Card, bool) {
	if m.state == nil {
		return protocol.Card{}, false
	}
	for _, c := range m.state.Cards {
		if c.ID == id {
			return c, true
		}
	}
	return protocol.Card{}, false
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}
