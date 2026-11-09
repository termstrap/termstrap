<?php

declare(strict_types=1);

namespace Termstrap\Renderer;

use Termstrap\Layout\RenderNode;
use Termstrap\Util;

final class TableRenderer
{
 /**
  * @param callable(RenderNode): string $renderNodeCallback
  */
 public static function renderTable(
  RenderNode $node,
  callable $renderNodeCallback,
 ): string {
  $trs = [];
  self::collectRows($node, $trs);

  $headers = [];
  $rows = [];
  $rowBgs = [];
  $headerSet = false;

  foreach ($trs as $tr) {
   [$cells, $allHeader] = self::rowCells($tr);
   if (empty($cells)) {
    continue;
   }

   $vals = [];
   foreach ($cells as $cell) {
    $cellNode = clone $cell;
    $cellNode->tag = 'div';
    $cellNode->style->paddingTop = 0;
    $cellNode->style->paddingBottom = 0;
    $cellNode->style->paddingLeft = 0;
    $cellNode->style->paddingRight = 0;
    $cellNode->style->marginTop = 0;
    $cellNode->style->marginBottom = 0;
    $vals[] = rtrim($renderNodeCallback($cellNode), "\n");
   }

   if ($allHeader && !$headerSet) {
    $headers = $vals;
    $headerSet = true;
    continue;
   }

   $rows[] = $vals;
   $rowBgs[] = $tr->style->bgColor;
  }

  $colCount = count($headers);
  foreach ($rows as $row) {
   if (count($row) > $colCount) {
    $colCount = count($row);
   }
  }
  if ($colCount === 0) {
   return '';
  }

  $tableW = $node->lipglossWidth > 0 ? $node->lipglossWidth : 80;
  $pattern = BorderPattern::forStyle($node->style->borderStyle, $node->style->rounded);

  $colWidths = self::computeColumnWidths($headers, $rows, $tableW, $colCount);
  $totalInnerW = array_sum($colWidths);

  $lines = [];

  // Border color: in Go, only set if node.Style.BorderColor != ""
  $borderColor = $node->style->borderColor;
  $styleBorder = function (string $text) use ($borderColor): string {
   if ($borderColor === '') {
    return $text;
   }
   $fg = Color::fgEscape($borderColor);
   return $fg !== '' ? $fg . $text . "\x1b[m" : $text;
  };

  $borderLeft = $styleBorder($pattern->left);
  $borderRight = $styleBorder($pattern->right) . "\x1b[m";

  // 1. Top border
  $topLine = $pattern->topLeft . str_repeat($pattern->top, $totalInnerW) . $pattern->topRight;
  $lines[] = $styleBorder($topLine);

  // 2. Header row
  if ($headerSet) {
   $headerCells = [];
   for ($c = 0; $c < $colCount; $c++) {
    $rawH = $headers[$c] ?? '';
    $cleanH = trim(Util::stripAnsi($rawH));
    $colW = $colWidths[$c];
    $innerColW = max(1, $colW - 2);

    $textW = Util::visualWidth($cleanH);
    $headerFg = $node->style->fgColor !== '' ? Color::fgEscape($node->style->fgColor) : '';

    if ($innerColW > $textW + 1) {
     $midSpaces = str_repeat(' ', $innerColW - $textW - 1);
     $headerContent = "\x1b[1m" . $headerFg . $cleanH . "\x1b[m" . $midSpaces . "…";
    } elseif ($innerColW > $textW) {
     $headerContent = "\x1b[1m" . $headerFg . $cleanH . "\x1b[m" . str_repeat(' ', $innerColW - $textW);
    } else {
     $headerContent = "\x1b[1m" . $headerFg . mb_strimwidth($cleanH, 0, $innerColW, '…', 'UTF-8') . "\x1b[m";
    }

    $styledH = ' ' . $headerContent . ' ';
    $headerCells[] = $styledH;
   }

   $joinedHeader = AnsiStyle::joinHorizontal('top', ...$headerCells);
   foreach (explode("\n", $joinedHeader) as $hl) {
    $lines[] = $borderLeft . $hl . $borderRight;
   }

   // Header separator
   $sepLine = $pattern->middleLeft . str_repeat($pattern->top, $totalInnerW) . $pattern->middleRight;
   $lines[] = $styleBorder($sepLine);
  }

  // 3. Body rows
  foreach ($rows as $rIdx => $row) {
   $bg = $rowBgs[$rIdx] ?? '';
   $bgEscape = $bg !== '' ? Color::bgEscape($bg) : '';

   $rowCells = [];
   for ($c = 0; $c < $colCount; $c++) {
    $cellRaw = $row[$c] ?? '';
    $cellText = (string) preg_replace('/ +(\x1b\[[0-9;]*m)*$/', '$1', rtrim($cellRaw, " \t\r\n"));
    $colW = $colWidths[$c];
    $innerColW = max(1, $colW - 2);

    $cellLines = explode("\n", TextWrapper::wrapText($cellText, $innerColW));
    if (count($cellLines) > 1 && $cellLines[array_key_last($cellLines)] === '') {
     array_pop($cellLines);
    }
    $paddedCellLines = [];
    foreach ($cellLines as $cl) {
     $curW = Util::visualWidth($cl);
     $spacePadding = max(0, $innerColW - $curW);
     if (!str_ends_with($cl, "\x1b[m") && !str_ends_with($cl, "\x1b[0m")) {
      $cl .= "\x1b[m";
     }
     $lineStr = ' ' . $cl . str_repeat(' ', $spacePadding) . ' ';
     if ($bg !== '') {
      $lineStr = Background::persistBackgroundColor($bgEscape . $lineStr . "\x1b[m", $bg);
     }
     $paddedCellLines[] = $lineStr;
    }
    $rowCells[] = implode("\n", $paddedCellLines);
   }

   $joinedRow = AnsiStyle::joinHorizontal('top', ...$rowCells);
   foreach (explode("\n", $joinedRow) as $rl) {
    $lines[] = $borderLeft . $rl . $borderRight;
   }
  }

  // 4. Bottom border
  $bottomLine = $pattern->bottomLeft . str_repeat($pattern->bottom, $totalInnerW) . $pattern->bottomRight;
  $lines[] = $styleBorder($bottomLine);

  return implode("\n", $lines);
 }

 /**
  * @param list<string> $headers
  * @param list<list<string>> $rows
  * @return list<int>
  */
 public static function computeColumnWidths(array $headers, array $rows, int $tableW, int $colCount): array
 {
  $innerBudget = max(1, $tableW - 2);

  $maxColWidths = [];
  for ($c = 0; $c < $colCount; $c++) {
   $maxW = Util::visualWidth($headers[$c] ?? '');
   foreach ($rows as $r) {
    $w = Util::visualWidth($r[$c] ?? '');
    if ($w > $maxW) {
     $maxW = $w;
    }
   }
   $maxColWidths[$c] = $maxW + 2; // include horizontal padding 1 on each side
  }

  $totalMax = array_sum($maxColWidths);
  $colWidths = $maxColWidths;

  if ($totalMax <= $innerBudget) {
   // Expand starting from shorter columns
   while (array_sum($colWidths) < $innerBudget) {
    $minIdx = 0;
    $minVal = PHP_INT_MAX;
    for ($j = 0; $j < $colCount; $j++) {
     if ($colWidths[$j] < $minVal) {
      $minVal = $colWidths[$j];
      $minIdx = $j;
     }
    }
    $colWidths[$minIdx]++;
   }
  } else {
   // Shrink starting from widest columns
   while (array_sum($colWidths) > $innerBudget) {
    $maxIdx = -1;
    $maxVal = -1;
    for ($j = 0; $j < $colCount; $j++) {
     if ($colWidths[$j] > 3 && $colWidths[$j] > $maxVal) {
      $maxVal = $colWidths[$j];
      $maxIdx = $j;
     }
    }
    if ($maxIdx < 0) {
     break;
    }
    $colWidths[$maxIdx]--;
   }
  }

  return $colWidths;
 }

 private static function truncateCell(string $text, int $maxLen): string
 {
  $curW = Util::visualWidth($text);
  if ($curW <= $maxLen) {
   return $text;
  }
  return mb_strimwidth($text, 0, $maxLen, '…', 'UTF-8');
 }

 /**
  * @param list<RenderNode> $out
  */
 public static function collectRows(RenderNode $node, array &$out): void
 {
  foreach ($node->children as $c) {
   match ($c->tag) {
    'tr' => $out[] = $c,
    'thead', 'tbody', 'tfoot' => self::collectRows($c, $out),
    default => null,
   };
  }
 }

 /**
  * @return array{0: list<RenderNode>, 1: bool} (cells, allHeader)
  */
 public static function rowCells(RenderNode $tr): array
 {
  $cells = [];
  $allHeader = true;

  foreach ($tr->children as $c) {
   if ($c->tag === 'th' || $c->tag === 'td') {
    $cells[] = $c;
    if ($c->tag !== 'th') {
     $allHeader = false;
    }
   }
  }

  if (empty($cells)) {
   $allHeader = false;
  }

  return [$cells, $allHeader];
 }
}
