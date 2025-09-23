package tui

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/helmedeiros/fastretro-cli/internal/client"
	"github.com/helmedeiros/fastretro-cli/internal/protocol"
	"github.com/helmedeiros/fastretro-cli/internal/styles"
)

// clearCopyMsg clears the copy confirmation after a delay.
type clearCopyMsg struct{}

func clearCopyAfter() tea.Cmd {
	return tea.Tick(2*time.Second, func(time.Time) tea.Msg {
		return clearCopyMsg{}
	})
}

func copyToClipboard(text string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("pbcopy")
	case "linux":
		cmd = exec.Command("xclip", "-selection", "clipboard")
	default:
		cmd = exec.Command("clip")
	}
	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}

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
	serverURL     string
	copyMsg       string
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
	case clearCopyMsg:
		m.copyMsg = ""
		return m, nil
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
	case "request-state":
		if m.client != nil && m.state != nil {
			m.client.SendState(m.state)
			if m.teamInfo != nil {
				if data, err := protocol.TeamInfoMessage(m.teamInfo); err == nil {
					m.client.Send(data)
				}
			}
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

	// Copy room code / share URL (available in all stages when not in input mode)
	if !m.inputMode && m.client != nil {
		switch msg.String() {
		case "c":
			if err := copyToClipboard(m.client.RoomCode); err == nil {
				m.copyMsg = "Room code copied!"
				return m, clearCopyAfter()
			}
		case "C":
			serverBase := m.serverURL
			if serverBase == "" {
				serverBase = "http://localhost:5173"
			}
			if err := copyToClipboard(m.client.ShareURL(serverBase)); err == nil {
				m.copyMsg = "Share URL copied!"
				return m, clearCopyAfter()
			}
		}
	}

	// Stage navigation (available in all stages when not in input mode)
	if !m.inputMode && m.state != nil {
		switch msg.String() {
		case "[":
			if idx := stageIndex(m.state.Stage); idx > 0 {
				m.state.Stage = allStages[idx-1]
				m.cursor = 0
				m.broadcastState()
			}
			return m, nil
		case "]":
			if idx := stageIndex(m.state.Stage); idx >= 0 && idx < len(allStages)-1 {
				m.state.Stage = allStages[idx+1]
				m.cursor = 0
				m.broadcastState()
			}
			return m, nil
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
		retroInfo += "  " + muted.Render(fmt.Sprintf("Room: %s | %d peers | [c] copy code  [C] copy URL", roomCode, m.peerCount))
	}

	// Show copy confirmation briefly
	copyMsg := ""
	if m.copyMsg != "" {
		copyMsg = "  " + lipgloss.NewStyle().Foreground(styles.Success).Render(m.copyMsg)
	}

	header := retroInfo + copyMsg + "\n" + m.renderStageBar()

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

func stageIndex(stage string) int {
	for i, s := range allStages {
		if s == stage {
			return i
		}
	}
	return -1
}

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
	bar := strings.Join(parts, "  ")
	muted := lipgloss.NewStyle().Foreground(styles.Muted)
	bar += "  " + muted.Render("← [  ] →")
	return bar
}
