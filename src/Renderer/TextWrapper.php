<?php

declare(strict_types=1);

namespace Termstrap\Renderer;

use Symfony\Component\Tui\Ansi\TextWrapper as SymfonyTextWrapper;

final class TextWrapper
{
    public static function wrapText(string $text, int $width): string
    {
        if ($width <= 0 || $text === '') {
            return $text;
        }

        $lines = SymfonyTextWrapper::wrapTextWithAnsi($text, $width);
        $lines = array_map(function (string $l): string {
            $l = rtrim($l, ' ');
            if (str_ends_with($l, "\x1b[24m")) {
                $l .= "\x1b[59m";
            }
            return $l;
        }, $lines);

        return implode("\n", $lines);
    }
}
