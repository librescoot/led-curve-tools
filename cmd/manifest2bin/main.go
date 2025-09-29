package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"led-curve-tools/pkg/json"
)

func main() {
	var (
		outputDir = flag.String("d", "output", "Output directory for binary files")
		help      = flag.Bool("h", false, "Show help")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <manifest.json>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nConverts a manifest.json file to binary fade/cue files\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s manifest.json\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -d /tmp/curves manifest.json\n", os.Args[0])
	}

	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	if flag.NArg() != 1 {
		fmt.Fprintf(os.Stderr, "Error: exactly one manifest file required\n\n")
		flag.Usage()
		os.Exit(1)
	}

	manifestFile := flag.Arg(0)

	if _, err := os.Stat(manifestFile); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: manifest file '%s' does not exist\n", manifestFile)
		os.Exit(1)
	}

	manifest, err := json.LoadManifestFromFile(manifestFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading manifest: %v\n", err)
		os.Exit(1)
	}

	err = os.MkdirAll(*outputDir, 0755)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	fadesDir := filepath.Join(*outputDir, "fades")
	cuesDir := filepath.Join(*outputDir, "cues")

	err = os.MkdirAll(fadesDir, 0755)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating fades directory: %v\n", err)
		os.Exit(1)
	}

	err = os.MkdirAll(cuesDir, 0755)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating cues directory: %v\n", err)
		os.Exit(1)
	}

	fadeCount := 0
	for _, fade := range manifest.Fades {
		var descriptiveName string
		if fade.DisplayName != "" {
			descriptiveName = fade.DisplayName
		} else {
			descriptiveName = fade.Name
		}
		filename := fmt.Sprintf("fade%d_%s", fade.FadeIndex, descriptiveName)
		outputPath := filepath.Join(fadesDir, filename)

		err = fade.WriteBinary(outputPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing fade '%s': %v\n", fade.Name, err)
			os.Exit(1)
		}
		fmt.Printf("Generated fade: %s -> %s\n", fade.Name, outputPath)
		fadeCount++
	}

	cueCount := 0
	for _, cue := range manifest.Cues {
		filename := fmt.Sprintf("cue%d_%s", cue.CueIndex, cue.Name)
		outputPath := filepath.Join(cuesDir, filename)

		err = cue.WriteBinary(outputPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing cue '%s': %v\n", cue.Name, err)
			os.Exit(1)
		}
		fmt.Printf("Generated cue: %s -> %s\n", cue.Name, outputPath)
		cueCount++
	}

	fmt.Printf("\nSuccessfully generated %d fades and %d cues to %s\n", fadeCount, cueCount, *outputDir)
}