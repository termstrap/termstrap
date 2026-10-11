// Example: ascii-all-fonts — Showcase all 140+ FIGlet ASCII fonts supported by Termstrap.
//
// Usage:
//
//	go run ./examples/ascii-all-fonts/
//	go run ./examples/ascii-all-fonts/ "MonTexte"
package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/go-scripts/termstrap"
	"golang.org/x/term"
)

func main() {
	width := 90
	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && w > 0 {
		width = w
	}

	word := "Hubs.sh"
	if len(os.Args) > 1 && os.Args[1] != "" {
		word = os.Args[1]
	}

	fonts := termstrap.AvailableAsciiFonts()
	sort.Strings(fonts)

	fmt.Printf("\033[1;36m================================================================================\033[0m\n")
	fmt.Printf("\033[1;32m  Termstrap ASCII Fonts Showcase (%d polices disponibles)\033[0m\n", len(fonts))
	fmt.Printf("  Texte affiché : %q | Largeur : %d colonnes\n", word, width)
	fmt.Printf("\033[1;36m================================================================================\033[0m\n\n")

	for i, f := range fonts {
		html := fmt.Sprintf(`
<div class="border rounded p-1 mb-1">
  <div class="text-warning fw-bold mb-1">#%d — Police: <span class="text-info">%s</span></div>
  <div class="text-center">
    <ascii font="%s" color="#7aa2f7" class="border p-1">%s</ascii>
  </div>
</div>`, i+1, f, f, word)

		m := termstrap.New(
			html,
			termstrap.WithWidth(width),
			termstrap.WithTheme(termstrap.ThemeTokyoNight),
		)

		rendered, err := m.Render()
		if err != nil {
			fmt.Printf("Erreur pour la police %q: %v\n", f, err)
			continue
		}

		fmt.Println(rendered)
	}

	fmt.Printf("\n\033[1;32m✔ %d polices affichées avec succès !\033[0m\n", len(fonts))
}
