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

func TestViewBrainstorm_NoColumns(t *testing.T) {
	m := testModel()
	m.state = &protocol.RetroState{Stage: "brainstorm"}
	// No cards = default columns shown
	view := m.viewBrainstorm()
	if !strings.Contains(view, "empty") {
		t.Error("expected empty columns to show '(empty)'")
	}
}
