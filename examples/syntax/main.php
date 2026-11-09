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
<h1>Syntax Highlighting</h1>

<p>Explicit language via <code>class="language-go"</code>:</p>
<pre><code class="language-go">package main

import "fmt"

// greet prints a friendly message.
func greet(name string) string {
	return fmt.Sprintf("Hello, %s!", name)
}

func main() {
	fmt.Println(greet("termstrap"))
}
</code></pre>

<p>Another language (Python):</p>
<pre><code class="language-python">def fib(n: int) -> int:
    a, b = 0, 1
    for _ in range(n):
        a, b = b, a + b
    return a

print([fib(i) for i in range(10)])
</code></pre>

<p>No class — language auto-detected from content (shell script):</p>
<pre><code>#!/usr/bin/env bash
set -euo pipefail

for file in *.txt; do
  echo "Processing $file"
  wc -l "$file"
done
</code></pre>

<p>Unified diff (exercises inserted/deleted tokens):</p>
<pre><code class="language-diff">--- a/main.go
+++ b/main.go
@@ -1,3 +1,3 @@
-func old()
+func new()
</code></pre>
HTML;

echo Termstrap::toAnsi($content, ['width' => $width, 'theme' => 'bootstrap']) . "\n";
