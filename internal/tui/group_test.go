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

// --- flatGroupItems ---

func TestFlatGroupItems_Structure(t *testing.T) {
	m := testGroupModel()
	items := m.flatGroupItems()

	// g1 header, c1, c2, c3 (ungrouped), c4 (ungrouped, start col)
	if len(items) != 5 {
		t.Fatalf("expected 5 items, got %d", len(items))
	}
	if items[0].kind != "group-header" {
		t.Error("first item should be group header")
	}
	if items[1].kind != "card" || items[1].cardID != "c1" {
		t.Error("second item should be card c1")
	}
	if items[1].grouped != true {
		t.Error("c1 should be marked as grouped")
	}
	if items[3].kind != "card" || items[3].cardID != "c3" {
		t.Error("fourth item should be ungrouped card c3")
	}
	if items[3].grouped != false {
		t.Error("c3 should not be grouped")
	}
}

func TestFlatGroupItems_NilState(t *testing.T) {
	m := testModel()
	items := m.flatGroupItems()
	if items != nil {
		t.Errorf("expected nil, got %v", items)
	}
}

func TestFlatGroupItems_NoGroups(t *testing.T) {
	m := testModel()
	m.state = &protocol.RetroState{
		Cards: []protocol.Card{
			{ID: "c1", ColumnID: "stop", Text: "a"},
			{ID: "c2", ColumnID: "stop", Text: "b"},
		},
	}
	items := m.flatGroupItems()

	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	for _, item := range items {
		if item.kind != "card" {
			t.Error("all items should be cards")
		}
		if item.grouped {
			t.Error("no items should be grouped")
		}
	}
}

// --- viewGroup ---

func TestViewGroup_ShowsGroupHeader(t *testing.T) {
	m := testGroupModel()
	view := m.viewGroup()

	if !strings.Contains(view, "Process Issues") {
		t.Error("expected group name in view")
	}
}

func TestViewGroup_ShowsGroupedCards(t *testing.T) {
	m := testGroupModel()
	view := m.viewGroup()

	if !strings.Contains(view, "long meetings") {
		t.Error("expected grouped card in view")
	}
}

func TestViewGroup_ShowsUngroupedCards(t *testing.T) {
	m := testGroupModel()
	view := m.viewGroup()

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
	if !strings.Contains(view, "ungroup") {
		t.Error("expected ungroup in help")
	}
	if !strings.Contains(view, "rename") {
		t.Error("expected rename in help")
	}
}

func TestViewGroup_MergeMode(t *testing.T) {
	m := testGroupModel()
	m.mergeSource = "c3"
	view := m.viewGroup()

	if !strings.Contains(view, "MERGE") {
		t.Error("expected MERGE MODE indicator")
	}
}

func TestViewGroup_RenameInput(t *testing.T) {
	m := testGroupModel()
	m.inputMode = true
	m.inputText = "New Name"
	view := m.viewGroup()

	if !strings.Contains(view, "Rename") {
		t.Error("expected rename prompt")
	}
	if !strings.Contains(view, "New Name") {
		t.Error("expected input text")
	}
}

func TestViewGroup_NilState(t *testing.T) {
	m := testModel()
	if view := m.viewGroup(); view != "" {
		t.Errorf("expected empty, got %q", view)
	}
}

func TestViewGroup_CursorHighlight(t *testing.T) {
	m := testGroupModel()
	m.cursor = 0 // group header
	view := m.viewGroup()

	if !strings.Contains(view, ">") {
		t.Error("expected cursor indicator")
	}
}

// --- handleGroupKeys ---

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
	m.cursor = 0

	result, _ := m.handleGroupKeys(keyMsg("down"))
	model := result.(Model)

	if model.cursor != 1 {
		t.Errorf("expected 1, got %d", model.cursor)
	}
}

func TestHandleGroupKeys_UpAtTop(t *testing.T) {
	m := testGroupModel()
	m.cursor = 0

	result, _ := m.handleGroupKeys(keyMsg("up"))
	model := result.(Model)

	if model.cursor != 0 {
		t.Errorf("should stay at 0, got %d", model.cursor)
	}
}

func TestHandleGroupKeys_DownAtBottom(t *testing.T) {
	m := testGroupModel()
	items := m.flatGroupItems()
	m.cursor = len(items) - 1

	result, _ := m.handleGroupKeys(keyMsg("down"))
	model := result.(Model)

	if model.cursor != len(items)-1 {
		t.Errorf("should stay at bottom, got %d", model.cursor)
	}
}

func TestHandleGroupKeys_MergeFirstSelect(t *testing.T) {
	m := testGroupModel()
	m.cursor = 3 // c3, ungrouped card

	result, _ := m.handleGroupKeys(keyMsg("m"))
	model := result.(Model)

	if model.mergeSource != "c3" {
		t.Errorf("expected merge source 'c3', got %q", model.mergeSource)
	}
}

func TestHandleGroupKeys_MergeSecondSelect(t *testing.T) {
	m := testGroupModel()
	m.mergeSource = "c3"
	m.cursor = 4 // c4

	result, _ := m.handleGroupKeys(keyMsg("m"))
	model := result.(Model)

	if model.mergeSource != "" {
		t.Error("merge source should be cleared after merge")
	}
	if len(model.state.Groups) != 2 {
		t.Errorf("expected new group, got %d groups", len(model.state.Groups))
	}
}

func TestHandleGroupKeys_MergeSameCard(t *testing.T) {
	m := testGroupModel()
	m.mergeSource = "c3"
	m.cursor = 3 // same card

	result, _ := m.handleGroupKeys(keyMsg("m"))
	model := result.(Model)

	if model.mergeSource != "c3" {
		t.Error("should not merge card with itself")
	}
}

func TestHandleGroupKeys_MergeOnGroupHeader(t *testing.T) {
	m := testGroupModel()
	m.cursor = 0 // group header

	result, _ := m.handleGroupKeys(keyMsg("m"))
	model := result.(Model)

	if model.mergeSource != "" {
		t.Error("should not select group header for merge")
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

func TestHandleGroupKeys_Ungroup(t *testing.T) {
	m := testGroupModel()
	m.cursor = 1 // c1, in group g1

	result, _ := m.handleGroupKeys(keyMsg("u"))
	model := result.(Model)

	// g1 had 2 cards, removing 1 leaves <2, group deleted
	if len(model.state.Groups) != 0 {
		t.Errorf("expected group deleted, got %d", len(model.state.Groups))
	}
}

func TestHandleGroupKeys_UngroupUngroupedCard(t *testing.T) {
	m := testGroupModel()
	m.cursor = 3 // c3, not grouped

	result, _ := m.handleGroupKeys(keyMsg("u"))
	model := result.(Model)

	if len(model.state.Groups) != 1 {
		t.Error("should not affect groups when ungrouping an ungrouped card")
	}
}

func TestHandleGroupKeys_Rename(t *testing.T) {
	m := testGroupModel()
	m.cursor = 0 // group header

	result, _ := m.handleGroupKeys(keyMsg("r"))
	model := result.(Model)

	if !model.inputMode {
		t.Error("r on group header should enter rename mode")
	}
}

func TestHandleGroupKeys_RenameOnCard(t *testing.T) {
	m := testGroupModel()
	m.cursor = 1 // card, not group header

	result, _ := m.handleGroupKeys(keyMsg("r"))
	model := result.(Model)

	if model.inputMode {
		t.Error("r on card should not enter rename mode")
	}
}

func TestHandleGroupKeys_VimK(t *testing.T) {
	m := testGroupModel()
	m.cursor = 2

	result, _ := m.handleGroupKeys(keyMsg("k"))
	model := result.(Model)

	if model.cursor != 1 {
		t.Errorf("expected 1, got %d", model.cursor)
	}
}

func TestHandleGroupKeys_VimJ(t *testing.T) {
	m := testGroupModel()
	m.cursor = 0

	result, _ := m.handleGroupKeys(keyMsg("j"))
	model := result.(Model)

	if model.cursor != 1 {
		t.Errorf("expected 1, got %d", model.cursor)
	}
}

// --- rename input ---

func TestHandleGroupRenameInput_Enter(t *testing.T) {
	m := testGroupModel()
	m.inputMode = true
	m.inputText = "New Name"
	m.cursor = 0 // group header

	result, _ := m.handleGroupRenameInput(keyMsg("enter"))
	model := result.(Model)

	if model.inputMode {
		t.Error("should exit input mode")
	}
	if model.state.Groups[0].Name != "New Name" {
		t.Errorf("expected 'New Name', got %q", model.state.Groups[0].Name)
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
		t.Error("empty name should not rename")
	}
}

func TestHandleGroupRenameInput_Escape(t *testing.T) {
	m := testGroupModel()
	m.inputMode = true
	m.inputText = "partial"

	result, _ := m.handleGroupRenameInput(keyMsg("esc"))
	model := result.(Model)

	if model.inputMode {
		t.Error("escape should exit rename")
	}
}

func TestHandleGroupRenameInput_Backspace(t *testing.T) {
	m := testGroupModel()
	m.inputMode = true
	m.inputText = "abc"

	result, _ := m.handleGroupRenameInput(keyMsg("backspace"))
	model := result.(Model)

	if model.inputText != "ab" {
		t.Errorf("expected 'ab', got %q", model.inputText)
	}
}

func TestHandleGroupRenameInput_TypeChar(t *testing.T) {
	m := testGroupModel()
	m.inputMode = true
	m.inputText = "Ne"

	result, _ := m.handleGroupRenameInput(keyMsg("w"))
	model := result.(Model)

	if model.inputText != "New" {
		t.Errorf("expected 'New', got %q", model.inputText)
	}
}

// --- merge/ungroup/rename logic ---

func TestMergeCards_CreateNewGroup(t *testing.T) {
	m := testGroupModel()
	m.mergeCards("c3", "c4")

	if len(m.state.Groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(m.state.Groups))
	}
}

func TestMergeCards_AddToExistingGroup(t *testing.T) {
	m := testGroupModel()
	m.mergeCards("c3", "c1") // c1 is in g1

	if len(m.state.Groups[0].CardIDs) != 3 {
		t.Errorf("expected 3 cards, got %d", len(m.state.Groups[0].CardIDs))
	}
}

func TestMergeCards_SourceAlreadyGrouped(t *testing.T) {
	m := testGroupModel()
	m.mergeCards("c1", "c3") // c1 already in group

	if len(m.state.Groups) != 1 {
		t.Error("should not create new group")
	}
}

func TestMergeCards_NilState(t *testing.T) {
	m := testModel()
	m.mergeCards("c1", "c2") // no panic
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
	m.renameGroupByID("nonexistent", "X")
	if m.state.Groups[0].Name != "Process Issues" {
		t.Error("should not change")
	}
}

func TestRenameGroupByID_NilState(t *testing.T) {
	m := testModel()
	m.renameGroupByID("g1", "X") // no panic
}

func TestUngroupCardByID_DeletesGroup(t *testing.T) {
	m := testGroupModel()
	m.ungroupCardByID("c1")

	if len(m.state.Groups) != 0 {
		t.Errorf("expected 0 groups, got %d", len(m.state.Groups))
	}
}

func TestUngroupCardByID_GroupSurvives(t *testing.T) {
	m := testGroupModel()
	m.state.Groups[0].CardIDs = []string{"c1", "c2", "c3"}
	m.ungroupCardByID("c1")

	if len(m.state.Groups) != 1 {
		t.Fatal("group should survive")
	}
	if len(m.state.Groups[0].CardIDs) != 2 {
		t.Errorf("expected 2 cards, got %d", len(m.state.Groups[0].CardIDs))
	}
}

func TestUngroupCardByID_NotInGroup(t *testing.T) {
	m := testGroupModel()
	m.ungroupCardByID("c3")
	if len(m.state.Groups) != 1 {
		t.Error("should not change")
	}
}

func TestUngroupCardByID_NilState(t *testing.T) {
	m := testModel()
	m.ungroupCardByID("c1") // no panic
}

func TestUngroupedCardsForColumn(t *testing.T) {
	m := testGroupModel()
	ungrouped := m.ungroupedCardsForColumn("stop")

	if len(ungrouped) != 1 || ungrouped[0].ID != "c3" {
		t.Errorf("expected [c3], got %v", ungrouped)
	}
}

func TestViewGroup_InAppView(t *testing.T) {
	m := testGroupModel()
	view := m.View()

	if !strings.Contains(view, "GROUP") {
		t.Error("expected GROUP in header")
	}
}

func TestHandleGroupKeys_DelegatesToRenameInput(t *testing.T) {
	m := testGroupModel()
	m.inputMode = true
	m.inputText = "x"

	result, _ := m.handleGroupKeys(keyMsg("y"))
	model := result.(Model)

	if model.inputText != "xy" {
		t.Errorf("expected 'xy', got %q", model.inputText)
	}
}

func TestBroadcastState_NilClient(t *testing.T) {
	m := testGroupModel()
	m.broadcastState() // no panic
}
