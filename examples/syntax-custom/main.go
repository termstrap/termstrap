// Example: syntax-custom — Fully custom palette without any built-in theme.
//
// Demonstrates overriding the --ts-code-* syntax palette and the document
// chrome variables (--ts-link-color, --ts-hr-color, --ts-blockquote-*) via
// WithStylesheets, on top of the default Bootstrap base (no WithTheme).
//
// Usage:
//
//	go run ./examples/syntax-custom/
//	go run ./examples/syntax-custom/ -w 100
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/termstrap/termstrap"
	"golang.org/x/term"
)

const content = `
<h3>Custom palette (no theme)</h3>

<p>An inline <code>var()</code> override and a
<a href="https://example.com">custom-colored link</a>.</p>

<pre><code class="language-python">def fib(n: int) -> int:
    # classic recursion
    if n < 2:
        return n
    return fib(n - 1) + fib(n - 2)

print([fib(i) for i in range(8)])
</code></pre>

<blockquote>Blockquote styled by --ts-blockquote-color / --ts-blockquote-border.</blockquote>

<hr>
`

// customPalette redefines every exposed terminal variable. Any var left out
// falls back to the Bootstrap default baked into css/termstrap.css.
const customPalette = `
:root {
  /* Code surfaces */
  --ts-code-bg: #0b1021;
  --ts-code-inline-bg: #1b2440;
  --ts-code-inline-fg: #ffd866;

  /* Syntax tokens */
  --ts-code-name: #e6e6fa;
  --ts-code-keyword: #ff6ac1;
  --ts-code-function: #57c7ff;
  --ts-code-type: #5af78e;
  --ts-code-builtin: #c792ea;
  --ts-code-string: #ffd866;
  --ts-code-number: #ff9e64;
  --ts-code-comment: #6b7394;
  --ts-code-operator: #89ddff;

  /* Document chrome */
  --ts-link-color: #57c7ff;
  --ts-hr-color: #ff6ac1;
  --ts-blockquote-color: #a0a8c0;
  --ts-blockquote-border: #5af78e;
}
`

func main() {
	width := 80
	flag.IntVar(&width, "w", width, "terminal width in columns")
	flag.Parse()

	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && w > 0 && !flagPassed("w") {
		width = w
	}

	m := termstrap.New(
		content,
		termstrap.WithWidth(width),
		termstrap.WithStylesheets(customPalette),
	)

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
