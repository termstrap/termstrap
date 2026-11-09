<?php

declare(strict_types=1);

namespace Termstrap\Tests;

use PHPUnit\Framework\Attributes\DataProvider;
use PHPUnit\Framework\TestCase;

final class E2eExamplesTest extends TestCase
{
 private string $examplesDir;

 protected function setUp(): void
 {
  parent::setUp();
  $this->examplesDir = dirname(__DIR__) . '/examples';
 }

 #[DataProvider('provideExampleNames')]
 public function testExampleExecutesAndProducesTerminalOutput(string $exampleName): void
 {
  $mainPhp = $this->examplesDir . '/' . $exampleName . '/main.php';
  $this->assertFileExists($mainPhp, "Example {$exampleName} must have main.php");

  $cmd = sprintf('%s %s', escapeshellcmd(PHP_BINARY), escapeshellarg($mainPhp));
  $output = (string) shell_exec($cmd);

  $this->assertNotEmpty($output, "Example {$exampleName} must produce non-empty terminal output");
 }

 /**
  * @return list<array{0: string}>
  */
 public static function provideExampleNames(): array
 {
  return [
   ['ascii'],
   ['ascii-all-fonts'],
   ['bootstrap'],
   ['borders'],
   ['breakpoints'],
   ['custom-css'],
   ['fiche'],
   ['glamour'],
   ['headings'],
   ['html'],
   ['lipgloss-features'],
   ['media-queries'],
   ['nested'],
   ['nested-no-md'],
   ['shadows'],
   ['spacing'],
   ['spinner-shimmer'],
   ['styling'],
   ['syntax'],
   ['syntax-custom'],
   ['syntax-themes'],
   ['themes'],
   ['torrent'],
   ['image/animated_gif'],
   ['image/caching'],
   ['image/colormodes'],
   ['image/detect'],
   ['image/formats'],
   ['image/grid'],
   ['image/local'],
   ['image/markdown'],
   ['image/protocols'],
   ['image/three-columns'],
  ];
 }
}
