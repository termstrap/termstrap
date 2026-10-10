package termstrap

import (
	"strings"
	"testing"
)

func TestUnits_ResolveLength(t *testing.T) {
	ctx := UnitContext{
		ParentWidth:    80,
		ParentHeight:   24,
		TerminalWidth:  100,
		TerminalHeight: 40,
		RootFontSize:   16.0,
	}

	tests := []struct {
		val      string
		axis     Axis
		expected int
	}{
		// Pixels
		{"8px", AxisX, 1},
		{"16px", AxisX, 2},
		{"24px", AxisX, 3},
		{"16px", AxisY, 1},
		{"32px", AxisY, 2},

		// Rem
		{"0.25rem", AxisX, 1}, // Minimum visual presence (0.5 cols -> 1)
		{"0.5rem", AxisX, 1},  // 1 col
		{"1rem", AxisX, 2},    // 2 cols
		{"1.5rem", AxisX, 3},  // 3 cols
		{"2rem", AxisX, 4},    // 4 cols
		{"0.25rem", AxisY, 0}, // Below sub-line threshold
		{"1rem", AxisY, 1},    // 1 line
		{"2rem", AxisY, 2},    // 2 lines
		{"5rem", AxisY, 3},    // Capped by damping rule (max 3 lines)

		// Character units (ch)
		{"10ch", AxisX, 10},
		{"25ch", AxisX, 25},

		// Points (pt)
		{"12pt", AxisX, 2}, // 12pt = 16px = 2 cols
		{"12pt", AxisY, 1}, // 12pt = 16px = 1 line

		// Percentage (%)
		{"50%", AxisX, 40}, // 50% of parent width 80 = 40 cols
		{"25%", AxisX, 20}, // 25% of parent width 80 = 20 cols
		{"50%", AxisY, 12}, // 50% of parent height 24 = 12 lines

		// Viewport units (vw, vh)
		{"25vw", AxisX, 25}, // 25% of terminal width 100 = 25 cols
		{"50vh", AxisY, 20}, // 50% of terminal height 40 = 20 lines

		// Unitless integers (backward compatibility with terminal framework CSS)
		{"0", AxisX, 0},
		{"2", AxisX, 2},
		{"1", AxisY, 1},
		{"3", AxisY, 3},
	}

	for _, tt := range tests {
		t.Run(tt.val+"_"+axisName(tt.axis), func(t *testing.T) {
			got := ResolveLength(tt.val, tt.axis, ctx)
			if got != tt.expected {
				t.Errorf("ResolveLength(%q, %v) = %d, expected %d", tt.val, axisName(tt.axis), got, tt.expected)
			}
		})
	}
}

func axisName(a Axis) string {
	if a == AxisX {
		return "AxisX"
	}
	return "AxisY"
}

func TestUnits_ResolveBox4D(t *testing.T) {
	ctx := UnitContext{
		ParentWidth:    80,
		ParentHeight:   24,
		TerminalWidth:  80,
		TerminalHeight: 24,
		RootFontSize:   16.0,
	}

	tests := []struct {
		name           string
		input          string
		topExpected    int
		rightExpected  int
		bottomExpected int
		leftExpected   int
	}{
		{
			name:           "1 value (1rem)",
			input:          "1rem",
			topExpected:    1,
			rightExpected:  2,
			bottomExpected: 1,
			leftExpected:   2,
		},
		{
			name:           "2 values (0.5rem 1rem)",
			input:          "0.5rem 1rem",
			topExpected:    1,
			rightExpected:  2,
			bottomExpected: 1,
			leftExpected:   2,
		},
		{
			name:           "3 values (10px 20px 30px)",
			input:          "16px 24px 32px",
			topExpected:    1,
			rightExpected:  3,
			bottomExpected: 2,
			leftExpected:   3,
		},
		{
			name:           "4 values (8px 16px 24px 32px)",
			input:          "8px 16px 24px 32px",
			topExpected:    1,
			rightExpected:  2,
			bottomExpected: 2,
			leftExpected:   4,
		},
		{
			name:           "Unitless values (framework compatibility: 1 3)",
			input:          "1 3",
			topExpected:    1,
			rightExpected:  3,
			bottomExpected: 1,
			leftExpected:   3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			top, right, bottom, left := ResolveBox4D(tt.input, ctx)
			if top != tt.topExpected || right != tt.rightExpected || bottom != tt.bottomExpected || left != tt.leftExpected {
				t.Errorf("ResolveBox4D(%q) = (%d, %d, %d, %d), expected (%d, %d, %d, %d)",
					tt.input, top, right, bottom, left, tt.topExpected, tt.rightExpected, tt.bottomExpected, tt.leftExpected)
			}
		})
	}
}

func TestUnits_ResolveFontSize(t *testing.T) {
	tests := []struct {
		val          string
		expectedBold bool
		expectedDim  bool
	}{
		{"2.5rem", true, false},
		{"24px", true, false},
		{"large", true, false},
		{"1rem", false, false},
		{"16px", false, false},
		{"0.75rem", false, true},
		{"12px", false, true},
		{"small", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.val, func(t *testing.T) {
			bold, dim, _ := ResolveFontSize(tt.val, 16.0)
			if bold != tt.expectedBold || dim != tt.expectedDim {
				t.Errorf("ResolveFontSize(%q) = (bold=%v, dim=%v), expected (bold=%v, dim=%v)",
					tt.val, bold, dim, tt.expectedBold, tt.expectedDim)
			}
		})
	}
}

func TestUnits_TailwindAndPicoRendering(t *testing.T) {
	// 1. Tailwind-like CSS with arbitrary units
	tailwindCSS := `
.tailwind-card {
  padding: 1rem;
  margin: 0.5rem;
  background-color: #212529;
  color: #ffffff;
  border-radius: 0.375rem;
}
.title {
  font-size: 1.5rem;
}
`
	htmlTailwind := `
<div class="tailwind-card">
  <div class="title">Tailwind Heading</div>
  <p>Universal unit rendering works cleanly.</p>
</div>`

	m1 := New(htmlTailwind, WithWidth(80), WithStylesheets(tailwindCSS))
	out1, err := m1.Render()
	if err != nil {
		t.Fatalf("Render failed for Tailwind snippet: %v", err)
	}

	if !strings.Contains(out1, "Tailwind Heading") {
		t.Errorf("expected 'Tailwind Heading' in output, got:\n%s", out1)
	}

	// 2. PicoCSS-like CSS with pixel units
	picoCSS := `
article {
  padding: 16px;
  margin-bottom: 24px;
  border: 1px solid #ced4da;
  border-radius: 4px;
}
`
	htmlPico := `
<article>
  <p>Pico article content in terminal.</p>
</article>`

	m2 := New(htmlPico, WithWidth(80), WithStylesheets(picoCSS))
	out2, err := m2.Render()
	if err != nil {
		t.Fatalf("Render failed for Pico snippet: %v", err)
	}

	if !strings.Contains(out2, "Pico article content in terminal.") {
		t.Errorf("expected Pico content in output, got:\n%s", out2)
	}
}
