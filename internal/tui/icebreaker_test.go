package tui

import (
	"strings"
	"testing"

	"github.com/helmedeiros/fastretro-cli/internal/protocol"
)

func testIcebreakerModel() Model {
	m := testModel()
	m.participantID = "p1"
	m.state = &protocol.RetroState{
		Stage: "icebreaker",
		Participants: []protocol.Participant{
			{ID: "p1", Name: "Alice"},
			{ID: "p2", Name: "Bob"},
			{ID: "p3", Name: "Carol"},
		},
		Icebreaker: &protocol.Icebreaker{
			Question:       "What's your favorite hobby?",
			Questions:      []string{"What's your favorite hobby?", "Best vacation?"},
			ParticipantIDs: []string{"p1", "p2", "p3"},
			CurrentIndex:   0,
		},
	}
	return m
}

func TestViewIcebreaker_ShowsQuestion(t *testing.T) {
	m := testIcebreakerModel()
	view := m.viewIcebreaker()

	if !strings.Contains(view, "favorite hobby") {
		t.Error("expected question in view")
	}
}

func TestViewIcebreaker_ShowsCurrentParticipant(t *testing.T) {
	m := testIcebreakerModel()
	view := m.viewIcebreaker()

	if !strings.Contains(view, "Alice") {
		t.Error("expected current participant 'Alice' in view")
	}
}

func TestViewIcebreaker_ShowsRoundProgress(t *testing.T) {
	m := testIcebreakerModel()
	view := m.viewIcebreaker()

	if !strings.Contains(view, "1 of 3") {
		t.Error("expected round progress '1 of 3'")
	}
}

func TestViewIcebreaker_SecondParticipant(t *testing.T) {
	m := testIcebreakerModel()
	m.state.Icebreaker.CurrentIndex = 1

	view := m.viewIcebreaker()

	if !strings.Contains(view, "2 of 3") {
		t.Error("expected round progress '2 of 3'")
	}
	if !strings.Contains(view, "Bob") {
		t.Error("expected Bob as current participant")
	}
}

func TestViewIcebreaker_DoneParticipants(t *testing.T) {
	m := testIcebreakerModel()
	m.state.Icebreaker.CurrentIndex = 1

	view := m.viewIcebreaker()

	if !strings.Contains(view, "done") {
		t.Error("expected 'done' for completed participants")
	}
}

func TestViewIcebreaker_NilIcebreaker(t *testing.T) {
	m := testModel()
	m.state = &protocol.RetroState{Stage: "icebreaker"}
	view := m.viewIcebreaker()

	if !strings.Contains(view, "Waiting") {
		t.Error("expected waiting message for nil icebreaker")
	}
}

func TestViewIcebreaker_NilState(t *testing.T) {
	m := testModel()
	view := m.viewIcebreaker()

	if !strings.Contains(view, "Waiting") {
		t.Error("expected waiting message for nil state")
	}
}

func TestViewIcebreaker_NoQuestion(t *testing.T) {
	m := testIcebreakerModel()
	m.state.Icebreaker.Question = ""
	m.state.Icebreaker.Questions = nil

	view := m.viewIcebreaker()

	if !strings.Contains(view, "Spin") {
		t.Error("expected spin prompt when no question")
	}
}

func TestViewIcebreaker_QuestionFromIndex(t *testing.T) {
	m := testIcebreakerModel()
	m.state.Icebreaker.Question = ""
	m.state.Icebreaker.CurrentIndex = 1

	view := m.viewIcebreaker()

	if !strings.Contains(view, "Best vacation") {
		t.Error("expected question from index when Question is empty")
	}
}

func TestParticipantName_Found(t *testing.T) {
	m := testIcebreakerModel()

	if name := m.participantName("p1"); name != "Alice" {
		t.Errorf("expected 'Alice', got %q", name)
	}
}

func TestParticipantName_NotFound(t *testing.T) {
	m := testIcebreakerModel()

	if name := m.participantName("unknown"); name != "unknown" {
		t.Errorf("expected 'unknown' as fallback, got %q", name)
	}
}

func TestParticipantName_NilState(t *testing.T) {
	m := testModel()

	if name := m.participantName("p1"); name != "p1" {
		t.Errorf("expected 'p1' as fallback, got %q", name)
	}
}

func TestViewIcebreaker_InAppView(t *testing.T) {
	m := testIcebreakerModel()
	view := m.View()

	if !strings.Contains(view, "ICEBREAKER") {
		t.Error("expected 'ICEBREAKER' in header")
	}
	if !strings.Contains(view, "favorite hobby") {
		t.Error("expected icebreaker content in full view")
	}
}
