<?php

declare(strict_types=1);

require_once dirname(__DIR__, 2) . '/vendor/autoload.php';

use Termstrap\Termstrap;

$content = <<<'HTML'
<h1>Responsive Breakpoints Demo</h1>

<h2>Breakpoint Thresholds</h2>

<table>
  <tr><th>Breakpoint</th><th>Prefix</th><th>Terminal Width</th></tr>
  <tr><td>xs</td><td>col- / col-xs-</td><td>&lt; 70 cols</td></tr>
  <tr><td>sm</td><td>col-sm-</td><td>&gt;= 70 cols</td></tr>
  <tr><td>md</td><td>col-md-</td><td>&gt;= 100 cols</td></tr>
  <tr><td>lg</td><td>col-lg-</td><td>&gt;= 130 cols</td></tr>
  <tr><td>xl</td><td>col-xl-</td><td>&gt;= 160 cols</td></tr>
  <tr><td>xxl</td><td>col-xxl-</td><td>&gt;= 190 cols</td></tr>
</table>

<hr />

<h2>col-md-6 — Stacks below 100 cols</h2>

<div class="row">
  <div class="col-md-6 border p-1">
    <div><b>Left (col-md-6)</b></div>
    <p>Visible side-by-side at 100+ cols.</p>
  </div>
  <div class="col-md-6 border p-1">
    <div><b>Right (col-md-6)</b></div>
    <p>Stacks vertically below 100 cols.</p>
  </div>
</div>

<hr />

<h2>col-lg-4 — Stacks below 130 cols</h2>

<div class="row">
  <div class="col-lg-4 border p-1">
    <div><b>A (col-lg-4)</b></div>
  </div>
  <div class="col-lg-4 border p-1">
    <div><b>B (col-lg-4)</b></div>
  </div>
  <div class="col-lg-4 border p-1">
    <div><b>C (col-lg-4)</b></div>
  </div>
</div>

<hr />

<h2>col-sm-6 — Side-by-side from 70 cols</h2>

<div class="row">
  <div class="col-sm-6 border p-1">
    <div><b>Left (col-sm-6)</b></div>
    <p>Side-by-side starting at 70 cols.</p>
  </div>
  <div class="col-sm-6 border p-1">
    <div><b>Right (col-sm-6)</b></div>
  </div>
</div>

<hr />

<h2>Mixed breakpoints — col-sm-12 / col-md-6</h2>

<div class="row">
  <div class="col-sm-12 col-md-6 border p-1">
    <div><b>Panel A</b></div>
    <p>Full width on small, half on medium+.</p>
  </div>
  <div class="col-sm-12 col-md-6 border p-1">
    <div><b>Panel B</b></div>
    <p>Full width on small, half on medium+.</p>
  </div>
</div>
HTML;

$widths = [
  ['name' => 'Very Narrow (55 cols - xs)', 'width' => 55],
  ['name' => 'Small (75 cols - sm)', 'width' => 75],
  ['name' => 'Medium / Standard (110 cols - md)', 'width' => 110],
  ['name' => 'Large (140 cols - lg)', 'width' => 140],
  ['name' => 'Extra Large (170 cols - xl)', 'width' => 170],
];

foreach ($widths as $w) {
  echo sprintf("\n\033[1;36m%s\033[0m\n", $w['name']);
  echo "════════════════════════════════════════\n";
  echo Termstrap::toAnsi($content, ['width' => $w['width'], 'theme' => 'bootstrap']);
}
