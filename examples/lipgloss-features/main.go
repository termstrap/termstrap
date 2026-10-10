// Example: lipgloss-features — Demonstrates the CSS features newly backed by
// lipgloss: text decorations (strikethrough/faint/blink/reverse), border styles
// (double/dashed/dotted/thick/hidden), per-side border colors, max-width
// clamping, flexbox justify-content/align-items, enumerated lists with
// list-style-type, trees, and tables rendered by the lipgloss table engine.
//
// Usage:
//
//	go run ./examples/lipgloss-features/
package main

import (
	"fmt"
	"os"

	"github.com/go-scripts/termstrap"
	"golang.org/x/term"
)

func main() {
	width := 90
	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && w > 0 {
		width = w
	}

	content := `<h1>Lipgloss Feature Showcase</h1>

<h2>Text Decorations & Underlines</h2>
<p>
  <del>deprecated price</del> <span class="text-success">new price</span>,
  <s>struck through</s>,
  <span style="opacity: 0.5;">faint muted</span>,
  <span style="text-decoration: underline; text-decoration-style: wavy; text-decoration-color: #ffc107;">curly warning underline</span>,
  <span style="text-decoration: underline; text-decoration-style: double; text-decoration-color: #0dcaf0;">double cyan underline</span>,
  <blink><span class="text-danger">blinking alert</span></blink>.
<h2>Border Styles</h2>
<div class="row">
  <div class="col-4 p-1" style="border-style: double;">double</div>
  <div class="col-4 p-1" style="border-style: dashed;">dashed</div>
  <div class="col-4 p-1" style="border-style: dotted;">dotted</div>
</div>
<div class="row">
  <div class="col-4 p-1" style="border-style: thick;">thick</div>
  <div class="col-4 p-1" style="border-style: block;">block</div>
  <div class="col-4 p-1" style="border: 1px solid; border-image: #00FA68 #9900FF #ED5353;">border blend</div>
</div>
<h2>Color Blending & Gradients</h2>
<div class="row">
  <div class="col-6 p-1 text-white" style="background: linear-gradient(#ff0055, #7928ca, #0070f3); border-radius: 4px;">
    Linear Gradient Background (CIELAB Blend)
  </div>
  <div class="col-6 p-1" style="background-color: lighten(#0d6efd, 40%); color: darken(#0d6efd, 40%); border: 1px solid complementary(#0d6efd);">
    Lighten / Darken / Complementary
  </div>
</div>
<h2>Max Width Clamp</h2>
<div class="p-1" style="max-width: 30ch; border: 1px solid;">
  This paragraph is clamped to a maximum width of 30 cells even though the viewport is wider.
</div>

<h2>Flex Alignment</h2>
<div class="d-flex" style="display:flex; justify-content: space-between;">
  <div class="col-3 border p-1">left</div>
  <div class="col-3 border p-1">right</div>
</div>
<div class="d-flex align-items-center" style="display:flex;">
  <div class="col-6 border p-1">tall<br>column<br>of text</div>
  <div class="col-6 border p-1">centered</div>
</div>

<h2>Lists</h2>
<ul>
  <li>Bulleted item
    <ul><li>nested one</li><li>nested two</li></ul>
  </li>
  <li>Second item</li>
</ul>
<ol>
  <li>First ordered</li>
  <li>Second ordered</li>
</ol>
<ul style="list-style-type: upper-roman;">
  <li>Chapter one</li>
  <li>Chapter two</li>
</ul>

<h2>Tree</h2>
<ul class="tree">
  <li>root
    <ul>
      <li>branch a
        <ul><li>leaf a1</li><li>leaf a2</li></ul>
      </li>
      <li>branch b</li>
    </ul>
  </li>
</ul>

<h2>Table</h2>
<table class="table">
  <thead><tr><th>#</th><th>Service</th><th>Env</th></tr></thead>
  <tbody>
    <tr><td>101</td><td>Billing Engine</td><td>Staging</td></tr>
    <tr style="background-color: #cfe2ff;"><td>104</td><td>Email Dispatcher</td><td>Production</td></tr>
    <tr><td>105</td><td>Legacy Cron</td><td>Production</td></tr>
  </tbody>
</table>
`

	m := termstrap.New(content, termstrap.WithWidth(width))
	output, err := m.Render()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Print(output)
}
