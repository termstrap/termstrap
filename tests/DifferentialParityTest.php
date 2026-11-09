<?php

declare(strict_types=1);

namespace Termstrap\Tests;

use PHPUnit\Framework\TestCase;
use Termstrap\Termstrap;

final class DifferentialParityTest extends TestCase
{
    public function testHeadingsWithMarkerTinting(): void
    {
        $html = '<h1>Title One</h1><h2>Title Two</h2><h3>Title Three</h3>';
        $opts = ['width' => 60, 'theme' => 'none'];

        $oracle = OracleHarness::renderOracle($html, $opts);
        $native = Termstrap::toAnsi($html, $opts);

        OracleHarness::assertCellGridEquals($oracle, $native, 60, 'Headings parity');
    }

    public function testInlineStylesAndColors(): void
    {
        $html = '<p>This is <strong style="color: #ff0000;">Bold Red</strong> and <em style="color: #00ff00;">Italic Green</em>.</p>';
        $opts = ['width' => 50, 'theme' => 'none'];

        $oracle = OracleHarness::renderOracle($html, $opts);
        $native = Termstrap::toAnsi($html, $opts);

        OracleHarness::assertCellGridEquals($oracle, $native, 50, 'Inline styles parity');
    }

    public function testBorderedBoxWithPadding(): void
    {
        $html = '<div class="border rounded p-1">Box Content</div>';
        $opts = ['width' => 40, 'theme' => 'none'];

        $oracle = OracleHarness::renderOracle($html, $opts);
        $native = Termstrap::toAnsi($html, $opts);

        OracleHarness::assertCellGridEquals($oracle, $native, 40, 'Bordered box parity');
    }

    public function testListsParity(): void
    {
        $html = '<ul><li>First item</li><li>Second item</li></ul><ol><li>One</li><li>Two</li></ol>';
        $opts = ['width' => 40, 'theme' => 'none'];

        $oracle = OracleHarness::renderOracle($html, $opts);
        $native = Termstrap::toAnsi($html, $opts);

        OracleHarness::assertCellGridEquals($oracle, $native, 40, 'Lists parity');
    }

    public function testTableParity(): void
    {
        $html = '<table><thead><tr><th>User</th><th>Role</th></tr></thead><tbody><tr><td>Alice</td><td>Admin</td></tr><tr><td>Bob</td><td>User</td></tr></tbody></table>';
        $opts = ['width' => 40, 'theme' => 'none'];

        $oracle = OracleHarness::renderOracle($html, $opts);
        $native = Termstrap::toAnsi($html, $opts);

        OracleHarness::assertCellGridEquals($oracle, $native, 40, 'Table parity');
    }

    public function testGridBreakpointsParityStackedVsSideBySide(): void
    {
        $html = '<div class="row"><div class="col-md-6 border">Col A</div><div class="col-md-6 border">Col B</div></div>';

        // 1. Stacked at width 70 (< MD 100)
        $opts70 = ['width' => 70, 'theme' => 'bootstrap'];
        $oracle70 = OracleHarness::renderOracle($html, $opts70);
        $native70 = Termstrap::toAnsi($html, $opts70);
        OracleHarness::assertCellGridEquals($oracle70, $native70, 70, 'Grid stacked at 70 cols');

        // 2. Side-by-side at width 120 (>= MD 100)
        $opts120 = ['width' => 120, 'theme' => 'bootstrap'];
        $oracle120 = OracleHarness::renderOracle($html, $opts120);
        $native120 = Termstrap::toAnsi($html, $opts120);
        OracleHarness::assertCellGridEquals($oracle120, $native120, 120, 'Grid side-by-side at 120 cols');
    }

    public function testBackgroundFillParity(): void
    {
        $html = '<div style="background-color: #212529; color: #ffffff;">Dark Box with <span style="font-weight: bold;">Bold</span> text</div>';
        $opts = ['width' => 40, 'theme' => 'none'];

        $oracle = OracleHarness::renderOracle($html, $opts);
        $native = Termstrap::toAnsi($html, $opts);

        OracleHarness::assertCellGridEquals($oracle, $native, 40, 'Background fill parity');
    }

    public function testDropShadowParity(): void
    {
        $html = '<div style="border: 1px solid #ffffff; box-shadow: 1;">Card</div>';
        $opts = ['width' => 30, 'theme' => 'none'];

        $oracle = OracleHarness::renderOracle($html, $opts);
        $native = Termstrap::toAnsi($html, $opts);

        OracleHarness::assertCellGridEquals($oracle, $native, 30, 'Drop shadow parity');
    }

    public function testCodeBlockParity(): void
    {
        $html = '<pre><code class="language-php">echo "Hello World";</code></pre>';
        $opts = ['width' => 50, 'theme' => 'none', 'disable_syntax_highlight' => true];

        $oracle = OracleHarness::renderOracle($html, $opts);
        $native = Termstrap::toAnsi($html, $opts);

        OracleHarness::assertCellGridEquals($oracle, $native, 50, 'Code block parity');
    }

    public function testOsc8HyperlinksParity(): void
    {
        $html = '<p>Visit <a href="https://example.com">Example</a> website.</p>';
        $opts = ['width' => 60, 'theme' => 'none', 'link_mode' => 'osc8'];

        $oracle = OracleHarness::renderOracle($html, $opts);
        $native = Termstrap::toAnsi($html, $opts);

        $this->assertStringContainsString('https://example.com', $native);
        $this->assertStringContainsString(']8;id=', $native);
    }

    public function testFootnotesLinkModeParity(): void
    {
        $html = '<p>See <a href="https://php.net">PHP</a> and <a href="https://symfony.com">Symfony</a>.</p>';
        $opts = ['width' => 60, 'theme' => 'none', 'link_mode' => 'footnotes'];

        $oracle = OracleHarness::renderOracle($html, $opts);
        $native = Termstrap::toAnsi($html, $opts);

        $this->assertStringContainsString('[1]', $native);
        $this->assertStringContainsString('[2]', $native);
        $this->assertStringContainsString('https://php.net', $native);
        $this->assertStringContainsString('https://symfony.com', $native);
    }

    public function testAllBreakpointBoundariesParity(): void
    {
        $html = '<div class="row">
            <div class="col-sm-6 col-md-4 col-lg-3 col-xl-2 col-xxl-1 border">Item A</div>
            <div class="col-sm-6 col-md-4 col-lg-3 col-xl-2 col-xxl-1 border">Item B</div>
        </div>';

        $boundaryWidths = [69, 70, 99, 100, 129, 130, 159, 160, 189, 190];
        foreach ($boundaryWidths as $w) {
            $opts = ['width' => $w, 'theme' => 'bootstrap'];
            $oracle = OracleHarness::renderOracle($html, $opts);
            $native = Termstrap::toAnsi($html, $opts);
            OracleHarness::assertCellGridEquals($oracle, $native, $w, "Boundary width {$w} parity");
        }
    }

    public function testNestedTwelveColumnRemainderParity(): void
    {
        $html = '<div class="row">
            <div class="col-7 border">
                <div class="row">
                    <div class="col-4 border">Sub 1</div>
                    <div class="col-4 border">Sub 2</div>
                    <div class="col-4 border">Sub 3</div>
                </div>
            </div>
            <div class="col-5 border">Side</div>
        </div>';

        $opts = ['width' => 83, 'theme' => 'none'];
        $oracle = OracleHarness::renderOracle($html, $opts);
        $native = Termstrap::toAnsi($html, $opts);
        OracleHarness::assertCellGridEquals($oracle, $native, 83, 'Nested 12-column remainder parity at 83 cols');
    }

    public function testCssOriginAndImportantCascadeParity(): void
    {
        $css = '
        #main-title { color: #0d6efd; font-weight: normal; }
        .override { color: #dc3545 !important; }
        ';
        $html = '<h2 id="main-title" class="override" style="color: #ffc107;">Styled Title</h2>';
        $opts = ['width' => 50, 'theme' => 'none', 'stylesheets' => [$css]];

        $oracle = OracleHarness::renderOracle($html, $opts);
        $native = Termstrap::toAnsi($html, $opts);
        OracleHarness::assertCellGridEquals($oracle, $native, 50, 'CSS origin & !important cascade parity');
    }

    public function testNestedVarFallbackAndCalcParity(): void
    {
        $css = '
        :root {
            --base-w: 20px;
            --computed-w: calc(var(--missing-var, var(--base-w)) + 10px);
        }
        .box {
            width: var(--computed-w);
            color: var(--unfound, #008080);
        }
        ';
        $html = '<div class="box border">Calc & Var</div>';
        $opts = ['width' => 60, 'theme' => 'none', 'stylesheets' => [$css]];

        $oracle = OracleHarness::renderOracle($html, $opts);
        $native = Termstrap::toAnsi($html, $opts);
        OracleHarness::assertCellGridEquals($oracle, $native, 60, 'Nested var() fallback & calc() parity');
    }

    public function testUnicodeCjkAndWrappingParity(): void
    {
        $html = '<p style="border: 1px solid #738087;">日本語の長いテキストの折り返しテストです。端末の幅に応じて綺麗に改行される必要があります。</p>';
        $opts = ['width' => 35, 'theme' => 'none'];

        $oracle = OracleHarness::renderOracle($html, $opts);
        $native = Termstrap::toAnsi($html, $opts);
        OracleHarness::assertCellGridEquals($oracle, $native, 35, 'Unicode CJK wrapping parity');
    }

    public function testOsc8MultilineWrappingParity(): void
    {
        $html = '<p>Check out <a href="https://very-long-domain-name-example.com/some/deep/resource/path">This is a rather long anchor link text that is intentionally designed to wrap across lines</a> inside the terminal.</p>';
        $opts = ['width' => 45, 'theme' => 'none', 'link_mode' => 'osc8'];

        $oracle = OracleHarness::renderOracle($html, $opts);
        $native = Termstrap::toAnsi($html, $opts);
        OracleHarness::assertCellGridEquals($oracle, $native, 45, 'OSC 8 multiline wrapping parity');
    }
}
