package termstrap

import (
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
)

// SpinnerFrames contains classic Braille spinner frames.
var SpinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// ShimmerOptions configures a shimmer gradient animation over text.
type ShimmerOptions struct {
	BaseColor      string // Color for idle characters (default #44475a)
	HighlightColor string // Peak glow color (default #ffffff)
	Bold           bool   // Whether to style text as bold
}

// Shimmer applies a dynamic light beam / wave across the given text for a given animation step.
func Shimmer(text string, step int, opts ...ShimmerOptions) string {
	if text == "" {
		return ""
	}
	opt := ShimmerOptions{
		BaseColor:      "#44475a",
		HighlightColor: "#ffffff",
		Bold:           true,
	}
	if len(opts) > 0 {
		if opts[0].BaseColor != "" {
			opt.BaseColor = opts[0].BaseColor
		}
		if opts[0].HighlightColor != "" {
			opt.HighlightColor = opts[0].HighlightColor
		}
		opt.Bold = opts[0].Bold
	}

	runes := []rune(text)
	n := len(runes)
	if n == 0 {
		return ""
	}

	baseStops := []color.Color{
		lipgloss.Color(opt.BaseColor),
		lipgloss.Color(opt.HighlightColor),
		lipgloss.Color(opt.BaseColor),
	}

	paletteSteps := n * 2
	if paletteSteps < 3 {
		paletteSteps = 3
	}
	palette := lipgloss.Blend1D(paletteSteps, baseStops...)

	var sb strings.Builder
	for i, r := range runes {
		colorIdx := (i - step) % len(palette)
		if colorIdx < 0 {
			colorIdx += len(palette)
		}
		c := palette[colorIdx]
		st := lipgloss.NewStyle().Foreground(c).Bold(opt.Bold)
		sb.WriteString(st.Render(string(r)))
	}
	return sb.String()
}

// Spinner returns the spinner frame glyph for a given step.
func Spinner(step int, frames ...[]string) string {
	f := SpinnerFrames
	if len(frames) > 0 && len(frames[0]) > 0 {
		f = frames[0]
	}
	idx := step % len(f)
	if idx < 0 {
		idx += len(f)
	}
	return f[idx]
}
