<?php

declare(strict_types=1);

namespace Termstrap\Css;

use Sabberworm\CSS\CSSList\AtRuleBlockList;
use Sabberworm\CSS\CSSList\Document;
use Sabberworm\CSS\OutputFormat;
use Sabberworm\CSS\RuleSet\DeclarationBlock;

final class Preprocessor
{
 /** @var array<string, string> */
 public array $globalVars = [];

 public function __construct(
  public string $theme = 'bootstrap',
 ) {}

 public function extractVariables(Document $sheet): void
 {
  $themeStr = strtolower(trim($this->theme));
  if ($themeStr === '') {
   $themeStr = 'bootstrap';
  }

  $format = new OutputFormat();

  foreach ($sheet->getAllDeclarationBlocks() as $block) {
   $isRootScope = false;
   foreach ($block->getSelectors() as $selObj) {
    $s = trim((string) $selObj->getSelector());
    if ($s === ':root' || $s === '::root' || $s === 'html' || $s === 'body' || $s === '*') {
     $isRootScope = true;
     break;
    }
    if (str_starts_with($s, '[data-bs-theme=')) {
     $target = trim(substr($s, 15), ']"\' ');
     if (strcasecmp($target, $themeStr) === 0) {
      $isRootScope = true;
      break;
     }
     if ($themeStr === 'light' && $target === 'light') {
      $isRootScope = true;
      break;
     }
     if (in_array($themeStr, ['bootstrap', 'dark', 'none', ''], true) && $target === 'dark') {
      $isRootScope = true;
      break;
     }
    }
   }

   if ($isRootScope) {
    foreach ($block->getRules() as $rule) {
     $prop = trim($rule->getRule());
     if (str_starts_with($prop, '--')) {
      $val = $rule->getValue();
      $strVal = is_string($val) ? $val : $val->render($format);
      $this->globalVars[$prop] = trim($strVal);
     }
    }
   }
  }

  $this->resolveGlobalVarDependencies();
 }

 private function resolveGlobalVarDependencies(): void
 {
  $lookup = fn(string $name): ?string => $this->globalVars[$name] ?? null;

  for ($pass = 0; $pass < 10; $pass++) {
   $changed = false;
   foreach ($this->globalVars as $k => $v) {
    if (str_contains($v, 'var(')) {
     $resolved = self::resolveVariables($v, $lookup);
     if ($resolved !== $v) {
      $this->globalVars[$k] = $resolved;
      $changed = true;
     }
    }
   }
   if (!$changed) {
    break;
   }
  }
 }

 /**
  * @param array<string, string> $localVars
  */
 public function preprocessValue(string $val, array $localVars = []): string
 {
  $val = trim($val);
  $lookup = fn(string $name): ?string => $localVars[$name] ?? ($this->globalVars[$name] ?? null);

  // 1. Resolve var(--*)
  if (str_contains($val, 'var(')) {
   $val = self::resolveVariables($val, $lookup);
  }

  // 2. Evaluate calc(...)
  if (str_contains(strtolower($val), 'calc(')) {
   $val = self::evaluateCalc($val);
  }

  // 3. Resolve rgba(...) and rgb(...)
  $valLower = strtolower($val);
  if (str_contains($valLower, 'rgba(') || str_contains($valLower, 'rgb(')) {
   $bgHex = in_array(strtolower($this->theme), ['dracula', 'tokyonight', 'glamour'], true)
    ? '#1a1b26'
    : '#ffffff';
   $val = self::resolveColors($val, $bgHex);
  }

  return $val;
 }

 /**
  * @param callable(string): (?string) $lookup
  */
 public static function resolveVariables(string $value, callable $lookup): string
 {
  return self::resolveVariablesWithDepth($value, $lookup, []);
 }

 /**
  * @param callable(string): (?string) $lookup
  * @param array<string, bool> $visited
  */
 private static function resolveVariablesWithDepth(string $value, callable $lookup, array $visited): string
 {
  if (!str_contains($value, 'var(')) {
   return $value;
  }

  $sb = '';
  $idx = 0;
  $n = strlen($value);

  while ($idx < $n) {
   $pos = strpos($value, 'var(', $idx);
   if ($pos === false) {
    $sb .= substr($value, $idx);
    break;
   }

   $sb .= substr($value, $idx, $pos - $idx);

   $contentStart = $pos + 4; // len("var(")
   $parenCount = 1;
   $closeIdx = -1;

   for ($i = $contentStart; $i < $n; $i++) {
    if ($value[$i] === '(') {
     $parenCount++;
    } elseif ($value[$i] === ')') {
     $parenCount--;
     if ($parenCount === 0) {
      $closeIdx = $i;
      break;
     }
    }
   }

   if ($closeIdx === -1) {
    // Malformed var()
    $sb .= substr($value, $pos);
    break;
   }

   $inside = substr($value, $contentStart, $closeIdx - $contentStart);
   [$varName, $fallback] = self::splitVarFallback($inside);
   $varName = trim($varName);

   $resolved = '';
   $found = false;

   if (!isset($visited[$varName])) {
    $val = $lookup($varName);
    if ($val !== null && $val !== '') {
     $visited[$varName] = true;
     $resolved = self::resolveVariablesWithDepth($val, $lookup, $visited);
     unset($visited[$varName]);
     $found = true;
    }
   }

   if (!$found && $fallback !== '') {
    $resolved = self::resolveVariablesWithDepth($fallback, $lookup, $visited);
   }

   $sb .= $resolved;
   $idx = $closeIdx + 1;
  }

  return $sb;
 }

 /**
  * @return array{0: string, 1: string}
  */
 private static function splitVarFallback(string $content): array
 {
  $parenCount = 0;
  $len = strlen($content);
  for ($i = 0; $i < $len; $i++) {
   if ($content[$i] === '(') {
    $parenCount++;
   } elseif ($content[$i] === ')') {
    $parenCount--;
   } elseif ($content[$i] === ',' && $parenCount === 0) {
    return [substr($content, 0, $i), trim(substr($content, $i + 1))];
   }
  }
  return [$content, ''];
 }

 public static function evaluateCalc(string $input): string
 {
  $lower = strtolower($input);
  if (!str_contains($lower, 'calc(')) {
   return $input;
  }

  $sb = '';
  $idx = 0;
  $n = strlen($input);

  while ($idx < $n) {
   $pos = strpos(strtolower(substr($input, $idx)), 'calc(');
   if ($pos === false) {
    $sb .= substr($input, $idx);
    break;
   }

   $calcStart = $idx + $pos;
   $sb .= substr($input, $idx, $calcStart - $idx);

   $contentStart = $calcStart + 5;
   $parenCount = 1;
   $closeIdx = -1;

   for ($i = $contentStart; $i < $n; $i++) {
    if ($input[$i] === '(') {
     $parenCount++;
    } elseif ($input[$i] === ')') {
     $parenCount--;
     if ($parenCount === 0) {
      $closeIdx = $i;
      break;
     }
    }
   }

   if ($closeIdx === -1) {
    $sb .= substr($input, $calcStart);
    break;
   }

   $inside = substr($input, $contentStart, $closeIdx - $contentStart);
   $res = self::evalCalcExpression($inside);
   $sb .= $res;
   $idx = $closeIdx + 1;
  }

  return $sb;
 }

 private static function evalCalcExpression(string $expr): string
 {
  $expr = trim($expr);
  if ($expr === '') {
   return '0';
  }

  $tokens = [];
  $i = 0;
  $m = strlen($expr);

  while ($i < $m) {
   $ch = $expr[$i];
   if (ctype_space($ch)) {
    $i++;
    continue;
   }

   // Operators
   if ($ch === '*' || $ch === '/' || $ch === '+') {
    $tokens[] = ['isOp' => true, 'op' => $ch, 'num' => 0.0, 'unit' => ''];
    $i++;
    continue;
   }

   if ($ch === '-') {
    if (!empty($tokens) && !$tokens[array_key_last($tokens)]['isOp']) {
     $tokens[] = ['isOp' => true, 'op' => '-', 'num' => 0.0, 'unit' => ''];
     $i++;
     continue;
    }
   }

   // Number parsing
   $start = $i;
   if ($ch === '-' || $ch === '+') {
    $i++;
   }
   $hasDigit = false;
   while ($i < $m && (ctype_digit($expr[$i]) || $expr[$i] === '.')) {
    $hasDigit = true;
    $i++;
   }

   if ($hasDigit) {
    $numStr = substr($expr, $start, $i - $start);
    $num = (float) $numStr;

    $unitStart = $i;
    while ($i < $m && (ctype_alpha($expr[$i]) || $expr[$i] === '%')) {
     $i++;
    }
    $unit = substr($expr, $unitStart, $i - $unitStart);

    $tokens[] = ['isOp' => false, 'op' => '', 'num' => $num, 'unit' => $unit];
    continue;
   }

   $i++;
  }

  if (empty($tokens)) {
   return $expr;
  }

  // Pass 1: Handle multiplication and division (*, /)
  $pass1 = [];
  $k = 0;
  $tCount = count($tokens);

  while ($k < $tCount) {
   $tok = $tokens[$k];
   if ($tok['isOp'] && ($tok['op'] === '*' || $tok['op'] === '/')) {
    if (!empty($pass1) && $k + 1 < $tCount && !$tokens[$k + 1]['isOp']) {
     $prev = array_pop($pass1);
     $next = $tokens[$k + 1];
     $resNum = 0.0;
     $resUnit = $prev['unit'] !== '' ? $prev['unit'] : $next['unit'];

     if ($tok['op'] === '*') {
      $resNum = $prev['num'] * $next['num'];
     } elseif ($next['num'] != 0.0) {
      $resNum = $prev['num'] / $next['num'];
     }

     $pass1[] = ['isOp' => false, 'op' => '', 'num' => $resNum, 'unit' => $resUnit];
     $k += 2;
     continue;
    }
   }
   $pass1[] = $tok;
   $k++;
  }

  if (empty($pass1)) {
   return '0';
  }

  // Pass 2: Handle addition and subtraction (+, -)
  $acc = $pass1[0];
  $k = 1;
  $pCount = count($pass1);

  while ($k < $pCount) {
   $opTok = $pass1[$k];
   if ($opTok['isOp'] && $k + 1 < $pCount) {
    $nextTok = $pass1[$k + 1];
    if (!$nextTok['isOp']) {
     $unit = $acc['unit'] !== '' ? $acc['unit'] : $nextTok['unit'];
     if ($opTok['op'] === '+') {
      $acc = ['isOp' => false, 'op' => '', 'num' => $acc['num'] + $nextTok['num'], 'unit' => $unit];
     } elseif ($opTok['op'] === '-') {
      $acc = ['isOp' => false, 'op' => '', 'num' => $acc['num'] - $nextTok['num'], 'unit' => $unit];
     }
     $k += 2;
     continue;
    }
   }
   $k++;
  }

  $resStr = rtrim(rtrim(sprintf('%.4f', $acc['num']), '0'), '.');
  if ($resStr === '' || $resStr === '-0') {
   $resStr = '0';
  }
  return $resStr . $acc['unit'];
 }

 public static function resolveColors(string $input, string $bgHex): string
 {
  $lower = strtolower($input);
  if (!str_contains($lower, 'rgb')) {
   return $input;
  }

  $sb = '';
  $idx = 0;
  $n = strlen($input);

  while ($idx < $n) {
   $sub = strtolower(substr($input, $idx));
   $posRGBA = strpos($sub, 'rgba(');
   $posRGB = strpos($sub, 'rgb(');

   $pos = -1;
   $isRGBA = false;
   if ($posRGBA !== false && ($posRGB === false || $posRGBA <= $posRGB)) {
    $pos = $posRGBA;
    $isRGBA = true;
   } elseif ($posRGB !== false) {
    $pos = $posRGB;
    $isRGBA = false;
   }

   if ($pos === -1) {
    $sb .= substr($input, $idx);
    break;
   }

   $callStart = $idx + $pos;
   $sb .= substr($input, $idx, $callStart - $idx);

   $fnLen = $isRGBA ? 5 : 4;
   $contentStart = $callStart + $fnLen;
   $parenCount = 1;
   $closeIdx = -1;

   for ($i = $contentStart; $i < $n; $i++) {
    if ($input[$i] === '(') {
     $parenCount++;
    } elseif ($input[$i] === ')') {
     $parenCount--;
     if ($parenCount === 0) {
      $closeIdx = $i;
      break;
     }
    }
   }

   if ($closeIdx === -1) {
    $sb .= substr($input, $callStart);
    break;
   }

   $inside = substr($input, $contentStart, $closeIdx - $contentStart);
   $parsed = self::parseRgba($inside);
   if ($parsed !== null) {
    [$r, $g, $b, $a] = $parsed;
    $sb .= self::blendRgba($r, $g, $b, $a, $bgHex);
   } else {
    $sb .= substr($input, $callStart, $closeIdx + 1 - $callStart);
   }

   $idx = $closeIdx + 1;
  }

  return $sb;
 }

 /**
  * @return array{0: int, 1: int, 2: int, 3: float}|null
  */
 public static function parseRgba(string $str): ?array
 {
  $str = str_replace('/', ',', $str);
  $rawParts = explode(',', $str);
  $parts = [];
  foreach ($rawParts as $p) {
   $f = preg_split('/\s+/', trim($p), -1, PREG_SPLIT_NO_EMPTY);
   if ($f !== false) {
    foreach ($f as $token) {
     $parts[] = $token;
    }
   }
  }

  if (count($parts) < 3) {
   return null;
  }

  $parseChan = function (string $s): ?int {
   $s = rtrim(trim($s), '%');
   if (!is_numeric($s)) {
    return null;
   }
   $v = (float) $s;
   $v = max(0.0, min(255.0, $v));
   return (int) round($v);
  };

  $rVal = $parseChan($parts[0]);
  $gVal = $parseChan($parts[1]);
  $bVal = $parseChan($parts[2]);

  if ($rVal === null || $gVal === null || $bVal === null) {
   return null;
  }

  $alpha = 1.0;
  if (isset($parts[3])) {
   $s = trim($parts[3]);
   if (str_ends_with($s, '%')) {
    $s = substr($s, 0, -1);
    if (is_numeric($s)) {
     $alpha = ((float) $s) / 100.0;
    }
   } elseif (is_numeric($s)) {
    $alpha = (float) $s;
   }
   $alpha = max(0.0, min(1.0, $alpha));
  }

  return [$rVal, $gVal, $bVal, $alpha];
 }

 public static function blendRgba(int $r, int $g, int $b, float $a, string $bgHex): string
 {
  if ($a >= 0.999) {
   return sprintf('#%02x%02x%02x', $r, $g, $b);
  }
  if ($a <= 0.001) {
   return 'transparent';
  }

  [$bgR, $bgG, $bgB] = self::parseHexRgb($bgHex);
  $blendR = (int) round($r * $a + $bgR * (1.0 - $a));
  $blendG = (int) round($g * $a + $bgG * (1.0 - $a));
  $blendB = (int) round($b * $a + $bgB * (1.0 - $a));

  return sprintf('#%02x%02x%02x', $blendR, $blendG, $blendB);
 }

 /**
  * @return array{0: int, 1: int, 2: int}
  */
 private static function parseHexRgb(string $hex): array
 {
  $hex = ltrim($hex, '#');
  if (strlen($hex) === 3) {
   $hex = $hex[0] . $hex[0] . $hex[1] . $hex[1] . $hex[2] . $hex[2];
  }
  if (strlen($hex) === 6) {
   $val = hexdec($hex);
   return [(int) ($val >> 16), (int) (($val >> 8) & 0xff), (int) ($val & 0xff)];
  }
  return [255, 255, 255];
 }
}
