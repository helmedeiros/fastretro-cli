package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/helmedeiros/fastretro-cli/internal/domain"
	"github.com/helmedeiros/fastretro-cli/internal/storage"
	"github.com/helmedeiros/fastretro-cli/internal/styles"
)

// HomeSection identifies which panel has focus.
type HomeSection int

const (
	SectionMembers HomeSection = iota
	SectionAgreements
	SectionActions
	SectionHistory
)

var sectionNames = []string{"MEMBERS", "AGREEMENTS", "ACTION ITEMS", "HISTORY"}

// HomeModel is the Bubble Tea model for the dashboard home screen.
type HomeModel struct {
	registry  *storage.JSONRegistryRepo
	repo      *storage.JSONTeamRepo
	teamEntry domain.TeamEntry
	team      domain.TeamState
	history   domain.RetroHistoryState
	section   HomeSection
	cursor    int
	inputMode bool
	inputText string
	inputAction string // "add-member", "add-agreement", "add-action", "edit-agreement", "reassign", "edit-action"
	editID    string   // ID of item being edited
	width     int
	height    int
	err       error
}

// NewHomeModel creates a home screen model for the given team.
func NewHomeModel(registry *storage.JSONRegistryRepo, entry domain.TeamEntry) HomeModel {
	teamDir := registry.TeamDir(entry.ID)
	repo := storage.NewJSONTeamRepo(teamDir)
	team, _ := repo.LoadTeam()
	history, _ := repo.LoadHistory()

	return HomeModel{
		registry:  registry,
		repo:      repo,
		teamEntry: entry,
		team:      team,
		history:   history,
		width:     80,
		height:    24,
	}
}

func (m HomeModel) Init() tea.Cmd {
	return nil
}

func (m HomeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m HomeModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.inputMode {
		return m.handleInput(msg)
	}

	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "tab":
		m.section = (m.section + 1) % 4
		m.cursor = 0
	case "shift+tab":
		m.section = (m.section + 3) % 4
		m.cursor = 0
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		max := m.sectionLen() - 1
		if max < 0 {
			max = 0
		}
		if m.cursor < max {
			m.cursor++
		}
	case "a":
		m.startAdd()
	case "d":
		m.deleteAtCursor()
	case "e":
		m.startEdit()
	case "enter", " ":
		m.toggleAtCursor()
	}
	return m, nil
}

func (m HomeModel) handleInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		m.commitInput()
		m.inputMode = false
		m.inputText = ""
		m.inputAction = ""
		m.editID = ""
	case "esc":
		m.inputMode = false
		m.inputText = ""
		m.inputAction = ""
		m.editID = ""
	case "backspace":
		if len(m.inputText) > 0 {
			m.inputText = m.inputText[:len(m.inputText)-1]
		}
	default:
		if len(msg.Runes) > 0 {
			m.inputText += string(msg.Runes)
		}
	}
	return m, nil
}

func (m HomeModel) sectionLen() int {
	switch m.section {
	case SectionMembers:
		return len(m.team.Members)
	case SectionAgreements:
		return len(m.team.Agreements)
	case SectionActions:
		return len(domain.GetAllActionItems(m.history))
	case SectionHistory:
		return len(m.history.Completed)
	}
	return 0
}

func (m *HomeModel) startAdd() {
	switch m.section {
	case SectionMembers:
		m.inputMode = true
		m.inputAction = "add-member"
		m.inputText = ""
	case SectionAgreements:
		m.inputMode = true
		m.inputAction = "add-agreement"
		m.inputText = ""
	case SectionActions:
		m.inputMode = true
		m.inputAction = "add-action"
		m.inputText = ""
	}
}

func (m *HomeModel) startEdit() {
	switch m.section {
	case SectionAgreements:
		if m.cursor < len(m.team.Agreements) {
			ag := m.team.Agreements[m.cursor]
			m.inputMode = true
			m.inputAction = "edit-agreement"
			m.inputText = ag.Text
			m.editID = ag.ID
		}
	case SectionActions:
		items := domain.GetAllActionItems(m.history)
		if m.cursor < len(items) {
			item := items[m.cursor]
			m.inputMode = true
			m.inputAction = "edit-action"
			m.inputText = item.Text
			m.editID = item.NoteID
		}
	}
}

func (m *HomeModel) commitInput() {
	text := strings.TrimSpace(m.inputText)
	if text == "" {
		return
	}
	switch m.inputAction {
	case "add-member":
		id := fmt.Sprintf("m-%d", len(m.team.Members)+1)
		if updated, err := domain.AddMember(m.team, id, text); err == nil {
			m.team = updated
			m.repo.SaveTeam(m.team)
		}
	case "add-agreement":
		id := fmt.Sprintf("a-%d", len(m.team.Agreements)+1)
		if updated, err := domain.AddAgreement(m.team, id, text, ""); err == nil {
			m.team = updated
			m.repo.SaveTeam(m.team)
		}
	case "edit-agreement":
		if updated, err := domain.EditAgreement(m.team, m.editID, text); err == nil {
			m.team = updated
			m.repo.SaveTeam(m.team)
		}
	case "add-action":
		id := fmt.Sprintf("manual-%d", len(domain.GetAllActionItems(m.history))+1)
		item := domain.FlatActionItem{NoteID: id, Text: text}
		m.history = domain.AddManualActionItem(m.history, item)
		m.repo.SaveHistory(m.history)
	case "edit-action":
		m.history = domain.EditActionItemText(m.history, m.editID, text)
		m.repo.SaveHistory(m.history)
	}
}

func (m *HomeModel) deleteAtCursor() {
	switch m.section {
	case SectionMembers:
		if m.cursor < len(m.team.Members) {
			id := m.team.Members[m.cursor].ID
			if updated, err := domain.RemoveMember(m.team, id); err == nil {
				m.team = updated
				m.repo.SaveTeam(m.team)
				if m.cursor >= len(m.team.Members) && m.cursor > 0 {
					m.cursor--
				}
			}
		}
	case SectionAgreements:
		if m.cursor < len(m.team.Agreements) {
			id := m.team.Agreements[m.cursor].ID
			m.team = domain.RemoveAgreement(m.team, id)
			m.repo.SaveTeam(m.team)
			if m.cursor >= len(m.team.Agreements) && m.cursor > 0 {
				m.cursor--
			}
		}
	case SectionActions:
		items := domain.GetAllActionItems(m.history)
		if m.cursor < len(items) {
			m.history = domain.RemoveActionItem(m.history, items[m.cursor].NoteID)
			m.repo.SaveHistory(m.history)
			if m.cursor >= len(domain.GetAllActionItems(m.history)) && m.cursor > 0 {
				m.cursor--
			}
		}
	}
}

func (m *HomeModel) toggleAtCursor() {
	if m.section == SectionActions {
		items := domain.GetAllActionItems(m.history)
		if m.cursor < len(items) {
			m.history = domain.ToggleActionItemDone(m.history, items[m.cursor].NoteID)
			m.repo.SaveHistory(m.history)
		}
	}
}

func (m HomeModel) View() string {
	accent := lipgloss.NewStyle().Foreground(styles.Accent).Bold(true)
	muted := lipgloss.NewStyle().Foreground(styles.Muted)

	var b strings.Builder

	// Header
	b.WriteString(accent.Render("fastRetro CLI"))
	b.WriteString("  " + muted.Render("team: "+m.teamEntry.Name))
	b.WriteString("\n")
	b.WriteString(muted.Render(strings.Repeat("─", 60)))
	b.WriteString("\n\n")

	// Panels: Members | Agreements | Action Items
	membersPanel := m.renderMembers()
	agreementsPanel := m.renderAgreements()
	actionsPanel := m.renderActions()

	panelContents := []string{membersPanel, agreementsPanel, actionsPanel}
	panelStyles := make([]lipgloss.Style, 3)
	for i := range panelStyles {
		style := styles.Column
		if int(m.section) == i {
			style = style.BorderForeground(styles.Accent)
		}
		panelStyles[i] = style
	}

	b.WriteString(joinColumnsEqualHeight(panelContents, panelStyles))
	b.WriteString("\n\n")

	// History section
	b.WriteString(m.renderHistory())
	b.WriteString("\n")

	// Input or help
	if m.inputMode {
		label := m.inputAction
		b.WriteString(fmt.Sprintf("\n  %s: %s▌\n", label, m.inputText))
		b.WriteString(muted.Render("  [Enter] save  [Esc] cancel"))
	} else {
		b.WriteString("\n")
		b.WriteString(muted.Render("[Tab] section  [a] add  [d] delete  [e] edit  [Enter] toggle done"))
		b.WriteString("\n")
		b.WriteString(muted.Render("[j] join retro  [q] quit"))
	}

	return b.String()
}

func (m HomeModel) renderMembers() string {
	accent := lipgloss.NewStyle().Foreground(styles.Accent).Bold(true)
	muted := lipgloss.NewStyle().Foreground(styles.Muted)
	isActive := m.section == SectionMembers

	header := fmt.Sprintf("MEMBERS (%d)", len(m.team.Members))
	if isActive {
		header = accent.Render(header)
	} else {
		header = muted.Render(header)
	}

	var lines []string
	lines = append(lines, header, "")

	if len(m.team.Members) == 0 {
		lines = append(lines, muted.Render("(empty)"))
	}
	for i, member := range m.team.Members {
		cursor := "  "
		if isActive && i == m.cursor {
			cursor = "> "
		}
		line := cursor + member.Name
		if isActive && i == m.cursor {
			lines = append(lines, styles.Selected.Render(line))
		} else {
			lines = append(lines, line)
		}
	}
	return strings.Join(lines, "\n")
}

func (m HomeModel) renderAgreements() string {
	accent := lipgloss.NewStyle().Foreground(styles.Accent).Bold(true)
	muted := lipgloss.NewStyle().Foreground(styles.Muted)
	isActive := m.section == SectionAgreements

	header := fmt.Sprintf("AGREEMENTS (%d)", len(m.team.Agreements))
	if isActive {
		header = accent.Render(header)
	} else {
		header = muted.Render(header)
	}

	var lines []string
	lines = append(lines, header, "")

	if len(m.team.Agreements) == 0 {
		lines = append(lines, muted.Render("(empty)"))
	}
	for i, ag := range m.team.Agreements {
		cursor := "  "
		if isActive && i == m.cursor {
			cursor = "> "
		}
		line := cursor + ag.Text
		if isActive && i == m.cursor {
			lines = append(lines, styles.Selected.Render(line))
		} else {
			lines = append(lines, line)
		}
	}
	return strings.Join(lines, "\n")
}

func (m HomeModel) renderActions() string {
	accent := lipgloss.NewStyle().Foreground(styles.Accent).Bold(true)
	muted := lipgloss.NewStyle().Foreground(styles.Muted)
	isActive := m.section == SectionActions

	items := domain.GetAllActionItems(m.history)
	header := fmt.Sprintf("ACTION ITEMS (%d)", len(items))
	if isActive {
		header = accent.Render(header)
	} else {
		header = muted.Render(header)
	}

	var lines []string
	lines = append(lines, header, "")

	if len(items) == 0 {
		lines = append(lines, muted.Render("(empty)"))
	}
	for i, item := range items {
		cursor := "  "
		if isActive && i == m.cursor {
			cursor = "> "
		}
		check := "[ ]"
		if item.Done {
			check = "[x]"
		}
		owner := ""
		if item.OwnerName != "" {
			owner = muted.Render(" → " + item.OwnerName)
		}
		line := fmt.Sprintf("%s%s %s%s", cursor, check, item.Text, owner)
		if isActive && i == m.cursor {
			lines = append(lines, styles.Selected.Render(line))
		} else {
			lines = append(lines, line)
		}
	}
	return strings.Join(lines, "\n")
}

func (m HomeModel) renderHistory() string {
	accent := lipgloss.NewStyle().Foreground(styles.Accent).Bold(true)
	muted := lipgloss.NewStyle().Foreground(styles.Muted)
	isActive := m.section == SectionHistory

	header := fmt.Sprintf("RETRO HISTORY (%d)", len(m.history.Completed))
	if isActive {
		header = accent.Render(header)
	} else {
		header = muted.Render(header)
	}

	var b strings.Builder
	b.WriteString(header)
	b.WriteString("\n")
	b.WriteString(muted.Render(strings.Repeat("─", 40)))
	b.WriteString("\n")

	if len(m.history.Completed) == 0 {
		b.WriteString(muted.Render("  No retros completed yet"))
	}
	for i, retro := range m.history.Completed {
		cursor := "  "
		if isActive && i == m.cursor {
			cursor = "> "
		}
		name := retro.FullState.Meta.Name
		if name == "" {
			name = retro.ID
		}
		line := fmt.Sprintf("%s%s — %s — %d action items",
			cursor, name, retro.CompletedAt, len(retro.ActionItems))
		if isActive && i == m.cursor {
			b.WriteString(styles.Selected.Render(line))
		} else {
			b.WriteString(line)
		}
		b.WriteString("\n")
	}
	return b.String()
}
