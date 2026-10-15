// Example: breakpoints — Demonstrates responsive grid behavior across
// different terminal widths using Bootstrap-like breakpoint classes.
//
// Usage:
//
//	go run ./examples/breakpoints/
package main

import (
	"fmt"
	"os"

	"github.com/termstrap/termstrap"
	termimage "github.com/termstrap/termstrap/image"
)

func main() {
	content := `<h1>Responsive Breakpoints Demo</h1>

<h2>Breakpoint Thresholds</h2>

<table>
  <tr><th>Breakpoint</th><th>Prefix</th><th>Terminal Width</th></tr>
  <tr><td>xs</td><td>col- / col-xs-</td><td>&lt; 70 cols</td></tr>
  <tr><td>sm</td><td>col-sm-</td><td>&gt;= 70 cols</td></tr>
  <tr><td>md</td><td>col-md-</td><td>&gt;= 100 cols</td></tr>
  <tr><td>lg</td><td>col-lg-</td><td>&gt;= 130 cols</td></tr>
  <tr><td>xl</td><td>col-xl-</td><td>&gt;= 160 cols</td></tr>
  <tr><td>xxl</td><td>col-xxl-</td><td>&gt;= 190 cols</td></tr>

<hr />

<h2>col-md-6 — Stacks below 100 cols</h2>

<div class="row">
  <div class="col-md-6 border p-1">
    <div><b>Left (col-md-6)</b></div>
    <p>Visible side-by-side at 100+ cols.</p>
  </div>
  <div class="col-md-6 border p-1">
    <div><b>Right (col-md-6)</b></div>
    <p>Stacks vertically below 100 cols.</p>
  </div>
</div>

<hr />

<h2>col-lg-4 — Stacks below 130 cols</h2>

<div class="row">
  <div class="col-lg-4 border p-1">
    <div><b>A (col-lg-4)</b></div>
  </div>
  <div class="col-lg-4 border p-1">
    <div><b>B (col-lg-4)</b></div>
  </div>
  <div class="col-lg-4 border p-1">
    <div><b>C (col-lg-4)</b></div>
  </div>
</div>

<hr />

<h2>col-sm-6 — Side-by-side from 70 cols</h2>

<div class="row">
  <div class="col-sm-6 border p-1">
    <div><b>Left (col-sm-6)</b></div>
    <p>Side-by-side starting at 70 cols.</p>
  </div>
  <div class="col-sm-6 border p-1">
    <div><b>Right (col-sm-6)</b></div>
  </div>
</div>
<hr />

<h2>Mixed breakpoints — col-sm-12 / col-md-6</h2>

<div class="row">
  <div class="col-sm-12 col-md-6 border p-1">
    <div><b>Panel A</b></div>
    <p>Full width on small, half on medium+.</p>
  </div>
  <div class="col-sm-12 col-md-6 border p-1">
    <div><b>Panel B</b></div>
    <p>Full width on small, half on medium+.</p>
  </div>
</div>
`

	widths := []struct {
		name  string
		width int
	}{
		{"Very Narrow (55 cols - xs)", 55},
		{"Small (75 cols - sm)", 75},
		{"Medium / Standard (110 cols - md)", 110},
		{"Large (140 cols - lg)", 140},
		{"Extra Large (170 cols - xl)", 170},
	}

	for _, w := range widths {
		fmt.Printf("\n\033[1;36m%s\033[0m\n", w.name)
		fmt.Println("════════════════════════════════════════")
		m := termstrap.New(
			content,
			termstrap.WithWidth(w.width),
			termstrap.WithImageRenderer(termimage.NewRenderer(termimage.WithProtocol(termimage.HalfBlock))),
		)
		output, err := m.Render()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error rendering at width %d: %v\n", w.width, err)
			continue
		}
		fmt.Print(output)
	}
}
