package tui

import (
	"strings"
	"testing"

	"github.com/helmedeiros/fastretro-cli/internal/protocol"
)

func testCloseModel() Model {
	m := testModel()
	m.participantID = "p1"
	m.state = &protocol.RetroState{
		Stage: "close",
		Meta:  protocol.RetroMeta{Name: "Sprint 42", Date: "2025-08-21"},
		Participants: []protocol.Participant{
			{ID: "p1", Name: "Alice"},
			{ID: "p2", Name: "Bob"},
		},
		Cards: []protocol.Card{
			{ID: "c1", ColumnID: "stop", Text: "long meetings"},
			{ID: "c2", ColumnID: "start", Text: "pair programming"},
			{ID: "c3", ColumnID: "stop", Text: "too many bugs"},
		},
		Groups: []protocol.Group{
			{ID: "g1", ColumnID: "stop", Name: "Process", CardIDs: []string{"c1", "c3"}},
		},
		Votes: []protocol.Vote{
			{ParticipantID: "p1", CardID: "g1"},
			{ParticipantID: "p2", CardID: "g1"},
			{ParticipantID: "p2", CardID: "c2"},
		},
		VoteBudget: 3,
		DiscussNotes: []protocol.DiscussNote{
			{ID: "n1", ParentCardID: "g1", Lane: "actions", Text: "Write runbook"},
		},
		ActionOwners: map[string]string{"n1": "p1"},
	}
	return m
}

func TestViewClose_ShowsStats(t *testing.T) {
	m := testCloseModel()
	view := m.viewClose()

	if !strings.Contains(view, "Participants") {
		t.Error("expected Participants stat")
	}
	if !strings.Contains(view, "2") {
		t.Error("expected participant count")
	}
	if !strings.Contains(view, "Cards") {
		t.Error("expected Cards stat")
	}
	if !strings.Contains(view, "3") {
		t.Error("expected card count")
	}
}

func TestViewClose_ShowsActionItems(t *testing.T) {
	m := testCloseModel()
	view := m.viewClose()

	if !strings.Contains(view, "Write runbook") {
		t.Error("expected action item")
	}
	if !strings.Contains(view, "Alice") {
		t.Error("expected resolved owner name")
	}
}

func TestViewClose_ShowsCheckmark(t *testing.T) {
	m := testCloseModel()
	view := m.viewClose()

	if !strings.Contains(view, "✓") {
		t.Error("expected checkmark")
	}
}

func TestViewClose_ShowsBoardOverview(t *testing.T) {
	m := testCloseModel()
	view := m.viewClose()

	if !strings.Contains(view, "Board overview") {
		t.Error("expected board overview section")
	}
	if !strings.Contains(view, "Process") {
		t.Error("expected group in board")
	}
}

func TestViewClose_NilState(t *testing.T) {
	m := testModel()
	if view := m.viewClose(); view != "" {
		t.Errorf("expected empty, got %q", view)
	}
}

func TestViewClose_NoActions(t *testing.T) {
	m := testModel()
	m.state = &protocol.RetroState{
		Stage: "close",
		Meta:  protocol.RetroMeta{Name: "Test"},
	}
	view := m.viewClose()

	if !strings.Contains(view, "Stats") {
		t.Error("expected stats section")
	}
	// No Action Items section when none exist
	if strings.Contains(view, "Action Items") {
		t.Error("should not show Action Items section when there are none")
	}
}

func TestViewClose_ShowsHelp(t *testing.T) {
	m := testCloseModel()
	view := m.viewClose()

	if !strings.Contains(view, "quit") {
		t.Error("expected quit hint")
	}
}

func TestViewClose_InAppView(t *testing.T) {
	m := testCloseModel()
	view := m.View()

	if !strings.Contains(view, "CLOSE") {
		t.Error("expected CLOSE in stage bar")
	}
}

func TestViewClose_UnassignedOwner(t *testing.T) {
	m := testCloseModel()
	m.state.ActionOwners = nil
	view := m.viewClose()

	if !strings.Contains(view, "unassigned") {
		t.Error("expected 'unassigned' for actions without owners")
	}
}
