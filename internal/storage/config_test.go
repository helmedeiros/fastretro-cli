package storage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/helmedeiros/fastretro-cli/internal/domain"
)

func TestJSONRegistryRepo_List_NotExists(t *testing.T) {
	dir := t.TempDir()
	repo := NewJSONRegistryRepo(dir)

	entries, err := repo.List()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0, got %d", len(entries))
	}
}

func TestJSONRegistryRepo_SaveAndList(t *testing.T) {
	dir := t.TempDir()
	repo := NewJSONRegistryRepo(dir)

	entries := []domain.TeamEntry{
		{ID: "t1", Name: "Alpha", CreatedAt: "2025-09-01"},
		{ID: "t2", Name: "Beta", CreatedAt: "2025-09-02"},
	}
	if err := repo.Save(entries); err != nil {
		t.Fatalf("save error: %v", err)
	}

	loaded, err := repo.List()
	if err != nil {
		t.Fatalf("load error: %v", err)
	}
	if len(loaded) != 2 {
		t.Fatalf("expected 2, got %d", len(loaded))
	}
	if loaded[0].Name != "Alpha" {
		t.Errorf("got %q", loaded[0].Name)
	}
}

func TestJSONRegistryRepo_SaveCreatesDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested")
	repo := NewJSONRegistryRepo(dir)

	if err := repo.Save([]domain.TeamEntry{}); err != nil {
		t.Fatalf("save error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "teams", "registry.json")); err != nil {
		t.Errorf("file should exist: %v", err)
	}
}

func TestJSONRegistryRepo_List_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "teams"), 0755)
	os.WriteFile(filepath.Join(dir, "teams", "registry.json"), []byte("bad"), 0644)

	repo := NewJSONRegistryRepo(dir)
	_, err := repo.List()
	if err == nil {
		t.Error("expected error")
	}
}

func TestJSONRegistryRepo_List_NullArray(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "teams"), 0755)
	os.WriteFile(filepath.Join(dir, "teams", "registry.json"), []byte("null"), 0644)

	repo := NewJSONRegistryRepo(dir)
	entries, err := repo.List()
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if entries == nil {
		t.Error("should not be nil")
	}
}

func TestJSONRegistryRepo_SelectedTeamID_NotExists(t *testing.T) {
	dir := t.TempDir()
	repo := NewJSONRegistryRepo(dir)

	id, err := repo.SelectedTeamID()
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if id != "" {
		t.Errorf("expected empty, got %q", id)
	}
}

func TestJSONRegistryRepo_SetAndGetSelectedTeamID(t *testing.T) {
	dir := t.TempDir()
	repo := NewJSONRegistryRepo(dir)

	if err := repo.SetSelectedTeamID("t1"); err != nil {
		t.Fatalf("set error: %v", err)
	}

	id, err := repo.SelectedTeamID()
	if err != nil {
		t.Fatalf("get error: %v", err)
	}
	if id != "t1" {
		t.Errorf("expected t1, got %q", id)
	}
}

func TestJSONRegistryRepo_SetSelectedTeamID_CreatesDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested")
	repo := NewJSONRegistryRepo(dir)

	if err := repo.SetSelectedTeamID("t1"); err != nil {
		t.Fatalf("error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "config.json")); err != nil {
		t.Errorf("file should exist: %v", err)
	}
}

func TestJSONRegistryRepo_SelectedTeamID_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "config.json"), []byte("bad"), 0644)

	repo := NewJSONRegistryRepo(dir)
	_, err := repo.SelectedTeamID()
	if err == nil {
		t.Error("expected error")
	}
}

func TestJSONRegistryRepo_TeamDir(t *testing.T) {
	repo := NewJSONRegistryRepo("/base")
	dir := repo.TeamDir("abc-123")
	expected := filepath.Join("/base", "teams", "abc-123")
	if dir != expected {
		t.Errorf("expected %q, got %q", expected, dir)
	}
}

func TestJSONRegistryRepo_SaveAndLoadIdentity(t *testing.T) {
	dir := t.TempDir()
	repo := NewJSONRegistryRepo(dir)

	repo.SaveIdentity("ABC-DEF-123", "p-1")

	got := repo.LoadIdentity("ABC-DEF-123")
	if got != "p-1" {
		t.Errorf("expected p-1, got %q", got)
	}
}

func TestJSONRegistryRepo_LoadIdentity_WrongRoom(t *testing.T) {
	dir := t.TempDir()
	repo := NewJSONRegistryRepo(dir)

	repo.SaveIdentity("ABC-DEF-123", "p-1")

	got := repo.LoadIdentity("OTHER-ROOM")
	if got != "" {
		t.Errorf("expected empty for wrong room, got %q", got)
	}
}

func TestJSONRegistryRepo_LoadIdentity_NotExists(t *testing.T) {
	dir := t.TempDir()
	repo := NewJSONRegistryRepo(dir)

	got := repo.LoadIdentity("ANY-ROOM")
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

func TestJSONRegistryRepo_LoadIdentity_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "identity.json"), []byte("bad"), 0644)

	repo := NewJSONRegistryRepo(dir)
	got := repo.LoadIdentity("ANY-ROOM")
	if got != "" {
		t.Errorf("expected empty for invalid JSON, got %q", got)
	}
}

func TestJSONRegistryRepo_SaveIdentity_Overwrites(t *testing.T) {
	dir := t.TempDir()
	repo := NewJSONRegistryRepo(dir)

	repo.SaveIdentity("ROOM-1", "p-1")
	repo.SaveIdentity("ROOM-2", "p-2")

	if got := repo.LoadIdentity("ROOM-1"); got != "" {
		t.Errorf("old room should be gone, got %q", got)
	}
	if got := repo.LoadIdentity("ROOM-2"); got != "p-2" {
		t.Errorf("expected p-2, got %q", got)
	}
}

func TestJSONRegistryRepo_SaveIdentity_CreatesDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested")
	repo := NewJSONRegistryRepo(dir)

	repo.SaveIdentity("ROOM", "p-1")

	if _, err := os.Stat(filepath.Join(dir, "identity.json")); err != nil {
		t.Errorf("file should exist: %v", err)
	}
}

func TestJSONRegistryRepo_SaveAndLoadDefaultMember(t *testing.T) {
	dir := t.TempDir()
	repo := NewJSONRegistryRepo(dir)

	repo.SaveDefaultMember("Alice")

	got := repo.LoadDefaultMember()
	if got != "Alice" {
		t.Errorf("expected Alice, got %q", got)
	}
}

func TestJSONRegistryRepo_LoadDefaultMember_NotExists(t *testing.T) {
	dir := t.TempDir()
	repo := NewJSONRegistryRepo(dir)

	got := repo.LoadDefaultMember()
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

func TestJSONRegistryRepo_LoadDefaultMember_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "config.json"), []byte("bad"), 0644)

	repo := NewJSONRegistryRepo(dir)
	got := repo.LoadDefaultMember()
	if got != "" {
		t.Errorf("expected empty for invalid JSON, got %q", got)
	}
}

func TestJSONRegistryRepo_SaveDefaultMember_PreservesSelectedTeam(t *testing.T) {
	dir := t.TempDir()
	repo := NewJSONRegistryRepo(dir)

	if err := repo.SetSelectedTeamID("t1"); err != nil {
		t.Fatalf("set error: %v", err)
	}
	repo.SaveDefaultMember("Bob")

	id, err := repo.SelectedTeamID()
	if err != nil {
		t.Fatalf("get error: %v", err)
	}
	if id != "t1" {
		t.Errorf("selected team lost, got %q", id)
	}
	if got := repo.LoadDefaultMember(); got != "Bob" {
		t.Errorf("expected Bob, got %q", got)
	}
}

func TestJSONRegistryRepo_SaveDefaultMember_ClearsWithEmpty(t *testing.T) {
	dir := t.TempDir()
	repo := NewJSONRegistryRepo(dir)

	repo.SaveDefaultMember("Alice")
	repo.SaveDefaultMember("")

	got := repo.LoadDefaultMember()
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

func TestJSONRegistryRepo_SaveDefaultMember_CreatesDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested")
	repo := NewJSONRegistryRepo(dir)

	repo.SaveDefaultMember("Alice")

	if _, err := os.Stat(filepath.Join(dir, "config.json")); err != nil {
		t.Errorf("file should exist: %v", err)
	}
}
