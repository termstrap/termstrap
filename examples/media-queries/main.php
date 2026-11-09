<?php

declare(strict_types=1);

require_once dirname(__DIR__, 2) . '/vendor/autoload.php';

use Termstrap\Termstrap;

$dir = __DIR__;
$html = (string) file_get_contents($dir . '/index.html');
$css = (string) file_get_contents($dir . '/style.css');

$widths = [
    ['name' => 'Narrow viewport (60 cols)', 'width' => 60],
    ['name' => 'Medium viewport (90 cols)', 'width' => 90],
    ['name' => 'Wide viewport (120 cols)', 'width' => 120],
];

foreach ($widths as $w) {
    echo sprintf("\n\033[1;36m=== %s ===\033[0m\n\n", $w['name']);
    echo Termstrap::toAnsi($html, [
        'width' => $w['width'],
        'theme' => 'none',
        'stylesheets' => [$css],
        'root_path' => $dir,
        'disable_images' => true,
    ]) . "\n";
}
