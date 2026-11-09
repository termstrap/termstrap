<?php

declare(strict_types=1);

namespace Termstrap\Tests;

use PHPUnit\Framework\TestCase;
use Termstrap\Animation\Shimmer;
use Termstrap\Animation\Spinner;
use Termstrap\Ascii\Figlet;
use Termstrap\Image\ColorMode;
use Termstrap\Image\HalfBlockRenderer;
use Termstrap\Image\ITermRenderer;
use Termstrap\Image\KittyRenderer;
use Termstrap\Image\SixelRenderer;
use Termstrap\StickyRegion;
use Termstrap\Termstrap;

final class ImagesAndAsciiTest extends TestCase
{
    public function testAvailableAsciiFonts(): void
    {
        $fonts = Termstrap::availableAsciiFonts();
        $this->assertGreaterThan(140, count($fonts));
        $this->assertContains('standard', $fonts);
        $this->assertContains('doh', $fonts);
        $this->assertContains('slant', $fonts);
    }

    public function testAsciiFigletRendering(): void
    {
        $html = '<ascii font="standard">Hello</ascii>';
        $out = Termstrap::toAnsi($html, ['width' => 80, 'theme' => 'none']);
        $this->assertNotEmpty($out);
        $this->assertStringContainsString('|', $out);
    }

    public function testAsciiFitScaleSextant(): void
    {
        $lines = Figlet::render('SUPER LONG TEXT THAT EXCEEDS WIDTH', 'doh', 'scale', false, 30);
        $this->assertNotEmpty($lines);
        foreach ($lines as $l) {
            $this->assertLessThanOrEqual(30, mb_strwidth($l, 'UTF-8'));
        }
    }

    public function testSpinnerFrames(): void
    {
        $this->assertSame('⠋', Spinner::spinner(0));
        $this->assertSame('⠙', Spinner::spinner(1));
        $this->assertSame('⠋', Spinner::spinner(10));
    }

    public function testShimmerAnimation(): void
    {
        $shimmer0 = Shimmer::shimmer('Testing', 0);
        $shimmer1 = Shimmer::shimmer('Testing', 1);

        $this->assertNotEmpty($shimmer0);
        $this->assertNotEmpty($shimmer1);
        $this->assertNotSame($shimmer0, $shimmer1);
        $this->assertStringContainsString('38;2;', $shimmer0);
    }

    public function testStickyRegionsAndPinning(): void
    {
        $html = '<body>
            <div class="section"><h1 class="sticky">Header 1</h1><p>Line 1<br>Line 2<br>Line 3</p></div>
            <div class="section"><h1 class="sticky">Header 2</h1><p>Line 4<br>Line 5<br>Line 6</p></div>
        </body>';

        [$body, $regions] = Termstrap::renderStickyRegions($html, ['width' => 40, 'theme' => 'none']);
        $this->assertCount(2, $regions);
        $this->assertSame(0, $regions[0]->startLine);
        $this->assertGreaterThan(0, $regions[1]->startLine);

        // Pinning test
        $view = "Line 1\nLine 2\nLine 3\nLine 4";
        $pinned = Termstrap::pinSticky($view, $regions, 2);
        $this->assertStringContainsString('Header 1', $pinned);
    }

    public function testImagePlaceholderModes(): void
    {
        $html = '<img src="https://example.com/logo.png" alt="My Logo">';

        // Placeholder mode
        $outPlaceholder = Termstrap::toAnsi($html, ['width' => 60, 'image_mode' => 'placeholder']);
        $this->assertStringContainsString('[My Logo](https://example.com/logo.png)', $outPlaceholder);

        // Glamour mode
        $outGlamour = Termstrap::toAnsi($html, ['width' => 60, 'image_mode' => 'glamour']);
        $this->assertStringContainsString('Image: My Logo →', $outGlamour);
        $this->assertStringContainsString('https://example.com/logo.png', $outGlamour);
    }

    public function testHalfBlockRendererDeterministic(): void
    {
        $im = imagecreatetruecolor(2, 2);
        $red = imagecolorallocate($im, 255, 0, 0);
        $blue = imagecolorallocate($im, 0, 0, 255);
        imagesetpixel($im, 0, 0, $red);
        imagesetpixel($im, 1, 0, $blue);
        imagesetpixel($im, 0, 1, $blue);
        imagesetpixel($im, 1, 1, $red);

        $rendered = HalfBlockRenderer::render($im, 2, ColorMode::TrueColor, false);
        $this->assertStringContainsString('38;2;255;0;0', $rendered);
        $this->assertStringContainsString('48;2;0;0;255', $rendered);
        $this->assertStringContainsString('▀', $rendered);
    }

    public function testKittyRendererSequence(): void
    {
        $im = imagecreatetruecolor(2, 2);
        $rendered = KittyRenderer::render($im, 2, 2, 1);
        $this->assertStringStartsWith("\x1b_G", $rendered);
        $this->assertStringEndsWith("\x1b\\", $rendered);
        $this->assertStringContainsString('a=T,f=100', $rendered);
    }

    public function testITermRendererSequence(): void
    {
        $im = imagecreatetruecolor(2, 2);
        $rendered = ITermRenderer::render($im, 2, 1);
        $this->assertStringStartsWith("\x1b]1337;File=inline=1", $rendered);
        $this->assertStringContainsString('preserveAspectRatio=1', $rendered);
    }

    public function testSixelRendererSequence(): void
    {
        $im = imagecreatetruecolor(2, 2);
        $rendered = SixelRenderer::render($im, 2);
        $this->assertStringStartsWith("\x1bPq", $rendered);
        $this->assertStringEndsWith("\x1b\\\n", $rendered);
    }

    public function testRenderWithOverlays(): void
    {
        $im = imagecreatetruecolor(4, 4);
        imagepng($im, '/tmp/test_ov.png');

        putenv('TERMSTRAP_IMAGE_PROTOCOL=kitty');
        $ts = new Termstrap();
        [$output, $overlays] = $ts->renderWithOverlays('<img src="/tmp/test_ov.png" width="4">', ['width' => 20]);

        $this->assertNotEmpty($overlays);
        $this->assertSame(4, $overlays[0]->width);
        $this->assertStringStartsWith("\x1b_G", $overlays[0]->escape);
    }
}
