<?php

declare(strict_types=1);

namespace Termstrap\Tests;

use PHPUnit\Framework\TestCase;

final class OracleHarnessTest extends TestCase
{
    public function testOracleRendersSampleHtml(): void
    {
        $html = '<h1>Hello World</h1>';
        $out = OracleHarness::renderOracle($html, ['width' => 80, 'theme' => 'none']);

        $this->assertNotEmpty($out);
        $this->assertStringContainsString('Hello World', $out);
    }

    public function testCellGridComparisonDetectsIdenticalOutput(): void
    {
        $ansi = "\x1b[1;38;2;255;0;0mHello\x1b[m";
        OracleHarness::assertCellGridEquals($ansi, $ansi, 40);
        $this->assertTrue(true);
    }
}
