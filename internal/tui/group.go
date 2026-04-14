package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/helmedeiros/fastretro-cli/internal/protocol"
	"github.com/helmedeiros/fastretro-cli/internal/styles"
)

// groupItem is a flat list entry for cursor navigation within a column.
type groupItem struct {
	kind    string // "card", "group-header"
	cardID  string
	groupID string
	label   string
	grouped bool
}

// columnGroupItems builds a flat navigable list for one column.
func (m Model) columnGroupItems(colID string) []groupItem {
	if m.state == nil {
		return nil
	}
	var items []groupItem
	grouped := m.groupedCardIDs()

	for _, g := range m.groupsForColumn(colID) {
		items = append(items, groupItem{
			kind:    "group-header",
			groupID: g.ID,
			label:   fmt.Sprintf("%s (%d)", g.Name, len(g.CardIDs)),
		})
		for _, cid := range g.CardIDs {
			if card, ok := m.cardByID(cid); ok {
				items = append(items, groupItem{
					kind:    "card",
					cardID:  cid,
					groupID: g.ID,
					label:   card.Text,
					grouped: true,
				})
			}
		}
	}

	for _, c := range m.cardsForColumn(colID) {
		if !grouped[c.ID] {
			items = append(items, groupItem{
				kind:   "card",
				cardID: c.ID,
				label:  c.Text,
			})
		}
	}
	return items
}

func (m Model) viewGroup() string {
	if m.state == nil {
		return ""
	}

	columns := m.getColumns()
	if len(columns) == 0 {
		return styles.Subtitle.Render("No cards to group")
	}

	var contents []string
	var colStyles []lipgloss.Style
	for ci, col := range columns {
		items := m.columnGroupItems(col.id)
		isActive := ci == m.activeCol

		var lines []string
		for i, item := range items {
			cursor := "  "
			if isActive && i == m.cursor {
				cursor = "> "
			}

			isMergeSource := item.kind == "card" && item.cardID == m.mergeSource

			switch item.kind {
			case "group-header":
				line := fmt.Sprintf("%s┌ %s", cursor, item.label)
				if isActive && i == m.cursor {
					lines = append(lines, styles.Selected.Render(line))
				} else {
					lines = append(lines, styles.Selected.Render("  ┌ "+item.label))
				}

			case "card":
				prefix := "• "
				if item.grouped {
					prefix = "│ "
				}
				text := item.label
				line := fmt.Sprintf("%s%s%s", cursor, prefix, text)

				if isMergeSource {
					lines = append(lines, styles.VoteBadge.Render(line))
				} else if isActive && i == m.cursor {
					lines = append(lines, styles.Selected.Render(line))
				} else {
					lines = append(lines, fmt.Sprintf("  %s%s", prefix, text))
				}
			}
		}

		if len(lines) == 0 {
			lines = append(lines, styles.Subtitle.Render("  (empty)"))
		}

		muted := lipgloss.NewStyle().Foreground(styles.Muted)
		header := col.title
		if isActive {
			header = styles.Selected.Render("▶ " + header)
		}
		if col.description != "" {
			header += "\n" + muted.Render(col.description) + "\n"
		}

		body := strings.Join(lines, "\n")
		content := header + "\n" + body

		contents = append(contents, content)
		style := styles.Column
		if isActive {
			style = style.BorderForeground(styles.Accent)
		}
		colStyles = append(colStyles, style)
	}

	board := joinColumnsEqualHeight(contents, colStyles)

	var help string
	if m.inputMode {
		help = fmt.Sprintf("Rename: %s▌  [Enter] save  [Esc] cancel", m.inputText)
	} else if m.mergeSource != "" {
		help = "MERGE: select target, press [m] to merge  [Esc] cancel"
	} else {
		help = "[j/k] navigate  [h/l] column  [m] merge  [u] ungroup  [e] rename  [q] back"
	}

	return board + "\n\n" + styles.StatusBar.Render(help)
}

func (m Model) handleGroupKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.inputMode {
		return m.handleGroupRenameInput(msg)
	}

	columns := m.getColumns()
	if len(columns) == 0 {
		return m, nil
	}

	items := m.columnGroupItems(columns[m.activeCol].id)

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
	case "m":
		if m.cursor < len(items) {
			item := items[m.cursor]
			if item.kind != "card" {
				break
			}
			if m.mergeSource == "" {
				m.mergeSource = item.cardID
			} else if m.mergeSource != item.cardID {
				m.mergeCards(m.mergeSource, item.cardID)
				m.mergeSource = ""
			}
		}
	case "u":
		if m.cursor < len(items) {
			item := items[m.cursor]
			if item.kind == "card" && item.grouped {
				m.ungroupCardByID(item.cardID)
			}
		}
	case "e":
		if m.cursor < len(items) {
			item := items[m.cursor]
			if item.kind == "group-header" {
				m.inputMode = true
				m.inputText = ""
			}
		}
	case "esc":
		m.mergeSource = ""
	}
	return m, nil
}

func (m Model) handleGroupRenameInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		name := strings.TrimSpace(m.inputText)
		if name != "" {
			columns := m.getColumns()
			if m.activeCol < len(columns) {
				items := m.columnGroupItems(columns[m.activeCol].id)
				if m.cursor < len(items) {
					item := items[m.cursor]
					if item.kind == "group-header" {
						m.renameGroupByID(item.groupID, name)
					}
				}
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

func (m *Model) mergeCards(sourceID, targetID string) {
	if m.state == nil {
		return
	}

	var targetGroup *protocol.Group
	for i, g := range m.state.Groups {
		for _, cid := range g.CardIDs {
			if cid == targetID {
				targetGroup = &m.state.Groups[i]
				break
			}
		}
	}

	for _, g := range m.state.Groups {
		for _, cid := range g.CardIDs {
			if cid == sourceID {
				return
			}
		}
	}

	if targetGroup != nil {
		targetGroup.CardIDs = append(targetGroup.CardIDs, sourceID)
	} else {
		sourceCard, _ := m.cardByID(sourceID)
		targetCard, _ := m.cardByID(targetID)
		newGroup := protocol.Group{
			ID:       fmt.Sprintf("g-cli-%d", len(m.state.Groups)+1),
			ColumnID: sourceCard.ColumnID,
			Name:     truncate(sourceCard.Text, 20) + " + " + truncate(targetCard.Text, 20),
			CardIDs:  []string{targetID, sourceID},
		}
		m.state.Groups = append(m.state.Groups, newGroup)
	}

	m.broadcastState()
}

func (m *Model) renameGroupByID(groupID, name string) {
	if m.state == nil {
		return
	}
	for i, g := range m.state.Groups {
		if g.ID == groupID {
			m.state.Groups[i].Name = strings.TrimSpace(name)
			m.broadcastState()
			return
		}
	}
}

func (m *Model) ungroupCardByID(cardID string) {
	if m.state == nil {
		return
	}
	for i, g := range m.state.Groups {
		for j, cid := range g.CardIDs {
			if cid == cardID {
				m.state.Groups[i].CardIDs = append(g.CardIDs[:j], g.CardIDs[j+1:]...)
				if len(m.state.Groups[i].CardIDs) < 2 {
					m.state.Groups = append(m.state.Groups[:i], m.state.Groups[i+1:]...)
				}
				m.broadcastState()
				return
			}
		}
	}
}

func (m *Model) broadcastState() {
	if m.client != nil {
		if err := m.client.SendState(m.state); err != nil {
			m.err = err
		}
	}
}

func (m Model) ungroupedCardsForColumn(colID string) []protocol.Card {
	grouped := m.groupedCardIDs()
	var result []protocol.Card
	for _, c := range m.cardsForColumn(colID) {
		if !grouped[c.ID] {
			result = append(result, c)
		}
	}
	return result
}
