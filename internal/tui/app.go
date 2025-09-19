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
	mergeSource   string // card ID selected as merge source
	teamInfo      *protocol.SyncTeamInfo
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
	case "team-info":
		if msg.TeamInfo != nil {
			m.teamInfo = msg.TeamInfo
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
		case "group":
			return m.handleGroupKeys(msg)
		case "vote":
			return m.handleVoteKeys(msg)
		case "discuss":
			return m.handleDiscussKeys(msg)
		case "review":
			return m.handleReviewKeys(msg)
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
	// Retro name and date
	titleStyle := lipgloss.NewStyle().Foreground(styles.Accent).Bold(true)
	muted := lipgloss.NewStyle().Foreground(styles.Muted)
	retroTitle := "fastRetro CLI"
	if m.state.Meta.Name != "" {
		retroTitle = m.state.Meta.Name
	}
	retroInfo := titleStyle.Render(retroTitle)
	if m.state.Meta.Date != "" {
		retroInfo += "  " + muted.Render(m.state.Meta.Date)
	}
	if roomCode != "" {
		retroInfo += "  " + muted.Render(fmt.Sprintf("Room: %s | %d peers", roomCode, m.peerCount))
	}

	header := retroInfo + "\n" + m.renderStageBar()

	var body string
	switch m.state.Stage {
	case "icebreaker":
		body = m.viewIcebreaker()
	case "brainstorm":
		body = m.viewBrainstorm()
	case "group":
		body = m.viewGroup()
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

// joinColumnsEqualHeight renders column content strings inside styles.Column
// with equal height, then joins them horizontally.
func joinColumnsEqualHeight(contents []string, colStyles []lipgloss.Style) string {
	// First render to measure heights
	var rendered []string
	maxH := 0
	for i, content := range contents {
		r := colStyles[i].Render(content)
		h := strings.Count(r, "\n") + 1
		if h > maxH {
			maxH = h
		}
		rendered = append(rendered, r)
	}
	// Re-render with equal height
	rendered = nil
	for i, content := range contents {
		r := colStyles[i].Height(maxH).Render(content)
		rendered = append(rendered, r)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, rendered...)
}

var allStages = []string{"icebreaker", "brainstorm", "group", "vote", "discuss", "review", "close"}

func (m Model) renderStageBar() string {
	if m.state == nil {
		return ""
	}

	active := lipgloss.NewStyle().Foreground(styles.Accent).Bold(true).Underline(true)
	inactive := lipgloss.NewStyle().Foreground(styles.Muted)

	var parts []string
	for _, s := range allStages {
		label := strings.ToUpper(s)
		if s == m.state.Stage {
			parts = append(parts, active.Render(label))
		} else {
			parts = append(parts, inactive.Render(label))
		}
	}
	return strings.Join(parts, "  ")
}
