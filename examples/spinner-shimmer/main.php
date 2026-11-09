<?php

declare(strict_types=1);

require_once dirname(__DIR__, 2) . '/vendor/autoload.php';

use Termstrap\Animation\Shimmer;
use Termstrap\Animation\Spinner;
use Termstrap\Termstrap;

$width = 80;

echo "Showing animated 'Working...' spinner indicator frames:\n\n";

for ($step = 0; $step < 5; $step++) {
  $spinner = Spinner::spinner($step);
  $shimmerText = Shimmer::shimmer('Working...', $step, '#6272a4', '#50fa7b', true);

  $html = sprintf(
    '<div class="border rounded p-1 mb-1" style="border-color: #bd93f9; width: 65ch;">
  <div><span class="text-warning">%s</span> %s</div>
  <div class="mt-1"><spinner class="text-info"></spinner> <shimmer class="text-primary">Declarative &lt;shimmer&gt; wave analyzing code...</shimmer></div>
  <div class="mt-1 text-muted"><span class="spinner text-success"></span> <span class="shimmer">Class-based .spinner & .shimmer in action...</span></div>
</div>',
    $spinner,
    $shimmerText
  );

  echo Termstrap::toAnsi($html, ['width' => $width, 'step' => $step]) . "\n";
}
