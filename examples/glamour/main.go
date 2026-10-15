// Example: glamour — Demonstrates all Markdown/HTML visual components
// styled according to the Glamour (Charmbracelet) design charter:
//   - Purple/yellow banner H1 (fit-content, no marker)
//   - Cyan H2-H5 with "## ".."##### " markers, and emerald H6 with "###### "
//   - Bold green links with terminal OSC 8 hyperlinks
//   - Coral inline code on dark charcoal background
//   - Citations with vertical bar "│ " and 1-space indentation
//   - Clean Lipgloss tables with custom border colors and cyan headers
//   - Full Chroma syntax-highlighted code blocks
//   - Bootstrap semantic cards matching the Glamour palette
//
// Usage:
//
//	go run ./examples/glamour/
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/termstrap/termstrap"
	"golang.org/x/term"
)

const document = `<div class="container p-1">
  <h1>Glamour Terminal Theme</h1>
  <p class="text-secondary">A Charm-inspired visual charter for HTML rendering in terminal interfaces.</p>

  <hr>

  <h2>Typography &amp; Inline Elements</h2>
  <p>
    TermStrap converts standard HTML into terminal representations matching the aesthetic of
    <a href="https://github.com/charmbracelet/glamour">Charm's Glamour</a> markdown engine.
    You can use <strong>bold text</strong>, <em>italic emphasis</em>, <del>strikethrough text</del>,
    and highlighted <span class="text-warning">warning</span> or <span class="text-info">informational</span> spans.
  </p>
  <p>
    Execute commands inline with <code>go get github.com/termstrap/termstrap</code> or review
    <code>termstrap.WithTheme(termstrap.ThemeGlamour)</code>.
  </p>

  <h3>Heading Hierarchy</h3>
  <p>Headings 2 through 5 use electric cyan Markdown prefixes, while H6 adopts an emerald green tint:</p>
  <h4>Heading Level 4</h4>
  <p class="text-muted">Sub-section content under H4.</p>
  <h5>Heading Level 5</h5>
  <p class="text-muted">Minor section note under H5.</p>
  <h6>Heading Level 6 (Minor Category)</h6>
  <p class="text-muted">Footnote or low-level heading without bold weight.</p>

  <hr>

  <h2>Blockquotes &amp; Citations</h2>
  <blockquote>
    "Simplicity is prerequisite for reliability."
    <br>
    — Edsger W. Dijkstra
  </blockquote>
  <hr>

  <h2>Image Handling Strategies</h2>
  <p>Glamour CLI mode represents images as text placeholders with target links:</p>
  <p><img src="https://golang.org/doc/gopher/frontpage.png" alt="The Go Gopher Logo"></p>

  <hr>
  <h2>Structured Lists</h2>
  <div class="row">
    <div class="col-md-6">
      <p class="fw-bold">Features (Unordered):</p>
      <ul>
        <li>Native terminal box-model &amp; auto-wrapping</li>
        <li>12-column responsive grid system</li>
        <li>Full ANSI TrueColor palette resolution</li>
        <li>Built-in Charm Glamour, Dracula &amp; Tokyo Night themes</li>
      </ul>
    </div>
    <div class="col-md-6">
      <p class="fw-bold">Getting Started (Ordered):</p>
      <ol>
        <li>Import <code>github.com/termstrap/termstrap</code></li>
        <li>Configure layout options via <code>WithWidth()</code></li>
        <li>Select <code>WithTheme(ThemeGlamour)</code></li>
        <li>Call <code>m.Render()</code> and print to stdout</li>
      </ol>
    </div>
  </div>

  <hr>

  <h2>Data Tables</h2>
  <p>Clean Lipgloss tables with header separation and Glamour borders:</p>
  <table class="table">
    <thead>
      <tr>
        <th>Package</th>
        <th>Role</th>
        <th>Status</th>
        <th>Version</th>
      </tr>
    </thead>
    <tbody>
      <tr>
        <td><code>termstrap</code></td>
        <td>HTML / CSS Engine</td>
        <td><span class="text-success">Active</span></td>
        <td>v1.2.0</td>
      </tr>
      <tr>
        <td><code>glamour</code></td>
        <td>Markdown Stylesheet</td>
        <td><span class="text-info">Reference</span></td>
        <td>v0.8.0</td>
      </tr>
      <tr>
        <td><code>lipgloss</code></td>
        <td>Style Primitives</td>
        <td><span class="text-warning">Upstream</span></td>
        <td>v2.0.0</td>
      </tr>
    </tbody>
  </table>

  <hr>

  <h2>Syntax Highlighting</h2>
  <p>Code blocks use the Glamour Chroma colorization palette:</p>
  <pre><code class="language-go">package main

import (
	"fmt"

	"github.com/termstrap/termstrap"
)

// main initializes TermStrap with the Glamour visual theme.
func main() {
	html := ` + "`" + `<h1>Hello Glamour!</h1>` + "`" + `
	out, err := termstrap.New(html, termstrap.WithTheme(termstrap.ThemeGlamour)).Render()
	if err != nil {
		panic(err)
	}
	fmt.Println(out)
}</code></pre>

  <hr>

  <h2>Bootstrap Semantic Palette (Glamour Tuned)</h2>
  <div class="row">
    <div class="col-md-4 mb-1">
      <div class="p-1 bg-primary text-white rounded">
        <strong>Primary (#5f5fff)</strong>
        <p class="mb-0">Royal Indigo banner accent</p>
      </div>
    </div>
    <div class="col-md-4 mb-1">
      <div class="p-1 bg-info text-dark rounded">
        <strong>Info (#00afff)</strong>
        <p class="mb-0">Electric Cyan heading color</p>
      </div>
    </div>
    <div class="col-md-4 mb-1">
      <div class="p-1 bg-success text-dark rounded">
        <strong>Success (#00d787)</strong>
        <p class="mb-0">Function &amp; positive status green</p>
      </div>
    </div>
  </div>

  <div class="row mt-1">
    <div class="col-md-4 mb-1">
      <div class="p-1 bg-warning text-dark rounded">
        <strong>Warning (#ffff87)</strong>
        <p class="mb-0">Title text &amp; decorator yellow</p>
      </div>
    </div>
    <div class="col-md-4 mb-1">
      <div class="p-1 bg-danger text-white rounded">
        <strong>Danger (#fd5b5b)</strong>
        <p class="mb-0">Coral red diff deleted &amp; error</p>
      </div>
    </div>
    <div class="col-md-4 mb-1">
      <div class="p-1 bg-dark text-white rounded border">
        <strong>Dark (#202020)</strong>
        <p class="mb-0">Charcoal background surface</p>
      </div>
    </div>
  </div>
</div>`

func main() {
	width := 84
	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && w > 0 {
		width = w
	}
	linkMode := "inline"
	imageMode := "glamour"
	flag.IntVar(&width, "w", width, "terminal width in columns")
	flag.StringVar(&linkMode, "link-mode", linkMode, "link presentation: osc8|inline|footnotes")
	flag.StringVar(&imageMode, "image-mode", imageMode, "image rendering: auto|glamour|placeholder")
	flag.Parse()

	m := termstrap.New(
		document,
		termstrap.WithWidth(width),
		termstrap.WithTheme(termstrap.ThemeGlamour),
		termstrap.WithLinkMode(termstrap.LinkMode(linkMode)),
		termstrap.WithImageMode(termstrap.ImageMode(imageMode)),
	)

	output, err := m.Render()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Render error: %v\n", err)
		os.Exit(1)
	}

	fmt.Print(output)
}
