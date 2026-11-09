<?php

declare(strict_types=1);

namespace Termstrap\Tests;

use Dom\HTMLDocument;
use PHPUnit\Framework\TestCase;
use Termstrap\Css\Axis;
use Termstrap\Css\CssMatcher;
use Termstrap\Css\MediaCondition;
use Termstrap\Css\Preprocessor;
use Termstrap\Css\Specificity;
use Termstrap\Css\UnitContext;
use Termstrap\Css\Units;
use Termstrap\Markdown;
use Termstrap\Util;

final class CssPipelineTest extends TestCase
{
    public function testSpecificityCalculations(): void
    {
        $cases = [
            '*' => [0, 0, 0],
            'li' => [0, 0, 1],
            'ul li' => [0, 0, 2],
            'ul ol+li' => [0, 0, 3],
            'h1 + *[rel=up]' => [0, 1, 1],
            'ul ol li.red' => [0, 1, 3],
            'li.red.level' => [0, 2, 1],
            '#x34y' => [1, 0, 0],
            '#s12:not(FOO)' => [1, 0, 1],
            '.card > .card-header #title.active' => [1, 3, 0],
            '::before' => [0, 0, 1],
        ];

        foreach ($cases as $sel => [$a, $b, $c]) {
            $spec = Specificity::compute($sel);
            $this->assertSame($a, $spec->a, "Specificity A for {$sel}");
            $this->assertSame($b, $spec->b, "Specificity B for {$sel}");
            $this->assertSame($c, $spec->c, "Specificity C for {$sel}");
        }
    }

    public function testSpecificityComparison(): void
    {
        $id = new Specificity(1, 0, 0);
        $manyClasses = new Specificity(0, 10, 10);
        $twoClasses = new Specificity(0, 2, 0);
        $oneClassFiveTags = new Specificity(0, 1, 5);

        $this->assertTrue($manyClasses->less($id));
        $this->assertTrue($oneClassFiveTags->less($twoClasses));
    }

    public function testUnitsResolution(): void
    {
        $ctx = UnitContext::default(80);

        $this->assertSame(0, Units::resolveLength('0', Axis::X, $ctx));
        $this->assertSame(10, Units::resolveLength('10ch', Axis::X, $ctx));
        $this->assertSame(40, Units::resolveLength('50%', Axis::X, $ctx));
        $this->assertSame(80, Units::resolveLength('100vw', Axis::X, $ctx));
        $this->assertSame(24, Units::resolveLength('100vh', Axis::Y, $ctx));
        $this->assertSame(Units::LENGTH_FIT_CONTENT, Units::resolveLength('fit-content', Axis::X, $ctx));

        [$top, $right, $bottom, $left] = Units::resolveBox4D('1rem 2rem', $ctx);
        $this->assertSame(1, $top);
        $this->assertSame(4, $right);
        $this->assertSame(1, $bottom);
        $this->assertSame(4, $left);
    }

    public function testPreprocessorVariables(): void
    {
        $css = '
        :root {
            --bs-blue: #0d6efd;
            --bs-primary: var(--bs-blue);
            --c: #198754;
            --b: var(--c);
            --a: var(--b);
            --res: var(--missing, var(--also-missing, #ffc107));
        }
        ';
        $matcher = new CssMatcher('bootstrap', [$css]);

        $this->assertSame('#0d6efd', $matcher->preprocessor->globalVars['--bs-primary'] ?? null);
        $this->assertSame('#198754', $matcher->preprocessor->globalVars['--a'] ?? null);
        $this->assertSame('#ffc107', $matcher->preprocessor->globalVars['--res'] ?? null);
    }

    public function testPreprocessorCalc(): void
    {
        $this->assertSame('0.75rem', Preprocessor::evaluateCalc('calc(1.5rem * 0.5)'));
        $this->assertSame('0rem', Preprocessor::evaluateCalc('calc(-1 * 0rem)'));
        $this->assertSame('30px', Preprocessor::evaluateCalc('calc(10px + 20px)'));
        $this->assertSame('25px', Preprocessor::evaluateCalc('calc(40px - 15px)'));
        $this->assertSame('8px', Preprocessor::evaluateCalc('calc(2 * 4px)'));
        $this->assertSame('5px', Preprocessor::evaluateCalc('calc(10px / 2)'));
    }

    public function testPreprocessorRgba(): void
    {
        $this->assertSame('#0d6efd', Preprocessor::resolveColors('rgba(13, 110, 253, 1)', '#ffffff'));
        $this->assertSame('#0d6efd', Preprocessor::resolveColors('rgb(13, 110, 253)', '#ffffff'));
        $this->assertSame('#0d6efd', Preprocessor::resolveColors('rgb(13 110 253)', '#ffffff'));
        $this->assertSame('#0d6efd', Preprocessor::resolveColors('rgba(13, 110, 253, 100%)', '#ffffff'));
        $this->assertSame('transparent', Preprocessor::resolveColors('rgba(0, 0, 0, 0)', '#ffffff'));
    }

    public function testMediaQueryMatching(): void
    {
        $cond = MediaCondition::parse('screen and (min-width: 70ch)');
        $this->assertNotNull($cond);
        $this->assertTrue($cond->matches(80, 24, 'bootstrap'));
        $this->assertFalse($cond->matches(60, 24, 'bootstrap'));

        $darkCond = MediaCondition::parse('(prefers-color-scheme: dark)');
        $this->assertNotNull($darkCond);
        $this->assertTrue($darkCond->matches(80, 24, 'tokyonight'));
        $this->assertFalse($darkCond->matches(80, 24, 'bootstrap'));
    }

    public function testCssCascadeAndSpecificityResolution(): void
    {
        $css = '
        #target {
            color: #0d6efd; /* Blue (A=1) */
        }
        .target-class {
            color: #dc3545; /* Red (B=1) */
        }
        div {
            color: #ffc107; /* Yellow (C=1) */
        }
        ';
        $html = '<div id="target" class="target-class">Specificity Winner</div>';
        $doc = HTMLDocument::createFromString($html, LIBXML_NOERROR);
        $div = $doc->getElementById('target');
        $this->assertNotNull($div);

        $matcher = new CssMatcher('bootstrap', [$css]);
        $style = $matcher->computeStyleForNode($div, 80);

        // #target must win: #0d6efd
        $this->assertSame('#0d6efd', $style->fgColor);
    }

    public function testImportantOverridesHigherSpecificity(): void
    {
        $css = '
        #target {
            color: #0d6efd;
        }
        .target-class {
            color: #198754 !important;
        }
        ';
        $html = '<div id="target" class="target-class">Important Winner</div>';
        $doc = HTMLDocument::createFromString($html, LIBXML_NOERROR);
        $div = $doc->getElementById('target');
        $this->assertNotNull($div);

        $matcher = new CssMatcher('bootstrap', [$css]);
        $style = $matcher->computeStyleForNode($div, 80);

        // .target-class !important must win: #198754
        $this->assertSame('#198754', $style->fgColor);
    }

    public function testInlineCodeEscaping(): void
    {
        $raw = 'Here is `<hr>` and `<h1>title</h1>` code span.';
        $escaped = Util::escapeInlineCodeHtml($raw);
        $this->assertSame('Here is `&lt;hr&gt;` and `&lt;h1&gt;title&lt;/h1&gt;` code span.', $escaped);
    }

    public function testMarkdownPreservesRawHtml(): void
    {
        $md = "# Heading\n\n<div class=\"alert\">Raw HTML</div>";
        $html = Markdown::toHtml($md);
        $this->assertStringContainsString('<h1>Heading</h1>', $html);
        $this->assertStringContainsString('<div class="alert">Raw HTML</div>', $html);
    }
}
