<?php

declare(strict_types=1);

namespace Termstrap\Image;

final readonly class ImageOverlay
{
    public function __construct(
        public int $row,
        public int $col,
        public int $width,
        public int $height,
        public string $escape,
    ) {}
}
