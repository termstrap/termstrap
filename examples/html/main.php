<?php

declare(strict_types=1);

require_once dirname(__DIR__, 2) . '/vendor/autoload.php';

use Termstrap\Termstrap;

$width = 100;
$htmlPath = __DIR__ . '/80s Huge Hits FLAC 2026.html';
$html = (string) file_get_contents($htmlPath);

echo Termstrap::toAnsi($html, ['width' => $width, 'theme' => 'bootstrap', 'disable_images' => true]) . "\n";
