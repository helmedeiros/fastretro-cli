package tui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/helmedeiros/fastretro-cli/internal/protocol"
)

func TestNewModel_Defaults(t *testing.T) {
	m := NewModel(nil)

	if m.participantID != "" {
		t.Errorf("expected empty participantID, got %q", m.participantID)
	}
	if m.takenIDs == nil {
		t.Error("expected takenIDs to be initialized")
	}
	if m.width != 80 {
		t.Errorf("expected width 80, got %d", m.width)
	}
	if m.height != 24 {
		t.Errorf("expected height 24, got %d", m.height)
	}
	if m.state != nil {
		t.Error("expected nil state")
	}
	if m.err != nil {
		t.Error("expected nil error")
	}
}

func TestView_Error(t *testing.T) {
	m := testModel()
	m.err = fmt.Errorf("connection lost")

	view := m.View()

	if !strings.Contains(view, "Error") {
		t.Error("expected 'Error' in view")
	}
	if !strings.Contains(view, "connection lost") {
		t.Error("expected error message in view")
	}
}

func TestView_WaitingForState(t *testing.T) {
	m := testModel()

	view := m.View()

	if !strings.Contains(view, "waiting") || !strings.Contains(view, "state") {
		t.Error("expected waiting message in view")
	}
}

func TestView_JoinScreen(t *testing.T) {
	m := testModelWithState()
	// participantID is empty → should show join screen
	view := m.View()

	if !strings.Contains(view, "Who are you?") {
		t.Error("expected join screen when participantID is empty")
	}
}

func TestView_BrainstormStage(t *testing.T) {
	m := testBrainstormModel()

	view := m.View()

	if !strings.Contains(view, "BRAINSTORM") {
		t.Error("expected 'BRAINSTORM' in header")
	}
}

func TestView_VoteStage(t *testing.T) {
	m := testVoteModel()

	view := m.View()

	if !strings.Contains(view, "VOTE") {
		t.Error("expected 'VOTE' in header")
	}
}

func TestView_DiscussStage(t *testing.T) {
	m := testDiscussModel()

	view := m.View()

	if !strings.Contains(view, "DISCUSS") {
		t.Error("expected 'DISCUSS' in header")
	}
}

func TestView_ReviewStage(t *testing.T) {
	m := testReviewModel()
	m.state.Stage = "review"

	view := m.View()

	if !strings.Contains(view, "REVIEW") {
		t.Error("expected 'REVIEW' in header")
	}
}

func TestView_CloseStage(t *testing.T) {
	m := testCloseModel()

	view := m.View()

	if !strings.Contains(view, "CLOSE") {
		t.Error("expected 'CLOSE' in header")
	}
}

func TestView_UnknownStage(t *testing.T) {
	m := testModel()
	m.participantID = "p1"
	m.state = &protocol.RetroState{Stage: "unknown"}

	view := m.View()

	if !strings.Contains(view, "view-only") {
		t.Error("expected 'view-only' for unknown stage")
	}
}

func TestView_GroupStageUsesBrainstormView(t *testing.T) {
	m := testBrainstormModel()
	m.state.Stage = "group"

	view := m.View()

	if !strings.Contains(view, "GROUP") {
		t.Error("expected 'GROUP' in header")
	}
}

func TestHandleWS_State(t *testing.T) {
	m := testModel()
	state := &protocol.RetroState{
		Stage: "brainstorm",
		Meta:  protocol.RetroMeta{Name: "Test"},
	}
	msg := protocol.IncomingMessage{
		Type:  "state",
		State: state,
	}

	result, _ := m.handleWS(msg)
	model := result.(Model)

	if model.state == nil {
		t.Fatal("expected state to be set")
	}
	if model.state.Stage != "brainstorm" {
		t.Errorf("expected stage 'brainstorm', got %q", model.state.Stage)
	}
}

func TestHandleWS_PeerCount(t *testing.T) {
	m := testModel()
	msg := protocol.IncomingMessage{
		Type:  "peer-count",
		Count: 5,
	}

	result, _ := m.handleWS(msg)
	model := result.(Model)

	if model.peerCount != 5 {
		t.Errorf("expected peerCount 5, got %d", model.peerCount)
	}
}

func TestHandleWS_TakenIDs(t *testing.T) {
	m := testModel()
	msg := protocol.IncomingMessage{
		Type: "taken-ids",
		IDs:  []string{"p1", "p2"},
	}

	result, _ := m.handleWS(msg)
	model := result.(Model)

	if !model.takenIDs["p1"] {
		t.Error("expected p1 to be taken")
	}
	if !model.takenIDs["p2"] {
		t.Error("expected p2 to be taken")
	}
	if model.takenIDs["p3"] {
		t.Error("p3 should not be taken")
	}
}

func TestHandleWS_NavigateStage(t *testing.T) {
	m := testModelWithState()
	msg := protocol.IncomingMessage{
		Type:  "navigate-stage",
		Stage: "vote",
	}

	result, _ := m.handleWS(msg)
	model := result.(Model)

	if model.state.Stage != "vote" {
		t.Errorf("expected stage 'vote', got %q", model.state.Stage)
	}
}

func TestHandleWS_NavigateStage_NoState(t *testing.T) {
	m := testModel()
	msg := protocol.IncomingMessage{
		Type:  "navigate-stage",
		Stage: "vote",
	}

	result, _ := m.handleWS(msg)
	model := result.(Model)

	if model.state != nil {
		t.Error("state should remain nil when navigating without existing state")
	}
}

func TestRenderStageBar(t *testing.T) {
	m := testBrainstormModel()
	bar := m.renderStageBar()

	if !strings.Contains(bar, "BRAINSTORM") {
		t.Error("expected BRAINSTORM in stage bar")
	}
	if !strings.Contains(bar, "VOTE") {
		t.Error("expected other stages in bar")
	}
}

func TestRenderStageBar_NilState(t *testing.T) {
	m := testModel()
	bar := m.renderStageBar()

	if bar != "" {
		t.Errorf("expected empty for nil state, got %q", bar)
	}
}

func TestView_WithMetaName(t *testing.T) {
	m := testModel()
	m.participantID = "p1"
	m.state = &protocol.RetroState{
		Stage: "brainstorm",
		Meta:  protocol.RetroMeta{Name: "Sprint 42", Date: "2025-09-07"},
		Cards: []protocol.Card{{ID: "c1", ColumnID: "stop", Text: "x"}},
	}
	view := m.View()

	if !strings.Contains(view, "Sprint 42") {
		t.Error("expected retro name in header")
	}
	if !strings.Contains(view, "2025-09-07") {
		t.Error("expected date in header")
	}
}

func TestView_FallbackTitle(t *testing.T) {
	m := testModel()
	m.participantID = "p1"
	m.state = &protocol.RetroState{Stage: "brainstorm"}
	view := m.View()

	if !strings.Contains(view, "fastRetro CLI") {
		t.Error("expected fallback title")
	}
}

func TestHandleKey_DiscussStage_InKeys(t *testing.T) {
	m := testDiscussModel()

	result, _ := m.handleKey(keyMsg("a"))
	model := result.(Model)

	if !model.inputMode {
		t.Error("expected discuss handler")
	}
}

func TestHandleKey_ReviewStage_InKeys(t *testing.T) {
	m := testReviewModel()

	result, _ := m.handleKey(keyMsg("o"))
	model := result.(Model)

	if !model.inputMode {
		t.Error("expected review handler")
	}
}

func TestHandleKey_GroupStage(t *testing.T) {
	m := testGroupModel()

	result, _ := m.handleKey(keyMsg(":"))
	// In group stage, ":" no longer enters command mode since redesign
	// but it shouldn't panic
	_ = result
}
