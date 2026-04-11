package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/helmedeiros/fastretro-cli/internal/domain"
	"github.com/helmedeiros/fastretro-cli/internal/storage"
)

func testHomeModel(t *testing.T) HomeModel {
	t.Helper()
	dir := t.TempDir()
	reg := storage.NewJSONRegistryRepo(dir)
	entry := domain.TeamEntry{ID: "t1", Name: "Test Squad"}

	repo := storage.NewJSONTeamRepo(reg.TeamDir("t1"))
	team := domain.NewTeam()
	team, _ = domain.AddMember(team, "m1", "Alice")
	team, _ = domain.AddMember(team, "m2", "Bob")
	team, _ = domain.AddAgreement(team, "a1", "Demo Fridays", "2025-09-01")
	repo.SaveTeam(team)

	h := domain.NewHistory()
	h = domain.AddCompletedRetro(h, domain.CompletedRetro{
		ID: "r1", CompletedAt: "2025-09-07",
		ActionItems: []domain.FlatActionItem{
			{NoteID: "n1", Text: "Fix bug", OwnerName: "Bob", Done: false},
			{NoteID: "n2", Text: "Update CI", Done: true},
		},
	})
	repo.SaveHistory(h)

	return NewHomeModel(reg, entry)
}

// --- View ---

func TestHomeView_ShowsTeamName(t *testing.T) {
	m := testHomeModel(t)
	view := m.View()
	if !strings.Contains(view, "Test Squad") {
		t.Error("expected team name")
	}
}

func TestHomeView_ShowsMembers(t *testing.T) {
	m := testHomeModel(t)
	view := m.View()
	if !strings.Contains(view, "Alice") || !strings.Contains(view, "Bob") {
		t.Error("expected members")
	}
	if !strings.Contains(view, "MEMBERS (2)") {
		t.Error("expected member count")
	}
}

func TestHomeView_ShowsAgreements(t *testing.T) {
	m := testHomeModel(t)
	view := m.View()
	if !strings.Contains(view, "Demo Fridays") {
		t.Error("expected agreement")
	}
}

func TestHomeView_ShowsActionItems(t *testing.T) {
	m := testHomeModel(t)
	view := m.View()
	if !strings.Contains(view, "Fix bug") {
		t.Error("expected action item")
	}
	if !strings.Contains(view, "[x]") {
		t.Error("expected done checkbox")
	}
	if !strings.Contains(view, "[ ]") {
		t.Error("expected open checkbox")
	}
}

func TestHomeView_ShowsHistory(t *testing.T) {
	m := testHomeModel(t)
	view := m.View()
	if !strings.Contains(view, "RETRO HISTORY") {
		t.Error("expected history section")
	}
	if !strings.Contains(view, "2025-09-07") {
		t.Error("expected retro date")
	}
}

func TestHomeView_ShowsHelp(t *testing.T) {
	m := testHomeModel(t)
	view := m.View()
	if !strings.Contains(view, "Tab") {
		t.Error("expected help keys")
	}
}

func TestHomeView_InputMode(t *testing.T) {
	m := testHomeModel(t)
	m.inputMode = true
	m.inputAction = "add-member"
	m.inputText = "Carol"
	view := m.View()
	if !strings.Contains(view, "Carol") {
		t.Error("expected input text")
	}
}

// --- Navigation ---

func TestHome_TabCyclesSections(t *testing.T) {
	m := testHomeModel(t)
	if m.section != SectionMembers {
		t.Error("should start at members")
	}

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = result.(HomeModel)
	if m.section != SectionAgreements {
		t.Errorf("expected agreements, got %d", m.section)
	}

	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = result.(HomeModel)
	if m.section != SectionActions {
		t.Errorf("expected actions, got %d", m.section)
	}

	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = result.(HomeModel)
	if m.section != SectionRetroHistory {
		t.Errorf("expected retro history, got %d", m.section)
	}

	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = result.(HomeModel)
	if m.section != SectionCheckHistory {
		t.Errorf("expected check history, got %d", m.section)
	}

	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = result.(HomeModel)
	if m.section != SectionMembers {
		t.Errorf("expected wrap to members, got %d", m.section)
	}
}

func TestHome_ShiftTabReverse(t *testing.T) {
	m := testHomeModel(t)
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	m = result.(HomeModel)
	if m.section != SectionCheckHistory {
		t.Errorf("expected check history, got %d", m.section)
	}
}

func TestHome_CursorUpDown(t *testing.T) {
	m := testHomeModel(t)
	// Members section, 2 members
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	m = result.(HomeModel)
	if m.cursor != 1 {
		t.Errorf("expected 1, got %d", m.cursor)
	}
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	m = result.(HomeModel)
	if m.cursor != 0 {
		t.Errorf("expected 0, got %d", m.cursor)
	}
}

func TestHome_CursorResetsOnTab(t *testing.T) {
	m := testHomeModel(t)
	m.cursor = 1
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = result.(HomeModel)
	if m.cursor != 0 {
		t.Error("cursor should reset on section change")
	}
}

func TestHome_WindowSize(t *testing.T) {
	m := testHomeModel(t)
	result, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = result.(HomeModel)
	if m.width != 120 || m.height != 40 {
		t.Errorf("got %dx%d", m.width, m.height)
	}
}

// --- Add member ---

func TestHome_AddMember(t *testing.T) {
	m := testHomeModel(t)
	m.section = SectionMembers

	// Press 'a'
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	m = result.(HomeModel)
	if !m.inputMode || m.inputAction != "add-member" {
		t.Error("expected add-member input mode")
	}

	// Type "Carol"
	for _, ch := range "Carol" {
		result, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
		m = result.(HomeModel)
	}

	// Press Enter
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = result.(HomeModel)

	if len(m.team.Members) != 3 {
		t.Errorf("expected 3 members, got %d", len(m.team.Members))
	}
	if m.inputMode {
		t.Error("should exit input mode")
	}
}

// --- Delete member ---

func TestHome_DeleteMember(t *testing.T) {
	m := testHomeModel(t)
	m.section = SectionMembers
	m.cursor = 0

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})
	m = result.(HomeModel)

	if len(m.team.Members) != 1 {
		t.Errorf("expected 1 member, got %d", len(m.team.Members))
	}
}

// --- Add agreement ---

func TestHome_AddAgreement(t *testing.T) {
	m := testHomeModel(t)
	m.section = SectionAgreements

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	m = result.(HomeModel)
	if m.inputAction != "add-agreement" {
		t.Error("expected add-agreement")
	}
}

// --- Edit agreement ---

func TestHome_EditAgreement(t *testing.T) {
	m := testHomeModel(t)
	m.section = SectionAgreements
	m.cursor = 0

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	m = result.(HomeModel)
	if m.inputAction != "edit-agreement" {
		t.Error("expected edit-agreement")
	}
	if m.inputText != "Demo Fridays" {
		t.Errorf("expected prefilled text, got %q", m.inputText)
	}
}

// --- Delete agreement ---

func TestHome_DeleteAgreement(t *testing.T) {
	m := testHomeModel(t)
	m.section = SectionAgreements
	m.cursor = 0

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})
	m = result.(HomeModel)

	if len(m.team.Agreements) != 0 {
		t.Errorf("expected 0, got %d", len(m.team.Agreements))
	}
}

// --- Toggle action item ---

func TestHome_ToggleActionItem(t *testing.T) {
	m := testHomeModel(t)
	m.section = SectionActions
	m.cursor = 0 // "Fix bug" is not done

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = result.(HomeModel)

	items := domain.GetAllActionItems(m.history)
	if !items[0].Done {
		t.Error("expected toggled to done")
	}
}

// --- Delete action item ---

func TestHome_DeleteActionItem(t *testing.T) {
	m := testHomeModel(t)
	m.section = SectionActions
	m.cursor = 0

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})
	m = result.(HomeModel)

	items := domain.GetAllActionItems(m.history)
	if len(items) != 1 {
		t.Errorf("expected 1, got %d", len(items))
	}
}

// --- Add manual action item ---

func TestHome_AddAction(t *testing.T) {
	m := testHomeModel(t)
	m.section = SectionActions

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	m = result.(HomeModel)
	if m.inputAction != "add-action" {
		t.Error("expected add-action")
	}
}

// --- Edit action item ---

func TestHome_EditAction(t *testing.T) {
	m := testHomeModel(t)
	m.section = SectionActions
	m.cursor = 0

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	m = result.(HomeModel)
	if m.inputAction != "edit-action" {
		t.Error("expected edit-action")
	}
	if m.inputText != "Fix bug" {
		t.Errorf("expected prefilled, got %q", m.inputText)
	}
}

// --- Input cancel ---

func TestHome_InputEscape(t *testing.T) {
	m := testHomeModel(t)
	m.inputMode = true
	m.inputText = "partial"

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = result.(HomeModel)
	if m.inputMode {
		t.Error("should exit input")
	}
}

func TestHome_InputBackspace(t *testing.T) {
	m := testHomeModel(t)
	m.inputMode = true
	m.inputText = "ab"

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	m = result.(HomeModel)
	if m.inputText != "a" {
		t.Errorf("got %q", m.inputText)
	}
}

// --- Quit ---

func TestHome_Quit(t *testing.T) {
	m := testHomeModel(t)
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	if cmd == nil {
		t.Error("expected quit command")
	}
}

// --- Empty state ---

func TestHome_EmptyTeam(t *testing.T) {
	dir := t.TempDir()
	reg := storage.NewJSONRegistryRepo(dir)
	entry := domain.TeamEntry{ID: "t1", Name: "Empty"}
	m := NewHomeModel(reg, entry)
	view := m.View()
	if !strings.Contains(view, "empty") {
		t.Error("expected empty indicators")
	}
}

// --- History add disabled ---

func TestHome_AddOnHistory_NoOp(t *testing.T) {
	m := testHomeModel(t)
	m.section = SectionRetroHistory
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	m = result.(HomeModel)
	if m.inputMode {
		t.Error("should not enter input on history section")
	}
}

// --- Commit empty input ---

func TestHome_CommitEmptyInput(t *testing.T) {
	m := testHomeModel(t)
	m.inputMode = true
	m.inputAction = "add-member"
	m.inputText = "  "

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = result.(HomeModel)
	if len(m.team.Members) != 2 {
		t.Error("empty input should not add member")
	}
}

func TestScrollWindow_AllVisible(t *testing.T) {
	start, end := scrollWindow(5, 0, 8)
	if start != 0 || end != 5 {
		t.Errorf("expected 0-5, got %d-%d", start, end)
	}
}

func TestScrollWindow_ScrollsDown(t *testing.T) {
	start, end := scrollWindow(20, 15, 8)
	if end-start != 8 {
		t.Errorf("window should be 8, got %d", end-start)
	}
	if start > 15 || end <= 15 {
		t.Errorf("cursor 15 should be visible in %d-%d", start, end)
	}
}

func TestScrollWindow_ClampsEnd(t *testing.T) {
	start, end := scrollWindow(10, 9, 8)
	if end != 10 {
		t.Errorf("expected end=10, got %d", end)
	}
	if start != 2 {
		t.Errorf("expected start=2, got %d", start)
	}
}

func TestScrollWindow_ClampsStart(t *testing.T) {
	start, end := scrollWindow(20, 0, 8)
	if start != 0 {
		t.Errorf("expected start=0, got %d", start)
	}
	if end != 8 {
		t.Errorf("expected end=8, got %d", end)
	}
}

func TestHome_ToggleDefaultMember(t *testing.T) {
	m := testHomeModel(t)
	m.section = SectionMembers
	m.cursor = 0 // Alice

	// Set as default
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("*")})
	m = result.(HomeModel)
	if got := m.registry.LoadDefaultMember(); got != "Alice" {
		t.Errorf("expected Alice, got %q", got)
	}

	// Toggle off
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("*")})
	m = result.(HomeModel)
	if got := m.registry.LoadDefaultMember(); got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

func TestHome_ToggleDefaultMember_WrongSection(t *testing.T) {
	m := testHomeModel(t)
	m.section = SectionAgreements

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("*")})
	m = result.(HomeModel)
	if got := m.registry.LoadDefaultMember(); got != "" {
		t.Errorf("should not set default from agreements section, got %q", got)
	}
}
