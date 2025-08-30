package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/helmedeiros/fastretro-cli/internal/client"
	"github.com/helmedeiros/fastretro-cli/internal/protocol"
	"github.com/helmedeiros/fastretro-cli/internal/styles"
)

// WSMsg wraps an incoming WebSocket message for Bubble Tea.
type WSMsg protocol.IncomingMessage

// ErrMsg wraps an error.
type ErrMsg struct{ Err error }

// Model is the main TUI model.
type Model struct {
	client        *client.Client
	state         *protocol.RetroState
	participantID string
	takenIDs      map[string]bool
	peerCount     int
	cursor        int
	inputText     string
	inputMode     bool
	activeCol     int
	err           error
	width         int
	height        int
}

// NewModel creates the initial model.
func NewModel(c *client.Client) Model {
	return Model{
		client:   c,
		takenIDs: make(map[string]bool),
		width:    80,
		height:   24,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		listenWS(m.client),
		requestState(m.client),
	)
}

func requestState(c *client.Client) tea.Cmd {
	return func() tea.Msg {
		if c != nil {
			c.RequestState()
		}
		return nil
	}
}

func listenWS(c *client.Client) tea.Cmd {
	return func() tea.Msg {
		msg, err := c.ReadMessage()
		if err != nil {
			return ErrMsg{Err: err}
		}
		return WSMsg(msg)
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)

	case WSMsg:
		return m.handleWS(protocol.IncomingMessage(msg))

	case ErrMsg:
		m.err = msg.Err
		return m, nil
	}
	return m, nil
}

func (m Model) handleWS(msg protocol.IncomingMessage) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case "state":
		if msg.State != nil {
			m.state = msg.State
		}
	case "peer-count":
		m.peerCount = msg.Count
	case "taken-ids":
		m.takenIDs = make(map[string]bool)
		for _, id := range msg.IDs {
			m.takenIDs[id] = true
		}
	case "navigate-stage":
		if m.state != nil {
			m.state.Stage = msg.Stage
		}
	}
	return m, listenWS(m.client)
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		if !m.inputMode {
			return m, tea.Quit
		}
	}

	// Identity selection
	if m.participantID == "" && m.state != nil {
		return m.handleJoinKeys(msg)
	}

	// Stage-specific
	if m.state != nil {
		switch m.state.Stage {
		case "brainstorm":
			return m.handleBrainstormKeys(msg)
		case "vote":
			return m.handleVoteKeys(msg)
		}
	}

	return m, nil
}

func (m Model) View() string {
	if m.err != nil {
		return styles.Title.Render("fastRetro CLI") + "\n\n" +
			lipgloss.NewStyle().Foreground(styles.Danger).Render(fmt.Sprintf("Error: %v", m.err)) + "\n"
	}

	if m.state == nil {
		roomCode := ""
		if m.client != nil {
			roomCode = m.client.RoomCode
		}
		return styles.Title.Render("fastRetro CLI") + "\n\n" +
			styles.Subtitle.Render(fmt.Sprintf("Connected to room %s — waiting for state...", roomCode)) + "\n" +
			styles.StatusBar.Render(fmt.Sprintf("%d peers in room", m.peerCount))
	}

	// Identity selection
	if m.participantID == "" {
		return m.viewJoin()
	}

	// Render based on stage
	roomCode := ""
	if m.client != nil {
		roomCode = m.client.RoomCode
	}
	header := styles.Title.Render("fastRetro CLI") + "  " +
		styles.StatusBar.Render(fmt.Sprintf("Room: %s | %d peers | Stage: %s",
			roomCode, m.peerCount, strings.ToUpper(m.state.Stage)))

	var body string
	switch m.state.Stage {
	case "icebreaker":
		body = m.viewIcebreaker()
	case "brainstorm", "group":
		body = m.viewBrainstorm()
	case "vote":
		body = m.viewVote()
	case "discuss":
		body = m.viewDiscuss()
	case "review":
		body = m.viewReview()
	case "close":
		body = m.viewClose()
	default:
		body = styles.Subtitle.Render(fmt.Sprintf("Stage: %s — view-only in CLI", m.state.Stage))
	}

	return header + "\n\n" + body + "\n"
}
