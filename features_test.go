package termstrap

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
)

// sgrParams returns the set of numeric SGR parameters present anywhere in s.
func sgrParams(s string) map[string]bool {
	out := map[string]bool{}
	for _, m := range ansiRe.FindAllStringSubmatch(s, -1) {
		for _, p := range strings.Split(m[1], ";") {
			if p != "" {
				out[p] = true
				if strings.Contains(p, ":") {
					for _, sub := range strings.Split(p, ":") {
						out[sub] = true
					}
				}
			}
		}
	}
	return out
}

func renderTC(t *testing.T, html string, width int) string {
	t.Helper()
	lipgloss.Writer.Profile = colorprofile.TrueColor
	out, err := New(html, WithWidth(width)).Render()
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	return out
}

func TestTextDecorations(t *testing.T) {
	tests := []struct {
		name string
		html string
		sgr  string // expected SGR param
	}{
		{"strikethrough", `<p style="text-decoration: line-through;">x</p>`, "9"},
		{"strike-tag", `<p><s>x</s></p>`, "9"},
		{"del-tag", `<p><del>x</del></p>`, "9"},
		{"faint-opacity", `<p style="opacity: 0.5;">x</p>`, "2"},
		{"blink", `<p style="text-decoration: blink;">x</p>`, "5"},
		{"blink-tag", `<p><blink>x</blink></p>`, "5"},
		{"underline", `<p style="text-decoration: underline;">x</p>`, "4"},
		{"underline-double", `<p style="text-decoration: underline; text-decoration-style: double;">x</p>`, "4:2"},
		{"underline-curly", `<p style="text-decoration: underline; text-decoration-style: wavy;">x</p>`, "4:3"},
		{"underline-dotted", `<p style="text-decoration: underline; text-decoration-style: dotted;">x</p>`, "4:4"},
		{"underline-dashed", `<p style="text-decoration: underline; text-decoration-style: dashed;">x</p>`, "4:5"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := renderTC(t, tt.html, 20)
			if !sgrParams(out)[tt.sgr] {
				t.Errorf("expected SGR %q in output, got params %v\n%q", tt.sgr, sgrParams(out), out)
			}
		})
	}
}

func TestBorderStyles(t *testing.T) {
	tests := []struct {
		name  string
		html  string
		glyph string
	}{
		{"double", `<div style="border: 1px double;">x</div>`, "═"},
		{"dashed", `<div style="border: 1px dashed;">x</div>`, "╌"},
		{"dotted", `<div style="border: 1px dotted;">x</div>`, "┄"},
		{"thick", `<div style="border: 3px solid;">x</div>`, "┃"},
		{"block", `<div style="border: 1px solid; border-style: block;">x</div>`, "█"},
		{"markdown", `<div style="border: 1px solid; border-style: markdown;">x</div>`, "-"},
		{"ascii", `<div style="border: 1px solid; border-style: ascii;">x</div>`, "+"},
		{"rounded", `<div style="border: 1px solid; border-radius: 4px;">x</div>`, "╭"},
		{"normal", `<div style="border: 1px solid;">x</div>`, "┌"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := renderTC(t, tt.html, 20)
			if !strings.Contains(out, tt.glyph) {
				t.Errorf("expected border glyph %q in output:\n%s", tt.glyph, out)
			}
		})
	}
}

func TestPerSideBorderColor(t *testing.T) {
	// A red top border and a blue bottom border must emit two distinct colors.
	out := renderTC(t, `<div style="border: 1px solid; border-top-color: #ff0000; border-bottom-color: #0000ff;">x</div>`, 20)
	p := sgrParams(out)
	if !p["255"] || !p["0"] {
		t.Fatalf("expected per-side RGB colors in output, got %v", p)
	}
	// Both a red (38/48;2;255;0;0) and blue (…0;0;255) sequence should be present.
	if !strings.Contains(out, "255;0;0") || !strings.Contains(out, "0;0;255") {
		t.Errorf("expected both red and blue border colors:\n%q", out)
	}
}
func TestUnderlineColor(t *testing.T) {
	out := renderTC(t, `<p style="text-decoration: underline; text-decoration-color: #ff00ff;">colored underline</p>`, 30)
	// Expect SGR 58;2;255;0;255 (ANSI underline color sequence)
	if !strings.Contains(out, "58;2;255;0;255") {
		t.Errorf("expected underline color sequence (58;2;255;0;255) in output:\n%q", out)
	}
}

func TestBorderBlend(t *testing.T) {
	out := renderTC(t, `<div style="border: 1px solid; border-image: #ff0000 #00ff00;">blended border</div>`, 30)
	// Expect ANSI RGB color sequences for blended border edges
	if !strings.Contains(out, "38;2;") {
		t.Errorf("expected 24-bit ANSI colors for blended border, got:\n%q", out)
	}
}
func TestBackgroundGradient(t *testing.T) {
	out := renderTC(t, `<div style="background: linear-gradient(#ff0000, #0000ff); width: 20px;">gradient text</div>`, 20)
	// Expect 24-bit background color escape sequences
	if !strings.Contains(out, "48;2;") {
		t.Errorf("expected 24-bit background gradient in output:\n%q", out)
	}
}

func TestColorTransformations(t *testing.T) {
	// Test lighten, darken, complementary
	out := renderTC(t, `<div style="color: lighten(#000000, 50%); background-color: complementary(#0000ff);">test</div>`, 20)
	if !strings.Contains(out, "38;2;") || !strings.Contains(out, "48;2;") {
		t.Errorf("expected transformed colors in output:\n%q", out)
	}
}

func TestMaxWidthClamp(t *testing.T) {
	clamped := renderTC(t, `<div style="max-width: 20px; border: 1px solid;">hi</div>`, 40)
	for _, l := range strings.Split(clamped, "\n") {
		if w := lipgloss.Width(strings.TrimRight(l, " ")); w > 20 {
			t.Errorf("box exceeds max-width 20: width=%d: %q", w, l)
		}
	}
	if !strings.Contains(clamped, "┌") {
		t.Fatalf("expected a border in clamped output:\n%s", clamped)
	}
}

func TestFlexJustifyContent(t *testing.T) {
	row := func(jc string) string {
		return renderTC(t, `<div class="d-flex" style="display:flex; justify-content:`+jc+`;"><div class="col-4 border">C</div></div>`, 40)
	}
	leftIndent := func(out string) int {
		for _, l := range strings.Split(out, "\n") {
			if strings.Contains(l, "┌") {
				return len(l) - len(strings.TrimLeft(l, " "))
			}
		}
		return -1
	}
	start := leftIndent(row("flex-start"))
	center := leftIndent(row("center"))
	end := leftIndent(row("flex-end"))
	if !(start < center && center < end) {
		t.Errorf("justify-content indents not ordered start<center<end: %d %d %d", start, center, end)
	}
}

func TestFlexAlignItems(t *testing.T) {
	// A short column next to a 3-line column: align-items-end pushes the short
	// box lower than align-items-start.
	html := func(ai string) string {
		return `<div class="d-flex ` + ai + `"><div class="col-6 border">a<br>b<br>c</div><div class="col-6 border">x</div></div>`
	}
	firstBoxRow := func(out string) int {
		// row index where the second column's top-left corner appears after col 18
		lines := strings.Split(out, "\n")
		for y, l := range lines {
			r := []rune(l)
			for x := 18; x < len(r); x++ {
				if r[x] == '┌' {
					return y
				}
			}
		}
		return -1
	}
	start := firstBoxRow(renderTC(t, html("align-items-start"), 40))
	end := firstBoxRow(renderTC(t, html("align-items-end"), 40))
	if !(end > start) {
		t.Errorf("align-items-end should place the short box lower: start=%d end=%d", start, end)
	}
}

func TestListRendering(t *testing.T) {
	out := renderTC(t, `<ul><li>Alpha</li><li>Beta</li></ul>`, 30)
	if !strings.Contains(out, "• Alpha") || !strings.Contains(out, "• Beta") {
		t.Errorf("expected bulleted list items:\n%s", out)
	}

	ol := renderTC(t, `<ol><li>First</li><li>Second</li></ol>`, 30)
	if !strings.Contains(ol, "1. First") || !strings.Contains(ol, "2. Second") {
		t.Errorf("expected numbered list items:\n%s", ol)
	}

	nested := renderTC(t, `<ul><li>Top<ul><li>Child</li></ul></li></ul>`, 30)
	if !strings.Contains(nested, "Top") || !strings.Contains(nested, "Child") {
		t.Errorf("expected nested list items:\n%s", nested)
	}
	// Nested item must be indented further than its parent.
	topCol := indentOf(nested, "Top")
	childCol := indentOf(nested, "Child")
	if childCol <= topCol {
		t.Errorf("nested item not indented: top=%d child=%d", topCol, childCol)
	}
}

func TestListStyleType(t *testing.T) {
	roman := renderTC(t, `<ul style="list-style-type: upper-roman;"><li>One</li><li>Two</li></ul>`, 30)
	if !strings.Contains(roman, "I.") || !strings.Contains(roman, "II.") {
		t.Errorf("expected roman enumerators:\n%s", roman)
	}
	none := renderTC(t, `<ul style="list-style: none;"><li>Bare</li></ul>`, 30)
	if strings.Contains(none, "•") {
		t.Errorf("list-style:none must not emit bullets:\n%s", none)
	}
}

func TestTreeRendering(t *testing.T) {
	out := renderTC(t, `<ul class="tree"><li>root<ul><li>a</li><li>b</li></ul></li></ul>`, 30)
	if !strings.Contains(out, "root") || !strings.Contains(out, "a") || !strings.Contains(out, "b") {
		t.Errorf("expected tree labels:\n%s", out)
	}
	if !strings.Contains(out, "├──") && !strings.Contains(out, "└──") {
		t.Errorf("expected tree branch connectors:\n%s", out)
	}
	// Tree must not introduce an empty-root node line.
	for _, l := range strings.Split(out, "\n") {
		trimmed := strings.TrimSpace(stripANSI(l))
		if trimmed == "└──" || trimmed == "├──" {
			t.Errorf("empty-label tree node found:\n%s", out)
		}
	}
}

func TestTableRendering(t *testing.T) {
	html := `<table class="table">
	  <thead><tr><th>ID</th><th>Name</th></tr></thead>
	  <tbody>
	    <tr><td>1</td><td>Alice</td></tr>
	    <tr><td>2</td><td>Bob</td></tr>
	  </tbody>
	</table>`
	out := renderTC(t, html, 40)
	for _, want := range []string{"ID", "Name", "Alice", "Bob"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in table:\n%s", want, out)
		}
	}
	// Boxed table with a header separator.
	if !strings.Contains(out, "┌") || !strings.Contains(out, "├") {
		t.Errorf("expected table box and header separator:\n%s", out)
	}
}

func TestTableRowBackground(t *testing.T) {
	html := `<table class="table"><tbody><tr style="background-color:#0d6efd;"><td>row</td></tr></tbody></table>`
	out := renderTC(t, html, 30)
	// A colored row must emit a background SGR (48;2;...).
	if !strings.Contains(out, "48;2;") {
		t.Errorf("expected background color on .table-primary row:\n%q", out)
	}
}

// indentOf returns the leading-space count of the first line containing sub.
func indentOf(out, sub string) int {
	for _, l := range strings.Split(out, "\n") {
		plain := ansiRe.ReplaceAllString(l, "")
		if strings.Contains(plain, sub) {
			return len(plain) - len(strings.TrimLeft(plain, " "))
		}
	}
	return -1
}
