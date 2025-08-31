package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/helmedeiros/fastretro-cli/internal/protocol"
	"github.com/helmedeiros/fastretro-cli/internal/styles"
)

func (m Model) viewGroup() string {
	if m.state == nil {
		return ""
	}

	columns := m.getColumns()
	if len(columns) == 0 {
		return styles.Subtitle.Render("No cards to group")
	}

	var b strings.Builder

	cardIndex := 1
	cardMap := m.buildCardIndex()

	for _, col := range columns {
		b.WriteString(styles.Selected.Render(fmt.Sprintf("Column: %s", col.title)))
		b.WriteString("\n\n")

		// Groups in this column
		groups := m.groupsForColumn(col.id)
		for _, g := range groups {
			b.WriteString(fmt.Sprintf("  [%s] %s (%d)\n", g.ID, g.Name, len(g.CardIDs)))
			for _, cid := range g.CardIDs {
				if card, ok := m.cardByID(cid); ok {
					idx := cardMap[cid]
					b.WriteString(fmt.Sprintf("      [%d] %s\n", idx, truncate(card.Text, 40)))
				}
			}
			b.WriteString("\n")
		}

		// Ungrouped cards
		ungrouped := m.ungroupedCardsForColumn(col.id)
		if len(ungrouped) > 0 {
			b.WriteString(styles.Subtitle.Render("  Ungrouped"))
			b.WriteString("\n")
			for _, c := range ungrouped {
				idx := cardMap[c.ID]
				b.WriteString(fmt.Sprintf("      [%d] %s\n", idx, truncate(c.Text, 40)))
				cardIndex++
			}
			b.WriteString("\n")
		}
	}
	_ = cardIndex

	// Commands help
	if m.inputMode {
		b.WriteString(fmt.Sprintf("\n  > %s▌\n", m.inputText))
		b.WriteString(styles.StatusBar.Render("[Enter] execute  [Esc] cancel"))
	} else {
		b.WriteString("\n")
		b.WriteString(styles.StatusBar.Render("Commands:"))
		b.WriteString("\n")
		b.WriteString(styles.StatusBar.Render("  m <a> <b>   merge notes into cluster"))
		b.WriteString("\n")
		b.WriteString(styles.StatusBar.Render("  r <id>      rename cluster"))
		b.WriteString("\n")
		b.WriteString(styles.StatusBar.Render("  u <id>      ungroup note from cluster"))
		b.WriteString("\n")
		b.WriteString(styles.StatusBar.Render("  :           enter command mode  [q] quit"))
	}

	return b.String()
}

func (m Model) handleGroupKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.inputMode {
		return m.handleGroupInput(msg)
	}

	switch msg.String() {
	case ":":
		m.inputMode = true
		m.inputText = ""
	}
	return m, nil
}

func (m Model) handleGroupInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		cmd := strings.TrimSpace(m.inputText)
		if cmd != "" {
			m.executeGroupCommand(cmd)
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

func (m *Model) executeGroupCommand(cmd string) {
	parts := strings.Fields(cmd)
	if len(parts) == 0 || m.state == nil {
		return
	}

	switch parts[0] {
	case "m":
		// m <cardNumA> <cardNumB> — merge two cards into a group
		if len(parts) < 3 {
			return
		}
		cardA := m.cardIDByIndex(parts[1])
		cardB := m.cardIDByIndex(parts[2])
		if cardA == "" || cardB == "" {
			return
		}
		m.mergeCards(cardA, cardB)
	case "r":
		// r <groupId> <new name...> — rename a group
		if len(parts) < 3 {
			return
		}
		groupID := parts[1]
		name := strings.Join(parts[2:], " ")
		m.renameGroupByID(groupID, name)
	case "u":
		// u <cardNum> — ungroup a card
		if len(parts) < 2 {
			return
		}
		cardID := m.cardIDByIndex(parts[1])
		if cardID == "" {
			return
		}
		m.ungroupCardByID(cardID)
	}
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
		// Create new group with both cards
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

	if m.client != nil {
		if err := m.client.SendState(m.state); err != nil {
			m.err = err
		}
	}
}

func (m *Model) renameGroupByID(groupID, name string) {
	if m.state == nil {
		return
	}
	for i, g := range m.state.Groups {
		if g.ID == groupID {
			m.state.Groups[i].Name = strings.TrimSpace(name)
			if m.client != nil {
				if err := m.client.SendState(m.state); err != nil {
					m.err = err
				}
			}
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
				// Remove group if fewer than 2 cards
				if len(m.state.Groups[i].CardIDs) < 2 {
					m.state.Groups = append(m.state.Groups[:i], m.state.Groups[i+1:]...)
				}
				if m.client != nil {
					if err := m.client.SendState(m.state); err != nil {
						m.err = err
					}
				}
				return
			}
		}
	}
}

// buildCardIndex assigns a sequential number to each card.
func (m Model) buildCardIndex() map[string]int {
	idx := make(map[string]int)
	if m.state == nil {
		return idx
	}
	n := 1
	for _, c := range m.state.Cards {
		idx[c.ID] = n
		n++
	}
	return idx
}

// cardIDByIndex looks up a card ID by its display index number (as string).
func (m Model) cardIDByIndex(numStr string) string {
	var num int
	if _, err := fmt.Sscanf(numStr, "%d", &num); err != nil {
		return ""
	}
	cardMap := m.buildCardIndex()
	for id, idx := range cardMap {
		if idx == num {
			return id
		}
	}
	return ""
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
