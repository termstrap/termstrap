package termstrap

import (
	"regexp"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
)

var ansiRe = regexp.MustCompile(`\x1b\[([0-9;:]*)m`)

// cellBgStates walks a rendered line and returns, per visible cell, the glyph
// and whether the card background is active on that cell.
func cellBgStates(line, bgCode string) []struct {
	ch string
	bg bool
} {
	var cells []struct {
		ch string
		bg bool
	}
	bg := false
	i := 0
	for i < len(line) {
		if loc := ansiRe.FindStringIndex(line[i:]); loc != nil && loc[0] == 0 {
			codes := ansiRe.FindStringSubmatch(line[i:])[1]
			if codes == "" {
				bg = false
			} else {
				for _, c := range strings.Split(codes, ";") {
					if c == "0" || c == "49" {
						bg = false
					}
				}
			}
			if strings.Contains(codes, bgCode) {
				bg = true
			}
			i += loc[1]
			continue
		}
		r := []rune(line[i:])
		ch := string(r[0])
		cells = append(cells, struct {
			ch string
			bg bool
		}{ch, bg})
		i += len(ch)
	}
	return cells
}

// TestBorderBackgroundCoversBorderBox verifies CSS border-box semantics: the
// element background extends under the border, so every cell from the left
// border glyph through the right border glyph (inclusive) carries the
// background, and no cell outside that span does (no bleed past the box). This
// prevents the dark seams that appear where colored boxes meet when border
// cells render on the terminal's default background.
func TestBorderBackgroundCoversBorderBox(t *testing.T) {
	lipgloss.Writer.Profile = colorprofile.TrueColor

	html := `<div class="border rounded p-2 bg-primary text-white">
  <h4 class="mb-1">CPU Load</h4>
  <p class="mb-0">8 Cores Active</p>
</div>`

	out, err := New(html, WithWidth(40)).Render()
	if err != nil {
		t.Fatal(err)
	}

	const bgCode = "48;2;13;110;253" // bg-primary #0d6efd
	borderGlyphs := "│╭╮╰╯┌┐└┘─━═┃║"
	sawBox := false

	for i, line := range strings.Split(out, "\n") {
		if !strings.ContainsAny(line, borderGlyphs) {
			continue
		}
		cells := cellBgStates(line, bgCode)

		first, last := -1, -1
		for idx, c := range cells {
			if strings.ContainsAny(c.ch, borderGlyphs) {
				if first < 0 {
					first = idx
				}
				last = idx
			}
		}
		if first < 0 || last <= first {
			continue
		}
		sawBox = true

		// Whole border box span (border cells included) must be colored.
		for idx := first; idx <= last; idx++ {
			if !cells[idx].bg {
				t.Fatalf("line %d: uncolored cell %d (%q) inside border box: %q", i, idx, cells[idx].ch, line)
			}
		}
		// No bleed: nothing outside the border box carries the background.
		for idx := 0; idx < first; idx++ {
			if cells[idx].bg {
				t.Fatalf("line %d: background bleeds left of border at cell %d: %q", i, idx, line)
			}
		}
		for idx := last + 1; idx < len(cells); idx++ {
			if cells[idx].bg {
				t.Fatalf("line %d: background bleeds right of border at cell %d: %q", i, idx, line)
			}
		}
	}
	if !sawBox {
		t.Fatal("expected a bordered card, found none")
	}
}
