// Example: headings — Differentiating h1..h6 without font-size.
//
// A terminal has no font-size, so TermStrap differentiates heading levels with
// other levers: a per-level color scale (bright h1 -> muted h6, var
// --ts-heading-1..6), bold (h1..h4), italic (h5/h6), and a Markdown-style
// marker (#..######) injected via hN::before. Everything is CSS, so you can
// push further: this example shows an opt-in "enhanced" variant that adds
// text-transform: uppercase and a border-bottom rule via WithStylesheets.
//
// Usage:
//
//	go run ./examples/headings/
//	go run ./examples/headings/ -w 80 -theme dracula -enhanced
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/termstrap/termstrap"
	"golang.org/x/term"
)

const content = `
<h1>Heading level 1</h1>
<p>Body text under the top-level title.</p>
<h2>Heading level 2</h2>
<p>Body text.</p>
<h3>Heading level 3</h3>
<p>Body text.</p>
<h4>Heading level 4</h4>
<p>Body text.</p>
<h5>Heading level 5</h5>
<p>Body text.</p>
<h6>Heading level 6</h6>
<p>Body text.</p>
`

// enhancedCSS opts into stronger differentiation on top of the defaults:
// uppercase h1/h2 and a horizontal rule under h1/h2 (border-bottom).
const enhancedCSS = `
h1, h2 {
  text-transform: uppercase;
  border-bottom: 1px solid;
  border-color: var(--ts-heading-rule) !important;
}
h6 {
  text-transform: uppercase;
}
`

func main() {
	width := 80
	theme := "bootstrap"
	enhanced := false
	flag.IntVar(&width, "w", width, "terminal width in columns")
	flag.StringVar(&theme, "theme", theme, "visual theme: bootstrap|dracula|tokyonight")
	flag.BoolVar(&enhanced, "enhanced", enhanced, "opt into uppercase + rule differentiation")
	flag.Parse()

	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && w > 0 && !flagPassed("w") {
		width = w
	}

	opts := []termstrap.Option{
		termstrap.WithWidth(width),
		termstrap.WithTheme(termstrap.Theme(theme)),
	}
	if enhanced {
		opts = append(opts, termstrap.WithStylesheets(enhancedCSS))
	}

	m := termstrap.New(content, opts...)
	output, err := m.Render()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Render error: %v\n", err)
		os.Exit(1)
	}
	fmt.Print(output)
}

func flagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}
