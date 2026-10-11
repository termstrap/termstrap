package termstrap

import (
	"strings"
	"sync"

	"github.com/common-nighthawk/go-figure"
)

var (
	availableFontsOnce sync.Once
	availableFontsMap  map[string]string // normalized lowercase -> exact asset font name
	allFontNames       []string
)

func initFonts() {
	availableFontsOnce.Do(func() {
		availableFontsMap = make(map[string]string)
		for _, name := range figure.AssetNames() {
			// AssetNames are formatted like "fonts/standard.flf"
			if strings.HasPrefix(name, "fonts/") && strings.HasSuffix(name, ".flf") {
				fontName := strings.TrimSuffix(strings.TrimPrefix(name, "fonts/"), ".flf")
				availableFontsMap[strings.ToLower(fontName)] = fontName
				allFontNames = append(allFontNames, fontName)
			}
		}
	})
}

// AvailableAsciiFonts returns the list of all embedded FIGlet fonts (over 140 fonts).
func AvailableAsciiFonts() []string {
	initFonts()
	return allFontNames
}

// resolveFontName normalizes a user font name and finds the exact matching FIGlet font name,
// falling back to "standard" if not found or empty.
func resolveFontName(name string) string {
	initFonts()
	name = strings.TrimSpace(strings.ToLower(name))
	if name == "" {
		return "standard"
	}
	if exact, ok := availableFontsMap[name]; ok {
		return exact
	}
	return "standard"
}

// defaultFontHierarchy lists crisp FIGlet fonts ordered from largest to smallest.
// They share a bold/solid aesthetic so that a responsive step-down preserves the
// overall lettering style while keeping every glyph sharp (no pixelation).
var defaultFontHierarchy = []string{
	"doh",      // ~123 cols — giant solid block
	"univers",  // ~78  cols — bold solid block
	"block",    // ~69  cols — block
	"colossal", // ~59  cols — bold solid block
	"big",      // ~49  cols — rounded bold
	"standard", // ~48  cols — classic
	"small",    // ~42  cols — compact
	"mini",     // ~30  cols — tiny
}

// resolveFontCandidates parses a font spec (e.g. "doh,doom,slant,small" or "auto" or "doh")
func resolveFontCandidates(spec string) []string {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return []string{"standard"}
	}
	if strings.EqualFold(spec, "auto") {
		return defaultFontHierarchy
	}
	parts := strings.Split(spec, ",")
	var candidates []string
	for _, p := range parts {
		f := strings.TrimSpace(p)
		if strings.EqualFold(f, "auto") {
			candidates = append(candidates, defaultFontHierarchy...)
		} else if f != "" {
			candidates = append(candidates, f)
		}
	}
	if len(candidates) == 0 {
		return []string{"standard"}
	}
	return candidates
}

// newFigureSafe safely constructs a figure without panicking on invalid font assets.
func newFigureSafe(phrase, fontName string, strict bool) []string {
	exact := resolveFontName(fontName)
	var lines []string
	func() {
		defer func() {
			if r := recover(); r != nil {
				// Fallback to standard font on unexpected asset panic
				fig := figure.NewFigure(phrase, "standard", false)
				lines = fig.Slicify()
			}
		}()
		fig := figure.NewFigure(phrase, exact, strict)
		lines = fig.Slicify()
	}()
	return lines
}

// newFigureResponsive chooses the largest font from the candidates that fits within maxCols.
// If none fits completely, it uses the smallest candidate.
func newFigureResponsive(phrase, fontSpec string, strict bool, maxCols int) []string {
	candidates := resolveFontCandidates(fontSpec)
	if maxCols <= 0 || len(candidates) == 1 {
		return newFigureSafe(phrase, candidates[0], strict)
	}

	for _, cand := range candidates {
		lines := newFigureSafe(phrase, cand, strict)
		w := 0
		for _, l := range lines {
			if len(l) > w {
				w = len(l)
			}
		}
		if w <= maxCols {
			return lines
		}
	}

	// None fit entirely: return the last candidate (the smallest)
	return newFigureSafe(phrase, candidates[len(candidates)-1], strict)
}

// asciiLines produces the final banner lines for an <ascii> node, honoring the
// fit mode. With fit="scale" the first chosen font is kept and its rendered art
// is downscaled (silhouette preserved) to fit maxCols. Otherwise the responsive
// font-stack fallback is used (the largest font in the list that fits).
func asciiLines(phrase, fontSpec, fit string, strict bool, maxCols int) []string {
	if fit == "scale" {
		candidates := resolveFontCandidates(fontSpec)
		lines := newFigureSafe(phrase, candidates[0], strict)
		return scaleAsciiArt(lines, maxCols)
	}
	return newFigureResponsive(phrase, fontSpec, strict, maxCols)
}
