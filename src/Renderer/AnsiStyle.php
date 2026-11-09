<?php

declare(strict_types=1);

namespace Termstrap\Renderer;

use Termstrap\Css\ComputedStyle;
use Termstrap\Util;

final class AnsiStyle
{
    public static function applyInlineStyle(ComputedStyle $s, string $text): string
    {
        if ($text === '') {
            return '';
        }

        $sgrParts = [];
        if ($s->bold) {
            $sgrParts[] = '1';
        }
        if ($s->faint) {
            $sgrParts[] = '2';
        }
        if ($s->italic) {
            $sgrParts[] = '3';
        }
        if ($s->underline) {
            $sgrParts[] = '4';
        }
        if ($s->blink) {
            $sgrParts[] = '5';
        }
        if ($s->strikethrough) {
            $sgrParts[] = '9';
        }

        if ($s->fgColor !== '') {
            [$r, $g, $b] = Color::parse($s->fgColor);
            $sgrParts[] = sprintf('38;2;%d;%d;%d', $r, $g, $b);
        }
        if ($s->bgColor !== '') {
            [$r, $g, $b] = Color::parse($s->bgColor);
            $sgrParts[] = sprintf('48;2;%d;%d;%d', $r, $g, $b);
        }
        if ($s->underlineColor !== '') {
            [$r, $g, $b] = Color::parse($s->underlineColor);
            $sgrParts[] = sprintf('58;2;%d;%d;%d', $r, $g, $b);
        }

        if (empty($sgrParts)) {
            return $text;
        }

        return "\x1b[" . implode(';', $sgrParts) . 'm' . $text . "\x1b[m";
    }

    public static function applyLipglossStyle(
        ComputedStyle $s,
        int $width,
        int $height,
        string $content,
        int $step = 0,
    ): string {
        $bgEscape = $s->bgColor !== '' ? Color::bgEscape($s->bgColor) : '';
        $fgEscape = $s->fgColor !== '' ? Color::fgEscape($s->fgColor) : '';

        // Box padding spaces use ONLY background color
        $padWhitespace = function (int $count) use ($bgEscape): string {
            if ($count <= 0) {
                return '';
            }
            $spaces = str_repeat(' ', $count);
            return $bgEscape !== '' ? $bgEscape . $spaces . "\x1b[m" : $spaces;
        };

        // Inner content line spaces carry foreground and background
        $padInner = function (int $count) use ($fgEscape, $bgEscape): string {
            if ($count <= 0) {
                return '';
            }
            $spaces = str_repeat(' ', $count);
            if ($bgEscape !== '') {
                return ($fgEscape !== '' ? $fgEscape : '') . $bgEscape . $spaces . "\x1b[m";
            }
            return $spaces;
        };

        $lines = explode("\n", $content);

        // 1. Render core text styling first
        $styledLines = [];
        foreach ($lines as $line) {
            $trimmed = rtrim($line, ' ');
            $trailingCount = strlen($line) - strlen($trimmed);
            $styled = self::applyInlineStyle($s, $trimmed);
            $styledLines[] = $styled . ($trailingCount > 0 ? $padWhitespace($trailingCount) : '');
        }
        $lines = $styledLines;

        // 2. Alignment & padding to targetInnerWidth
        $targetInnerWidth = $width;
        if ($targetInnerWidth > 0 && $s->display !== 'inline') {
            $borderH = ($s->border || $s->borderLeft ? 1 : 0) + ($s->border || $s->borderRight ? 1 : 0);
            $padH = $s->paddingLeft + $s->paddingRight;
            $innerWidth = max(1, $targetInnerWidth - $borderH - $padH);

            foreach ($lines as $i => $l) {
                $curW = Util::visualWidth($l);
                if ($curW === 0 && $l === '') {
                    $lines[$i] = $padInner($innerWidth);
                } elseif ($curW < $innerWidth) {
                    $diff = $innerWidth - $curW;
                    $lines[$i] = match ($s->textAlign) {
                        'center' => $padWhitespace((int) ($diff / 2)) . $l . $padWhitespace($diff - (int) ($diff / 2)),
                        'right' => $padWhitespace($diff) . $l,
                        default => $l . $padWhitespace($diff),
                    };
                }
            }
        }

        // 3. Apply box padding
        if ($s->display !== 'inline') {
            if ($s->paddingLeft > 0 || $s->paddingRight > 0) {
                $leftPad = $padWhitespace($s->paddingLeft);
                $rightPad = $padWhitespace($s->paddingRight);
                foreach ($lines as $i => $l) {
                    $lines[$i] = $leftPad . $l . $rightPad;
                }
            }

            if ($s->paddingTop > 0) {
                $topLineW = $targetInnerWidth > 0 ? max(1, $targetInnerWidth - (($s->border || $s->borderLeft ? 1 : 0) + ($s->border || $s->borderRight ? 1 : 0))) : 0;
                $topLines = array_fill(0, $s->paddingTop, $padWhitespace($topLineW));
                array_unshift($lines, ...$topLines);
            }

            if ($s->paddingBottom > 0) {
                $botLineW = $targetInnerWidth > 0 ? max(1, $targetInnerWidth - (($s->border || $s->borderLeft ? 1 : 0) + ($s->border || $s->borderRight ? 1 : 0))) : 0;
                $bottomLines = array_fill(0, $s->paddingBottom, $padWhitespace($botLineW));
                array_push($lines, ...$bottomLines);
            }
        }

        // 4. Height alignment
        if ($height > 0 && $s->display !== 'inline') {
            $borderV = ($s->border ? 2 : (($s->borderTop ? 1 : 0) + ($s->borderBottom ? 1 : 0)));
            $marginV = $s->marginTop + $s->marginBottom;
            $shadowV = Shadow::shadowBottomRows($s->shadow);
            $innerH = max(1, $height - $borderV - $marginV - $shadowV);

            $curH = count($lines);
            if (($s->bgColor !== '' || !empty($s->bgGradient)) && $curH < $innerH) {
                $diff = $innerH - $curH;
                $lineW = Util::visualWidth($lines[0] ?? '');
                $emptyLine = $padWhitespace($lineW);
                for ($k = 0; $k < $diff; $k++) {
                    $lines[] = $emptyLine;
                }
            }
        }

        $out = implode("\n", $lines);

        // Apply borders
        $hasTop = $s->border || $s->borderTop;
        $hasRight = $s->border || $s->borderRight;
        $hasBottom = $s->border || $s->borderBottom;
        $hasLeft = $s->border || $s->borderLeft;

        if ($hasTop || $hasRight || $hasBottom || $hasLeft) {
            $out = self::renderBorder($out, $s, $hasTop, $hasRight, $hasBottom, $hasLeft);
        }

        // Apply Margins
        if ($s->marginTop > 0 || $s->marginBottom > 0 || $s->marginLeft > 0 || $s->marginRight > 0) {
            $mOutLines = explode("\n", $out);
            if ($s->marginLeft > 0 || $s->marginRight > 0) {
                $leftM = str_repeat(' ', $s->marginLeft);
                $rightM = str_repeat(' ', $s->marginRight);
                foreach ($mOutLines as $i => $ml) {
                    $mOutLines[$i] = $leftM . $ml . $rightM;
                }
            }
            if ($s->marginTop > 0) {
                array_unshift($mOutLines, ...array_fill(0, $s->marginTop, ''));
            }
            if ($s->marginBottom > 0) {
                array_push($mOutLines, ...array_fill(0, $s->marginBottom, ''));
            }
            $out = implode("\n", $mOutLines);
        }

        // Background persistence or gradient
        if (!empty($s->bgGradient)) {
            $out = Background::applyBackgroundGradient($out, $s->bgGradient);
        } elseif ($s->bgColor !== '') {
            $out = Background::persistBackgroundColor($out, $s->bgColor);
        }

        // Drop shadow
        if ($s->shadow > 0) {
            $out = Shadow::applyShadowWithWidth($out, $s->shadow, 0);
        }

        return $out;
    }

    private static function renderBorder(
        string $content,
        ComputedStyle $s,
        bool $hasTop,
        bool $hasRight,
        bool $hasBottom,
        bool $hasLeft,
    ): string {
        $pattern = BorderPattern::forStyle($s->borderStyle, $s->rounded);
        $lines = explode("\n", $content);
        $contentW = max(array_map([Util::class, 'visualWidth'], $lines));

        $totalW = $contentW + ($hasLeft ? 1 : 0) + ($hasRight ? 1 : 0);

        // Resolve border colors: in Go, only set if borderColor != ""
        $defaultBorderColor = $s->borderColor;
        if (!empty($s->borderBlend)) {
            $defaultBorderColor = $s->borderBlend[0];
        }
        $topColor = $s->borderTopColor !== '' ? $s->borderTopColor : $defaultBorderColor;
        $rightColor = $s->borderRightColor !== '' ? $s->borderRightColor : (isset($s->borderBlend[1]) ? $s->borderBlend[1] : $defaultBorderColor);
        $bottomColor = $s->borderBottomColor !== '' ? $s->borderBottomColor : (isset($s->borderBlend[1]) ? $s->borderBlend[1] : $defaultBorderColor);
        $leftColor = $s->borderLeftColor !== '' ? $s->borderLeftColor : $defaultBorderColor;
        $bgEscape = $s->bgColor !== '' ? Color::bgEscape($s->bgColor) : '';

        $styleBorderPart = function (string $text, string $fgColor) use ($bgEscape): string {
            if ($text === '') {
                return '';
            }
            $fg = $fgColor !== '' ? Color::fgEscape($fgColor) : '';
            if ($fg === '' && $bgEscape === '') {
                return $text;
            }
            return $fg . $bgEscape . $text . "\x1b[m";
        };

        if (!empty($s->borderBlend) && count($s->borderBlend) >= 2) {
            $h = count($lines);
            $totalPerimeter = ($totalW + $h + 2) * 2;
            $gradient = Color::blend1D($totalPerimeter, $s->borderBlend);

            $topGrad = array_slice($gradient, 0, $totalW);
            $rightGrad = array_slice($gradient, $totalW + 2, $h);
            $bot32 = array_slice($gradient, $totalW + 2 + $h, $totalW + 2);
            $botGrad = array_slice(array_reverse($bot32), 0, $totalW);
            $leftH = array_slice($gradient, ($totalW + 2) * 2 + $h, $h);
            $leftGrad = array_reverse($leftH);

            $styleRunes = function (string $text, array $grad) use ($bgEscape): string {
                $chars = preg_split('//u', $text, -1, PREG_SPLIT_NO_EMPTY) ?: [];
                $out = '';
                foreach ($chars as $idx => $ch) {
                    $c = $grad[$idx] ?? ($grad[0] ?? [255, 255, 255]);
                    $fg = sprintf("\x1b[38;2;%d;%d;%dm", $c[0], $c[1], $c[2]);
                    $out .= $fg . $bgEscape . $ch . "\x1b[m";
                }
                return $out;
            };

            // Top
            if ($hasTop) {
                $tl = $hasLeft ? $pattern->topLeft : '';
                $tr = $hasRight ? $pattern->topRight : '';
                $midCount = max(0, $totalW - Util::visualWidth($tl) - Util::visualWidth($tr));
                $mid = str_repeat($pattern->top, $midCount);
                $result[] = $styleRunes($tl . $mid . $tr, $topGrad);
            }

            // Sides
            foreach ($lines as $i => $line) {
                $cLeft = $leftGrad[$i] ?? ($leftGrad[0] ?? [255, 255, 255]);
                $cRight = $rightGrad[$i] ?? ($rightGrad[0] ?? [255, 255, 255]);
                $leftStr = $hasLeft ? sprintf("\x1b[38;2;%d;%d;%dm", $cLeft[0], $cLeft[1], $cLeft[2]) . $bgEscape . $pattern->left . "\x1b[m" : '';
                $rightStr = $hasRight ? sprintf("\x1b[38;2;%d;%d;%dm", $cRight[0], $cRight[1], $cRight[2]) . $bgEscape . $pattern->right . "\x1b[m" : '';
                $result[] = $leftStr . $line . $rightStr;
            }

            // Bottom
            if ($hasBottom) {
                $bl = $hasLeft ? $pattern->bottomLeft : '';
                $br = $hasRight ? $pattern->bottomRight : '';
                $midCount = max(0, $totalW - Util::visualWidth($bl) - Util::visualWidth($br));
                $mid = str_repeat($pattern->bottom, $midCount);
                $result[] = $styleRunes($bl . $mid . $br, $botGrad);
            }

            return implode("\n", $result);
        }

        // Top border
        if ($hasTop) {
            $tl = $hasLeft ? $pattern->topLeft : '';
            $tr = $hasRight ? $pattern->topRight : '';
            $midCount = max(0, $totalW - Util::visualWidth($tl) - Util::visualWidth($tr));
            $mid = str_repeat($pattern->top, $midCount);

            $topLine = $styleBorderPart($tl . $mid . $tr, $topColor);
            $result[] = $topLine;
        }

        // Side borders
        foreach ($lines as $line) {
            $leftStr = $hasLeft ? $styleBorderPart($pattern->left, $leftColor) : '';
            $rightStr = $hasRight ? $styleBorderPart($pattern->right, $rightColor) : '';
            $result[] = $leftStr . $line . $rightStr;
        }

        // Bottom border
        if ($hasBottom) {
            $bl = $hasLeft ? $pattern->bottomLeft : '';
            $br = $hasRight ? $pattern->bottomRight : '';
            $midCount = max(0, $totalW - Util::visualWidth($bl) - Util::visualWidth($br));
            $mid = str_repeat($pattern->bottom, $midCount);

            $bottomLine = $styleBorderPart($bl . $mid . $br, $bottomColor);
            $result[] = $bottomLine;
        }

        return implode("\n", $result);
    }

    public static function joinVertical(string $align, string ...$parts): string
    {
        $filtered = array_values(array_filter($parts, fn($p) => $p !== ''));
        if (empty($filtered)) {
            return '';
        }

        // Split all parts into lines
        $allLines = [];
        $maxW = 0;
        foreach ($filtered as $part) {
            $lines = explode("\n", $part);
            foreach ($lines as $l) {
                $w = Util::visualWidth($l);
                if ($w > $maxW) {
                    $maxW = $w;
                }
            }
            $allLines[] = $lines;
        }

        $result = [];
        foreach ($allLines as $partLines) {
            foreach ($partLines as $l) {
                $w = Util::visualWidth($l);
                if ($w < $maxW) {
                    $diff = $maxW - $w;
                    $l = match ($align) {
                        'center' => str_repeat(' ', (int) ($diff / 2)) . $l . str_repeat(' ', $diff - (int) ($diff / 2)),
                        'right' => str_repeat(' ', $diff) . $l,
                        default => $l . str_repeat(' ', $diff),
                    };
                }
                $result[] = $l;
            }
        }

        return implode("\n", $result);
    }

    public static function joinHorizontal(string $align, string ...$parts): string
    {
        $filtered = array_values(array_filter($parts, fn($p) => $p !== ''));
        if (empty($filtered)) {
            return '';
        }

        $blockLines = [];
        $blockWidths = [];
        $maxHeight = 0;

        foreach ($filtered as $part) {
            $lines = explode("\n", $part);
            $w = 0;
            foreach ($lines as $l) {
                $lw = Util::visualWidth($l);
                if ($lw > $w) {
                    $w = $lw;
                }
            }
            $blockLines[] = $lines;
            $blockWidths[] = $w;
            if (count($lines) > $maxHeight) {
                $maxHeight = count($lines);
            }
        }

        // Align heights by adding padding lines
        foreach ($blockLines as $b => $lines) {
            $curH = count($lines);
            $w = $blockWidths[$b];
            $empty = str_repeat(' ', $w);
            if ($curH < $maxHeight) {
                $diff = $maxHeight - $curH;
                match ($align) {
                    'bottom' => array_unshift($blockLines[$b], ...array_fill(0, $diff, $empty)),
                    'center' => (function () use (&$blockLines, $b, $diff, $empty) {
                        $top = (int) ($diff / 2);
                        $bottom = $diff - $top;
                        if ($top > 0) array_unshift($blockLines[$b], ...array_fill(0, $top, $empty));
                        if ($bottom > 0) array_push($blockLines[$b], ...array_fill(0, $bottom, $empty));
                    })(),
                    default => array_push($blockLines[$b], ...array_fill(0, $diff, $empty)),
                };
            }
        }

        // Merge row by row
        $rows = [];
        for ($r = 0; $r < $maxHeight; $r++) {
            $rowStr = '';
            foreach ($blockLines as $b => $lines) {
                $l = $lines[$r] ?? '';
                $lw = Util::visualWidth($l);
                $w = $blockWidths[$b];
                if ($lw < $w) {
                    $l .= str_repeat(' ', $w - $lw);
                }
                $rowStr .= $l;
            }
            $rows[] = $rowStr;
        }

        return implode("\n", $rows);
    }
}
