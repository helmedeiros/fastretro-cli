package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/helmedeiros/fastretro-cli/internal/domain"
	"github.com/helmedeiros/fastretro-cli/internal/protocol"
	"github.com/helmedeiros/fastretro-cli/internal/storage"
	"github.com/helmedeiros/fastretro-cli/internal/styles"
)

// HomeSection identifies which panel has focus.
type HomeSection int

const (
	SectionMembers HomeSection = iota
	SectionAgreements
	SectionActions
	SectionRetroHistory
	SectionCheckHistory
	sectionCount
)

var sectionNames = []string{"MEMBERS", "AGREEMENTS", "ACTION ITEMS", "RETRO HISTORY", "CHECK HISTORY"}

// ViewHistoryMsg signals the shell to display a completed session's close view.
type ViewHistoryMsg struct {
	State *protocol.RetroState
}

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
		m.section = (m.section + 1) % sectionCount
		m.cursor = 0
	case "shift+tab":
		m.section = (m.section + sectionCount - 1) % sectionCount
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
		if m.section == SectionRetroHistory || m.section == SectionCheckHistory {
			return m.viewHistoryAtCursor()
		}
		m.toggleAtCursor()
	case "*":
		m.toggleDefaultMember()
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

func (m HomeModel) retroHistory() []domain.CompletedRetro {
	var result []domain.CompletedRetro
	for i := len(m.history.Completed) - 1; i >= 0; i-- {
		r := m.history.Completed[i]
		if r.FullState.Meta.Type != "check" {
			result = append(result, r)
		}
	}
	return result
}

func (m HomeModel) checkHistory() []domain.CompletedRetro {
	var result []domain.CompletedRetro
	for i := len(m.history.Completed) - 1; i >= 0; i-- {
		r := m.history.Completed[i]
		if r.FullState.Meta.Type == "check" {
			result = append(result, r)
		}
	}
	return result
}

func (m HomeModel) sectionLen() int {
	switch m.section {
	case SectionMembers:
		return len(m.team.Members)
	case SectionAgreements:
		return len(m.team.Agreements)
	case SectionActions:
		return len(domain.GetAllActionItems(m.history))
	case SectionRetroHistory:
		return len(m.retroHistory())
	case SectionCheckHistory:
		return len(m.checkHistory())
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

func (m *HomeModel) toggleDefaultMember() {
	if m.section != SectionMembers || m.cursor >= len(m.team.Members) {
		return
	}
	member := m.team.Members[m.cursor]
	current := m.registry.LoadDefaultMember()
	if strings.EqualFold(current, member.Name) {
		_ = m.registry.SaveDefaultMember("")
	} else {
		_ = m.registry.SaveDefaultMember(member.Name)
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

func (m HomeModel) viewHistoryAtCursor() (tea.Model, tea.Cmd) {
	var items []domain.CompletedRetro
	if m.section == SectionCheckHistory {
		items = m.checkHistory()
	} else {
		items = m.retroHistory()
	}
	if m.cursor >= len(items) {
		return m, nil
	}
	state := items[m.cursor].FullState
	return m, func() tea.Msg { return ViewHistoryMsg{State: &state} }
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

	topRow := joinColumnsEqualHeight(panelContents, panelStyles)
	b.WriteString(topRow)
	b.WriteString("\n\n")

	// Measure top row width to match bottom row
	topWidth := 0
	for _, line := range strings.Split(topRow, "\n") {
		w := lipgloss.Width(line)
		if w > topWidth {
			topWidth = w
		}
	}
	halfWidth := topWidth / 2
	if halfWidth < 40 {
		halfWidth = 40
	}

	// History sections side by side
	retroHist := m.renderFilteredHistory("RETRO HISTORY", m.retroHistory(), m.section == SectionRetroHistory)
	checkHist := m.renderFilteredHistory("CHECK HISTORY", m.checkHistory(), m.section == SectionCheckHistory)

	histContents := []string{retroHist, checkHist}
	histStyles := make([]lipgloss.Style, 2)
	for i := range histStyles {
		style := styles.HistoryColumn.Width(halfWidth)
		section := SectionRetroHistory + HomeSection(i)
		if m.section == section {
			style = style.BorderForeground(styles.Accent)
		}
		histStyles[i] = style
	}
	b.WriteString(joinColumnsEqualHeight(histContents, histStyles))
	b.WriteString("\n")

	// Input or help
	if m.inputMode {
		label := m.inputAction
		b.WriteString(fmt.Sprintf("\n  %s: %s▌\n", label, m.inputText))
		b.WriteString(muted.Render("  [Enter] save  [Esc] cancel"))
	} else {
		b.WriteString("\n")
		b.WriteString(muted.Render("[Tab] section  [a] add  [d] delete  [e] edit  [Enter] toggle done / view  [*] set me"))
		b.WriteString("\n")
		b.WriteString(muted.Render("[j] join  [n] new retro  [c] new check  [t] teams  [q] quit"))
	}

	return b.String()
}

const maxPanelItems = 8

// scrollWindow returns start/end indices for a visible window around the cursor.
func scrollWindow(total, cursor, maxVisible int) (int, int) {
	if total <= maxVisible {
		return 0, total
	}
	half := maxVisible / 2
	start := cursor - half
	if start < 0 {
		start = 0
	}
	end := start + maxVisible
	if end > total {
		end = total
		start = end - maxVisible
	}
	return start, end
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
	} else {
		cur := 0
		if isActive {
			cur = m.cursor
		}
		start, end := scrollWindow(len(m.team.Members), cur, maxPanelItems)
		if start > 0 {
			lines = append(lines, muted.Render(fmt.Sprintf("  ↑ %d more", start)))
		}
		defaultName := m.registry.LoadDefaultMember()
		for i := start; i < end; i++ {
			member := m.team.Members[i]
			cursor := "  "
			if isActive && i == m.cursor {
				cursor = "> "
			}
			suffix := ""
			if strings.EqualFold(defaultName, member.Name) {
				suffix = " " + accent.Render("(me)")
			}
			line := cursor + member.Name + suffix
			if isActive && i == m.cursor {
				lines = append(lines, styles.Selected.Render(line))
			} else {
				lines = append(lines, line)
			}
		}
		if end < len(m.team.Members) {
			lines = append(lines, muted.Render(fmt.Sprintf("  ↓ %d more", len(m.team.Members)-end)))
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
	} else {
		cur := 0
		if isActive {
			cur = m.cursor
		}
		start, end := scrollWindow(len(m.team.Agreements), cur, maxPanelItems)
		if start > 0 {
			lines = append(lines, muted.Render(fmt.Sprintf("  ↑ %d more", start)))
		}
		for i := start; i < end; i++ {
			ag := m.team.Agreements[i]
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
		if end < len(m.team.Agreements) {
			lines = append(lines, muted.Render(fmt.Sprintf("  ↓ %d more", len(m.team.Agreements)-end)))
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
	} else {
		cur := 0
		if isActive {
			cur = m.cursor
		}
		start, end := scrollWindow(len(items), cur, maxPanelItems)
		if start > 0 {
			lines = append(lines, muted.Render(fmt.Sprintf("  ↑ %d more", start)))
		}
		for i := start; i < end; i++ {
			item := items[i]
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
		if end < len(items) {
			lines = append(lines, muted.Render(fmt.Sprintf("  ↓ %d more", len(items)-end)))
		}
	}
	return strings.Join(lines, "\n")
}

func formatHistoryDate(metaDate, completedAt string) string {
	raw := metaDate
	if raw == "" {
		raw = completedAt
	}
	if len(raw) > 10 {
		raw = raw[:10]
	}
	t, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return raw
	}
	return t.Format("Mon, Jan 2")
}

func overallCheckScore(state protocol.RetroState, tmpl protocol.CheckTemplate) float64 {
	if len(tmpl.Questions) == 0 || len(state.SurveyResponses) == 0 {
		return 0
	}
	var sum float64
	var count int
	for _, q := range tmpl.Questions {
		var ratings []int
		for _, r := range state.SurveyResponses {
			if r.QuestionID == q.ID && r.Rating > 0 {
				ratings = append(ratings, r.Rating)
			}
		}
		if len(ratings) == 0 {
			continue
		}
		for i := 1; i < len(ratings); i++ {
			for j := i; j > 0 && ratings[j] < ratings[j-1]; j-- {
				ratings[j], ratings[j-1] = ratings[j-1], ratings[j]
			}
		}
		mid := len(ratings) / 2
		var median float64
		if len(ratings)%2 == 0 {
			median = float64(ratings[mid-1]+ratings[mid]) / 2.0
		} else {
			median = float64(ratings[mid])
		}
		sum += median
		count++
	}
	if count == 0 {
		return 0
	}
	return sum / float64(count)
}

func (m HomeModel) renderFilteredHistory(title string, items []domain.CompletedRetro, isActive bool) string {
	accent := lipgloss.NewStyle().Foreground(styles.Accent).Bold(true)
	muted := lipgloss.NewStyle().Foreground(styles.Muted)
	dim := lipgloss.NewStyle().Foreground(styles.Muted)

	header := fmt.Sprintf("%s (%d)", title, len(items))
	if isActive {
		header = accent.Render(header)
	} else {
		header = muted.Render(header)
	}

	var b strings.Builder
	b.WriteString(header)
	b.WriteString("\n")

	if len(items) == 0 {
		b.WriteString(muted.Render("  No sessions yet"))
		return b.String()
	}

	cur := 0
	if isActive {
		cur = m.cursor
	}
	start, end := scrollWindow(len(items), cur, 4) // fewer items, richer display
	if start > 0 {
		b.WriteString(muted.Render(fmt.Sprintf("  ↑ %d more", start)))
		b.WriteString("\n")
	}
	for i := start; i < end; i++ {
		retro := items[i]
		isCurrent := isActive && i == m.cursor

		name := retro.FullState.Meta.Name
		if name == "" {
			name = retro.ID
		}

		// Date formatting
		date := formatHistoryDate(retro.FullState.Meta.Date, retro.CompletedAt)

		// Stats line
		participants := len(retro.FullState.Participants)
		actionCount := len(retro.ActionItems)
		isCheck := retro.FullState.Meta.Type == "check"

		var statsLine string
		if isCheck {
			tmpl := protocol.GetCheckTemplate(retro.FullState.Meta.TemplateID)
			overallScore := overallCheckScore(retro.FullState, tmpl)
			scoreStr := "—"
			if overallScore > 0 {
				scoreStr = fmt.Sprintf("%.1f", overallScore)
			}
			statsLine = fmt.Sprintf("%d participants | score %s | %d actions",
				participants, scoreStr, actionCount)
		} else {
			cards := len(retro.FullState.Cards)
			votes := len(retro.FullState.Votes)
			statsLine = fmt.Sprintf("%d participants | %d cards | %d votes | %d actions",
				participants, cards, votes, actionCount)
		}

		// Template info
		var templateLine string
		if isCheck {
			tmpl := protocol.GetCheckTemplate(retro.FullState.Meta.TemplateID)
			templateLine = tmpl.Name
		} else {
			tmpl := protocol.GetTemplate(retro.FullState.Meta.TemplateID)
			var cols []string
			for _, col := range tmpl.Columns {
				cols = append(cols, col.Title)
			}
			templateLine = strings.Join(cols, " · ")
		}

		// Render card
		cursor := "  "
		if isCurrent {
			cursor = "> "
		}

		nameRendered := name
		if isCurrent {
			nameRendered = accent.Render(name)
		}

		b.WriteString(fmt.Sprintf("%s%s\n", cursor, nameRendered))
		if date != "" {
			b.WriteString(fmt.Sprintf("  %s  %s\n", dim.Render(date), dim.Render(templateLine)))
		} else {
			b.WriteString(fmt.Sprintf("  %s\n", dim.Render(templateLine)))
		}
		b.WriteString(fmt.Sprintf("  %s\n", dim.Render(statsLine)))
		if i < end-1 {
			b.WriteString("\n")
		}
	}
	if end < len(items) {
		b.WriteString(muted.Render(fmt.Sprintf("\n  ↓ %d more", len(items)-end)))
		b.WriteString("\n")
	}
	return b.String()
}
