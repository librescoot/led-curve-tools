package fade

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
)

const (
	PWMPeriod  = 12000
	SampleRate = 250
)

type CurveType string

const (
	CurveTypeLinear CurveType = "linear"
	CurveTypeBezier CurveType = "bezier"
)

type FadePoint struct {
	Time float64 `json:"time"`
	Duty float64 `json:"duty"`
}

type Fade struct {
	Name        string      `json:"name"`
	DisplayName string      `json:"displayName"`
	Points      []FadePoint `json:"points"`
	SampleRate  int         `json:"sampleRate"`
	CurveType   CurveType   `json:"curveType,omitempty"`
	FadeIndex   int         `json:"fadeIndex,omitempty"`
	Filename    string      `json:"filename,omitempty"`
	Samples     []uint16    `json:"-"`
	cachedBezierPoints []FadePoint // Cache for bezier curve points
}

func LoadBinary(filename string) (*Fade, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	sampleCount := stat.Size() / 2
	if stat.Size()%2 != 0 {
		return nil, fmt.Errorf("invalid fade file: odd number of bytes")
	}

	samples := make([]uint16, sampleCount)
	err = binary.Read(file, binary.LittleEndian, &samples)
	if err != nil {
		return nil, fmt.Errorf("failed to read samples: %w", err)
	}

	fade := &Fade{
		Name:       extractNameFromPath(filename),
		SampleRate: SampleRate,
		CurveType:  CurveTypeLinear, // Binary files don't store curve type, assume linear
		Samples:    samples,
		Filename:   strings.TrimSuffix(filepath.Base(filename), ".bin"),
	}

	fade.Points = fade.convertSamplesToPoints()
	fade.DisplayName = fade.Name

	return fade, nil
}

func (f *Fade) convertSamplesToPoints() []FadePoint {
	if len(f.Samples) == 0 {
		return []FadePoint{}
	}

	points := []FadePoint{}
	lastDuty := -1.0

	for i, sample := range f.Samples {
		duty := float64(sample) / float64(PWMPeriod)
		time := float64(i) * 1000.0 / float64(f.SampleRate)

		if i == 0 || i == len(f.Samples)-1 || abs(duty-lastDuty) > 0.001 {
			points = append(points, FadePoint{
				Time: time,
				Duty: duty,
			})
			lastDuty = duty
		}
	}

	return points
}

func (f *Fade) WriteBinary(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	if len(f.Samples) == 0 {
		f.generateSamplesFromPoints()
	}

	err = binary.Write(file, binary.LittleEndian, f.Samples)
	if err != nil {
		return fmt.Errorf("failed to write samples: %w", err)
	}

	return nil
}

func (f *Fade) generateSamplesFromPoints() {
	if len(f.Points) == 0 {
		f.Samples = []uint16{}
		return
	}

	duration := f.Points[len(f.Points)-1].Time
	sampleCount := int(duration*float64(f.SampleRate)/1000.0) + 1
	f.Samples = make([]uint16, sampleCount)

	for i := 0; i < sampleCount; i++ {
		time := float64(i) * 1000.0 / float64(f.SampleRate)
		duty := f.interpolateDutyAtTime(time)
		f.Samples[i] = uint16(duty * float64(PWMPeriod))
	}
}

func (f *Fade) interpolateDutyAtTime(time float64) float64 {
	if len(f.Points) == 0 {
		return 0.0
	}

	if len(f.Points) == 1 {
		return math.Max(0.0, math.Min(1.0, f.Points[0].Duty))
	}

	if time <= f.Points[0].Time {
		return math.Max(0.0, math.Min(1.0, f.Points[0].Duty))
	}

	if time >= f.Points[len(f.Points)-1].Time {
		return math.Max(0.0, math.Min(1.0, f.Points[len(f.Points)-1].Duty))
	}

	if f.CurveType == CurveTypeBezier && len(f.Points) >= 2 {
		// Use bezier curve interpolation
		bezierPoints := f.getBezierPoints()
		if len(bezierPoints) == 0 {
			return 0.0
		}
		if len(bezierPoints) == 1 {
			return math.Max(0.0, math.Min(1.0, bezierPoints[0].Duty))
		}

		// Find surrounding bezier points and interpolate
		i := 1
		for i < len(bezierPoints) && bezierPoints[i].Time < time {
			i++
		}

		if i >= len(bezierPoints) {
			return math.Max(0.0, math.Min(1.0, bezierPoints[len(bezierPoints)-1].Duty))
		}

		p0 := bezierPoints[i-1]
		p1 := bezierPoints[i]

		// Linear interpolation between bezier points
		if p1.Time == p0.Time {
			return math.Max(0.0, math.Min(1.0, p0.Duty))
		}
		t := (time - p0.Time) / (p1.Time - p0.Time)
		duty := p0.Duty + t*(p1.Duty-p0.Duty)
		return math.Max(0.0, math.Min(1.0, duty))
	} else {
		// Use linear interpolation
		for i := 0; i < len(f.Points)-1; i++ {
			p1, p2 := f.Points[i], f.Points[i+1]
			if time >= p1.Time && time <= p2.Time {
				if p2.Time == p1.Time {
					return math.Max(0.0, math.Min(1.0, p1.Duty))
				}
				ratio := (time - p1.Time) / (p2.Time - p1.Time)
				duty := p1.Duty + ratio*(p2.Duty-p1.Duty)
				return math.Max(0.0, math.Min(1.0, duty))
			}
		}
	}

	return 0.0
}

// binomialCoefficient calculates binomial coefficient for Bezier curves
func binomialCoefficient(n, k int) float64 {
	if k > n {
		return 0
	}
	if k == 0 || k == n {
		return 1
	}

	result := 1.0
	for i := 1; i <= k; i++ {
		result = result * float64(n-i+1) / float64(i)
	}
	return result
}

// calculateBezierCurve calculates Bezier curve points using control points
func (f *Fade) calculateBezierCurve(numPoints int) []FadePoint {
	if len(f.Points) < 2 {
		return append([]FadePoint{}, f.Points...)
	}

	// Use the existing fade points as control points
	controlPoints := f.Points
	n := len(controlPoints) - 1
	result := make([]FadePoint, 0, numPoints+1)

	// Get time bounds from original points
	startTime := controlPoints[0].Time
	endTime := controlPoints[len(controlPoints)-1].Time
	timeRange := endTime - startTime

	// Calculate curve points with proper time distribution
	for i := 0; i <= numPoints; i++ {
		t := float64(i) / float64(numPoints)

		// Calculate Bezier position in normalized space (0-1)
		var normalizedTime, duty float64

		for j := 0; j <= n; j++ {
			bernstein := binomialCoefficient(n, j) *
						math.Pow(1-t, float64(n-j)) *
						math.Pow(t, float64(j))

			// Normalize control point times to 0-1 range for Bezier calculation
			normalizedControlTime := (controlPoints[j].Time - startTime) / timeRange
			normalizedTime += bernstein * normalizedControlTime
			duty += bernstein * controlPoints[j].Duty
		}

		// Convert back to actual time scale
		actualTime := startTime + (normalizedTime * timeRange)
		clampedDuty := math.Max(0.0, math.Min(1.0, duty))
		result = append(result, FadePoint{Time: actualTime, Duty: clampedDuty})
	}

	// Sort points by time to ensure proper ordering
	for i := 0; i < len(result)-1; i++ {
		for j := i + 1; j < len(result); j++ {
			if result[i].Time > result[j].Time {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	return result
}

// getBezierPoints gets cached Bezier curve points or calculates them
func (f *Fade) getBezierPoints() []FadePoint {
	if f.cachedBezierPoints == nil || f.CurveType != CurveTypeBezier {
		if len(f.Points) == 0 {
			return []FadePoint{}
		}

		// Adaptive resolution based on duration
		duration := f.Points[len(f.Points)-1].Time
		numPoints := int(duration * 2)
		if numPoints < 100 {
			numPoints = 100
		}
		if numPoints > 1000 {
			numPoints = 1000
		}

		f.cachedBezierPoints = f.calculateBezierCurve(numPoints)
	}
	return f.cachedBezierPoints
}

// invalidateCache clears cached Bezier points when points change
func (f *Fade) invalidateCache() {
	f.cachedBezierPoints = nil
}

func extractNameFromPath(path string) string {
	base := filepath.Base(path)
	name := strings.TrimSuffix(base, ".bin")

	parts := strings.Split(name, "_")
	if len(parts) > 1 && strings.HasPrefix(parts[0], "fade") {
		return strings.Join(parts[1:], "_")
	}

	return name
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}