package viz

import (
	"image/color"
)

var DefaultColors = []color.RGBA{
	{255, 99, 132, 255},  // Red
	{54, 162, 235, 255},  // Blue
	{255, 205, 86, 255},  // Yellow
	{75, 192, 192, 255},  // Green
	{153, 102, 255, 255}, // Purple
	{255, 159, 64, 255},  // Orange
	{199, 199, 199, 255}, // Grey
	{83, 102, 255, 255},  // Light Blue
	{255, 99, 255, 255},  // Pink
	{99, 255, 132, 255},  // Light Green
}

func GetColor(index int) color.RGBA {
	return DefaultColors[index%len(DefaultColors)]
}

var (
	ColorBackground = color.RGBA{255, 255, 255, 255}
	ColorGrid       = color.RGBA{230, 230, 230, 255}
	ColorAxis       = color.RGBA{100, 100, 100, 255}
	ColorText       = color.RGBA{50, 50, 50, 255}
)