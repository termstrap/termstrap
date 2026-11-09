<?php

declare(strict_types=1);

require_once dirname(__DIR__, 3) . '/vendor/autoload.php';

use Termstrap\Termstrap;

$width = 80;
$content = <<<'HTML'
<h1>Markdown Image Integration</h1>
<p>Markdown documents can contain embedded images using <code>![alt](src)</code> syntax.</p>
<div><img src="logo.png" alt="Logo" /></div>
HTML;

echo Termstrap::toAnsi($content, ['width' => $width, 'theme' => 'none', 'image_mode' => 'placeholder']) . "\n";
