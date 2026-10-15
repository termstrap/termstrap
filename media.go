package termstrap

import (
	"math"
	"strconv"
	"strings"
)

// MediaClause represents an atomic media query conjunction (separated by "and").
type MediaClause struct {
	MediaType string // "screen", "all", "print", etc.
	Negate    bool   // "not screen"

	// Boundary constraints (-1 means unconstrained)
	MinWidth  int
	MaxWidth  int
	MinHeight int
	MaxHeight int

	Orientation string // "landscape", "portrait", or ""
	ColorScheme string // "dark", "light", or ""
}

// MediaCondition represents a set of media clauses combined with OR (comma-separated in CSS).
type MediaCondition struct {
	Clauses []MediaClause
}

// ParseMediaQuery parses a CSS @media prelude (e.g. "(min-width: 768px)", "screen and (max-width: 50ch)").
func ParseMediaQuery(prelude string) *MediaCondition {
	prelude = strings.TrimSpace(prelude)
	if prelude == "" {
		return nil
	}

	// Comma-separated media queries act as logical OR
	rawParts := strings.Split(prelude, ",")
	var clauses []MediaClause

	for _, part := range rawParts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		clause := parseMediaClause(part)
		clauses = append(clauses, clause)
	}

	if len(clauses) == 0 {
		return nil
	}

	return &MediaCondition{Clauses: clauses}
}

// parseMediaClause parses a single query expression with optional "and" operators.
func parseMediaClause(expr string) MediaClause {
	clause := MediaClause{
		MediaType: "all",
		MinWidth:  -1,
		MaxWidth:  -1,
		MinHeight: -1,
		MaxHeight: -1,
	}

	tokens := strings.Split(expr, "and")
	for _, rawTok := range tokens {
		tok := strings.TrimSpace(rawTok)
		if tok == "" {
			continue
		}

		// Check for parentheses expression e.g. (min-width: 768px)
		if strings.HasPrefix(tok, "(") && strings.HasSuffix(tok, ")") {
			inner := strings.TrimSpace(tok[1 : len(tok)-1])
			parseFeature(&clause, inner)
			continue
		}

		// Media type e.g. "screen", "only screen", "not print"
		lower := strings.ToLower(tok)
		if strings.HasPrefix(lower, "not ") {
			clause.Negate = true
			clause.MediaType = strings.TrimSpace(lower[4:])
		} else if strings.HasPrefix(lower, "only ") {
			clause.MediaType = strings.TrimSpace(lower[5:])
		} else {
			clause.MediaType = lower
		}
	}

	return clause
}

// parseFeature parses key:value pairs like "min-width: 768px" or "prefers-color-scheme: dark".
func parseFeature(clause *MediaClause, feature string) {
	parts := strings.SplitN(feature, ":", 2)
	if len(parts) != 2 {
		return
	}

	key := strings.TrimSpace(strings.ToLower(parts[0]))
	val := strings.TrimSpace(parts[1])

	ctx := DefaultUnitContext(80)

	switch key {
	case "min-width":
		clause.MinWidth = resolveMediaLengthFloorCeil(val, AxisX, ctx, true)
	case "max-width":
		clause.MaxWidth = resolveMediaLengthFloorCeil(val, AxisX, ctx, false)
	case "width":
		w := resolveMediaLengthFloorCeil(val, AxisX, ctx, true)
		clause.MinWidth = w
		clause.MaxWidth = w

	case "min-height":
		clause.MinHeight = resolveMediaLengthFloorCeil(val, AxisY, ctx, true)
	case "max-height":
		clause.MaxHeight = resolveMediaLengthFloorCeil(val, AxisY, ctx, false)
	case "height":
		h := resolveMediaLengthFloorCeil(val, AxisY, ctx, true)
		clause.MinHeight = h
		clause.MaxHeight = h

	case "orientation":
		clause.Orientation = strings.ToLower(val)
	case "prefers-color-scheme":
		clause.ColorScheme = strings.ToLower(val)
	}
}

// resolveMediaLengthFloorCeil converts CSS length to columns/lines with proper boundary rounding.
func resolveMediaLengthFloorCeil(val string, axis Axis, ctx UnitContext, isMin bool) int {
	numPart, unitPart := splitNumberAndUnit(val)
	if numPart == "" {
		return ResolveLength(val, axis, ctx)
	}

	num, err := strconv.ParseFloat(numPart, 64)
	if err != nil {
		return ResolveLength(val, axis, ctx)
	}

	rootFS := ctx.RootFontSize
	if rootFS <= 0 {
		rootFS = 16.0
	}

	var px float64
	unitPart = strings.ToLower(unitPart)
	switch unitPart {
	case "px":
		px = num
	case "rem", "em":
		px = num * rootFS
	case "ch":
		px = num * (rootFS * 0.5)
	case "pt":
		px = num * (96.0 / 72.0)
	default:
		return ResolveLength(val, axis, ctx)
	}

	if axis == AxisX {
		cols := px / (rootFS * 0.5)
		if isMin {
			// For min-width: round standard
			return int(math.Round(cols))
		}
		// For max-width: floor so max-width: 575.98px becomes 71 cols when min-width: 576px is 72 cols
		return int(math.Floor(cols))
	}

	lines := px / rootFS
	if isMin {
		return int(math.Round(lines))
	}
	return int(math.Floor(lines))
}

// Matches evaluates whether the media condition holds true for current terminal dimensions and theme.
func (m *MediaCondition) Matches(termWidth, termHeight int, theme Theme) bool {
	if m == nil || len(m.Clauses) == 0 {
		return true
	}

	if termWidth <= 0 {
		termWidth = 80
	}
	if termHeight <= 0 {
		termHeight = 24
	}

	themeStr := string(theme)
	if themeStr == "" {
		themeStr = "bootstrap"
	}
	isDark := theme == ThemeDracula || theme == ThemeTokyoNight || theme == ThemeGlamour || themeStr == "dark"

	// OR condition: matches if any clause matches
	for _, c := range m.Clauses {
		if matchClause(c, termWidth, termHeight, isDark) {
			return true
		}
	}

	return false
}

func matchClause(c MediaClause, termWidth, termHeight int, isDark bool) bool {
	matchesType := true
	switch c.MediaType {
	case "print":
		matchesType = false // Terminal rendering is a screen device
	case "speech":
		matchesType = false
	case "screen", "all", "":
		matchesType = true
	default:
		matchesType = true
	}

	if c.Negate {
		matchesType = !matchesType
	}
	if !matchesType {
		return false
	}

	// Width boundaries
	if c.MinWidth >= 0 && termWidth < c.MinWidth {
		return false
	}
	if c.MaxWidth >= 0 && termWidth > c.MaxWidth {
		return false
	}

	// Height boundaries
	if c.MinHeight >= 0 && termHeight < c.MinHeight {
		return false
	}
	if c.MaxHeight >= 0 && termHeight > c.MaxHeight {
		return false
	}

	// Orientation
	if c.Orientation != "" {
		if c.Orientation == "landscape" && termWidth < termHeight {
			return false
		}
		if c.Orientation == "portrait" && termWidth >= termHeight {
			return false
		}
	}

	// Color scheme
	if c.ColorScheme != "" {
		if c.ColorScheme == "dark" && !isDark {
			return false
		}
		if c.ColorScheme == "light" && isDark {
			return false
		}
	}

	return true
}

// Merge creates a combined MediaCondition from parent and child conditions (logical AND).
func (m *MediaCondition) Merge(child *MediaCondition) *MediaCondition {
	if m == nil {
		return child
	}
	if child == nil {
		return m
	}

	var combined []MediaClause
	for _, p := range m.Clauses {
		for _, c := range child.Clauses {
			merged := MediaClause{
				MediaType:   c.MediaType,
				Negate:      c.Negate,
				MinWidth:    maxInt(p.MinWidth, c.MinWidth),
				MaxWidth:    minNonNeg(p.MaxWidth, c.MaxWidth),
				MinHeight:   maxInt(p.MinHeight, c.MinHeight),
				MaxHeight:   minNonNeg(p.MaxHeight, c.MaxHeight),
				Orientation: c.Orientation,
				ColorScheme: c.ColorScheme,
			}
			if merged.Orientation == "" {
				merged.Orientation = p.Orientation
			}
			if merged.ColorScheme == "" {
				merged.ColorScheme = p.ColorScheme
			}
			combined = append(combined, merged)
		}
	}

	return &MediaCondition{Clauses: combined}
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minNonNeg(a, b int) int {
	if a < 0 {
		return b
	}
	if b < 0 {
		return a
	}
	if a < b {
		return a
	}
	return b
}
