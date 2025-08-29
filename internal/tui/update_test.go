package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/helmedeiros/fastretro-cli/internal/protocol"
)

func TestUpdate_KeyMsg(t *testing.T) {
	m := testBrainstormModel()

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	model := result.(Model)

	if !model.inputMode {
		t.Error("expected key message to be handled")
	}
}

func TestUpdate_WSMsg(t *testing.T) {
	m := testModel()
	state := &protocol.RetroState{Stage: "vote"}
	wsMsg := WSMsg(protocol.IncomingMessage{
		Type:  "state",
		State: state,
	})

	result, cmd := m.Update(wsMsg)
	model := result.(Model)

	if model.state == nil {
		t.Fatal("expected state to be set from WSMsg")
	}
	if model.state.Stage != "vote" {
		t.Errorf("expected stage 'vote', got %q", model.state.Stage)
	}
	// Should return a cmd to listen for next message
	if cmd == nil {
		t.Error("expected non-nil cmd to continue listening")
	}
}

func TestHandleVoteKeys_CursorAtTop(t *testing.T) {
	m := testVoteModel()
	m.cursor = 0

	result, _ := m.handleVoteKeys(keyMsg("up"))
	model := result.(Model)

	if model.cursor != 0 {
		t.Errorf("cursor should stay at 0, got %d", model.cursor)
	}
}

func TestHandleVoteKeys_CursorAtBottom(t *testing.T) {
	m := testVoteModel()
	items := m.voteItems()
	m.cursor = len(items) - 1

	result, _ := m.handleVoteKeys(keyMsg("down"))
	model := result.(Model)

	if model.cursor != len(items)-1 {
		t.Errorf("cursor should stay at bottom, got %d", model.cursor)
	}
}

func TestHandleJoinKeys_EnterNoState(t *testing.T) {
	m := testModel()
	m.state = &protocol.RetroState{} // empty participants

	result, _ := m.handleJoinKeys(keyMsg("enter"))
	model := result.(Model)

	if model.participantID != "" {
		t.Error("should not select when no participants")
	}
}

func TestHandleJoinKeys_DownNoState(t *testing.T) {
	m := testModel()
	m.state = nil

	// Should not panic
	result, _ := m.handleJoinKeys(keyMsg("down"))
	model := result.(Model)

	if model.cursor != 0 {
		t.Error("cursor should not change with nil state")
	}
}

func TestHandleJoinInput_MultiCharKey(t *testing.T) {
	m := testModelWithState()
	m.inputMode = true
	m.inputText = "x"

	// Multi-char key strings (like "tab") should not be appended
	result, _ := m.handleJoinInput(tea.KeyMsg{Type: tea.KeyTab})
	model := result.(Model)

	if model.inputText != "x" {
		t.Errorf("multi-char key should not append, got %q", model.inputText)
	}
}

func TestHandleBrainstormInput_MultiCharKey(t *testing.T) {
	m := testBrainstormModel()
	m.inputMode = true
	m.inputText = "x"

	result, _ := m.handleBrainstormInput(tea.KeyMsg{Type: tea.KeyTab})
	model := result.(Model)

	if model.inputText != "x" {
		t.Errorf("multi-char key should not append, got %q", model.inputText)
	}
}

func TestHandleBrainstormInput_BackspaceEmpty(t *testing.T) {
	m := testBrainstormModel()
	m.inputMode = true
	m.inputText = ""

	result, _ := m.handleBrainstormInput(keyMsg("backspace"))
	model := result.(Model)

	if model.inputText != "" {
		t.Errorf("expected empty, got %q", model.inputText)
	}
}
