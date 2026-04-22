package widgets

import (
	"fmt"
	"math"
	"strings"
)

// RadarChart renders a text-based radar/spider chart using Unicode characters.
// Labels and score values are placed directly on the chart at each axis endpoint.
func RadarChart(labels []string, values []float64, maxValue int, radius int) string {
	n := len(labels)
	if n < 3 || len(values) != n || maxValue <= 0 {
		return ""
	}

	// Find longest label for padding
	maxLabelLen := 0
	for _, l := range labels {
		if len(l) > maxLabelLen {
			maxLabelLen = len(l)
		}
	}
	labelPad := maxLabelLen + 6 // label + " 4.0" + margin

	// Grid dimensions — wider than tall because terminal chars are ~2:1 aspect
	width := radius*4 + labelPad*2
	height := radius*2 + 4
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

	// Convert polar to grid coordinates
	toGrid := func(index int, value float64) (int, int) {
		angle := 2*math.Pi*float64(index)/float64(n) - math.Pi/2
		r := (value / float64(maxValue)) * float64(radius)
		x := cx + int(math.Round(r*math.Cos(angle)*2))
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

	writeString := func(x, y int, s string) {
		for i, ch := range s {
			px := x + i
			if px >= 0 && px < width && y >= 0 && y < height {
				grid[y][px] = ch
			}
		}
	}

	// Draw reference ring at max level
	for i := 0; i < n; i++ {
		x, y := toGrid(i, float64(maxValue))
		setGrid(x, y, '·')
	}

	// Draw center
	setGrid(cx, cy, '+')

	// Draw axis dots from center to outer ring
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

	// Plot data points
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

	// Connect adjacent points
	for i := 0; i < n; i++ {
		next := (i + 1) % n
		drawLine(grid, pts[i].x, pts[i].y, pts[next].x, pts[next].y, width, height)
	}

	// Place labels at the outer edge of each axis
	for i, label := range labels {
		angle := 2*math.Pi*float64(i)/float64(n) - math.Pi/2
		lx, ly := toGrid(i, float64(maxValue))

		score := "—"
		if values[i] > 0 {
			score = fmt.Sprintf("%.1f", values[i])
		}

		text := label + " " + score

		// Position label based on which side of the chart it's on
		cosA := math.Cos(angle)
		if cosA > 0.3 {
			// Right side: place label to the right of the point
			writeString(lx+2, ly, text)
		} else if cosA < -0.3 {
			// Left side: place label to the left of the point
			startX := lx - len(text) - 1
			if startX < 0 {
				startX = 0
			}
			writeString(startX, ly, text)
		} else {
			// Top or bottom: center the label
			startX := lx - len(text)/2
			if startX < 0 {
				startX = 0
			}
			sinA := math.Sin(angle)
			offset := 1
			if sinA < 0 {
				offset = -1
			}
			writeString(startX, ly+offset, text)
		}
	}

	// Build output
	var b strings.Builder
	for _, row := range grid {
		line := strings.TrimRight(string(row), " ")
		b.WriteString(line)
		b.WriteString("\n")
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
