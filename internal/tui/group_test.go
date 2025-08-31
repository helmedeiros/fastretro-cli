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

func TestViewGroup_ShowsGroups(t *testing.T) {
	m := testGroupModel()
	view := m.viewGroup()

	if !strings.Contains(view, "Process Issues") {
		t.Error("expected group name in view")
	}
	if !strings.Contains(view, "long meetings") {
		t.Error("expected grouped card in view")
	}
}

func TestViewGroup_ShowsUngrouped(t *testing.T) {
	m := testGroupModel()
	view := m.viewGroup()

	if !strings.Contains(view, "Ungrouped") {
		t.Error("expected 'Ungrouped' section")
	}
	if !strings.Contains(view, "unclear requirements") {
		t.Error("expected ungrouped card in view")
	}
}

func TestViewGroup_ShowsNumberedCards(t *testing.T) {
	m := testGroupModel()
	view := m.viewGroup()

	if !strings.Contains(view, "[1]") {
		t.Error("expected numbered card [1]")
	}
}

func TestViewGroup_ShowsCommands(t *testing.T) {
	m := testGroupModel()
	view := m.viewGroup()

	if !strings.Contains(view, "merge") {
		t.Error("expected merge command in help")
	}
	if !strings.Contains(view, "rename") {
		t.Error("expected rename command in help")
	}
	if !strings.Contains(view, "ungroup") {
		t.Error("expected ungroup command in help")
	}
}

func TestViewGroup_InputMode(t *testing.T) {
	m := testGroupModel()
	m.inputMode = true
	m.inputText = "m 1 2"
	view := m.viewGroup()

	if !strings.Contains(view, "m 1 2") {
		t.Error("expected input text in view")
	}
}

func TestViewGroup_NilState(t *testing.T) {
	m := testModel()
	view := m.viewGroup()

	if view != "" {
		t.Errorf("expected empty view, got %q", view)
	}
}

func TestViewGroup_NoCards(t *testing.T) {
	m := testModel()
	m.state = &protocol.RetroState{Stage: "group"}
	view := m.viewGroup()

	// Default columns shown with no cards — commands still available
	if !strings.Contains(view, "merge") {
		t.Error("expected commands in view even with no cards")
	}
}

func TestHandleGroupKeys_Colon(t *testing.T) {
	m := testGroupModel()

	result, _ := m.handleGroupKeys(keyMsg(":"))
	model := result.(Model)

	if !model.inputMode {
		t.Error("expected input mode on ':'")
	}
}

func TestHandleGroupKeys_OtherKey(t *testing.T) {
	m := testGroupModel()

	result, _ := m.handleGroupKeys(keyMsg("a"))
	model := result.(Model)

	if model.inputMode {
		t.Error("'a' should not activate input mode in group stage")
	}
}

func TestHandleGroupInput_Escape(t *testing.T) {
	m := testGroupModel()
	m.inputMode = true
	m.inputText = "partial"

	result, _ := m.handleGroupInput(keyMsg("esc"))
	model := result.(Model)

	if model.inputMode {
		t.Error("escape should exit input mode")
	}
}

func TestHandleGroupInput_Backspace(t *testing.T) {
	m := testGroupModel()
	m.inputMode = true
	m.inputText = "abc"

	result, _ := m.handleGroupInput(keyMsg("backspace"))
	model := result.(Model)

	if model.inputText != "ab" {
		t.Errorf("expected 'ab', got %q", model.inputText)
	}
}

func TestHandleGroupInput_TypeChar(t *testing.T) {
	m := testGroupModel()
	m.inputMode = true
	m.inputText = "m "

	result, _ := m.handleGroupInput(keyMsg("1"))
	model := result.(Model)

	if model.inputText != "m 1" {
		t.Errorf("expected 'm 1', got %q", model.inputText)
	}
}

func TestBuildCardIndex(t *testing.T) {
	m := testGroupModel()
	idx := m.buildCardIndex()

	if idx["c1"] != 1 {
		t.Errorf("c1 should be 1, got %d", idx["c1"])
	}
	if idx["c4"] != 4 {
		t.Errorf("c4 should be 4, got %d", idx["c4"])
	}
}

func TestBuildCardIndex_NilState(t *testing.T) {
	m := testModel()
	idx := m.buildCardIndex()

	if len(idx) != 0 {
		t.Errorf("expected empty map, got %v", idx)
	}
}

func TestCardIDByIndex(t *testing.T) {
	m := testGroupModel()

	if id := m.cardIDByIndex("1"); id != "c1" {
		t.Errorf("expected 'c1', got %q", id)
	}
	if id := m.cardIDByIndex("4"); id != "c4" {
		t.Errorf("expected 'c4', got %q", id)
	}
	if id := m.cardIDByIndex("99"); id != "" {
		t.Errorf("expected empty for out-of-range, got %q", id)
	}
	if id := m.cardIDByIndex("abc"); id != "" {
		t.Errorf("expected empty for non-numeric, got %q", id)
	}
}

func TestMergeCards_CreateNewGroup(t *testing.T) {
	m := testGroupModel()
	// c3 and c4 are ungrouped
	// But c3 is stop, c4 is start — use two ungrouped stop cards
	m.state.Cards = append(m.state.Cards, protocol.Card{ID: "c5", ColumnID: "stop", Text: "no retros"})

	m.mergeCards("c5", "c3")

	if len(m.state.Groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(m.state.Groups))
	}
	newGroup := m.state.Groups[1]
	if len(newGroup.CardIDs) != 2 {
		t.Errorf("expected 2 cards in new group, got %d", len(newGroup.CardIDs))
	}
}

func TestMergeCards_AddToExistingGroup(t *testing.T) {
	m := testGroupModel()
	// c3 is ungrouped, g1 contains c1 and c2
	m.mergeCards("c3", "c1") // c1 is in g1

	if len(m.state.Groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(m.state.Groups))
	}
	if len(m.state.Groups[0].CardIDs) != 3 {
		t.Errorf("expected 3 cards in group, got %d", len(m.state.Groups[0].CardIDs))
	}
}

func TestMergeCards_SourceAlreadyGrouped(t *testing.T) {
	m := testGroupModel()
	// c1 is in g1, should be no-op
	m.mergeCards("c1", "c3")

	if len(m.state.Groups) != 1 {
		t.Errorf("expected no new group, got %d groups", len(m.state.Groups))
	}
}

func TestMergeCards_NilState(t *testing.T) {
	m := testModel()
	m.mergeCards("c1", "c2") // should not panic
}

func TestRenameGroupByID(t *testing.T) {
	m := testGroupModel()
	m.renameGroupByID("g1", "New Name")

	if m.state.Groups[0].Name != "New Name" {
		t.Errorf("expected 'New Name', got %q", m.state.Groups[0].Name)
	}
}

func TestRenameGroupByID_NotFound(t *testing.T) {
	m := testGroupModel()
	m.renameGroupByID("nonexistent", "New Name")

	// Should not panic or change anything
	if m.state.Groups[0].Name != "Process Issues" {
		t.Error("should not change existing group")
	}
}

func TestRenameGroupByID_NilState(t *testing.T) {
	m := testModel()
	m.renameGroupByID("g1", "New") // should not panic
}

func TestUngroupCardByID(t *testing.T) {
	m := testGroupModel()
	// g1 has c1 and c2, ungrouping c1 should keep g1 with 1 card → delete group
	m.ungroupCardByID("c1")

	if len(m.state.Groups) != 0 {
		t.Errorf("expected group to be deleted (< 2 cards), got %d groups", len(m.state.Groups))
	}
}

func TestUngroupCardByID_GroupSurvives(t *testing.T) {
	m := testGroupModel()
	m.state.Groups[0].CardIDs = []string{"c1", "c2", "c3"}

	m.ungroupCardByID("c1")

	if len(m.state.Groups) != 1 {
		t.Fatalf("expected group to survive, got %d groups", len(m.state.Groups))
	}
	if len(m.state.Groups[0].CardIDs) != 2 {
		t.Errorf("expected 2 cards remaining, got %d", len(m.state.Groups[0].CardIDs))
	}
}

func TestUngroupCardByID_NotInGroup(t *testing.T) {
	m := testGroupModel()
	initialGroups := len(m.state.Groups)

	m.ungroupCardByID("c3") // not in any group

	if len(m.state.Groups) != initialGroups {
		t.Error("should not change groups for ungrouped card")
	}
}

func TestUngroupCardByID_NilState(t *testing.T) {
	m := testModel()
	m.ungroupCardByID("c1") // should not panic
}

func TestUngroupedCardsForColumn(t *testing.T) {
	m := testGroupModel()
	ungrouped := m.ungroupedCardsForColumn("stop")

	if len(ungrouped) != 1 {
		t.Fatalf("expected 1 ungrouped stop card, got %d", len(ungrouped))
	}
	if ungrouped[0].ID != "c3" {
		t.Errorf("expected c3, got %q", ungrouped[0].ID)
	}
}

func TestUngroupedCardsForColumn_AllGrouped(t *testing.T) {
	m := testGroupModel()
	m.state.Groups[0].CardIDs = []string{"c1", "c2", "c3"}
	ungrouped := m.ungroupedCardsForColumn("stop")

	if len(ungrouped) != 0 {
		t.Errorf("expected 0 ungrouped, got %d", len(ungrouped))
	}
}

func TestExecuteGroupCommand_Merge(t *testing.T) {
	m := testGroupModel()
	m.state.Cards = append(m.state.Cards, protocol.Card{ID: "c5", ColumnID: "stop", Text: "no retros"})

	m.executeGroupCommand("m 5 3")

	if len(m.state.Groups) != 2 {
		t.Errorf("expected 2 groups after merge, got %d", len(m.state.Groups))
	}
}

func TestExecuteGroupCommand_Rename(t *testing.T) {
	m := testGroupModel()
	m.executeGroupCommand("r g1 Better Name")

	if m.state.Groups[0].Name != "Better Name" {
		t.Errorf("expected 'Better Name', got %q", m.state.Groups[0].Name)
	}
}

func TestExecuteGroupCommand_Ungroup(t *testing.T) {
	m := testGroupModel()
	m.executeGroupCommand("u 1") // c1 is in g1

	if len(m.state.Groups) != 0 {
		t.Errorf("expected group deleted, got %d", len(m.state.Groups))
	}
}

func TestExecuteGroupCommand_Empty(t *testing.T) {
	m := testGroupModel()
	m.executeGroupCommand("") // should not panic
}

func TestExecuteGroupCommand_Unknown(t *testing.T) {
	m := testGroupModel()
	m.executeGroupCommand("x 1 2") // unknown command, no-op
	if len(m.state.Groups) != 1 {
		t.Error("unknown command should not change state")
	}
}

func TestExecuteGroupCommand_NilState(t *testing.T) {
	m := testModel()
	m.executeGroupCommand("m 1 2") // should not panic
}

func TestViewGroup_InAppView(t *testing.T) {
	m := testGroupModel()
	view := m.View()

	if !strings.Contains(view, "GROUP") {
		t.Error("expected 'GROUP' in header")
	}
}
