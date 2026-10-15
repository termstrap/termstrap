# Documentation des Fonctionnalités — Termstrap

Ce document constitue la référence exhaustive et à jour de toutes les fonctionnalités, balises HTML, classes CSS Bootstrap et capacités graphiques prises en charge par le moteur **Termstrap**.

---

## 1. Vue d'Ensemble de l'Architecture

Termstrap est un moteur de rendu **HTML/CSS natif vers terminal ANSI**. Il analyse l'arbre DOM HTML, résout les règles CSS (feuille de style Bootstrap intégrée + styles personnalisés) et calcule un modèle de boîte (*Box Model*) précis en cellules de caractères pour produire un affichage terminal soigné et responsive.

```
[ HTML + CSS Input ]
         │
         ▼
[ CSS Matcher (Cascadia + Douceur) ]
         │
         ▼
[ Layout Engine (Box Model 2D) ]
         │
         ▼
[ ANSI Render Engine (Lipgloss) ] ──► [ Text + Image Overlays ]
```

---

## 2. Grille & Flexbox Responsive (`.row` / `.col-*`)

Le système de grille repose sur une division en 12 colonnes avec calcul dynamique selon les seuils de largeur (*breakpoints*) du terminal.

### Breakpoints pris en charge

| Breakpoint | Préfixe de classe | Largeur minimale requise |
| :--- | :--- | :--- |
| **xs** (Extra Small) | `.col-*`, `.col-xs-*` | `< 70` colonnes |
| **sm** (Small) | `.col-sm-*` | $\ge 70$ colonnes |
| **md** (Medium) | `.col-md-*` | $\ge 100$ colonnes |
| **lg** (Large) | `.col-lg-*` | $\ge 130$ colonnes |
| **xl** (Extra Large) | `.col-xl-*` | $\ge 160$ colonnes |
| **xxl** (Extra Extra Large) | `.col-xxl-*` | $\ge 190$ colonnes |
### Comportement
- **Empilement automatique (*Stacking*)** : Si la largeur du terminal est inférieure au breakpoint spécifié (ex: un `.col-md-6` affiché sur un terminal de 50 colonnes), la colonne s'empile verticalement à 100% de la largeur.
- **Multi-breakpoints** : Vous pouvez cumuler plusieurs classes responsives (ex: `class="col-sm-12 col-md-6"`).
- **Imbrication (*Nesting*)** : Une `.row` peut être imbriquée à l'intérieur d'une `.col-*` à n'importe quel niveau de profondeur. Les largeurs sont réparties proportionnellement à la largeur de la colonne parente.

```html
<div class="row">
  <div class="col-sm-12 col-md-6 border p-1">
    <h3>Colonne Gauche</h3>
  </div>
  <div class="col-sm-12 col-md-6 border p-1">
    <h3>Colonne Droite</h3>
  </div>
</div>
```

---

## 3. Modèle de Boîte : Marges & Paddings

Les classes de marges et de paddings s'appliquent sur **n'importe quelle balise HTML** (`<div>`, `<p>`, `<img>`, `<table>`, `<span>`, etc.).

### Marges Externes (`m-*`)

| Type | Classes | Description |
| :--- | :--- | :--- |
| **Globales** | `.m-0` à `.m-5` | Marge sur les 4 côtés (de 0 à 5 cellules/lignes) |
| **Horizontales** | `.mx-0` à `.mx-5` | Marges gauche et droite |
| **Verticales** | `.my-0` à `.my-5` | Marges haut et bas |
| **Haut** | `.mt-0` à `.mt-5` | Marge supérieure |
| **Bas** | `.mb-0` à `.mb-5` | Marge inférieure |
| **Gauche / Start** | `.ms-0` à `.ms-5`, `.ml-0` à `.ml-5` | Marge gauche |
| **Droite / End** | `.me-0` à `.me-5`, `.mr-0` à `.mr-5` | Marge droite |

### Paddings Internes (`p-*`)

| Type | Classes | Description |
| :--- | :--- | :--- |
| **Globaux** | `.p-0` à `.p-5` | Espacement interne sur les 4 côtés |
| **Horizontaux** | `.px-0` à `.px-5` | Paddings gauche et droite |
| **Verticaux** | `.py-0` à `.py-5` | Paddings haut et bas |
| **Haut** | `.pt-0` à `.pt-5` | Padding supérieur |
| **Bas** | `.pb-0` à `.pb-5` | Padding inférieur |
| **Gauche / Start** | `.ps-0` à `.ps-5`, `.pl-0` à `.pl-5` | Padding gauche |
| **Droite / End** | `.pe-0` à `.pe-5`, `.pr-0` à `.pr-5` | Padding droit |

---

## 4. Bordures & Coins Arrondis

| Classe | Description |
| :--- | :--- |
| `.border` | Bordure rectangulaire complète sur les 4 côtés |
| `.rounded` | Coins arrondis utilisant les caractères Unicode `╭ ╮ ╰ ╯` |
| `.border-top` | Bordure supérieure uniquement |
| `.border-bottom` | Bordure inférieure uniquement |
| `.border-left`, `.border-start` | Bordure gauche uniquement |
| `.border-right`, `.border-end` | Bordure droite uniquement |

*Note : Les bordures partielles peuvent être librement combinées (ex: `class="border-top border-bottom"` ou `class="border-left border-right"`).*

---

## 5. Ombres Portées (*Box Shadows*)

| Classe | Niveau d'ombre | Effet visuel |
| :--- | :--- | :--- |
| `.shadow-sm` | Taille 1 | Ombre légère et discrète sous la boîte |
| `.shadow`, `.shadow-md` | Taille 2 | Ombre moyenne standard |
| `.shadow-lg` | Taille 3 | Ombre prononcée avec relief accentué |
| `.shadow-none` | Taille 0 | Supprime toute ombre |

*Système d'auto-clamping : Si une boîte approche de la limite droite du terminal, l'ombre se compresse automatiquement pour éviter tout débordement ou retour à la ligne forcé.*

---

## 6. Palette de Couleurs Bootstrap

Les couleurs sont prises en charge en **TrueColor 24-bit** (ou fallback ANSI 256).

### Couleurs d'Arrière-Plan (`bg-*`)
- `.bg-primary` (Bleu #0d6efd)
- `.bg-secondary` (Gris #6c757d)
- `.bg-success` (Vert #198754)
- `.bg-danger` (Rouge #dc3545)
- `.bg-warning` (Jaune #ffc107)
- `.bg-info` (Cyan #0dcaf0)
- `.bg-dark` (Noir/Anthracite #212529)
- `.bg-light` (Gris clair #f8f9fa)

### Couleurs de Texte (`text-*`)
- `.text-primary`, `.text-secondary`, `.text-success`, `.text-danger`, `.text-warning`, `.text-info`, `.text-dark`, `.text-light`, `.text-white`, `.text-muted`.

---

## 7. Typographie & Alignement

### Alignement Textuel
- `.text-start`, `.text-left` : Alignement à gauche (comportement par défaut).
- `.text-center` : Centrage du texte dans son conteneur.
- `.text-end`, `.text-right` : Alignement à droite.

### Styles de Police & Décorations Avancées
- `.fw-bold`, `.text-bold`, `<b>`, `<strong>` : Texte en gras.
- `.fst-italic`, `.text-italic`, `<i>`, `<em>` : Texte en italique.
- `.text-decoration-underline` : Texte souligné (`lipgloss.UnderlineSingle`).
- `text-decoration-style` : Variantes de soulignement ANSI :
  - `wavy`, `curly` : Vaguelettes (`lipgloss.UnderlineCurly` / ANSI `4:3`).
  - `double` : Double trait (`lipgloss.UnderlineDouble` / ANSI `4:2`).
  - `dotted` : Points (`lipgloss.UnderlineDotted` / ANSI `4:4`).
  - `dashed` : Tirets (`lipgloss.UnderlineDashed` / ANSI `4:5`).
- `text-decoration-color: <color>` : Couleur de soulignement dédiée (`SGR 58;2;r;g;b`).
- `text-decoration: line-through`, `.text-decoration-line-through`, `<s>`, `<del>`, `<strike>` : Texte barré (`lipgloss.Strikethrough`).
- `text-decoration: blink`, `<blink>` : Clignotement (`lipgloss.Blink` / ANSI SGR 5).
- `opacity: < 1` : Texte atténué (`lipgloss.Faint`), pour `.text-muted`-like et `<hr>`.

Toutes ces décorations se cumulent et s'héritent sur les enfants inline.
---

## 8. Balises HTML Prises en Charge

| Balise | Rendu visuel dans le terminal |
| :--- | :--- |
| `<h1>` à `<h6>` | Titres en gras avec espacements verticaux automatiques |
| `<p>` | Paragraphe de texte avec retour à la ligne automatique (*word-wrap*) |
| `<table>`, `<tr>`, `<th>`, `<td>` | Tableau complet avec colonnes alignées et bordures d'en-tête |
| `<blockquote>` | Citation avec bordure latérale gauche et retrait |
| `<code>` | Code inline sur fond sombre avec texte contrasté |
| `<pre>` | Bloc de code préformaté préservant l'indentation, avec **coloration syntaxique** automatique |
| `<hr>` | Ligne horizontale de séparation |
| `<ul>`, `<ol>`, `<li>` | Listes ordonnées et non-ordonnées |
| `<img>` | Image rendue en demi-blocs ANSI ou transmise en overlay natif |
| `<ascii>` | Bannière texte en ASCII Art FIGlet (`width`, `color`, `font`) |
### Coloration Syntaxique des Blocs de Code

Les blocs `<pre>` sont colorés syntaxiquement via [chroma](https://github.com/alecthomas/chroma) (activé par défaut).

- **Intégration la plus simple** (convention HTML/Markdown GitHub) :

  ```html
  <pre><code class="language-go">package main
  func main() {}</code></pre>
  ```

  Le préfixe de classe `language-xxx` (ou `lang-xxx`) sur le `<code>` interne — ou sur le `<pre>` — fixe le langage.

- **Détection automatique** : sans classe de langage, le langage est déduit du contenu (`lexers.Analyse`). Un `<pre><code>…</code></pre>` nu est donc coloré automatiquement.

- **Couleurs pilotées par CSS (charte du thème)** : la palette de coloration est définie par des variables CSS `--ts-code-*`. Les défauts (dans `css/termstrap.css`) s'accordent aux couleurs sémantiques Bootstrap sur le fond sombre du `<pre>` ; chaque thème (`dracula`, `tokyonight`, `glamour`) surcharge ces variables selon sa charte graphique dans son propre `css/theme-*.css`. On peut aussi les surcharger via `WithStylesheets(...)` :

  ```css
  :root {
    /* Surfaces des blocs de code */
    --ts-code-bg: #212529;         /* fond du bloc <pre>            */
    --ts-code-inline-bg: #343a40;  /* fond du <code> inline         */
    --ts-code-inline-fg: #ffc107;  /* texte du <code> inline        */

    /* Couleurs des tokens (coloration du <pre>) */
    --ts-code-name: #f8f9fa;       /* identifiants / texte par défaut (= couleur du <pre>) */
    --ts-code-keyword: #e685b5;    /* mots-clés                     */
    --ts-code-function: #6ea8fe;   /* noms de fonctions / méthodes  */
    --ts-code-type: #6edff6;       /* types / classes               */
    --ts-code-builtin: #c29ffa;    /* builtins                      */
    --ts-code-string: #75b798;     /* chaînes                       */
    --ts-code-number: #ffda6a;     /* nombres                       */
    --ts-code-comment: #9aa0a6;    /* commentaires                  */
    --ts-code-operator: #dee2e6;   /* opérateurs / ponctuation      */
    --ts-code-variable: #f8f9fa;   /* variables / propriétés        */
    --ts-code-constant: #c29ffa;   /* constantes / symboles         */
    --ts-code-tag: #e685b5;        /* balises markup (HTML/XML)     */
    --ts-code-attribute: #6ea8fe;  /* attributs / décorateurs       */
    --ts-code-regex: #63e6be;      /* expressions régulières        */
    --ts-code-escape: #ffbf69;     /* séquences d'échappement / interpolation */
    --ts-code-deleted: #ea868f;    /* suppressions diff / erreurs   */
    --ts-code-inserted: #75b798;   /* insertions diff               */
    --ts-code-heading: #6ea8fe;    /* titres markup                 */
  }
  ```

  `--ts-code-name` sert aussi de couleur de texte par défaut du `<pre>`. Toute variable non surchargée retombe sur le défaut Bootstrap (côté Go également, donc la coloration fonctionne même sans CSS).

- **Style chroma imposé (échappatoire)** : `WithSyntaxTheme("monokai")` force un style chroma nommé (`github-dark`, `nord`, `catppuccin-mocha`, `dracula`, …) et court-circuite les variables CSS.

- **Désactivation** : `WithSyntaxHighlight(false)`.

Les formateurs terminaux n'émettent que des couleurs d'avant-plan ; l'arrière-plan du bloc `<pre>` est préservé. La profondeur de couleur suit `WithColorMode` (TrueColor / 256 / 16).

- **Chrome du document piloté par CSS** : les surfaces non-code suivent la même logique de variables `--ts-*`, surchargées par thème et par `WithStylesheets(...)` :

  ```css
  :root {
    --ts-link-color: #6ea8fe;        /* couleur des liens <a>            */
    --ts-hr-color: #6c757d;          /* trait horizontal <hr>            */
    --ts-blockquote-color: #adb5bd;  /* texte des <blockquote>           */
    --ts-blockquote-border: #6c757d; /* barre latérale des <blockquote>  */
  }
  ```

- **Exemples** :
  - `examples/syntax/` — bloc multi-langages + diff (option `-theme`).
  - `examples/syntax-themes/` — même document rendu sur chaque thème pour comparer les chartes (code, liens, `<hr>`, `<blockquote>`).
  - `examples/syntax-custom/` — palette entièrement personnalisée via `WithStylesheets(...)`, sans thème (base Bootstrap).

---

### Hiérarchie des Titres (`<h1>`…`<h6>`)

Un terminal n'a pas de `font-size` : TermStrap différencie les niveaux par d'autres leviers, tous pilotés par CSS.

- **Échelle de couleur par niveau** : de `--ts-heading-1` (le plus clair) à `--ts-heading-6` (le plus atténué), surchargée par chaque thème selon sa charte (accent sur `h1`).
- **Graisse / style** : `h1`–`h4` en gras, `h5`/`h6` en italique.
- **Marqueurs Markdown** (par défaut) : chaque titre est encadré par `#`…`######` (selon le niveau) dans la teinte atténuée `--ts-heading-rule`, et la position du marqueur suit l'alignement du titre :
  - **centré** (`text-center`) : marqueur de **chaque côté** de chaque ligne — `## Titre ##` ;
  - **aligné à droite** (`text-end`) : marqueur **à droite** uniquement — `Titre ###` ;
  - **sinon** (gauche) : marqueur **à gauche** uniquement — `## Titre`.
- **`text-transform`** (nouveau dans le moteur) : `uppercase` / `lowercase` / `capitalize`, utilisable sur n'importe quelle balise.

  ```css
  :root {
    --ts-heading-1: #f8f9fa;  /* titre principal, le plus clair */
    --ts-heading-6: #868e96;  /* le plus atténué                */
    --ts-heading-rule: #495057; /* trait optionnel sous un titre */
  }
  ```

Différenciation plus marquée en opt-in (majuscules + trait `border-bottom`), voir `examples/headings/` (`-enhanced`) :

  ```css
  h1, h2 {
    text-transform: uppercase;
    border-bottom: 1px solid;
    border-color: var(--ts-heading-rule) !important;
  }
  ```

### Pseudo-éléments `::before` / `::after` et `content`

Le moteur résout les pseudo-éléments `::before` et `::after` (ainsi que la syntaxe héritée `:before`/`:after`) et injecte leur `content` comme texte inline, avant/après le contenu de l'élément.

- La propriété `content` accepte les chaînes (`"— "`) et décode les échappements CSS unicode (`"\2014\00a0"`, `"\00bb"`). `content: none` / `normal` ou les formes fonctionnelles non supportées (`attr()`, `counters()`) ne génèrent rien.
- Le pseudo-élément hérite des propriétés typographiques de l'élément (graisse, italique, couleur) ; ses propres déclarations (ex. `color`) priment.
- Alimente nativement les composants Bootstrap qui en dépendent, p. ex. `.blockquote-footer::before { content: "— "; }`.

  ```css
  .note::before { content: "» "; color: #6c757d; }
  ```

## 8bis. Bordures, Dimensionnement, Flexbox, Listes & Tableaux

Ces fonctionnalités s'appuient directement sur lipgloss (styles de bordure, `Place`, `JoinHorizontal`, sous-paquets `list`/`tree`/`table`).

### Styles de Bordure (`border-style`)
- `solid` (défaut), `double`, `thick`, `block`, `outer-half-block`, `inner-half-block`, `markdown`, `ascii`, `dashed`, `dotted`, `hidden`.
- `border-width: medium|thick` ou `3px`+ promeut une bordure `solid` en `thick`.
- Les traits `dashed`/`dotted` sont construits à partir de glyphes box-drawing.
- `rounded` / `.rounded` reste géré (coins arrondis).

```html
<div style="border-style: double;">double</div>
<div style="border-style: block;">block</div>
<div style="border-style: dashed; border-color: #0d6efd;">dashed colorée</div>
```

### Couleurs et Dégradés de Bordure (`border-image` / `border-color-blend`)
- `border-top-color`, `border-right-color`, `border-bottom-color`, `border-left-color` surchargent la couleur partagée `border-color` côté par côté.
- `border-image: #col1 #col2 ...` ou `border-color-blend: #col1 #col2 ...` : applique un **dégradé fluide** le long de la bordure (`lipgloss.BorderForegroundBlend`).

### Dégradés d'Arrière-Plan & Fonctions de Couleur
- **Dégradés CIELAB** : `background: linear-gradient(#start, #mid, #end)` via `lipgloss.Blend1D`.
- **Fonctions de manipulation** :
  - `lighten(#color, percentage%)` : Éclaircit une couleur via `lipgloss.Lighten`.
  - `darken(#color, percentage%)` : Assombrit une couleur via `lipgloss.Darken`.
  - `complementary(#color)` : Calcule la couleur complémentaire via `lipgloss.Complementary`.
### Dimensionnement (`max-width` / `max-height`)
- `max-width` borne la largeur d'un bloc ; le contenu est re-flowé puis le bloc tronqué (lipgloss `MaxWidth`).
- `max-height` borne la hauteur (lipgloss `MaxHeight`).
- Les unités cellulaires (`ch`) mappent 1:1 sur les colonnes ; `px`/`rem` suivent la conversion habituelle (1 cellule ≈ 8px).

### Flexbox (`justify-content` / `align-items`)
- `justify-content`: `flex-start` (défaut), `center`, `flex-end`, et placement via `Place` lorsque la rangée est plus étroite que son conteneur.
- `align-items`: `stretch` (défaut, égalise les hauteurs), `center`, `flex-end`, `flex-start` — positionne les colonnes de hauteurs inégales sur l'axe transverse.
- Classes Bootstrap reconnues : `.justify-content-*`, `.align-items-*`.

### Listes Énumérées (`<ul>` / `<ol>`)
- Rendues via `lipgloss/list` avec indentation des sous-listes imbriquées.
- `list-style-type` : `disc`/`circle`/`square` (puces), `decimal` (défaut `<ol>`), `lower-alpha`/`upper-alpha`, `lower-roman`/`upper-roman`.
- `list-style: none` supprime les marqueurs.

### Arbres (`<ul class="tree">` / `role="tree"`)
- Rendus via `lipgloss/tree` avec connecteurs `├──` / `└──`, sans nœud racine vide.

### Tableaux (`<table>`)
- Rendus via le moteur `lipgloss/table` : cadre externe, séparateur d'en-tête, sans filets de colonnes internes (look Bootstrap épuré).
- Couleurs de ligne préservées (`background-color` sur `<tr>`, classes `.table-primary`, …).
- `border-color` et `border-style` du `<table>` pilotent le cadre.

Voir `examples/lipgloss-features/` pour une démonstration de l'ensemble.
## 8ter. Animations & Indicateurs (Shimmer / Spinner)

`termstrap` intègre nativement des indicateurs de chargement animés ("Working...") inspirés de **Oh My Pi (OMP)**.

### Syntaxe HTML/CSS Déclarative

```html
<div class="border rounded p-1" style="border-color: #bd93f9;">
  <spinner class="text-warning"></spinner>
  <shimmer class="text-primary">Working on analysis...</shimmer>
</div>
```

- **Balises dédiées** : `<spinner>` et `<shimmer>`.
- **Classes CSS équivalentes** : `.spinner` et `.shimmer` sur n'importe quel élément.
- **Propriété CSS** : `animation: shimmer` ou `animation: spinner`.

### Pilotage des Frames en Go

```go
// Dans votre boucle de rafraîchissement / tick TUI :
for step := range 50 {
    out, _ := termstrap.New(
        `<div><spinner class="text-info"></spinner> <shimmer>Working...</shimmer></div>`,
        termstrap.WithWidth(60),
        termstrap.WithStep(step), // 'step' fait avancer le shimmer et le spinner
    ).Render()

    fmt.Printf("\r\x1b[2K%s", out)
    time.Sleep(70 * time.Millisecond)
}
```

- **Fonctions Go autonomes** :
  - `termstrap.Shimmer(text string, step int, opts ...ShimmerOptions) string`
  - `termstrap.Spinner(step int, frames ...[]string) string`



## 8quater. Thème Glamour & Stratégies d'Affichage (Liens & Images)

TermStrap intègre le thème **Glamour** (`ThemeGlamour`) reprenant la charte graphique de Charmbracelet (Glow / Glamour), tout en découplant la présentation visuelle des stratégies d'affichage des liens et des images.

### 1. Charte Visuelle Glamour (`ThemeGlamour`)
- **`<h1>`** : bandeau violet royal / indigo (`#5f5fff`), texte jaune vif (`#ffff87`) en gras, ajusté à la taille du texte (`width: fit-content`), sans marqueur `#`.
- **`<h2>` à `<h5>`** : cyan électrique (`#00afff`) avec marqueurs `## `…`##### `.
- **`<h6>`** : vert émeraude (`#00af5f`) avec marqueur `###### ` et graisse standard.
- **`<blockquote>`** : barre latérale `│` avec retrait de 1 espace, texte en italique.
- **`<code>` inline** : texte corail (`#ff5f5f`) sur fond sombre (`#303030`).
- **`<pre>`** : fond `#2a2a2a` et coloration Chroma Glamour complète.
- **`<table>`** : bordures `#585858`, en-tête en gras cyan `#00afff`.

### 2. Stratégies de Liens (`LinkMode`)
- **`LinkModeOSC8`** *(défaut)* : hyperliens terminaux cliquables natifs sans URL apparente.
- **`LinkModeInline`** : affiche le texte du lien suivi de l'URL brute (`Texte https://...`).
- **`LinkModeFootnotes`** : insère des numéros de référence (`Texte [1]`) et liste les URLs en fin de document.

```go
termstrap.New(html, termstrap.WithLinkMode(termstrap.LinkModeFootnotes))
```

### 3. Stratégies d'Images (`ImageMode`)
- **`ImageModeAuto`** *(défaut)* : rendu graphique matériel complet (Kitty, iTerm2, Sixel, Half-blocks). Comportement natif non contraint.
- **`ImageModeSmart`** *(adaptatif)* : rendu graphique quand c'est optimal, avec bascule automatique sur le placeholder Glamour (`Image: alt → src`) si l'affichage graphique est dégradé :
  - Largeur de colonne trop étroite (`< 12` colonnes par défaut, configurable via `WithMinImageWidth(cols)`).
  - Image native microscopique (`< 32×32 px`, ex. favicons ou pixels de tracking).
  - Rapport d'aspect extrême (`ratio > 5:1` ou `< 1:4`).
  - Format vectoriel non rastérisé (`.svg`).
  - Erreur de chargement réseau ou fichier introuvable (404).
- **`ImageModeGlamour`** : remplace toutes les images par des placeholders Glamour textuels.
- **`ImageModePlaceholder`** : placeholders Markdown simples `[alt](src)`.

```go
// Rendu adaptatif intelligent (opt-in)
termstrap.New(html,
    termstrap.WithTheme(termstrap.ThemeGlamour),
    termstrap.WithImageMode(termstrap.ImageModeSmart),
    termstrap.WithMinImageWidth(15),
)
```

### 4. Preset Tout-en-un
```go
// Active le thème Glamour + liens inline + placeholders images
termstrap.New(html, termstrap.WithGlamourPreset())
```
## 8quinquies. Bannières ASCII Art (`<ascii>`)

La balise `<ascii>` permet de générer des bannières texte en ASCII art (polices FIGlet) intégrées au flux HTML/CSS :

```html
<ascii width="50" color="#ffffff">Hello</ascii>
```

### Attributs & Personnalisation
- **`width`** : Largeur allouée en colonnes (`"50"`) ou en pourcentage (`"100%"`). Si aucune largeur n'est spécifiée, le conteneur s'ajuste automatiquement (*shrink-wrap*) à la largeur exacte de la bannière.
- **`color`** : Couleur du texte ASCII (code hexadécimal `#ffffff`, couleur CSS ou nom). Peut également être défini via CSS (`color: ...` ou classes Bootstrap comme `.text-warning`).
- **`font`** : Nom de la police FIGlet, ou **échelle de repli responsive** (*fallback stack*) séparée par des virgules, ordonnée de la plus grande à la plus petite. C'est la méthode **recommandée** pour un responsive à **lettres nettes** :
  - `font="doh,univers,colossal,big,small"` ou `font="auto"`. Termstrap choisit la **plus grande police qui rentre entièrement** dans la largeur disponible, et descend d'un cran quand l'écran rétrécit. Les glyphes restent **toujours nets** (aucune pixelisation). Pour préserver le style, listez des polices du **même registre** (ici toutes *bloc/plein* comme `doh`).
  - Exemples de polices intégrées :
    - **Bloc plein / Géantes** : `doh`, `univers`, `colossal`, `block`, `banner3-D`, `epic`
    - **Classiques** : `standard`, `big`, `small`, `mini`
    - **Modernes / Italiques** : `slant`, `shadow`, `slant-relief`, `speed`
    - **Jeux / Stylisées** : `doom`, `graffiti`, `larry3d`, `isometric1`, `starwars`
- **`fit`** : Stratégie appliquée quand la bannière dépasse la largeur disponible :
  - **`fit="clip"`** *(défaut)* : troncature propre à droite.
  - **`fit="scale"`** : **downscale pixel à police unique**. La police choisie est conservée *à l'identique*, puis son dessin est ré-échantillonné comme une image 1-bit en blocs Unicode 2×3 (*sextants*). ⚠️ *Compromis* : cela garde exactement la même police mais **pixelise** les lettres aux petites tailles. Pour des lettres nettes, préférez l'échelle de polices (`font="a,b,c"`) ci-dessus.
- **`class="border-shimmer"`** : Applique une vague d'animation shimmer lumineuse animée directement le long de la bordure du bloc.
- Compatible avec les marges, paddings, bordures (`.border`, `.rounded`), ombres et alignements (`.text-center`, `.text-end`).
- **Fonction Go** : `termstrap.AvailableAsciiFonts()` retourne la liste complète des noms de polices intégrées.

## 8quater. En-têtes Épinglés (`position-sticky` / `.sticky`)


Un terminal n'a pas de `position: sticky` CSS : le rendu est **statique**, et le
défilement appartient au conteneur appelant (ex. un `viewport` Bubble Tea). Le
comportement sticky est donc fourni en deux temps, **entièrement calculés par
termstrap** ; l'appelant ne fournit que la position de scroll courante.

```go
// 1. Une fois (ou au resize) : rendu du corps + repérage des sections sticky.
body, regions, err := termstrap.Model{HTML: html, Width: w}.RenderStickyRegions()
viewport.SetContent(body)

// 2. À chaque frame : épingle la section active selon le scroll.
view := termstrap.PinSticky(viewport.View(), regions, viewport.YOffset())
```

- **Vrai `sticky`** : chaque en-tête défile normalement et ne se colle en haut
  qu'une fois sa position naturelle passée au-dessus du viewport ; il est
  ensuite **poussé** par l'en-tête de la section suivante.
- **Détection** : classes `.sticky`, `.sticky-top`, `.position-sticky`,
  `.fixed-top`… (Bootstrap) ou `style="position: sticky|fixed"`.
- **Sections** : chaque enfant de premier niveau du `body` est une section ;
  son en-tête épinglé est l'élément sticky rendu dans le wrapper de la section
  (padding/largeur conservés).

## 9. Rendu d'Images (`<img>`)

### Formats d'Images
- **PNG** (avec transparence alpha)
- **JPEG**
- **WebP**
- **GIF**

### Sources
- **Distantes** : URLs HTTP/HTTPS (`<img src="https://example.com/image.png" />`).
- **Locales** : Fichiers locaux résolus via `RootPath` (`<img src="poster.png" />`).

### Protocoles Terminaux
1. **HalfBlock (Unicode `▀`/`▄`)** : Protocole universel compatible avec tous les terminaux modernes. Rendu direct dans la grille de texte.
2. **Protocoles Graphiques Natifs (Kitty, iTerm2, Sixel)** : Système d'overlays retournant les coordonnées `(Row, Col, Width, Height)` pour intégration dans les frameworks TUI comme Bubble Tea.

---

## 10. Styles Personnalisés & Feuilles de Style Externes

### Styles Inline (`style="..."`)
Vous pouvez définir des propriétés CSS directement sur les éléments HTML :
```html
<div style="background-color: #ff5500; color: #ffffff; margin: 2; padding: 1;">
  Contenu stylé manuellement
</div>
```

### Feuilles de Style Externes (`WithStylesheets`)
Injectez vos propres règles CSS pour surcharger ou enrichir les classes Bootstrap :
```go
customCSS := `
  .badge-custom {
    background-color: #6f42c1;
    color: #ffffff;
    padding: 0 1;
    border-radius: true;
  }
`

m := termstrap.New(htmlContent, 
    termstrap.WithWidth(100),
    termstrap.WithStylesheets(customCSS),
)
output, err := m.Render()
```

`WithStylesheets` **auto-détecte** la nature de chaque argument :

| Argument | Détection | Chargement |
| :--- | :--- | :--- |
| CSS brut | contient `{` / `}` ou un saut de ligne | utilisé tel quel |
| URL | préfixe `http://` ou `https://` | `http.Get` (timeout 5 s, 2xx requis) |
| Chemin de fichier | sinon | `os.ReadFile` (relatif → résolu via `WithRootPath`) |

```go
m := termstrap.New(htmlContent,
    termstrap.WithWidth(100),
    termstrap.WithRootPath("./assets"),
    termstrap.WithStylesheets(
        ".badge-custom { background-color: #6f42c1; }", // CSS brut
        "theme.css",                                    // fichier (via RootPath)
        "https://cdn.example.com/site.css",             // URL
    ),
)
```

Un chemin/URL illisible est ignoré silencieusement (même politique que `<link>`).
La résolution a lieu au `Render()`, donc l'ordre vis-à-vis de `WithRootPath` n'importe pas.

---

## 11. Guide d'Utilisation de l'API Go

```go
package main

import (
    "fmt"
    "github.com/go-scripts/termstrap"
)

func main() {
    html := `
    <div class="row">
      <div class="col-md-6 bg-dark text-white p-2 border rounded shadow-lg">
        <h2 class="text-center">Serveur Actif</h2>
        <p>Le serveur fonctionne normalement.</p>
      </div>
      <div class="col-md-6 p-2 text-center">
        <img src="https://go.dev/doc/gopher/frontpage.png" alt="Gopher" />
      </div>
    </div>`

    // Initialisation du modèle
    m := termstrap.New(html, 
        termstrap.WithWidth(90),
        termstrap.WithRootPath("./assets"),
    )

    // Rendu en chaîne ANSI
    output, err := m.Render()
    if err != nil {
        panic(err)
    }

    fmt.Print(output)
}
```
