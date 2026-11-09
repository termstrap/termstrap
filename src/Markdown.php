<?php

declare(strict_types=1);

namespace Termstrap;

use League\CommonMark\GithubFlavoredMarkdownConverter;
use Throwable;

final class Markdown
{
    private static ?GithubFlavoredMarkdownConverter $converter = null;

    public static function toHtml(string $markdown): string
    {
        if ($markdown === '') {
            return '';
        }

        if (self::$converter === null) {
            self::$converter = new GithubFlavoredMarkdownConverter([
                'html_input' => 'allow',
                'allow_unsafe_links' => true,
            ]);
        }

        try {
            return (string) self::$converter->convert($markdown);
        } catch (Throwable) {
            return $markdown;
        }
    }
}
