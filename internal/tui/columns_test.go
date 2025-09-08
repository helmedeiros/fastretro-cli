package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/helmedeiros/fastretro-cli/internal/styles"
)

func TestJoinColumnsEqualHeight_EqualContent(t *testing.T) {
	contents := []string{"line1\nline2", "line1\nline2"}
	colStyles := []lipgloss.Style{styles.Column, styles.Column}

	result := joinColumnsEqualHeight(contents, colStyles)

	if result == "" {
		t.Error("expected non-empty result")
	}
}

func TestJoinColumnsEqualHeight_UnequalContent(t *testing.T) {
	short := "one line"
	long := "line1\nline2\nline3\nline4"
	contents := []string{short, long}
	colStyles := []lipgloss.Style{styles.Column, styles.Column}

	result := joinColumnsEqualHeight(contents, colStyles)

	// Both columns should have the same number of lines
	parts := strings.SplitN(result, "│", 3) // rough check that both render
	if len(parts) < 2 {
		t.Error("expected multiple columns in output")
	}
}

func TestJoinColumnsEqualHeight_SingleColumn(t *testing.T) {
	contents := []string{"content"}
	colStyles := []lipgloss.Style{styles.Column}

	result := joinColumnsEqualHeight(contents, colStyles)

	if !strings.Contains(result, "content") {
		t.Error("expected content in output")
	}
}

func TestJoinColumnsEqualHeight_EmptyContent(t *testing.T) {
	contents := []string{"", ""}
	colStyles := []lipgloss.Style{styles.Column, styles.Column}

	result := joinColumnsEqualHeight(contents, colStyles)

	if result == "" {
		t.Error("expected rendered empty columns")
	}
}

func TestJoinColumnsEqualHeight_ActiveStyle(t *testing.T) {
	contents := []string{"col1", "col2"}
	active := styles.Column.BorderForeground(styles.Accent)
	colStyles := []lipgloss.Style{active, styles.Column}

	result := joinColumnsEqualHeight(contents, colStyles)

	if !strings.Contains(result, "col1") || !strings.Contains(result, "col2") {
		t.Error("expected both columns")
	}
}
