package domain

import "testing"

func TestNewTeam(t *testing.T) {
	team := NewTeam()
	if len(team.Members) != 0 {
		t.Errorf("expected 0 members, got %d", len(team.Members))
	}
	if len(team.Agreements) != 0 {
		t.Errorf("expected 0 agreements, got %d", len(team.Agreements))
	}
}

// --- AddMember ---

func TestAddMember(t *testing.T) {
	team := NewTeam()
	result, err := AddMember(team, "m1", "Alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Members) != 1 {
		t.Fatalf("expected 1 member, got %d", len(result.Members))
	}
	if result.Members[0].Name != "Alice" {
		t.Errorf("expected Alice, got %q", result.Members[0].Name)
	}
}

func TestAddMember_Multiple(t *testing.T) {
	team := NewTeam()
	team, _ = AddMember(team, "m1", "Alice")
	team, _ = AddMember(team, "m2", "Bob")
	if len(team.Members) != 2 {
		t.Errorf("expected 2 members, got %d", len(team.Members))
	}
}

func TestAddMember_EmptyName(t *testing.T) {
	team := NewTeam()
	_, err := AddMember(team, "m1", "")
	if err != ErrEmptyName {
		t.Errorf("expected ErrEmptyName, got %v", err)
	}
}

func TestAddMember_WhitespaceName(t *testing.T) {
	team := NewTeam()
	_, err := AddMember(team, "m1", "   ")
	if err != ErrEmptyName {
		t.Errorf("expected ErrEmptyName, got %v", err)
	}
}

func TestAddMember_DuplicateName(t *testing.T) {
	team := NewTeam()
	team, _ = AddMember(team, "m1", "Alice")
	_, err := AddMember(team, "m2", "Alice")
	if err != ErrDuplicateName {
		t.Errorf("expected ErrDuplicateName, got %v", err)
	}
}

func TestAddMember_DuplicateNameCaseInsensitive(t *testing.T) {
	team := NewTeam()
	team, _ = AddMember(team, "m1", "Alice")
	_, err := AddMember(team, "m2", "alice")
	if err != ErrDuplicateName {
		t.Errorf("expected ErrDuplicateName, got %v", err)
	}
}

func TestAddMember_TrimName(t *testing.T) {
	team := NewTeam()
	result, _ := AddMember(team, "m1", "  Alice  ")
	if result.Members[0].Name != "Alice" {
		t.Errorf("expected trimmed name, got %q", result.Members[0].Name)
	}
}

func TestAddMember_Immutable(t *testing.T) {
	team := NewTeam()
	team, _ = AddMember(team, "m1", "Alice")
	result, _ := AddMember(team, "m2", "Bob")
	if len(team.Members) != 1 {
		t.Error("original should not be modified")
	}
	if len(result.Members) != 2 {
		t.Error("result should have 2 members")
	}
}

func TestAddMember_PreservesAgreements(t *testing.T) {
	team := NewTeam()
	team, _ = AddAgreement(team, "a1", "test", "2025-01-01")
	result, _ := AddMember(team, "m1", "Alice")
	if len(result.Agreements) != 1 {
		t.Error("agreements should be preserved")
	}
}

// --- RemoveMember ---

func TestRemoveMember(t *testing.T) {
	team := NewTeam()
	team, _ = AddMember(team, "m1", "Alice")
	team, _ = AddMember(team, "m2", "Bob")
	result, err := RemoveMember(team, "m1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Members) != 1 {
		t.Fatalf("expected 1 member, got %d", len(result.Members))
	}
	if result.Members[0].ID != "m2" {
		t.Errorf("expected m2, got %q", result.Members[0].ID)
	}
}

func TestRemoveMember_NotFound(t *testing.T) {
	team := NewTeam()
	_, err := RemoveMember(team, "nonexistent")
	if err != ErrMemberNotFound {
		t.Errorf("expected ErrMemberNotFound, got %v", err)
	}
}

func TestRemoveMember_LastMember(t *testing.T) {
	team := NewTeam()
	team, _ = AddMember(team, "m1", "Alice")
	result, _ := RemoveMember(team, "m1")
	if len(result.Members) != 0 {
		t.Errorf("expected 0 members, got %d", len(result.Members))
	}
}

func TestRemoveMember_PreservesAgreements(t *testing.T) {
	team := NewTeam()
	team, _ = AddMember(team, "m1", "Alice")
	team, _ = AddAgreement(team, "a1", "test", "2025-01-01")
	result, _ := RemoveMember(team, "m1")
	if len(result.Agreements) != 1 {
		t.Error("agreements should be preserved")
	}
}

// --- AddAgreement ---

func TestAddAgreement_DuplicateID(t *testing.T) {
	team := NewTeam()
	team, _ = AddAgreement(team, "a1", "first", "2025-01-01")
	_, err := AddAgreement(team, "a1", "second", "2025-01-01")
	if err != ErrDuplicateID {
		t.Errorf("expected ErrDuplicateID, got %v", err)
	}
}

func TestAddAgreement(t *testing.T) {
	team := NewTeam()
	result, err := AddAgreement(team, "a1", "We demo every Friday", "2025-09-01")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Agreements) != 1 {
		t.Fatalf("expected 1, got %d", len(result.Agreements))
	}
	if result.Agreements[0].Text != "We demo every Friday" {
		t.Errorf("got %q", result.Agreements[0].Text)
	}
}

func TestAddAgreement_EmptyText(t *testing.T) {
	team := NewTeam()
	_, err := AddAgreement(team, "a1", "", "2025-09-01")
	if err != ErrEmptyName {
		t.Errorf("expected ErrEmptyName, got %v", err)
	}
}

func TestAddAgreement_TrimText(t *testing.T) {
	team := NewTeam()
	result, _ := AddAgreement(team, "a1", "  trimmed  ", "2025-09-01")
	if result.Agreements[0].Text != "trimmed" {
		t.Errorf("got %q", result.Agreements[0].Text)
	}
}

func TestAddAgreement_PreservesMembers(t *testing.T) {
	team := NewTeam()
	team, _ = AddMember(team, "m1", "Alice")
	result, _ := AddAgreement(team, "a1", "test", "2025-01-01")
	if len(result.Members) != 1 {
		t.Error("members should be preserved")
	}
}

// --- EditAgreement ---

func TestEditAgreement(t *testing.T) {
	team := NewTeam()
	team, _ = AddAgreement(team, "a1", "old", "2025-01-01")
	result, err := EditAgreement(team, "a1", "new text")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Agreements[0].Text != "new text" {
		t.Errorf("got %q", result.Agreements[0].Text)
	}
}

func TestEditAgreement_NotFound(t *testing.T) {
	team := NewTeam()
	_, err := EditAgreement(team, "nonexistent", "text")
	if err != ErrAgreementNotFound {
		t.Errorf("expected ErrAgreementNotFound, got %v", err)
	}
}

func TestEditAgreement_EmptyText(t *testing.T) {
	team := NewTeam()
	team, _ = AddAgreement(team, "a1", "old", "2025-01-01")
	_, err := EditAgreement(team, "a1", "")
	if err != ErrEmptyName {
		t.Errorf("expected ErrEmptyName, got %v", err)
	}
}

func TestEditAgreement_Immutable(t *testing.T) {
	team := NewTeam()
	team, _ = AddAgreement(team, "a1", "old", "2025-01-01")
	result, _ := EditAgreement(team, "a1", "new")
	if team.Agreements[0].Text != "old" {
		t.Error("original should not be modified")
	}
	if result.Agreements[0].Text != "new" {
		t.Error("result should have new text")
	}
}

// --- RemoveAgreement ---

func TestRemoveAgreement(t *testing.T) {
	team := NewTeam()
	team, _ = AddAgreement(team, "a1", "one", "2025-01-01")
	team, _ = AddAgreement(team, "a2", "two", "2025-01-01")
	result := RemoveAgreement(team, "a1")
	if len(result.Agreements) != 1 {
		t.Fatalf("expected 1, got %d", len(result.Agreements))
	}
	if result.Agreements[0].ID != "a2" {
		t.Errorf("expected a2, got %q", result.Agreements[0].ID)
	}
}

func TestRemoveAgreement_NotFound(t *testing.T) {
	team := NewTeam()
	result := RemoveAgreement(team, "nonexistent")
	if len(result.Agreements) != 0 {
		t.Error("should return empty agreements")
	}
}

func TestRemoveAgreement_LastAgreement(t *testing.T) {
	team := NewTeam()
	team, _ = AddAgreement(team, "a1", "one", "2025-01-01")
	result := RemoveAgreement(team, "a1")
	if len(result.Agreements) != 0 {
		t.Errorf("expected 0, got %d", len(result.Agreements))
	}
}

func TestRemoveAgreement_PreservesMembers(t *testing.T) {
	team := NewTeam()
	team, _ = AddMember(team, "m1", "Alice")
	team, _ = AddAgreement(team, "a1", "test", "2025-01-01")
	result := RemoveAgreement(team, "a1")
	if len(result.Members) != 1 {
		t.Error("members should be preserved")
	}
}
