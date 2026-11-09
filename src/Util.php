<?php

declare(strict_types=1);

namespace Termstrap;

use Symfony\Component\Tui\Ansi\AnsiUtils;

final class Util
{
 /**
  * Matches ANSI escape sequences (CSI and OSC sequences).
  */
 public const ANSI_REGEX = '/\x1b(\[[0-9;:?]*[a-zA-Z]|\][^\x1b\x07]*(\x1b\\\\|\x07))/';

 /**
  * Space immediately preceding punctuation regex.
  */
 public const SPACE_BEFORE_PUNCT_REGEX = '/ (\x1b\[[0-9;:?]*[a-zA-Z]|\][^\x1b\x07]*(\x1b\\\\|\x07))*([,;:!?.])/';

 /**
  * Strip ANSI escape sequences from string.
  */
 public static function stripAnsi(string $s): string
 {
  $stripped = preg_replace(self::ANSI_REGEX, '', $s);
  return trim($stripped ?? '');
 }

 /**
  * Replace breaking space preceding punctuation with non-breaking space.
  */
 public static function gluePunctuation(string $s): string
 {
  return (string) preg_replace(self::SPACE_BEFORE_PUNCT_REGEX, "\u{00a0}$1$3", $s);
 }

 /**
  * HTML-escapes angle brackets and ampersands that appear inside Markdown code spans
  * so the HTML parser does not treat tag-like text such as <hr> or <h1> as real elements.
  */
 public static function escapeInlineCodeHtml(string $s): string
 {
  $len = strlen($s);
  $out = '';
  $i = 0;

  while ($i < $len) {
   if ($s[$i] !== '`') {
    $out .= $s[$i];
    $i++;
    continue;
   }

   $open = $i;
   while ($i < $len && $s[$i] === '`') {
    $i++;
   }
   $runLen = $i - $open;
   $contentStart = $i;
   $closeStart = -1;

   $j = $i;
   while ($j < $len) {
    if ($s[$j] !== '`') {
     $j++;
     continue;
    }
    $runStart = $j;
    while ($j < $len && $s[$j] === '`') {
     $j++;
    }
    if ($j - $runStart === $runLen) {
     $closeStart = $runStart;
     $i = $j;
     break;
    }
   }

   $out .= substr($s, $open, $contentStart - $open);
   if ($closeStart < 0) {
    // Unterminated backtick run
    continue;
   }

   $content = substr($s, $contentStart, $closeStart - $contentStart);
   $content = str_replace(['&', '<', '>'], ['&amp;', '&lt;', '&gt;'], $content);
   $out .= $content;
   $out .= substr($s, $closeStart, $i - $closeStart);
  }

  return $out;
 }

 /**
  * Visible width in terminal character columns.
  */
 public static function visualWidth(string $s): int
 {
  return AnsiUtils::visibleWidth($s);
 }

 /**
  * Parse hex color to RGB tuple (0-255).
  *
  * @return array{0: int, 1: int, 2: int}
  */
 public static function hexToRgb(string $hex): array
 {
  $hex = ltrim($hex, '#');
  if (strlen($hex) === 3) {
   $hex = $hex[0] . $hex[0] . $hex[1] . $hex[1] . $hex[2] . $hex[2];
  }

  if (strlen($hex) !== 6) {
   return [255, 255, 255];
  }

  $r = (int) hexdec(substr($hex, 0, 2));
  $g = (int) hexdec(substr($hex, 2, 2));
  $b = (int) hexdec(substr($hex, 4, 2));

  return [$r, $g, $b];
 }
}
