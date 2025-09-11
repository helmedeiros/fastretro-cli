package storage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/helmedeiros/fastretro-cli/internal/domain"
)

func TestJSONTeamRepo_LoadTeam_NotExists(t *testing.T) {
	dir := t.TempDir()
	repo := NewJSONTeamRepo(dir)

	state, err := repo.LoadTeam()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(state.Members) != 0 {
		t.Errorf("expected 0 members, got %d", len(state.Members))
	}
}

func TestJSONTeamRepo_SaveAndLoadTeam(t *testing.T) {
	dir := t.TempDir()
	repo := NewJSONTeamRepo(dir)

	team := domain.NewTeam()
	team, _ = domain.AddMember(team, "m1", "Alice")
	team, _ = domain.AddAgreement(team, "a1", "Demo Fridays", "2025-09-01")

	if err := repo.SaveTeam(team); err != nil {
		t.Fatalf("save error: %v", err)
	}

	loaded, err := repo.LoadTeam()
	if err != nil {
		t.Fatalf("load error: %v", err)
	}
	if len(loaded.Members) != 1 {
		t.Errorf("expected 1 member, got %d", len(loaded.Members))
	}
	if loaded.Members[0].Name != "Alice" {
		t.Errorf("got %q", loaded.Members[0].Name)
	}
	if len(loaded.Agreements) != 1 {
		t.Errorf("expected 1 agreement, got %d", len(loaded.Agreements))
	}
}

func TestJSONTeamRepo_SaveCreatesDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "team")
	repo := NewJSONTeamRepo(dir)

	if err := repo.SaveTeam(domain.NewTeam()); err != nil {
		t.Fatalf("save error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "team.json")); err != nil {
		t.Errorf("file should exist: %v", err)
	}
}

func TestJSONTeamRepo_LoadHistory_NotExists(t *testing.T) {
	dir := t.TempDir()
	repo := NewJSONTeamRepo(dir)

	state, err := repo.LoadHistory()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(state.Completed) != 0 {
		t.Errorf("expected 0, got %d", len(state.Completed))
	}
}

func TestJSONTeamRepo_SaveAndLoadHistory(t *testing.T) {
	dir := t.TempDir()
	repo := NewJSONTeamRepo(dir)

	h := domain.NewHistory()
	h = domain.AddCompletedRetro(h, domain.CompletedRetro{
		ID:          "r1",
		CompletedAt: "2025-09-07",
		ActionItems: []domain.FlatActionItem{
			{NoteID: "n1", Text: "Fix bug", Done: false},
		},
	})

	if err := repo.SaveHistory(h); err != nil {
		t.Fatalf("save error: %v", err)
	}

	loaded, err := repo.LoadHistory()
	if err != nil {
		t.Fatalf("load error: %v", err)
	}
	if len(loaded.Completed) != 1 {
		t.Fatalf("expected 1, got %d", len(loaded.Completed))
	}
	if len(loaded.Completed[0].ActionItems) != 1 {
		t.Errorf("expected 1 action item, got %d", len(loaded.Completed[0].ActionItems))
	}
}

func TestJSONTeamRepo_LoadTeam_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "team.json"), []byte("not json"), 0644)

	repo := NewJSONTeamRepo(dir)
	_, err := repo.LoadTeam()
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestJSONTeamRepo_LoadHistory_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "history.json"), []byte("not json"), 0644)

	repo := NewJSONTeamRepo(dir)
	_, err := repo.LoadHistory()
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestJSONTeamRepo_LoadTeam_NullArrays(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "team.json"), []byte(`{}`), 0644)

	repo := NewJSONTeamRepo(dir)
	state, err := repo.LoadTeam()
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if state.Members == nil {
		t.Error("members should not be nil")
	}
	if state.Agreements == nil {
		t.Error("agreements should not be nil")
	}
}

func TestJSONTeamRepo_LoadHistory_NullArray(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "history.json"), []byte(`{}`), 0644)

	repo := NewJSONTeamRepo(dir)
	state, err := repo.LoadHistory()
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if state.Completed == nil {
		t.Error("completed should not be nil")
	}
}

func TestJSONTeamRepo_AtomicWrite(t *testing.T) {
	dir := t.TempDir()
	repo := NewJSONTeamRepo(dir)

	repo.SaveTeam(domain.NewTeam())

	// Verify no .tmp file left behind
	matches, _ := filepath.Glob(filepath.Join(dir, "*.tmp"))
	if len(matches) != 0 {
		t.Errorf("temp file should be cleaned up, found: %v", matches)
	}
}
