package domain

import "testing"

func TestAddTeamEntry(t *testing.T) {
	entries, err := AddTeamEntry(nil, "t1", "Acme Squad", "2025-09-01")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1, got %d", len(entries))
	}
	if entries[0].Name != "Acme Squad" {
		t.Errorf("got %q", entries[0].Name)
	}
}

func TestAddTeamEntry_Multiple(t *testing.T) {
	entries, _ := AddTeamEntry(nil, "t1", "Alpha", "2025-09-01")
	entries, _ = AddTeamEntry(entries, "t2", "Beta", "2025-09-01")
	if len(entries) != 2 {
		t.Errorf("expected 2, got %d", len(entries))
	}
}

func TestAddTeamEntry_EmptyName(t *testing.T) {
	_, err := AddTeamEntry(nil, "t1", "", "2025-09-01")
	if err != ErrEmptyName {
		t.Errorf("expected ErrEmptyName, got %v", err)
	}
}

func TestAddTeamEntry_DuplicateName(t *testing.T) {
	entries, _ := AddTeamEntry(nil, "t1", "Alpha", "2025-09-01")
	_, err := AddTeamEntry(entries, "t2", "alpha", "2025-09-01")
	if err != ErrDuplicateName {
		t.Errorf("expected ErrDuplicateName, got %v", err)
	}
}

func TestAddTeamEntry_TrimName(t *testing.T) {
	entries, _ := AddTeamEntry(nil, "t1", "  Alpha  ", "2025-09-01")
	if entries[0].Name != "Alpha" {
		t.Errorf("got %q", entries[0].Name)
	}
}

func TestAddTeamEntry_Immutable(t *testing.T) {
	entries, _ := AddTeamEntry(nil, "t1", "Alpha", "2025-09-01")
	result, _ := AddTeamEntry(entries, "t2", "Beta", "2025-09-01")
	if len(entries) != 1 {
		t.Error("original should not be modified")
	}
	if len(result) != 2 {
		t.Error("result should have 2")
	}
}

func TestRemoveTeamEntry(t *testing.T) {
	entries, _ := AddTeamEntry(nil, "t1", "Alpha", "2025-09-01")
	entries, _ = AddTeamEntry(entries, "t2", "Beta", "2025-09-01")
	result := RemoveTeamEntry(entries, "t1")
	if len(result) != 1 {
		t.Fatalf("expected 1, got %d", len(result))
	}
	if result[0].ID != "t2" {
		t.Errorf("got %q", result[0].ID)
	}
}

func TestRemoveTeamEntry_NotFound(t *testing.T) {
	result := RemoveTeamEntry(nil, "nonexistent")
	if len(result) != 0 {
		t.Errorf("expected 0, got %d", len(result))
	}
}

func TestRemoveTeamEntry_Last(t *testing.T) {
	entries, _ := AddTeamEntry(nil, "t1", "Alpha", "2025-09-01")
	result := RemoveTeamEntry(entries, "t1")
	if len(result) != 0 {
		t.Errorf("expected 0, got %d", len(result))
	}
}

func TestRenameTeamEntry(t *testing.T) {
	entries, _ := AddTeamEntry(nil, "t1", "Alpha", "2025-09-01")
	result, err := RenameTeamEntry(entries, "t1", "Omega")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result[0].Name != "Omega" {
		t.Errorf("got %q", result[0].Name)
	}
}

func TestRenameTeamEntry_EmptyName(t *testing.T) {
	entries, _ := AddTeamEntry(nil, "t1", "Alpha", "2025-09-01")
	_, err := RenameTeamEntry(entries, "t1", "")
	if err != ErrEmptyName {
		t.Errorf("expected ErrEmptyName, got %v", err)
	}
}

func TestRenameTeamEntry_DuplicateName(t *testing.T) {
	entries, _ := AddTeamEntry(nil, "t1", "Alpha", "2025-09-01")
	entries, _ = AddTeamEntry(entries, "t2", "Beta", "2025-09-01")
	_, err := RenameTeamEntry(entries, "t1", "beta")
	if err != ErrDuplicateName {
		t.Errorf("expected ErrDuplicateName, got %v", err)
	}
}

func TestRenameTeamEntry_NotFound(t *testing.T) {
	_, err := RenameTeamEntry(nil, "nonexistent", "X")
	if err == nil {
		t.Error("expected error for not found")
	}
}

func TestRenameTeamEntry_SameNameAllowed(t *testing.T) {
	entries, _ := AddTeamEntry(nil, "t1", "Alpha", "2025-09-01")
	result, err := RenameTeamEntry(entries, "t1", "Alpha")
	if err != nil {
		t.Fatalf("renaming to same name should be allowed: %v", err)
	}
	if result[0].Name != "Alpha" {
		t.Errorf("got %q", result[0].Name)
	}
}

func TestRenameTeamEntry_Immutable(t *testing.T) {
	entries, _ := AddTeamEntry(nil, "t1", "Alpha", "2025-09-01")
	result, _ := RenameTeamEntry(entries, "t1", "Omega")
	if entries[0].Name != "Alpha" {
		t.Error("original should not be modified")
	}
	if result[0].Name != "Omega" {
		t.Error("result should have new name")
	}
}

func TestFindTeamEntryByName(t *testing.T) {
	entries, _ := AddTeamEntry(nil, "t1", "Alpha", "2025-09-01")
	entry, ok := FindTeamEntryByName(entries, "alpha")
	if !ok {
		t.Fatal("expected to find")
	}
	if entry.ID != "t1" {
		t.Errorf("got %q", entry.ID)
	}
}

func TestFindTeamEntryByName_NotFound(t *testing.T) {
	_, ok := FindTeamEntryByName(nil, "nope")
	if ok {
		t.Error("should not find")
	}
}
