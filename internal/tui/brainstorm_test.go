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
			{ID: "cli-p1-0", ColumnID: "stop", Text: "my card"},
		},
	}
	return m
}

// --- columnBrainstormItems ---

func TestColumnBrainstormItems_UngroupedCards(t *testing.T) {
	m := testBrainstormModel()
	items := m.columnBrainstormItems("stop")

	if len(items) != 3 {
		t.Fatalf("expected 3 stop items, got %d", len(items))
	}
	for _, item := range items {
		if item.kind != "card" {
			t.Errorf("expected all cards, got %q", item.kind)
		}
	}
}

func TestColumnBrainstormItems_WithGroups(t *testing.T) {
	m := testBrainstormModel()
	m.state.Groups = []protocol.Group{
		{ID: "g1", ColumnID: "stop", Name: "Issues", CardIDs: []string{"c1", "c2"}},
	}
	items := m.columnBrainstormItems("stop")

	// group-header, c1, c2 (group-card), cli-p1-0 (ungrouped card)
	if len(items) != 4 {
		t.Fatalf("expected 4 items, got %d", len(items))
	}
	if items[0].kind != "group-header" {
		t.Error("first should be group-header")
	}
	if items[1].kind != "group-card" {
		t.Error("second should be group-card")
	}
	if items[3].kind != "card" {
		t.Error("fourth should be ungrouped card")
	}
}

func TestColumnBrainstormItems_MineFlag(t *testing.T) {
	m := testBrainstormModel()
	items := m.columnBrainstormItems("stop")

	for _, item := range items {
		if item.cardID == "cli-p1-0" && !item.mine {
			t.Error("cli-p1-0 should be mine")
		}
		if item.cardID == "c1" && item.mine {
			t.Error("c1 should not be mine")
		}
	}
}

func TestColumnBrainstormItems_NilState(t *testing.T) {
	m := testModel()
	if items := m.columnBrainstormItems("stop"); items != nil {
		t.Errorf("expected nil")
	}
}

// --- isMyCard ---

func TestIsMyCard(t *testing.T) {
	m := testBrainstormModel()

	if !m.isMyCard("cli-p1-0") {
		t.Error("should be mine")
	}
	if m.isMyCard("c1") {
		t.Error("should not be mine")
	}
}

// --- viewBrainstorm ---

func TestViewBrainstorm_ShowsCards(t *testing.T) {
	m := testBrainstormModel()
	view := m.viewBrainstorm()

	if !strings.Contains(view, "long meetings") {
		t.Error("expected card text")
	}
	if !strings.Contains(view, "pair programming") {
		t.Error("expected start column card")
	}
}

func TestViewBrainstorm_ShowsCursor(t *testing.T) {
	m := testBrainstormModel()
	m.cursor = 0
	view := m.viewBrainstorm()

	if !strings.Contains(view, ">") {
		t.Error("expected cursor")
	}
}

func TestViewBrainstorm_ShowsHelp(t *testing.T) {
	m := testBrainstormModel()
	view := m.viewBrainstorm()

	if !strings.Contains(view, "delete") {
		t.Error("expected delete hint")
	}
	if !strings.Contains(view, "[a] add") {
		t.Error("expected add hint")
	}
}

func TestViewBrainstorm_InputMode(t *testing.T) {
	m := testBrainstormModel()
	m.inputMode = true
	m.inputText = "new card"
	view := m.viewBrainstorm()

	if !strings.Contains(view, "new card") {
		t.Error("expected input text")
	}
}

func TestViewBrainstorm_EmptyState(t *testing.T) {
	m := testModel()
	m.state = nil
	if view := m.viewBrainstorm(); view != "" {
		t.Errorf("expected empty, got %q", view)
	}
}

func TestViewBrainstorm_EmptyColumns(t *testing.T) {
	m := testModel()
	m.state = &protocol.RetroState{Stage: "brainstorm"}
	view := m.viewBrainstorm()

	if !strings.Contains(view, "empty") {
		t.Error("expected empty indicator")
	}
}

func TestViewBrainstorm_ActiveColumn(t *testing.T) {
	m := testBrainstormModel()
	m.activeCol = 0
	view := m.viewBrainstorm()

	if !strings.Contains(view, "[1]") {
		t.Error("expected numbered column title")
	}
}

// --- handleBrainstormKeys navigation ---

func TestHandleBrainstormKeys_Up(t *testing.T) {
	m := testBrainstormModel()
	m.cursor = 1

	result, _ := m.handleBrainstormKeys(keyMsg("up"))
	model := result.(Model)

	if model.cursor != 0 {
		t.Errorf("expected 0, got %d", model.cursor)
	}
}

func TestHandleBrainstormKeys_Down(t *testing.T) {
	m := testBrainstormModel()

	result, _ := m.handleBrainstormKeys(keyMsg("down"))
	model := result.(Model)

	if model.cursor != 1 {
		t.Errorf("expected 1, got %d", model.cursor)
	}
}

func TestHandleBrainstormKeys_UpAtTop(t *testing.T) {
	m := testBrainstormModel()

	result, _ := m.handleBrainstormKeys(keyMsg("up"))
	model := result.(Model)

	if model.cursor != 0 {
		t.Error("should stay at 0")
	}
}

func TestHandleBrainstormKeys_Tab(t *testing.T) {
	m := testBrainstormModel()
	m.cursor = 2

	result, _ := m.handleBrainstormKeys(keyMsg("tab"))
	model := result.(Model)

	if model.activeCol != 1 {
		t.Errorf("expected col 1, got %d", model.activeCol)
	}
	if model.cursor != 0 {
		t.Error("cursor should reset on column switch")
	}
}

func TestHandleBrainstormKeys_ShiftTab(t *testing.T) {
	m := testBrainstormModel()
	m.activeCol = 1

	result, _ := m.handleBrainstormKeys(keyMsg("shift+tab"))
	model := result.(Model)

	if model.activeCol != 0 {
		t.Errorf("expected col 0")
	}
}

func TestHandleBrainstormKeys_Right(t *testing.T) {
	m := testBrainstormModel()

	result, _ := m.handleBrainstormKeys(keyMsg("right"))
	model := result.(Model)

	if model.activeCol != 1 {
		t.Errorf("expected col 1")
	}
}

func TestHandleBrainstormKeys_Left(t *testing.T) {
	m := testBrainstormModel()
	m.activeCol = 1

	result, _ := m.handleBrainstormKeys(keyMsg("left"))
	model := result.(Model)

	if model.activeCol != 0 {
		t.Errorf("expected col 0")
	}
}

func TestHandleBrainstormKeys_AddCard(t *testing.T) {
	m := testBrainstormModel()

	result, _ := m.handleBrainstormKeys(keyMsg("a"))
	model := result.(Model)

	if !model.inputMode {
		t.Error("expected input mode")
	}
}

// --- delete ---

func TestHandleBrainstormKeys_DeleteMyCard(t *testing.T) {
	m := testBrainstormModel()
	// cli-p1-0 is at index 2 in stop column (c1, c2, cli-p1-0)
	m.cursor = 2

	initialCards := len(m.state.Cards)
	result, _ := m.handleBrainstormKeys(keyMsg("d"))
	model := result.(Model)

	if len(model.state.Cards) != initialCards-1 {
		t.Errorf("expected %d cards, got %d", initialCards-1, len(model.state.Cards))
	}
}

func TestHandleBrainstormKeys_DeleteOthersCard(t *testing.T) {
	m := testBrainstormModel()
	m.cursor = 0 // c1, not mine

	initialCards := len(m.state.Cards)
	result, _ := m.handleBrainstormKeys(keyMsg("d"))
	model := result.(Model)

	if len(model.state.Cards) != initialCards {
		t.Error("should not delete others' cards")
	}
}

func TestRemoveCard_Basic(t *testing.T) {
	m := testBrainstormModel()
	m.removeCard("cli-p1-0")

	for _, c := range m.state.Cards {
		if c.ID == "cli-p1-0" {
			t.Error("card should be removed")
		}
	}
}

func TestRemoveCard_FromGroup(t *testing.T) {
	m := testBrainstormModel()
	m.state.Groups = []protocol.Group{
		{ID: "g1", ColumnID: "stop", Name: "G", CardIDs: []string{"c1", "cli-p1-0", "c2"}},
	}

	m.removeCard("cli-p1-0")

	if len(m.state.Groups) != 1 {
		t.Fatal("group should survive")
	}
	if len(m.state.Groups[0].CardIDs) != 2 {
		t.Errorf("expected 2 cards in group, got %d", len(m.state.Groups[0].CardIDs))
	}
}

func TestRemoveCard_GroupDissolves(t *testing.T) {
	m := testBrainstormModel()
	m.state.Groups = []protocol.Group{
		{ID: "g1", ColumnID: "stop", Name: "G", CardIDs: []string{"c1", "cli-p1-0"}},
	}

	m.removeCard("cli-p1-0")

	if len(m.state.Groups) != 0 {
		t.Error("group should dissolve with < 2 cards")
	}
}

func TestRemoveCard_NilState(t *testing.T) {
	m := testModel()
	m.removeCard("c1") // no panic
}

// --- input ---

func TestHandleBrainstormInput_Enter(t *testing.T) {
	m := testBrainstormModel()
	m.inputMode = true
	m.inputText = "new card"

	initialCards := len(m.state.Cards)
	result, _ := m.handleBrainstormInput(keyMsg("enter"))
	model := result.(Model)

	if model.inputMode {
		t.Error("should exit input")
	}
	if len(model.state.Cards) != initialCards+1 {
		t.Errorf("expected %d cards", initialCards+1)
	}
}

func TestHandleBrainstormInput_EnterEmpty(t *testing.T) {
	m := testBrainstormModel()
	m.inputMode = true
	m.inputText = ""

	initialCards := len(m.state.Cards)
	result, _ := m.handleBrainstormInput(keyMsg("enter"))
	model := result.(Model)

	if len(model.state.Cards) != initialCards {
		t.Error("empty should not add")
	}
}

func TestHandleBrainstormInput_Escape(t *testing.T) {
	m := testBrainstormModel()
	m.inputMode = true

	result, _ := m.handleBrainstormInput(keyMsg("esc"))
	model := result.(Model)

	if model.inputMode {
		t.Error("should exit")
	}
}

func TestHandleBrainstormInput_Backspace(t *testing.T) {
	m := testBrainstormModel()
	m.inputMode = true
	m.inputText = "ab"

	result, _ := m.handleBrainstormInput(keyMsg("backspace"))
	model := result.(Model)

	if model.inputText != "a" {
		t.Errorf("got %q", model.inputText)
	}
}

func TestHandleBrainstormInput_Type(t *testing.T) {
	m := testBrainstormModel()
	m.inputMode = true
	m.inputText = "a"

	result, _ := m.handleBrainstormInput(keyMsg("b"))
	model := result.(Model)

	if model.inputText != "ab" {
		t.Errorf("got %q", model.inputText)
	}
}

func TestHandleBrainstormInput_MaxLength(t *testing.T) {
	m := testBrainstormModel()
	m.inputMode = true
	buf := make([]byte, 140)
	for i := range buf {
		buf[i] = 'a'
	}
	m.inputText = string(buf)

	result, _ := m.handleBrainstormInput(keyMsg("b"))
	model := result.(Model)

	if len(model.inputText) != 140 {
		t.Errorf("should not exceed 140, got %d", len(model.inputText))
	}
}

func TestHandleBrainstormKeys_DelegatesToInput(t *testing.T) {
	m := testBrainstormModel()
	m.inputMode = true
	m.inputText = "x"

	result, _ := m.handleBrainstormKeys(keyMsg("a"))
	model := result.(Model)

	if model.inputText != "xa" {
		t.Errorf("got %q", model.inputText)
	}
}

// --- helpers ---

func TestGetColumns_FromCards(t *testing.T) {
	m := testBrainstormModel()
	cols := m.getColumns()

	if len(cols) != 2 {
		t.Fatalf("expected 2, got %d", len(cols))
	}
}

func TestGetColumns_Default(t *testing.T) {
	m := testModel()
	m.state = &protocol.RetroState{Stage: "brainstorm"}
	cols := m.getColumns()

	if len(cols) != 2 {
		t.Fatalf("expected 2 defaults, got %d", len(cols))
	}
}

func TestGetColumns_NilState(t *testing.T) {
	m := testModel()
	if cols := m.getColumns(); cols != nil {
		t.Errorf("expected nil")
	}
}

func TestCardsForColumn(t *testing.T) {
	m := testBrainstormModel()
	if cards := m.cardsForColumn("stop"); len(cards) != 3 {
		t.Errorf("expected 3 stop cards, got %d", len(cards))
	}
	if cards := m.cardsForColumn("start"); len(cards) != 1 {
		t.Errorf("expected 1 start card, got %d", len(cards))
	}
}

func TestCardsForColumn_NilState(t *testing.T) {
	m := testModel()
	if cards := m.cardsForColumn("stop"); cards != nil {
		t.Errorf("expected nil")
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
		if got := truncate(tt.input, tt.max); got != tt.want {
			t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.max, got, tt.want)
		}
	}
}
