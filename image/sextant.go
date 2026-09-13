package image

import (
	"fmt"
	"image"
	"strings"
)

// sextantRunes maps a 6-bit mask (0..63) to its corresponding Unicode character.
// Sub-pixel positions in the 2x3 block:
// bit 0 (1): top-left (x=0, y=0)
// bit 1 (2): top-right (x=1, y=0)
// bit 2 (4): mid-left (x=0, y=1)
// bit 3 (8): mid-right (x=1, y=1)
// bit 4 (16): btm-left (x=0, y=2)
// bit 5 (32): btm-right (x=1, y=2)
var sextantRunes = [64]rune{
	0:  ' ',         // 000000
	1:  '\U0001FB00', // 1
	2:  '\U0001FB01', // 2
	3:  '\U0001FB02', // 12
	4:  '\U0001FB03', // 3
	5:  '\U0001FB04', // 13
	6:  '\U0001FB05', // 23
	7:  '\U0001FB06', // 123
	8:  '\U0001FB07', // 4
	9:  '\U0001FB08', // 14
	10: '\U0001FB09', // 24
	11: '\U0001FB0A', // 124
	12: '\U0001FB0B', // 34
	13: '\U0001FB0C', // 134
	14: '\U0001FB0D', // 234
	15: '\U0001FB0E', // 1234
	16: '\U0001FB0F', // 5
	17: '\U0001FB10', // 15
	18: '\U0001FB11', // 25
	19: '\U0001FB12', // 125
	20: '\U0001FB13', // 35
	21: '\u258C',     // 135 (LEFT HALF BLOCK ▌)
	22: '\U0001FB14', // 235
	23: '\U0001FB15', // 1235
	24: '\U0001FB16', // 45
	25: '\U0001FB17', // 145
	26: '\U0001FB18', // 245
	27: '\U0001FB19', // 1245
	28: '\U0001FB1A', // 345
	29: '\U0001FB1B', // 1345
	30: '\U0001FB1C', // 2345
	31: '\U0001FB1D', // 12345
	32: '\U0001FB1E', // 6
	33: '\U0001FB1F', // 16
	34: '\U0001FB20', // 26
	35: '\U0001FB21', // 126
	36: '\U0001FB22', // 36
	37: '\U0001FB23', // 136
	38: '\U0001FB24', // 236
	39: '\U0001FB25', // 1236
	40: '\U0001FB26', // 46
	41: '\U0001FB27', // 146
	42: '\u2590',     // 246 (RIGHT HALF BLOCK ▐)
	43: '\U0001FB28', // 1246
	44: '\U0001FB29', // 346
	45: '\U0001FB2A', // 1346
	46: '\U0001FB2B', // 2346
	47: '\U0001FB2C', // 12346
	48: '\U0001FB2D', // 56
	49: '\U0001FB2E', // 156
	50: '\U0001FB2F', // 256
	51: '\U0001FB30', // 1256
	52: '\U0001FB31', // 356
	53: '\U0001FB32', // 1356
	54: '\U0001FB33', // 2356
	55: '\U0001FB34', // 12356
	56: '\U0001FB35', // 456
	57: '\U0001FB36', // 1456
	58: '\U0001FB37', // 2456
	59: '\U0001FB38', // 12456
	60: '\U0001FB39', // 3456
	61: '\U0001FB3A', // 13456
	62: '\U0001FB3B', // 23456
	63: '\u2588',     // 123456 (FULL BLOCK █)
}

// sextantRenderer renders images using Unicode 13.0 Block Sextant characters (🬀..🬿, ▌, ▐, █)
// providing 2x3 (6 sub-pixels per character cell) high-resolution ANSI rendering.
type sextantRenderer struct {
	colorMode ColorMode
	optimize  bool
	dither    bool
}

// NewSextantRenderer creates a new sextant renderer with the specified options.
func NewSextantRenderer(colorMode ColorMode, optimize bool, dither ...bool) Renderer {
	d := false
	if len(dither) > 0 {
		d = dither[0]
	}
	return &sextantRenderer{
		colorMode: colorMode,
		optimize:  optimize,
		dither:    d,
	}
}

func (r *sextantRenderer) Protocol() Protocol { return Sextant }

// Render converts the image to colored block sextant characters at the given column width.
func (r *sextantRenderer) Render(img image.Image, width int) (string, error) {
	if img == nil {
		return "", fmt.Errorf("sextant: nil image")
	}
	if width <= 0 {
		return "", fmt.Errorf("sextant: invalid width %d", width)
	}

	// Each column cell holds 2 horizontal pixels
	scaled := ResizeToWidth(img, width*2)
	if r.dither && r.colorMode != ColorModeTrueColor {
		scaled = FloydSteinbergDither(scaled, r.colorMode)
	}
	return r.renderScaled(scaled)
}

// RenderConstrained fits the image within both column width and row height.
func (r *sextantRenderer) RenderConstrained(img image.Image, width, height int) (string, error) {
	if img == nil {
		return "", fmt.Errorf("sextant: nil image")
	}
	if width <= 0 || height <= 0 {
		return "", fmt.Errorf("sextant: invalid dimensions %dx%d", width, height)
	}

	// Each character cell holds 2 horizontal x 3 vertical pixels
	scaled := ResizeToFit(img, width*2, height*3)
	if r.dither && r.colorMode != ColorModeTrueColor {
		scaled = FloydSteinbergDither(scaled, r.colorMode)
	}
	return r.renderScaled(scaled)
}

func (r *sextantRenderer) renderScaled(scaled image.Image) (string, error) {
	bounds := scaled.Bounds()
	imgW := bounds.Dx()
	imgH := bounds.Dy()
	if imgW == 0 || imgH == 0 {
		return "", fmt.Errorf("sextant: rendered empty output")
	}

	var buf strings.Builder
	buf.Grow((imgW/2 + 1) * (imgH/3 + 1) * 20)

	type pixelInfo struct {
		r, g, b uint8
		solid   bool
	}

	// Step in blocks of 2 horizontal pixels and 3 vertical pixels
	for y := bounds.Min.Y; y < bounds.Max.Y; y += 3 {
		var lastFG, lastBG string
		hasFG, hasBG := false, false

		for x := bounds.Min.X; x < bounds.Max.X; x += 2 {
			var px [6]pixelInfo
			solidCount := 0

			// 6 sub-pixel positions:
			// 0: (x, y)     1: (x+1, y)
			// 2: (x, y+1)   3: (x+1, y+1)
			// 4: (x, y+2)   5: (x+1, y+2)
			offsets := [6][2]int{
				{0, 0}, {1, 0},
				{0, 1}, {1, 1},
				{0, 2}, {1, 2},
			}

			for i, off := range offsets {
				pxX := x + off[0]
				pxY := y + off[1]
				if pxX < bounds.Max.X && pxY < bounds.Max.Y {
					r16, g16, b16, a16 := scaled.At(pxX, pxY).RGBA()
					alpha := uint8(a16 >> 8)
					if alpha >= 128 {
						px[i] = pixelInfo{
							r:     uint8(r16 >> 8),
							g:     uint8(g16 >> 8),
							b:     uint8(b16 >> 8),
							solid: true,
						}
						solidCount++
					}
				}
			}

			if solidCount == 0 {
				// Completely transparent cell
				if r.optimize && hasBG {
					buf.WriteString("\x1b[49m")
					hasBG = false
					lastBG = ""
				}
				buf.WriteString(" ")
				continue
			}

			// Calculate average color and luminance
			var totalR, totalG, totalB, totalLuma int
			for _, p := range px {
				if p.solid {
					totalR += int(p.r)
					totalG += int(p.g)
					totalB += int(p.b)
					totalLuma += int(p.r)*299 + int(p.g)*587 + int(p.b)*114
				}
			}
			avgLuma := totalLuma / solidCount

			// Partition into two clusters (fg = brighter, bg = darker)
			var fgR, fgG, fgB, fgCount int
			var bgR, bgG, bgB, bgCount int

			for _, p := range px {
				if p.solid {
					luma := int(p.r)*299 + int(p.g)*587 + int(p.b)*114
					if luma >= avgLuma {
						fgR += int(p.r)
						fgG += int(p.g)
						fgB += int(p.b)
						fgCount++
					} else {
						bgR += int(p.r)
						bgG += int(p.g)
						bgB += int(p.b)
						bgCount++
					}
				}
			}

			var fgColor, bgColor [3]uint8
			if fgCount > 0 {
				fgColor = [3]uint8{uint8(fgR / fgCount), uint8(fgG / fgCount), uint8(fgB / fgCount)}
			} else {
				fgColor = [3]uint8{uint8(totalR / solidCount), uint8(totalG / solidCount), uint8(totalB / solidCount)}
			}

			if bgCount > 0 {
				bgColor = [3]uint8{uint8(bgR / bgCount), uint8(bgG / bgCount), uint8(bgB / bgCount)}
			} else {
				bgColor = fgColor
			}

			// Determine 6-bit mask
			mask := 0
			for i, p := range px {
				if p.solid {
					// Compare distance to FG vs BG
					dFG := distSq(p.r, p.g, p.b, fgColor[0], fgColor[1], fgColor[2])
					dBG := distSq(p.r, p.g, p.b, bgColor[0], bgColor[1], bgColor[2])
					if dFG <= dBG {
						mask |= 1 << i
					}
				}
			}

			fgCode := r.fgEscape(fgColor[0], fgColor[1], fgColor[2])
			bgCode := r.bgEscape(bgColor[0], bgColor[1], bgColor[2])

			switch mask {
			case 0:
				if r.optimize {
					if !hasBG || bgCode != lastBG {
						buf.WriteString(bgCode)
						lastBG = bgCode
						hasBG = true
					}
				} else {
					buf.WriteString(bgCode)
				}
				buf.WriteString(" ")
			case 63:
				if r.optimize {
					if !hasFG || fgCode != lastFG {
						buf.WriteString(fgCode)
						lastFG = fgCode
						hasFG = true
					}
					if hasBG {
						buf.WriteString("\x1b[49m")
						hasBG = false
						lastBG = ""
					}
				} else {
					buf.WriteString(fgCode)
					buf.WriteString("\x1b[49m")
				}
				buf.WriteString("█")
			default:
				if r.optimize {
					if !hasFG || fgCode != lastFG {
						buf.WriteString(fgCode)
						lastFG = fgCode
						hasFG = true
					}
					if !hasBG || bgCode != lastBG {
						buf.WriteString(bgCode)
						lastBG = bgCode
						hasBG = true
					}
				} else {
					buf.WriteString(fgCode)
					buf.WriteString(bgCode)
				}
				buf.WriteRune(sextantRunes[mask])
			}
		}

		// Reset formatting at end of each row
		buf.WriteString("\x1b[0m")
		if y+3 < bounds.Max.Y {
			buf.WriteString("\n")
		}
	}

	result := buf.String()
	if strings.TrimSpace(result) == "" {
		return "", fmt.Errorf("sextant: rendered empty output")
	}
	return result, nil
}

func distSq(r1, g1, b1, r2, g2, b2 uint8) int {
	dr := int(r1) - int(r2)
	dg := int(g1) - int(g2)
	db := int(b1) - int(b2)
	return dr*dr + dg*dg + db*db
}

func (r *sextantRenderer) fgEscape(cr, cg, cb uint8) string {
	switch r.colorMode {
	case ColorMode256:
		idx := RGBTo256(cr, cg, cb)
		return fmt.Sprintf("\x1b[38;5;%dm", idx)
	case ColorMode16:
		idx := RGBTo16(cr, cg, cb)
		if idx < 8 {
			return fmt.Sprintf("\x1b[%dm", 30+idx)
		}
		return fmt.Sprintf("\x1b[%dm", 90+(idx-8))
	default: // TrueColor
		return fmt.Sprintf("\x1b[38;2;%d;%d;%dm", cr, cg, cb)
	}
}

func (r *sextantRenderer) bgEscape(cr, cg, cb uint8) string {
	switch r.colorMode {
	case ColorMode256:
		idx := RGBTo256(cr, cg, cb)
		return fmt.Sprintf("\x1b[48;5;%dm", idx)
	case ColorMode16:
		idx := RGBTo16(cr, cg, cb)
		if idx < 8 {
			return fmt.Sprintf("\x1b[%dm", 40+idx)
		}
		return fmt.Sprintf("\x1b[%dm", 100+(idx-8))
	default: // TrueColor
		return fmt.Sprintf("\x1b[48;2;%d;%d;%dm", cr, cg, cb)
	}
}
