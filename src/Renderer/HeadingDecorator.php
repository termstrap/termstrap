<?php

declare(strict_types=1);

namespace Termstrap\Renderer;

final class HeadingDecorator
{
    public static function headingLevel(string $tag): int
    {
        $tag = strtolower($tag);
        if (strlen($tag) === 2 && $tag[0] === 'h' && $tag[1] >= '1' && $tag[1] <= '6') {
            return (int) $tag[1];
        }
        return 0;
    }

    public static function decorate(
        string $text,
        int $level,
        string $align,
        string $markerColor,
        string $h1Marker,
    ): string {
        $marker = str_repeat('#', $level);
        if ($level === 1) {
            if ($h1Marker === 'none' || $h1Marker === 'empty' || $h1Marker === '') {
                $marker = '';
            } elseif ($h1Marker !== '#') {
                $marker = $h1Marker;
            }
        }

        if ($marker !== '' && $markerColor !== '') {
            $markerEscape = Color::fgEscape($markerColor);
            if ($markerEscape !== '') {
                $marker = $markerEscape . $marker . "\x1b[m";
            }
        }

        $lines = explode("\n", $text);
        foreach ($lines as $i => $l) {
            if ($marker === '') {
                continue;
            }
            $lines[$i] = match ($align) {
                'center' => $marker . ' ' . $l . ' ' . $marker,
                'right' => $l . ' ' . $marker,
                default => $marker . ' ' . $l,
            };
        }

        return implode("\n", $lines);
    }
}
