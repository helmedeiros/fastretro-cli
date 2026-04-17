package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/helmedeiros/fastretro-cli/internal/protocol"
	"github.com/helmedeiros/fastretro-cli/internal/styles"
	"github.com/helmedeiros/fastretro-cli/internal/widgets"
)

func (m Model) viewClose() string {
	if m.state == nil {
		return ""
	}

	isCheck := m.state.Meta.Type == "check"
	accent := lipgloss.NewStyle().Foreground(styles.Accent).Bold(true)
	muted := lipgloss.NewStyle().Foreground(styles.Muted)

	var b strings.Builder

	// Stats
	b.WriteString(accent.Render("Stats"))
	b.WriteString("\n")
	b.WriteString(muted.Render("────────────────────────────────────────"))
	b.WriteString("\n\n")

	b.WriteString(fmt.Sprintf("  Participants: %s\n", accent.Render(fmt.Sprintf("%d", len(m.state.Participants)))))

	if isCheck {
		tmpl := protocol.GetCheckTemplate(m.state.Meta.TemplateID)
		b.WriteString(fmt.Sprintf("  Questions:    %s\n", accent.Render(fmt.Sprintf("%d", len(tmpl.Questions)))))
	} else {
		b.WriteString(fmt.Sprintf("  Cards:        %s\n", accent.Render(fmt.Sprintf("%d", len(m.state.Cards)))))
		b.WriteString(fmt.Sprintf("  Groups:       %s\n", accent.Render(fmt.Sprintf("%d", len(m.state.Groups)))))
		b.WriteString(fmt.Sprintf("  Votes cast:   %s\n", accent.Render(fmt.Sprintf("%d", len(m.state.Votes)))))
	}

	actions := m.actionNotes()
	b.WriteString(fmt.Sprintf("  Action items: %s\n", accent.Render(fmt.Sprintf("%d", len(actions)))))

	// Action items
	if len(actions) > 0 {
		b.WriteString("\n")
		b.WriteString(accent.Render("Action Items"))
		b.WriteString("\n")
		b.WriteString(muted.Render("────────────────────────────────────────"))
		b.WriteString("\n\n")

		for i, a := range actions {
			owner := m.ownerForAction(a.id)
			ownerStr := muted.Render("unassigned")
			if owner != "" {
				ownerStr = muted.Render(m.participantName(owner))
			}
			checkmark := accent.Render("✓")
			b.WriteString(fmt.Sprintf("  %s %d. %s\n", checkmark, i+1, a.text))
			b.WriteString(fmt.Sprintf("      %s\n\n", ownerStr))
		}
	}

	if isCheck {
		// Survey results
		tmpl := protocol.GetCheckTemplate(m.state.Meta.TemplateID)
		b.WriteString(accent.Render("Survey results"))
		b.WriteString("\n")
		b.WriteString(muted.Render("────────────────────────────────────────"))
		b.WriteString("\n\n")

		for _, q := range tmpl.Questions {
			median := m.medianForItem(q.ID)
			scoreStr := "—"
			if median > 0 {
				scoreStr = fmt.Sprintf("%.1f", median)
			}
			b.WriteString(fmt.Sprintf("  %s  %s\n", accent.Render(scoreStr), q.Title))
		}
	} else {
		// Board overview
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

	b.WriteString("\n")
	completeLabel := "Retro"
	if isCheck {
		completeLabel = "Check"
	}
	b.WriteString(muted.Render(fmt.Sprintf("%s complete! [q] back", completeLabel)))

	return b.String()
}
