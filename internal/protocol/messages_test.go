package protocol

import (
	"encoding/json"
	"testing"
)

func TestParseMessage_State(t *testing.T) {
	data := `{"type":"state","state":{"stage":"brainstorm","meta":{"name":"Sprint 42"},"participants":[],"cards":[],"groups":[],"votes":[],"voteBudget":3,"discussNotes":[]}}`
	msg, err := ParseMessage([]byte(data))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.Type != "state" {
		t.Errorf("expected type 'state', got %q", msg.Type)
	}
	if msg.State == nil {
		t.Fatal("expected state to be non-nil")
	}
	if msg.State.Stage != "brainstorm" {
		t.Errorf("expected stage 'brainstorm', got %q", msg.State.Stage)
	}
	if msg.State.Meta.Name != "Sprint 42" {
		t.Errorf("expected meta name 'Sprint 42', got %q", msg.State.Meta.Name)
	}
}

func TestParseMessage_PeerCount(t *testing.T) {
	data := `{"type":"peer-count","count":5}`
	msg, err := ParseMessage([]byte(data))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.Type != "peer-count" {
		t.Errorf("expected type 'peer-count', got %q", msg.Type)
	}
	if msg.Count != 5 {
		t.Errorf("expected count 5, got %d", msg.Count)
	}
}

func TestParseMessage_TakenIDs(t *testing.T) {
	data := `{"type":"taken-ids","ids":["alice","bob"]}`
	msg, err := ParseMessage([]byte(data))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.Type != "taken-ids" {
		t.Errorf("expected type 'taken-ids', got %q", msg.Type)
	}
	if len(msg.IDs) != 2 {
		t.Fatalf("expected 2 IDs, got %d", len(msg.IDs))
	}
	if msg.IDs[0] != "alice" || msg.IDs[1] != "bob" {
		t.Errorf("unexpected IDs: %v", msg.IDs)
	}
}

func TestParseMessage_NavigateStage(t *testing.T) {
	data := `{"type":"navigate-stage","stage":"vote"}`
	msg, err := ParseMessage([]byte(data))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.Type != "navigate-stage" {
		t.Errorf("expected type 'navigate-stage', got %q", msg.Type)
	}
	if msg.Stage != "vote" {
		t.Errorf("expected stage 'vote', got %q", msg.Stage)
	}
}

func TestParseMessage_InvalidJSON(t *testing.T) {
	_, err := ParseMessage([]byte("not json"))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestParseMessage_EmptyObject(t *testing.T) {
	msg, err := ParseMessage([]byte("{}"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.Type != "" {
		t.Errorf("expected empty type, got %q", msg.Type)
	}
}

func TestStateMessage(t *testing.T) {
	state := &RetroState{
		Stage: "brainstorm",
		Meta:  RetroMeta{Name: "Test"},
		Cards: []Card{{ID: "c1", ColumnID: "stop", Text: "bugs"}},
	}
	data, err := StateMessage(state)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}
	if result["type"] != "state" {
		t.Errorf("expected type 'state', got %v", result["type"])
	}
}

func TestVoteStageMessage(t *testing.T) {
	data, err := VoteStageMessage("discuss", "alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]string
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}
	if result["type"] != "vote-stage" {
		t.Errorf("expected type 'vote-stage', got %v", result["type"])
	}
	if result["stage"] != "discuss" {
		t.Errorf("expected stage 'discuss', got %v", result["stage"])
	}
	if result["participantId"] != "alice" {
		t.Errorf("expected participantId 'alice', got %v", result["participantId"])
	}
}

func TestClaimIdentityMessage(t *testing.T) {
	data, err := ClaimIdentityMessage("bob")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]string
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}
	if result["type"] != "claim-identity" {
		t.Errorf("expected type 'claim-identity', got %v", result["type"])
	}
	if result["participantId"] != "bob" {
		t.Errorf("expected participantId 'bob', got %v", result["participantId"])
	}
}

func TestStateMessage_RoundTrip(t *testing.T) {
	original := &RetroState{
		Stage:      "vote",
		Meta:       RetroMeta{Name: "Sprint 42", Date: "2025-08-21", TemplateID: "start-stop"},
		VoteBudget: 3,
		Participants: []Participant{
			{ID: "p1", Name: "Alice"},
			{ID: "p2", Name: "Bob"},
		},
		Cards: []Card{
			{ID: "c1", ColumnID: "stop", Text: "long meetings"},
			{ID: "c2", ColumnID: "start", Text: "pair programming"},
		},
		Groups: []Group{
			{ID: "g1", ColumnID: "stop", Name: "Meetings", CardIDs: []string{"c1"}},
		},
		Votes: []Vote{
			{ParticipantID: "p1", CardID: "g1"},
		},
	}

	data, err := StateMessage(original)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	msg, err := ParseMessage(data)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	if msg.State.Stage != "vote" {
		t.Errorf("stage mismatch: got %q", msg.State.Stage)
	}
	if len(msg.State.Cards) != 2 {
		t.Errorf("expected 2 cards, got %d", len(msg.State.Cards))
	}
	if len(msg.State.Groups) != 1 {
		t.Errorf("expected 1 group, got %d", len(msg.State.Groups))
	}
	if len(msg.State.Votes) != 1 {
		t.Errorf("expected 1 vote, got %d", len(msg.State.Votes))
	}
}

func TestRequestStateMessage(t *testing.T) {
	data, err := RequestStateMessage()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]string
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if result["type"] != "request-state" {
		t.Errorf("expected type 'request-state', got %v", result["type"])
	}
}

func TestParseMessage_TeamInfo(t *testing.T) {
	data := `{"type":"team-info","teamInfo":{"teamName":"Acme","members":[{"id":"m1","name":"Alice"}],"agreements":[{"id":"a1","text":"Ship daily"}]}}`
	msg, err := ParseMessage([]byte(data))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.Type != "team-info" {
		t.Errorf("expected team-info, got %q", msg.Type)
	}
	if msg.TeamInfo == nil {
		t.Fatal("expected teamInfo")
	}
	if msg.TeamInfo.TeamName != "Acme" {
		t.Errorf("got %q", msg.TeamInfo.TeamName)
	}
	if len(msg.TeamInfo.Members) != 1 || msg.TeamInfo.Members[0].Name != "Alice" {
		t.Errorf("unexpected members: %v", msg.TeamInfo.Members)
	}
	if len(msg.TeamInfo.Agreements) != 1 || msg.TeamInfo.Agreements[0].Text != "Ship daily" {
		t.Errorf("unexpected agreements: %v", msg.TeamInfo.Agreements)
	}
}
