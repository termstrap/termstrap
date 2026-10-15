package termstrap

import (
	"strings"
	"testing"
)

func TestThemeGlamour_Headings(t *testing.T) {
	html := `<h1>Title 1</h1>
<h2>Title 2</h2>
<h3>Title 3</h3>
<h6>Title 6</h6>`

	m := New(html, WithWidth(60), WithTheme(ThemeGlamour))
	out, err := m.Render()
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	stripped := stripANSI(out)

	// H1: No "# " marker in Glamour theme
	if strings.Contains(stripped, "# Title 1") {
		t.Errorf("expected H1 to omit '#' marker in Glamour theme, got:\n%s", stripped)
	}
	if !strings.Contains(stripped, "Title 1") {
		t.Errorf("expected 'Title 1' in output, got:\n%s", stripped)
	}

	// H2: Uses "## " marker in Glamour theme
	if !strings.Contains(stripped, "## Title 2") {
		t.Errorf("expected '## Title 2' in output, got:\n%s", stripped)
	}

	// H3: Uses "### " marker in Glamour theme
	if !strings.Contains(stripped, "### Title 3") {
		t.Errorf("expected '### Title 3' in output, got:\n%s", stripped)
	}

	// H6: Uses "###### " marker in Glamour theme
	if !strings.Contains(stripped, "###### Title 6") {
		t.Errorf("expected '###### Title 6' in output, got:\n%s", stripped)
	}

	// H1 colors: ANSI background 48;2;95;95;255 (#5f5fff) and foreground 38;2;255;255;135 (#ffff87)
	if !strings.Contains(out, "48;2;95;95;255") {
		t.Errorf("expected H1 background color #5f5fff (48;2;95;95;255) in output, got:\n%q", out)
	}
	if !strings.Contains(out, "38;2;255;255;135") {
		t.Errorf("expected H1 foreground color #ffff87 (38;2;255;255;135) in output, got:\n%q", out)
	}

	// H2 marker/text color: #00afff (38;2;0;175;255)
	if !strings.Contains(out, "38;2;0;175;255") {
		t.Errorf("expected H2 color #00afff (38;2;0;175;255) in output, got:\n%q", out)
	}

	// H6 color: #00af5f (38;2;0;175;95)
	if !strings.Contains(out, "38;2;0;175;95") {
		t.Errorf("expected H6 color #00af5f (38;2;0;175;95) in output, got:\n%q", out)
	}
}

func TestThemeGlamour_InlineCodeAndLinks(t *testing.T) {
	html := `<p>Here is a <a href="https://example.com">Glamour link</a> and <code>inline cmd</code>.</p>`

	m := New(html, WithWidth(60), WithTheme(ThemeGlamour))
	out, err := m.Render()
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	// Link color: #00af5f -> 38;2;0;175;95
	if !strings.Contains(out, "38;2;0;175;95") {
		t.Errorf("expected link color #00af5f (38;2;0;175;95) in output, got:\n%q", out)
	}

	// Inline code foreground: #ff5f5f -> 38;2;255;95;95
	if !strings.Contains(out, "38;2;255;95;95") {
		t.Errorf("expected code color #ff5f5f (38;2;255;95;95) in output, got:\n%q", out)
	}

	// Inline code background: #303030 -> 48;2;48;48;48
	if !strings.Contains(out, "48;2;48;48;48") {
		t.Errorf("expected code background #303030 (48;2;48;48;48) in output, got:\n%q", out)
	}
}

func TestThemeGlamour_Blockquote(t *testing.T) {
	html := `<blockquote>Glamour style citation quote.</blockquote>`

	m := New(html, WithWidth(60), WithTheme(ThemeGlamour))
	out, err := m.Render()
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	stripped := stripANSI(out)
	if !strings.Contains(stripped, "│ Glamour style citation quote.") {
		t.Errorf("expected '│ Glamour style citation quote.' with 1-space indent, got:\n%s", stripped)
	}

	// Border color: #676767 -> 38;2;103;103;103
	if !strings.Contains(out, "38;2;103;103;103") {
		t.Errorf("expected blockquote border color #676767 (38;2;103;103;103) in output, got:\n%q", out)
	}
}

func TestThemeGlamour_Table(t *testing.T) {
	html := `<table class="table">
  <thead><tr><th>Header 1</th><th>Header 2</th></tr></thead>
  <tbody><tr><td>Cell 1</td><td>Cell 2</td></tr></tbody>
</table>`

	m := New(html, WithWidth(50), WithTheme(ThemeGlamour))
	out, err := m.Render()
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	// Table border color #585858 -> 38;2;88;88;88
	if !strings.Contains(out, "38;2;88;88;88") {
		t.Errorf("expected table border color #585858 (38;2;88;88;88) in output, got:\n%q", out)
	}

	// Header color #00afff -> 38;2;0;175;255
	if !strings.Contains(out, "38;2;0;175;255") {
		t.Errorf("expected header color #00afff (38;2;0;175;255) in output, got:\n%q", out)
	}
}

func TestThemeGlamour_SyntaxHighlighting(t *testing.T) {
	html := `<pre><code class="language-go">package main

func main() {
}</code></pre>`

	m := New(html, WithWidth(60), WithTheme(ThemeGlamour))
	out, err := m.Render()
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	// Pre background #2a2a2a -> 48;2;42;42;42
	if !strings.Contains(out, "48;2;42;42;42") {
		t.Errorf("expected pre background color #2a2a2a (48;2;42;42;42) in output, got:\n%q", out)
	}

	// Keyword "func" / "package" -> #00aaff (38;2;0;170;255)
	if !strings.Contains(out, "38;2;0;170;255") {
		t.Errorf("expected syntax keyword color #00aaff (38;2;0;170;255) in output, got:\n%q", out)
	}

	// Function "main" -> #00d787 (38;2;0;215;135)
	if !strings.Contains(out, "38;2;0;215;135") {
		t.Errorf("expected syntax function color #00d787 (38;2;0;215;135) in output, got:\n%q", out)
	}
}

func TestLinkMode_Inline(t *testing.T) {
	html := `<p>Visit <a href="https://example.com">Example Site</a> for more info.</p>`
	m := New(html, WithWidth(70), WithTheme(ThemeGlamour), WithLinkMode(LinkModeInline))
	out, err := m.Render()
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	stripped := stripANSI(out)
	// In LinkModeInline, both the link text and raw URL appear in the text
	if !strings.Contains(stripped, "Example Site https://example.com") {
		t.Errorf("expected inline link with URL, got:\n%s", stripped)
	}
}

func TestLinkMode_Footnotes(t *testing.T) {
	html := `<p>Check <a href="https://first.org">First</a> and <a href="https://second.org">Second</a>.</p>`
	m := New(html, WithWidth(70), WithTheme(ThemeGlamour), WithLinkMode(LinkModeFootnotes))
	out, err := m.Render()
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	stripped := stripANSI(out)
	// Links should have [1] and [2] citations
	if !strings.Contains(stripped, "First [1]") {
		t.Errorf("expected 'First [1]' in body, got:\n%s", stripped)
	}
	if !strings.Contains(stripped, "Second [2]") {
		t.Errorf("expected 'Second [2]' in body, got:\n%s", stripped)
	}

	// Footnotes section at bottom
	if !strings.Contains(stripped, "[1] https://first.org") {
		t.Errorf("expected footnote '[1] https://first.org', got:\n%s", stripped)
	}
	if !strings.Contains(stripped, "[2] https://second.org") {
		t.Errorf("expected footnote '[2] https://second.org', got:\n%s", stripped)
	}
}

func TestImageMode_Glamour(t *testing.T) {
	html := `<p><img src="https://example.com/logo.png" alt="Company Logo"></p>`
	m := New(html, WithWidth(70), WithTheme(ThemeGlamour), WithImageMode(ImageModeGlamour))
	out, err := m.Render()
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	stripped := stripANSI(out)
	if !strings.Contains(stripped, "Image: Company Logo →") {
		t.Errorf("expected 'Image: Company Logo →' placeholder, got:\n%s", stripped)
	}
	if !strings.Contains(stripped, "https://example.com/logo.png") {
		t.Errorf("expected image source URL in output, got:\n%s", stripped)
	}
}

func TestGlamourPreset(t *testing.T) {
	html := `<h1>Preset Demo</h1>
<p>See <a href="https://example.com">Site</a>.</p>
<p><img src="https://example.com/banner.png" alt="Banner"></p>`

	m := New(html, WithWidth(70), WithGlamourPreset())
	out, err := m.Render()
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	stripped := stripANSI(out)
	// H1: No "# " marker
	if strings.Contains(stripped, "# Preset Demo") {
		t.Errorf("expected H1 without marker, got:\n%s", stripped)
	}
	// Inline link URL
	if !strings.Contains(stripped, "Site https://example.com") {
		t.Errorf("expected inline URL in link, got:\n%s", stripped)
	}
	// Glamour image placeholder
	if !strings.Contains(stripped, "Image: Banner →") {
		t.Errorf("expected glamour image placeholder, got:\n%s", stripped)
	}
}

func TestImageMode_Smart(t *testing.T) {
	t.Run("Narrow Width Fallback", func(t *testing.T) {
		// Content width is 8 cols (< min 12)
		html := `<div style="width: 8ch;"><img src="https://example.com/photo.png" alt="Tiny Col Photo"></div>`
		m := New(html, WithWidth(60), WithImageMode(ImageModeSmart), WithMinImageWidth(12))
		out, err := m.Render()
		if err != nil {
			t.Fatalf("Render failed: %v", err)
		}
		stripped := stripANSI(out)
		collapsed := strings.Join(strings.Fields(stripped), " ")
		if !strings.Contains(collapsed, "Image: Tiny Col Photo →") {
			t.Errorf("expected smart fallback for narrow column, got:\n%s", stripped)
		}
	})

	t.Run("SVG Vector Fallback", func(t *testing.T) {
		html := `<p><img src="https://example.com/vector.svg" alt="Vector Diagram"></p>`
		m := New(html, WithWidth(70), WithImageMode(ImageModeSmart))
		out, err := m.Render()
		if err != nil {
			t.Fatalf("Render failed: %v", err)
		}
		stripped := stripANSI(out)
		if !strings.Contains(stripped, "Image: Vector Diagram →") {
			t.Errorf("expected smart fallback for SVG vector, got:\n%s", stripped)
		}
	})

	t.Run("Load Error Fallback", func(t *testing.T) {
		html := `<p><img src="/non/existent/path/404.jpg" alt="Missing Photo"></p>`
		m := New(html, WithWidth(70), WithImageMode(ImageModeSmart))
		out, err := m.Render()
		if err != nil {
			t.Fatalf("Render failed: %v", err)
		}
		stripped := stripANSI(out)
		if !strings.Contains(stripped, "Image: Missing Photo →") {
			t.Errorf("expected smart fallback for 404 load error, got:\n%s", stripped)
		}
	})

	t.Run("Default Auto Keeps Graphics", func(t *testing.T) {
		// When ImageMode is Auto (default), it doesn't emit "Image: ... →" placeholder
		html := `<p><img src="/non/existent/path/404.jpg" alt="Missing Photo"></p>`
		m := New(html, WithWidth(70), WithImageMode(ImageModeAuto))
		out, err := m.Render()
		if err != nil {
			t.Fatalf("Render failed: %v", err)
		}
		stripped := stripANSI(out)
		if strings.Contains(stripped, "Image: Missing Photo →") {
			t.Errorf("expected Auto mode not to emit Glamour placeholder, got:\n%s", stripped)
		}
		if !strings.Contains(stripped, "[Missing Photo]") {
			t.Errorf("expected standard '[Missing Photo]' placeholder in Auto mode, got:\n%s", stripped)
		}
	})
}
