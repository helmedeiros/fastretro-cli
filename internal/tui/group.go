package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/helmedeiros/fastretro-cli/internal/protocol"
	"github.com/helmedeiros/fastretro-cli/internal/styles"
)

// groupItem is a flat list entry for cursor navigation.
type groupItem struct {
	kind    string // "card", "group-header"
	cardID  string
	groupID string
	label   string
	grouped bool
	colID   string
}

// flatGroupItems builds a flat navigable list of all cards and group headers.
func (m Model) flatGroupItems() []groupItem {
	if m.state == nil {
		return nil
	}
	var items []groupItem
	grouped := m.groupedCardIDs()

	for _, col := range m.getColumns() {
		// Groups in this column
		for _, g := range m.groupsForColumn(col.id) {
			items = append(items, groupItem{
				kind:    "group-header",
				groupID: g.ID,
				label:   fmt.Sprintf("%s (%d)", g.Name, len(g.CardIDs)),
				colID:   col.id,
			})
			for _, cid := range g.CardIDs {
				if card, ok := m.cardByID(cid); ok {
					items = append(items, groupItem{
						kind:    "card",
						cardID:  cid,
						groupID: g.ID,
						label:   card.Text,
						grouped: true,
						colID:   col.id,
					})
				}
			}
		}
		// Ungrouped cards in this column
		for _, c := range m.cardsForColumn(col.id) {
			if !grouped[c.ID] {
				items = append(items, groupItem{
					kind:   "card",
					cardID: c.ID,
					label:  c.Text,
					colID:  col.id,
				})
			}
		}
	}
	return items
}

func (m Model) viewGroup() string {
	if m.state == nil {
		return ""
	}

	items := m.flatGroupItems()
	if len(items) == 0 {
		return styles.Subtitle.Render("No cards to group")
	}

	var b strings.Builder

	currentCol := ""
	for i, item := range items {
		// Column header
		if item.colID != currentCol {
			if currentCol != "" {
				b.WriteString("\n")
			}
			currentCol = item.colID
			b.WriteString(styles.Subtitle.Render(fmt.Sprintf("── %s ──", currentCol)))
			b.WriteString("\n")
		}

		cursor := "  "
		if i == m.cursor {
			cursor = "> "
		}

		isMergeSource := item.kind == "card" && item.cardID == m.mergeSource

		switch item.kind {
		case "group-header":
			line := fmt.Sprintf("%s┌ %s", cursor, item.label)
			if i == m.cursor {
				b.WriteString(styles.Selected.Render(line))
			} else {
				b.WriteString(styles.Selected.Render("  ┌ " + item.label))
			}
			b.WriteString("\n")

		case "card":
			prefix := "  "
			if item.grouped {
				prefix = "│ "
			}
			text := truncate(item.label, 38)
			line := fmt.Sprintf("%s%s%s", cursor, prefix, text)

			if isMergeSource {
				b.WriteString(styles.VoteBadge.Render(line))
			} else if i == m.cursor {
				b.WriteString(styles.Selected.Render(line))
			} else if item.grouped {
				b.WriteString(fmt.Sprintf("  %s%s", prefix, text))
			} else {
				b.WriteString(fmt.Sprintf("  • %s", text))
			}
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")

	// Rename input mode
	if m.inputMode {
		b.WriteString(fmt.Sprintf("  Rename: %s▌\n", m.inputText))
		b.WriteString(styles.StatusBar.Render("[Enter] save  [Esc] cancel"))
	} else if m.mergeSource != "" {
		b.WriteString(styles.VoteBadge.Render(" MERGE MODE "))
		b.WriteString("  Select target card, then press ")
		b.WriteString(styles.Selected.Render("[m]"))
		b.WriteString("  or ")
		b.WriteString(styles.Selected.Render("[Esc]"))
		b.WriteString(" to cancel\n")
	} else {
		b.WriteString(styles.StatusBar.Render("[↑↓] navigate  [m] merge  [u] ungroup  [r] rename  [q] quit"))
	}

	return b.String()
}

func (m Model) handleGroupKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.inputMode {
		return m.handleGroupRenameInput(msg)
	}

	items := m.flatGroupItems()

	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(items)-1 {
			m.cursor++
		}
	case "m":
		if m.cursor < len(items) {
			item := items[m.cursor]
			if item.kind != "card" {
				break
			}
			if m.mergeSource == "" {
				// First selection
				m.mergeSource = item.cardID
			} else if m.mergeSource != item.cardID {
				// Second selection — merge
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
	case "r":
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
			items := m.flatGroupItems()
			if m.cursor < len(items) {
				item := items[m.cursor]
				if item.kind == "group-header" {
					m.renameGroupByID(item.groupID, name)
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
		if len(msg.String()) == 1 {
			m.inputText += msg.String()
		}
	}
	return m, nil
}

func (m *Model) mergeCards(sourceID, targetID string) {
	if m.state == nil {
		return
	}

	// Find if target is already in a group
	var targetGroup *protocol.Group
	for i, g := range m.state.Groups {
		for _, cid := range g.CardIDs {
			if cid == targetID {
				targetGroup = &m.state.Groups[i]
				break
			}
		}
	}

	// Check if source is already in a group
	for _, g := range m.state.Groups {
		for _, cid := range g.CardIDs {
			if cid == sourceID {
				return // source already grouped, no-op
			}
		}
	}

	if targetGroup != nil {
		// Add source to existing group
		targetGroup.CardIDs = append(targetGroup.CardIDs, sourceID)
	} else {
		// Create new group
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
