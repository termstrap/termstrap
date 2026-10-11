package termstrap

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestAsciiTag_Basic(t *testing.T) {
	html := `<ascii width="50" color="#ffffff">Hello</ascii>`
	m := New(html, WithWidth(60))
	out, err := m.Render()
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}

	// Output should contain ASCII art representation of "Hello" (letters H, e, l, o in figlet)
	if !strings.Contains(out, "|") && !strings.Contains(out, "_") {
		t.Fatalf("Expected ASCII banner glyphs in output, got:\n%s", out)
	}

	// Should contain ANSI color sequence for #ffffff
	if !strings.Contains(out, "255;255;255") && !strings.Contains(out, "ffffff") {
		t.Errorf("Expected foreground color #ffffff in output, got:\n%s", out)
	}
}

func TestAsciiTag_CustomFont(t *testing.T) {
	html := `<ascii font="slant" color="#ff0000">Hi</ascii>`
	m := New(html, WithWidth(60))
	out, err := m.Render()
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}

	// Slant font uses / and \ characters
	if !strings.Contains(out, "/") && !strings.Contains(out, `\`) {
		t.Fatalf("Expected slant banner glyphs in output, got:\n%s", out)
	}

	// Should contain ANSI color sequence for #ff0000 (red)
	if !strings.Contains(out, "255;0;0") && !strings.Contains(out, "ff0000") {
		t.Errorf("Expected foreground color #ff0000 in output, got:\n%s", out)
	}
}

func TestAsciiTag_WidthConstraint(t *testing.T) {
	// Constrained to width 20 columns
	html := `<ascii width="20">Hello World Long Text</ascii>`
	m := New(html, WithWidth(80))
	out, err := m.Render()
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}

	lines := strings.Split(out, "\n")
	for _, l := range lines {
		stripped := stripANSI(l)
		if len(stripped) > 20 {
			t.Errorf("Line exceeded width 20: %q (len %d)", stripped, len(stripped))
		}
	}
}

func TestAsciiTag_WithStyleAndBootstrap(t *testing.T) {
	html := `<div class="bg-dark p-2 border"><ascii class="text-warning" font="doom">GO</ascii></div>`
	m := New(html, WithWidth(80))
	out, err := m.Render()
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}

	if !strings.Contains(out, "|") && !strings.Contains(out, "_") {
		t.Fatalf("Expected ASCII art output, got:\n%s", out)
	}
}

func TestAsciiTag_PercentWidthTextCenterAndShimmer(t *testing.T) {
	html := `<ascii width="100%" color="#7aa2f7" class="text-center shimmer">Hubs.sh</ascii>`
	m := New(html, WithWidth(100), WithStep(2))
	out, err := m.Render()
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}

	// Output should contain ASCII banner
	if !strings.Contains(out, "|") && !strings.Contains(out, "_") {
		t.Fatalf("Expected ASCII art output, got:\n%s", out)
	}

	// Shimmer produces 24-bit ANSI codes (38;2;)
	if !strings.Contains(out, "38;2;") {
		t.Errorf("Expected shimmer 24-bit ANSI color sequences in output, got:\n%s", out)
	}

	// Text-center: check that first non-empty line has leading spaces for centering
	lines := strings.Split(out, "\n")
	hasCenteredLine := false
	for _, l := range lines {
		rawWithoutAnsi := ansiRegex.ReplaceAllString(l, "")
		if strings.TrimSpace(rawWithoutAnsi) != "" && strings.HasPrefix(rawWithoutAnsi, " ") {
			hasCenteredLine = true
			break
		}
	}
	if !hasCenteredLine {
		t.Errorf("Expected centered ASCII art line with leading spaces in output:\n%s", out)
	}
}

func TestAsciiTag_BorderShimmer(t *testing.T) {
	html := `<div class="w-100 text-center"><ascii class="shimmer border border-shimmer" color="#7aa2f7">Hubs.sh</ascii></div>`
	m := New(html, WithWidth(100), WithStep(3))
	out, err := m.Render()
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}

	// Border glyphs should be present
	if !strings.Contains(out, "┌") || !strings.Contains(out, "┐") {
		t.Fatalf("Expected border glyphs in output, got:\n%s", out)
	}

	// Border should carry 24-bit ANSI colors from shimmer
	if !strings.Contains(out, "38;2;") {
		t.Errorf("Expected 24-bit ANSI colors in border-shimmer output, got:\n%s", out)
	}
}

func TestAsciiFonts_ListAndResolution(t *testing.T) {
	fonts := AvailableAsciiFonts()
	if len(fonts) < 100 {
		t.Fatalf("Expected over 100 available FIGlet fonts, got %d", len(fonts))
	}

	// Test case-insensitivity: SLANT vs slant vs Slant should all work and not panic
	for _, f := range []string{"SLANT", "Slant", "dOoM", "BLOCK", "non-existent-fallback-font"} {
		html := `<ascii font="` + f + `">Test</ascii>`
		m := New(html, WithWidth(80))
		out, err := m.Render()
		if err != nil {
			t.Fatalf("Render failed for font %q: %v", f, err)
		}
		if len(strings.TrimSpace(out)) == 0 {
			t.Fatalf("Expected non-empty output for font %q", f)
		}
	}
}

func TestAsciiFonts_ResponsiveFallback(t *testing.T) {
	// At width 80, "doh" (123 cols) cannot fit. It must fallback to "small" or "doom".
	html := `<ascii font="doh,doom,small">Hubs.sh</ascii>`
	m := New(html, WithWidth(80))
	out, err := m.Render()
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	// Output lines must fit within 80 cols
	for _, l := range strings.Split(out, "\n") {
		stripped := stripANSI(l)
		if len(stripped) > 80 {
			t.Errorf("Line exceeded 80 cols in responsive fallback: %q", stripped)
		}
	}
}

func TestAsciiFonts_FitScale(t *testing.T) {
	// "doh" needs ~123 cols. fit="scale" must keep doh and downscale it (not swap font),
	// producing block art that fits within the width via Unicode sextant characters.
	html := `<ascii font="doh" fit="scale">Hubs.sh</ascii>`
	m := New(html, WithWidth(60))
	out, err := m.Render()
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	for _, l := range strings.Split(out, "\n") {
		if w := utf8.RuneCountInString(stripANSI(l)); w > 60 {
			t.Errorf("Line exceeded 60 cols under fit=scale: width=%d %q", w, l)
		}
	}

	// Must contain block/sextant glyphs proving the art was downscaled, not clipped.
	if !strings.ContainsAny(stripANSI(out), "█▌▐🬀🬁🬂🬃🬄🬅🬆🬇🬈🬉🬊🬋🬌🬍🬎🬏") {
		t.Errorf("Expected Unicode block glyphs from downscaling, got:\n%s", out)
	}
}
