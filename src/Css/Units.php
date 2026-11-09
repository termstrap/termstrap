<?php

declare(strict_types=1);

namespace Termstrap\Css;

final class Units
{
 public const LENGTH_FIT_CONTENT = -1;

 public static function resolveLength(string $val, Axis $axis, UnitContext $ctx): int
 {
  $val = trim($val);
  if ($val === 'fit-content' || $val === 'max-content' || $val === 'min-content') {
   return self::LENGTH_FIT_CONTENT;
  }
  if ($val === '' || $val === 'auto' || $val === 'inherit' || $val === 'initial') {
   return 0;
  }

  // 1. Pure unitless number (e.g. "0", "1", "2")
  if (is_numeric($val)) {
   $num = (float) $val;
   if ($num == 0.0) {
    return 0;
   }
   if ($num > 0 && $num < 1 && $axis === Axis::X) {
    return 1;
   }
   return (int) round($num);
  }

  // 2. Parse numeric value and unit suffix
  [$numPart, $unitPart] = self::splitNumberAndUnit($val);
  if ($numPart === '' || !is_numeric($numPart)) {
   return 0;
  }

  $num = (float) $numPart;
  if ($num == 0.0) {
   return 0;
  }

  $rootFS = $ctx->rootFontSize > 0 ? $ctx->rootFontSize : 16.0;
  $unitPart = strtolower($unitPart);
  $px = 0.0;
  $isRelativePercent = false;
  $percentFraction = 0.0;

  switch ($unitPart) {
   case 'px':
    $px = $num;
    break;
   case 'rem':
   case 'em':
    $px = $num * $rootFS;
    break;
   case 'ch':
    // 1ch is character advance width (approx 0.5rem = 8px)
    $px = $num * ($rootFS * 0.5);
    break;
   case 'pt':
    $px = $num * (96.0 / 72.0);
    break;
   case 'in':
    $px = $num * 96.0;
    break;
   case 'cm':
    $px = $num * (96.0 / 2.54);
    break;
   case 'mm':
    $px = $num * (96.0 / 25.4);
    break;
   case '%':
    $isRelativePercent = true;
    $percentFraction = $num / 100.0;
    break;
   case 'vw':
    $tw = $ctx->terminalWidth > 0 ? $ctx->terminalWidth : 80;
    return (int) round(($num / 100.0) * $tw);
   case 'vh':
    $th = $ctx->terminalHeight > 0 ? $ctx->terminalHeight : 24;
    return (int) round(($num / 100.0) * $th);
   default:
    $px = $num;
    break;
  }

  // 3. Project to terminal cells
  if ($axis === Axis::X) {
   if ($isRelativePercent) {
    $pw = $ctx->parentWidth > 0 ? $ctx->parentWidth : ($ctx->terminalWidth > 0 ? $ctx->terminalWidth : 80);
    $cols = round($percentFraction * $pw);
    if ($cols < 1 && $percentFraction > 0) {
     return 1;
    }
    if ($cols < 0) {
     return 0;
    }
    return (int) $cols;
   }

   // Horizontal: 1 column cell ≈ 8px (0.5rem)
   $colsFloat = $px / ($rootFS * 0.5);
   if ($colsFloat <= 0) {
    return 0;
   }
   if ($colsFloat < 1.0) {
    // Minimum visual presence rule
    return 1;
   }
   return (int) round($colsFloat);
  }

  // Axis::Y (vertical lines)
  if ($isRelativePercent) {
   $ph = $ctx->parentHeight > 0 ? $ctx->parentHeight : ($ctx->terminalHeight > 0 ? $ctx->terminalHeight : 24);
   $lines = round($percentFraction * $ph);
   if ($lines < 0) {
    return 0;
   }
   return (int) $lines;
  }

  // Vertical: 1 terminal row ≈ 16px (1.0rem)
  $linesFloat = $px / $rootFS;
  if ($linesFloat <= 0) {
   return 0;
  }
  // Sub-line vertical threshold: below 0.4rem return 0 to avoid wasteful blank lines
  if ($linesFloat < 0.4) {
   return 0;
  }

  $lines = (int) round($linesFloat);
  // Damping rule: limit maximum vertical padding/margin to 3 lines
  if ($lines > 3) {
   $lines = 3;
  }

  return $lines;
 }

 /**
  * @return array{0: string, 1: string}
  */
 public static function splitNumberAndUnit(string $val): array
 {
  $val = trim($val);
  if ($val === '') {
   return ['', ''];
  }

  $i = 0;
  $n = strlen($val);

  if ($i < $n && ($val[$i] === '+' || $val[$i] === '-')) {
   $i++;
  }

  $hasDigits = false;
  while ($i < $n && (ctype_digit($val[$i]) || $val[$i] === '.')) {
   $hasDigits = true;
   $i++;
  }

  if (!$hasDigits) {
   return ['', $val];
  }

  $num = substr($val, 0, $i);
  $unit = trim(substr($val, $i));
  return [$num, $unit];
 }

 /**
  * @return array{0: int, 1: int, 2: int, 3: int} (top, right, bottom, left)
  */
 public static function resolveBox4D(string $val, UnitContext $ctx): array
 {
  $val = trim($val);
  if ($val === '') {
   return [0, 0, 0, 0];
  }

  $fields = preg_split('/\s+/', $val, -1, PREG_SPLIT_NO_EMPTY);
  if (empty($fields)) {
   return [0, 0, 0, 0];
  }

  return match (count($fields)) {
   1 => [
    self::resolveLength($fields[0], Axis::Y, $ctx),
    self::resolveLength($fields[0], Axis::X, $ctx),
    self::resolveLength($fields[0], Axis::Y, $ctx),
    self::resolveLength($fields[0], Axis::X, $ctx),
   ],
   2 => [
    self::resolveLength($fields[0], Axis::Y, $ctx),
    self::resolveLength($fields[1], Axis::X, $ctx),
    self::resolveLength($fields[0], Axis::Y, $ctx),
    self::resolveLength($fields[1], Axis::X, $ctx),
   ],
   3 => [
    self::resolveLength($fields[0], Axis::Y, $ctx),
    self::resolveLength($fields[1], Axis::X, $ctx),
    self::resolveLength($fields[2], Axis::Y, $ctx),
    self::resolveLength($fields[1], Axis::X, $ctx),
   ],
   default => [
    self::resolveLength($fields[0], Axis::Y, $ctx),
    self::resolveLength($fields[1], Axis::X, $ctx),
    self::resolveLength($fields[2], Axis::Y, $ctx),
    self::resolveLength($fields[3], Axis::X, $ctx),
   ],
  };
 }

 /**
  * @return array{0: bool, 1: bool, 2: float} (bold, dim, scale)
  */
 public static function resolveFontSize(string $val, float $rootFS = 16.0): array
 {
  $val = trim($val);
  if ($val === '') {
   return [false, false, 1.0];
  }
  if ($rootFS <= 0) {
   $rootFS = 16.0;
  }

  [$numPart, $unitPart] = self::splitNumberAndUnit($val);
  if ($numPart === '' || !is_numeric($numPart)) {
   $lower = strtolower($val);
   return match ($lower) {
    'xx-large', 'x-large', 'larger', 'large' => [true, false, 1.5],
    'small', 'x-small', 'xx-small', 'smaller' => [false, true, 0.8],
    default => [false, false, 1.0],
   };
  }

  $num = (float) $numPart;
  $unitPart = strtolower($unitPart);
  $px = match ($unitPart) {
   'px' => $num,
   'rem', 'em' => $num * $rootFS,
   'pt' => $num * (96.0 / 72.0),
   '%' => ($num / 100.0) * $rootFS,
   default => $num,
  };

  $ratio = $px / $rootFS;
  if ($ratio >= 1.25) {
   return [true, false, $ratio];
  }
  if ($ratio <= 0.85) {
   return [false, true, $ratio];
  }

  return [false, false, $ratio];
 }

 public static function checkIsLengthUnit(string $token): bool
 {
  $token = strtolower(trim($token));
  foreach (['px', 'rem', 'em', 'pt', 'in', 'cm', 'mm', '%', 'vw', 'vh', 'ch'] as $u) {
   if (str_ends_with($token, $u)) {
    $prefix = substr($token, 0, -strlen($u));
    if (is_numeric($prefix)) {
     return true;
    }
   }
  }
  return false;
 }
}
