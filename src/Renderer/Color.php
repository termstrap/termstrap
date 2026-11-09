<?php

declare(strict_types=1);

namespace Termstrap\Renderer;

use Termstrap\Util;

final class Color
{
 /**
  * @return array{0: int, 1: int, 2: int}
  */
 public static function parse(string $color): array
 {
  return Util::hexToRgb($color);
 }

 public static function toHex(int $r, int $g, int $b): string
 {
  return sprintf('#%02x%02x%02x', $r, $g, $b);
 }

 public static function fgEscape(string $color): string
 {
  if ($color === '') {
   return '';
  }
  if ($color[0] === '#' || strlen($color) === 6 || strlen($color) === 3) {
   [$r, $g, $b] = self::parse($color);
   return sprintf("\x1b[38;2;%d;%d;%dm", $r, $g, $b);
  }
  return '';
 }

 public static function bgEscape(string $color): string
 {
  if ($color === '') {
   return '';
  }
  if ($color[0] === '#' || strlen($color) === 6 || strlen($color) === 3) {
   [$r, $g, $b] = self::parse($color);
   return sprintf("\x1b[48;2;%d;%d;%dm", $r, $g, $b);
  }
  return '';
 }

 private static function linearize(float $v): float
 {
  return ($v <= 0.04045) ? $v / 12.92 : pow(($v + 0.055) / 1.055, 2.4);
 }

 private static function delinearize(float $v): float
 {
  return ($v <= 0.0031308) ? 12.92 * $v : (1.055 * pow($v, 1.0 / 2.4) - 0.055);
 }

 /**
  * Convert sRGB [0..255] to CIE XYZ space matching go-colorful
  * @return array{0: float, 1: float, 2: float}
  */
 public static function rgbToXyz(int $r, int $g, int $b): array
 {
  $rf = self::linearize($r / 255.0);
  $gf = self::linearize($g / 255.0);
  $bf = self::linearize($b / 255.0);

  $x = 0.41239079926595948 * $rf + 0.35758433938387796 * $gf + 0.18048078840183429 * $bf;
  $y = 0.21263900587151036 * $rf + 0.71516867876775593 * $gf + 0.072192315360733715 * $bf;
  $z = 0.019330818715591851 * $rf + 0.11919477979462599 * $gf + 0.95053215224966058 * $bf;

  return [$x, $y, $z];
 }

 /**
  * Convert CIE XYZ to sRGB [0..255] matching go-colorful
  * @return array{0: int, 1: int, 2: int}
  */
 public static function xyzToRgb(float $x, float $y, float $z): array
 {
  $r = 3.2409699419045214 * $x - 1.5373831775700935 * $y - 0.49861076029300328 * $z;
  $g = -0.96924363628087983 * $x + 1.8759675015077207 * $y + 0.041555057407175613 * $z;
  $b = 0.055630079696993609 * $x - 0.20397695888897657 * $y + 1.0569715142428786 * $z;

  return [
   (int) round(max(0.0, min(1.0, self::delinearize($r))) * 255.0),
   (int) round(max(0.0, min(1.0, self::delinearize($g))) * 255.0),
   (int) round(max(0.0, min(1.0, self::delinearize($b))) * 255.0),
  ];
 }

 private static function labF(float $t): float
 {
  return ($t > (6.0 / 29.0) ** 3) ? pow($t, 1.0 / 3.0) : ($t / 3.0 * (29.0 / 6.0) ** 2 + 4.0 / 29.0);
 }

 private static function labFinv(float $t): float
 {
  return ($t > 6.0 / 29.0) ? $t ** 3 : (3.0 * (6.0 / 29.0) ** 2 * ($t - 4.0 / 29.0));
 }

 /**
  * Convert CIE XYZ to CIE L*a*b* matching go-colorful (D65 illuminant)
  * @return array{0: float, 1: float, 2: float}
  */
 public static function xyzToLab(float $x, float $y, float $z): array
 {
  $fy = self::labF($y / 1.00000);
  $l = 1.16 * $fy - 0.16;
  $a = 5.0 * (self::labF($x / 0.95047) - $fy);
  $b = 2.0 * ($fy - self::labF($z / 1.08883));

  return [$l, $a, $b];
 }

 /**
  * Convert CIE L*a*b* to CIE XYZ matching go-colorful (D65 illuminant)
  * @return array{0: float, 1: float, 2: float}
  */
 public static function labToXyz(float $l, float $a, float $b): array
 {
  $l2 = ($l + 0.16) / 1.16;
  $x = 0.95047 * self::labFinv($l2 + $a / 5.0);
  $y = 1.00000 * self::labFinv($l2);
  $z = 1.08883 * self::labFinv($l2 - $b / 2.0);

  return [$x, $y, $z];
 }

 /**
  * Blend two RGB colors in CIELAB space matching go-colorful BlendLab
  * @param array{0: int, 1: int, 2: int} $c1
  * @param array{0: int, 1: int, 2: int} $c2
  * @return array{0: int, 1: int, 2: int}
  */
 public static function blendLab(array $c1, array $c2, float $t): array
 {
  [$l1, $a1, $b1] = self::xyzToLab(...self::rgbToXyz($c1[0], $c1[1], $c1[2]));
  [$l2, $a2, $b2] = self::xyzToLab(...self::rgbToXyz($c2[0], $c2[1], $c2[2]));

  $l = $l1 + $t * ($l2 - $l1);
  $a = $a1 + $t * ($a2 - $a1);
  $b = $b1 + $t * ($b2 - $b1);

  return self::xyzToRgb(...self::labToXyz($l, $a, $b));
 }

 /**
  * Lipgloss Blend1D equivalent in CIELAB
  * @param list<string|array{0: int, 1: int, 2: int}> $stops
  * @return list<array{0: int, 1: int, 2: int}>
  */
 public static function blend1D(int $steps, array $stops): array
 {
  if ($steps <= 0) {
   return [];
  }
  if (empty($stops)) {
   return array_fill(0, $steps, [255, 255, 255]);
  }

  $rgbStops = [];
  foreach ($stops as $s) {
   $rgbStops[] = is_string($s) ? self::parse($s) : $s;
  }

  if ($steps <= count($rgbStops)) {
   return array_slice($rgbStops, 0, $steps);
  }
  if (count($rgbStops) === 1) {
   return array_fill(0, $steps, $rgbStops[0]);
  }

  $numSegments = count($rgbStops) - 1;
  $defaultSize = (int) ($steps / $numSegments);
  $remainingSteps = $steps % $numSegments;

  $blended = [];
  for ($i = 0; $i < $numSegments; $i++) {
   $from = $rgbStops[$i];
   $to = $rgbStops[$i + 1];

   $segmentSize = $defaultSize;
   if ($i < $remainingSteps) {
    $segmentSize++;
   }

   $divisor = (float) max(1, $segmentSize - 1);
   for ($j = 0; $j < $segmentSize; $j++) {
    $factor = $segmentSize > 1 ? ((float) $j) / $divisor : 0.0;
    $blended[] = self::blendLab($from, $to, $factor);
   }
  }

  return $blended;
 }
}
