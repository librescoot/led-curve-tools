# LED Curve Tools

A collection of command-line tools for working with LED fade curves and cue sequences used in unu Scooter Pro. These tools handle both linear and bezier curve interpolation, providing bidirectional conversion between binary and human-readable JSON formats, plus comprehensive visualization capabilities.

## What These Tools Do

The unu Scooter Pro uses PWM (Pulse Width Modulation) to control brightness across multiple LED channels like the front light ring, headlight beam, brake lights, and turn signals. The system stores lighting patterns as either:

- **Fade curves**: Smooth brightness transitions over time (like a gentle power-on sequence)
- **Cue sequences**: Instant commands that trigger multiple LEDs simultaneously

These tools let you convert between the binary format the hardware uses and human-readable JSON, visualize the curves, and work with complete lighting manifests from design tools.

## Tools Overview

- **bin2json** - Converts binary fade/cue files to JSON format
- **json2bin** - Converts JSON fade/cue files to binary format
- **ledviz** - Visualizes LED curves and cue sequences with graphical and ASCII output
- **manifest2bin** - Converts entire manifest.json files to binary fade/cue collections

## Building

```bash
go mod tidy
go build -o bin/bin2json ./cmd/bin2json
go build -o bin/json2bin ./cmd/json2bin
go build -o bin/ledviz ./cmd/ledviz
go build -o bin/manifest2bin ./cmd/manifest2bin
```

## Usage

### bin2json - Binary to JSON Converter

Convert binary fade/cue files to human-readable JSON format:

```bash
./bin/bin2json fade0
./bin/bin2json -o output.json cue01_standby
```

### json2bin - JSON to Binary Converter

Convert JSON fade/cue definitions to binary format. Supports both linear and bezier curve interpolation:

```bash
./bin/json2bin fade_config.json
./bin/json2bin -o fade0 fade_definition.json

# Bezier curves are automatically handled when curveType is specified in JSON
# Linear interpolation is used by default for compatibility
```

### ledviz - LED Curve Visualizer

Visualize LED fade curves and cue sequences. Uses logarithmic Y-axis scaling by default because LED brightness perception is logarithmic - small changes at low brightness are more visible than the same change at high brightness:

#### ASCII Visualization
```bash
# Single fade curve
./bin/ledviz -format ascii fade0

# Multiple fade comparison
./bin/ledviz -format ascii fade0 fade01 fade02

# Cue sequence
./bin/ledviz -format ascii cue01
```

#### Graphical Visualization
```bash
# Single fade curve (logarithmic Y-axis by default)
./bin/ledviz fade0

# Multiple fade comparison with descriptive title
./bin/ledviz -o comparison.png -title "Fade Comparison" fade*

# Linear Y-axis scale instead of logarithmic
./bin/ledviz -linear -o linear_comparison.png fade0 fade01

# High-resolution output with custom dimensions
./bin/ledviz -width 1920 -height 1080 -o hires.png fade0

# Disable grid lines or legend
./bin/ledviz -grid=false -legend=false fade0
```

### manifest2bin - Bulk Manifest Converter

Convert entire manifest.json files (from led-designer) to binary fade/cue collections:

```bash
# Convert manifest to binary files with descriptive names
./bin/manifest2bin manifest.json

# Specify custom output directory
./bin/manifest2bin -d /path/to/output manifest.json

# Creates organized directory structure:
# output/
# ├── fades/
# │   ├── fade0_smooth-on
# │   ├── fade1_smooth-off
# │   └── fade10_blink
# └── cues/
#     ├── cue0_all_off
#     ├── cue10_blink_left
#     └── cue11_blink_right
```

## File Formats

### Binary Format

#### Fade Files
- 16-bit little-endian samples
- Sample rate: 250Hz (4ms per sample)
- Value range: 0-12000 (PWM duty cycle)
- PWM period: 12000

#### Cue Files
- 4-byte action structures:
  - Byte 0: LED index (0-7)
  - Byte 1: Action type (0=fade, 1=duty)
  - Bytes 2-3: Value (little-endian, 16-bit)

### JSON Format

#### Fade JSON Schema
```json
{
  "name": "fade_name",
  "displayName": "Fade Display Name",
  "points": [
    {
      "time": 0.0,     // Time in milliseconds
      "duty": 0.0      // Duty cycle (0.0-1.0)
    }
  ],
  "sampleRate": 250,   // Samples per second
  "curveType": "bezier", // "linear" or "bezier" (optional, defaults to linear)
  "fadeIndex": 0,      // Index for binary export
  "filename": "fade0"  // Binary filename
}
```

#### Cue JSON Schema
```json
{
  "name": "cue_name",
  "cueIndex": 0,
  "actions": [
    {
      "ledIndex": 0,        // LED channel (0-7)
      "type": "fade",       // "fade", "duty", or "last_duty"
      "value": 0.0          // Fade index or duty cycle
    }
  ],
  "filename": "cue0"
}
```

#### Manifest JSON Schema
```json
{
  "fades": [
    {
      "name": "fade0",
      "displayName": "smooth-on",
      "points": [...],
      "sampleRate": 250,
      "curveType": "bezier",  // Curve interpolation type
      "fadeIndex": 0,
      "filename": "fade0"
    }
  ],
  "cues": [
    {
      "name": "all_off",
      "cueIndex": 0,
      "actions": [...],
      "filename": "cue0"
    }
  ]
}
```

## Examples

### Converting Binary to JSON
```bash
# Convert fade file
./bin/bin2json fade0
# Output: fade0.json

# Convert cue file
./bin/bin2json cue01
# Output: cue01.json
```

### Creating Visual Comparisons
```bash
# ASCII comparison of multiple fades
./bin/ledviz -format ascii fade0 fade10

# High-resolution PNG comparison with logarithmic scale (default)
./bin/ledviz -width 1920 -height 1080 -title "Ring vs Blinker Patterns" \
  -o comparison.png fade0 fade10

# Linear scale comparison for high-brightness curves
./bin/ledviz -linear -title "High Brightness Comparison" \
  -o linear_comparison.png fade6 fade8
```

### Bulk Processing with Manifest
```bash
# Convert entire manifest to organized binary collection
./bin/manifest2bin led-designer/example/manifest.json

# Generate all fades and cues from manifest with descriptive names:
# - fade0_smooth-on, fade1_smooth-off, fade10_blink
# - cue0_all_off, cue10_blink_left, cue11_blink_right
```

### Round-trip Conversion
```bash
# Binary -> JSON -> Binary
./bin/bin2json original          # Creates original.json
./bin/json2bin -o converted original.json
```

## Key Features

### Curve Interpolation
- **Linear interpolation**: Straight lines between control points (default for compatibility)
- **Bezier curves**: Smooth curves using all control points as Bezier curve anchors
- **Automatic detection**: JSON `curveType` field determines interpolation method
- **High precision**: Adaptive resolution ensures smooth curves at any scale

### Visualization
- **Logarithmic Y-axis scaling** by default - matches human LED brightness perception
- **Linear scale option** with `--linear` flag for technical analysis
- **Multi-fade visualization** with color-coded curves and endpoint markers
- **Professional graphical output** with legends positioned outside plot area
- **Detailed tick marks** - major ticks with labels, minor ticks for precision
- **Both ASCII and graphical** output formats for different use cases

### File Processing
- **Auto-detection** of file types (fade vs cue) based on content analysis
- **Bulk manifest processing** - convert entire led-designer manifests at once
- **Descriptive filenames** - automatically generates fade0_smooth-on, cue10_blink_left
- **Full round-trip** conversion support with data integrity preservation
- **Optimized point reduction** removes redundant data points while preserving curve shape
- **Error handling** with validation for LED indices, duty cycles, and curve parameters

## Dependencies

- Go 1.19 or later
- github.com/fogleman/gg (for graphical output)

## Understanding the LED System

### Hardware Constants
- **PWM Period**: 12000 ticks (hardware-specific timing)
- **Sample Rate**: 250Hz (4ms per sample) - smooth enough for human perception
- **Duty Cycle Range**: 0.0-1.0 (0-100% brightness)
- **LED Channels**: 8 total (0-7) mapped to different lights:

| Index | LED Name              | Description                    |
|-------|-----------------------|--------------------------------|
| 0     | Headlight            | Main front illumination       |
| 1     | Front ring           | Front accent lighting          |
| 2     | Brake light          | Rear brake indicator           |
| 3     | Blinker front left   | Left front turn signal        |
| 4     | Blinker front right  | Right front turn signal       |
| 5     | Number plates        | License plate illumination    |
| 6     | Blinker rear left    | Left rear turn signal         |
| 7     | Blinker rear right   | Right rear turn signal        |

Channels 3, 4, 6, and 7 are configured as blinker channels and do not use adaptive mode.

### Why Logarithmic Scaling?
Human eyes perceive LED brightness logarithmically - doubling from 1% to 2% is much more visible than doubling from 50% to 100%. The logarithmic Y-axis in visualizations matches this perception, making subtle low-brightness changes clearly visible while not wasting space on barely-perceptible high-brightness differences.

### Curve Types Explained
- **Linear**: Each control point connects with straight lines. Simple and predictable.
- **Bezier**: All control points work together to create a smooth curve. The curve doesn't necessarily pass through middle control points, but is influenced by all of them. Perfect for smooth, natural-looking LED transitions like gentle power-on sequences.