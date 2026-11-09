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

$dir = __DIR__;
$html = (string) file_get_contents($dir . '/index.html');
$css = (string) file_get_contents($dir . '/style.css');

echo Termstrap::toAnsi($html, [
    'width' => $width,
    'theme' => 'none',
    'stylesheets' => [$css],
    'root_path' => $dir,
    'disable_images' => true,
]) . "\n";
