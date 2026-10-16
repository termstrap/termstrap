// Example: animated_gif — Demonstrates rendering multi-frame animated GIFs using WithStep.
//
// Usage:
//
//	go run ./examples/image/animated_gif/
package main

import (
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"os"
	"path/filepath"
	"time"

	"github.com/termstrap/termstrap"
	"golang.org/x/term"
)

func main() {
	width := 80
	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && w > 0 {
		width = w
	}

	tmpDir, err := os.MkdirTemp("", "termstrap-gif-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create temp dir: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)

	gifPath := filepath.Join(tmpDir, "spinner.gif")
	createAnimatedSpinnerGIF(gifPath, 40, 20, 8)

	content := `<div class="p-2 border rounded">
  <h1>Animated GIF Demo</h1>
  <p>Step-based animation driving GIF frame selection via <code>WithStep(step)</code>.</p>
  <div class="text-center">
    <img src="spinner.gif" alt="Animated GIF" />
  </div>
  <p class="text-muted text-center">Press Ctrl+C to exit</p>
</div>`

	fmt.Print("\033[2J") // Clear screen once

	// Run animation loop for 24 steps in demo
	ticker := time.NewTicker(120 * time.Millisecond)
	defer ticker.Stop()

	step := 0
	for range ticker.C {
		m := termstrap.New(
			content,
			termstrap.WithWidth(width),
			termstrap.WithRootPath(tmpDir),
			termstrap.WithStep(step),
		)
		out, err := m.Render()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Render error: %v\n", err)
			return
		}

		// Move cursor to top-left and redraw
		fmt.Print("\033[H" + out)

		step++
		if step >= 24 {
			break
		}
	}
	fmt.Println()
}

// createAnimatedSpinnerGIF generates a synthetic multi-frame GIF showing a rotating dot.
func createAnimatedSpinnerGIF(path string, width, height, numFrames int) {
	palette := color.Palette{
		color.RGBA{20, 20, 25, 255},   // Background
		color.RGBA{80, 120, 200, 255}, // Inactive track
		color.RGBA{255, 180, 50, 255}, // Active pointer
	}

	g := &gif.GIF{
		Image: make([]*image.Paletted, numFrames),
		Delay: make([]int, numFrames),
	}

	for f := range numFrames {
		img := image.NewPaletted(image.Rect(0, 0, width, height), palette)
		// Background fill
		for y := range height {
			for x := range width {
				img.SetColorIndex(x, y, 0)
			}
		}

		// Draw track line in center
		midY := height / 2
		margin := 4
		trackLen := width - 2*margin
		for x := margin; x < width-margin; x++ {
			img.SetColorIndex(x, midY, 1)
		}

		// Active dot moving across track
		activeX := margin + (f * trackLen / numFrames)
		for dy := -2; dy <= 2; dy++ {
			for dx := -2; dx <= 2; dx++ {
				px := activeX + dx
				py := midY + dy
				if px >= 0 && px < width && py >= 0 && py < height {
					img.SetColorIndex(px, py, 2)
				}
			}
		}

		g.Image[f] = img
		g.Delay[f] = 10
	}

	f, err := os.Create(path)
	if err != nil {
		return
	}
	defer f.Close()

	_ = gif.EncodeAll(f, g)
}
