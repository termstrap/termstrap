<?php

declare(strict_types=1);

namespace Termstrap\Renderer;

final readonly class BorderPattern
{
    public function __construct(
        public string $top = '─',
        public string $bottom = '─',
        public string $left = '│',
        public string $right = '│',
        public string $topLeft = '┌',
        public string $topRight = '┐',
        public string $bottomLeft = '└',
        public string $bottomRight = '┘',
        public string $middleLeft = '├',
        public string $middleRight = '┤',
        public string $middle = '┼',
        public string $middleTop = '┬',
        public string $middleBottom = '┴',
    ) {}

    public static function forStyle(string $style, bool $rounded = false): self
    {
        return match ($style) {
            'double' => new self(
                top: '═',
                bottom: '═',
                left: '║',
                right: '║',
                topLeft: '╔',
                topRight: '╗',
                bottomLeft: '╚',
                bottomRight: '╝',
                middleLeft: '╠',
                middleRight: '╣',
                middle: '╬',
                middleTop: '╦',
                middleBottom: '╩'
            ),
            'thick' => new self(
                top: '━',
                bottom: '━',
                left: '┃',
                right: '┃',
                topLeft: '┏',
                topRight: '┓',
                bottomLeft: '┗',
                bottomRight: '┛',
                middleLeft: '┣',
                middleRight: '┫',
                middle: '╋',
                middleTop: '┳',
                middleBottom: '┻'
            ),
            'block' => new self(
                top: '█',
                bottom: '█',
                left: '█',
                right: '█',
                topLeft: '█',
                topRight: '█',
                bottomLeft: '█',
                bottomRight: '█',
                middleLeft: '█',
                middleRight: '█',
                middle: '█',
                middleTop: '█',
                middleBottom: '█'
            ),
            'outer-half-block' => new self(
                top: '▀',
                bottom: '▄',
                left: '▌',
                right: '▐',
                topLeft: '▛',
                topRight: '▜',
                bottomLeft: '▙',
                bottomRight: '▟',
            ),
            'inner-half-block' => new self(
                top: '▄',
                bottom: '▀',
                left: '▐',
                right: '▌',
                topLeft: '▗',
                topRight: '▖',
                bottomLeft: '▝',
                bottomRight: '▘',
            ),
            'markdown' => new self(
                top: '-',
                bottom: '-',
                left: '|',
                right: '|',
                topLeft: '|',
                topRight: '|',
                bottomLeft: '|',
                bottomRight: '|',
                middleLeft: '|',
                middleRight: '|',
                middle: '|',
                middleTop: '|',
                middleBottom: '|'
            ),
            'ascii' => new self(
                top: '-',
                bottom: '-',
                left: '|',
                right: '|',
                topLeft: '+',
                topRight: '+',
                bottomLeft: '+',
                bottomRight: '+',
                middleLeft: '+',
                middleRight: '+',
                middle: '+',
                middleTop: '+',
                middleBottom: '+'
            ),
            'hidden', 'none' => new self(
                top: ' ',
                bottom: ' ',
                left: ' ',
                right: ' ',
                topLeft: ' ',
                topRight: ' ',
                bottomLeft: ' ',
                bottomRight: ' ',
                middleLeft: ' ',
                middleRight: ' ',
                middle: ' ',
                middleTop: ' ',
                middleBottom: ' '
            ),
            'dashed' => $rounded
                ? new self(
                    top: '╌',
                    bottom: '╌',
                    left: '╎',
                    right: '╎',
                    topLeft: '╭',
                    topRight: '╮',
                    bottomLeft: '╰',
                    bottomRight: '╯',
                )
                : new self(
                    top: '╌',
                    bottom: '╌',
                    left: '╎',
                    right: '╎',
                    topLeft: '┌',
                    topRight: '┐',
                    bottomLeft: '└',
                    bottomRight: '┘',
                ),
            'dotted' => $rounded
                ? new self(
                    top: '┄',
                    bottom: '┄',
                    left: '┆',
                    right: '┆',
                    topLeft: '╭',
                    topRight: '╮',
                    bottomLeft: '╰',
                    bottomRight: '╯',
                )
                : new self(
                    top: '┄',
                    bottom: '┄',
                    left: '┆',
                    right: '┆',
                    topLeft: '┌',
                    topRight: '┐',
                    bottomLeft: '└',
                    bottomRight: '┘',
                ),
            default => $rounded
                ? new self(
                    top: '─',
                    bottom: '─',
                    left: '│',
                    right: '│',
                    topLeft: '╭',
                    topRight: '╮',
                    bottomLeft: '╰',
                    bottomRight: '╯',
                )
                : new self(
                    top: '─',
                    bottom: '─',
                    left: '│',
                    right: '│',
                    topLeft: '┌',
                    topRight: '┐',
                    bottomLeft: '└',
                    bottomRight: '┘',
                ),
        };
    }
}
