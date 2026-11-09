<?php

declare(strict_types=1);

namespace Termstrap\Image;

final class Detect
{
    public static function detectProtocol(): Protocol
    {
        $override = getenv('TERMSTRAP_IMAGE_PROTOCOL');
        if ($override !== false && $override !== '') {
            return match (strtolower(trim($override))) {
                'kitty' => Protocol::Kitty,
                'iterm2', 'iterm' => Protocol::ITerm2,
                'sixel' => Protocol::Sixel,
                'sextant', 'sextants', 'block6' => Protocol::Sextant,
                default => Protocol::HalfBlock,
            };
        }

        $term = (string) getenv('TERM');
        $termProgram = (string) getenv('TERM_PROGRAM');
        $lcTerminal = (string) getenv('LC_TERMINAL');

        if ($term === 'xterm-kitty' || strcasecmp($termProgram, 'kitty') === 0) {
            return Protocol::Kitty;
        }

        if (strcasecmp($termProgram, 'WezTerm') === 0) {
            return Protocol::Kitty;
        }

        if (strcasecmp($termProgram, 'iTerm.app') === 0 || strcasecmp($lcTerminal, 'iTerm2') === 0) {
            return Protocol::ITerm2;
        }

        $termLower = strtolower($term);
        $termProgLower = strtolower($termProgram);
        if (
            str_contains($termLower, 'mlterm')
            || str_contains($termLower, 'yaft')
            || in_array($termProgLower, ['foot', 'contour', 'konsole'], true)
        ) {
            return Protocol::Sixel;
        }

        return Protocol::HalfBlock;
    }
}
