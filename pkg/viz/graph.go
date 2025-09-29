package viz

import (
	"fmt"
	"image/color"
	"math"
	"path/filepath"
	"strings"

	"github.com/fogleman/gg"
	"led-curve-tools/pkg/cue"
	"led-curve-tools/pkg/fade"
)

type GraphConfig struct {
	Width       int
	Height      int
	Title       string
	XLabel      string
	YLabel      string
	ShowGrid    bool
	ShowLegend  bool
	LogScale    bool
	MarginLeft  float64
	MarginRight float64
	MarginTop   float64
	MarginBottom float64
	LineWidth   float64
}

func DefaultGraphConfig() *GraphConfig {
	return &GraphConfig{
		Width:        1200,
		Height:       800,
		Title:        "LED Fade Curves",
		XLabel:       "Time (ms)",
		YLabel:       "Duty Cycle (%)",
		ShowGrid:     true,
		ShowLegend:   true,
		LogScale:     true,
		MarginLeft:   80,
		MarginRight:  180,  // Increased for legend space
		MarginTop:    60,
		MarginBottom: 80,
		LineWidth:    2.0,
	}
}

func CreateFadeGraph(fades []*fade.Fade, config *GraphConfig, outputPath string) error {
	if len(fades) == 0 {
		return fmt.Errorf("no fades provided")
	}

	dc := gg.NewContext(config.Width, config.Height)

	dc.SetColor(ColorBackground)
	dc.Clear()

	plotArea := calculatePlotArea(config)

	maxTime, maxDuty := calculateMaxValues(fades)

	if config.ShowGrid {
		drawGrid(dc, plotArea, maxTime, maxDuty, config)
	}

	drawAxes(dc, plotArea, maxTime, maxDuty, config)

	for i, f := range fades {
		color := GetColor(i)
		drawFade(dc, f, plotArea, maxTime, maxDuty, color, config.LineWidth, config.LogScale)
	}

	if config.ShowLegend && len(fades) > 1 {
		drawLegend(dc, fades, config)
	}

	drawTitle(dc, config)

	return saveGraph(dc, outputPath)
}

func CreateCueVisualization(c *cue.Cue, config *GraphConfig, outputPath string) error {
	dc := gg.NewContext(config.Width, config.Height)

	dc.SetColor(ColorBackground)
	dc.Clear()

	plotArea := calculatePlotArea(config)

	drawCueTimeline(dc, c, plotArea)
	drawTitle(dc, config)

	return saveGraph(dc, outputPath)
}

type PlotArea struct {
	X, Y, Width, Height float64
}

func calculatePlotArea(config *GraphConfig) PlotArea {
	return PlotArea{
		X:      config.MarginLeft,
		Y:      config.MarginTop,
		Width:  float64(config.Width) - config.MarginLeft - config.MarginRight,
		Height: float64(config.Height) - config.MarginTop - config.MarginBottom,
	}
}

func calculateMaxValues(fades []*fade.Fade) (float64, float64) {
	maxTime := 0.0
	maxDuty := 1.0

	for _, f := range fades {
		if len(f.Points) > 0 {
			lastPoint := f.Points[len(f.Points)-1]
			if lastPoint.Time > maxTime {
				maxTime = lastPoint.Time
			}
		}
		for _, point := range f.Points {
			if point.Duty > maxDuty {
				maxDuty = point.Duty
			}
		}
	}

	maxTime = math.Ceil(maxTime/1000) * 1000
	maxDuty = math.Min(maxDuty*1.1, 1.0)

	return maxTime, maxDuty
}

func logScaleTransform(value, max float64) float64 {
	if value <= 0 {
		return 0
	}
	// Use log10 scale with a minimum floor to handle very small values
	minValue := 0.001
	if value < minValue {
		value = minValue
	}
	if max < minValue {
		max = 1.0
	}
	return math.Log10(value/minValue) / math.Log10(max/minValue)
}

func getYPosition(duty, maxDuty float64, plotArea PlotArea, logScale bool) float64 {
	var ratio float64
	if logScale {
		ratio = logScaleTransform(duty, maxDuty)
	} else {
		ratio = duty / maxDuty
	}
	return plotArea.Y + plotArea.Height - ratio*plotArea.Height
}

func drawGrid(dc *gg.Context, plotArea PlotArea, maxTime, maxDuty float64, config *GraphConfig) {
	dc.SetColor(ColorGrid)
	dc.SetLineWidth(1.0)

	timeSteps := int(maxTime / 1000)
	if timeSteps > 10 {
		timeSteps = (timeSteps + 4) / 5 * 5
	}

	for i := 0; i <= timeSteps; i++ {
		time := float64(i) * maxTime / float64(timeSteps)
		x := plotArea.X + (time/maxTime)*plotArea.Width
		dc.DrawLine(x, plotArea.Y, x, plotArea.Y+plotArea.Height)
		dc.Stroke()
	}

	if config.LogScale {
		// Major grid lines
		dutyValues := []float64{0.001, 0.002, 0.005, 0.01, 0.02, 0.05, 0.1, 0.2, 0.5, 1.0}
		for _, duty := range dutyValues {
			if duty <= maxDuty {
				y := getYPosition(duty, maxDuty, plotArea, config.LogScale)
				dc.DrawLine(plotArea.X, y, plotArea.X+plotArea.Width, y)
				dc.Stroke()
			}
		}

		// Minor grid lines (intermediate values)
		minorDutyValues := []float64{0.003, 0.004, 0.006, 0.007, 0.008, 0.009, 0.015, 0.03, 0.04, 0.06, 0.07, 0.08, 0.09, 0.15, 0.3, 0.4, 0.6, 0.7, 0.8, 0.9}
		dc.SetColor(ColorGrid)
		dc.SetLineWidth(0.5)
		for _, duty := range minorDutyValues {
			if duty <= maxDuty {
				y := getYPosition(duty, maxDuty, plotArea, config.LogScale)
				dc.DrawLine(plotArea.X, y, plotArea.X+plotArea.Width, y)
				dc.Stroke()
			}
		}
		dc.SetLineWidth(1.0)
	} else {
		dutySteps := 10
		for i := 0; i <= dutySteps; i++ {
			duty := float64(i) * maxDuty / float64(dutySteps)
			y := getYPosition(duty, maxDuty, plotArea, config.LogScale)
			dc.DrawLine(plotArea.X, y, plotArea.X+plotArea.Width, y)
			dc.Stroke()
		}
	}
}

func drawAxes(dc *gg.Context, plotArea PlotArea, maxTime, maxDuty float64, config *GraphConfig) {
	dc.SetColor(ColorAxis)
	dc.SetLineWidth(2.0)

	dc.DrawLine(plotArea.X, plotArea.Y, plotArea.X, plotArea.Y+plotArea.Height)
	dc.DrawLine(plotArea.X, plotArea.Y+plotArea.Height, plotArea.X+plotArea.Width, plotArea.Y+plotArea.Height)
	dc.Stroke()

	dc.SetColor(ColorText)
	dc.LoadFontFace("/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf", 12)

	// Draw minor ticks every 20ms
	dc.SetLineWidth(1.0)
	dc.SetColor(ColorGrid)
	for timeMs := 0.0; timeMs <= maxTime; timeMs += 20.0 {
		x := plotArea.X + (timeMs/maxTime)*plotArea.Width
		dc.DrawLine(x, plotArea.Y+plotArea.Height-5, x, plotArea.Y+plotArea.Height)
	}
	dc.Stroke()

	// Draw major ticks every 100ms with labels
	dc.SetLineWidth(2.0)
	dc.SetColor(ColorAxis)
	for timeMs := 0.0; timeMs <= maxTime; timeMs += 100.0 {
		x := plotArea.X + (timeMs/maxTime)*plotArea.Width
		dc.DrawLine(x, plotArea.Y+plotArea.Height-10, x, plotArea.Y+plotArea.Height)
	}
	dc.Stroke()

	// Add major tick labels
	dc.SetColor(ColorText)
	for timeMs := 0.0; timeMs <= maxTime; timeMs += 100.0 {
		x := plotArea.X + (timeMs/maxTime)*plotArea.Width
		label := fmt.Sprintf("%.0f", timeMs)
		dc.DrawStringAnchored(label, x, plotArea.Y+plotArea.Height+20, 0.5, 0)
	}

	// Y-axis ticks and labels
	dc.SetColor(ColorAxis)
	dc.SetLineWidth(1.0)

	if config.LogScale {
		// Major tick marks with labels
		dutyValues := []float64{0.001, 0.002, 0.005, 0.01, 0.02, 0.05, 0.1, 0.2, 0.5, 1.0}
		for _, duty := range dutyValues {
			if duty <= maxDuty {
				y := getYPosition(duty, maxDuty, plotArea, config.LogScale)
				// Major tick mark
				dc.DrawLine(plotArea.X-8, y, plotArea.X, y)
				dc.Stroke()
			}
		}

		// Minor tick marks (no labels, every 10% increment above 20%)
		minorDutyValues := []float64{0.3, 0.4, 0.6, 0.7, 0.8, 0.9}
		for _, duty := range minorDutyValues {
			if duty <= maxDuty {
				y := getYPosition(duty, maxDuty, plotArea, config.LogScale)
				// Minor tick mark
				dc.DrawLine(plotArea.X-4, y, plotArea.X, y)
				dc.Stroke()
			}
		}

		// Add labels for major ticks
		dc.SetColor(ColorText)
		for _, duty := range dutyValues {
			if duty <= maxDuty {
				y := getYPosition(duty, maxDuty, plotArea, config.LogScale)
				var label string
				if duty < 0.01 {
					label = fmt.Sprintf("%.1f", duty*100)
				} else {
					label = fmt.Sprintf("%.0f", duty*100)
				}
				dc.DrawStringAnchored(label, plotArea.X-15, y, 1.0, 0.5)
			}
		}
	} else {
		// Linear scale - more detailed ticks above 20%
		dutySteps := 10
		for i := 0; i <= dutySteps; i++ {
			duty := float64(i) * maxDuty / float64(dutySteps)
			y := getYPosition(duty, maxDuty, plotArea, config.LogScale)
			// Major tick mark
			dc.DrawLine(plotArea.X-8, y, plotArea.X, y)
			dc.Stroke()
		}

		// Add more minor ticks above 20%
		for pct := 30; pct <= 90; pct += 10 {
			duty := float64(pct) / 100.0
			if duty <= maxDuty {
				y := getYPosition(duty, maxDuty, plotArea, config.LogScale)
				// Minor tick mark
				dc.DrawLine(plotArea.X-4, y, plotArea.X, y)
				dc.Stroke()
			}
		}

		// Add labels for major ticks
		dc.SetColor(ColorText)
		for i := 0; i <= dutySteps; i++ {
			duty := float64(i) * maxDuty / float64(dutySteps)
			y := getYPosition(duty, maxDuty, plotArea, config.LogScale)
			label := fmt.Sprintf("%.0f", duty*100)
			dc.DrawStringAnchored(label, plotArea.X-15, y, 1.0, 0.5)
		}
	}

	dc.LoadFontFace("/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf", 14)
	dc.DrawStringAnchored(config.XLabel, plotArea.X+plotArea.Width/2, plotArea.Y+plotArea.Height+50, 0.5, 0)

	dc.Push()
	dc.RotateAbout(-math.Pi/2, plotArea.X-50, plotArea.Y+plotArea.Height/2)
	dc.DrawStringAnchored(config.YLabel, plotArea.X-50, plotArea.Y+plotArea.Height/2, 0.5, 0.5)
	dc.Pop()
}

func drawFade(dc *gg.Context, f *fade.Fade, plotArea PlotArea, maxTime, maxDuty float64, color color.RGBA, lineWidth float64, logScale bool) {
	if len(f.Points) < 2 {
		return
	}

	dc.SetColor(color)
	dc.SetLineWidth(lineWidth)

	for i := 0; i < len(f.Points)-1; i++ {
		p1, p2 := f.Points[i], f.Points[i+1]

		x1 := plotArea.X + (p1.Time/maxTime)*plotArea.Width
		y1 := getYPosition(p1.Duty, maxDuty, plotArea, logScale)
		x2 := plotArea.X + (p2.Time/maxTime)*plotArea.Width
		y2 := getYPosition(p2.Duty, maxDuty, plotArea, logScale)

		dc.DrawLine(x1, y1, x2, y2)
	}
	dc.Stroke()

	dc.SetLineWidth(1.0)
	for i, point := range f.Points {
		x := plotArea.X + (point.Time/maxTime)*plotArea.Width
		y := getYPosition(point.Duty, maxDuty, plotArea, logScale)

		if i == 0 {
			dc.DrawCircle(x, y, 4)
			dc.Fill()
		} else if i == len(f.Points)-1 {
			dc.DrawRectangle(x-4, y-4, 8, 8)
			dc.Fill()
		} else {
			dc.DrawCircle(x, y, 2)
			dc.Fill()
		}
	}
}

func drawLegend(dc *gg.Context, fades []*fade.Fade, config *GraphConfig) {
	plotArea := calculatePlotArea(config)
	legendX := plotArea.X + plotArea.Width + 20  // 20px from plot area edge
	legendY := config.MarginTop + 20

	legendWidth := 150.0
	legendHeight := float64(len(fades))*25 + 20

	dc.SetColor(ColorBackground)
	dc.DrawRectangle(legendX-10, legendY-10, legendWidth, legendHeight)
	dc.Fill()

	dc.SetColor(ColorAxis)
	dc.SetLineWidth(1.0)
	dc.DrawRectangle(legendX-10, legendY-10, legendWidth, legendHeight)
	dc.Stroke()

	dc.LoadFontFace("/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf", 12)

	for i, f := range fades {
		y := legendY + float64(i)*25

		color := GetColor(i)
		dc.SetColor(color)
		dc.SetLineWidth(3.0)
		dc.DrawLine(legendX, y, legendX+20, y)
		dc.Stroke()

		dc.SetColor(ColorText)
		name := f.Name
		if name == "" {
			name = f.Filename
		}
		dc.DrawString(name, legendX+30, y+5)
	}
}

func drawTitle(dc *gg.Context, config *GraphConfig) {
	if config.Title == "" {
		return
	}

	dc.SetColor(ColorText)
	dc.LoadFontFace("/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf", 18)
	dc.DrawStringAnchored(config.Title, float64(config.Width)/2, 30, 0.5, 0.5)
}

func drawCueTimeline(dc *gg.Context, c *cue.Cue, plotArea PlotArea) {
	dc.SetColor(ColorText)
	dc.LoadFontFace("/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf", 12)

	y := plotArea.Y + 50
	spacing := 30.0

	dc.DrawString(fmt.Sprintf("Cue: %s (%d actions)", c.Name, len(c.Actions)), plotArea.X, y)
	y += spacing * 2

	for i, action := range c.Actions {
		text := fmt.Sprintf("Action %d: LED %d, %s", i, action.LEDIndex, action.Type)
		if action.Type == cue.TypeDuty {
			text += fmt.Sprintf(" %.3f", action.Value)
		} else {
			text += fmt.Sprintf(" %d", int(action.Value))
		}

		color := GetColor(action.LEDIndex)
		dc.SetColor(color)
		dc.DrawCircle(plotArea.X, y-5, 5)
		dc.Fill()

		dc.SetColor(ColorText)
		dc.DrawString(text, plotArea.X+20, y)
		y += spacing
	}
}

func saveGraph(dc *gg.Context, outputPath string) error {
	ext := strings.ToLower(filepath.Ext(outputPath))

	switch ext {
	case ".png", "":
		if ext == "" {
			outputPath += ".png"
		}
		return dc.SavePNG(outputPath)
	default:
		return fmt.Errorf("unsupported output format: %s", ext)
	}
}