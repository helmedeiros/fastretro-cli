package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/helmedeiros/fastretro-cli/internal/protocol"
	"github.com/helmedeiros/fastretro-cli/internal/styles"
)

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

	items := m.voteItems()
	for i, item := range items {
		cursor := "  "
		if i == m.cursor {
			cursor = "> "
		}
		votes := m.votesForItem(item.id)
		myVotes := m.myVotesForItem(item.id)

		line := fmt.Sprintf("%s[%d] %s", cursor, i+1, truncate(item.label, 30))
		if votes > 0 {
			badge := styles.VoteBadge.Render(fmt.Sprintf("+%d", votes))
			line += "  " + badge
		}
		if myVotes > 0 {
			line += muted.Render(fmt.Sprintf(" (you: %d)", myVotes))
		}

		if i == m.cursor {
			b.WriteString(styles.Selected.Render(line))
		} else {
			b.WriteString(line)
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(muted.Render("[↑↓] navigate  [Enter/Space] vote  [u] unvote  [q] quit"))

	return b.String()
}

func (m Model) handleVoteKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	items := m.voteItems()

	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(items)-1 {
			m.cursor++
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

type voteItem struct {
	id    string
	label string
}

func (m Model) voteItems() []voteItem {
	if m.state == nil {
		return nil
	}

	var items []voteItem
	groupedCards := make(map[string]bool)

	for _, g := range m.state.Groups {
		items = append(items, voteItem{id: g.ID, label: g.Name})
		for _, cid := range g.CardIDs {
			groupedCards[cid] = true
		}
	}

	for _, c := range m.state.Cards {
		if !groupedCards[c.ID] {
			items = append(items, voteItem{id: c.ID, label: c.Text})
		}
	}

	return items
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
