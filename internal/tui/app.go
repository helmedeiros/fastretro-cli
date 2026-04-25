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
	client            *client.Client
	state             *protocol.RetroState
	participantID     string
	defaultMemberName string // auto-select participant matching this name
	takenIDs          map[string]bool
	peerCount         int
	cursor            int
	inputText         string
	inputMode         bool
	reviewPickMode    bool // true when showing participant picker
	reviewPickCursor  int  // cursor within participant list
	activeCol         int
	mergeSource       string // card ID selected as merge source
	teamInfo          *protocol.SyncTeamInfo
	serverURL         string
	copyMsg           string
	err               error
	width             int
	height            int
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

// SetParticipantID sets the chosen identity (used to restore persisted identity).
func (m *Model) SetParticipantID(id string) {
	m.participantID = id
}

// ParticipantID returns the current participant identity.
func (m Model) ParticipantID() string {
	return m.participantID
}

// SetDefaultMemberName sets the name to auto-match when state arrives.
func (m *Model) SetDefaultMemberName(name string) {
	m.defaultMemberName = name
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
			// Auto-select default member if no identity picked yet
			if m.participantID == "" && m.defaultMemberName != "" {
				for _, p := range m.state.Participants {
					if strings.EqualFold(p.Name, m.defaultMemberName) && !m.takenIDs[p.ID] {
						m.participantID = p.ID
						if m.client != nil {
							m.client.ClaimIdentity(p.ID)
						}
						break
					}
				}
			}
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
		stages := m.stagesForType()
		switch msg.String() {
		case "[":
			if idx := stageIndexIn(m.state.Stage, stages); idx > 0 {
				m.state.Stage = stages[idx-1]
				m.cursor = 0
				m.broadcastState()
			}
			return m, nil
		case "]":
			if idx := stageIndexIn(m.state.Stage, stages); idx >= 0 && idx < len(stages)-1 {
				m.state.Stage = stages[idx+1]
				m.cursor = 0
				m.initStage()
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
		case "icebreaker":
			return m.handleIcebreakerKeys(msg)
		case "brainstorm":
			return m.handleBrainstormKeys(msg)
		case "group":
			return m.handleGroupKeys(msg)
		case "vote":
			return m.handleVoteKeys(msg)
		case "survey":
			return m.handleSurveyKeys(msg)
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
		youLabel := ""
		if m.participantID != "" {
			youLabel = "You: " + m.participantName(m.participantID) + " | "
		}
		retroInfo += "  " + muted.Render(fmt.Sprintf("Room: %s | %d peers | %s[c] copy code  [C] copy URL", roomCode, m.peerCount, youLabel))
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
	case "survey":
		body = m.viewSurvey()
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

// initStage initializes stage-specific state when entering a new stage.
func (m *Model) initStage() {
	if m.state == nil {
		return
	}
	switch m.state.Stage {
	case "discuss":
		if m.state.Discuss == nil {
			m.state.Discuss = m.buildDiscussState()
		}
	}
}

// buildDiscussState creates the discuss order. For retros: votable items by votes desc.
// For checks: template questions by median asc (worst first).
func (m Model) buildDiscussState() *protocol.DiscussState {
	if m.state.Meta.Type == "check" {
		return m.buildCheckDiscussState()
	}

	type votable struct {
		id    string
		votes int
	}

	grouped := m.groupedCardIDs()
	var items []votable

	for _, g := range m.state.Groups {
		items = append(items, votable{id: g.ID, votes: m.votesForItem(g.ID)})
	}
	for _, c := range m.state.Cards {
		if !grouped[c.ID] {
			items = append(items, votable{id: c.ID, votes: m.votesForItem(c.ID)})
		}
	}

	for i := 1; i < len(items); i++ {
		for j := i; j > 0 && items[j].votes > items[j-1].votes; j-- {
			items[j], items[j-1] = items[j-1], items[j]
		}
	}

	var order []string
	for _, item := range items {
		order = append(order, item.id)
	}

	return &protocol.DiscussState{
		Order:        order,
		CurrentIndex: 0,
		Segment:      "context",
	}
}

func (m Model) buildCheckDiscussState() *protocol.DiscussState {
	tmpl := protocol.GetCheckTemplate(m.state.Meta.TemplateID)

	type item struct {
		id     string
		median float64
	}

	var items []item
	for _, q := range tmpl.Questions {
		items = append(items, item{id: q.ID, median: m.medianForItem(q.ID)})
	}

	// Sort ascending by median (worst first)
	for i := 1; i < len(items); i++ {
		for j := i; j > 0 && items[j].median < items[j-1].median; j-- {
			items[j], items[j-1] = items[j-1], items[j]
		}
	}

	var order []string
	for _, it := range items {
		order = append(order, it.id)
	}

	return &protocol.DiscussState{
		Order:        order,
		CurrentIndex: 0,
		Segment:      "actions",
	}
}

var retroStages = []string{"icebreaker", "brainstorm", "group", "vote", "discuss", "review", "close"}
var checkStages = []string{"icebreaker", "survey", "discuss", "review", "close"}

func (m Model) stagesForType() []string {
	if m.state != nil && m.state.Meta.Type == "check" {
		return checkStages
	}
	return retroStages
}

func stageIndexIn(stage string, stages []string) int {
	for i, s := range stages {
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

	stages := m.stagesForType()
	active := lipgloss.NewStyle().Foreground(styles.Accent).Bold(true).Underline(true)
	inactive := lipgloss.NewStyle().Foreground(styles.Muted)

	var parts []string
	for _, s := range stages {
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
