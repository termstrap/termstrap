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

$content = <<<'HTML'
<div class="container-fluid p-1">
  <div class="w-100 text-center mb-1">
    <ascii class="shimmer border border-shimmer" color="#7aa2f7">Hubs.sh</ascii>
    <div class="text-muted mt-1">Bordure ajustée au plus près + effet border-shimmer synchronisé</div>
  </div>

  <div class="row">
    <div class="col-12 col-md-6 border rounded p-1 mb-1">
      <div class="text-warning fw-bold mb-1">Police "slant" (border rounded) :</div>
      <ascii font="slant" color="#ff9e64" class="border rounded p-1">Hello</ascii>
    </div>

    <div class="col-12 col-md-6 border rounded p-1 mb-1">
      <div class="text-info fw-bold mb-1">Police "doom" (border-shimmer) :</div>
      <ascii font="doom" color="#7dcfff" class="border border-shimmer shimmer">World</ascii>
    </div>
  </div>

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

  <div class="border rounded p-1 bg-dark text-center" style="border-color: #9ece6a;">
    <ascii width="100%" color="#9ece6a" class="shimmer">Termstrap</ascii>
  </div>
</div>
HTML;

echo Termstrap::toAnsi($content, ['width' => $width, 'theme' => 'tokyonight']) . "\n";
