package widgets

import (
	"testing"
)

func TestWrapText_ShortText(t *testing.T) {
	lines := WrapText("hello", 20)
	if len(lines) != 1 || lines[0] != "hello" {
		t.Errorf("got %v", lines)
	}
}

func TestWrapText_ExactFit(t *testing.T) {
	lines := WrapText("12345", 5)
	if len(lines) != 1 || lines[0] != "12345" {
		t.Errorf("got %v", lines)
	}
}

func TestWrapText_WordBreak(t *testing.T) {
	lines := WrapText("hello world foo", 11)
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d: %v", len(lines), lines)
	}
	if lines[0] != "hello world" {
		t.Errorf("line 0: got %q", lines[0])
	}
	if lines[1] != "foo" {
		t.Errorf("line 1: got %q", lines[1])
	}
}

func TestWrapText_LongWord(t *testing.T) {
	lines := WrapText("abcdefghij", 5)
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d: %v", len(lines), lines)
	}
	if lines[0] != "abcde" {
		t.Errorf("line 0: got %q", lines[0])
	}
}

func TestWrapText_Empty(t *testing.T) {
	lines := WrapText("", 10)
	if len(lines) != 1 || lines[0] != "" {
		t.Errorf("got %v", lines)
	}
}

func TestWrapText_MultipleWraps(t *testing.T) {
	lines := WrapText("the quick brown fox jumps over the lazy dog", 15)
	if len(lines) < 3 {
		t.Errorf("expected at least 3 lines, got %d: %v", len(lines), lines)
	}
	for _, l := range lines {
		if len(l) > 15 {
			t.Errorf("line exceeds maxWidth: %q (%d chars)", l, len(l))
		}
	}
}
