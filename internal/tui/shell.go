package tui

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/helmedeiros/fastretro-cli/internal/client"
	"github.com/helmedeiros/fastretro-cli/internal/domain"
	"github.com/helmedeiros/fastretro-cli/internal/protocol"
	"github.com/helmedeiros/fastretro-cli/internal/storage"
	"github.com/helmedeiros/fastretro-cli/internal/styles"
)

// ShellMode indicates which screen is active.
type ShellMode int

const (
	ModeHome ShellMode = iota
	ModeJoinInput
	ModeNewRetro
	ModeNewCheck
	ModeTeamSelect
	ModeSession
)

// JoinStartMsg signals the shell to show the room code input.
type JoinStartMsg struct{}

// JoinConnectMsg signals the shell to connect and start a session.
type JoinConnectMsg struct {
	RoomCode  string
	ServerURL string
}

// SessionDoneMsg signals the session ended, with the final state for saving.
type SessionDoneMsg struct {
	FinalState interface{} // *protocol.RetroState or nil
}

// ShellModel is the top-level model that switches between home and session.
type ShellModel struct {
	mode      ShellMode
	home      HomeModel
	session   Model
	registry  *storage.JSONRegistryRepo
	teamEntry domain.TeamEntry
	serverURL string
	joinInput      string
	joinErr        string
	teamEntries    []domain.TeamEntry
	teamCursor     int
	teamInput      string
	teamInputMode  bool
	teamAction     string // "create", "rename"
	templateCursor      int
	checkTemplateCursor int
	retroName           string
	retroNameInput      bool
	width          int
	height         int
}

// NewShellModel creates the shell with a home screen.
func NewShellModel(registry *storage.JSONRegistryRepo, entry domain.TeamEntry, serverURL string) ShellModel {
	return ShellModel{
		mode:      ModeHome,
		home:      NewHomeModel(registry, entry),
		registry:  registry,
		teamEntry: entry,
		serverURL: serverURL,
		width:     80,
		height:    24,
	}
}

// StartInTeamSelect sets the initial mode to team selector (for first launch with no teams).
func (m *ShellModel) StartInTeamSelect() {
	entries, _ := m.registry.List()
	m.teamEntries = entries
	m.mode = ModeTeamSelect
}

func (m ShellModel) Init() tea.Cmd {
	return m.home.Init()
}

func (m ShellModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	switch m.mode {
	case ModeHome:
		return m.updateHome(msg)
	case ModeJoinInput:
		return m.updateJoinInput(msg)
	case ModeNewRetro:
		return m.updateNewRetro(msg)
	case ModeNewCheck:
		return m.updateNewCheck(msg)
	case ModeTeamSelect:
		return m.updateTeamSelect(msg)
	case ModeSession:
		return m.updateSession(msg)
	}
	return m, nil
}

func (m ShellModel) updateHome(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if !m.home.inputMode {
			switch msg.String() {
			case "j":
				m.mode = ModeJoinInput
				m.joinInput = ""
				m.joinErr = ""
				return m, nil
			case "n":
				m.mode = ModeNewRetro
				m.templateCursor = 0
				m.retroName = ""
				m.retroNameInput = false
				return m, nil
			case "c":
				m.mode = ModeNewCheck
				m.checkTemplateCursor = 0
				m.retroName = ""
				m.retroNameInput = false
				return m, nil
			case "t":
				return m.openTeamSelect()
			}
		}
	}

	updated, cmd := m.home.Update(msg)
	m.home = updated.(HomeModel)
	return m, cmd
}

func (m ShellModel) updateJoinInput(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if m.joinInput != "" {
				return m.connectToRoom()
			}
		case "esc":
			m.mode = ModeHome
			return m, nil
		case "backspace":
			if len(m.joinInput) > 0 {
				m.joinInput = m.joinInput[:len(m.joinInput)-1]
			}
		default:
			if len(msg.Runes) > 0 {
				m.joinInput += string(msg.Runes)
			}
		}
	}
	return m, nil
}

func (m ShellModel) connectToRoom() (tea.Model, tea.Cmd) {
	c, err := client.Connect(m.joinInput, m.serverURL)
	if err != nil {
		m.joinErr = err.Error()
		return m, nil
	}

	m.session = NewModel(c)
	m.session.serverURL = m.serverURL
	m.session.width = m.width
	m.session.height = m.height

	// Restore identity if reconnecting to the same room
	if saved := m.registry.LoadIdentity(c.RoomCode); saved != "" {
		m.session.participantID = saved
		c.ClaimIdentity(saved)
	}
	// Store default member name for auto-matching when state arrives
	m.session.defaultMemberName = m.registry.LoadDefaultMember()

	m.mode = ModeSession
	return m, m.session.Init()
}

func (m *ShellModel) rememberIdentity() {
	if m.session.client != nil && m.session.participantID != "" {
		_ = m.registry.SaveIdentity(m.session.client.RoomCode, m.session.participantID)
	}
}

func (m ShellModel) updateSession(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Intercept ctrl+c in session to return to home instead of quitting
		if msg.String() == "ctrl+c" {
			m.rememberIdentity()
			m.saveSessionToHistory()
			m.mode = ModeHome
			m.home = NewHomeModel(m.registry, m.teamEntry) // refresh data
			return m, nil
		}
		// Intercept q when not in input mode
		if msg.String() == "q" && !m.session.inputMode {
			m.rememberIdentity()
			m.saveSessionToHistory()
			m.mode = ModeHome
			m.home = NewHomeModel(m.registry, m.teamEntry)
			return m, nil
		}
	}

	prevID := m.session.participantID
	updated, cmd := m.session.Update(msg)
	m.session = updated.(Model)

	// Persist identity as soon as the user picks one
	if m.session.participantID != "" && m.session.participantID != prevID {
		m.rememberIdentity()
	}

	return m, cmd
}

func (m ShellModel) updateNewRetro(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.retroNameInput {
			switch msg.String() {
			case "enter":
				name := m.retroName
				if name == "" {
					name = protocol.Templates[m.templateCursor].Name
				}
				m.startLocalRetro(name)
				if m.session.client != nil {
					return m, m.session.Init()
				}
				return m, nil
			case "esc":
				m.retroNameInput = false
				m.retroName = ""
			case "backspace":
				if len(m.retroName) > 0 {
					m.retroName = m.retroName[:len(m.retroName)-1]
				}
			default:
				if len(msg.Runes) > 0 {
					m.retroName += string(msg.Runes)
				}
			}
			return m, nil
		}

		switch msg.String() {
		case "esc":
			m.mode = ModeHome
			return m, nil
		case "up", "k":
			if m.templateCursor > 0 {
				m.templateCursor--
			}
		case "down", "j":
			if m.templateCursor < len(protocol.Templates)-1 {
				m.templateCursor++
			}
		case "enter":
			m.retroNameInput = true
			m.retroName = ""
		}
	}
	return m, nil
}

func (m ShellModel) updateNewCheck(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.retroNameInput {
			switch msg.String() {
			case "enter":
				name := m.retroName
				if name == "" {
					name = protocol.CheckTemplates[m.checkTemplateCursor].Name
				}
				m.startLocalCheck(name)
				if m.session.client != nil {
					return m, m.session.Init()
				}
				return m, nil
			case "esc":
				m.retroNameInput = false
				m.retroName = ""
			case "backspace":
				if len(m.retroName) > 0 {
					m.retroName = m.retroName[:len(m.retroName)-1]
				}
			default:
				if len(msg.Runes) > 0 {
					m.retroName += string(msg.Runes)
				}
			}
			return m, nil
		}

		switch msg.String() {
		case "esc":
			m.mode = ModeHome
			return m, nil
		case "up", "k":
			if m.checkTemplateCursor > 0 {
				m.checkTemplateCursor--
			}
		case "down", "j":
			if m.checkTemplateCursor < len(protocol.CheckTemplates)-1 {
				m.checkTemplateCursor++
			}
		case "enter":
			m.retroNameInput = true
			m.retroName = ""
		}
	}
	return m, nil
}

func (m *ShellModel) startLocalCheck(name string) {
	tmpl := protocol.CheckTemplates[m.checkTemplateCursor]

	var participants []protocol.Participant
	for _, member := range m.home.team.Members {
		participants = append(participants, protocol.Participant{
			ID:   member.ID,
			Name: member.Name,
		})
	}

	var participantIDs []string
	for _, p := range participants {
		participantIDs = append(participantIDs, p.ID)
	}

	state := &protocol.RetroState{
		Stage: "icebreaker",
		Meta: protocol.RetroMeta{
			Type:       "check",
			Name:       name,
			TemplateID: tmpl.ID,
		},
		Participants: participants,
		Icebreaker: &protocol.Icebreaker{
			Questions: []string{
				"What is a book or show you would recommend right now?",
				"If you could have any superpower for a day, what would it be?",
				"What is something you learned recently that surprised you?",
				"If you could travel anywhere tomorrow, where would you go?",
				"What is your go-to comfort food?",
				"What is the best advice you have ever received?",
				"If you could master any skill instantly, what would it be?",
				"What is a small thing that made you happy this week?",
			},
			ParticipantIDs: participantIDs,
			CurrentIndex:   0,
		},
		Cards:           []protocol.Card{},
		Groups:          []protocol.Group{},
		Votes:           []protocol.Vote{},
		VoteBudget:      3,
		DiscussNotes:    []protocol.DiscussNote{},
		SurveyResponses: []protocol.SurveyResponse{},
	}

	var c *client.Client
	roomCode, err := client.CreateRoom(m.serverURL)
	if err == nil {
		c, err = client.Connect(roomCode, m.serverURL)
		if err == nil {
			c.SendState(state)
			teamInfo := protocol.SyncTeamInfo{TeamName: m.teamEntry.Name}
			for _, member := range m.home.team.Members {
				teamInfo.Members = append(teamInfo.Members, protocol.TeamInfoMember{ID: member.ID, Name: member.Name})
			}
			for _, ag := range m.home.team.Agreements {
				teamInfo.Agreements = append(teamInfo.Agreements, protocol.TeamInfoAgreement{ID: ag.ID, Text: ag.Text})
			}
			if msg, err := protocol.TeamInfoMessage(&teamInfo); err == nil {
				c.Send(msg)
			}
		}
	}

	sessionTeamInfo := &protocol.SyncTeamInfo{TeamName: m.teamEntry.Name}
	for _, member := range m.home.team.Members {
		sessionTeamInfo.Members = append(sessionTeamInfo.Members, protocol.TeamInfoMember{ID: member.ID, Name: member.Name})
	}
	for _, ag := range m.home.team.Agreements {
		sessionTeamInfo.Agreements = append(sessionTeamInfo.Agreements, protocol.TeamInfoAgreement{ID: ag.ID, Text: ag.Text})
	}

	m.session = Model{
		client:            c,
		state:             state,
		teamInfo:          sessionTeamInfo,
		defaultMemberName: m.registry.LoadDefaultMember(),
		serverURL:         m.serverURL,
		takenIDs:          make(map[string]bool),
		width:             m.width,
		height:            m.height,
	}

	if defaultName := m.session.defaultMemberName; defaultName != "" {
		for _, p := range state.Participants {
			if strings.EqualFold(p.Name, defaultName) {
				m.session.participantID = p.ID
				if c != nil {
					c.ClaimIdentity(p.ID)
				}
				break
			}
		}
	}

	m.mode = ModeSession
}

func (m ShellModel) viewNewCheck() string {
	accent := lipgloss.NewStyle().Foreground(styles.Accent).Bold(true)
	muted := lipgloss.NewStyle().Foreground(styles.Muted)

	var s string
	s += accent.Render("New Check") + "\n\n"

	if m.retroNameInput {
		tmpl := protocol.CheckTemplates[m.checkTemplateCursor]
		s += fmt.Sprintf("  Template: %s\n\n", accent.Render(tmpl.Name))
		s += fmt.Sprintf("  Check name: %s▌\n", m.retroName)
		s += "\n" + muted.Render("  [Enter] start  [Esc] back")
		return s
	}

	s += "  Pick a template:\n\n"
	for i, tmpl := range protocol.CheckTemplates {
		cursor := "  "
		if i == m.checkTemplateCursor {
			cursor = "> "
		}
		line := fmt.Sprintf("%s%s (%d questions)", cursor, tmpl.Name, len(tmpl.Questions))
		if i == m.checkTemplateCursor {
			s += styles.Selected.Render(line) + "\n"
			for _, q := range tmpl.Questions {
				s += muted.Render(fmt.Sprintf("    %s — %s", q.Title, q.Description)) + "\n"
			}
		} else {
			s += line + "\n"
		}
	}

	s += "\n" + muted.Render("  [↑↓] select  [Enter] choose  [Esc] back")
	return s
}

func (m *ShellModel) startLocalRetro(name string) {
	tmpl := protocol.Templates[m.templateCursor]

	// Build participants from team members
	var participants []protocol.Participant
	for _, member := range m.home.team.Members {
		participants = append(participants, protocol.Participant{
			ID:   member.ID,
			Name: member.Name,
		})
	}

	// Build participant IDs for icebreaker order
	var participantIDs []string
	for _, p := range participants {
		participantIDs = append(participantIDs, p.ID)
	}

	state := &protocol.RetroState{
		Stage: "icebreaker",
		Meta: protocol.RetroMeta{
			Type:       "retro",
			Name:       name,
			TemplateID: tmpl.ID,
		},
		Participants: participants,
		Icebreaker: &protocol.Icebreaker{
			Questions: []string{
				"What is a book or show you would recommend right now?",
				"If you could have any superpower for a day, what would it be?",
				"What is something you learned recently that surprised you?",
				"If you could travel anywhere tomorrow, where would you go?",
				"What is your go-to comfort food?",
				"What is the best advice you have ever received?",
				"If you could master any skill instantly, what would it be?",
				"What is a small thing that made you happy this week?",
			},
			ParticipantIDs: participantIDs,
			CurrentIndex:   0,
		},
		Cards:        []protocol.Card{},
		Groups:       []protocol.Group{},
		Votes:        []protocol.Vote{},
		VoteBudget:   3,
		DiscussNotes: []protocol.DiscussNote{},
	}

	// Try to create a room on the server for sharing
	var c *client.Client
	roomCode, err := client.CreateRoom(m.serverURL)
	if err == nil {
		c, err = client.Connect(roomCode, m.serverURL)
		if err == nil {
			// Broadcast initial state and team info
			c.SendState(state)
			teamInfo := protocol.SyncTeamInfo{
				TeamName: m.teamEntry.Name,
			}
			for _, member := range m.home.team.Members {
				teamInfo.Members = append(teamInfo.Members, protocol.TeamInfoMember{
					ID: member.ID, Name: member.Name,
				})
			}
			for _, ag := range m.home.team.Agreements {
				teamInfo.Agreements = append(teamInfo.Agreements, protocol.TeamInfoAgreement{
					ID: ag.ID, Text: ag.Text,
				})
			}
			if msg, err := protocol.TeamInfoMessage(&teamInfo); err == nil {
				c.Send(msg)
			}
		}
	}

	// Build team info for responding to request-state from peers
	sessionTeamInfo := &protocol.SyncTeamInfo{TeamName: m.teamEntry.Name}
	for _, member := range m.home.team.Members {
		sessionTeamInfo.Members = append(sessionTeamInfo.Members, protocol.TeamInfoMember{
			ID: member.ID, Name: member.Name,
		})
	}
	for _, ag := range m.home.team.Agreements {
		sessionTeamInfo.Agreements = append(sessionTeamInfo.Agreements, protocol.TeamInfoAgreement{
			ID: ag.ID, Text: ag.Text,
		})
	}

	m.session = Model{
		client:            c,
		state:             state,
		teamInfo:          sessionTeamInfo,
		defaultMemberName: m.registry.LoadDefaultMember(),
		serverURL:         m.serverURL,
		takenIDs:          make(map[string]bool),
		width:             m.width,
		height:            m.height,
	}

	// Auto-select default member if set
	if defaultName := m.session.defaultMemberName; defaultName != "" {
		for _, p := range state.Participants {
			if strings.EqualFold(p.Name, defaultName) {
				m.session.participantID = p.ID
				if c != nil {
					c.ClaimIdentity(p.ID)
				}
				break
			}
		}
	}

	m.mode = ModeSession
}

func (m ShellModel) openTeamSelect() (tea.Model, tea.Cmd) {
	entries, _ := m.registry.List()
	m.teamEntries = entries
	m.teamCursor = 0
	m.teamInputMode = false
	m.teamInput = ""
	m.teamAction = ""
	// Position cursor on currently selected team
	for i, e := range entries {
		if e.ID == m.teamEntry.ID {
			m.teamCursor = i
			break
		}
	}
	m.mode = ModeTeamSelect
	return m, nil
}

func (m ShellModel) updateTeamSelect(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.teamInputMode {
			return m.handleTeamInput(msg)
		}
		switch msg.String() {
		case "esc":
			if m.teamEntry.ID == "" {
				return m, tea.Quit // no team to go back to
			}
			m.mode = ModeHome
		case "up", "k":
			if m.teamCursor > 0 {
				m.teamCursor--
			}
		case "down", "j":
			if m.teamCursor < len(m.teamEntries)-1 {
				m.teamCursor++
			}
		case "enter":
			if m.teamCursor < len(m.teamEntries) {
				return m.selectTeam(m.teamEntries[m.teamCursor])
			}
		case "c":
			m.teamInputMode = true
			m.teamAction = "create"
			m.teamInput = ""
		case "r":
			if m.teamCursor < len(m.teamEntries) {
				m.teamInputMode = true
				m.teamAction = "rename"
				m.teamInput = m.teamEntries[m.teamCursor].Name
			}
		case "d":
			if m.teamCursor < len(m.teamEntries) {
				entry := m.teamEntries[m.teamCursor]
				m.teamEntries = domain.RemoveTeamEntry(m.teamEntries, entry.ID)
				m.registry.Save(m.teamEntries)
				// Remove data dir
				os.RemoveAll(m.registry.TeamDir(entry.ID))
				// Clear selection if deleted
				if entry.ID == m.teamEntry.ID {
					m.teamEntry = domain.TeamEntry{}
					m.registry.SetSelectedTeamID("")
				}
				if m.teamCursor >= len(m.teamEntries) && m.teamCursor > 0 {
					m.teamCursor--
				}
			}
		case "q":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m ShellModel) handleTeamInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		name := strings.TrimSpace(m.teamInput)
		if name != "" {
			switch m.teamAction {
			case "create":
				id := fmt.Sprintf("t-%d", len(m.teamEntries)+1)
				if entries, err := domain.AddTeamEntry(m.teamEntries, id, name, ""); err == nil {
					m.teamEntries = entries
					m.registry.Save(m.teamEntries)
					m.teamCursor = len(m.teamEntries) - 1
				}
			case "rename":
				if m.teamCursor < len(m.teamEntries) {
					id := m.teamEntries[m.teamCursor].ID
					if entries, err := domain.RenameTeamEntry(m.teamEntries, id, name); err == nil {
						m.teamEntries = entries
						m.registry.Save(m.teamEntries)
						// Update current entry if renamed
						if id == m.teamEntry.ID {
							m.teamEntry.Name = name
						}
					}
				}
			}
		}
		m.teamInputMode = false
		m.teamInput = ""
		m.teamAction = ""
	case "esc":
		m.teamInputMode = false
		m.teamInput = ""
		m.teamAction = ""
	case "backspace":
		if len(m.teamInput) > 0 {
			m.teamInput = m.teamInput[:len(m.teamInput)-1]
		}
	default:
		if len(msg.Runes) > 0 {
			m.teamInput += string(msg.Runes)
		}
	}
	return m, nil
}

func (m ShellModel) selectTeam(entry domain.TeamEntry) (tea.Model, tea.Cmd) {
	m.teamEntry = entry
	m.registry.SetSelectedTeamID(entry.ID)
	m.home = NewHomeModel(m.registry, entry)
	m.mode = ModeHome
	return m, nil
}

func (m ShellModel) viewTeamSelect() string {
	accent := lipgloss.NewStyle().Foreground(styles.Accent).Bold(true)
	muted := lipgloss.NewStyle().Foreground(styles.Muted)

	var s string
	s += accent.Render("Teams") + "\n"
	s += muted.Render(strings.Repeat("─", 40)) + "\n\n"

	if len(m.teamEntries) == 0 {
		s += muted.Render("  No teams yet. Press [c] to create one.") + "\n"
	}
	for i, entry := range m.teamEntries {
		cursor := "  "
		if i == m.teamCursor {
			cursor = "> "
		}
		marker := ""
		if entry.ID == m.teamEntry.ID {
			marker = "  " + accent.Render("*")
		}
		line := fmt.Sprintf("%s%s%s", cursor, entry.Name, marker)
		if i == m.teamCursor {
			s += styles.Selected.Render(line) + "\n"
		} else {
			s += line + "\n"
		}
	}

	s += "\n"
	if m.teamInputMode {
		label := m.teamAction
		s += fmt.Sprintf("  %s: %s▌\n", label, m.teamInput)
		s += muted.Render("  [Enter] save  [Esc] cancel") + "\n"
	} else {
		s += muted.Render("[↑↓] navigate  [Enter] select  [c] create  [d] delete  [r] rename  [Esc] back") + "\n"
	}

	return s
}

func (m *ShellModel) resolveOrCreateTeam(info *protocol.SyncTeamInfo) domain.TeamEntry {
	entries, _ := m.registry.List()

	// Find existing team by name
	if entry, ok := domain.FindTeamEntryByName(entries, info.TeamName); ok {
		return entry
	}

	// Create new team
	id := fmt.Sprintf("t-remote-%d", len(entries)+1)
	entries, err := domain.AddTeamEntry(entries, id, info.TeamName, "")
	if err != nil {
		return m.teamEntry // fallback to current
	}
	m.registry.Save(entries)
	return entries[len(entries)-1]
}

func (m *ShellModel) saveSessionToHistory() {
	state := m.session.state
	if state == nil {
		return
	}

	// Resolve target team: use team-info from remote if available
	targetEntry := m.teamEntry
	if m.session.teamInfo != nil && m.session.teamInfo.TeamName != "" {
		targetEntry = m.resolveOrCreateTeam(m.session.teamInfo)
	}

	if targetEntry.ID == "" {
		return
	}

	// Merge participants, members, and agreements into the target team
	teamDir := m.registry.TeamDir(targetEntry.ID)
	repo := storage.NewJSONTeamRepo(teamDir)
	team, _ := repo.LoadTeam()
	changed := false

	// Merge participants from retro session
	for _, p := range state.Participants {
		if updated, err := domain.AddMember(team, p.ID, p.Name); err == nil {
			team = updated
			changed = true
		}
	}

	// Merge members from team-info
	if m.session.teamInfo != nil {
		for _, member := range m.session.teamInfo.Members {
			if updated, err := domain.AddMember(team, member.ID, member.Name); err == nil {
				team = updated
				changed = true
			}
		}
		// Merge agreements from team-info
		for _, ag := range m.session.teamInfo.Agreements {
			if updated, err := domain.AddAgreement(team, ag.ID, ag.Text, ""); err == nil {
				team = updated
				changed = true
			}
		}
	}

	if changed {
		repo.SaveTeam(team)
	}

	// Only save if there are action items
	var actionItems []domain.FlatActionItem
	for _, n := range state.DiscussNotes {
		if n.Lane == "actions" {
			owner := ""
			if state.ActionOwners != nil {
				ownerID := state.ActionOwners[n.ID]
				if ownerID != "" {
					// Resolve owner name
					for _, p := range state.Participants {
						if p.ID == ownerID {
							owner = p.Name
							break
						}
					}
					if owner == "" {
						owner = ownerID
					}
				}
			}
			parentText := ""
			// Find parent card/group/question text
			if state.Meta.Type == "check" {
				tmpl := protocol.GetCheckTemplate(state.Meta.TemplateID)
				for _, q := range tmpl.Questions {
					if q.ID == n.ParentCardID {
						parentText = q.Title
						break
					}
				}
			}
			if parentText == "" {
				for _, g := range state.Groups {
					if g.ID == n.ParentCardID {
						parentText = g.Name
						break
					}
				}
			}
			if parentText == "" {
				for _, c := range state.Cards {
					if c.ID == n.ParentCardID {
						parentText = c.Text
						break
					}
				}
			}
			actionItems = append(actionItems, domain.FlatActionItem{
				NoteID:     n.ID,
				Text:       n.Text,
				ParentText: parentText,
				OwnerName:  owner,
			})
		}
	}

	if len(actionItems) == 0 && state.Meta.Name == "" {
		return // nothing worth saving
	}

	historyRepo := storage.NewJSONTeamRepo(m.registry.TeamDir(targetEntry.ID))
	history, _ := historyRepo.LoadHistory()

	retroID := state.Meta.Name
	if retroID == "" {
		retroID = fmt.Sprintf("retro-%d", len(history.Completed)+1)
	}

	entry := domain.CompletedRetro{
		ID:          retroID,
		CompletedAt: state.Meta.Date,
		ActionItems: actionItems,
		FullState:   *state,
	}
	history = domain.AddCompletedRetro(history, entry)
	historyRepo.SaveHistory(history)
}

func (m ShellModel) View() string {
	switch m.mode {
	case ModeJoinInput:
		return m.viewJoinInput()
	case ModeNewRetro:
		return m.viewNewRetro()
	case ModeNewCheck:
		return m.viewNewCheck()
	case ModeTeamSelect:
		return m.viewTeamSelect()
	case ModeSession:
		return m.session.View()
	default:
		return m.home.View()
	}
}

func (m ShellModel) viewNewRetro() string {
	accent := lipgloss.NewStyle().Foreground(styles.Accent).Bold(true)
	muted := lipgloss.NewStyle().Foreground(styles.Muted)

	var s string
	s += accent.Render("New Retrospective") + "\n\n"

	if m.retroNameInput {
		tmpl := protocol.Templates[m.templateCursor]
		s += fmt.Sprintf("  Template: %s\n\n", accent.Render(tmpl.Name))
		s += fmt.Sprintf("  Retro name: %s▌\n", m.retroName)
		s += "\n" + muted.Render("  [Enter] start  [Esc] back")
		return s
	}

	s += "  Pick a template:\n\n"
	for i, tmpl := range protocol.Templates {
		cursor := "  "
		if i == m.templateCursor {
			cursor = "> "
		}
		line := fmt.Sprintf("%s%s", cursor, tmpl.Name)
		if i == m.templateCursor {
			s += styles.Selected.Render(line) + "\n"
			// Show column descriptions
			for _, col := range tmpl.Columns {
				s += muted.Render(fmt.Sprintf("    %s — %s", col.Title, col.Description)) + "\n"
			}
		} else {
			s += line + "\n"
		}
	}

	s += "\n" + muted.Render("  [↑↓] select  [Enter] choose  [Esc] back")
	return s
}

func (m ShellModel) viewJoinInput() string {
	accent := lipgloss.NewStyle().Foreground(styles.Accent).Bold(true)
	muted := lipgloss.NewStyle().Foreground(styles.Muted)

	var s string
	s += accent.Render("Join Retrospective") + "\n\n"
	s += "  Enter room code or URL:\n\n"
	s += fmt.Sprintf("  > %s▌\n", m.joinInput)

	if m.joinErr != "" {
		s += "\n  " + lipgloss.NewStyle().Foreground(styles.Danger).Render(m.joinErr) + "\n"
	}

	s += "\n" + muted.Render("  [Enter] connect  [Esc] back")
	return s
}
