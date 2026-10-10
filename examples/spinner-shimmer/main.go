// Example: spinner-shimmer — Demonstrates animated "Working..." shimmering text
// and Braille spinner in Termstrap, inspired by Oh My Pi / coding agent indicators.
//
// Usage:
//
//	go run ./examples/spinner-shimmer/
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/go-scripts/termstrap"
	"golang.org/x/term"
)

func main() {
	width := 80
	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && w > 0 {
		width = w
	}

	fmt.Println("Showing animated 'Working...' spinner indicator (Press Ctrl+C to stop)...")
	fmt.Println()

	// Hide cursor during animation
	fmt.Print("\x1b[?25l")
	defer fmt.Print("\x1b[?25h\n")

	for step := range 50 {
		spinner := termstrap.Spinner(step)
		shimmerText := termstrap.Shimmer("Working...", step, termstrap.ShimmerOptions{
			BaseColor:      "#6272a4",
			HighlightColor: "#50fa7b", // Glow green/white
			Bold:           true,
		})

		// Declarative HTML syntax: <spinner> and <shimmer> tags or .spinner / .shimmer classes
		html := fmt.Sprintf(`<div class="border rounded p-1" style="border-color: #bd93f9; width: 65ch;">
  <div><span class="text-warning">%s</span> %s</div>
  <div class="mt-1"><spinner class="text-info"></spinner> <shimmer class="text-primary">Declarative &lt;shimmer&gt; wave analyzing code...</shimmer></div>
  <div class="mt-1 text-muted"><span class="spinner text-success"></span> <span class="shimmer">Class-based .spinner & .shimmer in action...</span></div>
</div>`, spinner, shimmerText)

		out, err := termstrap.New(html, termstrap.WithWidth(width), termstrap.WithStep(step)).Render()
		if err != nil {
			break
		}
		// Clear screen region and reprint
		lines := len(termstrap.SpinnerFrames)
		_ = lines
		fmt.Printf("\r\x1b[2K%s\x1b[3A", out)
		time.Sleep(70 * time.Millisecond)
	}

	fmt.Printf("\n\n\n\x1b[32m✔ Done!\x1b[0m\n")
}
