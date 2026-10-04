// Example: custom-css — Demonstrates using Termstrap with a custom external
// CSS stylesheet (.css file) and custom CSS classes/rules instead of Bootstrap classes.
//
// Usage:
//
//	go run ./examples/custom-css/
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-scripts/termstrap"
	termimage "github.com/go-scripts/termstrap/image"
	"golang.org/x/term"
)

func main() {
	// Detect current terminal width or fallback to 80 cols
	width := 80
	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && w > 0 {
		width = w
	}

	// 1. Read the custom CSS stylesheet
	cssPath := filepath.Join("examples", "custom-css", "style.css")
	customCSS, err := os.ReadFile(cssPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read CSS file %s: %v\n", cssPath, err)
		os.Exit(1)
	}

	// 2. Define HTML markup using our custom CSS classes (no Bootstrap classes)
	content, err := os.ReadFile(filepath.Join("examples", "custom-css", "index.html"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read HTML file: %v\n", err)
		os.Exit(1)
	}

	// 3. Initialize Termstrap with the custom stylesheet
	m := termstrap.New(
		string(content),
		termstrap.WithWidth(width),
		termstrap.WithStylesheets(string(customCSS)),
		termstrap.WithImageRenderer(termimage.NewRenderer(termimage.WithProtocol(termimage.HalfBlock))),
	)

	// 4. Render to ANSI output
	output, err := m.Render()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Render error: %v\n", err)
		os.Exit(1)
	}

	fmt.Print(output)
}
