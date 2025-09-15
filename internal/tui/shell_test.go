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

// --- New retro flow ---

func TestShell_NKeyOpensNewRetro(t *testing.T) {
	m := testShellModel(t)
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	shell := result.(ShellModel)
	if shell.mode != ModeNewRetro {
		t.Errorf("expected ModeNewRetro, got %d", shell.mode)
	}
}

func TestShell_NKeyIgnoredDuringInput(t *testing.T) {
	m := testShellModel(t)
	m.home.inputMode = true
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	shell := result.(ShellModel)
	if shell.mode != ModeHome {
		t.Error("n during input should not trigger new retro")
	}
}

func TestShell_NewRetro_NavigateTemplates(t *testing.T) {
	m := testShellModel(t)
	m.mode = ModeNewRetro

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	shell := result.(ShellModel)
	if shell.templateCursor != 1 {
		t.Errorf("expected 1, got %d", shell.templateCursor)
	}

	result, _ = shell.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	shell = result.(ShellModel)
	if shell.templateCursor != 0 {
		t.Errorf("expected 0, got %d", shell.templateCursor)
	}
}

func TestShell_NewRetro_EscReturnsHome(t *testing.T) {
	m := testShellModel(t)
	m.mode = ModeNewRetro

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	shell := result.(ShellModel)
	if shell.mode != ModeHome {
		t.Error("expected ModeHome")
	}
}

func TestShell_NewRetro_EnterShowsNameInput(t *testing.T) {
	m := testShellModel(t)
	m.mode = ModeNewRetro

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	shell := result.(ShellModel)
	if !shell.retroNameInput {
		t.Error("expected name input mode")
	}
}

func TestShell_NewRetro_TypeName(t *testing.T) {
	m := testShellModel(t)
	m.mode = ModeNewRetro
	m.retroNameInput = true

	for _, ch := range "Sprint 1" {
		result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
		m = result.(ShellModel)
	}
	if m.retroName != "Sprint 1" {
		t.Errorf("got %q", m.retroName)
	}
}

func TestShell_NewRetro_StartRetro(t *testing.T) {
	m := testShellModel(t)
	m.mode = ModeNewRetro
	m.retroNameInput = true
	m.retroName = "Sprint 42"

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	shell := result.(ShellModel)

	if shell.mode != ModeSession {
		t.Errorf("expected ModeSession, got %d", shell.mode)
	}
	if shell.session.state == nil {
		t.Fatal("session state should be set")
	}
	if shell.session.state.Meta.Name != "Sprint 42" {
		t.Errorf("got %q", shell.session.state.Meta.Name)
	}
	if shell.session.state.Meta.TemplateID != "start-stop" {
		t.Errorf("expected start-stop, got %q", shell.session.state.Meta.TemplateID)
	}
	if len(shell.session.state.Participants) != 1 {
		t.Errorf("expected 1 participant (Alice), got %d", len(shell.session.state.Participants))
	}
}

func TestShell_NewRetro_DefaultName(t *testing.T) {
	m := testShellModel(t)
	m.mode = ModeNewRetro
	m.retroNameInput = true
	m.retroName = "" // empty = use template name

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	shell := result.(ShellModel)
	if shell.session.state.Meta.Name != "Start / Stop" {
		t.Errorf("expected template name, got %q", shell.session.state.Meta.Name)
	}
}

func TestShell_NewRetro_NameBackspace(t *testing.T) {
	m := testShellModel(t)
	m.mode = ModeNewRetro
	m.retroNameInput = true
	m.retroName = "abc"

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	shell := result.(ShellModel)
	if shell.retroName != "ab" {
		t.Errorf("got %q", shell.retroName)
	}
}

func TestShell_NewRetro_NameEscBack(t *testing.T) {
	m := testShellModel(t)
	m.mode = ModeNewRetro
	m.retroNameInput = true

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	shell := result.(ShellModel)
	if shell.retroNameInput {
		t.Error("should exit name input")
	}
	if shell.mode != ModeNewRetro {
		t.Error("should stay in new retro (template picker)")
	}
}

func TestShell_ViewNewRetro_TemplatePicker(t *testing.T) {
	m := testShellModel(t)
	m.mode = ModeNewRetro
	view := m.View()

	if !strings.Contains(view, "New Retrospective") {
		t.Error("expected header")
	}
	if !strings.Contains(view, "Start / Stop") {
		t.Error("expected template name")
	}
	if !strings.Contains(view, "Starfish") {
		t.Error("expected all templates listed")
	}
}

func TestShell_ViewNewRetro_NameInput(t *testing.T) {
	m := testShellModel(t)
	m.mode = ModeNewRetro
	m.retroNameInput = true
	m.retroName = "My Retro"
	view := m.View()

	if !strings.Contains(view, "My Retro") {
		t.Error("expected name input")
	}
}

// --- Team selector ---

func TestShell_TKeyOpensTeamSelect(t *testing.T) {
	m := testShellModel(t)
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("t")})
	shell := result.(ShellModel)
	if shell.mode != ModeTeamSelect {
		t.Errorf("expected ModeTeamSelect, got %d", shell.mode)
	}
}

func TestShell_TeamSelect_ShowsTeams(t *testing.T) {
	m := testShellModel(t)
	// Add a team to registry first
	entries, _ := m.registry.List()
	entries, _ = domain.AddTeamEntry(entries, "t1", "Test Squad", "2025-01-01")
	entries, _ = domain.AddTeamEntry(entries, "t2", "Other Team", "2025-01-02")
	m.registry.Save(entries)

	m.mode = ModeTeamSelect
	m.teamEntries = entries
	view := m.View()

	if !strings.Contains(view, "Test Squad") {
		t.Error("expected team name")
	}
	if !strings.Contains(view, "Other Team") {
		t.Error("expected second team")
	}
}

func TestShell_TeamSelect_Navigate(t *testing.T) {
	m := testShellModel(t)
	m.mode = ModeTeamSelect
	m.teamEntries = []domain.TeamEntry{{ID: "t1", Name: "A"}, {ID: "t2", Name: "B"}}

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	shell := result.(ShellModel)
	if shell.teamCursor != 1 {
		t.Errorf("expected 1, got %d", shell.teamCursor)
	}

	result, _ = shell.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	shell = result.(ShellModel)
	if shell.teamCursor != 0 {
		t.Errorf("expected 0, got %d", shell.teamCursor)
	}
}

func TestShell_TeamSelect_SelectTeam(t *testing.T) {
	m := testShellModel(t)
	m.mode = ModeTeamSelect
	m.teamEntries = []domain.TeamEntry{{ID: "t1", Name: "A"}, {ID: "t2", Name: "B"}}
	m.teamCursor = 1

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	shell := result.(ShellModel)
	if shell.mode != ModeHome {
		t.Errorf("expected ModeHome, got %d", shell.mode)
	}
	if shell.teamEntry.ID != "t2" {
		t.Errorf("expected t2, got %q", shell.teamEntry.ID)
	}
}

func TestShell_TeamSelect_Create(t *testing.T) {
	m := testShellModel(t)
	m.mode = ModeTeamSelect
	m.teamEntries = []domain.TeamEntry{}

	// Press c
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")})
	shell := result.(ShellModel)
	if !shell.teamInputMode || shell.teamAction != "create" {
		t.Error("expected create input mode")
	}

	// Type name
	for _, ch := range "New Team" {
		result, _ = shell.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
		shell = result.(ShellModel)
	}

	// Enter
	result, _ = shell.Update(tea.KeyMsg{Type: tea.KeyEnter})
	shell = result.(ShellModel)
	if len(shell.teamEntries) != 1 {
		t.Fatalf("expected 1 team, got %d", len(shell.teamEntries))
	}
	if shell.teamEntries[0].Name != "New Team" {
		t.Errorf("got %q", shell.teamEntries[0].Name)
	}
}

func TestShell_TeamSelect_Delete(t *testing.T) {
	m := testShellModel(t)
	m.mode = ModeTeamSelect
	m.teamEntries = []domain.TeamEntry{{ID: "t1", Name: "A"}, {ID: "t2", Name: "B"}}
	m.teamCursor = 0

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})
	shell := result.(ShellModel)
	if len(shell.teamEntries) != 1 {
		t.Fatalf("expected 1 team, got %d", len(shell.teamEntries))
	}
	if shell.teamEntries[0].ID != "t2" {
		t.Errorf("expected t2 remaining")
	}
}

func TestShell_TeamSelect_Rename(t *testing.T) {
	m := testShellModel(t)
	m.mode = ModeTeamSelect
	m.teamEntries = []domain.TeamEntry{{ID: "t1", Name: "Old Name"}}
	m.teamCursor = 0

	// Press r
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	shell := result.(ShellModel)
	if !shell.teamInputMode || shell.teamAction != "rename" {
		t.Error("expected rename input mode")
	}
	if shell.teamInput != "Old Name" {
		t.Errorf("expected prefilled, got %q", shell.teamInput)
	}
}

func TestShell_TeamSelect_EscReturnsHome(t *testing.T) {
	m := testShellModel(t)
	m.mode = ModeTeamSelect

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	shell := result.(ShellModel)
	if shell.mode != ModeHome {
		t.Errorf("expected ModeHome, got %d", shell.mode)
	}
}

func TestShell_TeamSelect_InputEsc(t *testing.T) {
	m := testShellModel(t)
	m.mode = ModeTeamSelect
	m.teamInputMode = true

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	shell := result.(ShellModel)
	if shell.teamInputMode {
		t.Error("should exit input")
	}
}

func TestShell_TeamSelect_InputBackspace(t *testing.T) {
	m := testShellModel(t)
	m.mode = ModeTeamSelect
	m.teamInputMode = true
	m.teamInput = "abc"

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	shell := result.(ShellModel)
	if shell.teamInput != "ab" {
		t.Errorf("got %q", shell.teamInput)
	}
}

func TestShell_StartInTeamSelect(t *testing.T) {
	dir := t.TempDir()
	reg := storage.NewJSONRegistryRepo(dir)
	entry := domain.TeamEntry{} // no team
	m := NewShellModel(reg, entry, "http://localhost:5173")
	m.StartInTeamSelect()

	if m.mode != ModeTeamSelect {
		t.Errorf("expected ModeTeamSelect, got %d", m.mode)
	}
}

// --- Capture participants ---

func TestShell_SaveSessionCaptures_Participants(t *testing.T) {
	m := testShellModel(t)
	m.session.state = &protocol.RetroState{
		Stage: "close",
		Meta:  protocol.RetroMeta{Name: "Sprint 1"},
		Participants: []protocol.Participant{
			{ID: "p1", Name: "Alice"}, // already in team
			{ID: "p2", Name: "NewPerson"},
		},
		DiscussNotes: []protocol.DiscussNote{
			{ID: "n1", ParentCardID: "c1", Lane: "actions", Text: "Do stuff"},
		},
	}

	m.saveSessionToHistory()

	// Check team members were updated
	repo := storage.NewJSONTeamRepo(m.registry.TeamDir("t1"))
	team, _ := repo.LoadTeam()

	found := false
	for _, member := range team.Members {
		if member.Name == "NewPerson" {
			found = true
		}
	}
	if !found {
		t.Error("expected NewPerson to be added to team members")
	}
}

func TestShell_TeamSelect_EmptyView(t *testing.T) {
	m := testShellModel(t)
	m.mode = ModeTeamSelect
	m.teamEntries = []domain.TeamEntry{}
	view := m.View()

	if !strings.Contains(view, "No teams") {
		t.Error("expected empty state message")
	}
}
