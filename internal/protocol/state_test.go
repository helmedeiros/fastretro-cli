package protocol

import (
	"encoding/json"
	"testing"
)

func TestRetroState_JSONSerialization(t *testing.T) {
	state := RetroState{
		Stage:      "brainstorm",
		Meta:       RetroMeta{Name: "Sprint 42", Date: "2025-08-21", Context: "Team Alpha", TemplateID: "start-stop"},
		VoteBudget: 3,
		Participants: []Participant{
			{ID: "p1", Name: "Alice"},
		},
		Cards: []Card{
			{ID: "c1", ColumnID: "stop", Text: "too many meetings"},
		},
	}

	data, err := json.Marshal(state)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var decoded RetroState
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if decoded.Stage != "brainstorm" {
		t.Errorf("stage: got %q, want 'brainstorm'", decoded.Stage)
	}
	if decoded.Meta.Name != "Sprint 42" {
		t.Errorf("meta.name: got %q", decoded.Meta.Name)
	}
	if decoded.Meta.TemplateID != "start-stop" {
		t.Errorf("meta.templateId: got %q", decoded.Meta.TemplateID)
	}
	if decoded.VoteBudget != 3 {
		t.Errorf("voteBudget: got %d", decoded.VoteBudget)
	}
	if len(decoded.Participants) != 1 {
		t.Fatalf("participants: got %d", len(decoded.Participants))
	}
	if decoded.Participants[0].Name != "Alice" {
		t.Errorf("participant name: got %q", decoded.Participants[0].Name)
	}
}

func TestRetroState_EmptyState(t *testing.T) {
	data := `{"stage":"","meta":{"name":"","date":"","context":"","templateId":""},"participants":[],"cards":[],"groups":[],"votes":[],"voteBudget":0,"discussNotes":[]}`

	var state RetroState
	if err := json.Unmarshal([]byte(data), &state); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if state.Stage != "" {
		t.Errorf("expected empty stage, got %q", state.Stage)
	}
	if len(state.Participants) != 0 {
		t.Errorf("expected 0 participants, got %d", len(state.Participants))
	}
}

func TestCard_JSON(t *testing.T) {
	card := Card{ID: "c1", ColumnID: "start", Text: "ship faster"}
	data, err := json.Marshal(card)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var decoded Card
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if decoded.ID != "c1" {
		t.Errorf("id: got %q", decoded.ID)
	}
	if decoded.ColumnID != "start" {
		t.Errorf("columnId: got %q", decoded.ColumnID)
	}
	if decoded.Text != "ship faster" {
		t.Errorf("text: got %q", decoded.Text)
	}
}

func TestGroup_JSON(t *testing.T) {
	group := Group{
		ID:       "g1",
		ColumnID: "stop",
		Name:     "Process Issues",
		CardIDs:  []string{"c1", "c2"},
	}
	data, err := json.Marshal(group)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var decoded Group
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if decoded.Name != "Process Issues" {
		t.Errorf("name: got %q", decoded.Name)
	}
	if len(decoded.CardIDs) != 2 {
		t.Fatalf("cardIds: got %d", len(decoded.CardIDs))
	}
}

func TestVote_JSON(t *testing.T) {
	vote := Vote{ParticipantID: "p1", CardID: "c1"}
	data, err := json.Marshal(vote)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var decoded Vote
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if decoded.ParticipantID != "p1" {
		t.Errorf("participantId: got %q", decoded.ParticipantID)
	}
	if decoded.CardID != "c1" {
		t.Errorf("cardId: got %q", decoded.CardID)
	}
}

func TestDiscussState_JSON(t *testing.T) {
	ds := DiscussState{
		Order:        []string{"g1", "c2"},
		CurrentIndex: 1,
		Segment:      "actions",
	}
	data, err := json.Marshal(ds)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var decoded DiscussState
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if len(decoded.Order) != 2 {
		t.Fatalf("order: got %d", len(decoded.Order))
	}
	if decoded.CurrentIndex != 1 {
		t.Errorf("currentIndex: got %d", decoded.CurrentIndex)
	}
	if decoded.Segment != "actions" {
		t.Errorf("segment: got %q", decoded.Segment)
	}
}

func TestDiscussNote_JSON(t *testing.T) {
	note := DiscussNote{
		ID:           "n1",
		ParentCardID: "g1",
		Lane:         "actions",
		Text:         "Create runbook",
	}
	data, err := json.Marshal(note)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var decoded DiscussNote
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if decoded.Lane != "actions" {
		t.Errorf("lane: got %q", decoded.Lane)
	}
	if decoded.Text != "Create runbook" {
		t.Errorf("text: got %q", decoded.Text)
	}
}

func TestTimer_JSON(t *testing.T) {
	timer := Timer{
		Status:      "running",
		DurationMs:  300000,
		ElapsedMs:   60000,
		RemainingMs: 240000,
	}
	data, err := json.Marshal(timer)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var decoded Timer
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if decoded.Status != "running" {
		t.Errorf("status: got %q", decoded.Status)
	}
	if decoded.RemainingMs != 240000 {
		t.Errorf("remainingMs: got %d", decoded.RemainingMs)
	}
}

func TestRetroState_WithDiscuss(t *testing.T) {
	state := RetroState{
		Stage: "discuss",
		Discuss: &DiscussState{
			Order:        []string{"g1"},
			CurrentIndex: 0,
			Segment:      "context",
		},
		DiscussNotes: []DiscussNote{
			{ID: "n1", ParentCardID: "g1", Lane: "context", Text: "We need clarity"},
			{ID: "n2", ParentCardID: "g1", Lane: "actions", Text: "Write docs"},
		},
		ActionOwners: map[string]string{"n2": "Alice"},
	}

	data, err := json.Marshal(state)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var decoded RetroState
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if decoded.Discuss == nil {
		t.Fatal("discuss should be non-nil")
	}
	if len(decoded.DiscussNotes) != 2 {
		t.Fatalf("expected 2 discuss notes, got %d", len(decoded.DiscussNotes))
	}
	if decoded.ActionOwners["n2"] != "Alice" {
		t.Errorf("action owner: got %q", decoded.ActionOwners["n2"])
	}
}
