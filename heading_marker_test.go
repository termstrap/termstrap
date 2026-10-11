package termstrap

import (
	"strings"
	"testing"
)

// TestHeadingMarkers verifies the Markdown-style #/##/… markers are placed
// around <h1>-<h6> text according to the heading's text alignment:
//   - centered  -> marker on both sides of every line
//   - right      -> marker only on the right
//   - otherwise  -> marker only on the left
func TestHeadingMarkers(t *testing.T) {
	render := func(html string) string {
		out, err := Model{HTML: html, Width: 40}.Render()
		if err != nil {
			t.Fatalf("render %q: %v", html, err)
		}
		return stripANSI(out)
	}

	t.Run("left default", func(t *testing.T) {
		got := render(`<h1>Title</h1>`)
		if !strings.Contains(got, "# Title") {
			t.Fatalf("expected leading marker, got %q", got)
		}
		if strings.Contains(got, "Title #") {
			t.Fatalf("unexpected trailing marker, got %q", got)
		}
	})

	t.Run("right aligned", func(t *testing.T) {
		got := render(`<h3 class="text-end">Title</h3>`)
		if !strings.Contains(got, "Title ###") {
			t.Fatalf("expected trailing marker, got %q", got)
		}
		if strings.Contains(got, "### Title") {
			t.Fatalf("unexpected leading marker, got %q", got)
		}
	})

	t.Run("centered both sides", func(t *testing.T) {
		got := render(`<h2 class="text-center">Title</h2>`)
		if !strings.Contains(got, "## Title ##") {
			t.Fatalf("expected markers on both sides, got %q", got)
		}
	})

	t.Run("level drives marker width", func(t *testing.T) {
		got := render(`<h4>Title</h4>`)
		if !strings.Contains(got, "#### Title") {
			t.Fatalf("expected 4 hashes for h4, got %q", got)
		}
	})
}
