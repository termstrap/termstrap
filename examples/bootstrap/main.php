<?php

declare(strict_types=1);

require_once dirname(__DIR__, 2) . '/vendor/autoload.php';

use Termstrap\Termstrap;

$width = 100;
$dir = __DIR__;
$bsCss = (string) file_get_contents(dirname(__DIR__, 2) . '/assets/css/bootstrap.css');

$demos = [
    ['title' => '1. System Monitoring Dashboard (KPIs & Grid)', 'file' => 'dashboard.html'],
    ['title' => '2. Bootstrap Striped & Colored Tables', 'file' => 'tables.html'],
    ['title' => '3. Contextual Alerts & Callouts', 'file' => 'alerts.html'],
    ['title' => '4. Media Fiche & Responsive Grid', 'file' => 'media_card.html'],
];

foreach ($demos as $d) {
    echo "================================================================================\n";
    echo " {$d['title']}\n";
    echo "================================================================================\n\n";

    $html = (string) file_get_contents($dir . '/' . $d['file']);
    $opts = [
        'width' => $width,
        'theme' => 'bootstrap',
        'stylesheets' => [$bsCss],
        'disable_images' => true,
    ];

    echo Termstrap::toAnsi($html, $opts) . "\n\n";
}
