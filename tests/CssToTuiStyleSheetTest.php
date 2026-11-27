<?php

declare(strict_types=1);

namespace Termstrap\Tests;

use PHPUnit\Framework\TestCase;
use Symfony\Component\Tui\Style\Direction;
use Termstrap\Bridge\Symfony\CssToTuiStyleSheet;

final class CssToTuiStyleSheetTest extends TestCase
{
    public function testConvertsBasicCssRules(): void
    {
        $css = <<<'CSS'
.header {
    color: #89b4fa;
    font-weight: bold;
    padding: 1 2;
}

.modal {
    border: 1px solid #7aa2f7;
    border-radius: 4px;
    background: #1e1e2e;
}

.hidden {
    display: none;
}

.hsplit {
    flex-direction: row;
    gap: 1;
}
CSS;

        $sheet = CssToTuiStyleSheet::fromCss($css);
        $this->assertNotNull($sheet);

        $rulesProperty = new \ReflectionProperty($sheet, 'rules');
        /** @var array<string, \Symfony\Component\Tui\Style\Style> $rules */
        $rules = $rulesProperty->getValue($sheet);

        $this->assertArrayHasKey('.header', $rules);
        $this->assertSame(['r' => 137, 'g' => 180, 'b' => 250], $rules['.header']->getColor()?->toRgb());
        $this->assertTrue($rules['.header']->getBold());
        $this->assertSame(1, $rules['.header']->getPadding()?->top);
        $this->assertSame(2, $rules['.header']->getPadding()?->right);

        $this->assertArrayHasKey('.modal', $rules);
        $this->assertSame(['r' => 30, 'g' => 30, 'b' => 46], $rules['.modal']->getBackground()?->toRgb());
        $this->assertNotNull($rules['.modal']->getBorder());

        $this->assertArrayHasKey('.hidden', $rules);
        $this->assertTrue($rules['.hidden']->getHidden());

        $this->assertArrayHasKey('.hsplit', $rules);
        $this->assertSame(Direction::Horizontal, $rules['.hsplit']->getDirection());
        $this->assertSame(1, $rules['.hsplit']->getGap());
    }

    public function testConvertsBreakpoints(): void
    {
        $css = <<<'CSS'
.sidebar {
    display: none;
}

@media (min-width: 120) {
    .sidebar {
        display: block;
        flex: 1;
    }
}
CSS;

        $sheet = CssToTuiStyleSheet::fromCss($css);
        $breakpointsProperty = new \ReflectionProperty($sheet, 'breakpoints');
        /** @var array<int, array<string, \Symfony\Component\Tui\Style\Style>> $breakpoints */
        $breakpoints = $breakpointsProperty->getValue($sheet);

        $this->assertArrayHasKey(120, $breakpoints);
        $this->assertArrayHasKey('.sidebar', $breakpoints[120]);
        $this->assertSame(1, $breakpoints[120]['.sidebar']->getFlex());
    }
}
