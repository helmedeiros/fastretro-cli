package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/helmedeiros/fastretro-cli/internal/protocol"
	"github.com/helmedeiros/fastretro-cli/internal/styles"
)

type voteItem struct {
	id    string
	label string
}

// columnVoteItems returns votable items (groups + ungrouped cards) for a column.
func (m Model) columnVoteItems(colID string) []voteItem {
	if m.state == nil {
		return nil
	}
	var items []voteItem
	grouped := m.groupedCardIDs()

	for _, g := range m.groupsForColumn(colID) {
		items = append(items, voteItem{id: g.ID, label: g.Name})
	}
	for _, c := range m.cardsForColumn(colID) {
		if !grouped[c.ID] {
			items = append(items, voteItem{id: c.ID, label: c.Text})
		}
	}
	return items
}

func (m Model) viewVote() string {
	if m.state == nil {
		return ""
	}

	accent := lipgloss.NewStyle().Foreground(styles.Accent).Bold(true)
	muted := lipgloss.NewStyle().Foreground(styles.Muted)

	var b strings.Builder

	remaining := m.votesRemaining()
	name := m.participantName(m.participantID)
	b.WriteString(fmt.Sprintf("  Voting as: %s  %s",
		accent.Render(name),
		muted.Render(fmt.Sprintf("(%d/%d votes left)", remaining, m.state.VoteBudget))))
	b.WriteString("\n\n")

	columns := m.getColumns()
	var rendered []string

	for ci, col := range columns {
		items := m.columnVoteItems(col.id)
		isActive := ci == m.activeCol

		var lines []string
		for idx, item := range items {
			cursor := "  "
			if isActive && idx == m.cursor {
				cursor = "> "
			}
			votes := m.votesForItem(item.id)
			myVotes := m.myVotesForItem(item.id)

			text := truncate(item.label, 22)
			line := fmt.Sprintf("%s%s", cursor, text)

			if votes > 0 {
				line += "  " + styles.VoteBadge.Render(fmt.Sprintf("+%d", votes))
			}
			if myVotes > 0 {
				line += muted.Render(fmt.Sprintf(" (you: %d)", myVotes))
			}

			if isActive && idx == m.cursor {
				lines = append(lines, styles.Selected.Render(line))
			} else {
				lines = append(lines, line)
			}
		}

		header := col.title
		if isActive {
			header = styles.Selected.Render("▶ " + header)
		}

		body := strings.Join(lines, "\n")
		if len(lines) == 0 {
			body = muted.Render("  (no items)")
		}

		content := header + "\n" + body
		style := styles.Column
		if isActive {
			style = style.BorderForeground(styles.Accent)
		}
		rendered = append(rendered, style.Render(content))
	}

	if len(rendered) > 0 {
		b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, rendered...))
	}

	b.WriteString("\n\n")
	b.WriteString(muted.Render("[↑↓] navigate  [Tab/←→] column  [Enter/Space] vote  [u] unvote  [q] quit"))

	return b.String()
}

func (m Model) handleVoteKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	columns := m.getColumns()
	if len(columns) == 0 {
		return m, nil
	}

	items := m.columnVoteItems(columns[m.activeCol].id)

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
	case "enter", " ":
		if m.cursor < len(items) && m.votesRemaining() > 0 {
			item := items[m.cursor]
			vote := protocol.Vote{
				ParticipantID: m.participantID,
				CardID:        item.id,
			}
			m.state.Votes = append(m.state.Votes, vote)
			m.broadcastState()
		}
	case "u":
		if m.cursor < len(items) {
			item := items[m.cursor]
			m.removeMyVote(item.id)
			m.broadcastState()
		}
	}
	return m, nil
}

func (m Model) votesForItem(itemID string) int {
	if m.state == nil {
		return 0
	}
	count := 0
	for _, v := range m.state.Votes {
		if v.CardID == itemID {
			count++
		}
	}
	return count
}

func (m Model) myVotesForItem(itemID string) int {
	if m.state == nil {
		return 0
	}
	count := 0
	for _, v := range m.state.Votes {
		if v.CardID == itemID && v.ParticipantID == m.participantID {
			count++
		}
	}
	return count
}

func (m Model) votesRemaining() int {
	if m.state == nil {
		return 0
	}
	used := 0
	for _, v := range m.state.Votes {
		if v.ParticipantID == m.participantID {
			used++
		}
	}
	return m.state.VoteBudget - used
}

func (m *Model) removeMyVote(itemID string) {
	if m.state == nil {
		return
	}
	for i := len(m.state.Votes) - 1; i >= 0; i-- {
		v := m.state.Votes[i]
		if v.CardID == itemID && v.ParticipantID == m.participantID {
			m.state.Votes = append(m.state.Votes[:i], m.state.Votes[i+1:]...)
			return
		}
	}
}
