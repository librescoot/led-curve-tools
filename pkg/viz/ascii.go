package viz

import (
	"fmt"
	"strings"

	"led-curve-tools/pkg/cue"
	"led-curve-tools/pkg/fade"
)

const (
	DefaultWidth  = 80
	DefaultHeight = 20
)

func RenderFadeASCII(f *fade.Fade, width, height int) string {
	if len(f.Points) == 0 {
		return "No data points"
	}

	if width <= 0 {
		width = DefaultWidth
	}
	if height <= 0 {
		height = DefaultHeight
	}

	grid := make([][]rune, height)
	for i := range grid {
		grid[i] = make([]rune, width)
		for j := range grid[i] {
			grid[i][j] = ' '
		}
	}

	maxTime := f.Points[len(f.Points)-1].Time
	maxDuty := 1.0
	for _, point := range f.Points {
		if point.Duty > maxDuty {
			maxDuty = point.Duty
		}
	}

	for i := 0; i < len(f.Points)-1; i++ {
		p1, p2 := f.Points[i], f.Points[i+1]

		x1 := int((p1.Time / maxTime) * float64(width-1))
		y1 := height - 1 - int((p1.Duty/maxDuty)*float64(height-1))
		x2 := int((p2.Time / maxTime) * float64(width-1))
		y2 := height - 1 - int((p2.Duty/maxDuty)*float64(height-1))

		drawLine(grid, x1, y1, x2, y2, '█')
	}

	for i, point := range f.Points {
		x := int((point.Time / maxTime) * float64(width-1))
		y := height - 1 - int((point.Duty/maxDuty)*float64(height-1))
		if x >= 0 && x < width && y >= 0 && y < height {
			if i == 0 {
				grid[y][x] = '●'
			} else if i == len(f.Points)-1 {
				grid[y][x] = '■'
			} else {
				grid[y][x] = '●'
			}
		}
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Fade: %s\n", f.Name))
	result.WriteString(fmt.Sprintf("Duration: %.1fms, Max Duty: %.1f%%\n", maxTime, maxDuty*100))
	result.WriteString(strings.Repeat("─", width) + "\n")

	for _, row := range grid {
		result.WriteString(string(row) + "\n")
	}

	result.WriteString(strings.Repeat("─", width) + "\n")
	result.WriteString(fmt.Sprintf("0ms%s%.0fms\n",
		strings.Repeat(" ", width-10), maxTime))

	return result.String()
}

func RenderMultipleFadesASCII(fades []*fade.Fade, width, height int) string {
	if len(fades) == 0 {
		return "No fades provided"
	}

	if width <= 0 {
		width = DefaultWidth
	}
	if height <= 0 {
		height = DefaultHeight
	}

	grid := make([][]rune, height)
	for i := range grid {
		grid[i] = make([]rune, width)
		for j := range grid[i] {
			grid[i][j] = ' '
		}
	}

	maxTime := 0.0
	maxDuty := 1.0
	for _, f := range fades {
		if len(f.Points) > 0 {
			lastTime := f.Points[len(f.Points)-1].Time
			if lastTime > maxTime {
				maxTime = lastTime
			}
		}
		for _, point := range f.Points {
			if point.Duty > maxDuty {
				maxDuty = point.Duty
			}
		}
	}

	symbols := []rune{'█', '▓', '▒', '░', '●', '○', '▪', '▫', '◆', '◇'}
	endSymbols := []rune{'■', '▣', '▦', '▩', '◆', '◇', '◾', '◽', '♦', '♢'}

	for fadeIdx, f := range fades {
		symbol := symbols[fadeIdx%len(symbols)]
		endSymbol := endSymbols[fadeIdx%len(endSymbols)]

		for i := 0; i < len(f.Points)-1; i++ {
			p1, p2 := f.Points[i], f.Points[i+1]

			x1 := int((p1.Time / maxTime) * float64(width-1))
			y1 := height - 1 - int((p1.Duty/maxDuty)*float64(height-1))
			x2 := int((p2.Time / maxTime) * float64(width-1))
			y2 := height - 1 - int((p2.Duty/maxDuty)*float64(height-1))

			drawLine(grid, x1, y1, x2, y2, symbol)
		}

		for i, point := range f.Points {
			x := int((point.Time / maxTime) * float64(width-1))
			y := height - 1 - int((point.Duty/maxDuty)*float64(height-1))
			if x >= 0 && x < width && y >= 0 && y < height {
				if i == len(f.Points)-1 {
					grid[y][x] = endSymbol
				}
			}
		}
	}

	var result strings.Builder
	result.WriteString("Multi-Fade Comparison\n")
	result.WriteString(fmt.Sprintf("Duration: %.1fms, Max Duty: %.1f%%\n", maxTime, maxDuty*100))

	for i, f := range fades {
		symbol := symbols[i%len(symbols)]
		name := f.Name
		if name == "" {
			name = f.Filename
		}
		result.WriteString(fmt.Sprintf("%c %s\n", symbol, name))
	}

	result.WriteString(strings.Repeat("─", width) + "\n")

	for _, row := range grid {
		result.WriteString(string(row) + "\n")
	}

	result.WriteString(strings.Repeat("─", width) + "\n")
	result.WriteString(fmt.Sprintf("0ms%s%.0fms\n",
		strings.Repeat(" ", width-10), maxTime))

	return result.String()
}

func RenderCueASCII(c *cue.Cue) string {
	var result strings.Builder

	result.WriteString(fmt.Sprintf("Cue: %s\n", c.Name))
	result.WriteString(fmt.Sprintf("Actions: %d\n", len(c.Actions)))
	result.WriteString(strings.Repeat("─", 50) + "\n")

	for i, action := range c.Actions {
		result.WriteString(fmt.Sprintf("%2d: LED %d │ %-8s │ ",
			i, action.LEDIndex, action.Type))

		if action.Type == cue.TypeDuty {
			result.WriteString(fmt.Sprintf("%.3f (%.1f%%)", action.Value, action.Value*100))
		} else {
			result.WriteString(fmt.Sprintf("%d", int(action.Value)))
		}
		result.WriteString("\n")
	}

	result.WriteString(strings.Repeat("─", 50) + "\n")
	return result.String()
}

func drawLine(grid [][]rune, x0, y0, x1, y1 int, char rune) {
	height := len(grid)
	width := len(grid[0])

	dx := abs(x1 - x0)
	dy := abs(y1 - y0)
	sx := 1
	if x0 > x1 {
		sx = -1
	}
	sy := 1
	if y0 > y1 {
		sy = -1
	}
	err := dx - dy

	x, y := x0, y0
	for {
		if x >= 0 && x < width && y >= 0 && y < height {
			grid[y][x] = char
		}

		if x == x1 && y == y1 {
			break
		}

		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x += sx
		}
		if e2 < dx {
			err += dx
			y += sy
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}