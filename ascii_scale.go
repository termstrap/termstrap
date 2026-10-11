package termstrap

import (
	"math"
	"strings"
)

// scaleSextantRunes maps a 6-bit sub-pixel mask (0..63) to its Unicode block character.
// Sub-pixel bit positions in the 2x3 block:
//
//	bit 0 (1): top-left      bit 1 (2): top-right
//	bit 2 (4): mid-left      bit 3 (8): mid-right
//	bit 4 (16): btm-left     bit 5 (32): btm-right
var scaleSextantRunes = [64]rune{
	0: ' ', 1: '\U0001FB00', 2: '\U0001FB01', 3: '\U0001FB02',
	4: '\U0001FB03', 5: '\U0001FB04', 6: '\U0001FB05', 7: '\U0001FB06',
	8: '\U0001FB07', 9: '\U0001FB08', 10: '\U0001FB09', 11: '\U0001FB0A',
	12: '\U0001FB0B', 13: '\U0001FB0C', 14: '\U0001FB0D', 15: '\U0001FB0E',
	16: '\U0001FB0F', 17: '\U0001FB10', 18: '\U0001FB11', 19: '\U0001FB12',
	20: '\U0001FB13', 21: '\u258C', 22: '\U0001FB14', 23: '\U0001FB15',
	24: '\U0001FB16', 25: '\U0001FB17', 26: '\U0001FB18', 27: '\U0001FB19',
	28: '\U0001FB1A', 29: '\U0001FB1B', 30: '\U0001FB1C', 31: '\U0001FB1D',
	32: '\U0001FB1E', 33: '\U0001FB1F', 34: '\U0001FB20', 35: '\U0001FB21',
	36: '\U0001FB22', 37: '\U0001FB23', 38: '\U0001FB24', 39: '\U0001FB25',
	40: '\U0001FB26', 41: '\U0001FB27', 42: '\u2590', 43: '\U0001FB28',
	44: '\U0001FB29', 45: '\U0001FB2A', 46: '\U0001FB2B', 47: '\U0001FB2C',
	48: '\U0001FB2D', 49: '\U0001FB2E', 50: '\U0001FB2F', 51: '\U0001FB30',
	52: '\U0001FB31', 53: '\U0001FB32', 54: '\U0001FB33', 55: '\U0001FB34',
	56: '\U0001FB35', 57: '\U0001FB36', 58: '\U0001FB37', 59: '\U0001FB38',
	60: '\U0001FB39', 61: '\U0001FB3A', 62: '\U0001FB3B', 63: '\u2588',
}

// inkThreshold is the sub-pixel coverage fraction above which a sub-pixel is "on".
// Lower values preserve thin strokes at small sizes.
const inkThreshold = 0.42

// scaleAsciiArt downsamples monochrome FIGlet art so its widest line fits within
// targetW columns, preserving the silhouette of the chosen font by re-sampling it
// as a 1-bit image into Unicode 2x3 sextant block characters.
//
// If the art already fits (or targetW <= 0), the original lines are returned
// unchanged so the native glyphs are kept whenever possible.
func scaleAsciiArt(lines []string, targetW int) []string {
	if targetW <= 0 {
		return lines
	}

	// Trim leading/trailing all-blank lines to tighten the vertical silhouette.
	start, end := 0, len(lines)
	for start < end && strings.TrimSpace(lines[start]) == "" {
		start++
	}
	for end > start && strings.TrimSpace(lines[end-1]) == "" {
		end--
	}
	lines = lines[start:end]
	if len(lines) == 0 {
		return lines
	}

	// Build the source ink grid (true = ink, false = background).
	srcH := len(lines)
	srcW := 0
	rows := make([][]rune, srcH)
	for y, l := range lines {
		r := []rune(l)
		rows[y] = r
		if len(r) > srcW {
			srcW = len(r)
		}
	}
	if srcW == 0 {
		return lines
	}
	if srcW <= targetW {
		// Already fits: keep the native, full-resolution glyphs.
		return lines
	}

	grid := make([][]bool, srcH)
	for y := range srcH {
		g := make([]bool, srcW)
		for x, ch := range rows[y] {
			g[x] = ch != ' ' && ch != '\t'
		}
		grid[y] = g
	}

	// Output cell grid: targetW cells wide. Each cell holds 2x3 sub-pixels.
	// Row count is aspect-preserving: R = round(targetW * srcH / srcW).
	outCols := targetW
	outRows := int(math.Round(float64(targetW) * float64(srcH) / float64(srcW)))
	if outRows < 1 {
		outRows = 1
	}
	subW := outCols * 2
	subH := outRows * 3

	sxW := float64(srcW) / float64(subW)
	sxH := float64(srcH) / float64(subH)

	on := func(cx, cy int) bool {
		x0 := float64(cx) * sxW
		x1 := float64(cx+1) * sxW
		y0 := float64(cy) * sxH
		y1 := float64(cy+1) * sxH
		var area, ink float64
		for yy := int(math.Floor(y0)); yy < int(math.Ceil(y1)); yy++ {
			fy0 := math.Max(y0, float64(yy))
			fy1 := math.Min(y1, float64(yy+1))
			h := fy1 - fy0
			if h <= 0 || yy < 0 || yy >= srcH {
				continue
			}
			for xx := int(math.Floor(x0)); xx < int(math.Ceil(x1)); xx++ {
				fx0 := math.Max(x0, float64(xx))
				fx1 := math.Min(x1, float64(xx+1))
				w := fx1 - fx0
				if w <= 0 || xx < 0 || xx >= srcW {
					continue
				}
				a := w * h
				area += a
				if grid[yy][xx] {
					ink += a
				}
			}
		}
		if area <= 0 {
			return false
		}
		return ink/area >= inkThreshold
	}

	out := make([]string, outRows)
	// Sub-pixel (cx,cy) within a cell -> bit value.
	bitFor := func(cx, cy int) int { return 1 << (cy*2 + cx) }
	var sb strings.Builder
	for r := range outRows {
		sb.Reset()
		for c := range outCols {
			mask := 0
			for cy := range 3 {
				for cx := range 2 {
					if on(c*2+cx, r*3+cy) {
						mask |= bitFor(cx, cy)
					}
				}
			}
			sb.WriteRune(scaleSextantRunes[mask])
		}
		out[r] = strings.TrimRight(sb.String(), " ")
	}
	return out
}
