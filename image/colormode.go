package image

import (
	"strings"
)

// ColorMode defines the color palette depth used for ANSI rendering.
type ColorMode int

const (
	// ColorModeTrueColor uses 24-bit RGB ANSI escape sequences (\033[38;2;R;G;Bm).
	// Default mode with highest color fidelity (16.7M colors).
	ColorModeTrueColor ColorMode = iota

	// ColorMode256 quantizes colors to the standard xterm 256-color palette (\033[38;5;Nm).
	// Generates 50-60% fewer escape sequence bytes for faster rendering over SSH.
	ColorMode256

	// ColorMode16 quantizes colors to the 16 basic/bright ANSI colors.
	ColorMode16
)

// String returns the name of the color mode.
func (cm ColorMode) String() string {
	switch cm {
	case ColorMode256:
		return "256"
	case ColorMode16:
		return "16"
	default:
		return "truecolor"
	}
}

// ParseColorMode parses a string into a ColorMode.
func ParseColorMode(val string) (ColorMode, bool) {
	val = strings.ToLower(strings.TrimSpace(val))
	switch val {
	case "256", "ansi256", "256color", "8bit":
		return ColorMode256, true
	case "16", "ansi16", "16color", "4bit", "basic":
		return ColorMode16, true
	case "truecolor", "24bit", "rgb", "tc":
		return ColorModeTrueColor, true
	default:
		return ColorModeTrueColor, false
	}
}

// RGBTo256 converts 24-bit RGB values to the nearest xterm 256-color palette index (0..255).
func RGBTo256(r, g, b uint8) uint8 {
	// 1. Check if color is close to grayscale
	diffRG := int(r) - int(g)
	if diffRG < 0 {
		diffRG = -diffRG
	}
	diffGB := int(g) - int(b)
	if diffGB < 0 {
		diffGB = -diffGB
	}
	diffRB := int(r) - int(b)
	if diffRB < 0 {
		diffRB = -diffRB
	}

	if diffRG <= 10 && diffGB <= 10 && diffRB <= 10 {
		gray := (int(r) + int(g) + int(b)) / 3
		if gray < 4 {
			return 16 // black (cube 0,0,0)
		}
		if gray > 248 {
			return 231 // white (cube 5,5,5)
		}
		// Map to 24-step grayscale ramp: 232..255 (values 8..238 with step ~10)
		rampIdx := (gray - 8 + 5) / 10
		if rampIdx < 0 {
			rampIdx = 0
		} else if rampIdx > 23 {
			rampIdx = 23
		}
		return uint8(232 + rampIdx)
	}

	// 2. Map to 6x6x6 color cube (indices 16..231)
	rIdx := nearestCubeIndex(r)
	gIdx := nearestCubeIndex(g)
	bIdx := nearestCubeIndex(b)

	return uint8(16 + 36*rIdx + 6*gIdx + bIdx)
}

func nearestCubeIndex(c uint8) int {
	if c < 48 {
		return 0
	}
	if c < 115 {
		return 1
	}
	return int((c - 35) / 40)
}

// RGBTo16 converts 24-bit RGB values to the nearest 16 ANSI color index (0..15).
func RGBTo16(r, g, b uint8) uint8 {
	idx := 0
	if r > 64 {
		idx |= 1
	}
	if g > 64 {
		idx |= 2
	}
	if b > 64 {
		idx |= 4
	}
	// Bright color if any component is bright enough
	if r > 192 || g > 192 || b > 192 {
		idx |= 8
	}
	return uint8(idx)
}

var cubeSteps = [6]uint8{0, 95, 135, 175, 215, 255}

var standard16Colors = [16][3]uint8{
	{0, 0, 0},       // 0: black
	{128, 0, 0},     // 1: red
	{0, 128, 0},     // 2: green
	{128, 128, 0},   // 3: yellow
	{0, 0, 128},     // 4: blue
	{128, 0, 128},   // 5: magenta
	{0, 128, 128},   // 6: cyan
	{192, 192, 192}, // 7: light gray
	{128, 128, 128}, // 8: dark gray
	{255, 0, 0},     // 9: bright red
	{0, 255, 0},     // 10: bright green
	{255, 255, 0},   // 11: bright yellow
	{0, 0, 255},     // 12: bright blue
	{255, 0, 255},   // 13: bright magenta
	{0, 255, 255},   // 14: bright cyan
	{255, 255, 255}, // 15: bright white
}

// Palette256ToRGB converts an xterm 256-color index to its corresponding RGB values.
func Palette256ToRGB(idx uint8) (r, g, b uint8) {
	if idx < 16 {
		c := standard16Colors[idx]
		return c[0], c[1], c[2]
	}
	if idx >= 232 {
		// Grayscale ramp: 24 steps from 8 to 238
		gray := uint8(8 + 10*int(idx-232))
		return gray, gray, gray
	}
	// 6x6x6 color cube: indices 16..231
	cube := idx - 16
	rIdx := cube / 36
	gIdx := (cube % 36) / 6
	bIdx := cube % 6
	return cubeSteps[rIdx], cubeSteps[gIdx], cubeSteps[bIdx]
}

// Palette16ToRGB converts an ANSI 16-color index to its corresponding RGB values.
func Palette16ToRGB(idx uint8) (r, g, b uint8) {
	c := standard16Colors[idx%16]
	return c[0], c[1], c[2]
}
