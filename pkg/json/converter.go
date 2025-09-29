package json

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"led-curve-tools/pkg/cue"
	"led-curve-tools/pkg/fade"
)

type Manifest struct {
	Fades []fade.Fade `json:"fades"`
	Cues  []cue.Cue   `json:"cues"`
}

func FadeToJSON(f *fade.Fade) ([]byte, error) {
	return json.MarshalIndent(f, "", "  ")
}

func CueToJSON(c *cue.Cue) ([]byte, error) {
	return json.MarshalIndent(c, "", "  ")
}

func FadeFromJSON(data []byte) (*fade.Fade, error) {
	var f fade.Fade
	err := json.Unmarshal(data, &f)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal fade JSON: %w", err)
	}
	return &f, nil
}

func CueFromJSON(data []byte) (*cue.Cue, error) {
	var c cue.Cue
	err := json.Unmarshal(data, &c)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal cue JSON: %w", err)
	}
	return &c, nil
}

func LoadFadeFromJSONFile(filename string) (*fade.Fade, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	return FadeFromJSON(data)
}

func LoadCueFromJSONFile(filename string) (*cue.Cue, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	return CueFromJSON(data)
}

func SaveFadeToJSONFile(f *fade.Fade, filename string) error {
	data, err := FadeToJSON(f)
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}

func SaveCueToJSONFile(c *cue.Cue, filename string) error {
	data, err := CueToJSON(c)
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}

func DetectFileType(filename string) (string, error) {
	base := strings.ToLower(filepath.Base(filename))

	if strings.Contains(base, "fade") {
		return "fade", nil
	}
	if strings.Contains(base, "cue") {
		return "cue", nil
	}

	ext := strings.ToLower(filepath.Ext(filename))
	if ext == ".json" {
		data, err := os.ReadFile(filename)
		if err != nil {
			return "", fmt.Errorf("failed to read file: %w", err)
		}

		var temp map[string]interface{}
		if err := json.Unmarshal(data, &temp); err != nil {
			return "", fmt.Errorf("invalid JSON: %w", err)
		}

		if _, hasSamples := temp["sampleRate"]; hasSamples {
			return "fade", nil
		}
		if _, hasActions := temp["actions"]; hasActions {
			return "cue", nil
		}

		return "", fmt.Errorf("cannot determine file type from JSON content")
	}

	if ext == ".bin" {
		file, err := os.Open(filename)
		if err != nil {
			return "", fmt.Errorf("failed to open file: %w", err)
		}
		defer file.Close()

		stat, err := file.Stat()
		if err != nil {
			return "", fmt.Errorf("failed to stat file: %w", err)
		}

		if stat.Size()%4 == 0 {
			return "cue", nil
		}
		if stat.Size()%2 == 0 {
			return "fade", nil
		}

		return "", fmt.Errorf("invalid binary file size")
	}

	return "", fmt.Errorf("cannot determine file type")
}

func LoadManifestFromFile(filename string) (*Manifest, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var manifest Manifest
	err = json.Unmarshal(data, &manifest)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal manifest JSON: %w", err)
	}

	return &manifest, nil
}

func SaveManifestToFile(manifest *Manifest, filename string) error {
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}
	return os.WriteFile(filename, data, 0644)
}

func LoadFromFile(filename string) (interface{}, string, error) {
	fileType, err := DetectFileType(filename)
	if err != nil {
		return nil, "", err
	}

	ext := strings.ToLower(filepath.Ext(filename))

	switch fileType {
	case "fade":
		if ext == ".json" {
			f, err := LoadFadeFromJSONFile(filename)
			return f, "fade", err
		} else {
			f, err := fade.LoadBinary(filename)
			return f, "fade", err
		}
	case "cue":
		if ext == ".json" {
			c, err := LoadCueFromJSONFile(filename)
			return c, "cue", err
		} else {
			c, err := cue.LoadBinary(filename)
			return c, "cue", err
		}
	default:
		return nil, "", fmt.Errorf("unsupported file type: %s", fileType)
	}
}