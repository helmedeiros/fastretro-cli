package widgets

import (
	"strings"
	"testing"
)

func TestBarChart_BasicOutput(t *testing.T) {
	result := BarChart([]string{"A", "B"}, []float64{3, 5}, 5, 10)
	if result == "" {
		t.Fatal("expected non-empty output")
	}
}

func TestBarChart_ContainsLabels(t *testing.T) {
	result := BarChart([]string{"Ownership", "Value"}, []float64{4, 3}, 5, 10)
	if !strings.Contains(result, "Ownership") || !strings.Contains(result, "Value") {
		t.Error("expected labels in output")
	}
}

func TestBarChart_ContainsScores(t *testing.T) {
	result := BarChart([]string{"A"}, []float64{4.5}, 5, 10)
	if !strings.Contains(result, "4.5") {
		t.Error("expected score value")
	}
}

func TestBarChart_ContainsFilledAndEmpty(t *testing.T) {
	result := BarChart([]string{"A"}, []float64{3}, 5, 10)
	if !strings.Contains(result, "█") {
		t.Error("expected filled bar chars")
	}
	if !strings.Contains(result, "░") {
		t.Error("expected empty bar chars")
	}
}

func TestBarChart_FullBar(t *testing.T) {
	result := BarChart([]string{"A"}, []float64{5}, 5, 10)
	if strings.Contains(result, "░") {
		t.Error("full score should have no empty segments")
	}
}

func TestBarChart_ZeroValue(t *testing.T) {
	result := BarChart([]string{"A"}, []float64{0}, 5, 10)
	if !strings.Contains(result, "—") {
		t.Error("zero score should show dash")
	}
}

func TestBarChart_Empty(t *testing.T) {
	result := BarChart([]string{}, []float64{}, 5, 10)
	if result != "" {
		t.Error("expected empty for no data")
	}
}

func TestBarChart_Mismatched(t *testing.T) {
	result := BarChart([]string{"A"}, []float64{1, 2}, 5, 10)
	if result != "" {
		t.Error("expected empty for mismatched lengths")
	}
}
