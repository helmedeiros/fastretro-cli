package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/helmedeiros/fastretro-cli/internal/client"
	"github.com/helmedeiros/fastretro-cli/internal/domain"
	"github.com/helmedeiros/fastretro-cli/internal/storage"
	"github.com/helmedeiros/fastretro-cli/internal/styles"
)

// ShellMode indicates which screen is active.
type ShellMode int

const (
	ModeHome ShellMode = iota
	ModeJoinInput
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
	joinInput string
	joinErr   string
	width     int
	height    int
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
	case ModeSession:
		return m.updateSession(msg)
	}
	return m, nil
}

func (m ShellModel) updateHome(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if !m.home.inputMode && msg.String() == "j" {
			m.mode = ModeJoinInput
			m.joinInput = ""
			m.joinErr = ""
			return m, nil
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
			if len(msg.String()) == 1 {
				m.joinInput += msg.String()
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
	case ModeSession:
		return m.session.View()
	default:
		return m.home.View()
	}
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
