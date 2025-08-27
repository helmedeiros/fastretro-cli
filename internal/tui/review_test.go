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
			"n2": "Alice",
		},
	}
	return m
}

func TestViewReview_ShowsActionItems(t *testing.T) {
	m := testReviewModel()
	view := m.viewReview()

	if !strings.Contains(view, "Write runbook") {
		t.Error("expected action 'Write runbook' in view")
	}
	if !strings.Contains(view, "Set up pairing schedule") {
		t.Error("expected action 'Set up pairing schedule' in view")
	}
}

func TestViewReview_ShowsOwners(t *testing.T) {
	m := testReviewModel()
	view := m.viewReview()

	if !strings.Contains(view, "Alice") {
		t.Error("expected owner 'Alice' in view")
	}
}

func TestViewReview_ShowsBoardOverview(t *testing.T) {
	m := testReviewModel()
	view := m.viewReview()

	if !strings.Contains(view, "Board Overview") {
		t.Error("expected 'Board Overview' section")
	}
	if !strings.Contains(view, "1 groups") {
		t.Error("expected group count in overview")
	}
}

func TestViewReview_NilState(t *testing.T) {
	m := testModel()
	view := m.viewReview()

	if view != "" {
		t.Errorf("expected empty view for nil state, got %q", view)
	}
}

func TestViewReview_NoActions(t *testing.T) {
	m := testModel()
	m.state = &protocol.RetroState{
		Stage: "review",
		Cards: []protocol.Card{
			{ID: "c1", ColumnID: "stop", Text: "item"},
		},
	}
	view := m.viewReview()

	if !strings.Contains(view, "No action items") {
		t.Error("expected 'No action items' message")
	}
}

func TestActionNotes(t *testing.T) {
	m := testReviewModel()
	actions := m.actionNotes()

	if len(actions) != 2 {
		t.Fatalf("expected 2 action notes, got %d", len(actions))
	}
	if actions[0].text != "Write runbook" {
		t.Errorf("first action: got %q", actions[0].text)
	}
	if actions[1].text != "Set up pairing schedule" {
		t.Errorf("second action: got %q", actions[1].text)
	}
}

func TestActionNotes_NilState(t *testing.T) {
	m := testModel()
	actions := m.actionNotes()

	if actions != nil {
		t.Errorf("expected nil for nil state, got %v", actions)
	}
}

func TestOwnerForAction(t *testing.T) {
	m := testReviewModel()

	if got := m.ownerForAction("n2"); got != "Alice" {
		t.Errorf("owner for n2: got %q, want 'Alice'", got)
	}
	if got := m.ownerForAction("n3"); got != "" {
		t.Errorf("owner for n3: got %q, want empty", got)
	}
}

func TestOwnerForAction_NilState(t *testing.T) {
	m := testModel()
	if got := m.ownerForAction("n1"); got != "" {
		t.Errorf("expected empty for nil state, got %q", got)
	}
}

func TestOwnerForAction_NilOwners(t *testing.T) {
	m := testModel()
	m.state = &protocol.RetroState{Stage: "review"}
	if got := m.ownerForAction("n1"); got != "" {
		t.Errorf("expected empty for nil owners, got %q", got)
	}
}
