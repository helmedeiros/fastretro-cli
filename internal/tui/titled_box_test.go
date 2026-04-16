package tui

import (
	"strings"
	"testing"
)

func TestTitledBox_ContainsTitle(t *testing.T) {
	box := titledBox("MEMBERS", "Alice\nBob", "", 40, 0, false)
	if !strings.Contains(box, "MEMBERS") {
		t.Error("expected title in top border")
	}
}

func TestTitledBox_ContainsContent(t *testing.T) {
	box := titledBox("TEST", "hello world", "", 40, 0, false)
	if !strings.Contains(box, "hello world") {
		t.Error("expected content inside box")
	}
}

func TestTitledBox_ContainsBottomLabel(t *testing.T) {
	box := titledBox("TEST", "content", "5 total", 40, 0, false)
	if !strings.Contains(box, "5 total") {
		t.Error("expected bottom label in border")
	}
}

func TestTitledBox_NoBottomLabel(t *testing.T) {
	box := titledBox("TEST", "content", "", 40, 0, false)
	lines := strings.Split(box, "\n")
	last := lines[len(lines)-1]
	if strings.Contains(last, "total") {
		t.Error("expected no label in bottom border")
	}
}

func TestTitledBox_MinHeight_PadsShortContent(t *testing.T) {
	short := titledBox("TEST", "one", "", 30, 1, false)
	tall := titledBox("TEST", "one", "", 30, 5, false)
	shortLines := strings.Count(short, "\n")
	tallLines := strings.Count(tall, "\n")
	if tallLines <= shortLines {
		t.Errorf("expected tall (%d lines) > short (%d lines)", tallLines, shortLines)
	}
}

func TestTitledBox_MinHeight_DoesNotShrink(t *testing.T) {
	content := "a\nb\nc\nd\ne"
	box := titledBox("TEST", content, "", 30, 2, false)
	if !strings.Contains(box, "d") || !strings.Contains(box, "e") {
		t.Error("minHeight should not shrink content")
	}
}

func TestTitledBox_ActiveCallsWithoutPanic(t *testing.T) {
	// Active and inactive should both render without panicking.
	// Color differences depend on terminal, so just verify both produce output.
	active := titledBox("TEST", "x", "", 30, 0, true)
	inactive := titledBox("TEST", "x", "", 30, 0, false)
	if len(active) == 0 || len(inactive) == 0 {
		t.Error("expected non-empty output for both active states")
	}
}

func TestTitledBox_HasBorderChars(t *testing.T) {
	box := titledBox("T", "x", "", 20, 0, false)
	if !strings.Contains(box, "╭") || !strings.Contains(box, "╯") {
		t.Error("expected rounded border characters")
	}
	if !strings.Contains(box, "│") {
		t.Error("expected side border characters")
	}
}

func TestContentHeight(t *testing.T) {
	if h := contentHeight("a"); h != 1 {
		t.Errorf("single line: got %d, want 1", h)
	}
	if h := contentHeight("a\nb\nc"); h != 3 {
		t.Errorf("three lines: got %d, want 3", h)
	}
	if h := contentHeight(""); h != 1 {
		t.Errorf("empty string: got %d, want 1", h)
	}
}
