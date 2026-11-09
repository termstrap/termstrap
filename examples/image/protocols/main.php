<?php

declare(strict_types=1);

require_once dirname(__DIR__, 3) . '/vendor/autoload.php';

use Termstrap\Termstrap;

$width = 80;
$content = <<<'HTML'
<h1>Graphics Protocols</h1>
<p>Termstrap supports Kitty, iTerm2, Sixel, Sextant, and Unicode Half-Block protocols.</p>
<div class="row">
  <div class="col-4 border p-1">Kitty Graphics</div>
  <div class="col-4 border p-1">iTerm2 Inline</div>
  <div class="col-4 border p-1">Sixel Graphics</div>
</div>
HTML;

echo Termstrap::toAnsi($content, ['width' => $width, 'theme' => 'none']) . "\n";
