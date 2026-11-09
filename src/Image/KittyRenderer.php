<?php

declare(strict_types=1);

namespace Termstrap\Image;

use GdImage;

final class KittyRenderer
{
    private const CHUNK_SIZE = 4096;

    public static function render(
        GdImage $img,
        int $width,
        int $cols = 0,
        int $rows = 0,
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

        ob_start();
        imagepng($scaled);
        $pngBytes = (string) ob_get_clean();

        $payload = base64_encode($pngBytes);
        $w = imagesx($scaled);
        $h = imagesy($scaled);

        return self::buildSequence($payload, $w, $h, $cols, $rows);
    }

    private static function buildSequence(string $payload, int $w, int $h, int $cols, int $rows): string
    {
        $chunks = str_split($payload, self::CHUNK_SIZE) ?: [];
        $total = count($chunks);
        $out = '';

        foreach ($chunks as $i => $chunk) {
            $out .= "\x1b_G";
            if ($i === 0) {
                $out .= sprintf('a=T,f=100,q=2,s=%d,v=%d', $w, $h);
                if ($cols > 0) {
                    $out .= sprintf(',c=%d', $cols);
                }
                if ($rows > 0) {
                    $out .= sprintf(',r=%d', $rows);
                }
            }
            if ($total === 1) {
                $out .= ',m=0;' . $chunk;
            } elseif ($i < $total - 1) {
                if ($i === 0) {
                    $out .= ',m=1;' . $chunk;
                } else {
                    $out .= 'm=1;' . $chunk;
                }
            } else {
                $out .= 'm=0;' . $chunk;
            }
            $out .= "\x1b\\";
        }

        return $out;
    }
}
