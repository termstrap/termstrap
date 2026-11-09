<?php

declare(strict_types=1);

namespace Termstrap\Tests;

use PHPUnit\Framework\Assert;
use RuntimeException;
use Symfony\Component\Tui\Terminal\ScreenBuffer;

final class OracleHarness
{
 private static ?string $oracleCliPath = null;

 public static function getOracleCliPath(): string
 {
  if (self::$oracleCliPath !== null) {
   return self::$oracleCliPath;
  }

  $root = dirname(__DIR__, 2);
  $libPath = $root . '/termstrap.ffi/libtermstrap.dylib';
  if (PHP_OS_FAMILY === 'Linux') {
   $libPath = $root . '/termstrap.ffi/libtermstrap.so';
  } elseif (PHP_OS_FAMILY === 'Windows') {
   $libPath = $root . '/termstrap.ffi/libtermstrap.dll';
  }

  if (!file_exists($libPath)) {
   // Check if source exists and can be built
   $makePath = $root . '/termstrap.src/Makefile';
   if (file_exists($makePath)) {
    $cmd = sprintf('cd %s && make build-ffi', escapeshellarg($root . '/termstrap.src'));
    exec($cmd, $out, $ret);
    if ($ret !== 0 || !file_exists($libPath)) {
     throw new RuntimeException("Go oracle shared library unavailable at: {$libPath}. Build with 'cd termstrap.src && make build-ffi'");
    }
   } else {
    throw new RuntimeException("Go oracle shared library unavailable at: {$libPath}.");
   }
  }

  $cli = __DIR__ . '/oracle_cli.php';
  if (!file_exists($cli)) {
   throw new RuntimeException("Oracle CLI runner not found at: {$cli}");
  }

  self::$oracleCliPath = $cli;
  return self::$oracleCliPath;
 }

 /**
  * @param array<string, mixed> $options
  */
 public static function renderOracle(string $content, array $options = []): string
 {
  $cli = self::getOracleCliPath();
  $payload = json_encode(['content' => $content, 'options' => $options], JSON_THROW_ON_ERROR);

  $descriptors = [
   0 => ['pipe', 'r'],
   1 => ['pipe', 'w'],
   2 => ['pipe', 'w'],
  ];

  $process = proc_open(escapeshellcmd(PHP_BINARY) . ' ' . escapeshellarg($cli), $descriptors, $pipes);
  if (!is_resource($process)) {
   throw new RuntimeException('Failed to launch oracle CLI process');
  }

  fwrite($pipes[0], $payload);
  fclose($pipes[0]);

  $stdout = stream_get_contents($pipes[1]);
  fclose($pipes[1]);

  $stderr = stream_get_contents($pipes[2]);
  fclose($pipes[2]);

  $exitCode = proc_close($process);
  if ($exitCode !== 0) {
   throw new RuntimeException("Oracle process failed with exit code {$exitCode}: {$stderr}");
  }

  $result = json_decode((string) $stdout, true, 512, JSON_THROW_ON_ERROR);
  if (isset($result['error'])) {
   throw new RuntimeException("Oracle error: " . $result['error']);
  }

  return (string) ($result['output'] ?? '');
 }

 /**
  * Compare VT cell grids produced by oracle and native renderer.
  */
 public static function assertCellGridEquals(
  string $expectedAnsi,
  string $actualAnsi,
  int $width,
  string $message = '',
 ): void {
  $expectedLines = explode("\n", $expectedAnsi);
  $actualLines = explode("\n", $actualAnsi);
  $height = max(count($expectedLines), count($actualLines), 1) + 2;

  $expBuf = new ScreenBuffer($width, $height);
  $expBuf->write($expectedAnsi);

  $actBuf = new ScreenBuffer($width, $height);
  $actBuf->write($actualAnsi);

  $expText = $expBuf->getScreen();
  $actText = $actBuf->getScreen();

  if ($expText !== $actText) {
   $diffMsg = sprintf(
    "%s\n--- Plain Text Screen Mismatch (Width: %d) ---\nExpected:\n%s\nActual:\n%s\n",
    $message,
    $width,
    $expText,
    $actText
   );
   Assert::assertSame($expText, $actText, $diffMsg);
  }

  $expCells = $expBuf->getCells();
  $actCells = $actBuf->getCells();

  $maxRows = max(count($expCells), count($actCells));
  for ($r = 0; $r < $maxRows; $r++) {
   $expRow = $expCells[$r] ?? [];
   $actRow = $actCells[$r] ?? [];

   for ($c = 0; $c < $width; $c++) {
    $expCell = $expRow[$c] ?? ['char' => ' ', 'style' => ''];
    $actCell = $actRow[$c] ?? ['char' => ' ', 'style' => ''];

    if ($expCell['char'] !== $actCell['char']) {
     Assert::assertSame(
      $expCell['char'],
      $actCell['char'],
      sprintf(
       "%s\nCharacter mismatch at [row=%d, col=%d]. Expected '%s', got '%s'",
       $message,
       $r,
       $c,
       $expCell['char'],
       $actCell['char']
      )
     );
    }

    // Check normalized styles
    $expStyle = self::normalizeSgr($expCell['style']);
    $actStyle = self::normalizeSgr($actCell['style']);

    if (!self::stylesMatch($expStyle, $actStyle)) {
     Assert::assertSame(
      $expStyle,
      $actStyle,
      sprintf(
       "%s\nStyle mismatch at [row=%d, col=%d] for char '%s'. Expected SGR '%s', got '%s'",
       $message,
       $r,
       $c,
       $actCell['char'],
       addcslashes($expStyle, "\0..\37"),
       addcslashes($actStyle, "\0..\37")
      )
     );
    }
   }
  }
  Assert::assertTrue(true);
 }

 private static function normalizeSgr(string $sgr): string
 {
  if ($sgr === '') {
   return '';
  }
  // If it's a reset code or empty, return empty
  if ($sgr === "\x1b[m" || $sgr === "\x1b[0m") {
   return '';
  }
  return $sgr;
 }

 private static function stylesMatch(string $s1, string $s2): bool
 {
  if ($s1 === $s2) {
   return true;
  }

  // Attributes check
  $hasBold1 = str_contains($s1, '[1;') || str_contains($s1, '[1m');
  $hasBold2 = str_contains($s2, '[1;') || str_contains($s2, '[1m');
  if ($hasBold1 !== $hasBold2) {
   return false;
  }

  $extract = function (string $s): array {
   $fg = null;
   $bg = null;
   if (preg_match('/38;2;(\d+);(\d+);(\d+)/', $s, $m)) {
    $fg = [(int)$m[1], (int)$m[2], (int)$m[3]];
   }
   if (preg_match('/48;2;(\d+);(\d+);(\d+)/', $s, $m)) {
    $bg = [(int)$m[1], (int)$m[2], (int)$m[3]];
   }
   return [$fg, $bg];
  };

  [$fg1, $bg1] = $extract($s1);
  [$fg2, $bg2] = $extract($s2);

  if (($fg1 === null) !== ($fg2 === null)) {
   return false;
  }
  if (($bg1 === null) !== ($bg2 === null)) {
   return false;
  }

  if ($fg1 !== null && $fg2 !== null) {
   if (abs($fg1[0] - $fg2[0]) > 10 || abs($fg1[1] - $fg2[1]) > 10 || abs($fg1[2] - $fg2[2]) > 10) {
    return false;
   }
  }

  if ($bg1 !== null && $bg2 !== null) {
   if (abs($bg1[0] - $bg2[0]) > 10 || abs($bg1[1] - $bg2[1]) > 10 || abs($bg1[2] - $bg2[2]) > 10) {
    return false;
   }
  }

  return true;
 }
}
