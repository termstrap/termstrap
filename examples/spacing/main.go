// Example: spacing — Demonstrates all Bootstrap margin and padding utility variants:
// full padding/margin (p-*, m-*), directional (pt-*, pb-*, ps-*/pl-*, pe-*/pr-*, mt-*, mb-*, ms-*/ml-*, me-*/mr-*),
// axis-based (px-*, py-*, mx-*, my-*), and scales 0 to 5.
//
// Usage:
//
//	go run ./examples/spacing/
package main

import (
	"fmt"
	"os"

	"github.com/go-scripts/termstrap"
	"golang.org/x/term"
)

func main() {
	width := 105
	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && w > 0 {
		width = w
	}

	content := `
<h1>Bootstrap Spacing Utilities Demo (Padding &amp; Margin)</h1>

<h2>1. All-Sides Padding (p-0 to p-5)</h2>
<div class="row">
  <div class="col-sm-6 col-md-4 mb-2">
    <div class="border rounded bg-dark text-white p-0"><code>.p-0</code> (pad: 0)</div>
  </div>
  <div class="col-sm-6 col-md-4 mb-2">
    <div class="border rounded bg-dark text-white p-1"><code>.p-1</code> (pad: 0 1)</div>
  </div>
  <div class="col-sm-6 col-md-4 mb-2">
    <div class="border rounded bg-dark text-white p-2"><code>.p-2</code> (pad: 0 2)</div>
  </div>
  <div class="col-sm-6 col-md-4 mb-2">
    <div class="border rounded bg-dark text-white p-3"><code>.p-3</code> (pad: 1 3)</div>
  </div>
  <div class="col-sm-6 col-md-4 mb-2">
    <div class="border rounded bg-dark text-white p-4"><code>.p-4</code> (pad: 1 4)</div>
  </div>
  <div class="col-sm-6 col-md-4 mb-2">
    <div class="border rounded bg-dark text-white p-5"><code>.p-5</code> (pad: 2 5)</div>
  </div>
</div>

<hr />

<h2>2. Directional Padding (pt-*, pb-*, ps-*/pl-*, pe-*/pr-*)</h2>
<div class="row">
  <div class="col-sm-6 col-md-3 mb-2">
    <div class="border rounded bg-secondary text-white pt-3">
      <code>.pt-3</code> (top)
    </div>
  </div>
  <div class="col-sm-6 col-md-3 mb-2">
    <div class="border rounded bg-secondary text-white pb-3">
      <code>.pb-3</code> (bottom)
    </div>
  </div>
  <div class="col-sm-6 col-md-3 mb-2">
    <div class="border rounded bg-secondary text-white ps-4">
      <code>.ps-4 / .pl-4</code> (start/left)
    </div>
  </div>
  <div class="col-sm-6 col-md-3 mb-2">
    <div class="border rounded bg-secondary text-white pe-4">
      <code>.pe-4 / .pr-4</code> (end/right)
    </div>
  </div>
</div>

<hr />

<h2>3. Horizontal &amp; Vertical Padding (px-*, py-*)</h2>
<div class="row">
  <div class="col-md-6 mb-2">
    <div class="border rounded bg-dark text-white px-5">
      <code>.px-5</code> — Large horizontal padding (left &amp; right)
    </div>
  </div>
  <div class="col-md-6 mb-2">
    <div class="border rounded bg-dark text-white py-3">
      <code>.py-3</code> — Vertical padding (top &amp; bottom)
    </div>
  </div>
</div>

<hr />

<h2>4. Directional Margins (mt-*, mb-*, ms-*/ml-*, me-*/mr-*)</h2>
<div class="border rounded p-2 bg-dark text-white">
  <div>Outer Container Box</div>
  <div class="border rounded bg-primary text-white p-1 mt-2">
    <code>.mt-2</code> — Margin Top (shifts down)
  </div>
  <div class="border rounded bg-success text-white p-1 ms-4">
    <code>.ms-4 / .ml-4</code> — Margin Start / Left (indents right)
  </div>
  <div class="border rounded bg-warning text-dark p-1 me-5">
    <code>.me-5 / .mr-5</code> — Margin End / Right (reserves right space)
  </div>
  <div class="border rounded bg-info text-dark p-1 mb-2">
    <code>.mb-2</code> — Margin Bottom (reserves bottom space)
  </div>
</div>

<hr />

<h2>5. Axis Margins (mx-*, my-*) &amp; Centering</h2>
<div>
  <div class="border rounded bg-danger text-white p-1 my-2">
    <code>.my-2</code> — Vertical Margin (top &amp; bottom separation)
  </div>
  <div class="border rounded bg-secondary text-white p-1 mx-5">
    <code>.mx-5</code> — Horizontal Margin (indented from both left &amp; right edges)
  </div>
</div>
`

	m := termstrap.New(
		content,
		termstrap.WithWidth(width),
	)

	output, err := m.Render()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Render error: %v\n", err)
		os.Exit(1)
	}

	fmt.Print(output)
}
