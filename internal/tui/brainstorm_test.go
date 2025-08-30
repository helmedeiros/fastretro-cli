package tui

import (
	"strings"
	"testing"

	"github.com/helmedeiros/fastretro-cli/internal/protocol"
)

func testBrainstormModel() Model {
	m := testModel()
	m.participantID = "p1"
	m.state = &protocol.RetroState{
		Stage: "brainstorm",
		Cards: []protocol.Card{
			{ID: "c1", ColumnID: "stop", Text: "long meetings"},
			{ID: "c2", ColumnID: "stop", Text: "too many bugs"},
			{ID: "c3", ColumnID: "start", Text: "pair programming"},
		},
	}
	return m
}

func TestViewBrainstorm_ShowsCards(t *testing.T) {
	m := testBrainstormModel()
	view := m.viewBrainstorm()

	if !strings.Contains(view, "long meetings") {
		t.Error("expected 'long meetings' in view")
	}
	if !strings.Contains(view, "pair programming") {
		t.Error("expected 'pair programming' in view")
	}
}

func TestViewBrainstorm_InputMode(t *testing.T) {
	m := testBrainstormModel()
	m.inputMode = true
	m.inputText = "new card"

	view := m.viewBrainstorm()

	if !strings.Contains(view, "new card") {
		t.Error("expected input text in view")
	}
}

func TestViewBrainstorm_EmptyState(t *testing.T) {
	m := testModel()
	m.state = nil
	view := m.viewBrainstorm()

	if view != "" {
		t.Errorf("expected empty view for nil state, got %q", view)
	}
}

func TestGetColumns_FromCards(t *testing.T) {
	m := testBrainstormModel()
	cols := m.getColumns()

	if len(cols) != 2 {
		t.Fatalf("expected 2 columns, got %d", len(cols))
	}

	ids := make(map[string]bool)
	for _, c := range cols {
		ids[c.id] = true
	}
	if !ids["stop"] || !ids["start"] {
		t.Errorf("expected stop and start columns, got %v", cols)
	}
}

func TestGetColumns_Default(t *testing.T) {
	m := testModel()
	m.state = &protocol.RetroState{Stage: "brainstorm"}
	cols := m.getColumns()

	if len(cols) != 2 {
		t.Fatalf("expected 2 default columns, got %d", len(cols))
	}
	if cols[0].id != "stop" || cols[1].id != "start" {
		t.Errorf("unexpected default columns: %v", cols)
	}
}

func TestGetColumns_NilState(t *testing.T) {
	m := testModel()
	cols := m.getColumns()

	if cols != nil {
		t.Errorf("expected nil for nil state, got %v", cols)
	}
}

func TestCardsForColumn(t *testing.T) {
	m := testBrainstormModel()

	stopCards := m.cardsForColumn("stop")
	if len(stopCards) != 2 {
		t.Errorf("expected 2 stop cards, got %d", len(stopCards))
	}

	startCards := m.cardsForColumn("start")
	if len(startCards) != 1 {
		t.Errorf("expected 1 start card, got %d", len(startCards))
	}

	emptyCards := m.cardsForColumn("nonexistent")
	if len(emptyCards) != 0 {
		t.Errorf("expected 0 cards for nonexistent column, got %d", len(emptyCards))
	}
}

func TestCardsForColumn_NilState(t *testing.T) {
	m := testModel()
	cards := m.cardsForColumn("stop")
	if cards != nil {
		t.Errorf("expected nil for nil state, got %v", cards)
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input string
		max   int
		want  string
	}{
		{"short", 10, "short"},
		{"exactly10!", 10, "exactly10!"},
		{"this is too long", 10, "this is t…"},
		{"", 5, ""},
		{"ab", 2, "ab"},
	}

	for _, tt := range tests {
		got := truncate(tt.input, tt.max)
		if got != tt.want {
			t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.max, got, tt.want)
		}
	}
}

func TestViewBrainstorm_ShowsGroups(t *testing.T) {
	m := testModel()
	m.participantID = "p1"
	m.state = &protocol.RetroState{
		Stage: "brainstorm",
		Cards: []protocol.Card{
			{ID: "c1", ColumnID: "stop", Text: "long meetings"},
			{ID: "c2", ColumnID: "stop", Text: "too many bugs"},
			{ID: "c3", ColumnID: "start", Text: "pair programming"},
		},
		Groups: []protocol.Group{
			{ID: "g1", ColumnID: "stop", Name: "Process Issues", CardIDs: []string{"c1", "c2"}},
		},
	}
	view := m.viewBrainstorm()

	if !strings.Contains(view, "Process Issues") {
		t.Error("expected group name in view")
	}
	if !strings.Contains(view, "long meetings") {
		t.Error("expected grouped card text in view")
	}
	// Ungrouped card should still show
	if !strings.Contains(view, "pair programming") {
		t.Error("expected ungrouped card in view")
	}
}

func TestGroupsForColumn(t *testing.T) {
	m := testModel()
	m.state = &protocol.RetroState{
		Groups: []protocol.Group{
			{ID: "g1", ColumnID: "stop", Name: "A"},
			{ID: "g2", ColumnID: "start", Name: "B"},
		},
	}

	stop := m.groupsForColumn("stop")
	if len(stop) != 1 || stop[0].Name != "A" {
		t.Errorf("expected 1 stop group, got %v", stop)
	}

	start := m.groupsForColumn("start")
	if len(start) != 1 || start[0].Name != "B" {
		t.Errorf("expected 1 start group, got %v", start)
	}

	empty := m.groupsForColumn("nonexistent")
	if len(empty) != 0 {
		t.Errorf("expected 0 groups, got %d", len(empty))
	}
}

func TestGroupsForColumn_NilState(t *testing.T) {
	m := testModel()
	if groups := m.groupsForColumn("stop"); groups != nil {
		t.Errorf("expected nil, got %v", groups)
	}
}

func TestGroupedCardIDs(t *testing.T) {
	m := testModel()
	m.state = &protocol.RetroState{
		Groups: []protocol.Group{
			{ID: "g1", CardIDs: []string{"c1", "c2"}},
			{ID: "g2", CardIDs: []string{"c3"}},
		},
	}

	ids := m.groupedCardIDs()
	if !ids["c1"] || !ids["c2"] || !ids["c3"] {
		t.Error("expected c1, c2, c3 to be grouped")
	}
	if ids["c4"] {
		t.Error("c4 should not be grouped")
	}
}

func TestGroupedCardIDs_NilState(t *testing.T) {
	m := testModel()
	ids := m.groupedCardIDs()
	if len(ids) != 0 {
		t.Errorf("expected empty map, got %v", ids)
	}
}

func TestCardByID_Found(t *testing.T) {
	m := testBrainstormModel()
	card, ok := m.cardByID("c1")
	if !ok {
		t.Fatal("expected card to be found")
	}
	if card.Text != "long meetings" {
		t.Errorf("expected 'long meetings', got %q", card.Text)
	}
}

func TestCardByID_NotFound(t *testing.T) {
	m := testBrainstormModel()
	_, ok := m.cardByID("nonexistent")
	if ok {
		t.Error("expected card not found")
	}
}

func TestCardByID_NilState(t *testing.T) {
	m := testModel()
	_, ok := m.cardByID("c1")
	if ok {
		t.Error("expected card not found with nil state")
	}
}

func TestViewBrainstorm_GroupedCardsNotDuplicated(t *testing.T) {
	m := testModel()
	m.participantID = "p1"
	m.state = &protocol.RetroState{
		Stage: "brainstorm",
		Cards: []protocol.Card{
			{ID: "c1", ColumnID: "stop", Text: "unique-ungrouped"},
			{ID: "c2", ColumnID: "stop", Text: "in-group-card"},
		},
		Groups: []protocol.Group{
			{ID: "g1", ColumnID: "stop", Name: "MyGroup", CardIDs: []string{"c2"}},
		},
	}
	view := m.viewBrainstorm()

	// Ungrouped card shown as bullet
	if !strings.Contains(view, "unique-ungrouped") {
		t.Error("expected ungrouped card")
	}
	// Grouped card should appear inside group, not as standalone bullet
	count := strings.Count(view, "in-group-card")
	if count != 1 {
		t.Errorf("grouped card should appear exactly once, appeared %d times", count)
	}
}

func TestViewBrainstorm_NoColumns(t *testing.T) {
	m := testModel()
	m.state = &protocol.RetroState{Stage: "brainstorm"}
	// No cards = default columns shown
	view := m.viewBrainstorm()
	if !strings.Contains(view, "empty") {
		t.Error("expected empty columns to show '(empty)'")
	}
}
