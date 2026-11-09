<?php

declare(strict_types=1);

namespace Termstrap\Css;

final readonly class MediaClause
{
    public function __construct(
        public string $mediaType = 'all',
        public bool $negate = false,
        public int $minWidth = -1,
        public int $maxWidth = -1,
        public int $minHeight = -1,
        public int $maxHeight = -1,
        public string $orientation = '',
        public string $colorScheme = '',
    ) {}

    public function matches(int $termWidth, int $termHeight, bool $isDark): bool
    {
        $matchesType = match ($this->mediaType) {
            'print', 'speech' => false,
            default => true,
        };

        if ($this->negate) {
            $matchesType = !$matchesType;
        }
        if (!$matchesType) {
            return false;
        }

        // Width boundaries
        if ($this->minWidth >= 0 && $termWidth < $this->minWidth) {
            return false;
        }
        if ($this->maxWidth >= 0 && $termWidth > $this->maxWidth) {
            return false;
        }

        // Height boundaries
        if ($this->minHeight >= 0 && $termHeight < $this->minHeight) {
            return false;
        }
        if ($this->maxHeight >= 0 && $termHeight > $this->maxHeight) {
            return false;
        }

        // Orientation
        if ($this->orientation !== '') {
            if ($this->orientation === 'landscape' && $termWidth < $termHeight) {
                return false;
            }
            if ($this->orientation === 'portrait' && $termWidth >= $termHeight) {
                return false;
            }
        }

        // Color scheme
        if ($this->colorScheme !== '') {
            if ($this->colorScheme === 'dark' && !$isDark) {
                return false;
            }
            if ($this->colorScheme === 'light' && $isDark) {
                return false;
            }
        }

        return true;
    }
}
