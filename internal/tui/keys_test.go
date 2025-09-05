package tui

import (
	"fmt"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/helmedeiros/fastretro-cli/internal/protocol"
)

func keyMsg(s string) tea.KeyMsg {
	switch s {
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	case "up":
		return tea.KeyMsg{Type: tea.KeyUp}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	case "tab":
		return tea.KeyMsg{Type: tea.KeyTab}
	case "shift+tab":
		return tea.KeyMsg{Type: tea.KeyShiftTab}
	case "left":
		return tea.KeyMsg{Type: tea.KeyLeft}
	case "right":
		return tea.KeyMsg{Type: tea.KeyRight}
	case "backspace":
		return tea.KeyMsg{Type: tea.KeyBackspace}
	case "ctrl+c":
		return tea.KeyMsg{Type: tea.KeyCtrlC}
	case " ":
		return tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}}
	default:
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
	}
}

// --- Join key handler tests ---

func TestHandleJoinKeys_CursorUp(t *testing.T) {
	m := testModelWithState()
	m.cursor = 2

	result, _ := m.handleJoinKeys(keyMsg("up"))
	model := result.(Model)

	if model.cursor != 1 {
		t.Errorf("cursor should be 1, got %d", model.cursor)
	}
}

func TestHandleJoinKeys_CursorDown(t *testing.T) {
	m := testModelWithState()
	m.cursor = 0

	result, _ := m.handleJoinKeys(keyMsg("down"))
	model := result.(Model)

	if model.cursor != 1 {
		t.Errorf("cursor should be 1, got %d", model.cursor)
	}
}

func TestHandleJoinKeys_CursorUpAtTop(t *testing.T) {
	m := testModelWithState()
	m.cursor = 0

	result, _ := m.handleJoinKeys(keyMsg("up"))
	model := result.(Model)

	if model.cursor != 0 {
		t.Errorf("cursor should stay at 0, got %d", model.cursor)
	}
}

func TestHandleJoinKeys_CursorDownAtBottom(t *testing.T) {
	m := testModelWithState()
	m.cursor = 2 // last participant

	result, _ := m.handleJoinKeys(keyMsg("down"))
	model := result.(Model)

	if model.cursor != 2 {
		t.Errorf("cursor should stay at 2, got %d", model.cursor)
	}
}

func TestHandleJoinKeys_CursorWithK(t *testing.T) {
	m := testModelWithState()
	m.cursor = 1

	result, _ := m.handleJoinKeys(keyMsg("k"))
	model := result.(Model)

	if model.cursor != 0 {
		t.Errorf("k should move cursor up, got %d", model.cursor)
	}
}

func TestHandleJoinKeys_CursorWithJ(t *testing.T) {
	m := testModelWithState()
	m.cursor = 0

	result, _ := m.handleJoinKeys(keyMsg("j"))
	model := result.(Model)

	if model.cursor != 1 {
		t.Errorf("j should move cursor down, got %d", model.cursor)
	}
}

func TestHandleJoinKeys_EnterSelectsParticipant(t *testing.T) {
	m := testModelWithState()
	m.cursor = 1 // Bob

	result, _ := m.handleJoinKeys(keyMsg("enter"))
	model := result.(Model)

	if model.participantID != "p2" {
		t.Errorf("expected participantID 'p2', got %q", model.participantID)
	}
}

func TestHandleJoinKeys_EnterSkipsTaken(t *testing.T) {
	m := testModelWithState()
	m.cursor = 1
	m.takenIDs["p2"] = true

	result, _ := m.handleJoinKeys(keyMsg("enter"))
	model := result.(Model)

	if model.participantID != "" {
		t.Errorf("should not select taken participant, got %q", model.participantID)
	}
}

func TestHandleJoinKeys_NToInputMode(t *testing.T) {
	m := testModelWithState()

	result, _ := m.handleJoinKeys(keyMsg("n"))
	model := result.(Model)

	if !model.inputMode {
		t.Error("expected input mode to be enabled")
	}
	if model.inputText != "" {
		t.Error("expected empty input text")
	}
}

func TestHandleJoinInput_Enter(t *testing.T) {
	m := testModelWithState()
	m.inputMode = true
	m.inputText = "Dave"

	result, _ := m.handleJoinInput(keyMsg("enter"))
	model := result.(Model)

	if model.participantID != "Dave" {
		t.Errorf("expected participantID 'Dave', got %q", model.participantID)
	}
	if model.inputMode {
		t.Error("input mode should be disabled")
	}
}

func TestHandleJoinInput_EnterEmpty(t *testing.T) {
	m := testModelWithState()
	m.inputMode = true
	m.inputText = "   "

	result, _ := m.handleJoinInput(keyMsg("enter"))
	model := result.(Model)

	if model.participantID != "" {
		t.Errorf("should not set ID for blank input, got %q", model.participantID)
	}
}

func TestHandleJoinInput_Escape(t *testing.T) {
	m := testModelWithState()
	m.inputMode = true
	m.inputText = "Dave"

	result, _ := m.handleJoinInput(keyMsg("esc"))
	model := result.(Model)

	if model.inputMode {
		t.Error("input mode should be disabled on escape")
	}
	if model.inputText != "" {
		t.Error("input text should be cleared on escape")
	}
}

func TestHandleJoinInput_Backspace(t *testing.T) {
	m := testModelWithState()
	m.inputMode = true
	m.inputText = "Dave"

	result, _ := m.handleJoinInput(keyMsg("backspace"))
	model := result.(Model)

	if model.inputText != "Dav" {
		t.Errorf("expected 'Dav' after backspace, got %q", model.inputText)
	}
}

func TestHandleJoinInput_BackspaceEmpty(t *testing.T) {
	m := testModelWithState()
	m.inputMode = true
	m.inputText = ""

	result, _ := m.handleJoinInput(keyMsg("backspace"))
	model := result.(Model)

	if model.inputText != "" {
		t.Errorf("expected empty after backspace on empty, got %q", model.inputText)
	}
}

func TestHandleJoinInput_TypeChar(t *testing.T) {
	m := testModelWithState()
	m.inputMode = true
	m.inputText = "Dav"

	result, _ := m.handleJoinInput(keyMsg("e"))
	model := result.(Model)

	if model.inputText != "Dave" {
		t.Errorf("expected 'Dave', got %q", model.inputText)
	}
}

func TestHandleJoinKeys_DelegatesToInput(t *testing.T) {
	m := testModelWithState()
	m.inputMode = true
	m.inputText = "x"

	result, _ := m.handleJoinKeys(keyMsg("a"))
	model := result.(Model)

	if model.inputText != "xa" {
		t.Errorf("expected 'xa' in input mode, got %q", model.inputText)
	}
}

// --- handleKey tests ---

func TestHandleKey_QuitNotInInputMode(t *testing.T) {
	m := testModel()
	m.state = &protocol.RetroState{Stage: "brainstorm"}
	m.participantID = "p1"

	_, cmd := m.handleKey(keyMsg("q"))

	if cmd == nil {
		t.Error("expected quit command")
	}
}

func TestHandleKey_QuitInInputMode(t *testing.T) {
	m := testModel()
	m.inputMode = true
	m.state = &protocol.RetroState{Stage: "brainstorm"}
	m.participantID = "p1"

	_, cmd := m.handleKey(keyMsg("q"))

	// Should not quit when in input mode - q is treated as text
	if cmd != nil {
		t.Error("should not quit when in input mode")
	}
}

func TestHandleKey_CtrlC(t *testing.T) {
	m := testModel()

	_, cmd := m.handleKey(keyMsg("ctrl+c"))

	if cmd == nil {
		t.Error("expected quit command for ctrl+c")
	}
}

func TestHandleKey_JoinWhenNoParticipant(t *testing.T) {
	m := testModelWithState()
	// participantID is empty, state exists → should delegate to join handler

	result, _ := m.handleKey(keyMsg("down"))
	model := result.(Model)

	if model.cursor != 1 {
		t.Errorf("expected join key handler to move cursor, got %d", model.cursor)
	}
}

func TestHandleKey_BrainstormStage(t *testing.T) {
	m := testBrainstormModel()

	result, _ := m.handleKey(keyMsg("a"))
	model := result.(Model)

	if !model.inputMode {
		t.Error("expected brainstorm key handler to activate input")
	}
}

func TestHandleKey_VoteStage(t *testing.T) {
	m := testVoteModel()
	m.cursor = 1

	result, _ := m.handleKey(keyMsg("up"))
	model := result.(Model)

	if model.cursor != 0 {
		t.Errorf("expected vote key handler to move cursor, got %d", model.cursor)
	}
}

func TestHandleKey_NoState(t *testing.T) {
	m := testModel()

	result, _ := m.handleKey(keyMsg("a"))
	model := result.(Model)

	// Should not panic, just return
	if model.participantID != "" {
		t.Error("should not change state when no state exists")
	}
}

// --- Update tests ---

func TestUpdate_WindowSize(t *testing.T) {
	m := testModel()
	msg := tea.WindowSizeMsg{Width: 120, Height: 40}

	result, _ := m.Update(msg)
	model := result.(Model)

	if model.width != 120 || model.height != 40 {
		t.Errorf("expected 120x40, got %dx%d", model.width, model.height)
	}
}

func TestUpdate_ErrMsg(t *testing.T) {
	m := testModel()
	msg := ErrMsg{Err: fmt.Errorf("test error")}

	result, _ := m.Update(msg)
	model := result.(Model)

	if model.err == nil || model.err.Error() != "test error" {
		t.Errorf("expected error 'test error', got %v", model.err)
	}
}

func TestUpdate_UnknownMsg(t *testing.T) {
	m := testModel()

	result, cmd := m.Update("unknown message type")
	model := result.(Model)

	if cmd != nil {
		t.Error("expected nil cmd for unknown message")
	}
	if model.err != nil {
		t.Error("should not set error for unknown message")
	}
}
