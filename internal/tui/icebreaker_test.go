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

	if !strings.Contains(view, "spin") {
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

func TestViewIcebreaker_ShowsHelpKeys(t *testing.T) {
	m := testIcebreakerModel()
	view := m.viewIcebreaker()

	if !strings.Contains(view, "spin") {
		t.Error("expected spin hint")
	}
	if !strings.Contains(view, "next") {
		t.Error("expected next hint")
	}
}

// --- handleIcebreakerKeys ---

func TestHandleIcebreakerKeys_Spin(t *testing.T) {
	m := testIcebreakerModel()
	m.state.Icebreaker.Question = ""

	result, _ := m.handleIcebreakerKeys(keyMsg("s"))
	model := result.(Model)

	if model.state.Icebreaker.Question == "" {
		t.Error("expected question to be set after spin")
	}
}

func TestHandleIcebreakerKeys_SpinPicksFromPool(t *testing.T) {
	m := testIcebreakerModel()

	result, _ := m.handleIcebreakerKeys(keyMsg("s"))
	model := result.(Model)

	q := model.state.Icebreaker.Question
	found := false
	for _, qq := range model.state.Icebreaker.Questions {
		if q == qq {
			found = true
		}
	}
	if !found {
		t.Errorf("question %q not in pool", q)
	}
}

func TestHandleIcebreakerKeys_Next(t *testing.T) {
	m := testIcebreakerModel()
	m.state.Icebreaker.CurrentIndex = 0

	result, _ := m.handleIcebreakerKeys(keyMsg("n"))
	model := result.(Model)

	if model.state.Icebreaker.CurrentIndex != 1 {
		t.Errorf("expected 1, got %d", model.state.Icebreaker.CurrentIndex)
	}
	if model.state.Icebreaker.Question != "" {
		t.Error("question should be cleared on next")
	}
}

func TestHandleIcebreakerKeys_NextMarksLastDone(t *testing.T) {
	m := testIcebreakerModel()
	m.state.Icebreaker.CurrentIndex = 2 // last participant

	result, _ := m.handleIcebreakerKeys(keyMsg("n"))
	model := result.(Model)

	if model.state.Icebreaker.CurrentIndex != 3 {
		t.Errorf("expected 3 (past end), got %d", model.state.Icebreaker.CurrentIndex)
	}
}

func TestHandleIcebreakerKeys_NextPastEnd(t *testing.T) {
	m := testIcebreakerModel()
	m.state.Icebreaker.CurrentIndex = 3 // already past end

	result, _ := m.handleIcebreakerKeys(keyMsg("n"))
	model := result.(Model)

	if model.state.Icebreaker.CurrentIndex != 3 {
		t.Error("should not go further past end")
	}
}

func TestViewIcebreaker_AllDone(t *testing.T) {
	m := testIcebreakerModel()
	m.state.Icebreaker.CurrentIndex = 3 // past all 3 participants
	view := m.viewIcebreaker()

	if !strings.Contains(view, "complete") {
		t.Error("expected 'complete' message")
	}
	// All should show as done
	if !strings.Contains(view, "done") {
		t.Error("expected done markers")
	}
}

func TestHandleIcebreakerKeys_Prev(t *testing.T) {
	m := testIcebreakerModel()
	m.state.Icebreaker.CurrentIndex = 1

	result, _ := m.handleIcebreakerKeys(keyMsg("p"))
	model := result.(Model)

	if model.state.Icebreaker.CurrentIndex != 0 {
		t.Errorf("expected 0, got %d", model.state.Icebreaker.CurrentIndex)
	}
}

func TestHandleIcebreakerKeys_PrevAtStart(t *testing.T) {
	m := testIcebreakerModel()
	m.state.Icebreaker.CurrentIndex = 0

	result, _ := m.handleIcebreakerKeys(keyMsg("p"))
	model := result.(Model)

	if model.state.Icebreaker.CurrentIndex != 0 {
		t.Error("should stay at start")
	}
}

func TestHandleIcebreakerKeys_NilState(t *testing.T) {
	m := testModel()
	result, _ := m.handleIcebreakerKeys(keyMsg("s"))
	_ = result // should not panic
}

func TestHandleIcebreakerKeys_NilIcebreaker(t *testing.T) {
	m := testModel()
	m.state = &protocol.RetroState{Stage: "icebreaker"}
	result, _ := m.handleIcebreakerKeys(keyMsg("s"))
	_ = result // should not panic
}

func TestHandleKey_IcebreakerStage(t *testing.T) {
	m := testIcebreakerModel()
	result, _ := m.handleKey(keyMsg("s"))
	model := result.(Model)
	if model.state.Icebreaker.Question == "" {
		t.Error("expected icebreaker handler to fire")
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
