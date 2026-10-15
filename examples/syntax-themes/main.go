// Example: syntax-themes — Same document rendered across every built-in theme.
//
// Shows how syntax highlighting and the document chrome (links, <hr>,
// <blockquote>) follow each theme's charter. Pass -theme to render only one.
//
// Usage:
//
//	go run ./examples/syntax-themes/
//	go run ./examples/syntax-themes/ -w 100 -theme dracula
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/termstrap/termstrap"
	"golang.org/x/term"
)

const content = `
<h3>Highlighting &amp; chrome</h3>

<p>Inline <code>termstrap.New()</code> plus a block below. See the
<a href="https://github.com/termstrap/termstrap">project link</a> color.</p>

<pre><code class="language-go">package main

import "fmt"

// greet prints a friendly message.
func greet(name string) string {
	return fmt.Sprintf("hello, %s", name)
}
</code></pre>

<blockquote>Blockquote border + text follow the theme accent.</blockquote>

<hr>

<pre><code class="language-diff">--- a/app.go
+++ b/app.go
@@ -1,2 +1,2 @@
-	return err
+	return nil
</code></pre>
`

func main() {
	width := 80
	only := ""
	flag.IntVar(&width, "w", width, "terminal width in columns")
	flag.StringVar(&only, "theme", only, "render only this theme (bootstrap|dracula|tokyonight)")
	flag.Parse()

	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && w > 0 && !flagPassed("w") {
		width = w
	}

	themes := []string{"bootstrap", "dracula", "tokyonight"}
	if only != "" {
		themes = []string{only}
	}

	for _, theme := range themes {
		fmt.Printf("\n=== theme: %s ===\n\n", theme)
		m := termstrap.New(
			content,
			termstrap.WithWidth(width),
			termstrap.WithTheme(termstrap.Theme(theme)),
		)
		out, err := m.Render()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Render error (%s): %v\n", theme, err)
			os.Exit(1)
		}
		fmt.Print(strings.TrimRight(out, "\n"))
		fmt.Println()
	}
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
