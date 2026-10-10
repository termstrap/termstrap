// Example: bootstrap — Demonstrates rendering multiple rich HTML layouts
// using the complete official Bootstrap CSS stylesheet (css/bootstrap.css).
//
// Usage:
//
//	go run ./examples/bootstrap/              # Renders all examples
//	go run ./examples/bootstrap/ dashboard    # System monitoring dashboard
//	go run ./examples/bootstrap/ tables       # Bootstrap styled tables
//	go run ./examples/bootstrap/ alerts       # Contextual alerts and callouts
//	go run ./examples/bootstrap/ media        # Media album detail card
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-scripts/termstrap"
	"golang.org/x/term"
)

func main() {
	width := 100
	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && w > 0 {
		width = w
	}

	// 1. Locate and load the complete official Bootstrap CSS stylesheet
	cssCandidates := []string{
		filepath.Join("css", "bootstrap.css"),
		filepath.Join("..", "..", "css", "bootstrap.css"),
		filepath.Join("..", "css", "bootstrap.css"),
	}

	var bootstrapCSS []byte
	var loadedPath string
	for _, candidate := range cssCandidates {
		data, err := os.ReadFile(candidate)
		if err == nil && len(data) > 0 {
			bootstrapCSS = data
			loadedPath = candidate
			break
		}
	}

	if len(bootstrapCSS) == 0 {
		fmt.Fprintf(os.Stderr, "Error: unable to locate css/bootstrap.css in candidate paths\n")
		os.Exit(1)
	}

	examplesDir := filepath.Join("examples", "bootstrap")
	if _, err := os.Stat(examplesDir); os.IsNotExist(err) {
		examplesDir = "."
	}

	type demo struct {
		id       string
		title    string
		fileName string
	}

	demos := []demo{
		{id: "dashboard", title: "1. System Monitoring Dashboard (KPIs & Grid)", fileName: "dashboard.html"},
		{id: "tables", title: "2. Bootstrap Striped & Colored Tables", fileName: "tables.html"},
		{id: "alerts", title: "3. Contextual Alerts & Callouts", fileName: "alerts.html"},
		{id: "media", title: "4. Media Fiche & Responsive Grid", fileName: "media_card.html"},
	}

	filter := ""
	if len(os.Args) > 1 {
		filter = strings.ToLower(strings.TrimSpace(os.Args[1]))
	}

	fmt.Printf("Loaded Bootstrap CSS (%d bytes from %s)\n", len(bootstrapCSS), loadedPath)
	fmt.Printf("Rendering terminal view at width %d cols...\n\n", width)

	for _, d := range demos {
		if filter != "" && !strings.Contains(d.id, filter) {
			continue
		}

		filePath := filepath.Join(examplesDir, d.fileName)
		htmlData, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", filePath, err)
			continue
		}

		fmt.Printf("================================================================================\n")
		fmt.Printf(" %s\n", d.title)
		fmt.Printf("================================================================================\n\n")

		m := termstrap.New(
			string(htmlData),
			termstrap.WithWidth(width),
			termstrap.WithStylesheets(string(bootstrapCSS)),
		)

		output, err := m.Render()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Render error in %s: %v\n", d.fileName, err)
			continue
		}

		fmt.Println(output)
		fmt.Println()
	}
}
