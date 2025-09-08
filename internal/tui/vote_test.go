package tui

import (
	"strings"
	"testing"

	"github.com/helmedeiros/fastretro-cli/internal/protocol"
)

func testVoteModel() Model {
	m := testModel()
	m.participantID = "p1"
	m.state = &protocol.RetroState{
		Stage:      "vote",
		VoteBudget: 3,
		Cards: []protocol.Card{
			{ID: "c1", ColumnID: "stop", Text: "long meetings"},
			{ID: "c2", ColumnID: "stop", Text: "too many bugs"},
			{ID: "c3", ColumnID: "start", Text: "pair programming"},
		},
		Groups: []protocol.Group{
			{ID: "g1", ColumnID: "stop", Name: "Process", CardIDs: []string{"c1", "c2"}},
		},
		Votes: []protocol.Vote{
			{ParticipantID: "p1", CardID: "g1"},
			{ParticipantID: "p2", CardID: "g1"},
			{ParticipantID: "p2", CardID: "c3"},
		},
	}
	return m
}

// --- columnVoteItems ---

func TestColumnVoteItems_Stop(t *testing.T) {
	m := testVoteModel()
	items := m.columnVoteItems("stop")

	// g1 (group), no ungrouped stop cards
	if len(items) != 1 {
		t.Fatalf("expected 1 item (group), got %d", len(items))
	}
	if items[0].id != "g1" {
		t.Errorf("expected g1, got %q", items[0].id)
	}
}

func TestColumnVoteItems_Start(t *testing.T) {
	m := testVoteModel()
	items := m.columnVoteItems("start")

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].id != "c3" {
		t.Errorf("expected c3, got %q", items[0].id)
	}
}

func TestColumnVoteItems_NilState(t *testing.T) {
	m := testModel()
	if items := m.columnVoteItems("stop"); items != nil {
		t.Errorf("expected nil")
	}
}

// --- viewVote ---

func TestViewVote_ShowsColumns(t *testing.T) {
	m := testVoteModel()
	view := m.viewVote()

	if !strings.Contains(view, "Stop") {
		t.Error("expected Stop column")
	}
	if !strings.Contains(view, "Start") {
		t.Error("expected Start column")
	}
}

func TestViewVote_ShowsItems(t *testing.T) {
	m := testVoteModel()
	view := m.viewVote()

	if !strings.Contains(view, "Process") {
		t.Error("expected group name")
	}
	if !strings.Contains(view, "pair programming") {
		t.Error("expected ungrouped card")
	}
}

func TestViewVote_ShowsVoteBudget(t *testing.T) {
	m := testVoteModel()
	view := m.viewVote()

	if !strings.Contains(view, "2/3") {
		t.Error("expected remaining votes")
	}
}

func TestViewVote_ShowsVoteCounts(t *testing.T) {
	m := testVoteModel()
	view := m.viewVote()

	if !strings.Contains(view, "+2") {
		t.Error("expected +2 votes on g1")
	}
}

func TestViewVote_ShowsHelp(t *testing.T) {
	m := testVoteModel()
	view := m.viewVote()

	if !strings.Contains(view, "column") {
		t.Error("expected column navigation hint")
	}
	if !strings.Contains(view, "vote") {
		t.Error("expected vote hint")
	}
}

func TestViewVote_ActiveColumn(t *testing.T) {
	m := testVoteModel()
	view := m.viewVote()

	if !strings.Contains(view, "▶") {
		t.Error("expected active column indicator")
	}
}

func TestViewVote_NilState(t *testing.T) {
	m := testModel()
	if view := m.viewVote(); view != "" {
		t.Errorf("expected empty")
	}
}

// --- handleVoteKeys navigation ---

func TestHandleVoteKeys_Up(t *testing.T) {
	m := testVoteModel()
	m.cursor = 1

	result, _ := m.handleVoteKeys(keyMsg("up"))
	model := result.(Model)

	if model.cursor != 0 {
		t.Errorf("expected 0, got %d", model.cursor)
	}
}

func TestHandleVoteKeys_Down(t *testing.T) {
	m := testVoteModel()
	// stop column has 1 item (g1), can't go down
	result, _ := m.handleVoteKeys(keyMsg("down"))
	model := result.(Model)

	if model.cursor != 0 {
		t.Errorf("expected 0 (only 1 item), got %d", model.cursor)
	}
}

func TestHandleVoteKeys_Tab(t *testing.T) {
	m := testVoteModel()

	result, _ := m.handleVoteKeys(keyMsg("tab"))
	model := result.(Model)

	if model.activeCol != 1 {
		t.Errorf("expected col 1, got %d", model.activeCol)
	}
	if model.cursor != 0 {
		t.Error("cursor should reset")
	}
}

func TestHandleVoteKeys_ShiftTab(t *testing.T) {
	m := testVoteModel()
	m.activeCol = 1

	result, _ := m.handleVoteKeys(keyMsg("shift+tab"))
	model := result.(Model)

	if model.activeCol != 0 {
		t.Errorf("expected col 0")
	}
}

func TestHandleVoteKeys_Right(t *testing.T) {
	m := testVoteModel()

	result, _ := m.handleVoteKeys(keyMsg("right"))
	model := result.(Model)

	if model.activeCol != 1 {
		t.Errorf("expected col 1")
	}
}

func TestHandleVoteKeys_Left(t *testing.T) {
	m := testVoteModel()
	m.activeCol = 1

	result, _ := m.handleVoteKeys(keyMsg("left"))
	model := result.(Model)

	if model.activeCol != 0 {
		t.Errorf("expected col 0")
	}
}

func TestHandleVoteKeys_VimK(t *testing.T) {
	m := testVoteModel()
	m.cursor = 1

	result, _ := m.handleVoteKeys(keyMsg("k"))
	model := result.(Model)

	if model.cursor != 0 {
		t.Errorf("expected 0")
	}
}

func TestHandleVoteKeys_VimJ(t *testing.T) {
	m := testVoteModel()
	// Switch to start column which has 1 item
	m.activeCol = 1
	result, _ := m.handleVoteKeys(keyMsg("j"))
	model := result.(Model)

	if model.cursor != 0 {
		t.Errorf("expected 0 (only 1 item)")
	}
}

// --- vote / unvote ---

func TestHandleVoteKeys_Vote(t *testing.T) {
	m := testVoteModel()
	initialVotes := len(m.state.Votes)

	result, _ := m.handleVoteKeys(keyMsg("enter"))
	model := result.(Model)

	if len(model.state.Votes) != initialVotes+1 {
		t.Errorf("expected %d votes, got %d", initialVotes+1, len(model.state.Votes))
	}
}

func TestHandleVoteKeys_VoteWithSpace(t *testing.T) {
	m := testVoteModel()
	initialVotes := len(m.state.Votes)

	result, _ := m.handleVoteKeys(keyMsg(" "))
	model := result.(Model)

	if len(model.state.Votes) != initialVotes+1 {
		t.Errorf("expected %d votes", initialVotes+1)
	}
}

func TestHandleVoteKeys_VoteNoBudget(t *testing.T) {
	m := testVoteModel()
	m.state.VoteBudget = 1 // already used 1
	initialVotes := len(m.state.Votes)

	result, _ := m.handleVoteKeys(keyMsg("enter"))
	model := result.(Model)

	if len(model.state.Votes) != initialVotes {
		t.Error("should not add vote when exhausted")
	}
}

func TestHandleVoteKeys_Unvote(t *testing.T) {
	m := testVoteModel()
	// cursor 0 = g1, p1 has voted for g1

	result, _ := m.handleVoteKeys(keyMsg("u"))
	model := result.(Model)

	if model.myVotesForItem("g1") != 0 {
		t.Error("vote should be removed")
	}
}

func TestHandleVoteKeys_UnvoteNoVote(t *testing.T) {
	m := testVoteModel()
	m.activeCol = 1 // start column, c3, p1 has NOT voted
	initialVotes := len(m.state.Votes)

	result, _ := m.handleVoteKeys(keyMsg("u"))
	model := result.(Model)

	if len(model.state.Votes) != initialVotes {
		t.Error("should not remove any vote")
	}
}

// --- helpers ---

func TestVotesForItem(t *testing.T) {
	m := testVoteModel()
	if got := m.votesForItem("g1"); got != 2 {
		t.Errorf("expected 2, got %d", got)
	}
	if got := m.votesForItem("c3"); got != 1 {
		t.Errorf("expected 1, got %d", got)
	}
	if got := m.votesForItem("x"); got != 0 {
		t.Errorf("expected 0, got %d", got)
	}
}

func TestVotesForItem_NilState(t *testing.T) {
	m := testModel()
	if got := m.votesForItem("c1"); got != 0 {
		t.Errorf("expected 0")
	}
}

func TestMyVotesForItem(t *testing.T) {
	m := testVoteModel()
	if got := m.myVotesForItem("g1"); got != 1 {
		t.Errorf("expected 1, got %d", got)
	}
	if got := m.myVotesForItem("c3"); got != 0 {
		t.Errorf("expected 0, got %d", got)
	}
}

func TestMyVotesForItem_NilState(t *testing.T) {
	m := testModel()
	if got := m.myVotesForItem("c1"); got != 0 {
		t.Errorf("expected 0")
	}
}

func TestVotesRemaining(t *testing.T) {
	m := testVoteModel()
	if got := m.votesRemaining(); got != 2 {
		t.Errorf("expected 2, got %d", got)
	}
}

func TestVotesRemaining_AllUsed(t *testing.T) {
	m := testVoteModel()
	m.state.Votes = append(m.state.Votes,
		protocol.Vote{ParticipantID: "p1", CardID: "c3"},
		protocol.Vote{ParticipantID: "p1", CardID: "c3"},
	)
	if got := m.votesRemaining(); got != 0 {
		t.Errorf("expected 0, got %d", got)
	}
}

func TestVotesRemaining_NilState(t *testing.T) {
	m := testModel()
	if got := m.votesRemaining(); got != 0 {
		t.Errorf("expected 0")
	}
}

func TestRemoveMyVote(t *testing.T) {
	m := testVoteModel()
	initial := len(m.state.Votes)
	m.removeMyVote("g1")
	if len(m.state.Votes) != initial-1 {
		t.Errorf("expected %d, got %d", initial-1, len(m.state.Votes))
	}
}

func TestRemoveMyVote_NotFound(t *testing.T) {
	m := testVoteModel()
	initial := len(m.state.Votes)
	m.removeMyVote("nonexistent")
	if len(m.state.Votes) != initial {
		t.Error("should not remove")
	}
}

func TestRemoveMyVote_NilState(t *testing.T) {
	m := testModel()
	m.removeMyVote("c1") // no panic
}

func TestViewVote_InAppView(t *testing.T) {
	m := testVoteModel()
	if view := m.View(); !strings.Contains(view, "VOTE") {
		t.Error("expected VOTE in stage bar")
	}
}
