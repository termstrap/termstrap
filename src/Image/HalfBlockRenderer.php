<?php

declare(strict_types=1);

namespace Termstrap\Image;

use GdImage;
use RuntimeException;

final class HalfBlockRenderer
{
    public static function render(
        GdImage $img,
        int $width,
        ColorMode $colorMode = ColorMode::TrueColor,
        bool $optimize = true,
    ): string {
        $origW = imagesx($img);
        $origH = imagesy($img);

        if ($origW <= 0 || $origH <= 0 || $width <= 0) {
            throw new RuntimeException("Invalid image dimensions: {$origW}x{$origH}");
        }

        $scaledW = $width;
        $scaledH = (int) round(((float) $origH * $width) / (float) $origW);
        if ($scaledH < 2) {
            $scaledH = 2;
        }
        if ($scaledH % 2 !== 0) {
            $scaledH++;
        }

        if ($scaledW === $origW && $scaledH === $origH) {
            $scaled = $img;
        } else {
            $scaled = @imagescale($img, $scaledW, $scaledH, IMG_BICUBIC);
            if ($scaled === false) {
                $scaled = imagecreatetruecolor($scaledW, $scaledH);
                imagealphablending($scaled, false);
                imagesavealpha($scaled, true);
                imagecopyresampled($scaled, $img, 0, 0, 0, 0, $scaledW, $scaledH, $origW, $origH);
            }
        }
        imagealphablending($scaled, false);
        imagesavealpha($scaled, true);

        $out = '';
        for ($y = 0; $y < $scaledH; $y += 2) {
            $lastFG = '';
            $lastBG = '';
            $hasFG = false;
            $hasBG = false;

            for ($x = 0; $x < $scaledW; $x++) {
                $rgba1 = imagecolorat($scaled, $x, $y);
                $topAlpha = ($rgba1 >> 24) & 0x7F;
                $topR = ($rgba1 >> 16) & 0xFF;
                $topG = ($rgba1 >> 8) & 0xFF;
                $topB = $rgba1 & 0xFF;

                $btmAlpha = 127;
                $btmR = 0;
                $btmG = 0;
                $btmB = 0;
                if ($y + 1 < $scaledH) {
                    $rgba2 = imagecolorat($scaled, $x, $y + 1);
                    $btmAlpha = ($rgba2 >> 24) & 0x7F;
                    $btmR = ($rgba2 >> 16) & 0xFF;
                    $btmG = ($rgba2 >> 8) & 0xFF;
                    $btmB = $rgba2 & 0xFF;
                }

                $topSolid = $topAlpha < 64;
                $btmSolid = $btmAlpha < 64;

                switch (true) {
                    case $topSolid && $btmSolid:
                        $fgCode = self::fgEscape($topR, $topG, $topB, $colorMode);
                        $bgCode = self::bgEscape($btmR, $btmG, $btmB, $colorMode);

                        if ($optimize) {
                            if (!$hasFG || $fgCode !== $lastFG) {
                                $out .= $fgCode;
                                $lastFG = $fgCode;
                                $hasFG = true;
                            }
                            if (!$hasBG || $bgCode !== $lastBG) {
                                $out .= $bgCode;
                                $lastBG = $bgCode;
                                $hasBG = true;
                            }
                        } else {
                            $out .= $fgCode . $bgCode;
                        }
                        $out .= '▀';
                        break;

                    case $topSolid && !$btmSolid:
                        $fgCode = self::fgEscape($topR, $topG, $topB, $colorMode);
                        if ($optimize) {
                            if (!$hasFG || $fgCode !== $lastFG) {
                                $out .= $fgCode;
                                $lastFG = $fgCode;
                                $hasFG = true;
                            }
                            if ($hasBG) {
                                $out .= "\x1b[49m";
                                $hasBG = false;
                                $lastBG = '';
                            }
                        } else {
                            $out .= $fgCode . "\x1b[49m";
                        }
                        $out .= '▀';
                        break;

                    case !$topSolid && $btmSolid:
                        $fgCode = self::fgEscape($btmR, $btmG, $btmB, $colorMode);
                        if ($optimize) {
                            if (!$hasFG || $fgCode !== $lastFG) {
                                $out .= $fgCode;
                                $lastFG = $fgCode;
                                $hasFG = true;
                            }
                            if ($hasBG) {
                                $out .= "\x1b[49m";
                                $hasBG = false;
                                $lastBG = '';
                            }
                        } else {
                            $out .= $fgCode . "\x1b[49m";
                        }
                        $out .= '▄';
                        break;

                    default: // Both transparent
                        if ($optimize && $hasBG) {
                            $out .= "\x1b[49m";
                            $hasBG = false;
                            $lastBG = '';
                        }
                        $out .= ' ';
                        break;
                }
            }

            $out .= "\x1b[0m";
            if ($y + 2 < $scaledH) {
                $out .= "\n";
            }
        }

        return $out;
    }

    private static function fgEscape(int $r, int $g, int $b, ColorMode $mode): string
    {
        return match ($mode) {
            ColorMode::Ansi256 => sprintf("\x1b[38;5;%dm", self::rgbToAnsi256($r, $g, $b)),
            ColorMode::Ansi16 => sprintf("\x1b[%dm", self::rgbToAnsi16($r, $g, $b, false)),
            default => sprintf("\x1b[38;2;%d;%d;%dm", $r, $g, $b),
        };
    }

    private static function bgEscape(int $r, int $g, int $b, ColorMode $mode): string
    {
        return match ($mode) {
            ColorMode::Ansi256 => sprintf("\x1b[48;5;%dm", self::rgbToAnsi256($r, $g, $b)),
            ColorMode::Ansi16 => sprintf("\x1b[%dm", self::rgbToAnsi16($r, $g, $b, true)),
            default => sprintf("\x1b[48;2;%d;%d;%dm", $r, $g, $b),
        };
    }

    public static function rgbToAnsi256(int $r, int $g, int $b): int
    {
        // 6x6x6 color cube: 16 + 36*r + 6*g + b where r,g,b in 0..5
        $rIdx = (int) round($r / 255.0 * 5);
        $gIdx = (int) round($g / 255.0 * 5);
        $bIdx = (int) round($b / 255.0 * 5);
        return 16 + (36 * $rIdx) + (6 * $gIdx) + $bIdx;
    }

    public static function rgbToAnsi16(int $r, int $g, int $b, bool $isBg): int
    {
        $bright = ($r > 128 || $g > 128 || $b > 128);
        $code = 0;
        if ($r >= 64) $code |= 1;
        if ($g >= 64) $code |= 2;
        if ($b >= 64) $code |= 4;

        if ($isBg) {
            return ($bright ? 100 : 40) + $code;
        }
        return ($bright ? 90 : 30) + $code;
    }
}
