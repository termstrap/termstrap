package termstrap

import (
	"strings"
	"testing"
)

func TestRenderStickyRegions_InlineWithOffsets(t *testing.T) {
	html := `<div class="p-1"><div class="position-sticky"><h1>ONE</h1></div><div><p>alpha body</p></div></div>` +
		`<div class="p-1"><div class="position-sticky"><h1>TWO</h1></div><div><p>omega body</p></div></div>`

	body, regions, err := Model{HTML: html, Theme: ThemeTokyoNight, Width: 40, DisableSyntaxHighlight: true}.RenderStickyRegions()
	if err != nil {
		t.Fatal(err)
	}
	if len(regions) != 2 {
		t.Fatalf("expected 2 sticky regions, got %d", len(regions))
	}
	// Both sticky headers stay inline in the scrollable body.
	b := stripANSI(body)
	for _, want := range []string{"ONE", "alpha body", "TWO", "omega body"} {
		if !strings.Contains(b, want) {
			t.Fatalf("body missing %q: %q", want, b)
		}
	}
	// Offsets are monotonic and the reported header line actually contains the title.
	bodyLines := strings.Split(body, "\n")
	for i, r := range regions {
		if r.StartLine < 0 || r.StartLine >= len(bodyLines) {
			t.Fatalf("region %d StartLine %d out of range (body %d lines)", i, r.StartLine, len(bodyLines))
		}
		if r.Height <= 0 || r.Header == "" {
			t.Fatalf("region %d has empty header", i)
		}
	}
	if regions[1].StartLine <= regions[0].StartLine {
		t.Fatalf("expected increasing offsets, got %d then %d", regions[0].StartLine, regions[1].StartLine)
	}
	// The second region's header should sit at its reported offset in the body.
	if got := stripANSI(strings.Join(bodyLines[regions[1].StartLine:regions[1].StartLine+regions[1].Height], " ")); !strings.Contains(got, "TWO") {
		t.Fatalf("region 1 header not found at offset %d: %q", regions[1].StartLine, got)
	}
}

func TestRenderStickyRegions_NoSticky(t *testing.T) {
	body, regions, err := Model{HTML: `<div><p>plain content</p></div>`, Width: 40, DisableSyntaxHighlight: true}.RenderStickyRegions()
	if err != nil {
		t.Fatal(err)
	}
	if regions != nil {
		t.Fatalf("expected no regions, got %d", len(regions))
	}
	if !strings.Contains(stripANSI(body), "plain content") {
		t.Fatalf("body missing content: %q", stripANSI(body))
	}
}

func TestPinSticky_ActivationAndPush(t *testing.T) {
	regions := []StickyRegion{
		{StartLine: 0, Height: 1, Header: "AAA"},
		{StartLine: 10, Height: 1, Header: "BBB"},
	}
	rows := func(n int) string {
		ls := make([]string, n)
		for i := range ls {
			ls[i] = "row"
		}
		return strings.Join(ls, "\n")
	}
	first := func(s string) string { return strings.SplitN(s, "\n", 2)[0] }

	// At the very top the first section's header sits naturally on row 0 — not pinned.
	if got := first(PinSticky(rows(5), regions, 0)); got != "row" {
		t.Fatalf("yoff=0 should not pin, got %q", got)
	}
	// Once section 0 scrolls above the top, its header pins.
	if got := first(PinSticky(rows(5), regions, 5)); got != "AAA" {
		t.Fatalf("yoff=5 want pinned AAA, got %q", got)
	}
	// Once section 1 scrolls above the top, its header replaces section 0's.
	if got := first(PinSticky(rows(5), regions, 12)); got != "BBB" {
		t.Fatalf("yoff=12 want pinned BBB, got %q", got)
	}
	// Push-off: a 2-line header whose successor is 1 line below the top slides up.
	pushed := []StickyRegion{{0, 2, "A0\nA1"}, {10, 2, "B0\nB1"}}
	if got := first(PinSticky(rows(5), pushed, 9)); got != "A1" {
		t.Fatalf("push want A1 on row 0, got %q", got)
	}
	// No regions: view passes through untouched.
	if got := PinSticky(rows(3), nil, 7); got != rows(3) {
		t.Fatalf("nil regions must pass through, got %q", got)
	}
}
