<?php

declare(strict_types=1);

require_once dirname(__DIR__, 2) . '/vendor/autoload.php';

use Termstrap\Termstrap;

$width = 80;
if (stream_isatty(STDOUT)) {
  $term = new \Symfony\Component\Console\Terminal();
  $w = $term->getWidth();
  if ($w > 0) {
    $width = $w;
  }
}

$htmlCard = <<<'HTML'
<div class="container p-2 border rounded">
  <div class="row">
    <div class="col-md-6 p-2">
      <div class="p-2 bg-primary rounded">
        <h2>Primary Card Header</h2>
        <p>This card demonstrates the primary background and text styling.</p>
        <p class="text-white">Highlighted status: <span class="fw-bold">Active</span></p>
      </div>
    </div>
    <div class="col-md-6 p-2">
      <div class="p-2 bg-dark rounded">
        <h3>Code & Details</h3>
        <p>Inline command: <code>go run ./main.go</code></p>
        <blockquote>Tokyo Night, Dracula, Bootstrap, and Glamour themes adapt automatically.</blockquote>
        <div class="p-1 bg-secondary rounded mt-1">
          <span class="text-warning">Warning text</span> | <span class="text-success">Success text</span> | <span class="text-info">Info text</span>
        </div>
      </div>
    </div>
  </div>
</div>
HTML;

$themes = [
  ['name' => 'Default / Bootstrap Theme', 'theme' => 'bootstrap'],
  ['name' => 'Tokyo Night Theme', 'theme' => 'tokyonight'],
  ['name' => 'Dracula Theme', 'theme' => 'dracula'],
  ['name' => 'Glamour Theme', 'theme' => 'glamour'],
];

foreach ($themes as $t) {
  echo sprintf("\n=== %s (Theme: \"%s\") ===\n\n", $t['name'], $t['theme']);
  echo Termstrap::toAnsi($htmlCard, ['width' => $width, 'theme' => $t['theme']]) . "\n";
}
