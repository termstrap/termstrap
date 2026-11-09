<?php

declare(strict_types=1);

namespace Termstrap;

final readonly class StickyRegion
{
    public function __construct(
        public int $startLine,
        public int $height,
        public string $header,
    ) {}
}
