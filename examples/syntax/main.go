// Example: syntax — Demonstrates syntax-highlighted code blocks.
//
// Termstrap colorizes <pre> blocks with chroma. The language is read from the
// inner <code class="language-xxx"> (the GitHub/CommonMark convention); when no
// class is present the language is auto-detected from the code itself.
//
// Usage:
//
//	go run ./examples/syntax/
//	go run ./examples/syntax/ -w 100 -theme dracula
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/termstrap/termstrap"
	"golang.org/x/term"
)

const content = `
<h1>Syntax Highlighting</h1>

<p>Explicit language via <code>class="language-go"</code>:</p>
<pre><code class="language-go">package main

import "fmt"

// greet prints a friendly message.
func greet(name string) string {
	return fmt.Sprintf("Hello, %s!", name)
}

func main() {
	fmt.Println(greet("termstrap"))
}
</code></pre>

<p>Another language (Python):</p>
<pre><code class="language-python">def fib(n: int) -> int:
    a, b = 0, 1
    for _ in range(n):
        a, b = b, a + b
    return a

print([fib(i) for i in range(10)])
</code></pre>

<p>No class — language auto-detected from content (shell script):</p>
<pre><code>#!/usr/bin/env bash
set -euo pipefail

for file in *.txt; do
  echo "Processing $file"
  wc -l "$file"
done
</code></pre>

<p>Unified diff (exercises inserted/deleted tokens):</p>
<pre><code class="language-diff">--- a/main.go
+++ b/main.go
@@ -1,3 +1,3 @@
-	fmt.Println("hello")
+	fmt.Println("hello, world")
 	return nil
</code></pre>
`

func main() {
	width := 80
	theme := "bootstrap"
	flag.IntVar(&width, "w", width, "terminal width in columns")
	flag.StringVar(&theme, "theme", theme, "visual theme: bootstrap|dracula|tokyonight")
	flag.Parse()

	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && w > 0 && !flagPassed("w") {
		width = w
	}

	m := termstrap.New(
		content,
		termstrap.WithWidth(width),
		termstrap.WithTheme(termstrap.Theme(theme)),
		// Highlighting is on by default. Disable with WithSyntaxHighlight(false)
		// or force a specific chroma style with WithSyntaxTheme("monokai").
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
