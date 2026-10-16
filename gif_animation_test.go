package termstrap

import (
	"image"
	"image/color"
	"image/gif"
	"os"
	"path/filepath"
	"testing"
)

// createMultiFrameGIF generates a temporary animated GIF with N distinct colored frames.
func createMultiFrameGIF(t *testing.T, numFrames int) string {
	t.Helper()

	palette := color.Palette{
		color.RGBA{0, 0, 0, 255},       // Black
		color.RGBA{255, 0, 0, 255},     // Red
		color.RGBA{0, 255, 0, 255},     // Green
		color.RGBA{0, 0, 255, 255},     // Blue
	}

	g := &gif.GIF{
		Image: make([]*image.Paletted, numFrames),
		Delay: make([]int, numFrames),
	}

	for i := range numFrames {
		img := image.NewPaletted(image.Rect(0, 0, 32, 32), palette)
		// Paint with color index corresponding to frame
		colIdx := uint8((i % (len(palette) - 1)) + 1)
		for y := range 32 {
			for x := range 32 {
				img.SetColorIndex(x, y, colIdx)
			}
		}
		g.Image[i] = img
		g.Delay[i] = 10
	}

	tmpDir := t.TempDir()
	gifPath := filepath.Join(tmpDir, "test_anim.gif")
	f, err := os.Create(gifPath)
	if err != nil {
		t.Fatalf("failed to create temp gif file: %v", err)
	}
	defer f.Close()

	if err := gif.EncodeAll(f, g); err != nil {
		t.Fatalf("failed to encode gif: %v", err)
	}

	return gifPath
}

func TestAnimatedGIF_StepFrameSelection(t *testing.T) {
	numFrames := 3
	gifPath := createMultiFrameGIF(t, numFrames)

	// Step 0 -> frame 0 (Red)
	m0 := New("<img src=\""+gifPath+"\" />", WithStep(0), WithWidth(40))
	img0, err := m0.loadImage(gifPath)
	if err != nil {
		t.Fatalf("loadImage step 0 failed: %v", err)
	}

	// Step 1 -> frame 1 (Green)
	m1 := New("<img src=\""+gifPath+"\" />", WithStep(1), WithWidth(40))
	img1, err := m1.loadImage(gifPath)
	if err != nil {
		t.Fatalf("loadImage step 1 failed: %v", err)
	}

	// Step 3 -> loops back to frame 0
	m3 := New("<img src=\""+gifPath+"\" />", WithStep(3), WithWidth(40))
	img3, err := m3.loadImage(gifPath)
	if err != nil {
		t.Fatalf("loadImage step 3 failed: %v", err)
	}

	// Verify frames have distinct colors
	c0 := img0.At(0, 0)
	c1 := img1.At(0, 0)
	c3 := img3.At(0, 0)

	r0, g0, b0, _ := c0.RGBA()
	r1, g1, b1, _ := c1.RGBA()
	r3, g3, b3, _ := c3.RGBA()

	if r0 == r1 && g0 == g1 && b0 == b1 {
		t.Errorf("frame 0 and frame 1 should have different colors, got c0=%v c1=%v", c0, c1)
	}

	if r0 != r3 || g0 != g3 || b0 != b3 {
		t.Errorf("frame 3 should wrap to frame 0, got c0=%v c3=%v", c0, c3)
	}

	// Test full Render() does not fail
	out0, err := m0.Render()
	if err != nil {
		t.Fatalf("m0.Render() failed: %v", err)
	}
	out1, err := m1.Render()
	if err != nil {
		t.Fatalf("m1.Render() failed: %v", err)
	}

	if out0 == "" || out1 == "" {
		t.Errorf("rendered output should not be empty")
	}
	if out0 == out1 {
		t.Errorf("rendered outputs for different frames should differ")
	}
}
