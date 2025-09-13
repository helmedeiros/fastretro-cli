package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/helmedeiros/fastretro-cli/internal/domain"
	"github.com/helmedeiros/fastretro-cli/internal/protocol"
	"github.com/helmedeiros/fastretro-cli/internal/storage"
)

func testShellModel(t *testing.T) ShellModel {
	t.Helper()
	dir := t.TempDir()
	reg := storage.NewJSONRegistryRepo(dir)
	entry := domain.TeamEntry{ID: "t1", Name: "Test Squad"}

	repo := storage.NewJSONTeamRepo(reg.TeamDir("t1"))
	team := domain.NewTeam()
	team, _ = domain.AddMember(team, "m1", "Alice")
	repo.SaveTeam(team)

	return NewShellModel(reg, entry, "http://localhost:5173")
}

// --- Mode switching ---

func TestShell_StartsInHomeMode(t *testing.T) {
	m := testShellModel(t)
	if m.mode != ModeHome {
		t.Errorf("expected ModeHome, got %d", m.mode)
	}
}

func TestShell_JKeyOpensJoinInput(t *testing.T) {
	m := testShellModel(t)
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	shell := result.(ShellModel)
	if shell.mode != ModeJoinInput {
		t.Errorf("expected ModeJoinInput, got %d", shell.mode)
	}
}

func TestShell_JoinInput_EscReturnsHome(t *testing.T) {
	m := testShellModel(t)
	m.mode = ModeJoinInput

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	shell := result.(ShellModel)
	if shell.mode != ModeHome {
		t.Errorf("expected ModeHome, got %d", shell.mode)
	}
}

func TestShell_JoinInput_TypesText(t *testing.T) {
	m := testShellModel(t)
	m.mode = ModeJoinInput

	for _, ch := range "ABC" {
		result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
		m = result.(ShellModel)
	}
	if m.joinInput != "ABC" {
		t.Errorf("expected 'ABC', got %q", m.joinInput)
	}
}

func TestShell_JoinInput_Backspace(t *testing.T) {
	m := testShellModel(t)
	m.mode = ModeJoinInput
	m.joinInput = "ABC"

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	shell := result.(ShellModel)
	if shell.joinInput != "AB" {
		t.Errorf("expected 'AB', got %q", shell.joinInput)
	}
}

func TestShell_JoinInput_EmptyEnterNoOp(t *testing.T) {
	m := testShellModel(t)
	m.mode = ModeJoinInput
	m.joinInput = ""

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	shell := result.(ShellModel)
	if shell.mode != ModeJoinInput {
		t.Error("should stay in join input on empty enter")
	}
}

func TestShell_JoinInput_ConnectError(t *testing.T) {
	m := testShellModel(t)
	m.mode = ModeJoinInput
	m.joinInput = "INVALID"
	m.serverURL = "http://localhost:1" // nothing running

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	shell := result.(ShellModel)
	if shell.joinErr == "" {
		t.Error("expected connection error")
	}
	if shell.mode != ModeJoinInput {
		t.Error("should stay in join input on error")
	}
}

// --- View ---

func TestShell_ViewHome(t *testing.T) {
	m := testShellModel(t)
	view := m.View()
	if !strings.Contains(view, "Test Squad") {
		t.Error("expected home view with team name")
	}
}

func TestShell_ViewJoinInput(t *testing.T) {
	m := testShellModel(t)
	m.mode = ModeJoinInput
	m.joinInput = "ABC-123"
	view := m.View()

	if !strings.Contains(view, "Join") {
		t.Error("expected join header")
	}
	if !strings.Contains(view, "ABC-123") {
		t.Error("expected input text")
	}
}

func TestShell_ViewJoinInputError(t *testing.T) {
	m := testShellModel(t)
	m.mode = ModeJoinInput
	m.joinErr = "connection failed"
	view := m.View()

	if !strings.Contains(view, "connection failed") {
		t.Error("expected error message")
	}
}

func TestShell_ViewSession(t *testing.T) {
	m := testShellModel(t)
	m.mode = ModeSession
	m.session = testModel()
	view := m.View()

	// Session with nil state shows waiting
	if !strings.Contains(view, "waiting") {
		t.Error("expected session view")
	}
}

// --- Window size ---

func TestShell_WindowSize(t *testing.T) {
	m := testShellModel(t)
	result, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	shell := result.(ShellModel)
	if shell.width != 120 || shell.height != 40 {
		t.Errorf("got %dx%d", shell.width, shell.height)
	}
}

// --- Save session to history ---

func TestShell_SaveSessionToHistory(t *testing.T) {
	m := testShellModel(t)
	m.session.state = &protocol.RetroState{
		Stage: "close",
		Meta:  protocol.RetroMeta{Name: "Sprint 42", Date: "2025-09-10"},
		Participants: []protocol.Participant{
			{ID: "p1", Name: "Alice"},
		},
		DiscussNotes: []protocol.DiscussNote{
			{ID: "n1", ParentCardID: "g1", Lane: "actions", Text: "Fix bug"},
		},
		ActionOwners: map[string]string{"n1": "p1"},
		Groups: []protocol.Group{
			{ID: "g1", Name: "Issues"},
		},
	}

	m.saveSessionToHistory()

	// Verify saved
	repo := storage.NewJSONTeamRepo(m.registry.TeamDir("t1"))
	history, _ := repo.LoadHistory()
	if len(history.Completed) != 1 {
		t.Fatalf("expected 1 completed retro, got %d", len(history.Completed))
	}
	if history.Completed[0].ID != "Sprint 42" {
		t.Errorf("expected 'Sprint 42', got %q", history.Completed[0].ID)
	}
	if len(history.Completed[0].ActionItems) != 1 {
		t.Fatalf("expected 1 action item, got %d", len(history.Completed[0].ActionItems))
	}
	item := history.Completed[0].ActionItems[0]
	if item.Text != "Fix bug" {
		t.Errorf("got %q", item.Text)
	}
	if item.OwnerName != "Alice" {
		t.Errorf("expected 'Alice', got %q", item.OwnerName)
	}
	if item.ParentText != "Issues" {
		t.Errorf("expected 'Issues', got %q", item.ParentText)
	}
}

func TestShell_SaveSessionToHistory_NilState(t *testing.T) {
	m := testShellModel(t)
	m.session.state = nil
	m.saveSessionToHistory() // should not panic
}

func TestShell_SaveSessionToHistory_NoActions(t *testing.T) {
	m := testShellModel(t)
	m.session.state = &protocol.RetroState{Stage: "close"}
	m.saveSessionToHistory()

	repo := storage.NewJSONTeamRepo(m.registry.TeamDir("t1"))
	history, _ := repo.LoadHistory()
	if len(history.Completed) != 0 {
		t.Error("should not save empty retro")
	}
}

// --- Home input doesn't trigger join ---

func TestShell_JKeyIgnoredDuringHomeInput(t *testing.T) {
	m := testShellModel(t)
	m.home.inputMode = true // simulate add-member input

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	shell := result.(ShellModel)
	if shell.mode != ModeHome {
		t.Error("j during input should not trigger join")
	}
}
