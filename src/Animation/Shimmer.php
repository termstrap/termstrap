<?php

declare(strict_types=1);

namespace Termstrap\Animation;

use Termstrap\Renderer\Color;

final class Shimmer
{
    public static function shimmer(
        string $text,
        int $step,
        string $baseColor = '#44475a',
        string $highlightColor = '#ffffff',
        bool $bold = true,
    ): string {
        if ($text === '') {
            return '';
        }

        $chars = preg_split('//u', $text, -1, PREG_SPLIT_NO_EMPTY) ?: [];
        $n = count($chars);
        if ($n === 0) {
            return '';
        }

        $baseStops = [
            $baseColor,
            $highlightColor,
            $baseColor,
        ];

        $paletteSteps = max(3, $n * 2);
        $palette = Color::blend1D($paletteSteps, $baseStops);
        $pCount = count($palette);

        $sb = '';
        foreach ($chars as $i => $char) {
            $colorIdx = ($i - $step) % $pCount;
            if ($colorIdx < 0) {
                $colorIdx += $pCount;
            }
            [$r, $g, $b] = $palette[$colorIdx];
            $fg = sprintf("\x1b[38;2;%d;%d;%dm", $r, $g, $b);
            $bld = $bold ? "\x1b[1m" : '';
            $sb .= $fg . $bld . $char . "\x1b[m";
        }

        return $sb;
    }
}
