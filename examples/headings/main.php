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

$content = <<<'HTML'
<h1>Heading level 1</h1>
<p>Body text under the top-level title.</p>
<h2>Heading level 2</h2>
<p>Body text.</p>
<h3>Heading level 3</h3>
<p>Body text.</p>
<h4>Heading level 4</h4>
<p>Body text.</p>
<h5>Heading level 5</h5>
<p>Body text.</p>
<h6>Heading level 6</h6>
<p>Body text.</p>
HTML;

echo Termstrap::toAnsi($content, ['width' => $width, 'theme' => 'bootstrap']);
