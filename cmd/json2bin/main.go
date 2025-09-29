package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"led-curve-tools/pkg/cue"
	"led-curve-tools/pkg/fade"
	"led-curve-tools/pkg/json"
)

func main() {
	var (
		outputFile = flag.String("o", "", "Output binary file (default: input.bin)")
		help       = flag.Bool("h", false, "Show help")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <json-file>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nConverts JSON fade/cue files to binary format\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s fade_config.json\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -o fade00.bin fade_definition.json\n", os.Args[0])
	}

	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	if flag.NArg() != 1 {
		fmt.Fprintf(os.Stderr, "Error: exactly one input file required\n\n")
		flag.Usage()
		os.Exit(1)
	}

	inputFile := flag.Arg(0)

	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: input file '%s' does not exist\n", inputFile)
		os.Exit(1)
	}

	data, fileType, err := json.LoadFromFile(inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading file: %v\n", err)
		os.Exit(1)
	}

	output := *outputFile
	if output == "" {
		base := strings.TrimSuffix(inputFile, filepath.Ext(inputFile))
		output = base + ".bin"
	}

	switch fileType {
	case "fade":
		f := data.(*fade.Fade)
		err = f.WriteBinary(output)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing fade binary: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Successfully converted fade JSON to %s (%d samples)\n", output, len(f.Samples))

	case "cue":
		c := data.(*cue.Cue)
		err = c.WriteBinary(output)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing cue binary: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Successfully converted cue JSON to %s (%d actions)\n", output, len(c.Actions))

	default:
		fmt.Fprintf(os.Stderr, "Error: unsupported file type: %s\n", fileType)
		os.Exit(1)
	}
}