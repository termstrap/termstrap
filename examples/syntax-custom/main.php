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
<h3>Custom palette (no theme)</h3>

<p>An inline <code>var()</code> override and a
<a href="https://example.com">custom-colored link</a>.</p>

<pre><code class="language-python">def fib(n: int) -> int:
    # classic recursion
    if n < 2:
        return n
    return fib(n - 1) + fib(n - 2)

print([fib(i) for i in range(8)])
</code></pre>

<blockquote>Blockquote styled by --ts-blockquote-color / --ts-blockquote-border.</blockquote>

<hr>
HTML;

$customPalette = <<<'CSS'
:root {
  --ts-code-bg: #0b1021;
  --ts-code-inline-bg: #1b2440;
  --ts-code-inline-fg: #ffd866;

  --ts-code-name: #e6e6fa;
  --ts-code-keyword: #ff6ac1;
  --ts-code-function: #57c7ff;
  --ts-code-type: #5af78e;
  --ts-code-builtin: #c792ea;
  --ts-code-string: #ffd866;
  --ts-code-number: #ff9e64;
  --ts-code-comment: #6b7394;
  --ts-code-operator: #89ddff;

  --ts-link-color: #57c7ff;
  --ts-hr-color: #ff6ac1;
  --ts-blockquote-color: #5af78e;
  --ts-blockquote-border: #ff6ac1;
}
CSS;

echo Termstrap::toAnsi($content, [
  'width' => $width,
  'theme' => 'none',
  'stylesheets' => [$customPalette],
]) . "\n";
