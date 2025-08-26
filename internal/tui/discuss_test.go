package tui

import (
	"strings"
	"testing"

	"github.com/helmedeiros/fastretro-cli/internal/protocol"
)

func testDiscussModel() Model {
	m := testModel()
	m.participantID = "p1"
	m.state = &protocol.RetroState{
		Stage: "discuss",
		Cards: []protocol.Card{
			{ID: "c1", ColumnID: "stop", Text: "long meetings"},
		},
		Groups: []protocol.Group{
			{ID: "g1", ColumnID: "stop", Name: "Process", CardIDs: []string{"c1"}},
		},
		Discuss: &protocol.DiscussState{
			Order:        []string{"g1"},
			CurrentIndex: 0,
			Segment:      "context",
		},
		DiscussNotes: []protocol.DiscussNote{
			{ID: "n1", ParentCardID: "g1", Lane: "context", Text: "Need clarity on process"},
			{ID: "n2", ParentCardID: "g1", Lane: "actions", Text: "Write runbook"},
		},
	}
	return m
}

func TestViewDiscuss_ShowsCurrentItem(t *testing.T) {
	m := testDiscussModel()
	view := m.viewDiscuss()

	if !strings.Contains(view, "Process") {
		t.Error("expected current discuss item 'Process' in view")
	}
	if !strings.Contains(view, "1 of 1") {
		t.Error("expected '1 of 1' progress indicator")
	}
}

func TestViewDiscuss_ShowsContextNotes(t *testing.T) {
	m := testDiscussModel()
	view := m.viewDiscuss()

	if !strings.Contains(view, "Need clarity on process") {
		t.Error("expected context note in view")
	}
}

func TestViewDiscuss_NoDiscussState(t *testing.T) {
	m := testModel()
	m.state = &protocol.RetroState{Stage: "discuss"}
	view := m.viewDiscuss()

	if !strings.Contains(view, "Waiting") {
		t.Error("expected waiting message when discuss state is nil")
	}
}

func TestViewDiscuss_NilState(t *testing.T) {
	m := testModel()
	view := m.viewDiscuss()

	if !strings.Contains(view, "Waiting") {
		t.Error("expected waiting message for nil state")
	}
}

func TestViewDiscuss_WithTimer(t *testing.T) {
	m := testDiscussModel()
	m.state.Timer = &protocol.Timer{
		Status:      "running",
		RemainingMs: 150000,
	}
	view := m.viewDiscuss()

	if !strings.Contains(view, "2:30") {
		t.Error("expected timer display '2:30'")
	}
}

func TestViewDiscuss_CompletedDiscussion(t *testing.T) {
	m := testDiscussModel()
	m.state.Discuss.CurrentIndex = 1 // past end of order

	view := m.viewDiscuss()

	if !strings.Contains(view, "complete") {
		t.Error("expected 'complete' when index past end")
	}
}

func TestLabelForItem_Group(t *testing.T) {
	m := testDiscussModel()
	label := m.labelForItem("g1")

	if label != "Process" {
		t.Errorf("expected 'Process', got %q", label)
	}
}

func TestLabelForItem_Card(t *testing.T) {
	m := testDiscussModel()
	label := m.labelForItem("c1")

	if label != "long meetings" {
		t.Errorf("expected 'long meetings', got %q", label)
	}
}

func TestLabelForItem_Unknown(t *testing.T) {
	m := testDiscussModel()
	label := m.labelForItem("unknown-id")

	if label != "unknown-id" {
		t.Errorf("expected 'unknown-id', got %q", label)
	}
}

func TestLabelForItem_NilState(t *testing.T) {
	m := testModel()
	label := m.labelForItem("c1")

	if label != "c1" {
		t.Errorf("expected 'c1' as fallback, got %q", label)
	}
}

func TestNotesForItem_Context(t *testing.T) {
	m := testDiscussModel()
	notes := m.notesForItem("g1", "context")

	if len(notes) != 1 {
		t.Fatalf("expected 1 context note, got %d", len(notes))
	}
	if notes[0].Text != "Need clarity on process" {
		t.Errorf("unexpected note text: %q", notes[0].Text)
	}
}

func TestNotesForItem_Actions(t *testing.T) {
	m := testDiscussModel()
	notes := m.notesForItem("g1", "actions")

	if len(notes) != 1 {
		t.Fatalf("expected 1 action note, got %d", len(notes))
	}
	if notes[0].Text != "Write runbook" {
		t.Errorf("unexpected note text: %q", notes[0].Text)
	}
}

func TestNotesForItem_Empty(t *testing.T) {
	m := testDiscussModel()
	notes := m.notesForItem("nonexistent", "context")

	if len(notes) != 0 {
		t.Errorf("expected 0 notes, got %d", len(notes))
	}
}

func TestNotesForItem_NilState(t *testing.T) {
	m := testModel()
	notes := m.notesForItem("g1", "context")

	if notes != nil {
		t.Errorf("expected nil for nil state, got %v", notes)
	}
}

func TestViewDiscuss_DefaultSegment(t *testing.T) {
	m := testDiscussModel()
	m.state.Discuss.Segment = ""
	view := m.viewDiscuss()

	if !strings.Contains(view, "context") {
		t.Error("expected 'context' as default segment")
	}
}
