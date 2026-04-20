package widgets

import (
	"fmt"
	"math"
	"strings"
)

// RadarChart renders a text-based radar/spider chart using Unicode characters.
// labels and values must have the same length. radius controls the chart size in characters.
func RadarChart(labels []string, values []float64, maxValue int, radius int) string {
	n := len(labels)
	if n < 3 || len(values) != n || maxValue <= 0 {
		return ""
	}

	// Grid dimensions — wider than tall because terminal chars are ~2:1 aspect
	width := radius*4 + 2
	height := radius*2 + 2
	cx := width / 2
	cy := height / 2

	// Initialize grid with spaces
	grid := make([][]rune, height)
	for y := range grid {
		grid[y] = make([]rune, width)
		for x := range grid[y] {
			grid[y][x] = ' '
		}
	}

	// Helper to convert polar to grid coordinates
	toGrid := func(index int, value float64) (int, int) {
		angle := 2*math.Pi*float64(index)/float64(n) - math.Pi/2
		r := (value / float64(maxValue)) * float64(radius)
		x := cx + int(math.Round(r*math.Cos(angle)*2)) // *2 for aspect ratio
		y := cy + int(math.Round(r*math.Sin(angle)))
		if x < 0 {
			x = 0
		}
		if x >= width {
			x = width - 1
		}
		if y < 0 {
			y = 0
		}
		if y >= height {
			y = height - 1
		}
		return x, y
	}

	setGrid := func(x, y int, ch rune) {
		if x >= 0 && x < width && y >= 0 && y < height {
			grid[y][x] = ch
		}
	}

	// Draw reference ring at max level using dots
	for i := 0; i < n; i++ {
		x, y := toGrid(i, float64(maxValue))
		setGrid(x, y, '·')
	}

	// Draw center dot
	setGrid(cx, cy, '+')

	// Draw axis lines from center to outer ring
	for i := 0; i < n; i++ {
		steps := radius
		for s := 1; s <= steps; s++ {
			v := float64(maxValue) * float64(s) / float64(steps)
			x, y := toGrid(i, v)
			if grid[y][x] == ' ' {
				grid[y][x] = '·'
			}
		}
	}

	// Plot data points and connect them
	type point struct{ x, y int }
	pts := make([]point, n)
	for i, v := range values {
		if v > float64(maxValue) {
			v = float64(maxValue)
		}
		x, y := toGrid(i, v)
		pts[i] = point{x, y}
		setGrid(x, y, '●')
	}

	// Connect adjacent points with line characters
	for i := 0; i < n; i++ {
		next := (i + 1) % n
		drawLine(grid, pts[i].x, pts[i].y, pts[next].x, pts[next].y, width, height)
	}

	// Build output string from grid
	var b strings.Builder
	for _, row := range grid {
		line := strings.TrimRight(string(row), " ")
		b.WriteString(line)
		b.WriteString("\n")
	}

	// Add labels below the chart
	b.WriteString("\n")
	for i, label := range labels {
		score := "—"
		if values[i] > 0 {
			score = fmt.Sprintf("%.1f", values[i])
		}
		b.WriteString(fmt.Sprintf("  %-22s %s\n", label, score))
	}

	return b.String()
}

// drawLine draws a line between two points on the grid using basic characters.
func drawLine(grid [][]rune, x0, y0, x1, y1, w, h int) {
	dx := x1 - x0
	dy := y1 - y0
	steps := int(math.Max(math.Abs(float64(dx)), math.Abs(float64(dy))))
	if steps == 0 {
		return
	}
	for s := 1; s < steps; s++ {
		x := x0 + dx*s/steps
		y := y0 + dy*s/steps
		if x >= 0 && x < w && y >= 0 && y < h && grid[y][x] == ' ' {
			// Pick character based on line direction
			if dy == 0 {
				grid[y][x] = '─'
			} else if dx == 0 {
				grid[y][x] = '│'
			} else if (dx > 0 && dy > 0) || (dx < 0 && dy < 0) {
				grid[y][x] = '╲'
			} else {
				grid[y][x] = '╱'
			}
		}
	}
}
