<?php

declare(strict_types=1);

require_once dirname(__DIR__, 3) . '/vendor/autoload.php';

use Termstrap\Termstrap;

$width = 80;
$content = <<<'HTML'
<h1>Animated GIF Support</h1>
<p>Animated GIFs are automatically adapted with smart placeholders or frame stepping.</p>
<div><img src="test.gif" alt="Animation demo" /></div>
HTML;

echo Termstrap::toAnsi($content, [
    'width' => $width,
    'theme' => 'none',
    'image_mode' => 'glamour',
]) . "\n";
