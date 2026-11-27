<?php

declare(strict_types=1);

namespace Termstrap\Css;

final class ComputedStyle
{
    public string $display = 'block';

    // Margins (terminal cells/lines)
    public int $marginTop = 0;
    public int $marginBottom = 0;
    public int $marginLeft = 0;
    public int $marginRight = 0;

    // Paddings (terminal cells/lines)
    public int $paddingTop = 0;
    public int $paddingBottom = 0;
    public int $paddingLeft = 0;
    public int $paddingRight = 0;

    // Borders
    public bool $border = false;
    public bool $borderTop = false;
    public bool $borderBottom = false;
    public bool $borderLeft = false;
    public bool $borderRight = false;
    public string $borderColor = '';
    public string $borderStyle = ''; // "", "solid", "double", "dashed", "dotted", "hidden", "thick", etc.
    public bool $rounded = false;

    // Per-side border colors
    public string $borderTopColor = '';
    public string $borderRightColor = '';
    public string $borderBottomColor = '';
    public string $borderLeftColor = '';
    public bool $borderShimmer = false;
    /** @var list<string> */
    public array $borderBlend = [];

    // Shadow (0 = none, 1 = sm, 2 = md, 3 = lg)
    public int $shadow = 0;

    // Colors
    public string $fgColor = '';
    public bool $fgColorSet = false;
    public string $bgColor = '';
    /** @var list<string> */
    public array $bgGradient = [];

    // Typography
    public bool $bold = false;
    public bool $boldSet = false;
    public bool $italic = false;
    public bool $italicSet = false;
    public bool $underline = false;
    public bool $underlineSet = false;
    public string $underlineStyle = ''; // single, double, curly, dotted, dashed, none
    public bool $underlineStyleSet = false;
    public string $underlineColor = '';
    public bool $underlineColorSet = false;
    public bool $strikethrough = false;
    public bool $strikethroughSet = false;
    public bool $faint = false;
    public bool $faintSet = false;
    public bool $blink = false;
    public bool $blinkSet = false;
    public bool $shimmer = false;
    public bool $shimmerSet = false;
    public string $shimmerHighlight = '';
    public bool $spinner = false;
    public bool $spinnerSet = false;

    public string $textAlign = 'left'; // left, center, right
    public bool $textAlignSet = false;
    public string $textTransform = ''; // "", "uppercase", "lowercase", "capitalize"
    public bool $textTransformSet = false;
    public string $whiteSpace = ''; // "", "normal", "nowrap", "pre", "pre-wrap", "pre-line", "break-spaces"
    public bool $whiteSpaceSet = false;

    // Flex alignment
    public string $alignItems = 'top'; // top, center, bottom
    public bool $alignItemsSet = false;
    public string $justifyContent = 'left'; // left, center, right, space-between
    public string $justifyContentMode = ''; // "", "space-between", "space-around"
    public bool $justifyContentSet = false;
    public string $verticalAlign = 'top'; // top, center, bottom
    public bool $verticalAlignSet = false;

    // List rendering
    public string $listStyle = ''; // "", "disc", "circle", "square", "decimal", etc.

    // Grid / Sizing
    public int $colSpan = 0; // 1-12 for flex items
    public bool $colAuto = false;
    public int $rowCols = 0; // 1-12 items per row
    public int $width = 0; // explicit width if specified (or LENGTH_FIT_CONTENT = -1)
    public int $maxWidth = 0;
    public int $maxHeight = 0;

    public function clone(): self
    {
        return clone $this;
    }
}
