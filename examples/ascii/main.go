// Example: ascii — Demonstrates rendering FIGlet ASCII art text banners
// using the declarative <ascii> HTML tag in Termstrap.
//
// Usage:
//
//	go run ./examples/ascii/
package main

import (
	"fmt"
	"os"

	"github.com/go-scripts/termstrap"
	"golang.org/x/term"
)

func main() {
	width := 80
	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && w > 0 {
		width = w
	}

	html := `
<div class="container-fluid p-1">
  <!-- 1. En-tête centré avec bordure collante et animation border-shimmer + shimmer -->
  <div class="w-100 text-center mb-1">
    <ascii class="shimmer border border-shimmer" color="#7aa2f7">Hubs.sh</ascii>
    <div class="text-muted mt-1">Bordure ajustée au plus près + effet border-shimmer synchronisé</div>
  </div>

  <div class="row">
    <!-- 2. Police "slant" avec bordure arrondie et text-warning -->
    <div class="col-12 col-md-6 border rounded p-1 mb-1">
      <div class="text-warning fw-bold mb-1">Police "slant" (border rounded) :</div>
      <ascii font="slant" color="#ff9e64" class="border rounded p-1">Hello</ascii>
    </div>

    <!-- 3. Police "doom" avec border-shimmer et text-info -->
    <div class="col-12 col-md-6 border rounded p-1 mb-1">
      <div class="text-info fw-bold mb-1">Police "doom" (border-shimmer) :</div>
      <ascii font="doom" color="#7dcfff" class="border border-shimmer shimmer">World</ascii>
    </div>
  </div>

  <!-- 4. Démonstration de diverses polices FIGlet populaires -->
  <div class="row">
    <div class="col-12 col-md-4 border rounded p-1 mb-1">
      <div class="text-warning fw-bold mb-1">Police "block" :</div>
      <ascii font="block" color="#f7768e">BLOCK</ascii>
    </div>
    <div class="col-12 col-md-4 border rounded p-1 mb-1">
      <div class="text-success fw-bold mb-1">Police "small" :</div>
      <ascii font="small" color="#9ece6a">SMALL</ascii>
    </div>
    <div class="col-12 col-md-4 border rounded p-1 mb-1">
      <div class="text-primary fw-bold mb-1">Police "digital" :</div>
      <ascii font="digital" color="#7aa2f7">DIGI</ascii>
    </div>
  </div>

  <!-- 5. Bannière pleine largeur 100% avec fond sombre et dégradé -->
  <div class="border rounded p-1 bg-dark text-center" style="border-color: #9ece6a;">
    <ascii width="100%" color="#9ece6a" class="shimmer">Termstrap</ascii>
    <p class="text-success mb-0 mt-1">Plus de 140 polices FIGlet disponibles via <code>termstrap.AvailableAsciiFonts()</code></p>
  </div>
</div>
`

	m := termstrap.New(
		html,
		termstrap.WithWidth(width),
		termstrap.WithTheme(termstrap.ThemeTokyoNight),
	)

	rendered, err := m.Render()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erreur de rendu : %v\n", err)
		os.Exit(1)
	}

	fmt.Println(rendered)
}
