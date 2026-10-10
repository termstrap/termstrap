package termstrap

import (
	"math"
	"strconv"
	"strings"
)

// Axis specifies the spatial dimension for unit projection.
type Axis int

const (
	// AxisX represents horizontal projection (columns / character cells).
	AxisX Axis = iota
	// AxisY represents vertical projection (lines / terminal rows).
	AxisY
)

// UnitContext holds layout dimensions for resolving relative and percentage units.
type UnitContext struct {
	ParentWidth    int     // Width of parent container in character columns
	ParentHeight   int     // Height of parent container in lines
	TerminalWidth  int     // Full terminal width in character columns
	TerminalHeight int     // Full terminal height in lines
	RootFontSize   float64 // Root font size in pixels (default: 16.0px = 1.0rem)
}

// DefaultUnitContext returns a fallback context when layout dimensions are not yet known.
func DefaultUnitContext(termWidth int) UnitContext {
	if termWidth <= 0 {
		termWidth = 80
	}
	return UnitContext{
		ParentWidth:    termWidth,
		ParentHeight:   24,
		TerminalWidth:  termWidth,
		TerminalHeight: 24,
		RootFontSize:   16.0,
	}
}

// ResolveLength converts any CSS length (px, rem, em, pt, in, ch, %, vw, vh, unitless)
// to terminal discrete units (character columns for AxisX, lines for AxisY).
func ResolveLength(val string, axis Axis, ctx UnitContext) int {
	val = strings.TrimSpace(val)
	if val == "" || val == "auto" || val == "inherit" || val == "initial" {
		return 0
	}

	// 1. Check for pure unitless number (e.g. "0", "1", "2")
	// Common in terminal-specific CSS or SVG attributes
	if num, err := strconv.ParseFloat(val, 64); err == nil {
		if num == 0 {
			return 0
		}
		if num > 0 && num < 1 && axis == AxisX {
			return 1
		}
		return int(math.Round(num))
	}

	// 2. Parse numeric value and unit suffix
	numPart, unitPart := splitNumberAndUnit(val)
	if numPart == "" {
		return 0
	}

	num, err := strconv.ParseFloat(numPart, 64)
	if err != nil {
		return 0
	}

	if num == 0 {
		return 0
	}

	rootFS := ctx.RootFontSize
	if rootFS <= 0 {
		rootFS = 16.0
	}

	unitPart = strings.ToLower(unitPart)
	var px float64
	isRelativePercent := false
	var percentFraction float64

	switch unitPart {
	case "px":
		px = num
	case "rem":
		px = num * rootFS
	case "em":
		px = num * rootFS
	case "ch":
		// 1ch is character advance width (approx 0.5rem = 8px)
		px = num * (rootFS * 0.5)
	case "pt":
		// 1pt = 1/72 inch = 1.3333px
		px = num * (96.0 / 72.0)
	case "in":
		px = num * 96.0
	case "cm":
		px = num * (96.0 / 2.54)
	case "mm":
		px = num * (96.0 / 25.4)
	case "%":
		isRelativePercent = true
		percentFraction = num / 100.0
	case "vw":
		tw := ctx.TerminalWidth
		if tw <= 0 {
			tw = 80
		}
		return int(math.Round((num / 100.0) * float64(tw)))
	case "vh":
		th := ctx.TerminalHeight
		if th <= 0 {
			th = 24
		}
		return int(math.Round((num / 100.0) * float64(th)))
	default:
		// Unknown unit or raw number
		px = num
	}

	// 3. Project to terminal cells
	if axis == AxisX {
		if isRelativePercent {
			pw := ctx.ParentWidth
			if pw <= 0 {
				pw = ctx.TerminalWidth
			}
			if pw <= 0 {
				pw = 80
			}
			cols := math.Round(percentFraction * float64(pw))
			if cols < 1 && percentFraction > 0 {
				return 1
			}
			if cols < 0 {
				return 0
			}
			return int(cols)
		}

		// Horizontal: 1 column cell ≈ 8px (0.5rem)
		colsFloat := px / (rootFS * 0.5)
		if colsFloat <= 0 {
			return 0
		}
		if colsFloat < 1.0 {
			// Minimum visual presence rule
			return 1
		}
		return int(math.Round(colsFloat))
	}

	// AxisY (vertical lines)
	if isRelativePercent {
		ph := ctx.ParentHeight
		if ph <= 0 {
			ph = ctx.TerminalHeight
		}
		if ph <= 0 {
			ph = 24
		}
		lines := math.Round(percentFraction * float64(ph))
		if lines < 0 {
			return 0
		}
		return int(lines)
	}

	// Vertical: 1 terminal row ≈ 16px (1.0rem)
	linesFloat := px / rootFS

	if linesFloat <= 0 {
		return 0
	}
	// Sub-line vertical threshold
	// Below 0.4rem (approx 6px, such as 0.25rem padding), return 0 to avoid wasteful blank lines
	if linesFloat < 0.4 {
		return 0
	}

	lines := int(math.Round(linesFloat))

	// Damping rule: limit maximum vertical padding/margin to prevent vertical screen overflow
	if lines > 3 {
		lines = 3
	}

	return lines
}

// splitNumberAndUnit separates a string like "1.5rem" into ("1.5", "rem").
func splitNumberAndUnit(val string) (string, string) {
	val = strings.TrimSpace(val)
	if val == "" {
		return "", ""
	}

	i := 0
	n := len(val)

	// Optional leading sign
	if i < n && (val[i] == '+' || val[i] == '-') {
		i++
	}

	hasDigits := false
	for i < n && ((val[i] >= '0' && val[i] <= '9') || val[i] == '.') {
		hasDigits = true
		i++
	}

	if !hasDigits {
		return "", val
	}

	num := val[:i]
	unit := strings.TrimSpace(val[i:])
	return num, unit
}

// ResolveBox4D parses 1, 2, 3, or 4 CSS values (e.g. "10px", "0.5rem 1rem", "10px 20px 30px")
// into (top, right, bottom, left) using axis-aware conversion.
func ResolveBox4D(val string, ctx UnitContext) (int, int, int, int) {
	val = strings.TrimSpace(val)
	if val == "" {
		return 0, 0, 0, 0
	}

	fields := strings.Fields(val)
	if len(fields) == 0 {
		return 0, 0, 0, 0
	}

	switch len(fields) {
	case 1:
		v := ResolveLength(fields[0], AxisY, ctx)
		h := ResolveLength(fields[0], AxisX, ctx)
		return v, h, v, h
	case 2:
		v := ResolveLength(fields[0], AxisY, ctx)
		h := ResolveLength(fields[1], AxisX, ctx)
		return v, h, v, h
	case 3:
		top := ResolveLength(fields[0], AxisY, ctx)
		h := ResolveLength(fields[1], AxisX, ctx)
		bottom := ResolveLength(fields[2], AxisY, ctx)
		return top, h, bottom, h
	default:
		top := ResolveLength(fields[0], AxisY, ctx)
		right := ResolveLength(fields[1], AxisX, ctx)
		bottom := ResolveLength(fields[2], AxisY, ctx)
		left := ResolveLength(fields[3], AxisX, ctx)
		return top, right, bottom, left
	}
}

// ResolveFontSize parses CSS font-size and returns typographic flags for terminal rendering.
func ResolveFontSize(val string, rootFS float64) (bold bool, dim bool, scale float64) {
	val = strings.TrimSpace(val)
	if val == "" {
		return false, false, 1.0
	}

	if rootFS <= 0 {
		rootFS = 16.0
	}

	numPart, unitPart := splitNumberAndUnit(val)
	if numPart == "" {
		lower := strings.ToLower(val)
		switch lower {
		case "xx-large", "x-large", "larger", "large":
			return true, false, 1.5
		case "small", "x-small", "xx-small", "smaller":
			return false, true, 0.8
		}
		return false, false, 1.0
	}

	num, err := strconv.ParseFloat(numPart, 64)
	if err != nil {
		return false, false, 1.0
	}

	var px float64
	unitPart = strings.ToLower(unitPart)
	switch unitPart {
	case "px":
		px = num
	case "rem", "em":
		px = num * rootFS
	case "pt":
		px = num * (96.0 / 72.0)
	case "%":
		px = (num / 100.0) * rootFS
	default:
		px = num
	}

	ratio := px / rootFS

	// Headers / large fonts (e.g. h1, h2, h3 or font-size >= 1.25rem / 20px)
	if ratio >= 1.25 {
		return true, false, ratio
	}

	// Small captions (e.g. font-size <= 0.85rem / 13px)
	if ratio <= 0.85 {
		return false, true, ratio
	}

	return false, false, ratio
}

// CheckIsLengthUnit determines if a token contains a recognized CSS unit.
func CheckIsLengthUnit(token string) bool {
	token = strings.ToLower(strings.TrimSpace(token))
	for _, u := range []string{"px", "rem", "em", "pt", "in", "cm", "mm", "%", "vw", "vh", "ch"} {
		if strings.HasSuffix(token, u) {
			prefix := strings.TrimSuffix(token, u)
			if _, err := strconv.ParseFloat(prefix, 64); err == nil {
				return true
			}
		}
	}
	return false
}
