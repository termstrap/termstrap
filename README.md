# Termstrap (PHP)

Moteur de rendu terminal natif en PHP capable de transformer du contenu **HTML**, **CSS** et **Markdown** en sortie texte enrichie avec codes d'échappement ANSI.

Portage PHP natif autonome basé sur le moteur de rendu terminal Termstrap (avec parité différentielle avec l'implémentation Go d'origine).

---

## Fonctionnalités

- **HTML & CSS dans le terminal** : Support des balises courantes (`<h1>`-`<h6>`, `<p>`, `<a>`, `<table>`, `<ul>`/`<ol>`, `<blockquote>`, `<code>`, `<pre>`, `<hr>`).
- **Support CSS étendu** :
  - Couleurs (hexadécimal, RGB, HSL, noms CSS, truecolor / 24-bit, 256 couleurs, 16 couleurs ANSI).
  - Propriétés de boîte : `margin`, `padding`, `border` (styles `solid`, `dashed`, `double`, `rounded`, `heavy`, etc.).
  - Ombres (`box-shadow`), alignements de texte, transformations de texte.
  - Media queries et breakpoints selon la largeur du terminal (`@media (min-width: ...)`).
- **Markdown vers ANSI** : Conversion transparente via League CommonMark avant rendu.
- **Coloration syntaxique** : Intégration de Tempest Highlight pour les blocs de code.
- **Rendu d'images dans le terminal** :
  - Protocoles modernes : Kitty Graphics, iTerm2, Sixel.
  - Replis en caractères : Half-block (haut/bas) et sextants.
  - Détection automatique des capacités du terminal.
- **Titres & bannières ASCII FIGlet** : Rendu des polices ASCII FIGlet (`.flf`).
- **Composants TUI dynamiques** : Prise en charge des spinners et shimmers animés.
- **Thèmes prêts à l'emploi** : `bootstrap`, `tokyonight`, `dracula`, `glamour` ou styles personnalisés.
- **CLI & Bibliothèque** : Utilisable directement en ligne de commande ou intégré comme dépendance PHP.

---

## Prérequis

- **PHP** : >= 8.4.1
- **Extensions requises** :
  - `ext-dom` (avec la nouvelle API HTML5 DOM de PHP 8.4)
  - `ext-mbstring`
  - `ext-gd`

---

## Installation

```bash
composer require codes-sh/termstrap
```

Ou en clonant le dépôt pour contribuer :

```bash
git clone https://github.com/codes-sh/termstrap.git
cd termstrap
composer install
```

---

## Utilisation en CLI

Le binaire CLI `bin/termstrap` permet de rendre des fichiers ou du flux standard (`stdin`).

### Rendre un fichier HTML ou Markdown

```bash
# Rendu d'un fichier HTML
./bin/termstrap document.html

# Rendu d'un fichier Markdown
./bin/termstrap README.md

# Rendu depuis l'entrée standard
echo "<h1>Bonjour <em>monde</em> !</h1>" | ./bin/termstrap -
```

### Options principales

| Option | Raccourci | Description |
|---|---|---|
| `--width` | `-w` | Largeur du terminal en colonnes (0 = détection automatique). |
| `--theme` | `-t` | Thème visuel (`bootstrap`, `tokyonight`, `dracula`, `glamour`, `none`). Défaut: `bootstrap`. |
| `--markdown` | `-m` | Force le mode Markdown. |
| `--css` | | Chemin vers une feuille de style CSS personnalisée. |
| `--color` | `-c` | Mode de couleurs (`truecolor`, `256`, `16`). |
| `--protocol` | `-p` | Force le protocole image (`halfblock`, `sextant`, `kitty`, `iterm`, `sixel`). |
| `--no-images` | | Désactive l'affichage des images (remplacées par du texte alt). |
| `--link-mode` | | Mode d'affichage des liens hypertexte (`osc8`, `inline`, `footnotes`). |
| `--watch` | `-W` | Surveille le fichier et recharge automatiquement l'affichage à chaque modification. |
| `--preset` | | Préréglage de présentation (ex: `glamour`). |
| `--root` | `-r` | Répertoire racine pour la résolution des chemins d'images relatifs. |

---

## Utilisation en Bibliothèque PHP

### Exemple rapide

```php
use Termstrap\Termstrap;

require __DIR__ . '/vendor/autoload.php';

$html = <<<HTML
<div style="background-color: #1e1e2e; color: #cdd6f4; padding: 1; border: 1px solid #cba6f7;">
    <h1 style="color: #89b4fa;">Termstrap PHP</h1>
    <p>Rendu natif <strong>HTML</strong> et <strong>CSS</strong> dans votre terminal.</p>
</div>
HTML;

echo Termstrap::toAnsi($html, [
    'width' => 80,
    'theme' => 'dracula',
]);
```

### Avec Markdown

````php
use Termstrap\Termstrap;

$markdown = <<<'MD'
# Mon titre Markdown

Voici du code PHP :

```php
echo "Hello Terminal!";
```
MD;

echo Termstrap::toAnsi($markdown, [
    'markdown' => true,
    'theme' => 'tokyonight',
]);
````

### Régions collantes (*Sticky*) pour TUI

```php
use Termstrap\Termstrap;

[$body, $stickyRegions] = Termstrap::renderStickyRegions($htmlContent, [
    'width' => 100,
]);
```

---

## Exemples inclus

Le projet comprend de nombreux exemples dans le dossier `examples/` :

```bash
# Lancer tous les exemples locaux
make examples

# Lancer un exemple spécifique
make example NAME=borders
make example NAME=themes
make example NAME=glamour
make example NAME=syntax
make example NAME=spinner-shimmer
make example NAME=breakpoints
```

---

## Tests & Qualité

Exécuter la suite de tests complète avec PHPUnit :

```bash
# Lancer les tests unitaires et fonctionnels
make test

# Lancer les tests différentiels de parité Go / PHP
make test-parity

# Lancer les tests end-to-end
make test-e2e

# Validation composer + tests
make check
```

---

## Licence

Ce projet est sous licence [MIT](LICENSE).
