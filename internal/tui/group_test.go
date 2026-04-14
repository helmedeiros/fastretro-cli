package tui

import (
	"strings"
	"testing"

	"github.com/helmedeiros/fastretro-cli/internal/protocol"
)

func testGroupModel() Model {
	m := testModel()
	m.participantID = "p1"
	m.state = &protocol.RetroState{
		Stage: "group",
		Cards: []protocol.Card{
			{ID: "c1", ColumnID: "stop", Text: "long meetings"},
			{ID: "c2", ColumnID: "stop", Text: "too many bugs"},
			{ID: "c3", ColumnID: "stop", Text: "unclear requirements"},
			{ID: "c4", ColumnID: "start", Text: "pair programming"},
		},
		Groups: []protocol.Group{
			{ID: "g1", ColumnID: "stop", Name: "Process Issues", CardIDs: []string{"c1", "c2"}},
		},
	}
	return m
}

// --- columnGroupItems ---

func TestColumnGroupItems_StopColumn(t *testing.T) {
	m := testGroupModel()
	items := m.columnGroupItems("stop")

	// g1 header, c1, c2 (grouped), c3 (ungrouped)
	if len(items) != 4 {
		t.Fatalf("expected 4 items, got %d", len(items))
	}
	if items[0].kind != "group-header" {
		t.Error("first should be group header")
	}
	if items[1].cardID != "c1" || !items[1].grouped {
		t.Error("second should be grouped card c1")
	}
	if items[3].cardID != "c3" || items[3].grouped {
		t.Error("fourth should be ungrouped card c3")
	}
}

func TestColumnGroupItems_StartColumn(t *testing.T) {
	m := testGroupModel()
	items := m.columnGroupItems("start")

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].cardID != "c4" {
		t.Errorf("expected c4, got %q", items[0].cardID)
	}
}

func TestColumnGroupItems_NilState(t *testing.T) {
	m := testModel()
	if items := m.columnGroupItems("stop"); items != nil {
		t.Errorf("expected nil, got %v", items)
	}
}

func TestColumnGroupItems_NoGroups(t *testing.T) {
	m := testModel()
	m.state = &protocol.RetroState{
		Cards: []protocol.Card{
			{ID: "c1", ColumnID: "stop", Text: "a"},
			{ID: "c2", ColumnID: "stop", Text: "b"},
		},
	}
	items := m.columnGroupItems("stop")

	if len(items) != 2 {
		t.Fatalf("expected 2, got %d", len(items))
	}
	for _, item := range items {
		if item.grouped {
			t.Error("no items should be grouped")
		}
	}
}

// --- viewGroup ---

func TestViewGroup_ShowsColumns(t *testing.T) {
	m := testGroupModel()
	view := m.viewGroup()

	if !strings.Contains(view, "Stop") {
		t.Error("expected Stop column")
	}
	if !strings.Contains(view, "Start") {
		t.Error("expected Start column")
	}
}

func TestViewGroup_ShowsGroupHeader(t *testing.T) {
	m := testGroupModel()
	view := m.viewGroup()

	if !strings.Contains(view, "Process Issues") {
		t.Error("expected group name")
	}
}

func TestViewGroup_ShowsCards(t *testing.T) {
	m := testGroupModel()
	view := m.viewGroup()

	if !strings.Contains(view, "long meetings") {
		t.Error("expected grouped card")
	}
	if !strings.Contains(view, "unclear requirements") {
		t.Error("expected ungrouped card")
	}
}

func TestViewGroup_ShowsHelp(t *testing.T) {
	m := testGroupModel()
	view := m.viewGroup()

	if !strings.Contains(view, "merge") {
		t.Error("expected merge in help")
	}
	if !strings.Contains(view, "column") {
		t.Error("expected column navigation in help")
	}
}

func TestViewGroup_MergeMode(t *testing.T) {
	m := testGroupModel()
	m.mergeSource = "c3"
	view := m.viewGroup()

	if !strings.Contains(view, "MERGE") {
		t.Error("expected MERGE indicator")
	}
}

func TestViewGroup_RenameInput(t *testing.T) {
	m := testGroupModel()
	m.inputMode = true
	m.inputText = "New"
	view := m.viewGroup()

	if !strings.Contains(view, "Rename") {
		t.Error("expected rename prompt")
	}
}

func TestViewGroup_NilState(t *testing.T) {
	m := testModel()
	if view := m.viewGroup(); view != "" {
		t.Errorf("expected empty, got %q", view)
	}
}

func TestViewGroup_ActiveColumnHighlight(t *testing.T) {
	m := testGroupModel()
	m.activeCol = 0
	view := m.viewGroup()

	if !strings.Contains(view, "▶") {
		t.Error("expected active column indicator")
	}
}

// --- handleGroupKeys navigation ---

func TestHandleGroupKeys_Up(t *testing.T) {
	m := testGroupModel()
	m.cursor = 2

	result, _ := m.handleGroupKeys(keyMsg("up"))
	model := result.(Model)

	if model.cursor != 1 {
		t.Errorf("expected 1, got %d", model.cursor)
	}
}

func TestHandleGroupKeys_Down(t *testing.T) {
	m := testGroupModel()

	result, _ := m.handleGroupKeys(keyMsg("down"))
	model := result.(Model)

	if model.cursor != 1 {
		t.Errorf("expected 1, got %d", model.cursor)
	}
}

func TestHandleGroupKeys_UpAtTop(t *testing.T) {
	m := testGroupModel()

	result, _ := m.handleGroupKeys(keyMsg("up"))
	model := result.(Model)

	if model.cursor != 0 {
		t.Errorf("should stay at 0")
	}
}

func TestHandleGroupKeys_Tab(t *testing.T) {
	m := testGroupModel()
	m.activeCol = 0
	m.cursor = 2

	result, _ := m.handleGroupKeys(keyMsg("tab"))
	model := result.(Model)

	if model.activeCol != 1 {
		t.Errorf("expected col 1, got %d", model.activeCol)
	}
	if model.cursor != 0 {
		t.Error("cursor should reset to 0 on column switch")
	}
}

func TestHandleGroupKeys_ShiftTab(t *testing.T) {
	m := testGroupModel()
	m.activeCol = 1

	result, _ := m.handleGroupKeys(keyMsg("shift+tab"))
	model := result.(Model)

	if model.activeCol != 0 {
		t.Errorf("expected col 0, got %d", model.activeCol)
	}
}

func TestHandleGroupKeys_Right(t *testing.T) {
	m := testGroupModel()

	result, _ := m.handleGroupKeys(keyMsg("right"))
	model := result.(Model)

	if model.activeCol != 1 {
		t.Errorf("expected col 1")
	}
}

func TestHandleGroupKeys_Left(t *testing.T) {
	m := testGroupModel()
	m.activeCol = 1

	result, _ := m.handleGroupKeys(keyMsg("left"))
	model := result.(Model)

	if model.activeCol != 0 {
		t.Errorf("expected col 0")
	}
}

// --- merge ---

func TestHandleGroupKeys_MergeFirstSelect(t *testing.T) {
	m := testGroupModel()
	m.cursor = 3 // c3, ungrouped

	result, _ := m.handleGroupKeys(keyMsg("m"))
	model := result.(Model)

	if model.mergeSource != "c3" {
		t.Errorf("expected 'c3', got %q", model.mergeSource)
	}
}

func TestHandleGroupKeys_MergeComplete(t *testing.T) {
	m := testGroupModel()
	// Switch to start column for a simple merge
	m.activeCol = 1
	m.state.Cards = append(m.state.Cards, protocol.Card{ID: "c5", ColumnID: "start", Text: "deploy more"})
	m.mergeSource = "c4"
	m.cursor = 1 // c5

	result, _ := m.handleGroupKeys(keyMsg("m"))
	model := result.(Model)

	if model.mergeSource != "" {
		t.Error("merge source should be cleared")
	}
	if len(model.state.Groups) != 2 {
		t.Errorf("expected 2 groups, got %d", len(model.state.Groups))
	}
}

func TestHandleGroupKeys_MergeSameCard(t *testing.T) {
	m := testGroupModel()
	m.mergeSource = "c3"
	m.cursor = 3

	result, _ := m.handleGroupKeys(keyMsg("m"))
	model := result.(Model)

	if model.mergeSource != "c3" {
		t.Error("should not merge with self")
	}
}

func TestHandleGroupKeys_MergeOnHeader(t *testing.T) {
	m := testGroupModel()
	m.cursor = 0 // group header

	result, _ := m.handleGroupKeys(keyMsg("m"))
	model := result.(Model)

	if model.mergeSource != "" {
		t.Error("should not select header for merge")
	}
}

func TestHandleGroupKeys_Escape(t *testing.T) {
	m := testGroupModel()
	m.mergeSource = "c3"

	result, _ := m.handleGroupKeys(keyMsg("esc"))
	model := result.(Model)

	if model.mergeSource != "" {
		t.Error("escape should cancel merge")
	}
}

// --- ungroup ---

func TestHandleGroupKeys_Ungroup(t *testing.T) {
	m := testGroupModel()
	m.cursor = 1 // c1, grouped

	result, _ := m.handleGroupKeys(keyMsg("u"))
	model := result.(Model)

	if len(model.state.Groups) != 0 {
		t.Errorf("expected 0 groups, got %d", len(model.state.Groups))
	}
}

func TestHandleGroupKeys_UngroupUngrouped(t *testing.T) {
	m := testGroupModel()
	m.cursor = 3 // c3, not grouped

	result, _ := m.handleGroupKeys(keyMsg("u"))
	model := result.(Model)

	if len(model.state.Groups) != 1 {
		t.Error("should not affect groups")
	}
}

// --- rename ---

func TestHandleGroupKeys_Rename(t *testing.T) {
	m := testGroupModel()
	m.cursor = 0 // header

	result, _ := m.handleGroupKeys(keyMsg("e"))
	model := result.(Model)

	if !model.inputMode {
		t.Error("expected rename mode")
	}
}

func TestHandleGroupKeys_RenameOnCard(t *testing.T) {
	m := testGroupModel()
	m.cursor = 1 // card

	result, _ := m.handleGroupKeys(keyMsg("e"))
	model := result.(Model)

	if model.inputMode {
		t.Error("should not rename on card")
	}
}

func TestHandleGroupRenameInput_Enter(t *testing.T) {
	m := testGroupModel()
	m.inputMode = true
	m.inputText = "New Name"
	m.cursor = 0

	result, _ := m.handleGroupRenameInput(keyMsg("enter"))
	model := result.(Model)

	if model.inputMode {
		t.Error("should exit rename")
	}
	if model.state.Groups[0].Name != "New Name" {
		t.Errorf("got %q", model.state.Groups[0].Name)
	}
}

func TestHandleGroupRenameInput_EnterEmpty(t *testing.T) {
	m := testGroupModel()
	m.inputMode = true
	m.inputText = "  "
	m.cursor = 0

	result, _ := m.handleGroupRenameInput(keyMsg("enter"))
	model := result.(Model)

	if model.state.Groups[0].Name != "Process Issues" {
		t.Error("empty should not rename")
	}
}

func TestHandleGroupRenameInput_Escape(t *testing.T) {
	m := testGroupModel()
	m.inputMode = true

	result, _ := m.handleGroupRenameInput(keyMsg("esc"))
	model := result.(Model)

	if model.inputMode {
		t.Error("should exit")
	}
}

func TestHandleGroupRenameInput_Type(t *testing.T) {
	m := testGroupModel()
	m.inputMode = true
	m.inputText = "N"

	result, _ := m.handleGroupRenameInput(keyMsg("e"))
	model := result.(Model)

	if model.inputText != "Ne" {
		t.Errorf("got %q", model.inputText)
	}
}

func TestHandleGroupRenameInput_Backspace(t *testing.T) {
	m := testGroupModel()
	m.inputMode = true
	m.inputText = "ab"

	result, _ := m.handleGroupRenameInput(keyMsg("backspace"))
	model := result.(Model)

	if model.inputText != "a" {
		t.Errorf("got %q", model.inputText)
	}
}

// --- logic ---

func TestMergeCards_CreateNew(t *testing.T) {
	m := testGroupModel()
	m.mergeCards("c3", "c4")

	if len(m.state.Groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(m.state.Groups))
	}
}

func TestMergeCards_AddToExisting(t *testing.T) {
	m := testGroupModel()
	m.mergeCards("c3", "c1")

	if len(m.state.Groups[0].CardIDs) != 3 {
		t.Errorf("expected 3 cards, got %d", len(m.state.Groups[0].CardIDs))
	}
}

func TestMergeCards_SourceGrouped(t *testing.T) {
	m := testGroupModel()
	m.mergeCards("c1", "c3")
	if len(m.state.Groups) != 1 {
		t.Error("no-op expected")
	}
}

func TestMergeCards_NilState(t *testing.T) {
	m := testModel()
	m.mergeCards("c1", "c2")
}

func TestRenameGroupByID(t *testing.T) {
	m := testGroupModel()
	m.renameGroupByID("g1", "Better")
	if m.state.Groups[0].Name != "Better" {
		t.Errorf("got %q", m.state.Groups[0].Name)
	}
}

func TestRenameGroupByID_NotFound(t *testing.T) {
	m := testGroupModel()
	m.renameGroupByID("x", "Y")
	if m.state.Groups[0].Name != "Process Issues" {
		t.Error("should not change")
	}
}

func TestUngroupCardByID_Deletes(t *testing.T) {
	m := testGroupModel()
	m.ungroupCardByID("c1")
	if len(m.state.Groups) != 0 {
		t.Errorf("got %d groups", len(m.state.Groups))
	}
}

func TestUngroupCardByID_Survives(t *testing.T) {
	m := testGroupModel()
	m.state.Groups[0].CardIDs = []string{"c1", "c2", "c3"}
	m.ungroupCardByID("c1")
	if len(m.state.Groups[0].CardIDs) != 2 {
		t.Errorf("got %d", len(m.state.Groups[0].CardIDs))
	}
}

func TestUngroupCardByID_NotInGroup(t *testing.T) {
	m := testGroupModel()
	m.ungroupCardByID("c3")
	if len(m.state.Groups) != 1 {
		t.Error("should not change")
	}
}

func TestUngroupedCardsForColumn(t *testing.T) {
	m := testGroupModel()
	u := m.ungroupedCardsForColumn("stop")
	if len(u) != 1 || u[0].ID != "c3" {
		t.Errorf("expected [c3], got %v", u)
	}
}

func TestViewGroup_InAppView(t *testing.T) {
	m := testGroupModel()
	if view := m.View(); !strings.Contains(view, "GROUP") {
		t.Error("expected GROUP in header")
	}
}

func TestHandleGroupKeys_DelegatesToRename(t *testing.T) {
	m := testGroupModel()
	m.inputMode = true
	m.inputText = "x"

	result, _ := m.handleGroupKeys(keyMsg("y"))
	model := result.(Model)

	if model.inputText != "xy" {
		t.Errorf("got %q", model.inputText)
	}
}

func TestBroadcastState_NilClient(t *testing.T) {
	m := testGroupModel()
	m.broadcastState()
}
