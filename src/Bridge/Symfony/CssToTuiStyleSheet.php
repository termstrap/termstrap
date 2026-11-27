<?php

declare(strict_types=1);

namespace Termstrap\Bridge\Symfony;

use Symfony\Component\Tui\Style\Border;
use Symfony\Component\Tui\Style\Color;
use Symfony\Component\Tui\Style\Direction;
use Symfony\Component\Tui\Style\Padding;
use Symfony\Component\Tui\Style\Style;
use Symfony\Component\Tui\Style\StyleSheet as TuiStyleSheet;
use Termstrap\Css\CssMatcher;
use Termstrap\Css\CssRule;
use Termstrap\Css\MediaClause;

/**
 * Bridges CSS stylesheets into Symfony Tui StyleSheet value objects.
 */
final class CssToTuiStyleSheet
{
    /**
     * Parse CSS text and compile into a Symfony Tui StyleSheet.
     */
    public static function fromCss(string $css, ?string $rootPath = null): TuiStyleSheet
    {
        $matcher = CssMatcher::create('none', [$css], $rootPath);
        $tuiStyleSheet = new TuiStyleSheet();

        foreach ($matcher->rules as $rule) {
            if ($rule->origin !== CssRule::ORIGIN_CUSTOM) {
                continue;
            }

            $selector = self::normalizeSelector($rule->selector, $rule->pseudo);
            if ($selector === '') {
                continue;
            }

            $style = self::createStyleFromDecls($rule->decls);
            if ($style === null) {
                continue;
            }

            $minWidth = self::extractMinWidth($rule);
            if ($minWidth > 0) {
                $tuiStyleSheet->addBreakpoint($minWidth, $selector, $style);
            } else {
                $tuiStyleSheet->addRule($selector, $style);
            }
        }

        return $tuiStyleSheet;
    }

    /**
     * Load CSS from a file and compile into a Symfony Tui StyleSheet.
     */
    public static function fromFile(string $path): TuiStyleSheet
    {
        if (!is_file($path)) {
            throw new \InvalidArgumentException(sprintf('CSS file "%s" not found.', $path));
        }

        $content = (string) file_get_contents($path);
        return self::fromCss($content, dirname($path));
    }

    private static function normalizeSelector(string $base, string $pseudo): string
    {
        $base = trim($base);
        if ($base === '') {
            return '';
        }

        // Handle :root pseudo-class
        if ($base === ':root' || $base === 'html' || $base === 'body') {
            return ':root';
        }

        // Pseudo class / element (e.g. :focus, ::spinner)
        if ($pseudo !== '') {
            $prefix = str_starts_with($pseudo, ':') ? $pseudo : ':' . $pseudo;
            return $base . $prefix;
        }

        return $base;
    }

    /**
     * @param list<array{property: string, value: string, important: bool}> $decls
     */
    private static function createStyleFromDecls(array $decls): ?Style
    {
        $color = null;
        $background = null;
        $bold = null;
        $italic = null;
        $underline = null;
        $strikethrough = null;
        $faint = null;
        $hidden = null;
        $direction = null;
        $gap = null;
        $flex = null;

        $paddingTop = 0;
        $paddingRight = 0;
        $paddingBottom = 0;
        $paddingLeft = 0;
        $hasPadding = false;

        $hasBorder = false;
        $borderStyle = 'solid';
        $borderColor = null;

        foreach ($decls as $decl) {
            $prop = strtolower($decl['property']);
            $val = trim($decl['value']);

            switch ($prop) {
                case 'color':
                    $color = $val;
                    break;

                case 'background-color':
                case 'background':
                    if (!str_contains($val, 'url(') && !str_contains($val, 'gradient')) {
                        $background = $val;
                    }
                    break;

                case 'font-weight':
                    $bold = in_array(strtolower($val), ['bold', '700', '800', '900', 'bolder'], true);
                    break;

                case 'font-style':
                    $italic = strtolower($val) === 'italic';
                    break;

                case 'text-decoration':
                case 'text-decoration-line':
                    $lowerVal = strtolower($val);
                    if (str_contains($lowerVal, 'underline')) {
                        $underline = true;
                    }
                    if (str_contains($lowerVal, 'line-through')) {
                        $strikethrough = true;
                    }
                    if ($lowerVal === 'none') {
                        $underline = false;
                        $strikethrough = false;
                    }
                    break;

                case 'opacity':
                    if (is_numeric($val) && (float) $val < 0.8) {
                        $faint = true;
                    }
                    break;

                case 'visibility':
                    if (strtolower($val) === 'hidden') {
                        $hidden = true;
                    } elseif (strtolower($val) === 'visible') {
                        $hidden = false;
                    }
                    break;

                case 'display':
                    if (strtolower($val) === 'none') {
                        $hidden = true;
                    }
                    break;

                case 'flex-direction':
                    $direction = strtolower($val) === 'row' ? Direction::Horizontal : Direction::Vertical;
                    break;

                case 'gap':
                    $gap = (int) $val;
                    break;

                case 'flex':
                case 'flex-grow':
                    $flex = (int) $val;
                    break;

                case 'padding':
                    $parts = preg_split('/\s+/', $val);
                    if ($parts && count($parts) === 1) {
                        $v = (int) $parts[0];
                        $paddingTop = $paddingRight = $paddingBottom = $paddingLeft = $v;
                    } elseif ($parts && count($parts) === 2) {
                        $paddingTop = $paddingBottom = (int) $parts[0];
                        $paddingRight = $paddingLeft = (int) $parts[1];
                    } elseif ($parts && count($parts) === 4) {
                        $paddingTop = (int) $parts[0];
                        $paddingRight = (int) $parts[1];
                        $paddingBottom = (int) $parts[2];
                        $paddingLeft = (int) $parts[3];
                    }
                    $hasPadding = true;
                    break;

                case 'padding-top':
                    $paddingTop = (int) $val;
                    $hasPadding = true;
                    break;
                case 'padding-right':
                    $paddingRight = (int) $val;
                    $hasPadding = true;
                    break;
                case 'padding-bottom':
                    $paddingBottom = (int) $val;
                    $hasPadding = true;
                    break;
                case 'padding-left':
                    $paddingLeft = (int) $val;
                    $hasPadding = true;
                    break;

                case 'border':
                    $hasBorder = true;
                    $parts = preg_split('/\s+/', $val);
                    if ($parts) {
                        foreach ($parts as $p) {
                            $lp = strtolower($p);
                            if (in_array($lp, ['solid', 'rounded', 'double', 'dashed', 'dotted', 'heavy'], true)) {
                                $borderStyle = $lp;
                            } elseif (!is_numeric($p) && !str_ends_with($p, 'px')) {
                                $borderColor = $p;
                            }
                        }
                    }
                    break;

                case 'border-color':
                    $borderColor = $val;
                    $hasBorder = true;
                    break;

                case 'border-style':
                    $borderStyle = strtolower($val);
                    $hasBorder = true;
                    break;

                case 'border-radius':
                    if ((int) $val > 0) {
                        $borderStyle = 'rounded';
                        $hasBorder = true;
                    }
                    break;
            }
        }

        $style = new Style();

        if ($color !== null) {
            $style = $style->withColor($color);
        }
        if ($background !== null) {
            $style = $style->withBackground($background);
        }
        if ($bold !== null) {
            $style = $style->withBold($bold);
        }
        if ($italic !== null) {
            $style = $style->withItalic($italic);
        }
        if ($underline !== null) {
            $style = $style->withUnderline($underline);
        }
        if ($strikethrough !== null) {
            $style = $style->withStrikethrough($strikethrough);
        }
        if ($faint !== null) {
            $style = $style->withFaint($faint);
        }
        if ($hidden !== null) {
            $style = $style->withHidden($hidden);
        }
        if ($direction !== null) {
            $style = $style->withDirection($direction);
        }
        if ($gap !== null) {
            $style = $style->withGap($gap);
        }
        if ($flex !== null) {
            $style = $style->withFlex($flex);
        }
        if ($hasPadding) {
            $style = $style->withPadding(new Padding($paddingTop, $paddingRight, $paddingBottom, $paddingLeft));
        }
        if ($hasBorder) {
            $pattern = match ($borderStyle) {
                'rounded' => 'rounded',
                'double' => 'double',
                'dashed' => 'dashed',
                'dotted' => 'dotted',
                'heavy' => 'heavy',
                default => 'single',
            };
            $style = $style->withBorder(Border::all(1, $pattern, $borderColor));
        }

        return $style;
    }

    private static function extractMinWidth(CssRule $rule): int
    {
        $cond = $rule->mediaCond;
        if ($cond === null) {
            return 0;
        }

        foreach ($cond->clauses as $clause) {
            if ($clause instanceof MediaClause && $clause->minWidth > 0) {
                return $clause->minWidth;
            }
        }

        return 0;
    }
}
