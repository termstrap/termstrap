<?php

declare(strict_types=1);

namespace Termstrap\Image;

use GdImage;

final class SixelRenderer
{
    public static function render(
        GdImage $img,
        int $width,
        int $cellWidthPx = 8,
    ): string {
        $origW = imagesx($img);
        $origH = imagesy($img);

        $cellW = $cellWidthPx > 0 ? $cellWidthPx : 8;
        $targetPxW = $width * $cellW;

        $scaled = $img;
        if ($origW > $targetPxW) {
            $ratio = $targetPxW / $origW;
            $targetPxH = (int) max(1, round($origH * $ratio));
            $scaled = imagescale($img, $targetPxW, $targetPxH, IMG_BICUBIC);
            imagealphablending($scaled, false);
            imagesavealpha($scaled, true);
        }

        $w = imagesx($scaled);
        $h = imagesy($scaled);

        // Collect unique colors for palette (up to 256 colors)
        $palette = [];
        $paletteIndex = 0;
        $colorMap = [];

        for ($y = 0; $y < $h; $y++) {
            for ($x = 0; $x < $w; $x++) {
                $rgba = imagecolorat($scaled, $x, $y);
                $a = ($rgba >> 24) & 0x7F;
                if ($a >= 64) {
                    continue; // transparent
                }
                $r = ($rgba >> 16) & 0xFF;
                $g = ($rgba >> 8) & 0xFF;
                $b = $rgba & 0xFF;
                // Quantize to 5-bit channels for palette compression
                $key = sprintf('%d,%d,%d', $r >> 3, $g >> 3, $b >> 3);
                if (!isset($colorMap[$key]) && $paletteIndex < 256) {
                    $colorMap[$key] = $paletteIndex;
                    $palette[$paletteIndex] = [$r, $g, $b];
                    $paletteIndex++;
                }
            }
        }

        $out = "\x1bPq\"1;1";
        // Emit palette definitions: #P;2;R;G;B (0-100 scale)
        foreach ($palette as $idx => [$r, $g, $b]) {
            $r100 = (int) round(($r / 255.0) * 100);
            $g100 = (int) round(($g / 255.0) * 100);
            $b100 = (int) round(($b / 255.0) * 100);
            $out .= sprintf('#%d;2;%d;%d;%d', $idx, $r100, $g100, $b100);
        }

        // Process in 6-pixel horizontal bands
        for ($bandY = 0; $bandY < $h; $bandY += 6) {
            foreach ($palette as $pIdx => $c) {
                $hasColorInBand = false;
                $bandMasks = [];

                for ($x = 0; $x < $w; $x++) {
                    $mask = 0;
                    for ($dy = 0; $dy < 6; $dy++) {
                        $y = $bandY + $dy;
                        if ($y >= $h) {
                            break;
                        }
                        $rgba = imagecolorat($scaled, $x, $y);
                        $a = ($rgba >> 24) & 0x7F;
                        if ($a >= 64) {
                            continue;
                        }
                        $r = ($rgba >> 16) & 0xFF;
                        $g = ($rgba >> 8) & 0xFF;
                        $b = $rgba & 0xFF;
                        $key = sprintf('%d,%d,%d', $r >> 3, $g >> 3, $b >> 3);
                        if (($colorMap[$key] ?? -1) === $pIdx) {
                            $mask |= (1 << $dy);
                        }
                    }
                    $bandMasks[$x] = $mask;
                    if ($mask > 0) {
                        $hasColorInBand = true;
                    }
                }

                if ($hasColorInBand) {
                    $out .= sprintf('#%d', $pIdx);
                    // RLE compression of identical sixel characters
                    $lastChar = -1;
                    $repeat = 0;

                    for ($x = 0; $x < $w; $x++) {
                        $ch = chr(63 + $bandMasks[$x]);
                        if ($ch === $lastChar) {
                            $repeat++;
                        } else {
                            if ($repeat > 3) {
                                $out .= sprintf('!%d%s', $repeat, $lastChar);
                            } elseif ($repeat > 0) {
                                $out .= str_repeat($lastChar, $repeat);
                            }
                            $lastChar = $ch;
                            $repeat = 1;
                        }
                    }
                    if ($repeat > 3) {
                        $out .= sprintf('!%d%s', $repeat, $lastChar);
                    } elseif ($repeat > 0) {
                        $out .= str_repeat($lastChar, $repeat);
                    }
                    $out .= '$'; // Return to start of line
                }
            }
            $out .= '-'; // Line feed (next sixel band)
        }

        $out .= "\x1b\\\n";
        return $out;
    }
}
