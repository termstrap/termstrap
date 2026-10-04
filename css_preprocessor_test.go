package termstrap

import (
	"strings"
	"testing"
)

func TestPreprocessor_Variables(t *testing.T) {
	tests := []struct {
		name     string
		cssInput string
		varQuery string
		expected string
	}{
		{
			name: "Basic root variables",
			cssInput: `:root {
				--bs-blue: #0d6efd;
				--bs-primary: var(--bs-blue);
			}`,
			varQuery: "--bs-primary",
			expected: "#0d6efd",
		},
		{
			name: "Multi-level variable dependencies",
			cssInput: `:root {
				--c: #198754;
				--b: var(--c);
				--a: var(--b);
			}`,
			varQuery: "--a",
			expected: "#198754",
		},
		{
			name: "Nested fallback",
			cssInput: `:root {
				--res: var(--missing, var(--also-missing, #ffc107));
			}`,
			varQuery: "--res",
			expected: "#ffc107",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := NewCSSMatcher(ThemeBootstrap, tt.cssInput)
			if err != nil {
				t.Fatalf("NewCSSMatcher failed: %v", err)
			}
			val := m.Preprocessor.GlobalVars[tt.varQuery]
			if val != tt.expected {
				t.Errorf("expected %q for %s, got %q", tt.expected, tt.varQuery, val)
			}
		})
	}
}

func TestPreprocessor_Calc(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"calc(1.5rem * 0.5)", "0.75rem"},
		{"calc(-1 * 0rem)", "0rem"},
		{"calc(10px + 20px)", "30px"},
		{"calc(40px - 15px)", "25px"},
		{"calc(2 * 4px)", "8px"},
		{"calc(10px / 2)", "5px"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := EvaluateCalc(tt.input)
			if got != tt.expected {
				t.Errorf("EvaluateCalc(%q) = %q, expected %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestPreprocessor_RGBA(t *testing.T) {
	tests := []struct {
		input    string
		bgHex    string
		expected string
	}{
		{"rgba(13, 110, 253, 1)", "#ffffff", "#0d6efd"},
		{"rgb(13, 110, 253)", "#ffffff", "#0d6efd"},
		{"rgb(13 110 253)", "#ffffff", "#0d6efd"},
		{"rgba(13, 110, 253, 100%)", "#ffffff", "#0d6efd"},
		{"rgba(0, 0, 0, 0)", "#ffffff", "transparent"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ResolveColors(tt.input, tt.bgHex)
			if strings.ToLower(got) != strings.ToLower(tt.expected) {
				t.Errorf("ResolveColors(%q, %q) = %q, expected %q", tt.input, tt.bgHex, got, tt.expected)
			}
		})
	}
}

func TestPreprocessor_BootstrapPatterns(t *testing.T) {
	cssBootstrapSnippet := `
:root {
  --bs-primary: #0d6efd;
  --bs-primary-rgb: 13, 110, 253;
  --bs-border-width: 1px;
  --bs-border-style: solid;
  --bs-border-color: #dee2e6;
  --bs-border-radius: 0.375rem;
}

.custom-card {
  --bs-bg-opacity: 1;
  background-color: rgba(var(--bs-primary-rgb), var(--bs-bg-opacity)) !important;
  color: #ffffff;
  border: var(--bs-border-width) var(--bs-border-style) var(--bs-border-color) !important;
  border-radius: var(--bs-border-radius) !important;
}
`
	html := `<div class="custom-card p-1">Bootstrap Variable Test</div>`

	m := New(html, WithWidth(80), WithStylesheets(cssBootstrapSnippet))
	out, err := m.Render()
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	// #0d6efd in ANSI TrueColor is 48;2;13;110;253
	expectedSeq := "48;2;13;110;253"
	if !strings.Contains(out, expectedSeq) {
		t.Errorf("expected ANSI TrueColor sequence %q for background-color, got output:\n%q", expectedSeq, out)
	}
	if !strings.Contains(out, "Bootstrap Variable Test") {
		t.Errorf("expected content 'Bootstrap Variable Test' in output, got:\n%s", out)
	}
}

func TestPreprocessor_InlineStyleVariables(t *testing.T) {
	customCSS := `:root { --highlight: #ffc107; }`
	html := `<div style="background-color: var(--highlight); color: #000000;">Inline Var</div>`

	m := New(html, WithWidth(60), WithStylesheets(customCSS))
	out, err := m.Render()
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	// #ffc107 in ANSI TrueColor is 48;2;255;193;7
	expectedSeq := "48;2;255;193;7"
	if !strings.Contains(out, expectedSeq) {
		t.Errorf("expected ANSI TrueColor sequence %q for inline style variable, got output:\n%q", expectedSeq, out)
	}
}
