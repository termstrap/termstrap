<?php

declare(strict_types=1);

namespace Termstrap\Image;

use GdImage;
use RuntimeException;
use Throwable;

final class ImageLoader
{
 /**
  * @return array{image: GdImage, isAnimated: bool, frameCount: int}
  */
 public static function load(string $src, ?string $rootPath = null): array
 {
  $src = trim($src);
  if ($src === '') {
   throw new RuntimeException('Empty image source');
  }

  $data = null;
  if (str_starts_with($src, 'http://') || str_starts_with($src, 'https://')) {
   $ctx = stream_context_create([
    'http' => [
     'timeout' => 10,
     'user_agent' => 'termstrap/1.0',
    ],
   ]);
   $data = @file_get_contents($src, false, $ctx);
   if ($data === false) {
    throw new RuntimeException("Failed to fetch image: {$src}");
   }
  } else {
   $path = $src;
   if ($rootPath !== null && $rootPath !== '' && !str_starts_with($path, '/')) {
    $path = rtrim($rootPath, '/') . '/' . $path;
   }
   if (!file_exists($path)) {
    throw new RuntimeException("Image file not found: {$path}");
   }
   $data = @file_get_contents($path);
   if ($data === false) {
    throw new RuntimeException("Failed to read image file: {$path}");
   }
  }

  if (str_contains(strtolower(substr($data, 0, 256)), '<svg') || str_ends_with(strtolower($src), '.svg')) {
   throw new RuntimeException("SVG images are not supported: {$src}");
  }

  // Check if animated GIF
  $isAnimated = false;
  $frameCount = 1;
  if (str_starts_with($data, 'GIF87a') || str_starts_with($data, 'GIF89a')) {
   // Count Graphics Control Extension blocks
   $gceCount = substr_count($data, "\x00\x21\xF9\x04");
   if ($gceCount > 1) {
    $isAnimated = true;
    $frameCount = $gceCount;
   }
  }

  try {
   $im = @imagecreatefromstring($data);
   if ($im === false) {
    throw new RuntimeException("Failed to decode image from {$src}");
   }
   imagealphablending($im, false);
   imagesavealpha($im, true);

   return [
    'image' => $im,
    'isAnimated' => $isAnimated,
    'frameCount' => $frameCount,
   ];
  } catch (Throwable $e) {
   throw new RuntimeException("Failed to decode image: " . $e->getMessage(), 0, $e);
  }
 }
}
