package widgets

import "testing"

func TestScrollWindow_FitsAll(t *testing.T) {
	start, end := ScrollWindow(3, 0, 10)
	if start != 0 || end != 3 {
		t.Errorf("got (%d,%d), want (0,3)", start, end)
	}
}

func TestScrollWindow_CursorAtStart(t *testing.T) {
	start, end := ScrollWindow(20, 0, 5)
	if start != 0 || end != 5 {
		t.Errorf("got (%d,%d), want (0,5)", start, end)
	}
}

func TestScrollWindow_CursorAtEnd(t *testing.T) {
	start, end := ScrollWindow(20, 19, 5)
	if start != 15 || end != 20 {
		t.Errorf("got (%d,%d), want (15,20)", start, end)
	}
}

func TestScrollWindow_CursorMiddle(t *testing.T) {
	start, end := ScrollWindow(20, 10, 5)
	if start != 8 || end != 13 {
		t.Errorf("got (%d,%d), want (8,13)", start, end)
	}
}

func TestScrollWindow_ExactFit(t *testing.T) {
	start, end := ScrollWindow(5, 2, 5)
	if start != 0 || end != 5 {
		t.Errorf("got (%d,%d), want (0,5)", start, end)
	}
}

func TestScrollWindow_Empty(t *testing.T) {
	start, end := ScrollWindow(0, 0, 5)
	if start != 0 || end != 0 {
		t.Errorf("got (%d,%d), want (0,0)", start, end)
	}
}
