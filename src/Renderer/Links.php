<?php

declare(strict_types=1);

namespace Termstrap\Renderer;

final class Links
{
 public const MODE_OSC8 = 'osc8';
 public const MODE_INLINE = 'inline';
 public const MODE_FOOTNOTES = 'footnotes';

 /**
  * @return array{0: string, 1: string}
  */
 public static function linkPlaceholders(int $id): array
 {
  $idBase = 1000;
  return [
   sprintf("\x1b[999000;%dm", $id + $idBase),
   sprintf("\x1b[999001;%dm", $id + $idBase),
  ];
 }

 /**
  * @param list<string> $links
  */
 public static function decorateLink(
  string $content,
  string $href,
  array &$links,
  string $linkMode,
  bool $disableHyperlinks,
 ): string {
  if ($href === '') {
   return $content;
  }

  switch ($linkMode) {
   case self::MODE_INLINE:
    $urlStyled = Color::fgEscape('#008787') . "\x1b[4m" . $href . "\x1b[m";
    $full = $content . ' ' . $urlStyled;
    if (!$disableHyperlinks) {
     $id = count($links);
     $links[] = $href;
     [$startMarker, $endMarker] = self::linkPlaceholders($id);
     return $startMarker . $full . $endMarker;
    }
    return $full;

   case self::MODE_FOOTNOTES:
    $id = count($links);
    $links[] = $href;
    $ref = Color::fgEscape('#767676') . sprintf('[%d]', $id + 1) . "\x1b[m";
    if (!$disableHyperlinks) {
     [$startMarker, $endMarker] = self::linkPlaceholders($id);
     return $startMarker . $content . $endMarker . ' ' . $ref;
    }
    return $content . ' ' . $ref;

   default: // OSC8
    if (!$disableHyperlinks) {
     $id = count($links);
     $links[] = $href;
     [$startMarker, $endMarker] = self::linkPlaceholders($id);
     return $startMarker . $content . $endMarker;
    }
    return $content;
  }
 }

 /**
  * @param list<string> $links
  */
 public static function expandLinks(
  string $output,
  array $links,
  string $linkMode,
  bool $disableHyperlinks,
 ): string {
  foreach ($links as $id => $url) {
   [$startPlaceholder, $endPlaceholder] = self::linkPlaceholders($id);
   $osc8Start = sprintf("\x1b]8;id=link_%d;%s\x1b\\", $id, $url);
   $osc8End = "\x1b]8;;\x1b\\";

   $lines = [];
   $inLink = false;
   foreach (explode("\n", $output) as $line) {
    if (str_contains($line, $startPlaceholder)) {
     $inLink = true;
    }
    if ($inLink) {
     $hasEnd = str_contains($line, $endPlaceholder);
     $line = str_replace($startPlaceholder, $osc8Start, $line);
     if (!str_starts_with($line, $osc8Start) && !str_contains($line, $osc8Start)) {
      $line = $osc8Start . $line;
     }
     if ($hasEnd) {
      $line = str_replace($endPlaceholder, $osc8End, $line);
      $inLink = false;
     } else {
      $trimmed = rtrim($line, ' ');
      $spaces = substr($line, strlen($trimmed));
      $line = $trimmed . $osc8End . "\x1b[24m\x1b[m" . $spaces;
     }
    }
    $lines[] = $line;
   }
   $output = implode("\n", $lines);
  }

  if ($linkMode === self::MODE_FOOTNOTES && !empty($links)) {
   $fn = "\n\n";
   foreach ($links as $i => $u) {
    $ref = Color::fgEscape('#767676') . sprintf('[%d]', $i + 1) . "\x1b[m";
    $urlStyled = Color::fgEscape('#008787') . "\x1b[4m" . $u . "\x1b[m";
    if (!$disableHyperlinks) {
     $urlStyled = sprintf("\x1b]8;id=fn_%d;%s\x1b\\%s\x1b]8;;\x1b\\", $i, $u, $urlStyled);
    }
    $fn .= sprintf(" %s %s\n", $ref, $urlStyled);
   }
   $output = rtrim($output, "\n") . $fn;
  }

  return $output;
 }
}
