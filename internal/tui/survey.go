package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/helmedeiros/fastretro-cli/internal/protocol"
	"github.com/helmedeiros/fastretro-cli/internal/styles"
)

func (m Model) viewSurvey() string {
	if m.state == nil || m.state.Meta.Type != "check" {
		return styles.Subtitle.Render("Waiting for survey to start...")
	}

	tmpl := protocol.GetCheckTemplate(m.state.Meta.TemplateID)
	if len(tmpl.Questions) == 0 {
		return styles.Subtitle.Render("No questions in template")
	}

	var b strings.Builder

	for i, q := range tmpl.Questions {
		isCurrent := i == m.cursor

		// Title line
		title := q.Title
		if isCurrent {
			title = styles.Selected.Render("> " + title)
		} else {
			title = "  " + title
		}
		b.WriteString(title)
		b.WriteString("\n")

		// Description (only for current)
		if isCurrent {
			b.WriteString(styles.Subtitle.Render("  " + q.Description))
			b.WriteString("\n")
		}

		// Rating display
		response := m.surveyResponseFor(q.ID)
		b.WriteString("  ")
		for _, opt := range q.Options {
			if response != nil && response.Rating == opt.Value {
				b.WriteString(styles.VoteBadge.Render(fmt.Sprintf("[%s]", opt.Label)))
			} else {
				b.WriteString(fmt.Sprintf(" %s ", opt.Label))
			}
			b.WriteString(" ")
		}

		// Comment indicator
		if response != nil && response.Comment != "" {
			b.WriteString(styles.Subtitle.Render(" // " + response.Comment))
		}

		// Saved indicator
		if response != nil {
			b.WriteString("  ")
			b.WriteString(styles.Subtitle.Render("SAVED"))
		}

		b.WriteString("\n\n")
	}

	// Status
	answered := m.surveyAnsweredCount()
	total := len(tmpl.Questions)
	b.WriteString(styles.StatusBar.Render(
		fmt.Sprintf("%d/%d answered  [j/k] navigate  [1-%d] rate  [e] comment  [q] back",
			answered, total, len(tmpl.Questions[0].Options))))

	return b.String()
}

func (m Model) surveyResponseFor(questionID string) *protocol.SurveyResponse {
	if m.state == nil {
		return nil
	}
	for i := range m.state.SurveyResponses {
		r := &m.state.SurveyResponses[i]
		if r.ParticipantID == m.participantID && r.QuestionID == questionID {
			return r
		}
	}
	return nil
}

func (m Model) surveyAnsweredCount() int {
	if m.state == nil {
		return 0
	}
	count := 0
	for _, r := range m.state.SurveyResponses {
		if r.ParticipantID == m.participantID {
			count++
		}
	}
	return count
}

func (m Model) handleSurveyKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.inputMode {
		return m.handleSurveyCommentInput(msg)
	}

	if m.state == nil || m.state.Meta.Type != "check" {
		return m, nil
	}

	tmpl := protocol.GetCheckTemplate(m.state.Meta.TemplateID)
	if len(tmpl.Questions) == 0 {
		return m, nil
	}

	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(tmpl.Questions)-1 {
			m.cursor++
		}
	case "e":
		m.inputMode = true
		m.inputText = ""
		existing := m.surveyResponseFor(tmpl.Questions[m.cursor].ID)
		if existing != nil {
			m.inputText = existing.Comment
		}
	default:
		// Number key rating
		if len(msg.Runes) == 1 {
			r := msg.Runes[0]
			if r >= '1' && r <= '9' {
				rating := int(r - '0')
				q := tmpl.Questions[m.cursor]
				valid := false
				for _, opt := range q.Options {
					if opt.Value == rating {
						valid = true
						break
					}
				}
				if valid {
					m.submitSurveyRating(q.ID, rating)
				}
			}
		}
	}
	return m, nil
}

func (m *Model) submitSurveyRating(questionID string, rating int) {
	if m.state == nil || m.participantID == "" {
		return
	}

	// Upsert
	found := false
	for i := range m.state.SurveyResponses {
		r := &m.state.SurveyResponses[i]
		if r.ParticipantID == m.participantID && r.QuestionID == questionID {
			r.Rating = rating
			found = true
			break
		}
	}
	if !found {
		m.state.SurveyResponses = append(m.state.SurveyResponses, protocol.SurveyResponse{
			ID:            fmt.Sprintf("cli-%s-%s", m.participantID, questionID),
			ParticipantID: m.participantID,
			QuestionID:    questionID,
			Rating:        rating,
		})
	}
	m.broadcastState()
}

func (m Model) handleSurveyCommentInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.state == nil {
		return m, nil
	}

	tmpl := protocol.GetCheckTemplate(m.state.Meta.TemplateID)
	if m.cursor >= len(tmpl.Questions) {
		return m, nil
	}
	questionID := tmpl.Questions[m.cursor].ID

	switch msg.String() {
	case "enter":
		comment := strings.TrimSpace(m.inputText)
		// Update comment on existing response, or create one with rating 0 (to be rated later)
		found := false
		for i := range m.state.SurveyResponses {
			r := &m.state.SurveyResponses[i]
			if r.ParticipantID == m.participantID && r.QuestionID == questionID {
				r.Comment = comment
				found = true
				break
			}
		}
		if !found && comment != "" {
			m.state.SurveyResponses = append(m.state.SurveyResponses, protocol.SurveyResponse{
				ID:            fmt.Sprintf("cli-%s-%s", m.participantID, questionID),
				ParticipantID: m.participantID,
				QuestionID:    questionID,
				Rating:        0,
				Comment:       comment,
			})
		}
		m.broadcastState()
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
