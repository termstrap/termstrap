<?php

declare(strict_types=1);

namespace Termstrap\Css;

final readonly class UnitContext
{
    public function __construct(
        public int $parentWidth = 80,
        public int $parentHeight = 24,
        public int $terminalWidth = 80,
        public int $terminalHeight = 24,
        public float $rootFontSize = 16.0,
    ) {}

    public static function default(int $termWidth): self
    {
        $tw = $termWidth > 0 ? $termWidth : 80;
        return new self(
            parentWidth: $tw,
            parentHeight: 24,
            terminalWidth: $tw,
            terminalHeight: 24,
            rootFontSize: 16.0,
        );
    }
}
