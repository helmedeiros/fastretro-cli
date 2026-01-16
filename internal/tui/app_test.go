package tui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
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

func TestHandleWS_RequestState(t *testing.T) {
	m := testBrainstormModel()
	// No client, so request-state is a no-op (doesn't panic)
	msg := protocol.IncomingMessage{Type: "request-state"}
	result, _ := m.handleWS(msg)
	model := result.(Model)
	// State should be unchanged
	if model.state.Stage != "brainstorm" {
		t.Error("state should be unchanged")
	}
}

func TestBuildDiscussState_SortsByVotes(t *testing.T) {
	m := testModel()
	m.state = &protocol.RetroState{
		Cards: []protocol.Card{
			{ID: "c1", ColumnID: "stop", Text: "low votes"},
			{ID: "c2", ColumnID: "stop", Text: "high votes"},
		},
		Groups: []protocol.Group{},
		Votes: []protocol.Vote{
			{ParticipantID: "p1", CardID: "c2"},
			{ParticipantID: "p2", CardID: "c2"},
			{ParticipantID: "p1", CardID: "c1"},
		},
	}

	ds := m.buildDiscussState()

	if len(ds.Order) != 2 {
		t.Fatalf("expected 2 items, got %d", len(ds.Order))
	}
	if ds.Order[0] != "c2" {
		t.Errorf("expected c2 first (most votes), got %q", ds.Order[0])
	}
	if ds.Segment != "context" {
		t.Errorf("expected context segment, got %q", ds.Segment)
	}
}

func TestBuildDiscussState_GroupsAndCards(t *testing.T) {
	m := testModel()
	m.state = &protocol.RetroState{
		Cards: []protocol.Card{
			{ID: "c1", ColumnID: "stop", Text: "in group"},
			{ID: "c2", ColumnID: "stop", Text: "ungrouped"},
		},
		Groups: []protocol.Group{
			{ID: "g1", ColumnID: "stop", Name: "Group", CardIDs: []string{"c1"}},
		},
	}

	ds := m.buildDiscussState()

	if len(ds.Order) != 2 {
		t.Fatalf("expected 2 (1 group + 1 card), got %d", len(ds.Order))
	}
	ids := make(map[string]bool)
	for _, id := range ds.Order {
		ids[id] = true
	}
	if !ids["g1"] || !ids["c2"] {
		t.Errorf("expected g1 and c2, got %v", ds.Order)
	}
}

func TestInitStage_Discuss(t *testing.T) {
	m := testModel()
	m.state = &protocol.RetroState{
		Stage: "discuss",
		Cards: []protocol.Card{{ID: "c1", ColumnID: "stop", Text: "item"}},
	}

	m.initStage()

	if m.state.Discuss == nil {
		t.Fatal("expected discuss state to be initialized")
	}
	if len(m.state.Discuss.Order) != 1 {
		t.Errorf("expected 1 item, got %d", len(m.state.Discuss.Order))
	}
}

func TestInitStage_DiscussAlreadySet(t *testing.T) {
	m := testModel()
	existing := &protocol.DiscussState{Order: []string{"x"}, CurrentIndex: 2}
	m.state = &protocol.RetroState{
		Stage:   "discuss",
		Discuss: existing,
	}

	m.initStage()

	// Should not overwrite existing
	if m.state.Discuss.CurrentIndex != 2 {
		t.Error("should not overwrite existing discuss state")
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

func TestHandleWS_StateAutoSelectsDefaultMember(t *testing.T) {
	m := testModel()
	m.defaultMemberName = "Alice"

	msg := protocol.IncomingMessage{
		Type: "state",
		State: &protocol.RetroState{
			Stage: "brainstorm",
			Participants: []protocol.Participant{
				{ID: "p1", Name: "Alice"},
				{ID: "p2", Name: "Bob"},
			},
		},
	}

	result, _ := m.handleWS(msg)
	model := result.(Model)

	if model.participantID != "p1" {
		t.Errorf("expected auto-select p1 (Alice), got %q", model.participantID)
	}
}

func TestHandleWS_StateAutoSelectCaseInsensitive(t *testing.T) {
	m := testModel()
	m.defaultMemberName = "alice"

	msg := protocol.IncomingMessage{
		Type: "state",
		State: &protocol.RetroState{
			Stage: "brainstorm",
			Participants: []protocol.Participant{
				{ID: "p1", Name: "Alice"},
			},
		},
	}

	result, _ := m.handleWS(msg)
	model := result.(Model)

	if model.participantID != "p1" {
		t.Errorf("expected auto-select p1, got %q", model.participantID)
	}
}

func TestHandleWS_StateNoAutoSelectWhenAlreadyPicked(t *testing.T) {
	m := testModel()
	m.defaultMemberName = "Bob"
	m.participantID = "p1" // already picked Alice

	msg := protocol.IncomingMessage{
		Type: "state",
		State: &protocol.RetroState{
			Stage: "brainstorm",
			Participants: []protocol.Participant{
				{ID: "p1", Name: "Alice"},
				{ID: "p2", Name: "Bob"},
			},
		},
	}

	result, _ := m.handleWS(msg)
	model := result.(Model)

	if model.participantID != "p1" {
		t.Errorf("should keep existing pick p1, got %q", model.participantID)
	}
}

func TestHandleWS_StateNoAutoSelectWhenTaken(t *testing.T) {
	m := testModel()
	m.defaultMemberName = "Alice"
	m.takenIDs["p1"] = true

	msg := protocol.IncomingMessage{
		Type: "state",
		State: &protocol.RetroState{
			Stage: "brainstorm",
			Participants: []protocol.Participant{
				{ID: "p1", Name: "Alice"},
			},
		},
	}

	result, _ := m.handleWS(msg)
	model := result.(Model)

	if model.participantID != "" {
		t.Errorf("should not auto-select taken participant, got %q", model.participantID)
	}
}

func TestHandleWS_StateNoAutoSelectNoMatch(t *testing.T) {
	m := testModel()
	m.defaultMemberName = "Charlie"

	msg := protocol.IncomingMessage{
		Type: "state",
		State: &protocol.RetroState{
			Stage: "brainstorm",
			Participants: []protocol.Participant{
				{ID: "p1", Name: "Alice"},
			},
		},
	}

	result, _ := m.handleWS(msg)
	model := result.(Model)

	if model.participantID != "" {
		t.Errorf("should not auto-select when no match, got %q", model.participantID)
	}
}

func TestRenderStageBar_AllStagesPresent(t *testing.T) {
	m := testModelWithState()
	m.participantID = "p1"
	m.state.Stage = "brainstorm"

	bar := m.renderStageBar()

	for _, stage := range []string{"ICEBREAKER", "BRAINSTORM", "GROUP", "VOTE", "DISCUSS", "REVIEW", "CLOSE"} {
		if !strings.Contains(bar, stage) {
			t.Errorf("expected %s in stage bar", stage)
		}
	}
}

func TestStageIndex(t *testing.T) {
	tests := []struct {
		stage string
		want  int
	}{
		{"icebreaker", 0},
		{"brainstorm", 1},
		{"close", 6},
		{"unknown", -1},
	}
	for _, tt := range tests {
		got := stageIndex(tt.stage)
		if got != tt.want {
			t.Errorf("stageIndex(%q) = %d, want %d", tt.stage, got, tt.want)
		}
	}
}

func TestJoinColumnsEqualHeight(t *testing.T) {
	contents := []string{"A\nB", "C"}
	colStyles := []lipgloss.Style{
		lipgloss.NewStyle().Width(10),
		lipgloss.NewStyle().Width(10),
	}
	result := joinColumnsEqualHeight(contents, colStyles)
	if result == "" {
		t.Error("expected non-empty result")
	}
}

func TestSetAndGetParticipantID(t *testing.T) {
	m := NewModel(nil)
	if m.ParticipantID() != "" {
		t.Errorf("expected empty, got %q", m.ParticipantID())
	}
	m.SetParticipantID("p-1")
	if m.ParticipantID() != "p-1" {
		t.Errorf("expected p-1, got %q", m.ParticipantID())
	}
}
