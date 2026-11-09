<?php

declare(strict_types=1);

require_once dirname(__DIR__, 2) . '/vendor/autoload.php';

use Termstrap\Termstrap;

$width = 100;
$dir = __DIR__;
$htmlPath = dirname($dir) . '/html/80s Huge Hits FLAC 2026.html';

if (file_exists($htmlPath)) {
    $html = (string) file_get_contents($htmlPath);
} else {
    $html = (string) file_get_contents($dir . '/torrent_detail.html');
}

$opts = [
    'width' => $width,
    'theme' => 'bootstrap',
    'disable_images' => true,
];

echo Termstrap::toAnsi($html, $opts) . "\n";
