<?php

declare(strict_types=1);

namespace Termstrap;

use Dom\Element;
use Dom\HTMLDocument;
use Termstrap\Ascii\Figlet;
use Termstrap\Image\ImageOverlay;
use Termstrap\Renderer\DocumentRenderer;
use Throwable;

final class Termstrap
{
 public const STICKY_SELECTOR = '.sticky, .sticky-top, .sticky-bottom, .position-sticky, .position-fixed, .fixed-top, .fixed-bottom, [style*="position:sticky"], [style*="position: sticky"], [style*="position:fixed"], [style*="position: fixed"]';

 /**
  * @param string|null $libraryPath Kept for backward compatibility with callers
  */
 public function __construct(private ?string $libraryPath = null) {}

 /**
  * Render HTML/CSS/Markdown content to terminal ANSI output.
  *
  * @param string $content HTML or Markdown content
  * @param array<string, mixed> $options
  * @return string ANSI rendered string
  */
 public function render(string $content, array $options = []): string
 {
  return self::toAnsi($content, $options);
 }

 /**
  * Static helper for quick rendering.
  *
  * @param string $content HTML or Markdown content
  * @param array<string, mixed> $options
  */
 public static function toAnsi(string $content, array $options = []): string
 {
  [$rendered] = DocumentRenderer::render($content, $options);
  return $rendered;
 }

 /**
  * Render with separated image overlays.
  *
  * @param string $content
  * @param array<string, mixed> $options
  * @return array{0: string, 1: list<ImageOverlay>}
  */
 public function renderWithOverlays(string $content, array $options = []): array
 {
  $rawOverlays = [];
  [$rendered, $rawOverlays] = DocumentRenderer::render($content, $options, $rawOverlays);

  $overlays = [];
  foreach ($rawOverlays as $o) {
   $overlays[] = new ImageOverlay(
    row: $o['row'],
    col: $o['col'],
    width: $o['width'],
    height: $o['height'],
    escape: $o['escape'],
   );
  }

  return [$rendered, $overlays];
 }

 /**
  * Render document into scrollable body and return line offsets of sticky sections.
  *
  * @param string $content
  * @param array<string, mixed> $options
  * @return array{0: string, 1: list<StickyRegion>}
  */
 public static function renderStickyRegions(string $content, array $options = []): array
 {
  $isMarkdown = (bool) ($options['markdown'] ?? false);
  if ($isMarkdown) {
   $content = Markdown::toHtml($content);
   $options['markdown'] = false;
  }

  try {
   $escaped = Util::escapeInlineCodeHtml($content);
   $doc = HTMLDocument::createFromString($escaped, LIBXML_NOERROR);
  } catch (Throwable) {
   return [self::toAnsi($content, $options), []];
  }

  $bodies = $doc->getElementsByTagName('body');
  $body = $bodies->length > 0 ? $bodies->item(0) : null;
  if ($body === null) {
   return [self::toAnsi($content, $options), []];
  }

  $sections = [];
  for ($c = $body->firstChild; $c !== null; $c = $c->nextSibling) {
   if ($c instanceof Element) {
    $sections[] = $c;
   }
  }

  if (empty($sections)) {
   return [self::toAnsi($content, $options), []];
  }

  // Check if any element matches sticky selector
  $hasSticky = false;
  foreach ($sections as $s) {
   if (self::elementMatchesSticky($s)) {
    $hasSticky = true;
    break;
   }
  }
  if (!$hasSticky) {
   return [self::toAnsi($content, $options), []];
  }

  $parts = [];
  $regions = [];
  $line = 0;

  foreach ($sections as $s) {
   $outerHtml = $s->ownerDocument->saveHTML($s);
   $segText = self::toAnsi($outerHtml, $options);

   $stickyEl = self::findStickyElement($s);
   if ($stickyEl !== null) {
    $headerHtml = $outerHtml;
    if ($stickyEl !== $s) {
     $stickyOuter = $s->ownerDocument->saveHTML($stickyEl);
     $cls = $s->getAttribute('class') ?? '';
     $open = '<' . $s->tagName;
     if ($cls !== '') {
      $open .= ' class="' . htmlspecialchars($cls) . '"';
     }
     $headerHtml = $open . '>' . $stickyOuter . '</' . $s->tagName . '>';
    }

    $hdr = rtrim(self::toAnsi($headerHtml, $options), "\n");
    $hdrLines = explode("\n", $hdr);
    $hdrHeight = count($hdrLines);

    $regions[] = new StickyRegion(
     startLine: $line,
     height: $hdrHeight,
     header: $hdr,
    );
   }

   $parts[] = $segText;
   $line += count(explode("\n", $segText));
  }

  return [implode("\n", $parts), $regions];
 }

 private static function elementMatchesSticky(Element $el): bool
 {
  try {
   if ($el->matches(self::STICKY_SELECTOR)) {
    return true;
   }
  } catch (Throwable) {
  }

  foreach ($el->getElementsByTagName('*') as $descendant) {
   try {
    if ($descendant->matches(self::STICKY_SELECTOR)) {
     return true;
    }
   } catch (Throwable) {
   }
  }

  return false;
 }

 private static function findStickyElement(Element $el): ?Element
 {
  try {
   if ($el->matches(self::STICKY_SELECTOR)) {
    return $el;
   }
  } catch (Throwable) {
  }

  foreach ($el->getElementsByTagName('*') as $descendant) {
   try {
    if ($descendant->matches(self::STICKY_SELECTOR)) {
     return $descendant;
    }
   } catch (Throwable) {
   }
  }

  return null;
 }

 /**
  * Composites scrolled viewport view with CSS position: sticky behavior.
  *
  * @param list<StickyRegion> $regions
  */
 public static function pinSticky(string $view, array $regions, int $yOffset): string
 {
  if (empty($regions)) {
   return $view;
  }

  $active = -1;
  foreach ($regions as $i => $r) {
   if ($r->startLine < $yOffset) {
    $active = $i;
   } else {
    break;
   }
  }

  if ($active < 0) {
   return $view;
  }

  $hdr = explode("\n", $regions[$active]->header);
  $n = count($hdr);

  $delta = 0;
  if (isset($regions[$active + 1])) {
   $gap = $regions[$active + 1]->startLine - $yOffset;
   if ($gap < $n) {
    $delta = $n - $gap;
   }
  }
  if ($delta >= $n) {
   return $view;
  }

  $lines = explode("\n", $view);
  for ($i = 0; $i < $n - $delta; $i++) {
   if (isset($lines[$i])) {
    $lines[$i] = $hdr[$i + $delta];
   }
  }

  return implode("\n", $lines);
 }

 /**
  * @return list<string>
  */
 public static function availableAsciiFonts(): array
 {
  return Figlet::availableAsciiFonts();
 }
}
