<?php

declare(strict_types=1);

namespace Termstrap\Renderer;

use Dom\HTMLDocument;
use Termstrap\Css\ComputedStyle;
use Termstrap\Css\CssMatcher;
use Termstrap\Layout\LayoutEngine;
use Termstrap\Layout\NodeType;
use Termstrap\Layout\RenderNode;
use Termstrap\Markdown;
use Termstrap\Util;
use Throwable;

final class DocumentRenderer
{
 /**
  * @param array<string, mixed> $options
  * @param list<array{row: int, col: int, width: int, height: int, escape: string}> $overlays
  * @return array{0: string, 1: list<array{row: int, col: int, width: int, height: int, escape: string}>}
  */
 public static function render(string $content, array $options = [], array &$overlays = []): array
 {
  $rawContent = trim($content);
  if ($rawContent === '') {
   return ['', []];
  }

  $width = (int) ($options['width'] ?? 80);
  if ($width <= 0) {
   $width = 80;
  }

  $theme = (string) ($options['theme'] ?? 'bootstrap');
  $rootPath = isset($options['root_path']) ? (string) $options['root_path'] : null;
  $isMarkdown = (bool) ($options['markdown'] ?? false);
  $linkMode = (string) ($options['link_mode'] ?? Links::MODE_OSC8);
  $disableHyperlinks = (bool) ($options['disable_hyperlinks'] ?? false);
  $disableSyntaxHighlight = (bool) ($options['disable_syntax_highlight'] ?? false);
  $step = (int) ($options['step'] ?? 0);

  if ($isMarkdown) {
   $content = Markdown::toHtml($content);
  }
  $content = Util::escapeInlineCodeHtml($content);

  try {
   $doc = HTMLDocument::createFromString($content, LIBXML_NOERROR);
  } catch (Throwable) {
   return [$content, []];
  }

  $stylesheets = [];
  if (isset($options['stylesheets']) && is_array($options['stylesheets'])) {
   foreach ($options['stylesheets'] as $s) {
    if (is_string($s) && $s !== '') {
     $stylesheets[] = self::resolveStylesheetArg($s, $rootPath);
    }
   }
  }

  // Extract embedded <style> tags
  $styles = $doc->getElementsByTagName('style');
  foreach ($styles as $st) {
   $txt = trim($st->textContent);
   if ($txt !== '') {
    $stylesheets[] = $txt;
   }
  }

  // Extract <link rel="stylesheet"> tags
  $linksElements = $doc->getElementsByTagName('link');
  foreach ($linksElements as $lnk) {
   $rel = strtolower(trim($lnk->getAttribute('rel') ?? ''));
   if ($rel === 'stylesheet') {
    $href = trim($lnk->getAttribute('href') ?? '');
    if ($href !== '') {
     $css = self::loadExternalCss($href, $rootPath);
     if ($css !== '') {
      $stylesheets[] = $css;
     }
    }
   }
  }

  $matcher = CssMatcher::create($theme, $stylesheets, $rootPath);

  // Heading marker color from preprocessor
  $headingMarkerColor = '#495057';
  $heading1Marker = '#';
  $heading6MarkerColor = '';
  if ($matcher->preprocessor !== null) {
   if (isset($matcher->preprocessor->globalVars['--ts-heading-rule'])) {
    $headingMarkerColor = trim($matcher->preprocessor->globalVars['--ts-heading-rule']);
   }
   if (isset($matcher->preprocessor->globalVars['--ts-heading-1-marker'])) {
    $heading1Marker = trim($matcher->preprocessor->globalVars['--ts-heading-1-marker']);
   }
   if (isset($matcher->preprocessor->globalVars['--ts-heading-6'])) {
    $heading6MarkerColor = trim($matcher->preprocessor->globalVars['--ts-heading-6']);
   }
  }

  $bodies = $doc->getElementsByTagName('body');
  $rootEl = $bodies->length > 0 ? $bodies->item(0) : $doc->documentElement;
  if ($rootEl === null) {
   return ['', []];
  }

  $tree = LayoutEngine::buildRenderTree($rootEl, $matcher, $width);
  if ($tree === null) {
   return ['', []];
  }

  LayoutEngine::computeLayout($tree, $width);

  $collectedLinks = [];
  $ctx = [
   'width' => $width,
   'matcher' => $matcher,
   'linkMode' => $linkMode,
   'disableHyperlinks' => $disableHyperlinks,
   'disableSyntaxHighlight' => $disableSyntaxHighlight,
   'step' => $step,
   'headingMarkerColor' => $headingMarkerColor,
   'heading1Marker' => $heading1Marker,
   'heading6MarkerColor' => $heading6MarkerColor,
   'options' => $options,
  ];

  $output = self::renderNode($tree, 0, 0, $overlays, 0, $collectedLinks, $ctx);

  // Expand links
  $output = Links::expandLinks($output, $collectedLinks, $linkMode, $disableHyperlinks);

  return [$output, $overlays];
 }

 private static function resolveStylesheetArg(string $src, ?string $rootPath): string
 {
  $s = trim($src);
  if ($s === '') {
   return '';
  }
  $isUrl = str_starts_with($s, 'http://') || str_starts_with($s, 'https://');
  if (!$isUrl && (str_contains($s, '{') || str_contains($s, '}') || str_contains($s, "\n"))) {
   return $src;
  }
  return self::loadExternalCss($s, $rootPath);
 }

 private static function loadExternalCss(string $src, ?string $rootPath): string
 {
  if (str_starts_with($src, 'http://') || str_starts_with($src, 'https://')) {
   $ctx = stream_context_create(['http' => ['timeout' => 5]]);
   $data = @file_get_contents($src, false, $ctx);
   return $data !== false ? (string) $data : '';
  }
  $path = $src;
  if ($rootPath !== null && $rootPath !== '' && !str_starts_with($path, '/')) {
   $path = rtrim($rootPath, '/') . '/' . $path;
  }
  if (file_exists($path)) {
   $data = @file_get_contents($path);
   return $data !== false ? (string) $data : '';
  }
  return '';
 }

 /**
  * @param list<array{row: int, col: int, width: int, height: int, escape: string}> $overlays
  * @param list<string> $links
  * @param array<string, mixed> $ctx
  */
 private static function renderNode(
  RenderNode $node,
  int $currentX,
  int $currentY,
  array &$overlays,
  int $targetHeight,
  array &$links,
  array $ctx,
 ): string {
  if ($node->tag === 'hr') {
   $w = $node->contentWidth > 0 ? $node->contentWidth : 80;
   $rule = str_repeat('─', $w);
   if ($node->style->borderColor !== '') {
    $rule = Color::fgEscape($node->style->borderColor) . $rule . "\x1b[m";
   }
   return AnsiStyle::applyLipglossStyle($node->style, $node->lipglossWidth, $targetHeight, $rule, $ctx['step']);
  }

  if ($node->tag === 'pre') {
   $text = $node->text;
   if (!$ctx['disableSyntaxHighlight']) {
    $vars = $ctx['matcher']->preprocessor !== null ? $ctx['matcher']->preprocessor->globalVars : [];
    $text = CodeHighlighter::highlight($text, $node->lang, $vars);
   }
   return AnsiStyle::applyLipglossStyle($node->style, $node->lipglossWidth, $targetHeight, $text, $ctx['step']);
  }
  if ($node->tag === 'ascii') {
   return self::renderAsciiNode($node, $currentX, $currentY, $targetHeight, $ctx);
  }
  if ($node->type === NodeType::Text) {
   return $node->text;
  }

  if ($node->type === NodeType::Image) {
   return self::renderImageNode($node, $currentX, $currentY, $overlays, $links, false, $ctx);
  }

  if ($node->tag === 'ul' || $node->tag === 'ol') {
   $renderSub = fn(RenderNode $n) => self::renderNode($n, $currentX, $currentY, $overlays, 0, $links, $ctx);
   $content = ListRenderer::renderList($node, $renderSub);
   return AnsiStyle::applyLipglossStyle($node->style, $node->lipglossWidth, $targetHeight, $content, $ctx['step']);
  }

  if ($node->tag === 'table') {
   $renderSub = fn(RenderNode $n) => self::renderNode($n, $currentX, $currentY, $overlays, 0, $links, $ctx);
   $content = TableRenderer::renderTable($node, $renderSub);
   $st = clone $node->style;
   $st->border = false;
   $st->borderTop = false;
   $st->borderBottom = false;
   $st->borderLeft = false;
   $st->borderRight = false;
   return AnsiStyle::applyLipglossStyle($st, 0, $targetHeight, $content, $ctx['step']);
  }

  $content = '';
  if ($node->style->display === 'flex') {
   /** @var list<array{children: list<RenderNode>, spans: list<int>, spanSum: int}> $subRows */
   $subRows = [];
   $currentSubRow = ['children' => [], 'spans' => [], 'spanSum' => 0];

   foreach ($node->children as $child) {
    $span = $child->style->colSpan > 0 ? $child->style->colSpan : 12;
    if (!empty($currentSubRow['children']) && $currentSubRow['spanSum'] + $span > 12) {
     $subRows[] = $currentSubRow;
     $currentSubRow = ['children' => [], 'spans' => [], 'spanSum' => 0];
    }
    $currentSubRow['children'][] = $child;
    $currentSubRow['spans'][] = $span;
    $currentSubRow['spanSum'] += $span;
   }
   if (!empty($currentSubRow['children'])) {
    $subRows[] = $currentSubRow;
   }

   $rowOutputs = [];
   foreach ($subRows as $row) {
    $renderedItems = [];
    $maxH = 0;
    foreach ($row['children'] as $child) {
     $out = self::renderNode($child, $currentX, $currentY, $overlays, 0, $links, $ctx);
     $renderedItems[] = $out;
     $h = count(explode("\n", $out));
     if ($h > $maxH) {
      $maxH = $h;
     }
    }

    $crossPos = 'top';
    $stretch = true;
    if ($node->style->alignItemsSet && $node->style->alignItems !== 'top') {
     $crossPos = $node->style->alignItems;
     $stretch = false;
    }

    $itemOutputs = [];
    if ($stretch && count($row['children']) > 1 && $maxH > 0) {
     foreach ($row['children'] as $child) {
      $out = self::renderNode($child, $currentX, $currentY, $overlays, $maxH, $links, $ctx);
      if ($out !== '') {
       $itemOutputs[] = $out;
      }
     }
    } else {
     $itemOutputs = $renderedItems;
    }

    $rowOutputs[] = AnsiStyle::joinHorizontal($crossPos, ...$itemOutputs);
   }

   $content = AnsiStyle::joinVertical('left', ...$rowOutputs);
  } else {
   // Block
   $blockX = $currentX + $node->style->marginLeft + $node->style->paddingLeft;
   $blockY = $currentY + $node->style->marginTop + $node->style->paddingTop;

   $childOutputs = [];
   $inlineRun = [];

   $flushInlineRun = function () use (&$inlineRun, &$childOutputs, &$blockY, $blockX, &$overlays, &$links, $node, $ctx): void {
    if (empty($inlineRun)) {
     return;
    }
    $sb = '';
    foreach ($inlineRun as $inlineNode) {
     $sb .= self::renderInline($inlineNode, $blockX, $blockY, $overlays, $links, false, $ctx);
    }
    $inlineRun = [];

    $rawText = $sb;
    $trimmed = trim($rawText);
    if ($trimmed === '') {
     return;
    }

    $glued = Util::gluePunctuation($trimmed);
    $wrapped = TextWrapper::wrapText($glued, $node->contentWidth);
    $wrapped = str_replace("\u{00a0}", ' ', $wrapped);

    $lvl = HeadingDecorator::headingLevel($node->tag);
    if ($lvl > 0) {
     $hColor = $ctx['headingMarkerColor'];
     if ($lvl === 6 && $ctx['heading6MarkerColor'] !== '') {
      $hColor = $ctx['heading6MarkerColor'];
     }
     $wrapped = HeadingDecorator::decorate($wrapped, $lvl, $node->style->textAlign, $hColor, $ctx['heading1Marker']);
    }

    if ($node->tag === 'li') {
     $liLines = explode("\n", $wrapped);
     foreach ($liLines as $i => $l) {
      $liLines[$i] = $i === 0 ? '• ' . $l : '  ' . $l;
     }
     $wrapped = implode("\n", $liLines);
    }

    if ($node->style->textAlign === 'center' || $node->style->textAlign === 'right') {
     $wLines = explode("\n", $wrapped);
     foreach ($wLines as $wi => $wl) {
      $curW = Util::visualWidth($wl);
      if ($curW < $node->contentWidth) {
       $diff = $node->contentWidth - $curW;
       $wLines[$wi] = match ($node->style->textAlign) {
        'center' => str_repeat(' ', (int) ($diff / 2)) . $wl . str_repeat(' ', $diff - (int) ($diff / 2)),
        'right' => str_repeat(' ', $diff) . $wl,
        default => $wl,
       };
      }
     }
     $wrapped = implode("\n", $wLines);
    }

    $childOutputs[] = $wrapped;
    $blockY += substr_count($wrapped, "\n") + 1;
   };

   foreach ($node->children as $child) {
    if ($child->style->display === 'inline' || $child->type === NodeType::Text) {
     $inlineRun[] = $child;
    } else {
     $flushInlineRun();
     $out = self::renderNode($child, $blockX, $blockY, $overlays, 0, $links, $ctx);
     if ($out !== '') {
      if ($node->style->textAlign === 'center' && $node->contentWidth > $child->boxWidth) {
       $cLines = explode("\n", $out);
       foreach ($cLines as $ci => $cl) {
        $curW = Util::visualWidth($cl);
        if ($curW < $node->contentWidth) {
         $diff = $node->contentWidth - $curW;
         $cLines[$ci] = str_repeat(' ', (int) ($diff / 2)) . $cl . str_repeat(' ', $diff - (int) ($diff / 2));
        }
       }
       $out = implode("\n", $cLines);
      } elseif ($node->style->textAlign === 'right' && $node->contentWidth > $child->boxWidth) {
       $cLines = explode("\n", $out);
       foreach ($cLines as $ci => $cl) {
        $curW = Util::visualWidth($cl);
        if ($curW < $node->contentWidth) {
         $diff = $node->contentWidth - $curW;
         $cLines[$ci] = str_repeat(' ', $diff) . $cl;
        }
       }
       $out = implode("\n", $cLines);
      }
      $childOutputs[] = $out;
      $blockY += substr_count($out, "\n") + 1;
     }
    }
   }
   $flushInlineRun();
   $content = AnsiStyle::joinVertical('left', ...$childOutputs);
  }

  $out = AnsiStyle::applyLipglossStyle($node->style, $node->lipglossWidth, $targetHeight, $content, $ctx['step']);
  if ($node->tag === 'a' && $node->href !== '') {
   $out = Links::decorateLink($out, $node->href, $links, $ctx['linkMode'], $ctx['disableHyperlinks']);
  }

  return $out;
 }

 /**
  * @param list<array{row: int, col: int, width: int, height: int, escape: string}> $overlays
  * @param list<string> $links
  * @param array<string, mixed> $ctx
  */
 private static function renderInline(
  RenderNode $node,
  int $currentX,
  int $currentY,
  array &$overlays,
  array &$links,
  bool $inLink,
  array $ctx,
 ): string {
  if ($node->tag === 'spinner' || $node->style->spinner) {
   return \Termstrap\Animation\Spinner::spinner($ctx['step'] ?? 0);
  }

  if ($node->tag === 'ascii') {
   return self::renderAsciiNode($node, $currentX, $currentY, 0, $ctx);
  }

  if ($node->type === NodeType::Text) {
   if ($node->style->shimmer) {
    $base = $node->style->fgColor !== '' ? $node->style->fgColor : '#44475a';
    $high = $node->style->shimmerHighlight !== '' ? $node->style->shimmerHighlight : '#ffffff';
    return \Termstrap\Animation\Shimmer::shimmer($node->text, $ctx['step'] ?? 0, $base, $high, $node->style->bold);
   }
   return AnsiStyle::applyInlineStyle($node->style, $node->text);
  }

  if ($node->tag === 'br') {
   return "\n";
  }

  if ($node->type === NodeType::Image) {
   return self::renderImageNode($node, $currentX, $currentY, $overlays, $links, $inLink, $ctx);
  }

  $linkActive = $inLink || ($node->tag === 'a' && $node->href !== '');

  $sb = '';
  foreach ($node->children as $child) {
   $sb .= self::renderInline($child, $currentX, $currentY, $overlays, $links, $linkActive, $ctx);
  }
  $content = $sb;

  if ($node->style->paddingLeft > 0 || $node->style->paddingRight > 0 || $node->style->bgColor !== '') {
   $bg = $node->style->bgColor !== '' ? Color::bgEscape($node->style->bgColor) : '';
   $leftPad = $node->style->paddingLeft > 0 ? ($bg . str_repeat(' ', $node->style->paddingLeft) . "\x1b[m") : '';
   $rightPad = $node->style->paddingRight > 0 ? ($bg . str_repeat(' ', $node->style->paddingRight) . "\x1b[m") : '';
   if ($bg !== '') {
    $content = $bg . $content . "\x1b[m";
   }
   $content = $leftPad . $content . $rightPad;
  }

  if ($node->tag === 'a' && $node->href !== '') {
   $content = Links::decorateLink($content, $node->href, $links, $ctx['linkMode'], $ctx['disableHyperlinks']);
  }

  return $content;
 }

 /**
  * Placeholder image node renderer (graphics protocols covered in Step 5)
  */
 private static function renderGlamourImage(
  RenderNode $node,
  array &$links,
  bool $inLink,
  array $ctx,
 ): string {
  $alt = $node->alt !== '' ? $node->alt : 'Image';
  if ($inLink) {
   return '[' . $alt . ']';
  }

  $prefix = Color::fgEscape('#767676') . 'Image: ' . $alt . " →\x1b[m";
  $urlStyle = "\x1b[38;2;255;135;215m\x1b[4m";

  $displayURL = $node->src;
  $termWidth = $ctx['width'] ?? 80;
  $truncate = (bool) ($ctx['options']['truncate_urls'] ?? false);

  if ($truncate && $termWidth > 0) {
   $prefixW = Util::visualWidth($prefix);
   $avail = $termWidth - $prefixW - 2;
   if ($avail >= 20 && strlen($displayURL) > $avail) {
    $displayURL = self::truncateUrl($displayURL, $avail);
   }
  }

  $urlText = $urlStyle . $displayURL . "\x1b[m";
  if (!$ctx['disableHyperlinks'] && $node->src !== '') {
   $id = count($links);
   $links[] = $node->src;
   [$startMarker, $endMarker] = Links::linkPlaceholders($id);
   $urlText = $startMarker . $urlText . $endMarker;
  }

  $prefixW = Util::visualWidth($prefix);
  if ($termWidth > 0 && $prefixW + 1 + strlen($displayURL) > $termWidth) {
   return $prefix . "\n  " . $urlText;
  }
  return $prefix . ' ' . $urlText;
 }

 private static function truncateUrl(string $u, int $maxLen): string
 {
  if (strlen($u) <= $maxLen || $maxLen < 16) {
   return $u;
  }
  $part1 = (int) ($maxLen / 2) - 2;
  $part2 = $maxLen - $part1 - 3;
  return substr($u, 0, $part1) . '...' . substr($u, -$part2);
 }

 private static function renderImageNode(
  RenderNode $node,
  int $currentX,
  int $currentY,
  array &$overlays,
  array &$links,
  bool $inLink,
  array $ctx,
 ): string {
  $alt = $node->alt !== '' ? $node->alt : 'Image';
  if ($inLink) {
   return AnsiStyle::applyInlineStyle($node->style, '[' . $alt . ']');
  }

  $imageMode = (string) ($ctx['options']['image_mode'] ?? 'auto');
  $disableImages = (bool) ($ctx['options']['disable_images'] ?? false);
  $disableAnimated = (bool) ($ctx['options']['disable_animated_images'] ?? false);

  if ($imageMode === 'glamour') {
   return self::renderGlamourImage($node, $links, $inLink, $ctx);
  }

  if ($disableImages || $imageMode === 'placeholder') {
   return '[' . $alt . '](' . $node->src . ')';
  }

  $imgWidth = $node->contentWidth;
  $maxImageWidth = (int) ($ctx['options']['max_image_width'] ?? 0);
  if ($maxImageWidth > 0 && $imgWidth > $maxImageWidth) {
   $imgWidth = $maxImageWidth;
  }
  if ($imgWidth < 1) {
   $imgWidth = 1;
  }

  if ($imageMode === 'smart') {
   $minW = (int) ($ctx['options']['min_image_width'] ?? 12);
   if ($minW <= 0) {
    $minW = 12;
   }
   if ($imgWidth < $minW) {
    return self::renderGlamourImage($node, $links, $inLink, $ctx);
   }
   $srcLower = strtolower($node->src);
   if (str_ends_with($srcLower, '.svg') || str_contains($srcLower, '.svg?')) {
    return self::renderGlamourImage($node, $links, $inLink, $ctx);
   }
  }

  $rootPath = $ctx['options']['root_path'] ?? null;
  try {
   $loaded = \Termstrap\Image\ImageLoader::load($node->src, $rootPath);
   if ($loaded['isAnimated'] && $disableAnimated) {
    return self::renderGlamourImage($node, $links, $inLink, $ctx);
   }
   $img = $loaded['image'];
  } catch (Throwable) {
   if ($imageMode === 'smart') {
    return self::renderGlamourImage($node, $links, $inLink, $ctx);
   }
   return '[' . $alt . ']';
  }

  if ($imageMode === 'smart') {
   $w = imagesx($img);
   $h = imagesy($img);
   if ($w < 32 && $h < 32) {
    return self::renderGlamourImage($node, $links, $inLink, $ctx);
   }
   if ($h > 0 && $w > 0) {
    $ratio = (float) $w / (float) $h;
    if ($ratio > 5.0 || $ratio < 0.25) {
     return self::renderGlamourImage($node, $links, $inLink, $ctx);
    }
   }
  }

  $protocol = \Termstrap\Image\Detect::detectProtocol();
  $colorMode = isset($ctx['options']['color_mode'])
   ? \Termstrap\Image\ColorMode::parse((string) $ctx['options']['color_mode'])
   : \Termstrap\Image\ColorMode::TrueColor;

  $rendered = match ($protocol) {
   \Termstrap\Image\Protocol::Kitty => \Termstrap\Image\KittyRenderer::render($img, $imgWidth),
   \Termstrap\Image\Protocol::ITerm2 => \Termstrap\Image\ITermRenderer::render($img, $imgWidth),
   \Termstrap\Image\Protocol::Sixel => \Termstrap\Image\SixelRenderer::render($img, $imgWidth),
   default => \Termstrap\Image\HalfBlockRenderer::render($img, $imgWidth, $colorMode),
  };

  $lines = explode("\n", $rendered);
  $imgHeight = count($lines);

  if ($protocol->isGraphic()) {
   $overlays[] = [
    'row' => $currentY,
    'col' => $currentX,
    'width' => $imgWidth,
    'height' => $imgHeight,
    'escape' => $rendered,
   ];
   $blankLine = str_repeat(' ', $imgWidth);
   return implode("\n", array_fill(0, $imgHeight, $blankLine));
  }

  return $rendered;
 }

 private static function renderAsciiNode(
  RenderNode $node,
  int $currentX,
  int $currentY,
  int $targetHeight,
  array $ctx,
 ): string {
  $targetW = $node->contentWidth > 0 ? $node->contentWidth : $node->lipglossWidth;
  $lines = \Termstrap\Ascii\Figlet::render(
   $node->text,
   $node->font !== '' ? $node->font : 'standard',
   $node->fit !== '' ? $node->fit : 'clip',
   $node->strict,
   $targetW
  );

  $maxBannerW = 0;
  foreach ($lines as $line) {
   $w = Util::visualWidth($line);
   if ($w > $maxBannerW) {
    $maxBannerW = $w;
   }
  }

  $formattedLines = [];
  $step = $ctx['step'] ?? 0;
  foreach ($lines as $line) {
   $lw = Util::visualWidth($line);
   if ($maxBannerW > 0 && $lw < $maxBannerW) {
    $line .= str_repeat(' ', $maxBannerW - $lw);
   }
   if ($node->style->shimmer) {
    $base = $node->style->fgColor !== '' ? $node->style->fgColor : '#44475a';
    $high = $node->style->shimmerHighlight !== '' ? $node->style->shimmerHighlight : '#ffffff';
    $line = \Termstrap\Animation\Shimmer::shimmer($line, $step, $base, $high, $node->style->bold);
   }
   if ($targetW > 0) {
    $lw = Util::visualWidth($line);
    if ($lw < $targetW) {
     $diff = $targetW - $lw;
     $line = match ($node->style->textAlign) {
      'center' => str_repeat(' ', (int) ($diff / 2)) . $line . str_repeat(' ', $diff - (int) ($diff / 2)),
      'right' => str_repeat(' ', $diff) . $line,
      default => $line,
     };
    }
   }
   $formattedLines[] = $line;
  }

  $asciiText = implode("\n", $formattedLines);
  if ($node->style->display === 'inline') {
   if ($node->style->shimmer) {
    return $asciiText;
   }
   return AnsiStyle::applyInlineStyle($node->style, $asciiText);
  }

  $styleCopy = clone $node->style;
  if ($styleCopy->shimmer) {
   $styleCopy->fgColor = '';
  }
  return AnsiStyle::applyLipglossStyle($styleCopy, $node->lipglossWidth, $targetHeight, $asciiText, $step);
 }
}
