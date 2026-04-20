package widgets

import (
	"strings"
	"testing"
)

func TestRadarChart_BasicOutput(t *testing.T) {
	labels := []string{"A", "B", "C", "D", "E"}
	values := []float64{3, 4, 5, 2, 4}
	result := RadarChart(labels, values, 5, 8)
	if result == "" {
		t.Fatal("expected non-empty output")
	}
}

func TestRadarChart_ContainsLabels(t *testing.T) {
	labels := []string{"Ownership", "Value", "Fun"}
	values := []float64{4, 3, 5}
	result := RadarChart(labels, values, 5, 8)
	for _, l := range labels {
		if !strings.Contains(result, l) {
			t.Errorf("expected label %q in output", l)
		}
	}
}

func TestRadarChart_ContainsScoreValues(t *testing.T) {
	labels := []string{"A", "B", "C"}
	values := []float64{4, 3, 5}
	result := RadarChart(labels, values, 5, 8)
	if !strings.Contains(result, "4.0") || !strings.Contains(result, "3.0") || !strings.Contains(result, "5.0") {
		t.Errorf("expected score values in output")
	}
}

func TestRadarChart_ContainsDataPoints(t *testing.T) {
	labels := []string{"A", "B", "C"}
	values := []float64{4, 3, 5}
	result := RadarChart(labels, values, 5, 8)
	if !strings.Contains(result, "●") {
		t.Error("expected data point markers")
	}
}

func TestRadarChart_ContainsCenterDot(t *testing.T) {
	labels := []string{"A", "B", "C"}
	values := []float64{1, 1, 1}
	result := RadarChart(labels, values, 5, 8)
	if !strings.Contains(result, "+") {
		t.Error("expected center marker")
	}
}

func TestRadarChart_ZeroValue(t *testing.T) {
	labels := []string{"A", "B", "C"}
	values := []float64{0, 3, 0}
	result := RadarChart(labels, values, 5, 8)
	if strings.Contains(result, "0.0") {
		t.Error("zero scores should show dash, not 0.0")
	}
}

func TestRadarChart_TooFewAxes(t *testing.T) {
	result := RadarChart([]string{"A", "B"}, []float64{1, 2}, 5, 8)
	if result != "" {
		t.Error("expected empty for fewer than 3 axes")
	}
}

func TestRadarChart_MismatchedLengths(t *testing.T) {
	result := RadarChart([]string{"A", "B", "C"}, []float64{1, 2}, 5, 8)
	if result != "" {
		t.Error("expected empty for mismatched lengths")
	}
}
