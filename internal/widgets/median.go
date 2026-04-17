package widgets

// MedianInt computes the median of a slice of integers.
// Returns 0 for empty input. Does not modify the input slice.
func MedianInt(values []int) float64 {
	if len(values) == 0 {
		return 0
	}
	sorted := make([]int, len(values))
	copy(sorted, values)
	// Insertion sort (small slices in TUI context)
	for i := 1; i < len(sorted); i++ {
		for j := i; j > 0 && sorted[j] < sorted[j-1]; j-- {
			sorted[j], sorted[j-1] = sorted[j-1], sorted[j]
		}
	}
	mid := len(sorted) / 2
	if len(sorted)%2 == 0 {
		return float64(sorted[mid-1]+sorted[mid]) / 2.0
	}
	return float64(sorted[mid])
}
