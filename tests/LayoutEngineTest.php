<?php

declare(strict_types=1);

namespace Termstrap\Tests;

use Dom\HTMLDocument;
use PHPUnit\Framework\TestCase;
use Termstrap\Css\CssMatcher;
use Termstrap\Layout\LayoutEngine;
use Termstrap\Layout\NodeType;

final class LayoutEngineTest extends TestCase
{
    public function testThreeColumnGridRemainderPreserving(): void
    {
        $html = '<div class="row">
            <div class="col-4 border rounded p-1">Col 1</div>
            <div class="col-4 border rounded p-1">Col 2</div>
            <div class="col-4 border rounded p-1">Col 3</div>
        </div>';

        $doc = HTMLDocument::createFromString($html, LIBXML_NOERROR);
        $body = $doc->getElementsByTagName('body')->item(0);
        $this->assertNotNull($body);

        $matcher = CssMatcher::create('none');
        $tree = LayoutEngine::buildRenderTree($body, $matcher, 100);
        $this->assertNotNull($tree);

        LayoutEngine::computeLayout($tree, 100);

        // Target row is the first child of body (the div.row)
        $row = $tree->children[0];
        $this->assertSame(3, count($row->children));

        $col1 = $row->children[0];
        $col2 = $row->children[1];
        $col3 = $row->children[2];

        // 100 / 3 = 33, 33, 34 (sum must be exactly 100)
        $sum = $col1->allocatedWidth + $col2->allocatedWidth + $col3->allocatedWidth;
        $this->assertSame(100, $sum);
        $this->assertSame(33, $col1->allocatedWidth);
        $this->assertSame(33, $col2->allocatedWidth);
        $this->assertSame(34, $col3->allocatedWidth);
    }

    public function testResponsiveBreakpointStackingVsSideBySide(): void
    {
        $html = '<div class="row">
            <div class="col-md-6">Col 1</div>
            <div class="col-md-6">Col 2</div>
        </div>';

        $doc = HTMLDocument::createFromString($html, LIBXML_NOERROR);
        $body = $doc->getElementsByTagName('body')->item(0);
        $this->assertNotNull($body);

        $matcher = CssMatcher::create('bootstrap');

        // At width 70 (< MD breakpoint 100), col-md-6 stacks at 12 cols each
        $tree70 = LayoutEngine::buildRenderTree($body, $matcher, 70);
        $this->assertNotNull($tree70);
        LayoutEngine::computeLayout($tree70, 70);
        $row70 = $tree70->children[0];
        $this->assertSame(70, $row70->children[0]->allocatedWidth);
        $this->assertSame(70, $row70->children[1]->allocatedWidth);

        // At width 120 (>= MD breakpoint 100), col-md-6 sits side-by-side at 6 cols each (60 cols each)
        $tree120 = LayoutEngine::buildRenderTree($body, $matcher, 120);
        $this->assertNotNull($tree120);
        LayoutEngine::computeLayout($tree120, 120);
        $row120 = $tree120->children[0];
        $this->assertSame(60, $row120->children[0]->allocatedWidth);
        $this->assertSame(60, $row120->children[1]->allocatedWidth);
    }

    public function testBoxDimensionsLadder(): void
    {
        $html = '<div style="margin: 2px 4px; padding: 1px 3px; border: 1px solid #fff; box-shadow: 1;">Box</div>';
        $doc = HTMLDocument::createFromString($html, LIBXML_NOERROR);
        $body = $doc->getElementsByTagName('body')->item(0);
        $this->assertNotNull($body);

        $matcher = CssMatcher::create('none');
        $tree = LayoutEngine::buildRenderTree($body, $matcher, 80);
        $this->assertNotNull($tree);

        LayoutEngine::computeLayout($tree, 80);
        $box = $tree->children[0];

        $this->assertSame(80, $box->allocatedWidth);
        // marginLeft = 1, marginRight = 1 (2px / 8px approx 1 cell), shadow = 1
        // BoxWidth = AllocatedWidth - marginLeft - marginRight - shadow
        $this->assertSame(77, $box->boxWidth);
        $this->assertSame(77, $box->lipglossWidth);
        // ContentWidth = LipglossWidth - borderH(2) - paddingH(2: padding-left 3px=1, padding-right 3px=1)
        $this->assertSame(73, $box->contentWidth);
    }

    public function testPictureSourceMediaSelection(): void
    {
        $html = '<picture>
            <source media="(min-width: 100ch)" src="large.png">
            <source media="(max-width: 99ch)" src="small.png">
            <img src="fallback.png" alt="Test">
        </picture>';

        $doc = HTMLDocument::createFromString($html, LIBXML_NOERROR);
        $body = $doc->getElementsByTagName('body')->item(0);
        $this->assertNotNull($body);

        $matcher = CssMatcher::create('none');

        $treeLarge = LayoutEngine::buildRenderTree($body, $matcher, 120);
        $this->assertNotNull($treeLarge);
        $imgLarge = $treeLarge->children[0];
        $this->assertSame(NodeType::Image, $imgLarge->type);
        $this->assertSame('large.png', $imgLarge->src);

        $treeSmall = LayoutEngine::buildRenderTree($body, $matcher, 80);
        $this->assertNotNull($treeSmall);
        $imgSmall = $treeSmall->children[0];
        $this->assertSame(NodeType::Image, $imgSmall->type);
        $this->assertSame('small.png', $imgSmall->src);
    }
}
