<?php

declare(strict_types=1);

require_once dirname(__DIR__, 3) . '/vendor/autoload.php';

use Termstrap\Termstrap;

$width = 80;
$content = <<<'HTML'
<h1>Image Grid</h1>
<div class="row">
  <div class="col-6 border p-1 text-center">Image 1</div>
  <div class="col-6 border p-1 text-center">Image 2</div>
</div>
HTML;

echo Termstrap::toAnsi($content, ['width' => $width, 'theme' => 'none']) . "\n";
