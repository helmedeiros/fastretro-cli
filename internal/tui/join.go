package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/helmedeiros/fastretro-cli/internal/styles"
)

func (m Model) viewJoin() string {
	var b strings.Builder

	muted := lipgloss.NewStyle().Foreground(styles.Muted)

	// Show retro name if available, otherwise room code
	title := "fastRetro CLI"
	if m.state != nil && m.state.Meta.Name != "" {
		title = m.state.Meta.Name
	}
	b.WriteString(styles.Title.Render(title))
	if m.state != nil && m.state.Meta.Date != "" {
		b.WriteString("  " + muted.Render(m.state.Meta.Date))
	}
	b.WriteString("\n")

	roomCode := ""
	if m.client != nil {
		roomCode = m.client.RoomCode
	}
	b.WriteString(muted.Render(fmt.Sprintf("Room: %s | %d peers", roomCode, m.peerCount)))
	b.WriteString("\n\n")
	b.WriteString(lipgloss.NewStyle().Bold(true).Render("Who are you?"))
	b.WriteString("\n\n")

	if m.state != nil {
		for i, p := range m.state.Participants {
			cursor := "  "
			if i == m.cursor {
				cursor = "> "
			}
			if m.takenIDs[p.ID] {
				b.WriteString(styles.Taken.Render(fmt.Sprintf("%s%s (taken)", cursor, p.Name)))
			} else if i == m.cursor {
				b.WriteString(styles.Selected.Render(fmt.Sprintf("%s%s", cursor, p.Name)))
			} else {
				b.WriteString(fmt.Sprintf("%s%s", cursor, p.Name))
			}
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	if m.inputMode {
		b.WriteString(fmt.Sprintf("  New name: %s▌", m.inputText))
	} else {
		b.WriteString("  [n] Add new name  [Enter] Select  [q] Quit")
	}
	b.WriteString("\n")

	return b.String()
}

func (m Model) handleJoinKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.inputMode {
		return m.handleJoinInput(msg)
	}

	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.state != nil && m.cursor < len(m.state.Participants)-1 {
			m.cursor++
		}
	case "enter":
		if m.state != nil && m.cursor < len(m.state.Participants) {
			p := m.state.Participants[m.cursor]
			if !m.takenIDs[p.ID] {
				m.participantID = p.ID
				if m.client != nil {
					if err := m.client.ClaimIdentity(p.ID); err != nil {
						m.err = err
					}
				}
			}
		}
	case "n":
		m.inputMode = true
		m.inputText = ""
	}
	return m, nil
}

func (m Model) handleJoinInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		name := strings.TrimSpace(m.inputText)
		if name != "" {
			m.participantID = name
			if m.client != nil {
				if err := m.client.ClaimIdentity(name); err != nil {
					m.err = err
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
