<?php

declare(strict_types=1);

namespace Termstrap\Renderer;

use Termstrap\Util;

final class Background
{
 public static function persistBackgroundColor(string $out, string $bgColor): string
 {
  if ($bgColor === '' || $out === '') {
   return $out;
  }

  $bgSeq = Color::bgEscape($bgColor);
  if ($bgSeq === '') {
   return $out;
  }

  $lines = explode("\n", $out);
  $bgCode = trim($bgSeq, "\x1b[m");

  foreach ($lines as $i => $line) {
   if (!str_contains($line, $bgSeq) && !str_contains($line, $bgCode)) {
    continue;
   }

   $line = str_replace("\x1b[49m", $bgSeq, $line);

   $lastReset = strrpos($line, "\x1b[0m");
   if ($lastReset === false) {
    $lastReset = strrpos($line, "\x1b[m");
    if ($lastReset === false) {
     $lines[$i] = $line;
     continue;
    }
   }

   $prefix = substr($line, 0, $lastReset);
   $suffix = substr($line, $lastReset);

   $prefix = str_replace("\x1b[0m", "\x1b[0m" . $bgSeq, $prefix);
   $prefix = str_replace("\x1b[m", "\x1b[m" . $bgSeq, $prefix);
   $prefix = str_replace($bgSeq . $bgSeq, $bgSeq, $prefix);

   $lines[$i] = $prefix . $suffix;
  }

  return implode("\n", $lines);
 }

 /**
  * @param list<string> $stopHexes
  */
 public static function applyBackgroundGradient(string $out, array $stopHexes): string
 {
  if ($out === '' || count($stopHexes) < 2) {
   return $out;
  }

  $lines = explode("\n", $out);
  $maxW = 0;
  foreach ($lines as $l) {
   $w = Util::visualWidth($l);
   if ($w > $maxW) {
    $maxW = $w;
   }
  }

  if ($maxW <= 0) {
   return $out;
  }

  $blended = Color::blend1D($maxW, $stopHexes);
  $bgSeqs = [];
  foreach ($blended as [$r, $g, $b]) {
   $bgSeqs[] = sprintf("\x1b[48;2;%d;%d;%dm", $r, $g, $b);
  }

  foreach ($lines as $lineIdx => $line) {
   $sb = '';
   $col = 0;
   $i = 0;
   $len = strlen($line);

   while ($i < $len) {
    if (preg_match(Util::ANSI_REGEX, substr($line, $i), $m, PREG_OFFSET_CAPTURE) && $m[0][1] === 0) {
     $sb .= $m[0][0];
     $i += strlen($m[0][0]);
     continue;
    }

    $ch = mb_substr(substr($line, $i), 0, 1, 'UTF-8');
    $chLen = strlen($ch);

    if (isset($bgSeqs[$col])) {
     $sb .= $bgSeqs[$col];
    }
    $sb .= $ch;
    $col += Util::visualWidth($ch);
    $i += $chLen;
   }

   $sb .= "\x1b[0m";
   $lines[$lineIdx] = $sb;
  }

  return implode("\n", $lines);
 }
}
