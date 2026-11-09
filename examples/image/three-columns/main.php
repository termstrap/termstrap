<?php

declare(strict_types=1);

require_once dirname(__DIR__, 3) . '/vendor/autoload.php';

use Termstrap\Termstrap;

$width = 90;
$content = <<<'HTML'
<h1>Three Column Image Grid</h1>
<div class="row">
  <div class="col-4 border rounded p-1 text-center">
    <div><img src="img1.png" alt="Photo 1" /></div>
    <div class="fw-bold">Item 1</div>
  </div>
  <div class="col-4 border rounded p-1 text-center">
    <div><img src="img2.png" alt="Photo 2" /></div>
    <div class="fw-bold">Item 2</div>
  </div>
  <div class="col-4 border rounded p-1 text-center">
    <div><img src="img3.png" alt="Photo 3" /></div>
    <div class="fw-bold">Item 3</div>
  </div>
</div>
HTML;

echo Termstrap::toAnsi($content, [
  'width' => $width,
  'theme' => 'bootstrap',
  'image_mode' => 'placeholder',
]) . "\n";
