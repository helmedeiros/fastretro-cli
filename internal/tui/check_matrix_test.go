package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/helmedeiros/fastretro-cli/internal/domain"
	"github.com/helmedeiros/fastretro-cli/internal/protocol"
)

func testCheckHistory() domain.RetroHistoryState {
	// History stores entries in append order (oldest first)
	return domain.RetroHistoryState{
		Completed: []domain.CompletedRetro{
			{
				ID:          "hc2",
				CompletedAt: "2026-03-20",
				ActionItems: []domain.FlatActionItem{},
				FullState: protocol.RetroState{
					Stage: "close",
					Meta:  protocol.RetroMeta{Type: "check", Name: "HC March", Date: "2026-03-20", TemplateID: "health-check"},
					Participants: []protocol.Participant{
						{ID: "p1", Name: "Alice"},
					},
					SurveyResponses: []protocol.SurveyResponse{
						{ID: "r5", ParticipantID: "p1", QuestionID: "ownership", Rating: 3, Comment: ""},
					},
				},
			},
			{
				ID:          "retro1",
				CompletedAt: "2026-04-15",
				FullState: protocol.RetroState{
					Stage: "close",
					Meta:  protocol.RetroMeta{Type: "retro", Name: "Sprint 1", TemplateID: "start-stop"},
				},
			},
			{
				ID:          "hc1",
				CompletedAt: "2026-04-20",
				ActionItems: []domain.FlatActionItem{{NoteID: "a1", Text: "Fix ownership"}},
				FullState: protocol.RetroState{
					Stage: "close",
					Meta:  protocol.RetroMeta{Type: "check", Name: "HC April", Date: "2026-04-20", TemplateID: "health-check"},
					Participants: []protocol.Participant{
						{ID: "p1", Name: "Alice"},
						{ID: "p2", Name: "Bob"},
					},
					SurveyResponses: []protocol.SurveyResponse{
						{ID: "r1", ParticipantID: "p1", QuestionID: "ownership", Rating: 4, Comment: ""},
						{ID: "r2", ParticipantID: "p2", QuestionID: "ownership", Rating: 2, Comment: ""},
						{ID: "r3", ParticipantID: "p1", QuestionID: "value", Rating: 5, Comment: ""},
						{ID: "r4", ParticipantID: "p2", QuestionID: "value", Rating: 3, Comment: ""},
					},
				},
			},
		},
	}
}

func TestNewCheckMatrixModel(t *testing.T) {
	m := NewCheckMatrixModel(testCheckHistory())
	if len(m.templates) == 0 {
		t.Fatal("expected templates")
	}
	if m.tmplCursor != 0 {
		t.Errorf("expected tmplCursor 0, got %d", m.tmplCursor)
	}
	if m.colCursor != 0 {
		t.Errorf("expected colCursor 0, got %d", m.colCursor)
	}
}

func TestCheckMatrix_Sessions_FiltersByTemplate(t *testing.T) {
	m := NewCheckMatrixModel(testCheckHistory())
	sessions := m.sessions()
	// Should only include health-check sessions, not retros
	if len(sessions) != 2 {
		t.Errorf("expected 2 health-check sessions, got %d", len(sessions))
	}
	// Should be in reverse order (newest first)
	if sessions[0].FullState.Meta.Name != "HC April" {
		t.Errorf("expected newest first, got %q", sessions[0].FullState.Meta.Name)
	}
}

func TestCheckMatrix_Sessions_EmptyForUnusedTemplate(t *testing.T) {
	m := NewCheckMatrixModel(testCheckHistory())
	m.tmplCursor = 1 // DORA template
	sessions := m.sessions()
	if len(sessions) != 0 {
		t.Errorf("expected 0 DORA sessions, got %d", len(sessions))
	}
}

func TestCheckMatrix_HandleKey_LeftRight(t *testing.T) {
	m := NewCheckMatrixModel(testCheckHistory())
	result, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	m = result.(CheckMatrixModel)
	if m.colCursor != 1 {
		t.Errorf("expected colCursor 1 after right, got %d", m.colCursor)
	}
	result, _ = m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	m = result.(CheckMatrixModel)
	if m.colCursor != 0 {
		t.Errorf("expected colCursor 0 after left, got %d", m.colCursor)
	}
}

func TestCheckMatrix_HandleKey_LeftBound(t *testing.T) {
	m := NewCheckMatrixModel(testCheckHistory())
	result, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	m = result.(CheckMatrixModel)
	if m.colCursor != 0 {
		t.Errorf("expected colCursor to stay at 0, got %d", m.colCursor)
	}
}

func TestCheckMatrix_HandleKey_Tab(t *testing.T) {
	m := NewCheckMatrixModel(testCheckHistory())
	if m.tmplCursor != 0 {
		t.Fatal("expected starting at template 0")
	}
	result, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyTab})
	m = result.(CheckMatrixModel)
	if m.tmplCursor != 1 {
		t.Errorf("expected tmplCursor 1, got %d", m.tmplCursor)
	}
	if m.colCursor != 0 {
		t.Errorf("expected colCursor reset to 0, got %d", m.colCursor)
	}
}

func TestCheckMatrix_HandleKey_TabWraps(t *testing.T) {
	m := NewCheckMatrixModel(testCheckHistory())
	for range m.templates {
		result, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyTab})
		m = result.(CheckMatrixModel)
	}
	if m.tmplCursor != 0 {
		t.Errorf("expected wrap to 0, got %d", m.tmplCursor)
	}
}

func TestCheckMatrix_HandleKey_Enter(t *testing.T) {
	m := NewCheckMatrixModel(testCheckHistory())
	_, cmd := m.handleKey(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected command from Enter")
	}
	msg := cmd()
	if _, ok := msg.(ViewHistoryMsg); !ok {
		t.Errorf("expected ViewHistoryMsg, got %T", msg)
	}
}

func TestCheckMatrix_HandleKey_Quit(t *testing.T) {
	m := NewCheckMatrixModel(testCheckHistory())
	_, cmd := m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Fatal("expected command from q")
	}
	msg := cmd()
	if _, ok := msg.(checkMatrixDoneMsg); !ok {
		t.Errorf("expected checkMatrixDoneMsg, got %T", msg)
	}
}

func TestCheckMatrix_HandleKey_Esc(t *testing.T) {
	m := NewCheckMatrixModel(testCheckHistory())
	_, cmd := m.handleKey(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("expected command from Esc")
	}
}

func TestCheckMatrix_View_ShowsTemplateTabs(t *testing.T) {
	m := NewCheckMatrixModel(testCheckHistory())
	view := m.View()
	if !strings.Contains(view, "Health Check") {
		t.Error("expected Health Check in view")
	}
	if !strings.Contains(view, "DORA Metrics Quiz") {
		t.Error("expected DORA Metrics Quiz in view")
	}
}

func TestCheckMatrix_View_ShowsScores(t *testing.T) {
	m := NewCheckMatrixModel(testCheckHistory())
	view := m.View()
	if !strings.Contains(view, "Ownership") {
		t.Error("expected Ownership question in view")
	}
	if !strings.Contains(view, "Overall") {
		t.Error("expected Overall row in view")
	}
}

func TestCheckMatrix_View_ShowsSessionHeaders(t *testing.T) {
	m := NewCheckMatrixModel(testCheckHistory())
	view := m.View()
	if !strings.Contains(view, "HC April") {
		t.Error("expected session name in view")
	}
}

func TestCheckMatrix_View_EmptyTemplate(t *testing.T) {
	m := NewCheckMatrixModel(testCheckHistory())
	m.tmplCursor = 1 // DORA — no sessions
	view := m.View()
	if !strings.Contains(view, "No completed sessions") {
		t.Error("expected empty message")
	}
}

func TestCheckMatrix_View_HelpAtBottom(t *testing.T) {
	m := NewCheckMatrixModel(testCheckHistory())
	view := m.View()
	lines := strings.Split(view, "\n")
	// Help should be near the end
	found := false
	for _, line := range lines[len(lines)-3:] {
		if strings.Contains(line, "[Esc] back") {
			found = true
		}
	}
	if !found {
		t.Error("expected help text near bottom of view")
	}
}

func TestMedianFromResponses(t *testing.T) {
	responses := []protocol.SurveyResponse{
		{QuestionID: "q1", Rating: 1},
		{QuestionID: "q1", Rating: 3},
		{QuestionID: "q1", Rating: 5},
		{QuestionID: "q2", Rating: 2},
		{QuestionID: "q2", Rating: 4},
	}
	if got := medianFromResponses(responses, "q1"); got != 3 {
		t.Errorf("q1 median: got %.1f, want 3", got)
	}
	if got := medianFromResponses(responses, "q2"); got != 3 {
		t.Errorf("q2 median: got %.1f, want 3", got)
	}
	if got := medianFromResponses(responses, "missing"); got != 0 {
		t.Errorf("missing median: got %.1f, want 0", got)
	}
}

func TestMedianFromResponses_SkipsZeroRating(t *testing.T) {
	responses := []protocol.SurveyResponse{
		{QuestionID: "q1", Rating: 0},
		{QuestionID: "q1", Rating: 4},
	}
	if got := medianFromResponses(responses, "q1"); got != 4 {
		t.Errorf("got %.1f, want 4", got)
	}
}

func TestBuildCheckDiscussState(t *testing.T) {
	m := checkSurveyModel()
	m.state.SurveyResponses = []protocol.SurveyResponse{
		{ID: "r1", ParticipantID: "p1", QuestionID: "ownership", Rating: 4},
		{ID: "r2", ParticipantID: "p1", QuestionID: "value", Rating: 1},
		{ID: "r3", ParticipantID: "p1", QuestionID: "fun", Rating: 5},
	}
	ds := m.buildCheckDiscussState()
	if ds == nil {
		t.Fatal("expected discuss state")
	}
	if len(ds.Order) != 9 {
		t.Errorf("expected 9 questions in order, got %d", len(ds.Order))
	}
	if ds.Segment != "actions" {
		t.Errorf("expected actions segment, got %q", ds.Segment)
	}
	// First should be lowest median — unrated (0) questions come first
	// value (1) should come before ownership (4) which should come before fun (5)
	valueIdx := -1
	ownershipIdx := -1
	funIdx := -1
	for i, id := range ds.Order {
		switch id {
		case "value":
			valueIdx = i
		case "ownership":
			ownershipIdx = i
		case "fun":
			funIdx = i
		}
	}
	if valueIdx >= ownershipIdx {
		t.Errorf("value (median 1) should come before ownership (median 4)")
	}
	if ownershipIdx >= funIdx {
		t.Errorf("ownership (median 4) should come before fun (median 5)")
	}
}
