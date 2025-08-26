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

func TestViewVote_ShowsItems(t *testing.T) {
	m := testVoteModel()
	view := m.viewVote()

	if !strings.Contains(view, "Process") {
		t.Error("expected group 'Process' in view")
	}
	if !strings.Contains(view, "pair programming") {
		t.Error("expected ungrouped card in view")
	}
}

func TestViewVote_ShowsVoteBudget(t *testing.T) {
	m := testVoteModel()
	view := m.viewVote()

	if !strings.Contains(view, "2/3") {
		t.Error("expected remaining votes '2/3' in view")
	}
}

func TestVoteItems_GroupsFirst(t *testing.T) {
	m := testVoteModel()
	items := m.voteItems()

	if len(items) != 2 {
		t.Fatalf("expected 2 items (1 group + 1 ungrouped card), got %d", len(items))
	}
	if items[0].id != "g1" {
		t.Errorf("first item should be group, got %q", items[0].id)
	}
	if items[1].id != "c3" {
		t.Errorf("second item should be ungrouped card, got %q", items[1].id)
	}
}

func TestVoteItems_NoGroups(t *testing.T) {
	m := testModel()
	m.state = &protocol.RetroState{
		Cards: []protocol.Card{
			{ID: "c1", Text: "item 1"},
			{ID: "c2", Text: "item 2"},
		},
	}
	items := m.voteItems()

	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
}

func TestVoteItems_NilState(t *testing.T) {
	m := testModel()
	items := m.voteItems()

	if items != nil {
		t.Errorf("expected nil for nil state, got %v", items)
	}
}

func TestVotesForItem(t *testing.T) {
	m := testVoteModel()

	if got := m.votesForItem("g1"); got != 2 {
		t.Errorf("votes for g1: got %d, want 2", got)
	}
	if got := m.votesForItem("c3"); got != 1 {
		t.Errorf("votes for c3: got %d, want 1", got)
	}
	if got := m.votesForItem("nonexistent"); got != 0 {
		t.Errorf("votes for nonexistent: got %d, want 0", got)
	}
}

func TestVotesForItem_NilState(t *testing.T) {
	m := testModel()
	if got := m.votesForItem("c1"); got != 0 {
		t.Errorf("expected 0 for nil state, got %d", got)
	}
}

func TestMyVotesForItem(t *testing.T) {
	m := testVoteModel()

	if got := m.myVotesForItem("g1"); got != 1 {
		t.Errorf("my votes for g1: got %d, want 1", got)
	}
	if got := m.myVotesForItem("c3"); got != 0 {
		t.Errorf("my votes for c3: got %d, want 0", got)
	}
}

func TestMyVotesForItem_NilState(t *testing.T) {
	m := testModel()
	if got := m.myVotesForItem("c1"); got != 0 {
		t.Errorf("expected 0 for nil state, got %d", got)
	}
}

func TestVotesRemaining(t *testing.T) {
	m := testVoteModel()

	if got := m.votesRemaining(); got != 2 {
		t.Errorf("votes remaining: got %d, want 2", got)
	}
}

func TestVotesRemaining_AllUsed(t *testing.T) {
	m := testVoteModel()
	m.state.Votes = append(m.state.Votes,
		protocol.Vote{ParticipantID: "p1", CardID: "c3"},
		protocol.Vote{ParticipantID: "p1", CardID: "c3"},
	)

	if got := m.votesRemaining(); got != 0 {
		t.Errorf("votes remaining: got %d, want 0", got)
	}
}

func TestVotesRemaining_NilState(t *testing.T) {
	m := testModel()
	if got := m.votesRemaining(); got != 0 {
		t.Errorf("expected 0 for nil state, got %d", got)
	}
}

func TestRemoveMyVote(t *testing.T) {
	m := testVoteModel()
	initialLen := len(m.state.Votes)

	m.removeMyVote("g1")

	if len(m.state.Votes) != initialLen-1 {
		t.Errorf("expected %d votes after removal, got %d", initialLen-1, len(m.state.Votes))
	}

	// Only removes one vote
	for _, v := range m.state.Votes {
		if v.ParticipantID == "p1" && v.CardID == "g1" {
			t.Error("p1's vote for g1 should have been removed")
		}
	}
}

func TestRemoveMyVote_NotFound(t *testing.T) {
	m := testVoteModel()
	initialLen := len(m.state.Votes)

	m.removeMyVote("nonexistent")

	if len(m.state.Votes) != initialLen {
		t.Error("should not remove any votes for nonexistent item")
	}
}

func TestRemoveMyVote_NilState(t *testing.T) {
	m := testModel()
	m.removeMyVote("c1") // should not panic
}

func TestViewVote_NilState(t *testing.T) {
	m := testModel()
	view := m.viewVote()
	if view != "" {
		t.Errorf("expected empty view for nil state, got %q", view)
	}
}
