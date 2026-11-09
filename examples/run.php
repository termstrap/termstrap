<?php

declare(strict_types=1);

require_once dirname(__DIR__) . '/vendor/autoload.php';

use Symfony\Component\Console\Terminal;

$width = 80;
if (stream_isatty(STDOUT)) {
    $term = new Terminal();
    $w = $term->getWidth();
    if ($w > 0) {
        $width = $w;
    }
}

$examplesDir = __DIR__;

$allExamples = [
    'ascii',
    'ascii-all-fonts',
    'bootstrap',
    'borders',
    'breakpoints',
    'custom-css',
    'fiche',
    'glamour',
    'headings',
    'html',
    'image/animated_gif',
    'image/caching',
    'image/colormodes',
    'image/detect',
    'image/formats',
    'image/grid',
    'image/local',
    'image/markdown',
    'image/protocols',
    'image/three-columns',
    'lipgloss-features',
    'media-queries',
    'nested',
    'nested-no-md',
    'shadows',
    'spacing',
    'spinner-shimmer',
    'styling',
    'syntax',
    'syntax-custom',
    'syntax-themes',
    'themes',
    'torrent',
];

$target = $argv[1] ?? '';
if ($target === '--list') {
    echo "Available examples:\n";
    foreach ($allExamples as $name) {
        echo "  - {$name}\n";
    }
    exit(0);
}

$runExample = function (string $name) use ($examplesDir): void {
    $dir = $examplesDir . '/' . $name;
    if (!is_dir($dir)) {
        fwrite(STDERR, "Example not found: {$name}\n");
        return;
    }

    $mainPhp = $dir . '/main.php';
    if (file_exists($mainPhp)) {
        require $mainPhp;
        return;
    }

    fwrite(STDERR, "No main.php found for example: {$name}\n");
};

if ($target !== '') {
    $runExample($target);
} else {
    foreach ($allExamples as $name) {
        echo "\n\033[1;36m════════════════════════════════════════════════════════════════════════════════\033[0m\n";
        echo "\033[1;32m ═══ examples/{$name} ═══\033[0m\n";
        echo "\033[1;36m════════════════════════════════════════════════════════════════════════════════\033[0m\n\n";
        $runExample($name);
    }
}
