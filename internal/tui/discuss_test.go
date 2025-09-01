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
			{ID: "c2", ColumnID: "stop", Text: "too many bugs"},
		},
		Groups: []protocol.Group{
			{ID: "g1", ColumnID: "stop", Name: "Process", CardIDs: []string{"c1", "c2"}},
		},
		Votes: []protocol.Vote{
			{ParticipantID: "p1", CardID: "g1"},
			{ParticipantID: "p2", CardID: "g1"},
		},
		Discuss: &protocol.DiscussState{
			Order:        []string{"g1"},
			CurrentIndex: 0,
			Segment:      "context",
		},
		DiscussNotes: []protocol.DiscussNote{
			{ID: "n1", ParentCardID: "g1", Lane: "context", Text: "Need clarity on process"},
			{ID: "n2", ParentCardID: "g1", Lane: "context", Text: "Second context note"},
			{ID: "n3", ParentCardID: "g1", Lane: "actions", Text: "Write runbook"},
		},
	}
	return m
}

func testDiscussModelMulti() Model {
	m := testDiscussModel()
	m.state.Cards = append(m.state.Cards, protocol.Card{ID: "c3", ColumnID: "start", Text: "pair programming"})
	m.state.Discuss.Order = []string{"g1", "c3"}
	return m
}

// --- viewDiscuss ---

func TestViewDiscuss_ShowsDots(t *testing.T) {
	m := testDiscussModelMulti()
	view := m.viewDiscuss()

	if !strings.Contains(view, "●") {
		t.Error("expected filled dot for current item")
	}
	if !strings.Contains(view, "○") {
		t.Error("expected empty dot for other item")
	}
}

func TestViewDiscuss_ShowsVoteCount(t *testing.T) {
	m := testDiscussModel()
	view := m.viewDiscuss()

	if !strings.Contains(view, "Votes: 2") {
		t.Error("expected vote count")
	}
}

func TestViewDiscuss_ShowsSubcards(t *testing.T) {
	m := testDiscussModel()
	view := m.viewDiscuss()

	if !strings.Contains(view, "long meetings") {
		t.Error("expected subcard text for group")
	}
}

func TestViewDiscuss_ShowsPrevNext(t *testing.T) {
	m := testDiscussModelMulti()
	view := m.viewDiscuss()

	if !strings.Contains(view, "Prev") {
		t.Error("expected Prev")
	}
	if !strings.Contains(view, "Next") {
		t.Error("expected Next")
	}
}

func TestViewDiscuss_SideBySideLanes(t *testing.T) {
	m := testDiscussModel()
	view := m.viewDiscuss()

	if !strings.Contains(view, "CONTEXT") {
		t.Error("expected CONTEXT column")
	}
	if !strings.Contains(view, "ACTIONS") {
		t.Error("expected ACTIONS column")
	}
}

func TestViewDiscuss_ShowsContextNotes(t *testing.T) {
	m := testDiscussModel()
	view := m.viewDiscuss()

	if !strings.Contains(view, "Need clarity") {
		t.Error("expected context note")
	}
}

func TestViewDiscuss_ShowsHelp(t *testing.T) {
	m := testDiscussModel()
	view := m.viewDiscuss()

	if !strings.Contains(view, "prev/next") {
		t.Error("expected prev/next hint")
	}
	if !strings.Contains(view, "lane") {
		t.Error("expected lane hint")
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
}

func TestViewDiscuss_NilState(t *testing.T) {
	m := testModel()
	if view := m.viewDiscuss(); !strings.Contains(view, "Waiting") {
		t.Error("expected waiting")
	}
}

func TestViewDiscuss_NoDiscussState(t *testing.T) {
	m := testModel()
	m.state = &protocol.RetroState{Stage: "discuss"}
	if view := m.viewDiscuss(); !strings.Contains(view, "Waiting") {
		t.Error("expected waiting")
	}
}

func TestViewDiscuss_Complete(t *testing.T) {
	m := testDiscussModel()
	m.state.Discuss.CurrentIndex = 1
	if view := m.viewDiscuss(); !strings.Contains(view, "complete") {
		t.Error("expected complete")
	}
}

func TestViewDiscuss_WithTimer(t *testing.T) {
	m := testDiscussModel()
	m.state.Timer = &protocol.Timer{Status: "running", RemainingMs: 150000}
	if view := m.viewDiscuss(); !strings.Contains(view, "2:30") {
		t.Error("expected timer")
	}
}

func TestViewDiscuss_EmptyNotes(t *testing.T) {
	m := testDiscussModel()
	m.state.DiscussNotes = nil
	view := m.viewDiscuss()

	if !strings.Contains(view, "empty") {
		t.Error("expected empty indicator")
	}
}

// --- handleDiscussKeys navigation ---

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
	m.cursor = 0

	result, _ := m.handleDiscussKeys(keyMsg("down"))
	model := result.(Model)

	if model.cursor != 1 {
		t.Errorf("expected 1, got %d", model.cursor)
	}
}

func TestHandleDiscussKeys_DownAtBottom(t *testing.T) {
	m := testDiscussModel()
	m.cursor = 1 // 2 context notes, max index 1

	result, _ := m.handleDiscussKeys(keyMsg("down"))
	model := result.(Model)

	if model.cursor != 1 {
		t.Errorf("should stay at 1, got %d", model.cursor)
	}
}

func TestHandleDiscussKeys_Tab(t *testing.T) {
	m := testDiscussModel()

	result, _ := m.handleDiscussKeys(keyMsg("tab"))
	model := result.(Model)

	if model.state.Discuss.Segment != "actions" {
		t.Errorf("expected actions, got %q", model.state.Discuss.Segment)
	}
	if model.cursor != 0 {
		t.Error("cursor should reset")
	}
}

func TestHandleDiscussKeys_TabBack(t *testing.T) {
	m := testDiscussModel()
	m.state.Discuss.Segment = "actions"

	result, _ := m.handleDiscussKeys(keyMsg("tab"))
	model := result.(Model)

	if model.state.Discuss.Segment != "context" {
		t.Errorf("expected context, got %q", model.state.Discuss.Segment)
	}
}

func TestHandleDiscussKeys_Left(t *testing.T) {
	m := testDiscussModel()
	m.state.Discuss.Segment = "actions"

	result, _ := m.handleDiscussKeys(keyMsg("left"))
	model := result.(Model)

	if model.state.Discuss.Segment != "context" {
		t.Error("expected context")
	}
}

func TestHandleDiscussKeys_Right(t *testing.T) {
	m := testDiscussModel()

	result, _ := m.handleDiscussKeys(keyMsg("right"))
	model := result.(Model)

	if model.state.Discuss.Segment != "actions" {
		t.Error("expected actions")
	}
}

func TestHandleDiscussKeys_Next(t *testing.T) {
	m := testDiscussModelMulti()

	result, _ := m.handleDiscussKeys(keyMsg("n"))
	model := result.(Model)

	if model.state.Discuss.CurrentIndex != 1 {
		t.Errorf("expected 1, got %d", model.state.Discuss.CurrentIndex)
	}
	if model.cursor != 0 {
		t.Error("cursor should reset")
	}
}

func TestHandleDiscussKeys_NextAtEnd(t *testing.T) {
	m := testDiscussModelMulti()
	m.state.Discuss.CurrentIndex = 1

	result, _ := m.handleDiscussKeys(keyMsg("n"))
	model := result.(Model)

	if model.state.Discuss.CurrentIndex != 1 {
		t.Error("should stay at end")
	}
}

func TestHandleDiscussKeys_Prev(t *testing.T) {
	m := testDiscussModelMulti()
	m.state.Discuss.CurrentIndex = 1

	result, _ := m.handleDiscussKeys(keyMsg("p"))
	model := result.(Model)

	if model.state.Discuss.CurrentIndex != 0 {
		t.Errorf("expected 0, got %d", model.state.Discuss.CurrentIndex)
	}
}

func TestHandleDiscussKeys_PrevAtStart(t *testing.T) {
	m := testDiscussModel()

	result, _ := m.handleDiscussKeys(keyMsg("p"))
	model := result.(Model)

	if model.state.Discuss.CurrentIndex != 0 {
		t.Error("should stay at start")
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

func TestHandleDiscussKeys_NoState(t *testing.T) {
	m := testModel()
	result, _ := m.handleDiscussKeys(keyMsg("a"))
	model := result.(Model)

	if model.inputMode {
		t.Error("should not activate with no state")
	}
}

func TestHandleDiscussKeys_DelegatesToInput(t *testing.T) {
	m := testDiscussModel()
	m.inputMode = true
	m.inputText = "x"

	result, _ := m.handleDiscussKeys(keyMsg("y"))
	model := result.(Model)

	if model.inputText != "xy" {
		t.Errorf("got %q", model.inputText)
	}
}

// --- handleDiscussInput ---

func TestHandleDiscussInput_Enter(t *testing.T) {
	m := testDiscussModel()
	m.inputMode = true
	m.inputText = "new note"

	initial := len(m.state.DiscussNotes)
	result, _ := m.handleDiscussInput(keyMsg("enter"))
	model := result.(Model)

	if model.inputMode {
		t.Error("should exit input")
	}
	if len(model.state.DiscussNotes) != initial+1 {
		t.Errorf("expected %d, got %d", initial+1, len(model.state.DiscussNotes))
	}
	last := model.state.DiscussNotes[len(model.state.DiscussNotes)-1]
	if last.Lane != "context" {
		t.Errorf("expected context, got %q", last.Lane)
	}
}

func TestHandleDiscussInput_EnterActions(t *testing.T) {
	m := testDiscussModel()
	m.state.Discuss.Segment = "actions"
	m.inputMode = true
	m.inputText = "action"

	result, _ := m.handleDiscussInput(keyMsg("enter"))
	model := result.(Model)

	last := model.state.DiscussNotes[len(model.state.DiscussNotes)-1]
	if last.Lane != "actions" {
		t.Errorf("expected actions, got %q", last.Lane)
	}
}

func TestHandleDiscussInput_EnterEmpty(t *testing.T) {
	m := testDiscussModel()
	m.inputMode = true
	m.inputText = "  "

	initial := len(m.state.DiscussNotes)
	result, _ := m.handleDiscussInput(keyMsg("enter"))
	model := result.(Model)

	if len(model.state.DiscussNotes) != initial {
		t.Error("empty should not add")
	}
}

func TestHandleDiscussInput_Escape(t *testing.T) {
	m := testDiscussModel()
	m.inputMode = true

	result, _ := m.handleDiscussInput(keyMsg("esc"))
	model := result.(Model)

	if model.inputMode {
		t.Error("should exit")
	}
}

func TestHandleDiscussInput_Backspace(t *testing.T) {
	m := testDiscussModel()
	m.inputMode = true
	m.inputText = "ab"

	result, _ := m.handleDiscussInput(keyMsg("backspace"))
	model := result.(Model)

	if model.inputText != "a" {
		t.Errorf("got %q", model.inputText)
	}
}

func TestHandleDiscussInput_Type(t *testing.T) {
	m := testDiscussModel()
	m.inputMode = true
	m.inputText = "a"

	result, _ := m.handleDiscussInput(keyMsg("b"))
	model := result.(Model)

	if model.inputText != "ab" {
		t.Errorf("got %q", model.inputText)
	}
}

// --- helpers ---

func TestLabelForItem_Group(t *testing.T) {
	m := testDiscussModel()
	if l := m.labelForItem("g1"); l != "Process" {
		t.Errorf("got %q", l)
	}
}

func TestLabelForItem_Card(t *testing.T) {
	m := testDiscussModel()
	if l := m.labelForItem("c1"); l != "long meetings" {
		t.Errorf("got %q", l)
	}
}

func TestLabelForItem_Unknown(t *testing.T) {
	m := testDiscussModel()
	if l := m.labelForItem("x"); l != "x" {
		t.Errorf("got %q", l)
	}
}

func TestLabelForItem_NilState(t *testing.T) {
	m := testModel()
	if l := m.labelForItem("c1"); l != "c1" {
		t.Errorf("got %q", l)
	}
}

func TestSubcardsForItem_Group(t *testing.T) {
	m := testDiscussModel()
	sc := m.subcardsForItem("g1")

	if len(sc) != 2 {
		t.Fatalf("expected 2, got %d", len(sc))
	}
	if sc[0] != "long meetings" {
		t.Errorf("got %q", sc[0])
	}
}

func TestSubcardsForItem_Card(t *testing.T) {
	m := testDiscussModel()
	sc := m.subcardsForItem("c1")

	if sc != nil {
		t.Errorf("expected nil for card, got %v", sc)
	}
}

func TestSubcardsForItem_NilState(t *testing.T) {
	m := testModel()
	if sc := m.subcardsForItem("g1"); sc != nil {
		t.Errorf("expected nil, got %v", sc)
	}
}

func TestNotesForItem_Context(t *testing.T) {
	m := testDiscussModel()
	if notes := m.notesForItem("g1", "context"); len(notes) != 2 {
		t.Errorf("expected 2, got %d", len(notes))
	}
}

func TestNotesForItem_Actions(t *testing.T) {
	m := testDiscussModel()
	if notes := m.notesForItem("g1", "actions"); len(notes) != 1 {
		t.Errorf("expected 1, got %d", len(notes))
	}
}

func TestNotesForItem_NilState(t *testing.T) {
	m := testModel()
	if notes := m.notesForItem("g1", "context"); notes != nil {
		t.Errorf("expected nil")
	}
}

func TestViewDiscuss_InAppView(t *testing.T) {
	m := testDiscussModel()
	if view := m.View(); !strings.Contains(view, "DISCUSS") {
		t.Error("expected DISCUSS in header")
	}
}

func TestHandleKey_DiscussStage(t *testing.T) {
	m := testDiscussModel()
	result, _ := m.handleKey(keyMsg("a"))
	model := result.(Model)
	if !model.inputMode {
		t.Error("expected discuss key handler")
	}
}
