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
<h3>Highlighting &amp; chrome</h3>

<p>Inline <code>termstrap.New()</code> plus a block below. See the
<a href="https://github.com/termstrap/termstrap">project link</a> color.</p>

<pre><code class="language-go">package main

import "fmt"

// greet prints a friendly message.
func greet(name string) string {
	return fmt.Sprintf("hello, %s", name)
}
</code></pre>

<blockquote>Blockquote border + text follow the theme accent.</blockquote>

<hr>

<pre><code class="language-diff">--- a/app.go
+++ b/app.go
@@ -1,2 +1,2 @@
-	return err
+	return nil
</code></pre>
HTML;

$themes = ['bootstrap', 'dracula', 'tokyonight'];
foreach ($themes as $theme) {
    echo sprintf("\n=== %s ===\n\n", $theme);
    echo Termstrap::toAnsi($content, ['width' => $width, 'theme' => $theme]) . "\n";
}
