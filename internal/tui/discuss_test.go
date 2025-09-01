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

func testDiscussModelMultiItems() Model {
	m := testDiscussModel()
	m.state.Discuss.Order = []string{"g1", "c1"}
	return m
}

// --- viewDiscuss ---

func TestViewDiscuss_ShowsCurrentItem(t *testing.T) {
	m := testDiscussModel()
	view := m.viewDiscuss()

	if !strings.Contains(view, "Process") {
		t.Error("expected current item label")
	}
	if !strings.Contains(view, "1 of 1") {
		t.Error("expected progress")
	}
}

func TestViewDiscuss_ShowsCarousel(t *testing.T) {
	m := testDiscussModelMultiItems()
	view := m.viewDiscuss()

	if !strings.Contains(view, "Process") {
		t.Error("expected first carousel item")
	}
}

func TestViewDiscuss_ShowsContextNotes(t *testing.T) {
	m := testDiscussModel()
	view := m.viewDiscuss()

	if !strings.Contains(view, "Need clarity") {
		t.Error("expected context note")
	}
}

func TestViewDiscuss_ShowsSegmentTabs(t *testing.T) {
	m := testDiscussModel()
	view := m.viewDiscuss()

	if !strings.Contains(view, "context") {
		t.Error("expected context tab")
	}
	if !strings.Contains(view, "actions") {
		t.Error("expected actions tab")
	}
}

func TestViewDiscuss_ShowsHelp(t *testing.T) {
	m := testDiscussModel()
	view := m.viewDiscuss()

	if !strings.Contains(view, "notes") {
		t.Error("expected navigation hint")
	}
	if !strings.Contains(view, "segment") {
		t.Error("expected segment hint")
	}
	if !strings.Contains(view, "add note") {
		t.Error("expected add note hint")
	}
}

func TestViewDiscuss_InputMode(t *testing.T) {
	m := testDiscussModel()
	m.inputMode = true
	m.inputText = "my note"
	view := m.viewDiscuss()

	if !strings.Contains(view, "my note") {
		t.Error("expected input text")
	}
	if !strings.Contains(view, "Add note") {
		t.Error("expected add note prompt")
	}
}

func TestViewDiscuss_NoDiscussState(t *testing.T) {
	m := testModel()
	m.state = &protocol.RetroState{Stage: "discuss"}
	view := m.viewDiscuss()

	if !strings.Contains(view, "Waiting") {
		t.Error("expected waiting message")
	}
}

func TestViewDiscuss_NilState(t *testing.T) {
	m := testModel()
	view := m.viewDiscuss()

	if !strings.Contains(view, "Waiting") {
		t.Error("expected waiting message")
	}
}

func TestViewDiscuss_Complete(t *testing.T) {
	m := testDiscussModel()
	m.state.Discuss.CurrentIndex = 1

	view := m.viewDiscuss()

	if !strings.Contains(view, "complete") {
		t.Error("expected complete message")
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
		t.Error("expected timer")
	}
}

func TestViewDiscuss_DefaultSegment(t *testing.T) {
	m := testDiscussModel()
	m.state.Discuss.Segment = ""
	view := m.viewDiscuss()

	if !strings.Contains(view, "context") {
		t.Error("expected context as default")
	}
}

func TestViewDiscuss_NoNotes(t *testing.T) {
	m := testDiscussModel()
	m.state.DiscussNotes = nil
	view := m.viewDiscuss()

	if !strings.Contains(view, "No notes") {
		t.Error("expected no notes message")
	}
}

func TestViewDiscuss_CursorOnNote(t *testing.T) {
	m := testDiscussModel()
	m.cursor = 0
	view := m.viewDiscuss()

	if !strings.Contains(view, ">") {
		t.Error("expected cursor on note")
	}
}

// --- handleDiscussKeys ---

func TestHandleDiscussKeys_Up(t *testing.T) {
	m := testDiscussModel()
	m.cursor = 1

	result, _ := m.handleDiscussKeys(keyMsg("up"))
	model := result.(Model)

	if model.cursor != 0 {
		t.Errorf("expected 0, got %d", model.cursor)
	}
}

func TestHandleDiscussKeys_Down(t *testing.T) {
	m := testDiscussModel()
	// context has 1 note
	m.cursor = 0

	result, _ := m.handleDiscussKeys(keyMsg("down"))
	model := result.(Model)

	// Only 1 context note, can't go further
	if model.cursor != 0 {
		t.Errorf("expected 0 (only 1 note), got %d", model.cursor)
	}
}

func TestHandleDiscussKeys_UpAtTop(t *testing.T) {
	m := testDiscussModel()
	m.cursor = 0

	result, _ := m.handleDiscussKeys(keyMsg("up"))
	model := result.(Model)

	if model.cursor != 0 {
		t.Errorf("should stay at 0")
	}
}

func TestHandleDiscussKeys_SwitchToActions(t *testing.T) {
	m := testDiscussModel()

	result, _ := m.handleDiscussKeys(keyMsg("right"))
	model := result.(Model)

	if model.state.Discuss.Segment != "actions" {
		t.Errorf("expected actions, got %q", model.state.Discuss.Segment)
	}
	if model.cursor != 0 {
		t.Error("cursor should reset on segment switch")
	}
}

func TestHandleDiscussKeys_SwitchToContext(t *testing.T) {
	m := testDiscussModel()
	m.state.Discuss.Segment = "actions"

	result, _ := m.handleDiscussKeys(keyMsg("left"))
	model := result.(Model)

	if model.state.Discuss.Segment != "context" {
		t.Errorf("expected context, got %q", model.state.Discuss.Segment)
	}
}

func TestHandleDiscussKeys_LeftOnContext(t *testing.T) {
	m := testDiscussModel()
	m.state.Discuss.Segment = "context"

	result, _ := m.handleDiscussKeys(keyMsg("left"))
	model := result.(Model)

	if model.state.Discuss.Segment != "context" {
		t.Error("should stay on context")
	}
}

func TestHandleDiscussKeys_RightOnActions(t *testing.T) {
	m := testDiscussModel()
	m.state.Discuss.Segment = "actions"

	result, _ := m.handleDiscussKeys(keyMsg("right"))
	model := result.(Model)

	if model.state.Discuss.Segment != "actions" {
		t.Error("should stay on actions")
	}
}

func TestHandleDiscussKeys_AddNote(t *testing.T) {
	m := testDiscussModel()

	result, _ := m.handleDiscussKeys(keyMsg("a"))
	model := result.(Model)

	if !model.inputMode {
		t.Error("expected input mode")
	}
}

func TestHandleDiscussKeys_VimK(t *testing.T) {
	m := testDiscussModel()
	m.cursor = 1

	result, _ := m.handleDiscussKeys(keyMsg("k"))
	model := result.(Model)

	if model.cursor != 0 {
		t.Errorf("expected 0, got %d", model.cursor)
	}
}

func TestHandleDiscussKeys_VimL(t *testing.T) {
	m := testDiscussModel()

	result, _ := m.handleDiscussKeys(keyMsg("l"))
	model := result.(Model)

	if model.state.Discuss.Segment != "actions" {
		t.Error("l should switch to actions")
	}
}

func TestHandleDiscussKeys_VimH(t *testing.T) {
	m := testDiscussModel()
	m.state.Discuss.Segment = "actions"

	result, _ := m.handleDiscussKeys(keyMsg("h"))
	model := result.(Model)

	if model.state.Discuss.Segment != "context" {
		t.Error("h should switch to context")
	}
}

func TestHandleDiscussKeys_NoState(t *testing.T) {
	m := testModel()
	result, _ := m.handleDiscussKeys(keyMsg("a"))
	model := result.(Model)

	if model.inputMode {
		t.Error("should not activate input with no state")
	}
}

func TestHandleDiscussKeys_DelegatesToInput(t *testing.T) {
	m := testDiscussModel()
	m.inputMode = true
	m.inputText = "x"

	result, _ := m.handleDiscussKeys(keyMsg("y"))
	model := result.(Model)

	if model.inputText != "xy" {
		t.Errorf("expected 'xy', got %q", model.inputText)
	}
}

// --- handleDiscussInput ---

func TestHandleDiscussInput_Enter(t *testing.T) {
	m := testDiscussModel()
	m.inputMode = true
	m.inputText = "new note"

	initialNotes := len(m.state.DiscussNotes)
	result, _ := m.handleDiscussInput(keyMsg("enter"))
	model := result.(Model)

	if model.inputMode {
		t.Error("should exit input mode")
	}
	if len(model.state.DiscussNotes) != initialNotes+1 {
		t.Errorf("expected %d notes, got %d", initialNotes+1, len(model.state.DiscussNotes))
	}
	last := model.state.DiscussNotes[len(model.state.DiscussNotes)-1]
	if last.Text != "new note" {
		t.Errorf("expected 'new note', got %q", last.Text)
	}
	if last.Lane != "context" {
		t.Errorf("expected lane 'context', got %q", last.Lane)
	}
	if last.ParentCardID != "g1" {
		t.Errorf("expected parent 'g1', got %q", last.ParentCardID)
	}
}

func TestHandleDiscussInput_EnterOnActions(t *testing.T) {
	m := testDiscussModel()
	m.state.Discuss.Segment = "actions"
	m.inputMode = true
	m.inputText = "action item"

	result, _ := m.handleDiscussInput(keyMsg("enter"))
	model := result.(Model)

	last := model.state.DiscussNotes[len(model.state.DiscussNotes)-1]
	if last.Lane != "actions" {
		t.Errorf("expected lane 'actions', got %q", last.Lane)
	}
}

func TestHandleDiscussInput_EnterEmpty(t *testing.T) {
	m := testDiscussModel()
	m.inputMode = true
	m.inputText = "  "

	initialNotes := len(m.state.DiscussNotes)
	result, _ := m.handleDiscussInput(keyMsg("enter"))
	model := result.(Model)

	if len(model.state.DiscussNotes) != initialNotes {
		t.Error("empty note should not be added")
	}
}

func TestHandleDiscussInput_Escape(t *testing.T) {
	m := testDiscussModel()
	m.inputMode = true
	m.inputText = "partial"

	result, _ := m.handleDiscussInput(keyMsg("esc"))
	model := result.(Model)

	if model.inputMode {
		t.Error("should exit input mode")
	}
}

func TestHandleDiscussInput_Backspace(t *testing.T) {
	m := testDiscussModel()
	m.inputMode = true
	m.inputText = "abc"

	result, _ := m.handleDiscussInput(keyMsg("backspace"))
	model := result.(Model)

	if model.inputText != "ab" {
		t.Errorf("expected 'ab', got %q", model.inputText)
	}
}

func TestHandleDiscussInput_TypeChar(t *testing.T) {
	m := testDiscussModel()
	m.inputMode = true
	m.inputText = "a"

	result, _ := m.handleDiscussInput(keyMsg("b"))
	model := result.(Model)

	if model.inputText != "ab" {
		t.Errorf("expected 'ab', got %q", model.inputText)
	}
}

// --- helpers ---

func TestLabelForItem_Group(t *testing.T) {
	m := testDiscussModel()
	if label := m.labelForItem("g1"); label != "Process" {
		t.Errorf("expected 'Process', got %q", label)
	}
}

func TestLabelForItem_Card(t *testing.T) {
	m := testDiscussModel()
	if label := m.labelForItem("c1"); label != "long meetings" {
		t.Errorf("expected 'long meetings', got %q", label)
	}
}

func TestLabelForItem_Unknown(t *testing.T) {
	m := testDiscussModel()
	if label := m.labelForItem("unknown"); label != "unknown" {
		t.Errorf("expected 'unknown', got %q", label)
	}
}

func TestLabelForItem_NilState(t *testing.T) {
	m := testModel()
	if label := m.labelForItem("c1"); label != "c1" {
		t.Errorf("expected 'c1', got %q", label)
	}
}

func TestNotesForItem_Context(t *testing.T) {
	m := testDiscussModel()
	notes := m.notesForItem("g1", "context")
	if len(notes) != 1 {
		t.Fatalf("expected 1, got %d", len(notes))
	}
}

func TestNotesForItem_Actions(t *testing.T) {
	m := testDiscussModel()
	notes := m.notesForItem("g1", "actions")
	if len(notes) != 1 {
		t.Fatalf("expected 1, got %d", len(notes))
	}
}

func TestNotesForItem_Empty(t *testing.T) {
	m := testDiscussModel()
	notes := m.notesForItem("nonexistent", "context")
	if len(notes) != 0 {
		t.Errorf("expected 0, got %d", len(notes))
	}
}

func TestNotesForItem_NilState(t *testing.T) {
	m := testModel()
	if notes := m.notesForItem("g1", "context"); notes != nil {
		t.Errorf("expected nil, got %v", notes)
	}
}

func TestViewDiscuss_InAppView(t *testing.T) {
	m := testDiscussModel()
	view := m.View()

	if !strings.Contains(view, "DISCUSS") {
		t.Error("expected DISCUSS in header")
	}
}

func TestHandleKey_DiscussStage(t *testing.T) {
	m := testDiscussModel()

	result, _ := m.handleKey(keyMsg("a"))
	model := result.(Model)

	if !model.inputMode {
		t.Error("expected discuss key handler to activate")
	}
}
