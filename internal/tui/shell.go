package tui

import (
	"fmt"

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
	templateCursor int
	retroName      string
	retroNameInput bool
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
	m.session.width = m.width
	m.session.height = m.height
	m.mode = ModeSession
	return m, m.session.Init()
}

func (m ShellModel) updateSession(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Intercept ctrl+c in session to return to home instead of quitting
		if msg.String() == "ctrl+c" {
			m.saveSessionToHistory()
			m.mode = ModeHome
			m.home = NewHomeModel(m.registry, m.teamEntry) // refresh data
			return m, nil
		}
		// Intercept q when not in input mode
		if msg.String() == "q" && !m.session.inputMode {
			m.saveSessionToHistory()
			m.mode = ModeHome
			m.home = NewHomeModel(m.registry, m.teamEntry)
			return m, nil
		}
	}

	updated, cmd := m.session.Update(msg)
	m.session = updated.(Model)

	// Check if session issued a quit command
	if cmd != nil {
		// We can't easily inspect tea.Cmd, so let the session handle quit
		// The user will use ctrl+c or q to return to home
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

	state := &protocol.RetroState{
		Stage: "brainstorm",
		Meta: protocol.RetroMeta{
			Name:       name,
			TemplateID: tmpl.ID,
		},
		Participants: participants,
		Cards:        []protocol.Card{},
		Groups:       []protocol.Group{},
		Votes:        []protocol.Vote{},
		VoteBudget:   3,
		DiscussNotes: []protocol.DiscussNote{},
	}

	m.session = Model{
		state:    state,
		takenIDs: make(map[string]bool),
		width:    m.width,
		height:   m.height,
	}
	m.mode = ModeSession
}

func (m *ShellModel) saveSessionToHistory() {
	state := m.session.state
	if state == nil {
		return
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
			// Find parent card/group text
			for _, g := range state.Groups {
				if g.ID == n.ParentCardID {
					parentText = g.Name
					break
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

	teamDir := m.registry.TeamDir(m.teamEntry.ID)
	repo := storage.NewJSONTeamRepo(teamDir)
	history, _ := repo.LoadHistory()

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
	repo.SaveHistory(history)
}

func (m ShellModel) View() string {
	switch m.mode {
	case ModeJoinInput:
		return m.viewJoinInput()
	case ModeNewRetro:
		return m.viewNewRetro()
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
