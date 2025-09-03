package tui

import (
	"strings"
	"testing"

	"github.com/helmedeiros/fastretro-cli/internal/protocol"
)

func testReviewModel() Model {
	m := testModel()
	m.participantID = "p1"
	m.state = &protocol.RetroState{
		Stage: "review",
		Participants: []protocol.Participant{
			{ID: "p1", Name: "Alice"},
			{ID: "p2", Name: "Bob"},
		},
		Cards: []protocol.Card{
			{ID: "c1", ColumnID: "stop", Text: "long meetings"},
			{ID: "c2", ColumnID: "start", Text: "pair programming"},
		},
		Groups: []protocol.Group{
			{ID: "g1", ColumnID: "stop", Name: "Process", CardIDs: []string{"c1"}},
		},
		DiscussNotes: []protocol.DiscussNote{
			{ID: "n1", ParentCardID: "g1", Lane: "context", Text: "We need clarity"},
			{ID: "n2", ParentCardID: "g1", Lane: "actions", Text: "Write runbook"},
			{ID: "n3", ParentCardID: "c2", Lane: "actions", Text: "Set up pairing schedule"},
		},
		ActionOwners: map[string]string{
			"n2": "p1",
		},
	}
	return m
}

// --- viewReview ---

func TestViewReview_ShowsActionItems(t *testing.T) {
	m := testReviewModel()
	view := m.viewReview()

	if !strings.Contains(view, "Write runbook") {
		t.Error("expected action item")
	}
	if !strings.Contains(view, "Set up pairing schedule") {
		t.Error("expected second action item")
	}
}

func TestViewReview_ShowsOwner(t *testing.T) {
	m := testReviewModel()
	view := m.viewReview()

	if !strings.Contains(view, "Alice") {
		t.Error("expected owner name resolved to Alice")
	}
}

func TestViewReview_ShowsUnassigned(t *testing.T) {
	m := testReviewModel()
	view := m.viewReview()

	if !strings.Contains(view, "unassigned") {
		t.Error("expected 'unassigned' for action without owner")
	}
}

func TestViewReview_ShowsBoardOverview(t *testing.T) {
	m := testReviewModel()
	view := m.viewReview()

	if !strings.Contains(view, "Board overview") {
		t.Error("expected board overview section")
	}
	if !strings.Contains(view, "Process") {
		t.Error("expected group in board overview")
	}
}

func TestViewReview_ShowsHelp(t *testing.T) {
	m := testReviewModel()
	view := m.viewReview()

	if !strings.Contains(view, "assign owner") {
		t.Error("expected assign owner hint")
	}
}

func TestViewReview_ShowsCursor(t *testing.T) {
	m := testReviewModel()
	m.cursor = 0
	view := m.viewReview()

	if !strings.Contains(view, ">") {
		t.Error("expected cursor indicator")
	}
}

func TestViewReview_ShowsCheckmark(t *testing.T) {
	m := testReviewModel()
	view := m.viewReview()

	if !strings.Contains(view, "✓") {
		t.Error("expected checkmark")
	}
}

func TestViewReview_InputMode(t *testing.T) {
	m := testReviewModel()
	m.inputMode = true
	m.inputText = "Bob"
	view := m.viewReview()

	if !strings.Contains(view, "Assign owner") {
		t.Error("expected assign prompt")
	}
	if !strings.Contains(view, "Bob") {
		t.Error("expected input text")
	}
	if !strings.Contains(view, "Alice") && !strings.Contains(view, "Participants") {
		t.Error("expected participant list hint")
	}
}

func TestViewReview_NilState(t *testing.T) {
	m := testModel()
	if view := m.viewReview(); view != "" {
		t.Errorf("expected empty, got %q", view)
	}
}

func TestViewReview_NoActions(t *testing.T) {
	m := testModel()
	m.state = &protocol.RetroState{Stage: "review"}
	view := m.viewReview()

	if !strings.Contains(view, "No action items") {
		t.Error("expected no actions message")
	}
}

// --- handleReviewKeys ---

func TestHandleReviewKeys_Up(t *testing.T) {
	m := testReviewModel()
	m.cursor = 1

	result, _ := m.handleReviewKeys(keyMsg("up"))
	model := result.(Model)

	if model.cursor != 0 {
		t.Errorf("expected 0, got %d", model.cursor)
	}
}

func TestHandleReviewKeys_Down(t *testing.T) {
	m := testReviewModel()

	result, _ := m.handleReviewKeys(keyMsg("down"))
	model := result.(Model)

	if model.cursor != 1 {
		t.Errorf("expected 1, got %d", model.cursor)
	}
}

func TestHandleReviewKeys_UpAtTop(t *testing.T) {
	m := testReviewModel()

	result, _ := m.handleReviewKeys(keyMsg("up"))
	model := result.(Model)

	if model.cursor != 0 {
		t.Error("should stay at 0")
	}
}

func TestHandleReviewKeys_DownAtBottom(t *testing.T) {
	m := testReviewModel()
	m.cursor = 1

	result, _ := m.handleReviewKeys(keyMsg("down"))
	model := result.(Model)

	if model.cursor != 1 {
		t.Error("should stay at bottom")
	}
}

func TestHandleReviewKeys_AssignOwner(t *testing.T) {
	m := testReviewModel()

	result, _ := m.handleReviewKeys(keyMsg("o"))
	model := result.(Model)

	if !model.inputMode {
		t.Error("expected input mode")
	}
}

func TestHandleReviewKeys_VimK(t *testing.T) {
	m := testReviewModel()
	m.cursor = 1

	result, _ := m.handleReviewKeys(keyMsg("k"))
	model := result.(Model)

	if model.cursor != 0 {
		t.Errorf("expected 0")
	}
}

func TestHandleReviewKeys_VimJ(t *testing.T) {
	m := testReviewModel()

	result, _ := m.handleReviewKeys(keyMsg("j"))
	model := result.(Model)

	if model.cursor != 1 {
		t.Errorf("expected 1")
	}
}

func TestHandleReviewKeys_DelegatesToInput(t *testing.T) {
	m := testReviewModel()
	m.inputMode = true
	m.inputText = "A"

	result, _ := m.handleReviewKeys(keyMsg("l"))
	model := result.(Model)

	if model.inputText != "Al" {
		t.Errorf("got %q", model.inputText)
	}
}

// --- handleReviewInput ---

func TestHandleReviewInput_Enter(t *testing.T) {
	m := testReviewModel()
	m.inputMode = true
	m.inputText = "Bob"
	m.cursor = 1 // n3, unassigned

	result, _ := m.handleReviewInput(keyMsg("enter"))
	model := result.(Model)

	if model.inputMode {
		t.Error("should exit input")
	}
	owner := model.state.ActionOwners["n3"]
	if owner != "p2" {
		t.Errorf("expected owner 'p2' (Bob), got %q", owner)
	}
}

func TestHandleReviewInput_EnterMatchesName(t *testing.T) {
	m := testReviewModel()
	m.inputMode = true
	m.inputText = "alice" // case insensitive
	m.cursor = 1

	result, _ := m.handleReviewInput(keyMsg("enter"))
	model := result.(Model)

	if model.state.ActionOwners["n3"] != "p1" {
		t.Errorf("expected p1 (Alice), got %q", model.state.ActionOwners["n3"])
	}
}

func TestHandleReviewInput_EnterEmpty(t *testing.T) {
	m := testReviewModel()
	m.inputMode = true
	m.inputText = "  "

	result, _ := m.handleReviewInput(keyMsg("enter"))
	model := result.(Model)

	if _, ok := model.state.ActionOwners["n3"]; ok {
		t.Error("empty should not assign")
	}
}

func TestHandleReviewInput_Escape(t *testing.T) {
	m := testReviewModel()
	m.inputMode = true

	result, _ := m.handleReviewInput(keyMsg("esc"))
	model := result.(Model)

	if model.inputMode {
		t.Error("should exit")
	}
}

func TestHandleReviewInput_Backspace(t *testing.T) {
	m := testReviewModel()
	m.inputMode = true
	m.inputText = "Bo"

	result, _ := m.handleReviewInput(keyMsg("backspace"))
	model := result.(Model)

	if model.inputText != "B" {
		t.Errorf("got %q", model.inputText)
	}
}

func TestHandleReviewInput_Type(t *testing.T) {
	m := testReviewModel()
	m.inputMode = true
	m.inputText = "B"

	result, _ := m.handleReviewInput(keyMsg("o"))
	model := result.(Model)

	if model.inputText != "Bo" {
		t.Errorf("got %q", model.inputText)
	}
}

// --- helpers ---

func TestActionNotes(t *testing.T) {
	m := testReviewModel()
	actions := m.actionNotes()

	if len(actions) != 2 {
		t.Fatalf("expected 2, got %d", len(actions))
	}
	if actions[0].text != "Write runbook" {
		t.Errorf("got %q", actions[0].text)
	}
}

func TestActionNotes_NilState(t *testing.T) {
	m := testModel()
	if actions := m.actionNotes(); actions != nil {
		t.Errorf("expected nil")
	}
}

func TestOwnerForAction(t *testing.T) {
	m := testReviewModel()
	if got := m.ownerForAction("n2"); got != "p1" {
		t.Errorf("got %q", got)
	}
	if got := m.ownerForAction("n3"); got != "" {
		t.Errorf("got %q", got)
	}
}

func TestOwnerForAction_NilState(t *testing.T) {
	m := testModel()
	if got := m.ownerForAction("n1"); got != "" {
		t.Errorf("got %q", got)
	}
}

func TestViewReview_InAppView(t *testing.T) {
	m := testReviewModel()
	if view := m.View(); !strings.Contains(view, "REVIEW") {
		t.Error("expected REVIEW in header")
	}
}

func TestHandleKey_ReviewStage(t *testing.T) {
	m := testReviewModel()
	result, _ := m.handleKey(keyMsg("o"))
	model := result.(Model)
	if !model.inputMode {
		t.Error("expected review key handler")
	}
}
