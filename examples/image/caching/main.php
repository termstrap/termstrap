<?php

declare(strict_types=1);

require_once dirname(__DIR__, 3) . '/vendor/autoload.php';

use Termstrap\Termstrap;

$width = 80;
$content = <<<'HTML'
<h1>Image Caching Policy Demo</h1>
<p>Images loaded by Termstrap can be cached with a specified TTL.</p>
<div><img src="avatar.png" alt="Avatar" /></div>
HTML;

echo Termstrap::toAnsi($content, [
    'width' => $width,
    'theme' => 'none',
    'image_mode' => 'placeholder',
]) . "\n";
