<?php

declare(strict_types=1);

namespace Termstrap\Layout;

use Termstrap\Css\ComputedStyle;

final class RenderNode
{
    public NodeType $type = NodeType::Element;
    public string $tag = '';
    public string $text = '';
    public string $src = '';
    public string $alt = '';
    public string $href = ''; // URL for <a> tags
    public string $lang = ''; // Source language hint for <pre> code blocks
    public string $font = ''; // Font hint for <ascii> art blocks
    public bool $strict = false;
    public string $fit = ''; // <ascii> fit mode: "clip" or "scale"
    public bool $isTree = false; // ul/ol flagged as a tree

    public ComputedStyle $style;

    /** @var list<RenderNode> */
    public array $children = [];

    // Layout Geometry (calculated in character cells)
    public int $allocatedWidth = 0; // Parent allocated width
    public int $boxWidth = 0;       // Width between outer border edges
    public int $lipglossWidth = 0;  // BoxWidth - Borders
    public int $contentWidth = 0;   // LipglossWidth - Padding
    public int $height = 0;
    public int $x = 0;              // Relative offset in parent
    public int $y = 0;

    public function __construct(
        NodeType $type = NodeType::Element,
        ?ComputedStyle $style = null,
    ) {
        $this->type = $type;
        $this->style = $style ?? new ComputedStyle();
    }

    public function __clone()
    {
        $this->style = clone $this->style;
    }
}
