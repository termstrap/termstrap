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
