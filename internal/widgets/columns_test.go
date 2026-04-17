package widgets

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestJoinColumnsEqualHeight_MatchesHeight(t *testing.T) {
	style := lipgloss.NewStyle().Width(10)
	contents := []string{"A\nB", "C"}
	result := JoinColumnsEqualHeight(contents, []lipgloss.Style{style, style})
	// Both columns should have the same number of lines
	lines := strings.Split(result, "\n")
	if len(lines) < 2 {
		t.Errorf("expected at least 2 lines, got %d", len(lines))
	}
}

func TestJoinColumnsEqualHeight_PreservesContent(t *testing.T) {
	style := lipgloss.NewStyle().Width(10)
	contents := []string{"hello", "world"}
	result := JoinColumnsEqualHeight(contents, []lipgloss.Style{style, style})
	if !strings.Contains(result, "hello") || !strings.Contains(result, "world") {
		t.Error("expected both columns' content in output")
	}
}

func TestJoinColumnsEqualHeight_SingleColumn(t *testing.T) {
	style := lipgloss.NewStyle().Width(10)
	result := JoinColumnsEqualHeight([]string{"one"}, []lipgloss.Style{style})
	if !strings.Contains(result, "one") {
		t.Error("expected content in single column")
	}
}
