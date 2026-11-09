<?php

declare(strict_types=1);

require_once dirname(__DIR__, 3) . '/vendor/autoload.php';

use Termstrap\Termstrap;

$width = 80;
if (stream_isatty(STDOUT)) {
  $term = new \Symfony\Component\Console\Terminal();
  $w = $term->getWidth();
  if ($w > 0) {
    $width = $w;
  }
}

$tmpDir = sys_get_temp_dir() . '/termstrap_img_' . uniqid();
mkdir($tmpDir);

$createPng = function (string $path, int $r, int $g, int $b): void {
  $im = imagecreatetruecolor(100, 60);
  $col = imagecolorallocate($im, $r, $g, $b);
  imagefill($im, 0, 0, $col);
  imagepng($im, $path);
};

$createPng($tmpDir . '/red.png', 220, 50, 50);
$createPng($tmpDir . '/green.png', 50, 180, 50);
$createPng($tmpDir . '/blue.png', 50, 100, 220);

$content = <<<'HTML'
<h1>Local Image Files</h1>

<h2>Gradient (generated PNG)</h2>

<div><img src="red.png" alt="red" /></div>

<hr />

<h2>Color Swatches in Grid</h2>

<div class="row">
  <div class="col-md-4 border rounded p-1 text-center">
    <div><img src="red.png" alt="red" /></div>
    <div><b>Red</b></div>
  </div>
  <div class="col-md-4 border rounded p-1 text-center">
    <div><img src="green.png" alt="green" /></div>
    <div><b>Green</b></div>
  </div>
  <div class="col-md-4 border rounded p-1 text-center">
    <div><img src="blue.png" alt="blue" /></div>
    <div><b>Blue</b></div>
  </div>
</div>

<hr />

<p>These images were generated on the fly and loaded from disk using <code>RootPath</code>.</p>
HTML;

echo Termstrap::toAnsi($content, [
  'width' => $width,
  'theme' => 'bootstrap',
  'root_path' => $tmpDir,
]) . "\n";
