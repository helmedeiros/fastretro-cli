package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/helmedeiros/fastretro-cli/internal/protocol"
)

func checkSurveyModel() Model {
	return Model{
		state: &protocol.RetroState{
			Stage: "survey",
			Meta:  protocol.RetroMeta{Type: "check", Name: "HC", TemplateID: "health-check"},
			Participants: []protocol.Participant{
				{ID: "p1", Name: "Alice"},
				{ID: "p2", Name: "Bob"},
			},
			SurveyResponses: []protocol.SurveyResponse{},
		},
		participantID: "p1",
		takenIDs:      map[string]bool{"p1": true},
		width:         80,
		height:        24,
	}
}

func TestViewSurvey_RendersQuestions(t *testing.T) {
	m := checkSurveyModel()
	view := m.viewSurvey()
	if !strings.Contains(view, "Ownership") {
		t.Error("expected Ownership in survey view")
	}
	if !strings.Contains(view, "Value") {
		t.Error("expected Value in survey view")
	}
	if !strings.Contains(view, "Fun") {
		t.Error("expected Fun in survey view")
	}
}

func TestViewSurvey_ShowsAnswerCount(t *testing.T) {
	m := checkSurveyModel()
	view := m.viewSurvey()
	if !strings.Contains(view, "0/9 answered") {
		t.Errorf("expected '0/9 answered' in view, got: %s", view)
	}
}

func TestViewSurvey_ShowsSavedIndicator(t *testing.T) {
	m := checkSurveyModel()
	m.state.SurveyResponses = []protocol.SurveyResponse{
		{ID: "r1", ParticipantID: "p1", QuestionID: "ownership", Rating: 4, Comment: ""},
	}
	view := m.viewSurvey()
	if !strings.Contains(view, "SAVED") {
		t.Error("expected SAVED indicator in view")
	}
}

func TestHandleSurveyKeys_NavigateUpDown(t *testing.T) {
	m := checkSurveyModel()
	if m.cursor != 0 {
		t.Fatalf("expected cursor 0, got %d", m.cursor)
	}

	result, _ := m.handleSurveyKeys(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = result.(Model)
	if m.cursor != 1 {
		t.Errorf("expected cursor 1, got %d", m.cursor)
	}

	result, _ = m.handleSurveyKeys(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = result.(Model)
	if m.cursor != 0 {
		t.Errorf("expected cursor 0, got %d", m.cursor)
	}
}

func TestHandleSurveyKeys_NumberKeyRating(t *testing.T) {
	m := checkSurveyModel()
	result, _ := m.handleSurveyKeys(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'4'}})
	m = result.(Model)
	if len(m.state.SurveyResponses) != 1 {
		t.Fatalf("expected 1 response, got %d", len(m.state.SurveyResponses))
	}
	if m.state.SurveyResponses[0].Rating != 4 {
		t.Errorf("expected rating 4, got %d", m.state.SurveyResponses[0].Rating)
	}
	if m.state.SurveyResponses[0].QuestionID != "ownership" {
		t.Errorf("expected questionID 'ownership', got %q", m.state.SurveyResponses[0].QuestionID)
	}
}

func TestHandleSurveyKeys_InvalidRatingIgnored(t *testing.T) {
	m := checkSurveyModel()
	// Health check options are 1-5, so 7 should be ignored
	result, _ := m.handleSurveyKeys(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'7'}})
	m = result.(Model)
	if len(m.state.SurveyResponses) != 0 {
		t.Errorf("expected 0 responses for invalid rating, got %d", len(m.state.SurveyResponses))
	}
}

func TestHandleSurveyKeys_UpsertRating(t *testing.T) {
	m := checkSurveyModel()
	result, _ := m.handleSurveyKeys(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}})
	m = result.(Model)
	result, _ = m.handleSurveyKeys(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'5'}})
	m = result.(Model)
	if len(m.state.SurveyResponses) != 1 {
		t.Fatalf("expected 1 response (upsert), got %d", len(m.state.SurveyResponses))
	}
	if m.state.SurveyResponses[0].Rating != 5 {
		t.Errorf("expected rating 5, got %d", m.state.SurveyResponses[0].Rating)
	}
}

func TestHandleSurveyKeys_CommentMode(t *testing.T) {
	m := checkSurveyModel()
	// First rate so there's a response to attach comment to
	result, _ := m.handleSurveyKeys(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'4'}})
	m = result.(Model)

	// Enter comment mode
	result, _ = m.handleSurveyKeys(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	m = result.(Model)
	if !m.inputMode {
		t.Fatal("expected input mode")
	}

	// Type comment
	result, _ = m.handleSurveyCommentInput(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}})
	m = result.(Model)
	result, _ = m.handleSurveyCommentInput(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}})
	m = result.(Model)

	// Submit
	result, _ = m.handleSurveyCommentInput(tea.KeyMsg{Type: tea.KeyEnter})
	m = result.(Model)
	if m.inputMode {
		t.Error("expected input mode to be off after enter")
	}
	if m.state.SurveyResponses[0].Comment != "Go" {
		t.Errorf("expected comment 'Go', got %q", m.state.SurveyResponses[0].Comment)
	}
}

func TestHandleSurveyKeys_CursorBounds(t *testing.T) {
	m := checkSurveyModel()
	// Try going up at top
	result, _ := m.handleSurveyKeys(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = result.(Model)
	if m.cursor != 0 {
		t.Errorf("expected cursor to stay at 0, got %d", m.cursor)
	}

	// Go to last question
	for i := 0; i < 20; i++ {
		result, _ = m.handleSurveyKeys(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		m = result.(Model)
	}
	tmpl := protocol.GetCheckTemplate("health-check")
	if m.cursor != len(tmpl.Questions)-1 {
		t.Errorf("expected cursor at last question (%d), got %d", len(tmpl.Questions)-1, m.cursor)
	}
}
