package termstrap

import (
	"strings"
	"testing"
)

func TestShimmer_Basic(t *testing.T) {
	text := "Working..."
	s0 := Shimmer(text, 0)
	s1 := Shimmer(text, 1)

	if s0 == "" || s1 == "" {
		t.Fatal("expected non-empty shimmer strings")
	}

	// Shimmer should contain ANSI 24-bit color sequences
	if !strings.Contains(s0, "38;2;") {
		t.Errorf("expected 24-bit ANSI color sequence in shimmer output, got %q", s0)
	}

	// Different steps should yield different ANSI color waves
	if s0 == s1 {
		t.Error("expected different color phase between step 0 and step 1")
	}

	// Plain text characters should be preserved in order
	plain := stripANSI(s0)
	if plain != text {
		t.Errorf("expected plain text %q, got %q", text, plain)
	}
}

func TestShimmer_Empty(t *testing.T) {
	if got := Shimmer("", 0); got != "" {
		t.Errorf("expected empty string for empty input, got %q", got)
	}
}

func TestSpinner_Frames(t *testing.T) {
	for i := range len(SpinnerFrames) {
		f := Spinner(i)
		if f != SpinnerFrames[i] {
			t.Errorf("expected frame %q at index %d, got %q", SpinnerFrames[i], i, f)
		}
	}

	// Cycle test
	fCycle := Spinner(len(SpinnerFrames))
	if fCycle != SpinnerFrames[0] {
		t.Errorf("expected cycle back to frame 0, got %q", fCycle)
	}

	// Custom frames
	custom := []string{"|", "/", "-", "\\"}
	if got := Spinner(1, custom); got != "/" {
		t.Errorf("expected custom frame '/', got %q", got)
	}
}

func TestDeclarativeSpinnerAndShimmer(t *testing.T) {
	html := `<div><spinner></spinner> <shimmer>Loading data...</shimmer></div>`
	m0 := New(html, WithWidth(40), WithStep(0))
	out0, err := m0.Render()
	if err != nil {
		t.Fatalf("render step 0: %v", err)
	}

	m1 := New(html, WithWidth(40), WithStep(1))
	out1, err := m1.Render()
	if err != nil {
		t.Fatalf("render step 1: %v", err)
	}

	// Should contain the first spinner frame
	if !strings.Contains(out0, SpinnerFrames[0]) {
		t.Errorf("expected spinner frame %q in out0:\n%q", SpinnerFrames[0], out0)
	}
	// Should contain the second spinner frame at step 1
	if !strings.Contains(out1, SpinnerFrames[1]) {
		t.Errorf("expected spinner frame %q in out1:\n%q", SpinnerFrames[1], out1)
	}

	// Shimmer should have emitted 24-bit ANSI colors
	if !strings.Contains(out0, "38;2;") {
		t.Errorf("expected 24-bit ANSI in out0:\n%q", out0)
	}

	// Text content should be preserved
	if !strings.Contains(stripANSI(out0), "Loading data...") {
		t.Errorf("expected plain text 'Loading data...' in output:\n%q", out0)
	}
}

func TestClassSpinnerAndShimmer(t *testing.T) {
	html := `<div><span class="spinner"></span> <span class="shimmer" style="color: #6272a4;">Analyzing...</span></div>`
	m := New(html, WithWidth(40), WithStep(2))
	out, err := m.Render()
	if err != nil {
		t.Fatalf("render: %v", err)
	}

	if !strings.Contains(out, SpinnerFrames[2]) {
		t.Errorf("expected spinner frame %q in output:\n%q", SpinnerFrames[2], out)
	}
	if !strings.Contains(stripANSI(out), "Analyzing...") {
		t.Errorf("expected plain text in output:\n%q", out)
	}
}

func TestShimmer_WithClassColor(t *testing.T) {
	html := `<h1 class="text-warning shimmer">Hubs.sh</h1>`
	m := New(html, WithWidth(40), WithTheme(ThemeTokyoNight), WithStep(0))
	out, err := m.Render()
	if err != nil {
	}
	t.Logf("OUTPUT:\n%q\n%s", out, out)

	// TokyoNight warning is #e0af68 (RGB 224;175;104)
	if !strings.Contains(out, "224;175;104") {
		t.Errorf("expected TokyoNight text-warning RGB 224;175;104 in shimmer output, got:\n%q", out)
	}
}
