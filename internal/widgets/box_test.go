package widgets

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

var testCfg = BoxConfig{
	ActiveColor:   lipgloss.Color("#5ec4c8"),
	InactiveColor: lipgloss.Color("#2a3a44"),
}

func TestTitledBox_ContainsTitle(t *testing.T) {
	box := TitledBox(testCfg, "MEMBERS", "Alice\nBob", "", 40, 0, false)
	if !strings.Contains(box, "MEMBERS") {
		t.Error("expected title in top border")
	}
}

func TestTitledBox_ContainsContent(t *testing.T) {
	box := TitledBox(testCfg, "TEST", "hello world", "", 40, 0, false)
	if !strings.Contains(box, "hello world") {
		t.Error("expected content inside box")
	}
}

func TestTitledBox_ContainsBottomLabel(t *testing.T) {
	box := TitledBox(testCfg, "TEST", "content", "5 total", 40, 0, false)
	if !strings.Contains(box, "5 total") {
		t.Error("expected bottom label in border")
	}
}

func TestTitledBox_NoBottomLabel(t *testing.T) {
	box := TitledBox(testCfg, "TEST", "content", "", 40, 0, false)
	lines := strings.Split(box, "\n")
	last := lines[len(lines)-1]
	if strings.Contains(last, "total") {
		t.Error("expected no label in bottom border")
	}
}

func TestTitledBox_MinHeight_PadsShortContent(t *testing.T) {
	short := TitledBox(testCfg, "TEST", "one", "", 30, 1, false)
	tall := TitledBox(testCfg, "TEST", "one", "", 30, 5, false)
	shortLines := strings.Count(short, "\n")
	tallLines := strings.Count(tall, "\n")
	if tallLines <= shortLines {
		t.Errorf("expected tall (%d lines) > short (%d lines)", tallLines, shortLines)
	}
}

func TestTitledBox_MinHeight_DoesNotShrink(t *testing.T) {
	content := "a\nb\nc\nd\ne"
	box := TitledBox(testCfg, "TEST", content, "", 30, 2, false)
	if !strings.Contains(box, "d") || !strings.Contains(box, "e") {
		t.Error("minHeight should not shrink content")
	}
}

func TestTitledBox_ActiveCallsWithoutPanic(t *testing.T) {
	active := TitledBox(testCfg, "TEST", "x", "", 30, 0, true)
	inactive := TitledBox(testCfg, "TEST", "x", "", 30, 0, false)
	if len(active) == 0 || len(inactive) == 0 {
		t.Error("expected non-empty output for both active states")
	}
}

func TestTitledBox_HasBorderChars(t *testing.T) {
	box := TitledBox(testCfg, "T", "x", "", 20, 0, false)
	if !strings.Contains(box, "╭") || !strings.Contains(box, "╯") {
		t.Error("expected rounded border characters")
	}
	if !strings.Contains(box, "│") {
		t.Error("expected side border characters")
	}
}

func TestTitledBox_BottomLabelPosition(t *testing.T) {
	box := TitledBox(testCfg, "TEST", "x", "3 total", 40, 0, false)
	lines := strings.Split(box, "\n")
	last := lines[len(lines)-1]
	// Label should be in the last 1/4, so there should be dashes before it
	idx := strings.Index(last, "3 total")
	if idx < 0 {
		t.Fatal("label not found")
	}
	// At least half the line should be dashes before the label
	if idx < 20 {
		t.Errorf("label at position %d, expected further right (last 1/4 of width 40)", idx)
	}
}

func TestContentHeight(t *testing.T) {
	if h := ContentHeight("a"); h != 1 {
		t.Errorf("single line: got %d, want 1", h)
	}
	if h := ContentHeight("a\nb\nc"); h != 3 {
		t.Errorf("three lines: got %d, want 3", h)
	}
	if h := ContentHeight(""); h != 1 {
		t.Errorf("empty string: got %d, want 1", h)
	}
}
