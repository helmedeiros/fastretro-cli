package tui

import (
	"fmt"
	"math"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/helmedeiros/fastretro-cli/internal/domain"
	"github.com/helmedeiros/fastretro-cli/internal/protocol"
	"github.com/helmedeiros/fastretro-cli/internal/styles"
	"github.com/helmedeiros/fastretro-cli/internal/widgets"
)

// CheckMatrixModel displays a comparison matrix of check sessions.
type CheckMatrixModel struct {
	history    domain.RetroHistoryState
	templates  []protocol.CheckTemplate
	tmplCursor int // which template is selected
	colCursor  int // which session column is highlighted
	width      int
	height     int
}

// NewCheckMatrixModel creates a matrix view from history.
func NewCheckMatrixModel(history domain.RetroHistoryState) CheckMatrixModel {
	return CheckMatrixModel{
		history:   history,
		templates: protocol.CheckTemplates,
		width:     80,
		height:    24,
	}
}

func (m CheckMatrixModel) sessions() []domain.CompletedRetro {
	tmpl := m.templates[m.tmplCursor]
	var result []domain.CompletedRetro
	for i := len(m.history.Completed) - 1; i >= 0; i-- {
		r := m.history.Completed[i]
		if r.FullState.Meta.Type == "check" && r.FullState.Meta.TemplateID == tmpl.ID {
			result = append(result, r)
		}
	}
	return result
}

func (m CheckMatrixModel) Init() tea.Cmd { return nil }

func (m CheckMatrixModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m CheckMatrixModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	sessions := m.sessions()
	switch msg.String() {
	case "q", "esc":
		return m, func() tea.Msg { return checkMatrixDoneMsg{} }
	case "left", "h":
		if m.colCursor > 0 {
			m.colCursor--
		}
	case "right", "l":
		if m.colCursor < len(sessions)-1 {
			m.colCursor++
		}
	case "tab":
		m.tmplCursor = (m.tmplCursor + 1) % len(m.templates)
		m.colCursor = 0
	case "enter":
		if m.colCursor < len(sessions) {
			state := sessions[m.colCursor].FullState
			return m, func() tea.Msg { return ViewHistoryMsg{State: &state} }
		}
	}
	return m, nil
}

type checkMatrixDoneMsg struct{}

func medianFromResponses(responses []protocol.SurveyResponse, questionID string) float64 {
	var ratings []int
	for _, r := range responses {
		if r.QuestionID == questionID && r.Rating > 0 {
			ratings = append(ratings, r.Rating)
		}
	}
	return widgets.MedianInt(ratings)
}

func scoreStyle(score float64, maxLevel int) lipgloss.Style {
	base := lipgloss.NewStyle().Width(10).Align(lipgloss.Center).Bold(true)
	if score == 0 {
		return base.Foreground(styles.Muted)
	}
	ratio := score / float64(maxLevel)
	if ratio >= 0.8 {
		return base.Foreground(styles.Success)
	}
	if ratio >= 0.6 {
		return base.Foreground(lipgloss.Color("#b4c850"))
	}
	if ratio >= 0.4 {
		return base.Foreground(lipgloss.Color("#d4a84e"))
	}
	return base.Foreground(styles.Danger)
}

func (m CheckMatrixModel) View() string {
	accent := lipgloss.NewStyle().Foreground(styles.Accent).Bold(true)
	muted := lipgloss.NewStyle().Foreground(styles.Muted)

	tmpl := m.templates[m.tmplCursor]
	sessions := m.sessions()
	maxLevel := 1
	for _, q := range tmpl.Questions {
		for _, o := range q.Options {
			if o.Value > maxLevel {
				maxLevel = o.Value
			}
		}
	}

	var b strings.Builder

	// Title
	b.WriteString(accent.Render("Check Comparison"))
	b.WriteString("\n\n")

	// Template tabs
	activeTab := lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.Accent).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Accent).
		Padding(0, 2)
	inactiveTab := lipgloss.NewStyle().
		Foreground(styles.Muted).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Border).
		Padding(0, 2)

	var tabs []string
	for i, t := range m.templates {
		// Count sessions for this template
		count := 0
		for j := len(m.history.Completed) - 1; j >= 0; j-- {
			r := m.history.Completed[j]
			if r.FullState.Meta.Type == "check" && r.FullState.Meta.TemplateID == t.ID {
				count++
			}
		}
		label := fmt.Sprintf("%s (%d)", t.Name, count)
		if i == m.tmplCursor {
			tabs = append(tabs, activeTab.Render(label))
		} else {
			tabs = append(tabs, inactiveTab.Render(label))
		}
	}
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, tabs...))
	b.WriteString("\n\n")

	if len(sessions) == 0 {
		b.WriteString(muted.Render("  No completed sessions for this template."))
		b.WriteString("\n\n")
		b.WriteString(muted.Render("[Tab] template  [Esc] back"))
		return b.String()
	}

	// Column headers
	qColWidth := 24
	cellWidth := 10
	header := lipgloss.NewStyle().Width(qColWidth).Render("")
	for i, s := range sessions {
		name := s.FullState.Meta.Name
		if name == "" {
			name = s.ID
		}
		if len(name) > cellWidth-1 {
			name = name[:cellWidth-2] + ".."
		}

		colStyle := lipgloss.NewStyle().Width(cellWidth).Align(lipgloss.Center)
		if i == m.colCursor {
			colStyle = colStyle.Foreground(styles.Accent).Bold(true).Underline(true)
		} else {
			colStyle = colStyle.Foreground(styles.Muted)
		}
		header += colStyle.Render(name)
	}
	b.WriteString(header)
	b.WriteString("\n")

	// Date row
	dateRow := lipgloss.NewStyle().Width(qColWidth).Render("")
	for i, s := range sessions {
		date := s.FullState.Meta.Date
		if date == "" {
			date = s.CompletedAt
		}
		if len(date) > 10 {
			date = date[:10]
		}
		colStyle := lipgloss.NewStyle().Width(cellWidth).Align(lipgloss.Center).Foreground(styles.Muted)
		if i == m.colCursor {
			colStyle = colStyle.Foreground(styles.Accent)
		}
		dateRow += colStyle.Render(date)
	}
	b.WriteString(dateRow)
	b.WriteString("\n")
	b.WriteString(muted.Render(strings.Repeat("─", qColWidth+cellWidth*len(sessions))))
	b.WriteString("\n")

	// Question rows
	for _, q := range tmpl.Questions {
		title := q.Title
		if len(title) > qColWidth-2 {
			title = title[:qColWidth-3] + ".."
		}
		row := lipgloss.NewStyle().Width(qColWidth).Render(title)

		for _, s := range sessions {
			median := medianFromResponses(s.FullState.SurveyResponses, q.ID)
			cellText := "—"
			if median > 0 {
				cellText = fmt.Sprintf("%.1f", median)
			}
			row += scoreStyle(median, maxLevel).Render(cellText)
		}
		b.WriteString(row)
		b.WriteString("\n")
	}

	// Overall score row
	b.WriteString(muted.Render(strings.Repeat("─", qColWidth+cellWidth*len(sessions))))
	b.WriteString("\n")
	overallRow := lipgloss.NewStyle().Width(qColWidth).Bold(true).Render("Overall")
	for _, s := range sessions {
		var sum float64
		var count int
		for _, q := range tmpl.Questions {
			med := medianFromResponses(s.FullState.SurveyResponses, q.ID)
			if med > 0 {
				sum += med
				count++
			}
		}
		overall := 0.0
		if count > 0 {
			overall = math.Round(sum/float64(count)*10) / 10
		}
		cellText := "—"
		if overall > 0 {
			cellText = fmt.Sprintf("%.1f", overall)
		}
		overallRow += scoreStyle(overall, maxLevel).Render(cellText)
	}
	b.WriteString(overallRow)
	b.WriteString("\n\n")

	// Radar chart for selected session
	if m.colCursor < len(sessions) {
		selected := sessions[m.colCursor]
		selName := selected.FullState.Meta.Name
		if selName == "" {
			selName = selected.ID
		}
		b.WriteString(accent.Render(fmt.Sprintf("  %s", selName)))
		b.WriteString("\n\n")

		var labels []string
		var values []float64
		for _, q := range tmpl.Questions {
			labels = append(labels, q.Title)
			values = append(values, medianFromResponses(selected.FullState.SurveyResponses, q.ID))
		}
		b.WriteString(widgets.RadarChart(labels, values, maxLevel, 8))
	}

	b.WriteString(muted.Render("[h/l] select  [Tab] template  [Enter] view  [Esc] back"))
	b.WriteString("\n")

	return b.String()
}
