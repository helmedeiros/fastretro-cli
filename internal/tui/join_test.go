package tui

import (
	"strings"
	"testing"

	"github.com/helmedeiros/fastretro-cli/internal/protocol"
)

func testModel() Model {
	return Model{
		takenIDs: make(map[string]bool),
		width:    80,
		height:   24,
	}
}

func testModelWithState() Model {
	m := testModel()
	m.state = &protocol.RetroState{
		Stage: "brainstorm",
		Participants: []protocol.Participant{
			{ID: "p1", Name: "Alice"},
			{ID: "p2", Name: "Bob"},
			{ID: "p3", Name: "Carol"},
		},
	}
	return m
}

func TestViewJoin_ShowsParticipants(t *testing.T) {
	m := testModelWithState()
	m.peerCount = 2

	view := m.viewJoin()

	if !strings.Contains(view, "Alice") {
		t.Error("expected view to contain 'Alice'")
	}
	if !strings.Contains(view, "Bob") {
		t.Error("expected view to contain 'Bob'")
	}
	if !strings.Contains(view, "Carol") {
		t.Error("expected view to contain 'Carol'")
	}
	if !strings.Contains(view, "2 peers") {
		t.Error("expected view to contain peer count")
	}
}

func TestViewJoin_TakenParticipant(t *testing.T) {
	m := testModelWithState()
	m.takenIDs["p3"] = true

	view := m.viewJoin()

	if !strings.Contains(view, "taken") {
		t.Error("expected view to show 'taken' for Carol")
	}
}

func TestViewJoin_InputMode(t *testing.T) {
	m := testModelWithState()
	m.inputMode = true
	m.inputText = "Dave"

	view := m.viewJoin()

	if !strings.Contains(view, "Dave") {
		t.Error("expected view to show input text")
	}
}

func TestViewJoin_NoState(t *testing.T) {
	m := testModel()
	view := m.viewJoin()

	if !strings.Contains(view, "Who are you?") {
		t.Error("expected view to show prompt even without state")
	}
}

func TestViewJoin_CursorNavigation(t *testing.T) {
	m := testModelWithState()
	m.cursor = 1

	view := m.viewJoin()

	lines := strings.Split(view, "\n")
	foundCursor := false
	for _, line := range lines {
		if strings.Contains(line, "> Bob") || strings.Contains(line, ">") && strings.Contains(line, "Bob") {
			foundCursor = true
		}
	}
	if !foundCursor {
		t.Error("expected cursor on Bob (index 1)")
	}
}
