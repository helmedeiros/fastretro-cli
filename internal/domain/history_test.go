package domain

import "testing"

func testHistory() RetroHistoryState {
	return AddCompletedRetro(NewHistory(), CompletedRetro{
		ID:          "r1",
		CompletedAt: "2025-09-07",
		ActionItems: []FlatActionItem{
			{NoteID: "n1", Text: "Fix login bug", ParentText: "Auth issues", OwnerName: "Bob", Done: false},
			{NoteID: "n2", Text: "Update CI", ParentText: "DevOps", OwnerName: "Alice", Done: true},
			{NoteID: "n3", Text: "Write docs", ParentText: "Onboarding", Done: false},
		},
	})
}

func TestNewHistory(t *testing.T) {
	h := NewHistory()
	if len(h.Completed) != 0 {
		t.Errorf("expected 0, got %d", len(h.Completed))
	}
}

func TestAddCompletedRetro(t *testing.T) {
	h := NewHistory()
	entry := CompletedRetro{ID: "r1", CompletedAt: "2025-09-07"}
	result := AddCompletedRetro(h, entry)
	if len(result.Completed) != 1 {
		t.Fatalf("expected 1, got %d", len(result.Completed))
	}
	if len(h.Completed) != 0 {
		t.Error("original should not be modified")
	}
}

func TestAddCompletedRetro_Multiple(t *testing.T) {
	h := NewHistory()
	h = AddCompletedRetro(h, CompletedRetro{ID: "r1"})
	h = AddCompletedRetro(h, CompletedRetro{ID: "r2"})
	if len(h.Completed) != 2 {
		t.Errorf("expected 2, got %d", len(h.Completed))
	}
}

func TestGetAllActionItems(t *testing.T) {
	h := testHistory()
	items := GetAllActionItems(h)
	if len(items) != 3 {
		t.Fatalf("expected 3, got %d", len(items))
	}
}

func TestGetAllActionItems_Empty(t *testing.T) {
	h := NewHistory()
	items := GetAllActionItems(h)
	if len(items) != 0 {
		t.Errorf("expected 0, got %d", len(items))
	}
}

func TestGetAllActionItems_MultipleRetros(t *testing.T) {
	h := testHistory()
	h = AddCompletedRetro(h, CompletedRetro{
		ID: "r2",
		ActionItems: []FlatActionItem{
			{NoteID: "n4", Text: "Extra item"},
		},
	})
	items := GetAllActionItems(h)
	if len(items) != 4 {
		t.Errorf("expected 4, got %d", len(items))
	}
}

func TestGetOpenActionItems(t *testing.T) {
	h := testHistory()
	items := GetOpenActionItems(h)
	if len(items) != 2 {
		t.Fatalf("expected 2 open items, got %d", len(items))
	}
	for _, item := range items {
		if item.Done {
			t.Error("should only return open items")
		}
	}
}

func TestGetOpenActionItems_AllDone(t *testing.T) {
	h := AddCompletedRetro(NewHistory(), CompletedRetro{
		ID: "r1",
		ActionItems: []FlatActionItem{
			{NoteID: "n1", Done: true},
		},
	})
	items := GetOpenActionItems(h)
	if len(items) != 0 {
		t.Errorf("expected 0, got %d", len(items))
	}
}

func TestToggleActionItemDone(t *testing.T) {
	h := testHistory()
	result := ToggleActionItemDone(h, "n1")
	items := GetAllActionItems(result)
	for _, item := range items {
		if item.NoteID == "n1" && !item.Done {
			t.Error("n1 should be done")
		}
	}
	// Original unchanged
	orig := GetAllActionItems(h)
	for _, item := range orig {
		if item.NoteID == "n1" && item.Done {
			t.Error("original should not be modified")
		}
	}
}

func TestToggleActionItemDone_BackToOpen(t *testing.T) {
	h := testHistory()
	result := ToggleActionItemDone(h, "n2") // n2 is done=true
	items := GetAllActionItems(result)
	for _, item := range items {
		if item.NoteID == "n2" && item.Done {
			t.Error("n2 should be toggled to open")
		}
	}
}

func TestToggleActionItemDone_NotFound(t *testing.T) {
	h := testHistory()
	result := ToggleActionItemDone(h, "nonexistent")
	if len(GetAllActionItems(result)) != 3 {
		t.Error("should not change anything")
	}
}

func TestReassignActionItem(t *testing.T) {
	h := testHistory()
	result := ReassignActionItem(h, "n1", "Carol")
	items := GetAllActionItems(result)
	for _, item := range items {
		if item.NoteID == "n1" && item.OwnerName != "Carol" {
			t.Errorf("expected Carol, got %q", item.OwnerName)
		}
	}
}

func TestReassignActionItem_ClearOwner(t *testing.T) {
	h := testHistory()
	result := ReassignActionItem(h, "n1", "")
	items := GetAllActionItems(result)
	for _, item := range items {
		if item.NoteID == "n1" && item.OwnerName != "" {
			t.Errorf("expected empty owner, got %q", item.OwnerName)
		}
	}
}

func TestEditActionItemText(t *testing.T) {
	h := testHistory()
	result := EditActionItemText(h, "n1", "New text")
	items := GetAllActionItems(result)
	for _, item := range items {
		if item.NoteID == "n1" && item.Text != "New text" {
			t.Errorf("expected 'New text', got %q", item.Text)
		}
	}
}

func TestRemoveActionItem(t *testing.T) {
	h := testHistory()
	result := RemoveActionItem(h, "n1")
	items := GetAllActionItems(result)
	if len(items) != 2 {
		t.Fatalf("expected 2, got %d", len(items))
	}
	for _, item := range items {
		if item.NoteID == "n1" {
			t.Error("n1 should be removed")
		}
	}
}

func TestRemoveActionItem_NotFound(t *testing.T) {
	h := testHistory()
	result := RemoveActionItem(h, "nonexistent")
	if len(GetAllActionItems(result)) != 3 {
		t.Error("should not change anything")
	}
}

func TestAddManualActionItem(t *testing.T) {
	h := NewHistory()
	item := FlatActionItem{NoteID: "manual-1", Text: "Manual task", CompletedAt: "2025-09-10"}
	result := AddManualActionItem(h, item)

	items := GetAllActionItems(result)
	if len(items) != 1 {
		t.Fatalf("expected 1, got %d", len(items))
	}
	if items[0].Text != "Manual task" {
		t.Errorf("got %q", items[0].Text)
	}
}

func TestAddManualActionItem_AppendsToExisting(t *testing.T) {
	h := NewHistory()
	h = AddManualActionItem(h, FlatActionItem{NoteID: "m1", Text: "First", CompletedAt: "2025-09-10"})
	h = AddManualActionItem(h, FlatActionItem{NoteID: "m2", Text: "Second", CompletedAt: "2025-09-10"})

	items := GetAllActionItems(h)
	if len(items) != 2 {
		t.Fatalf("expected 2, got %d", len(items))
	}
	// Should still be 1 completed retro (the manual one)
	if len(h.Completed) != 1 {
		t.Errorf("expected 1 retro entry, got %d", len(h.Completed))
	}
}

func TestAddManualActionItem_WithExistingRetros(t *testing.T) {
	h := testHistory()
	h = AddManualActionItem(h, FlatActionItem{NoteID: "m1", Text: "Manual", CompletedAt: "2025-09-10"})

	if len(h.Completed) != 2 {
		t.Errorf("expected 2 retro entries, got %d", len(h.Completed))
	}
	items := GetAllActionItems(h)
	if len(items) != 4 {
		t.Errorf("expected 4 total items, got %d", len(items))
	}
}
