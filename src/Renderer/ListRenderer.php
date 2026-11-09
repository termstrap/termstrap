<?php

declare(strict_types=1);

namespace Termstrap\Renderer;

use Termstrap\Layout\RenderNode;

final class ListRenderer
{
 /**
  * @param callable(RenderNode): string $renderNodeCallback
  */
 public static function renderList(RenderNode $node, callable $renderNodeCallback): string
 {
  if ($node->isTree) {
   return self::renderTree($node, $renderNodeCallback, '');
  }

  $enumFn = self::getEnumerator($node);
  $lines = [];
  $itemIndex = 1;

  foreach ($node->children as $child) {
   if ($child->tag !== 'li') {
    continue;
   }

   [$contentNodes, $nestedLists] = self::splitListItem($child);

   // Render item content
   $itemNode = clone $child;
   $itemNode->tag = 'div';
   $itemNode->children = $contentNodes;
   $renderedItem = rtrim($renderNodeCallback($itemNode), "\n");

   $prefix = $enumFn($itemIndex);
   $prefixLen = mb_strwidth($prefix, 'UTF-8');
   $indentSpaces = str_repeat(' ', $prefixLen);

   $itemLines = explode("\n", $renderedItem);
   foreach ($itemLines as $idx => $line) {
    if ($idx === 0) {
     $lines[] = $prefix . $line;
    } else {
     $lines[] = $indentSpaces . $line;
    }
   }

   // Render nested sublists indented
   foreach ($nestedLists as $sublist) {
    $subRendered = self::renderList($sublist, $renderNodeCallback);
    foreach (explode("\n", $subRendered) as $subLine) {
     if ($subLine !== '') {
      $lines[] = '  ' . $subLine;
     }
    }
   }

   $itemIndex++;
  }

  return implode("\n", $lines);
 }

 /**
  * @param callable(RenderNode): string $renderNodeCallback
  */
 public static function renderTree(RenderNode $node, callable $renderNodeCallback, string $prefix): string
 {
  $lines = [];
  $childCount = count($node->children);
  $actualLis = array_values(array_filter($node->children, fn($c) => $c->tag === 'li'));
  $total = count($actualLis);

  foreach ($actualLis as $i => $child) {
   $isLast = ($i === $total - 1);
   $branch = $isLast ? '└── ' : '├── ';
   $nextPrefix = $prefix . ($isLast ? '    ' : '│   ');

   [$contentNodes, $nestedLists] = self::splitListItem($child);

   $itemNode = clone $child;
   $itemNode->tag = 'div';
   $itemNode->children = $contentNodes;
   $renderedItem = rtrim($renderNodeCallback($itemNode), "\n");

   $itemLines = explode("\n", $renderedItem);
   foreach ($itemLines as $idx => $line) {
    if ($idx === 0) {
     $lines[] = $prefix . $branch . $line;
    } else {
     $lines[] = $nextPrefix . $line;
    }
   }

   foreach ($nestedLists as $sublist) {
    $subLines = self::renderTree($sublist, $renderNodeCallback, $nextPrefix);
    if ($subLines !== '') {
     $lines[] = $subLines;
    }
   }
  }

  return implode("\n", $lines);
 }

 /**
  * @return array{0: list<RenderNode>, 1: list<RenderNode>}
  */
 public static function splitListItem(RenderNode $li): array
 {
  $content = [];
  $nested = [];
  foreach ($li->children as $c) {
   if ($c->tag === 'ul' || $c->tag === 'ol') {
    $nested[] = $c;
   } else {
    $content[] = $c;
   }
  }
  return [$content, $nested];
 }

 /**
  * @return callable(int): string
  */
 public static function getEnumerator(RenderNode $node): callable
 {
  $fg = $node->style->fgColor !== '' ? Color::fgEscape($node->style->fgColor) : '';
  $wrap = fn(string $s) => $fg !== '' ? $fg . $s . "\x1b[m" : $s;

  $style = $node->style->listStyle;
  if ($style === 'none') {
   return fn(int $i) => $wrap(' ');
  }

  $liCount = count(array_filter($node->children, fn($c) => $c->tag === 'li'));
  if ($style === 'decimal' || ($style === '' && $node->tag === 'ol')) {
   $numWidth = max(1, strlen((string) $liCount));
   return fn(int $i) => $wrap(sprintf('%' . $numWidth . 'd. ', $i));
  }
  if (in_array($style, ['lower-alpha', 'lower-latin'], true)) {
   return fn(int $i) => $wrap(self::toAlphabet($i, false) . '. ');
  }
  if (in_array($style, ['upper-alpha', 'upper-latin'], true)) {
   return fn(int $i) => $wrap(self::toAlphabet($i, true) . '. ');
  }
  if ($style === 'lower-roman' || $style === 'upper-roman') {
   $upper = $style === 'upper-roman';
   $romanWidth = max(1, strlen(self::toRoman($liCount, $upper)));
   return fn(int $i) => $wrap(sprintf('%' . $romanWidth . 's. ', self::toRoman($i, $upper)));
  }
  if ($style === 'circle') {
   return fn(int $i) => $wrap('◦ ');
  }
  if ($style === 'square') {
   return fn(int $i) => $wrap('▪ ');
  }

  return fn(int $i) => $wrap('• ');
 }

 private static function toAlphabet(int $n, bool $upper): string
 {
  $n = max(1, $n);
  $s = '';
  while ($n > 0) {
   $rem = ($n - 1) % 26;
   $s = chr(($upper ? 65 : 97) + $rem) . $s;
   $n = (int) (($n - 1) / 26);
  }
  return $s;
 }

 private static function toRoman(int $num, bool $upper): string
 {
  $lookup = [
   'M' => 1000,
   'CM' => 900,
   'D' => 500,
   'CD' => 400,
   'C' => 100,
   'XC' => 90,
   'L' => 50,
   'XL' => 40,
   'X' => 10,
   'IX' => 9,
   'V' => 5,
   'IV' => 4,
   'I' => 1,
  ];
  $roman = '';
  foreach ($lookup as $sym => $val) {
   while ($num >= $val) {
    $roman .= $sym;
    $num -= $val;
   }
  }
  return $upper ? $roman : strtolower($roman);
 }
}
