<?php

declare(strict_types=1);

namespace Termstrap\Renderer;

use Termstrap\Util;

final class Shadow
{
    public static function shadowBottomRows(int $shadowSize): int
    {
        if ($shadowSize <= 0) {
            return 0;
        }
        return (int) (($shadowSize + 1) / 2);
    }

    public static function applyShadowWithWidth(string $content, int $shadowSize, int $maxWidth): string
    {
        if ($shadowSize <= 0 || $content === '') {
            return $content;
        }

        $lines = explode("\n", $content);
        $start = 0;
        $total = count($lines);

        while ($start < $total && trim($lines[$start]) === '') {
            $start++;
        }
        $end = $total;
        while ($end > $start && trim($lines[$end - 1]) === '') {
            $end--;
        }

        $leading = array_slice($lines, 0, $start);
        $visible = array_slice($lines, $start, $end - $start);
        $trailing = array_slice($lines, $end);

        if (empty($visible)) {
            return $content;
        }

        $contentWidth = Util::visualWidth($visible[0]);
        $effectiveShadow = $shadowSize;
        if ($maxWidth > 0 && $contentWidth + $shadowSize > maxWidth) {
            $effectiveShadow = $maxWidth - $contentWidth;
            if ($effectiveShadow < 1) {
                $effectiveShadow = 1;
            }
        }

        $shadowChar = $effectiveShadow >= 3 ? '▒' : '░';
        $shadowEscape = "\x1b[38;5;240m" . str_repeat($shadowChar, $effectiveShadow) . "\x1b[m";

        $result = $leading;
        foreach ($visible as $line) {
            $result[] = $line . $shadowEscape;
        }

        $bottomShadowEscape = "\x1b[38;5;240m" . str_repeat($shadowChar, $contentWidth) . "\x1b[m";
        $bottomRows = self::shadowBottomRows($effectiveShadow);
        for ($r = 0; $r < $bottomRows; $r++) {
            $result[] = str_repeat(' ', $effectiveShadow) . $bottomShadowEscape;
        }

        foreach ($trailing as $line) {
            $result[] = $line;
        }

        return implode("\n", $result);
    }
}
