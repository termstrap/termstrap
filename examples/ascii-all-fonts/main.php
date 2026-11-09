<?php

declare(strict_types=1);

require_once dirname(__DIR__, 2) . '/vendor/autoload.php';

use Termstrap\Termstrap;

$width = 90;
$word = $argv[1] ?? 'Hubs.sh';

$fonts = Termstrap::availableAsciiFonts();
sort($fonts);

echo "\033[1;36m================================================================================\033[0m\n";
echo sprintf("\033[1;32m  Termstrap ASCII Fonts Showcase (%d polices disponibles)\033[0m\n", count($fonts));
echo sprintf("  Texte affiché : %s | Largeur : %d colonnes\n", $word, $width);
echo "\033[1;36m================================================================================\033[0m\n\n";

// Show top 8 representative fonts when run in preview mode
$selectedFonts = array_slice($fonts, 0, 8);
foreach ($selectedFonts as $i => $f) {
  $html = sprintf(
    '<div class="border rounded p-1 mb-1">
  <div class="text-warning fw-bold mb-1">#%d — Police: <span class="text-info">%s</span></div>
  <div class="text-center">
    <ascii font="%s" color="#7aa2f7" class="border p-1">%s</ascii>
  </div>
</div>',
    $i + 1,
    $f,
    $f,
    $word
  );

  echo Termstrap::toAnsi($html, ['width' => $width, 'theme' => 'tokyonight']) . "\n";
}

echo sprintf("\n\033[1;32m✔ %d polices présentées avec succès !\033[0m\n", count($selectedFonts));
