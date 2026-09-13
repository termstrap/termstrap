package image

import (
	"image"
	"image/color"
)

type rgbErr struct {
	r, g, b float64
}

// FloydSteinbergDither applies Floyd-Steinberg error diffusion dithering to an image
// according to the target ColorMode (ColorMode256 or ColorMode16).
// For ColorModeTrueColor, returns a copy of the image.
func FloydSteinbergDither(img image.Image, mode ColorMode) *image.RGBA {
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	dst := image.NewRGBA(image.Rect(0, 0, w, h))

	if mode == ColorModeTrueColor {
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				dst.Set(x, y, img.At(bounds.Min.X+x, bounds.Min.Y+y))
			}
		}
		return dst
	}

	errBuf := make([][]rgbErr, h+2)
	for i := range errBuf {
		errBuf[i] = make([]rgbErr, w+2)
	}

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			origC := img.At(bounds.Min.X+x, bounds.Min.Y+y)
			r32, g32, b32, a32 := origC.RGBA()
			a := uint8(a32 >> 8)
			if a < 128 {
				dst.SetRGBA(x, y, color.RGBA{0, 0, 0, 0})
				continue
			}

			curR := clamp(float64(r32>>8) + errBuf[y][x+1].r)
			curG := clamp(float64(g32>>8) + errBuf[y][x+1].g)
			curB := clamp(float64(b32>>8) + errBuf[y][x+1].b)

			var quantR, quantG, quantB uint8
			if mode == ColorMode256 {
				idx := RGBTo256(uint8(curR), uint8(curG), uint8(curB))
				quantR, quantG, quantB = Palette256ToRGB(idx)
			} else { // ColorMode16
				idx := RGBTo16(uint8(curR), uint8(curG), uint8(curB))
				quantR, quantG, quantB = Palette16ToRGB(idx)
			}

			dst.SetRGBA(x, y, color.RGBA{R: quantR, G: quantG, B: quantB, A: a})

			errR := curR - float64(quantR)
			errG := curG - float64(quantG)
			errB := curB - float64(quantB)

			// Floyd-Steinberg diffusion coefficients:
			// (x+1, y)   : 7/16
			// (x-1, y+1) : 3/16
			// (x, y+1)   : 5/16
			// (x+1, y+1) : 1/16
			distributeDitherError(errBuf, x+2, y, errR*7.0/16.0, errG*7.0/16.0, errB*7.0/16.0)
			distributeDitherError(errBuf, x, y+1, errR*3.0/16.0, errG*3.0/16.0, errB*3.0/16.0)
			distributeDitherError(errBuf, x+1, y+1, errR*5.0/16.0, errG*5.0/16.0, errB*5.0/16.0)
			distributeDitherError(errBuf, x+2, y+1, errR*1.0/16.0, errG*1.0/16.0, errB*1.0/16.0)
		}
	}

	return dst
}

func distributeDitherError(buf [][]rgbErr, x, y int, er, eg, eb float64) {
	if y < len(buf) && x >= 0 && x < len(buf[y]) {
		buf[y][x].r += er
		buf[y][x].g += eg
		buf[y][x].b += eb
	}
}

func clamp(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return v
}
