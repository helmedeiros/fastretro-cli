package tui

import (
	"fmt"
	"strings"

	"github.com/helmedeiros/fastretro-cli/internal/styles"
)

func (m Model) viewIcebreaker() string {
	if m.state == nil || m.state.Icebreaker == nil {
		return styles.Subtitle.Render("Waiting for icebreaker to start...")
	}

	var b strings.Builder

	ib := m.state.Icebreaker

	b.WriteString(styles.Subtitle.Render("Icebreaker"))
	b.WriteString("\n\n")

	// Current question
	question := ib.Question
	if question == "" && len(ib.Questions) > 0 && ib.CurrentIndex < len(ib.Questions) {
		question = ib.Questions[ib.CurrentIndex]
	}
	if question != "" {
		b.WriteString(styles.ActiveCard.Render(fmt.Sprintf("  %s  ", question)))
	} else {
		b.WriteString(styles.Subtitle.Render("  Spin the wheel to get a question!"))
	}
	b.WriteString("\n\n")

	// Current participant
	if ib.CurrentIndex < len(ib.ParticipantIDs) {
		pid := ib.ParticipantIDs[ib.CurrentIndex]
		name := m.participantName(pid)
		b.WriteString(fmt.Sprintf("  Answering: %s", name))
		b.WriteString("\n")
	}

	// Participant order
	if len(ib.ParticipantIDs) > 0 {
		b.WriteString(fmt.Sprintf("\n  Round %d of %d\n", ib.CurrentIndex+1, len(ib.ParticipantIDs)))
		for i, pid := range ib.ParticipantIDs {
			name := m.participantName(pid)
			marker := "  "
			if i == ib.CurrentIndex {
				b.WriteString(styles.Selected.Render(fmt.Sprintf("  > %s\n", name)))
			} else if i < ib.CurrentIndex {
				b.WriteString(styles.Taken.Render(fmt.Sprintf("  %s%s (done)\n", marker, name)))
			} else {
				b.WriteString(fmt.Sprintf("  %s%s\n", marker, name))
			}
		}
	}

	b.WriteString("\n")
	b.WriteString(styles.StatusBar.Render("View only — use web app to spin and advance"))

	return b.String()
}

func (m Model) participantName(id string) string {
	if m.state == nil {
		return id
	}
	for _, p := range m.state.Participants {
		if p.ID == id {
			return p.Name
		}
	}
	return id
}
