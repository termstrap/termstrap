package termstrap

import (
	"strings"
	"testing"
)

func TestSpecificity_W3CExamples(t *testing.T) {
	tests := []struct {
		selector string
		expected Specificity
	}{
		{"*", Specificity{A: 0, B: 0, C: 0}},
		{"li", Specificity{A: 0, B: 0, C: 1}},
		{"ul li", Specificity{A: 0, B: 0, C: 2}},
		{"ul ol+li", Specificity{A: 0, B: 0, C: 3}},
		{"h1 + *[rel=up]", Specificity{A: 0, B: 1, C: 1}},
		{"ul ol li.red", Specificity{A: 0, B: 1, C: 3}},
		{"li.red.level", Specificity{A: 0, B: 2, C: 1}},
		{"#x34y", Specificity{A: 1, B: 0, C: 0}},
		{"#s12:not(FOO)", Specificity{A: 1, B: 0, C: 1}},
		{".card > .card-header #title.active", Specificity{A: 1, B: 3, C: 0}},
		{"::before", Specificity{A: 0, B: 0, C: 1}},
	}

	for _, tt := range tests {
		t.Run(tt.selector, func(t *testing.T) {
			got := ComputeSpecificity(tt.selector)
			if got != tt.expected {
				t.Errorf("ComputeSpecificity(%q) = %+v, expected %+v", tt.selector, got, tt.expected)
			}
		})
	}
}

func TestSpecificity_Comparison(t *testing.T) {
	id := Specificity{A: 1, B: 0, C: 0}
	manyClasses := Specificity{A: 0, B: 10, C: 10}
	twoClasses := Specificity{A: 0, B: 2, C: 0}
	oneClassFiveTags := Specificity{A: 0, B: 1, C: 5}

	if !manyClasses.Less(id) {
		t.Errorf("expected (0, 10, 10) < (1, 0, 0)")
	}
	if !oneClassFiveTags.Less(twoClasses) {
		t.Errorf("expected (0, 1, 5) < (0, 2, 0)")
	}
}

func TestSpecificity_CascadeDocumentOrderOverride(t *testing.T) {
	// In document order, 'div' and '.target-class' appear AFTER '#target'.
	// But '#target' has higher specificity (1, 0, 0) vs (0, 1, 0) and (0, 0, 1).
	// Therefore, '#target' MUST win and color must be Blue (#0d6efd -> 13;110;253).
	css := `
#target {
  color: #0d6efd; /* Blue (A=1) */
}
.target-class {
  color: #dc3545; /* Red (B=1) */
}
div {
  color: #ffc107; /* Yellow (C=1) */
}
`
	html := `<div id="target" class="target-class">Specificity Winner</div>`

	m := New(html, WithWidth(60), WithStylesheets(css))
	out, err := m.Render()
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	if !strings.Contains(out, "13;110;253") {
		t.Errorf("expected ID selector #target (Blue: 13;110;253) to win over later class and tag rules, got:\n%q", out)
	}
}

func TestSpecificity_ImportantOverridesHigherSpecificity(t *testing.T) {
	// '#target' has higher specificity (1, 0, 0), but '.target-class' has !important.
	// Therefore, '.target-class' (!important) MUST win and color must be Green (#198754 -> 25;135;84).
	css := `
#target {
  color: #0d6efd; /* Blue (A=1, normal) */
}
.target-class {
  color: #198754 !important; /* Green (B=1, !important) */
}
`
	html := `<div id="target" class="target-class">Important Winner</div>`

	m := New(html, WithWidth(60), WithStylesheets(css))
	out, err := m.Render()
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	if !strings.Contains(out, "25;135;84") {
		t.Errorf("expected !important class (Green: 25;135;84) to win over normal #target, got:\n%q", out)
	}
}

func TestSpecificity_InlineStyleCascade(t *testing.T) {
	t.Run("Normal inline style beats normal stylesheet rule", func(t *testing.T) {
		css := `.btn { color: #0d6efd; }` // Blue
		html := `<div class="btn" style="color: #ffc107;">Inline Yellow</div>` // Yellow (255;193;7)

		m := New(html, WithWidth(60), WithStylesheets(css))
		out, err := m.Render()
		if err != nil {
			t.Fatalf("Render failed: %v", err)
		}
		if !strings.Contains(out, "255;193;7") {
			t.Errorf("expected normal inline style (Yellow: 255;193;7) to win, got:\n%q", out)
		}
	})

	t.Run("Stylesheet !important beats normal inline style", func(t *testing.T) {
		css := `.btn { color: #dc3545 !important; }` // Red !important
		html := `<div class="btn" style="color: #ffc107;">Important Stylesheet</div>`

		m := New(html, WithWidth(60), WithStylesheets(css))
		out, err := m.Render()
		if err != nil {
			t.Fatalf("Render failed: %v", err)
		}
		if !strings.Contains(out, "220;52;69") {
			t.Errorf("expected stylesheet !important (Red: 220;52;69) to win over normal inline style, got:\n%q", out)
		}
	})

	t.Run("Inline style !important beats stylesheet !important", func(t *testing.T) {
		css := `.btn { color: #dc3545 !important; }` // Red !important
		html := `<div class="btn" style="color: #0dcaf0 !important;">Inline Important Cyan</div>` // Cyan (13;202;240)

		m := New(html, WithWidth(60), WithStylesheets(css))
		out, err := m.Render()
		if err != nil {
			t.Fatalf("Render failed: %v", err)
		}
		if !strings.Contains(out, "13;202;240") {
			t.Errorf("expected inline !important (Cyan: 13;202;240) to win over stylesheet !important, got:\n%q", out)
		}
	})
}
