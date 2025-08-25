package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/helmedeiros/fastretro-cli/internal/protocol"
	"github.com/helmedeiros/fastretro-cli/internal/styles"
)

func (m Model) viewBrainstorm() string {
	if m.state == nil {
		return ""
	}

	columns := m.getColumns()
	if len(columns) == 0 {
		return styles.Subtitle.Render("No columns defined")
	}

	var rendered []string
	for i, col := range columns {
		cards := m.cardsForColumn(col.id)
		var lines []string
		for _, c := range cards {
			text := truncate(c.Text, 32)
			lines = append(lines, fmt.Sprintf("  • %s", text))
		}

		header := col.title
		if i == m.activeCol {
			header = styles.Selected.Render("▶ " + header)
		}

		body := strings.Join(lines, "\n")
		if len(lines) == 0 {
			body = styles.Subtitle.Render("  (empty)")
		}

		if m.inputMode && i == m.activeCol {
			body += fmt.Sprintf("\n\n  Add: %s▌", m.inputText)
		}

		content := header + "\n" + body
		style := styles.Column
		if i == m.activeCol {
			style = style.BorderForeground(styles.Accent)
		}
		rendered = append(rendered, style.Render(content))
	}

	board := lipgloss.JoinHorizontal(lipgloss.Top, rendered...)

	help := "[Tab] switch column  [a] add card  [←→] navigate  [q] quit"
	if m.inputMode {
		help = "[Enter] submit  [Esc] cancel"
	}

	return board + "\n\n" + styles.StatusBar.Render(help)
}

func (m Model) handleBrainstormKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.inputMode {
		return m.handleBrainstormInput(msg)
	}

	columns := m.getColumns()

	switch msg.String() {
	case "tab", "right", "l":
		if len(columns) > 0 {
			m.activeCol = (m.activeCol + 1) % len(columns)
		}
	case "shift+tab", "left", "h":
		if len(columns) > 0 {
			m.activeCol = (m.activeCol - 1 + len(columns)) % len(columns)
		}
	case "a":
		m.inputMode = true
		m.inputText = ""
	}
	return m, nil
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
				if m.client != nil {
					if err := m.client.SendState(m.state); err != nil {
						m.err = err
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
		if len(msg.String()) == 1 && len(m.inputText) < 140 {
			m.inputText += msg.String()
		}
	}
	return m, nil
}

type columnInfo struct {
	id    string
	title string
}

func (m Model) getColumns() []columnInfo {
	if m.state == nil {
		return nil
	}
	seen := make(map[string]bool)
	var cols []columnInfo
	for _, c := range m.state.Cards {
		if !seen[c.ColumnID] {
			seen[c.ColumnID] = true
			cols = append(cols, columnInfo{id: c.ColumnID, title: c.ColumnID})
		}
	}
	if len(cols) == 0 {
		cols = []columnInfo{{id: "stop", title: "Stop"}, {id: "start", title: "Start"}}
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

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}
