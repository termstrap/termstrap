<?php

declare(strict_types=1);

namespace Termstrap\Ascii;

use Symfony\Component\Tui\Widget\Figlet\FigletFont;
use Symfony\Component\Tui\Widget\Figlet\FigletRenderer;
use Termstrap\Util;
use Throwable;

final class Figlet
{
 public const DEFAULT_FONT_HIERARCHY = [
  'doh',
  'univers',
  'block',
  'colossal',
  'big',
  'standard',
  'small',
  'mini',
 ];

 private const INK_THRESHOLD = 0.42;

 /** @var list<string> */
 private const SCALE_SEXTANT_RUNES = [
  " ",
  "\u{1FB00}",
  "\u{1FB01}",
  "\u{1FB02}",
  "\u{1FB03}",
  "\u{1FB04}",
  "\u{1FB05}",
  "\u{1FB06}",
  "\u{1FB07}",
  "\u{1FB08}",
  "\u{1FB09}",
  "\u{1FB0A}",
  "\u{1FB0B}",
  "\u{1FB0C}",
  "\u{1FB0D}",
  "\u{1FB0E}",
  "\u{1FB0F}",
  "\u{1FB10}",
  "\u{1FB11}",
  "\u{1FB12}",
  "\u{1FB13}",
  "\u{258C}",
  "\u{1FB14}",
  "\u{1FB15}",
  "\u{1FB16}",
  "\u{1FB17}",
  "\u{1FB18}",
  "\u{1FB19}",
  "\u{1FB1A}",
  "\u{1FB1B}",
  "\u{1FB1C}",
  "\u{1FB1D}",
  "\u{1FB1E}",
  "\u{1FB1F}",
  "\u{1FB20}",
  "\u{1FB21}",
  "\u{1FB22}",
  "\u{1FB23}",
  "\u{1FB24}",
  "\u{1FB25}",
  "\u{1FB26}",
  "\u{1FB27}",
  "\u{2590}",
  "\u{1FB28}",
  "\u{1FB29}",
  "\u{1FB2A}",
  "\u{1FB2B}",
  "\u{1FB2C}",
  "\u{1FB2D}",
  "\u{1FB2E}",
  "\u{1FB2F}",
  "\u{1FB30}",
  "\u{1FB31}",
  "\u{1FB32}",
  "\u{1FB33}",
  "\u{1FB34}",
  "\u{1FB35}",
  "\u{1FB36}",
  "\u{1FB37}",
  "\u{1FB38}",
  "\u{1FB39}",
  "\u{1FB3A}",
  "\u{1FB3B}",
  "\u{2588}",
 ];

 /** @var array<string, FigletFont> */
 private static array $fontCache = [];

 /** @var array<string, string> */
 private static array $availableFontsMap = [];

 /** @var list<string> */
 private static array $allFontNames = [];

 public static function initFonts(): void
 {
  if (!empty(self::$availableFontsMap)) {
   return;
  }

  $fontsDir = dirname(__DIR__, 2) . '/assets/fonts';
  $files = glob($fontsDir . '/*.flf') ?: [];
  foreach ($files as $file) {
   $base = basename($file, '.flf');
   self::$availableFontsMap[strtolower($base)] = $base;
   self::$allFontNames[] = $base;
  }
  sort(self::$allFontNames);
 }

 /**
  * @return list<string>
  */
 public static function availableAsciiFonts(): array
 {
  self::initFonts();
  return self::$allFontNames;
 }

 public static function resolveFontName(string $name): string
 {
  self::initFonts();
  $name = strtolower(trim($name));
  if ($name === '') {
   return 'standard';
  }
  return self::$availableFontsMap[$name] ?? 'standard';
 }

 /**
  * @return list<string>
  */
 public static function resolveFontCandidates(string $spec): array
 {
  $spec = trim($spec);
  if ($spec === '') {
   return ['standard'];
  }
  if (strcasecmp($spec, 'auto') === 0) {
   return self::DEFAULT_FONT_HIERARCHY;
  }

  $parts = explode(',', $spec);
  $candidates = [];
  foreach ($parts as $p) {
   $f = trim($p);
   if (strcasecmp($f, 'auto') === 0) {
    array_push($candidates, ...self::DEFAULT_FONT_HIERARCHY);
   } elseif ($f !== '') {
    $candidates[] = $f;
   }
  }

  return !empty($candidates) ? $candidates : ['standard'];
 }

 /**
  * @return list<string>
  */
 public static function render(
  string $phrase,
  string $fontSpec = 'standard',
  string $fit = 'clip',
  bool $strict = false,
  int $maxCols = 0,
 ): array {
  $phrase = trim($phrase);
  if ($phrase === '') {
   return [];
  }

  $candidates = self::resolveFontCandidates($fontSpec);
  if ($maxCols <= 0 || count($candidates) === 1) {
   $lines = self::renderFont($phrase, $candidates[0]);
  } else {
   $lines = self::renderResponsive($phrase, $candidates, $maxCols);
  }

  if ($fit === 'scale' && $maxCols > 0) {
   $lines = self::scaleAsciiArt($lines, $maxCols);
  }

  return $lines;
 }

 /**
  * @param list<string> $candidates
  * @return list<string>
  */
 private static function renderResponsive(string $phrase, array $candidates, int $maxCols): array
 {
  foreach ($candidates as $cand) {
   $lines = self::renderFont($phrase, $cand);
   $w = 0;
   foreach ($lines as $l) {
    $lw = Util::visualWidth($l);
    if ($lw > $w) {
     $w = $lw;
    }
   }
   if ($w <= $maxCols) {
    return $lines;
   }
  }

  return self::renderFont($phrase, $candidates[count($candidates) - 1]);
 }

 /**
  * @return list<string>
  */
 public static function renderFont(string $phrase, string $fontName): array
 {
  $resolved = self::resolveFontName($fontName);
  $fontsDir = dirname(__DIR__, 2) . '/assets/fonts';
  $fontPath = $fontsDir . '/' . $resolved . '.flf';

  if (!file_exists($fontPath)) {
   $fontPath = $fontsDir . '/standard.flf';
  }

  if (!isset(self::$fontCache[$fontPath])) {
   try {
    self::$fontCache[$fontPath] = FigletFont::load($fontPath);
   } catch (Throwable) {
    $fontPath = $fontsDir . '/standard.flf';
    if (!isset(self::$fontCache[$fontPath])) {
     self::$fontCache[$fontPath] = FigletFont::load($fontPath);
    }
   }
  }

  $font = self::$fontCache[$fontPath];
  $renderer = new FigletRenderer($font);
  return $renderer->render($phrase);
 }

 /**
  * @param list<string> $lines
  * @return list<string>
  */
 public static function scaleAsciiArt(array $lines, int $targetW): array
 {
  if ($targetW <= 0) {
   return $lines;
  }

  // Trim leading/trailing blank lines
  $start = 0;
  $total = count($lines);
  while ($start < $total && trim($lines[$start]) === '') {
   $start++;
  }
  $end = $total;
  while ($end > $start && trim($lines[$end - 1]) === '') {
   $end--;
  }
  $lines = array_slice($lines, $start, $end - $start);
  if (empty($lines)) {
   return $lines;
  }

  $srcH = count($lines);
  $srcW = 0;
  $rows = [];
  foreach ($lines as $l) {
   $chars = preg_split('//u', $l, -1, PREG_SPLIT_NO_EMPTY) ?: [];
   $rows[] = $chars;
   $len = count($chars);
   if ($len > $srcW) {
    $srcW = $len;
   }
  }
  if ($srcW === 0 || $srcW <= $targetW) {
   return $lines;
  }

  $grid = [];
  for ($y = 0; $y < $srcH; $y++) {
   $rowChars = $rows[$y];
   $g = [];
   for ($x = 0; $x < $srcW; $x++) {
    $ch = $rowChars[$x] ?? ' ';
    $g[$x] = ($ch !== ' ' && $ch !== "\t");
   }
   $grid[$y] = $g;
  }

  $outCols = $targetW;
  $outRows = (int) round(($targetW * $srcH) / (float) $srcW);
  if ($outRows < 1) {
   $outRows = 1;
  }
  $subW = $outCols * 2;
  $subH = $outRows * 3;

  $sxW = (float) $srcW / (float) $subW;
  $sxH = (float) $srcH / (float) $subH;

  $on = function (int $cx, int $cy) use ($grid, $srcW, $srcH, $sxW, $sxH): bool {
   $x0 = $cx * $sxW;
   $x1 = ($cx + 1) * $sxW;
   $y0 = $cy * $sxH;
   $y1 = ($cy + 1) * $sxH;
   $area = 0.0;
   $ink = 0.0;

   for ($yy = (int) floor($y0); $yy < (int) ceil($y1); $yy++) {
    $fy0 = max($y0, (float) $yy);
    $fy1 = min($y1, (float) ($yy + 1));
    $h = $fy1 - $fy0;
    if ($h <= 0 || $yy < 0 || $yy >= $srcH) {
     continue;
    }
    for ($xx = (int) floor($x0); $xx < (int) ceil($x1); $xx++) {
     $fx0 = max($x0, (float) $xx);
     $fx1 = min($x1, (float) ($xx + 1));
     $w = $fx1 - $fx0;
     if ($w <= 0 || $xx < 0 || $xx >= $srcW) {
      continue;
     }
     $a = $w * $h;
     $area += $a;
     if ($grid[$yy][$xx] ?? false) {
      $ink += $a;
     }
    }
   }

   if ($area <= 0) {
    return false;
   }
   return ($ink / $area) >= self::INK_THRESHOLD;
  };

  $out = [];
  for ($row = 0; $row < $outRows; $row++) {
   $sb = '';
   for ($col = 0; $col < $outCols; $col++) {
    $mask = 0;
    if ($on($col * 2, $row * 3)) {
     $mask |= 1;
    }
    if ($on($col * 2 + 1, $row * 3)) {
     $mask |= 2;
    }
    if ($on($col * 2, $row * 3 + 1)) {
     $mask |= 4;
    }
    if ($on($col * 2 + 1, $row * 3 + 1)) {
     $mask |= 8;
    }
    if ($on($col * 2, $row * 3 + 2)) {
     $mask |= 16;
    }
    if ($on($col * 2 + 1, $row * 3 + 2)) {
     $mask |= 32;
    }
    $sb .= self::SCALE_SEXTANT_RUNES[$mask];
   }
   $out[] = $sb;
  }

  return $out;
 }
}
