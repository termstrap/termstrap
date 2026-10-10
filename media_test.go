package termstrap

import (
	"strings"
	"testing"
)

func TestMedia_ParseMediaQuery(t *testing.T) {
	tests := []struct {
		prelude          string
		expectedClauses  int
		expectedMinWidth int
		expectedMaxWidth int
	}{
		{
			prelude:          "(min-width: 768px)",
			expectedClauses:  1,
			expectedMinWidth: 96, // 768 / 8 = 96 cols
			expectedMaxWidth: -1,
		},
		{
			prelude:          "(min-width: 576px) and (max-width: 767.98px)",
			expectedClauses:  1,
			expectedMinWidth: 72, // 576 / 8 = 72 cols
			expectedMaxWidth: 95, // floor(767.98 / 8) = 95 cols
		},
		{
			prelude:          "screen and (min-width: 100ch)",
			expectedClauses:  1,
			expectedMinWidth: 100,
			expectedMaxWidth: -1,
		},
		{
			prelude:          "(max-width: 50ch), (min-width: 120ch)",
			expectedClauses:  2,
			expectedMinWidth: -1,
			expectedMaxWidth: 50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.prelude, func(t *testing.T) {
			cond := ParseMediaQuery(tt.prelude)
			if cond == nil {
				t.Fatalf("ParseMediaQuery(%q) returned nil", tt.prelude)
			}
			if len(cond.Clauses) != tt.expectedClauses {
				t.Errorf("expected %d clauses, got %d", tt.expectedClauses, len(cond.Clauses))
			}
			c0 := cond.Clauses[0]
			if c0.MinWidth != tt.expectedMinWidth {
				t.Errorf("expected MinWidth=%d, got %d", tt.expectedMinWidth, c0.MinWidth)
			}
			if c0.MaxWidth != tt.expectedMaxWidth {
				t.Errorf("expected MaxWidth=%d, got %d", tt.expectedMaxWidth, c0.MaxWidth)
			}
		})
	}
}

func TestMedia_Matching(t *testing.T) {
	cond := ParseMediaQuery("(min-width: 80ch) and (max-width: 120ch)")
	if cond == nil {
		t.Fatal("failed to parse media query")
	}

	if cond.Matches(60, 24, ThemeBootstrap) {
		t.Errorf("width 60 should NOT match [80ch, 120ch]")
	}
	if !cond.Matches(80, 24, ThemeBootstrap) {
		t.Errorf("width 80 SHOULD match [80ch, 120ch]")
	}
	if !cond.Matches(100, 24, ThemeBootstrap) {
		t.Errorf("width 100 SHOULD match [80ch, 120ch]")
	}
	if !cond.Matches(120, 24, ThemeBootstrap) {
		t.Errorf("width 120 SHOULD match [80ch, 120ch]")
	}
	if cond.Matches(140, 24, ThemeBootstrap) {
		t.Errorf("width 140 should NOT match [80ch, 120ch]")
	}
}

func TestMedia_EndToEndResponsiveRender(t *testing.T) {
	css := `
.responsive-badge {
  padding: 0 1;
  background-color: #dc3545; /* Red default */
  color: #ffffff;
}

@media (min-width: 90ch) {
  .responsive-badge {
    background-color: #198754; /* Green on wide terminal (>=90 cols) */
  }
}

@media (max-width: 60ch) {
  .responsive-badge {
    background-color: #0d6efd; /* Blue on narrow terminal (<=60 cols) */
  }
}
`
	html := `<div class="responsive-badge">Status</div>`

	// 1. Narrow terminal (50 cols): should be Blue (#0d6efd -> 13;110;253)
	mNarrow := New(html, WithWidth(50), WithStylesheets(css))
	outNarrow, err := mNarrow.Render()
	if err != nil {
		t.Fatalf("Render failed for narrow: %v", err)
	}
	if !strings.Contains(outNarrow, "13;110;253") {
		t.Errorf("expected blue color (13;110;253) at width 50, got:\n%q", outNarrow)
	}

	// 2. Medium terminal (75 cols): should be Red (#dc3545 -> 220;53;69)
	mMedium := New(html, WithWidth(75), WithStylesheets(css))
	outMedium, err := mMedium.Render()
	if err != nil {
		t.Fatalf("Render failed for medium: %v", err)
	}
	if !strings.Contains(outMedium, "220;53;69") {
		t.Errorf("expected red color (220;53;69) at width 75, got:\n%q", outMedium)
	}

	// 3. Wide terminal (100 cols): should be Green (#198754 -> 25;135;84)
	mWide := New(html, WithWidth(100), WithStylesheets(css))
	outWide, err := mWide.Render()
	if err != nil {
		t.Fatalf("Render failed for wide: %v", err)
	}
	if !strings.Contains(outWide, "25;135;84") {
		t.Errorf("expected green color (25;135;84) at width 100, got:\n%q", outWide)
	}
}

func TestMedia_DisplayNoneResponsive(t *testing.T) {
	css := `
.sidebar {
  padding: 0 1;
  background-color: #343a40;
  color: #ffffff;
}
@media (max-width: 60ch) {
  .sidebar {
    display: none;
  }
}
`
	html := `<div><div class="sidebar">Sidebar Content</div><p>Main Body</p></div>`

	// Width 50: sidebar hidden
	mSmall := New(html, WithWidth(50), WithStylesheets(css))
	outSmall, err := mSmall.Render()
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}
	if strings.Contains(outSmall, "Sidebar Content") {
		t.Errorf("expected sidebar to be hidden at width 50, got:\n%s", outSmall)
	}
	if !strings.Contains(outSmall, "Main Body") {
		t.Errorf("expected 'Main Body' to remain visible at width 50, got:\n%s", outSmall)
	}

	// Width 80: sidebar visible
	mNormal := New(html, WithWidth(80), WithStylesheets(css))
	outNormal, err := mNormal.Render()
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}
	if !strings.Contains(outNormal, "Sidebar Content") {
		t.Errorf("expected sidebar to be visible at width 80, got:\n%s", outNormal)
	}
}
