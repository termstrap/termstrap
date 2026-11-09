<?php

declare(strict_types=1);

require_once dirname(__DIR__, 3) . '/vendor/autoload.php';

use Termstrap\Termstrap;

$width = 80;
$content = <<<'HTML'
<h1>Supported Image Formats</h1>
<p>Termstrap supports PNG, JPEG, WebP, and GIF formats.</p>
<div class="row">
  <div class="col-3 border p-1 text-center">[PNG]</div>
  <div class="col-3 border p-1 text-center">[JPEG]</div>
  <div class="col-3 border p-1 text-center">[WebP]</div>
  <div class="col-3 border p-1 text-center">[GIF]</div>
</div>
HTML;

echo Termstrap::toAnsi($content, ['width' => $width, 'theme' => 'none']) . "\n";
