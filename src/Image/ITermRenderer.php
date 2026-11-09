<?php

declare(strict_types=1);

namespace Termstrap\Image;

use GdImage;

final class ITermRenderer
{
    public static function render(
        GdImage $img,
        int $width,
        int $height = 0,
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
        $size = strlen($pngBytes);

        if ($height > 0) {
            return sprintf(
                "\x1b]1337;File=inline=1;size=%d;width=%d;height=%d;preserveAspectRatio=1:%s\x07\n",
                $size,
                $width,
                $height,
                $payload
            );
        }

        return sprintf(
            "\x1b]1337;File=inline=1;size=%d;width=%d;preserveAspectRatio=1:%s\x07\n",
            $size,
            $width,
            $payload
        );
    }
}
