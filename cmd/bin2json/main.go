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
		outputFile = flag.String("o", "", "Output JSON file (default: input.json)")
		help       = flag.Bool("h", false, "Show help")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <binary-file>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nConverts binary fade/cue files to JSON format\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s fade00_ring_off_to_full.bin\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -o output.json cue00_all_off.bin\n", os.Args[0])
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

	var jsonData []byte

	switch fileType {
	case "fade":
		jsonData, err = json.FadeToJSON(data.(*fade.Fade))
	case "cue":
		jsonData, err = json.CueToJSON(data.(*cue.Cue))
	default:
		fmt.Fprintf(os.Stderr, "Error: unsupported file type: %s\n", fileType)
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error converting to JSON: %v\n", err)
		os.Exit(1)
	}

	output := *outputFile
	if output == "" {
		base := strings.TrimSuffix(inputFile, filepath.Ext(inputFile))
		output = base + ".json"
	}

	err = os.WriteFile(output, jsonData, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully converted %s to %s (%s)\n", inputFile, output, fileType)
}