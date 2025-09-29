package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"led-curve-tools/pkg/cue"
	"led-curve-tools/pkg/fade"
	"led-curve-tools/pkg/json"
	"led-curve-tools/pkg/viz"
)

type OutputFormat string

const (
	FormatPNG   OutputFormat = "png"
	FormatASCII OutputFormat = "ascii"
)

func main() {
	var (
		outputFile   = flag.String("o", "", "Output file (default: visualization.png)")
		format       = flag.String("format", "png", "Output format: png, ascii")
		width        = flag.Int("width", 1200, "Image width for graphical output")
		height       = flag.Int("height", 800, "Image height for graphical output")
		title        = flag.String("title", "", "Graph title (default: auto-generated)")
		asciiWidth   = flag.Int("ascii-width", 80, "ASCII visualization width")
		asciiHeight  = flag.Int("ascii-height", 20, "ASCII visualization height")
		showGrid     = flag.Bool("grid", true, "Show grid lines")
		showLegend   = flag.Bool("legend", true, "Show legend for multi-fade graphs")
		linearScale  = flag.Bool("linear", false, "Use linear Y-axis scale (default: logarithmic)")
		help         = flag.Bool("h", false, "Show help")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <file1> [file2] [file3] ...\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nVisualizes LED fade curves and cue sequences\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nSupported file formats:\n")
		fmt.Fprintf(os.Stderr, "  .bin - Binary fade/cue files\n")
		fmt.Fprintf(os.Stderr, "  .json - JSON fade/cue files\n")
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s fade00.bin\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -format ascii fade00.bin fade01.bin\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -o comparison.png -title \"Fade Comparison\" fade*.bin\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -width 1920 -height 1080 cue00.bin\n", os.Args[0])
	}

	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	if flag.NArg() == 0 {
		fmt.Fprintf(os.Stderr, "Error: at least one input file required\n\n")
		flag.Usage()
		os.Exit(1)
	}

	outputFormat := OutputFormat(strings.ToLower(*format))
	if outputFormat != FormatPNG && outputFormat != FormatASCII {
		fmt.Fprintf(os.Stderr, "Error: unsupported format '%s'\n", *format)
		os.Exit(1)
	}

	inputFiles := flag.Args()
	fades := []*fade.Fade{}
	cues := []*cue.Cue{}

	for _, filename := range inputFiles {
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Error: file '%s' does not exist\n", filename)
			os.Exit(1)
		}

		data, fileType, err := json.LoadFromFile(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading file '%s': %v\n", filename, err)
			os.Exit(1)
		}

		switch fileType {
		case "fade":
			fades = append(fades, data.(*fade.Fade))
		case "cue":
			cues = append(cues, data.(*cue.Cue))
		default:
			fmt.Fprintf(os.Stderr, "Error: unsupported file type '%s' for file '%s'\n", fileType, filename)
			os.Exit(1)
		}
	}

	if len(fades) > 0 && len(cues) > 0 {
		fmt.Fprintf(os.Stderr, "Error: cannot mix fade and cue files in same visualization\n")
		os.Exit(1)
	}

	if len(fades) > 0 {
		err := visualizeFades(fades, outputFormat, *outputFile, *width, *height, *title, *asciiWidth, *asciiHeight, *showGrid, *showLegend, *linearScale)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating fade visualization: %v\n", err)
			os.Exit(1)
		}
	} else if len(cues) > 0 {
		err := visualizeCues(cues, outputFormat, *outputFile, *width, *height, *title, *asciiWidth, *asciiHeight)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating cue visualization: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Fprintf(os.Stderr, "Error: no valid fade or cue files found\n")
		os.Exit(1)
	}
}

func visualizeFades(fades []*fade.Fade, format OutputFormat, outputFile string, width, height int, title string, asciiWidth, asciiHeight int, showGrid, showLegend, linearScale bool) error {
	if format == FormatASCII {
		if len(fades) == 1 {
			fmt.Print(viz.RenderFadeASCII(fades[0], asciiWidth, asciiHeight))
		} else {
			fmt.Print(viz.RenderMultipleFadesASCII(fades, asciiWidth, asciiHeight))
		}
		return nil
	}

	config := viz.DefaultGraphConfig()
	config.Width = width
	config.Height = height
	config.ShowGrid = showGrid
	config.ShowLegend = showLegend && len(fades) > 1
	config.LogScale = !linearScale

	if title != "" {
		config.Title = title
	} else if len(fades) == 1 {
		name := fades[0].Name
		if name == "" {
			name = fades[0].Filename
		}
		config.Title = fmt.Sprintf("Fade: %s", name)
	} else {
		config.Title = fmt.Sprintf("Fade Comparison (%d curves)", len(fades))
	}

	output := outputFile
	if output == "" {
		if len(fades) == 1 {
			output = fades[0].Filename + "_visualization.png"
		} else {
			output = "fade_comparison.png"
		}
	}

	err := viz.CreateFadeGraph(fades, config, output)
	if err != nil {
		return err
	}

	fmt.Printf("Successfully created fade visualization: %s\n", output)
	return nil
}

func visualizeCues(cues []*cue.Cue, format OutputFormat, outputFile string, width, height int, title string, asciiWidth, asciiHeight int) error {
	if format == FormatASCII {
		for i, c := range cues {
			if i > 0 {
				fmt.Println()
			}
			fmt.Print(viz.RenderCueASCII(c))
		}
		return nil
	}

	if len(cues) > 1 {
		fmt.Fprintf(os.Stderr, "Warning: graphical visualization of multiple cues not fully implemented, showing first cue only\n")
	}

	config := viz.DefaultGraphConfig()
	config.Width = width
	config.Height = height

	if title != "" {
		config.Title = title
	} else {
		name := cues[0].Name
		if name == "" {
			name = cues[0].Filename
		}
		config.Title = fmt.Sprintf("Cue: %s", name)
	}

	output := outputFile
	if output == "" {
		output = cues[0].Filename + "_visualization.png"
	}

	err := viz.CreateCueVisualization(cues[0], config, output)
	if err != nil {
		return err
	}

	fmt.Printf("Successfully created cue visualization: %s\n", output)
	return nil
}