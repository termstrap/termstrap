// Example: media-queries — Demonstrates responsive terminal rendering
// using standard CSS @media queries (min-width, max-width, ch units).
//
// Usage:
//
//	go run ./examples/media-queries/
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-scripts/termstrap"
	"golang.org/x/term"
)

func main() {
	// 1. Locate files
	baseDir := filepath.Join("examples", "media-queries")
	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		baseDir = "."
	}

	htmlBytes, err := os.ReadFile(filepath.Join(baseDir, "index.html"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading index.html: %v\n", err)
		os.Exit(1)
	}

	cssBytes, err := os.ReadFile(filepath.Join(baseDir, "style.css"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading style.css: %v\n", err)
		os.Exit(1)
	}

	// 2. Multi-width comparison scenarios
	scenarios := []struct {
		name  string
		width int
		desc  string
	}{
		{
			name:  "Narrow Terminal / Mobile (< 70 cols)",
			width: 55,
			desc:  "Base CSS active: Blue banner, sidebar hidden, compact view.",
		},
		{
			name:  "Medium Terminal / Tablet (70 - 104 cols)",
			width: 80,
			desc:  "@media (min-width: 70ch) active: Purple banner, sidebar revealed.",
		},
		{
			name:  "Wide Terminal / Desktop (>= 105 cols)",
			width: 115,
			desc:  "@media (min-width: 105ch) active: Green banner, full telemetry notice.",
		},
	}

	for _, sc := range scenarios {
		fmt.Printf("================================================================================\n")
		fmt.Printf(" Scenario: %s (Simulated width: %d cols)\n", sc.name, sc.width)
		fmt.Printf(" %s\n", sc.desc)
		fmt.Printf("================================================================================\n\n")

		m := termstrap.New(
			string(htmlBytes),
			termstrap.WithWidth(sc.width),
			termstrap.WithStylesheets(string(cssBytes)),
		)

		output, err := m.Render()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Render error: %v\n", err)
			continue
		}

		fmt.Println(output)
		fmt.Println()
	}

	// 3. Current Live Terminal Window
	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && w > 0 {
		fmt.Printf("================================================================================\n")
		fmt.Printf(" Live Terminal Render (Detected width: %d cols)\n", w)
		fmt.Printf("================================================================================\n\n")

		mLive := termstrap.New(
			string(htmlBytes),
			termstrap.WithWidth(w),
			termstrap.WithStylesheets(string(cssBytes)),
		)
		outLive, err := mLive.Render()
		if err == nil {
			fmt.Println(outLive)
		}
	}
}
