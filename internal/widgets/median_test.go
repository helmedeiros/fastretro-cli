package widgets

import "testing"

func TestMedianInt_OddCount(t *testing.T) {
	if got := MedianInt([]int{1, 3, 5}); got != 3 {
		t.Errorf("got %.1f, want 3", got)
	}
}

func TestMedianInt_EvenCount(t *testing.T) {
	if got := MedianInt([]int{2, 4}); got != 3 {
		t.Errorf("got %.1f, want 3", got)
	}
}

func TestMedianInt_Single(t *testing.T) {
	if got := MedianInt([]int{7}); got != 7 {
		t.Errorf("got %.1f, want 7", got)
	}
}

func TestMedianInt_Empty(t *testing.T) {
	if got := MedianInt(nil); got != 0 {
		t.Errorf("got %.1f, want 0", got)
	}
}

func TestMedianInt_Unsorted(t *testing.T) {
	if got := MedianInt([]int{5, 1, 3}); got != 3 {
		t.Errorf("got %.1f, want 3", got)
	}
}

func TestMedianInt_DoesNotMutateInput(t *testing.T) {
	input := []int{5, 1, 3}
	MedianInt(input)
	if input[0] != 5 || input[1] != 1 || input[2] != 3 {
		t.Errorf("input was mutated: %v", input)
	}
}
