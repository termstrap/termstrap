<?php

declare(strict_types=1);

namespace Termstrap\Image;

enum Protocol: string
{
    case HalfBlock = 'halfblock';
    case Sextant = 'sextant';
    case Kitty = 'kitty';
    case ITerm2 = 'iterm2';
    case Sixel = 'sixel';

    public function isGraphic(): bool
    {
        return $this === self::Kitty || $this === self::ITerm2 || $this === self::Sixel;
    }
}
