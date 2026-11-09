<?php

declare(strict_types=1);

require_once dirname(__DIR__, 3) . '/vendor/autoload.php';

use Termstrap\Termstrap;

$width = 80;
$content = <<<'HTML'
<h1>Color Modes Demonstration</h1>
<p>Termstrap supports TrueColor, ANSI 256, and ANSI 16 color depth output.</p>
<div class="row">
  <div class="col-4 border p-1">TrueColor (24-bit)</div>
  <div class="col-4 border p-1">ANSI 256</div>
  <div class="col-4 border p-1">ANSI 16</div>
</div>
HTML;

echo Termstrap::toAnsi($content, ['width' => $width, 'theme' => 'none']) . "\n";
