package cue

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	ActionFade = 0
	ActionDuty = 1

	MaxLEDs = 8
)

type ActionType string

const (
	TypeFade     ActionType = "fade"
	TypeDuty     ActionType = "duty"
	TypeLastDuty ActionType = "last_duty"
)

type Action struct {
	LEDIndex int        `json:"ledIndex"`
	Type     ActionType `json:"type"`
	Value    float64    `json:"value"`
}

type Cue struct {
	Name     string   `json:"name"`
	CueIndex int      `json:"cueIndex"`
	Actions  []Action `json:"actions"`
	Filename string   `json:"filename"`
}

type BinaryAction struct {
	LEDIndex   uint8
	ActionFlag uint8
	Value      uint16
}

func LoadBinary(filename string) (*Cue, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	actionCount := stat.Size() / 4
	if stat.Size()%4 != 0 {
		return nil, fmt.Errorf("invalid cue file: size not multiple of 4 bytes")
	}

	actions := make([]Action, 0, actionCount)

	for i := int64(0); i < actionCount; i++ {
		var binAction BinaryAction
		err = binary.Read(file, binary.LittleEndian, &binAction)
		if err != nil {
			return nil, fmt.Errorf("failed to read action %d: %w", i, err)
		}

		if binAction.LEDIndex >= MaxLEDs {
			return nil, fmt.Errorf("invalid LED index %d in action %d", binAction.LEDIndex, i)
		}

		action := Action{
			LEDIndex: int(binAction.LEDIndex),
		}

		switch binAction.ActionFlag {
		case ActionFade:
			action.Type = TypeFade
			action.Value = float64(binAction.Value)
		case ActionDuty:
			action.Type = TypeDuty
			action.Value = float64(binAction.Value) / 12000.0
		default:
			return nil, fmt.Errorf("invalid action flag %d in action %d", binAction.ActionFlag, i)
		}

		actions = append(actions, action)
	}

	cue := &Cue{
		Name:     extractNameFromPath(filename),
		Actions:  actions,
		Filename: strings.TrimSuffix(filepath.Base(filename), ".bin"),
	}

	return cue, nil
}

func (c *Cue) WriteBinary(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	for i, action := range c.Actions {
		if action.LEDIndex < 0 || action.LEDIndex >= MaxLEDs {
			return fmt.Errorf("invalid LED index %d in action %d", action.LEDIndex, i)
		}

		binAction := BinaryAction{
			LEDIndex: uint8(action.LEDIndex),
		}

		switch action.Type {
		case TypeFade:
			binAction.ActionFlag = ActionFade
			binAction.Value = uint16(action.Value)
		case TypeDuty:
			binAction.ActionFlag = ActionDuty
			binAction.Value = uint16(action.Value * 12000.0)
		case TypeLastDuty:
			binAction.ActionFlag = ActionFade
			binAction.Value = uint16(action.Value)
		default:
			return fmt.Errorf("invalid action type %s in action %d", action.Type, i)
		}

		err = binary.Write(file, binary.LittleEndian, &binAction)
		if err != nil {
			return fmt.Errorf("failed to write action %d: %w", i, err)
		}
	}

	return nil
}

func (c *Cue) String() string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("Cue: %s\n", c.Name))
	for i, action := range c.Actions {
		builder.WriteString(fmt.Sprintf("  Action %d: LED %d, %s", i, action.LEDIndex, action.Type))
		if action.Type == TypeDuty {
			builder.WriteString(fmt.Sprintf(" %.3f", action.Value))
		} else {
			builder.WriteString(fmt.Sprintf(" %d", int(action.Value)))
		}
		builder.WriteString("\n")
	}
	return builder.String()
}

func extractNameFromPath(path string) string {
	base := filepath.Base(path)
	name := strings.TrimSuffix(base, ".bin")

	parts := strings.Split(name, "_")
	if len(parts) > 1 && strings.HasPrefix(parts[0], "cue") {
		return strings.Join(parts[1:], "_")
	}

	return name
}