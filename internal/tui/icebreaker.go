package tui

import (
	"fmt"
	"math/rand"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/helmedeiros/fastretro-cli/internal/styles"
)

func (m Model) viewIcebreaker() string {
	if m.state == nil || m.state.Icebreaker == nil {
		return lipgloss.NewStyle().Foreground(styles.Muted).Render("Waiting for icebreaker to start...")
	}

	accent := lipgloss.NewStyle().Foreground(styles.Accent).Bold(true)
	muted := lipgloss.NewStyle().Foreground(styles.Muted)

	var b strings.Builder

	ib := m.state.Icebreaker

	// Current question
	question := ib.Question
	if question == "" && len(ib.Questions) > 0 && ib.CurrentIndex < len(ib.Questions) {
		question = ib.Questions[ib.CurrentIndex]
	}
	if question != "" {
		b.WriteString(styles.ActiveCard.Render(fmt.Sprintf("  %s  ", question)))
	} else {
		b.WriteString(muted.Render("  Press [s] to spin a question!"))
	}
	b.WriteString("\n\n")

	// Current participant
	allDone := ib.CurrentIndex >= len(ib.ParticipantIDs)
	if !allDone {
		pid := ib.ParticipantIDs[ib.CurrentIndex]
		name := m.participantName(pid)
		b.WriteString(fmt.Sprintf("  Answering: %s", accent.Render(name)))
		b.WriteString("\n")
	} else {
		b.WriteString(accent.Render("  Icebreaker complete!"))
		b.WriteString("\n")
	}

	// Participant order
	if len(ib.ParticipantIDs) > 0 {
		round := ib.CurrentIndex + 1
		if allDone {
			round = len(ib.ParticipantIDs)
		}
		b.WriteString(fmt.Sprintf("\n  %s\n\n", muted.Render(
			fmt.Sprintf("Round %d of %d", round, len(ib.ParticipantIDs)))))
		for i, pid := range ib.ParticipantIDs {
			name := m.participantName(pid)
			if i == ib.CurrentIndex {
				b.WriteString(accent.Render(fmt.Sprintf("  > %s", name)))
			} else if i < ib.CurrentIndex {
				b.WriteString(muted.Render(fmt.Sprintf("    %s (done)", name)))
			} else {
				b.WriteString(fmt.Sprintf("    %s", name))
			}
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(muted.Render("[s] spin question  [n] next person  [p] prev person"))

	return b.String()
}

func (m Model) handleIcebreakerKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.state == nil || m.state.Icebreaker == nil {
		return m, nil
	}

	ib := m.state.Icebreaker

	switch msg.String() {
	case "s":
		// Spin: pick a random question
		if len(ib.Questions) > 0 {
			m.state.Icebreaker.Question = ib.Questions[rand.Intn(len(ib.Questions))]
			m.broadcastState()
		}
	case "n":
		// Next participant (allows going one past the end to mark last as done)
		if ib.CurrentIndex < len(ib.ParticipantIDs) {
			m.state.Icebreaker.CurrentIndex++
			m.state.Icebreaker.Question = ""
			m.broadcastState()
		}
	case "p":
		// Previous participant
		if ib.CurrentIndex > 0 {
			m.state.Icebreaker.CurrentIndex--
			m.state.Icebreaker.Question = ""
			m.broadcastState()
		}
	}
	return m, nil
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
