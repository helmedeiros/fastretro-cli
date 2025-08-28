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
		ActionOwners: map[string]string{"n1": "Alice"},
	}
	return m
}

func TestViewClose_ShowsMeta(t *testing.T) {
	m := testCloseModel()
	view := m.viewClose()

	if !strings.Contains(view, "Sprint 42") {
		t.Error("expected name 'Sprint 42' in view")
	}
	if !strings.Contains(view, "2025-08-21") {
		t.Error("expected date in view")
	}
}

func TestViewClose_ShowsStats(t *testing.T) {
	m := testCloseModel()
	view := m.viewClose()

	if !strings.Contains(view, "Participants: 2") {
		t.Error("expected 'Participants: 2'")
	}
	if !strings.Contains(view, "Cards: 3") {
		t.Error("expected 'Cards: 3'")
	}
	if !strings.Contains(view, "Groups: 1") {
		t.Error("expected 'Groups: 1'")
	}
	if !strings.Contains(view, "Votes cast: 3") {
		t.Error("expected 'Votes cast: 3'")
	}
	if !strings.Contains(view, "Action items: 1") {
		t.Error("expected 'Action items: 1'")
	}
}

func TestViewClose_ShowsActionItems(t *testing.T) {
	m := testCloseModel()
	view := m.viewClose()

	if !strings.Contains(view, "Write runbook") {
		t.Error("expected action item in view")
	}
	if !strings.Contains(view, "Alice") {
		t.Error("expected owner in view")
	}
}

func TestViewClose_ShowsBoard(t *testing.T) {
	m := testCloseModel()
	view := m.viewClose()

	if !strings.Contains(view, "Board") {
		t.Error("expected 'Board' section")
	}
}

func TestViewClose_NilState(t *testing.T) {
	m := testModel()
	view := m.viewClose()

	if view != "" {
		t.Errorf("expected empty view for nil state, got %q", view)
	}
}

func TestViewClose_NoActions(t *testing.T) {
	m := testModel()
	m.state = &protocol.RetroState{
		Stage: "close",
		Meta:  protocol.RetroMeta{Name: "Test"},
	}
	view := m.viewClose()

	if !strings.Contains(view, "Summary") {
		t.Error("expected summary header")
	}
	// Should not contain Action Items section
	if strings.Contains(view, "Action Items") {
		t.Error("should not show Action Items section when there are none")
	}
}

func TestViewClose_NoName(t *testing.T) {
	m := testModel()
	m.state = &protocol.RetroState{
		Stage: "close",
		Meta:  protocol.RetroMeta{Date: "2025-08-21"},
	}
	view := m.viewClose()

	if strings.Contains(view, "Name:") {
		t.Error("should not show Name field when empty")
	}
	if !strings.Contains(view, "2025-08-21") {
		t.Error("expected date to be shown")
	}
}
