<?php

declare(strict_types=1);

namespace Termstrap\Image;

enum ColorMode: string
{
    case TrueColor = 'truecolor';
    case Ansi256 = 'ansi256';
    case Ansi16 = 'ansi16';

    public static function parse(string $val): self
    {
        return match (strtolower(trim($val))) {
            'ansi256', '256' => self::Ansi256,
            'ansi16', '16' => self::Ansi16,
            default => self::TrueColor,
        };
    }
}
