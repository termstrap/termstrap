<?php

declare(strict_types=1);

namespace Termstrap\Layout;

use Dom\Element;
use Dom\Node;
use Dom\Text;
use Termstrap\Css\Axis;
use Termstrap\Css\ComputedStyle;
use Termstrap\Css\CssMatcher;
use Termstrap\Css\MediaCondition;
use Termstrap\Css\UnitContext;
use Termstrap\Css\Units;
use Termstrap\Util;

final class LayoutEngine
{
    public static function buildRenderTree(
        Node $node,
        CssMatcher $matcher,
        int $termWidth,
        ?ComputedStyle $parentStyle = null,
    ): ?RenderNode {
        if ($node instanceof Text) {
            $raw = $node->data;
            $trimmed = trim($raw);
            if ($trimmed === '') {
                if (strlen($raw) > 0) {
                    $st = new ComputedStyle();
                    $st->display = 'inline';
                    $r = new RenderNode(NodeType::Text, $st);
                    $r->text = ' ';
                    return $r;
                }
                return null;
            }

            $words = preg_split('/\s+/', $raw, -1, PREG_SPLIT_NO_EMPTY) ?: [];
            $hasLeading = ctype_space($raw[0]);
            $hasTrailing = ctype_space($raw[strlen($raw) - 1]);

            $text = ($hasLeading ? ' ' : '') . implode(' ', $words) . ($hasTrailing ? ' ' : '');
            if ($parentStyle !== null && $parentStyle->textTransform !== '') {
                $text = self::applyTextTransform($text, $parentStyle->textTransform);
            }

            $st = new ComputedStyle();
            $st->display = 'inline';
            if ($parentStyle !== null) {
                $st->fgColor = $parentStyle->fgColor;
                $st->fgColorSet = $parentStyle->fgColorSet;
                $st->bold = $parentStyle->bold;
                $st->boldSet = $parentStyle->boldSet;
                $st->italic = $parentStyle->italic;
                $st->italicSet = $parentStyle->italicSet;
                $st->underline = $parentStyle->underline;
                $st->underlineSet = $parentStyle->underlineSet;
                $st->underlineStyle = $parentStyle->underlineStyle;
                $st->underlineStyleSet = $parentStyle->underlineStyleSet;
                $st->underlineColor = $parentStyle->underlineColor;
                $st->underlineColorSet = $parentStyle->underlineColorSet;
                $st->strikethrough = $parentStyle->strikethrough;
                $st->strikethroughSet = $parentStyle->strikethroughSet;
                $st->faint = $parentStyle->faint;
                $st->faintSet = $parentStyle->faintSet;
                $st->blink = $parentStyle->blink;
                $st->blinkSet = $parentStyle->blinkSet;
                $st->shimmer = $parentStyle->shimmer;
                $st->shimmerSet = $parentStyle->shimmerSet;
                $st->shimmerHighlight = $parentStyle->shimmerHighlight;
                $st->textTransform = $parentStyle->textTransform;
                $st->textTransformSet = $parentStyle->textTransformSet;
            }

            $rNode = new RenderNode(NodeType::Text, $st);
            $rNode->text = $text;
            return $rNode;
        }

        if (!$node instanceof Element) {
            return null;
        }

        $tag = strtolower($node->tagName);
        if (in_array($tag, ['head', 'style', 'script', 'meta', 'title'], true)) {
            return null;
        }

        $style = $matcher->computeStyleForNode($node, $termWidth, $parentStyle);
        if ($style->display === 'none') {
            return null;
        }

        if ($tag === 'tr') {
            $style->display = 'flex';
        }
        if ($tag === 'th' || $tag === 'td') {
            $style->colAuto = true;
        }
        if ($tag === 'br') {
            $style->display = 'inline';
        }

        $rNode = new RenderNode(NodeType::Element, $style);
        $rNode->tag = $tag;

        if ($node->hasAttribute('href')) {
            $rNode->href = $node->getAttribute('href');
        }

        if ($tag === 'ul' || $tag === 'ol') {
            $classAttr = $node->getAttribute('class') ?? '';
            $role = $node->getAttribute('role') ?? '';
            if ($role === 'tree') {
                $rNode->isTree = true;
            } else {
                $classes = preg_split('/\s+/', trim($classAttr), -1, PREG_SPLIT_NO_EMPTY) ?: [];
                if (in_array('tree', $classes, true)) {
                    $rNode->isTree = true;
                }
            }
        }

        if ($tag === 'picture') {
            $rNode->type = NodeType::Image;
            self::resolvePictureNode($node, $rNode, $matcher, $termWidth);
            return $rNode;
        }

        if ($tag === 'img') {
            $rNode->type = NodeType::Image;
            $rNode->src = $node->getAttribute('src') ?? '';
            if ($rNode->src === '' && $node->hasAttribute('srcset')) {
                $rNode->src = self::parseSrcset($node->getAttribute('srcset'));
            }
            $rNode->alt = $node->getAttribute('alt') ?? '';
            if ($rNode->style->width <= 0 && $node->hasAttribute('width')) {
                $w = $node->getAttribute('width');
                if ($w !== '') {
                    $ctx = UnitContext::default($termWidth);
                    $rNode->style->width = Units::resolveLength($w, Axis::X, $ctx);
                }
            }
            return $rNode;
        }

        if ($tag === 'pre') {
            $rNode->text = rtrim($node->textContent, "\n");
            $preClass = $node->getAttribute('class') ?? '';
            $codeClass = '';
            $codeNodes = $node->getElementsByTagName('code');
            if ($codeNodes->length > 0) {
                $codeClass = $codeNodes->item(0)->getAttribute('class') ?? '';
            }
            $rNode->lang = self::resolveLang($codeClass, $preClass);
            return $rNode;
        }

        if ($tag === 'ascii') {
            $rNode->text = $node->textContent;
            if ($node->hasAttribute('font')) {
                $rNode->font = $node->getAttribute('font');
            }
            if ($node->hasAttribute('fit')) {
                $rNode->fit = strtolower(trim($node->getAttribute('fit')));
            }
            return $rNode;
        }

        // ::before pseudo-element
        $before = $matcher->pseudoText($node, $termWidth, $style, 'before');
        if ($before !== null) {
            [$beforeText, $beforeStyle] = $before;
            $beforeText = self::applyTextTransform($beforeText, $beforeStyle->textTransform);
            $beforeNode = new RenderNode(NodeType::Text, $beforeStyle);
            $beforeNode->text = $beforeText;
            $rNode->children[] = $beforeNode;
        }

        // Recursively build children
        for ($c = $node->firstChild; $c !== null; $c = $c->nextSibling) {
            $childRNode = self::buildRenderTree($c, $matcher, $termWidth, $style);
            if ($childRNode !== null) {
                if ($rNode->style->display === 'flex' && $childRNode->type === NodeType::Text && trim($childRNode->text) === '') {
                    continue;
                }
                $rNode->children[] = $childRNode;
            }
        }

        // ::after pseudo-element
        $after = $matcher->pseudoText($node, $termWidth, $style, 'after');
        if ($after !== null) {
            [$afterText, $afterStyle] = $after;
            $afterText = self::applyTextTransform($afterText, $afterStyle->textTransform);
            $afterNode = new RenderNode(NodeType::Text, $afterStyle);
            $afterNode->text = $afterText;
            $rNode->children[] = $afterNode;
        }

        return $rNode;
    }

    public static function computeLayout(RenderNode $root, int $availableWidth): void
    {
        if ($availableWidth <= 0) {
            return;
        }
        self::computeWidths($root, $availableWidth);
    }

    private static function computeWidths(RenderNode $node, int $parentWidth): void
    {
        // Reduce available width by margins and shadow
        $avail = $parentWidth - $node->style->marginLeft - $node->style->marginRight - $node->style->shadow;
        if ($avail < 1) {
            $avail = 1;
        }

        if ($node->style->width > 0) {
            $explicitW = $node->style->width - $node->style->marginLeft - $node->style->marginRight - $node->style->shadow;
            if ($explicitW > 0 && $explicitW < $avail) {
                $avail = $explicitW;
            }
        }
        if ($node->style->maxWidth > 0) {
            $maxW = $node->style->maxWidth - $node->style->marginLeft - $node->style->marginRight - $node->style->shadow;
            if ($maxW > 0 && $maxW < $avail) {
                $avail = $maxW;
            }
        }

        // For blocks with fit-content/max-content width, shrink-wrap avail to the natural content width
        if ($node->style->width === Units::LENGTH_FIT_CONTENT) {
            $borderExtra = 0;
            if ($node->style->border || $node->style->borderLeft) {
                $borderExtra++;
            }
            if ($node->style->border || $node->style->borderRight) {
                $borderExtra++;
            }
            $paddingExtra = $node->style->paddingLeft + $node->style->paddingRight;
            $naturalW = self::naturalContentWidth($node);
            $needed = $naturalW + $borderExtra + $paddingExtra;
            if ($needed > 0 && $needed < $avail) {
                $avail = $needed;
            }
        }

        $borderH = 0;
        if ($node->style->border || $node->style->borderLeft) {
            $borderH++;
        }
        if ($node->style->border || $node->style->borderRight) {
            $borderH++;
        }

        $lipglossW = $avail > 0 ? $avail : 1;
        $paddingH = $node->style->paddingLeft + $node->style->paddingRight;
        $contentW = $lipglossW - $borderH - $paddingH;
        if ($contentW < 1) {
            $contentW = 1;
        }

        $node->allocatedWidth = $parentWidth;
        $node->boxWidth = $avail;
        $node->lipglossWidth = $lipglossW;
        $node->contentWidth = $contentW;

        if ($node->style->display === 'flex') {
            // Count auto columns
            $autoCount = 0;
            foreach ($node->children as $child) {
                if ($child->style->colSpan === 0 || $child->style->colAuto) {
                    $autoCount++;
                }
            }

            // Resolve column spans for all children
            foreach ($node->children as $child) {
                if ($child->style->colSpan === 0 || $child->style->colAuto) {
                    if ($node->style->rowCols > 0) {
                        $child->style->colSpan = (int) (12 / $node->style->rowCols);
                    } elseif ($autoCount > 0 && $autoCount <= 12) {
                        $child->style->colSpan = (int) (12 / $autoCount);
                    } else {
                        $child->style->colSpan = 12;
                    }
                }
                if ($child->style->colSpan < 1) {
                    $child->style->colSpan = 1;
                }
                if ($child->style->colSpan > 12) {
                    $child->style->colSpan = 12;
                }
            }

            /** @var list<array{children: list<RenderNode>, spans: list<int>, spanSum: int}> $subRows */
            $subRows = [];
            $currentSubRow = ['children' => [], 'spans' => [], 'spanSum' => 0];

            foreach ($node->children as $child) {
                $span = $child->style->colSpan;
                if (!empty($currentSubRow['children']) && $currentSubRow['spanSum'] + $span > 12) {
                    $subRows[] = $currentSubRow;
                    $currentSubRow = ['children' => [], 'spans' => [], 'spanSum' => 0];
                }
                $currentSubRow['children'][] = $child;
                $currentSubRow['spans'][] = $span;
                $currentSubRow['spanSum'] += $span;
            }
            if (!empty($currentSubRow['children'])) {
                $subRows[] = $currentSubRow;
            }

            foreach ($subRows as $row) {
                $targetWidth = (int) (($node->contentWidth * $row['spanSum']) / 12);
                if ($row['spanSum'] >= 12) {
                    $targetWidth = $node->contentWidth;
                }
                $remWidth = $targetWidth;
                $remSpan = $row['spanSum'];

                foreach ($row['children'] as $i => $child) {
                    $span = $row['spans'][$i];
                    $colWidth = (int) (($remWidth * $span) / $remSpan);
                    if ($colWidth < 1) {
                        $colWidth = 1;
                    }
                    $remWidth -= $colWidth;
                    $remSpan -= $span;
                    self::computeWidths($child, $colWidth);
                }
            }
        } elseif ($node->tag === 'ul' || $node->tag === 'ol') {
            $enumWidth = $node->tag === 'ol' ? 4 : 2;
            foreach ($node->children as $child) {
                $w = $node->contentWidth - $enumWidth;
                if ($w < 1) {
                    $w = 1;
                }
                self::computeWidths($child, $w);
            }
        } else {
            foreach ($node->children as $child) {
                self::computeWidths($child, $node->contentWidth);
            }
        }
    }

    public static function naturalContentWidth(RenderNode $node): int
    {
        if ($node->type === NodeType::Text) {
            $lines = explode("\n", $node->text);
            $maxW = 0;
            foreach ($lines as $l) {
                $w = Util::visualWidth($l);
                if ($w > $maxW) {
                    $maxW = $w;
                }
            }
            return $maxW;
        }

        $maxW = 0;
        $currLine = 0;
        foreach ($node->children as $c) {
            $w = self::naturalContentWidth($c);
            if ($c->style->display === 'block') {
                if ($currLine > $maxW) {
                    $maxW = $currLine;
                }
                $currLine = 0;
                if ($w > $maxW) {
                    $maxW = $w;
                }
            } else {
                $currLine += $w;
            }
        }
        if ($currLine > $maxW) {
            $maxW = $currLine;
        }
        return $maxW;
    }

    public static function resolvePictureNode(
        Element $el,
        RenderNode $rNode,
        CssMatcher $matcher,
        int $termWidth,
    ): void {
        $chosenSrc = '';
        $chosenWidth = '';
        $theme = $matcher->theme !== '' ? $matcher->theme : 'bootstrap';

        $sources = $el->getElementsByTagName('source');
        foreach ($sources as $source) {
            $matched = true;
            if ($source->hasAttribute('media')) {
                $mediaAttr = trim($source->getAttribute('media'));
                if ($mediaAttr !== '') {
                    $cond = MediaCondition::parse($mediaAttr);
                    if ($cond !== null && !$cond->matches($termWidth, 24, $theme)) {
                        $matched = false;
                    }
                }
            }

            if ($matched) {
                if ($source->hasAttribute('srcset')) {
                    $chosenSrc = self::parseSrcset($source->getAttribute('srcset'));
                } elseif ($source->hasAttribute('src')) {
                    $chosenSrc = trim($source->getAttribute('src'));
                }

                if ($chosenSrc !== '') {
                    if ($source->hasAttribute('width')) {
                        $chosenWidth = $source->getAttribute('width');
                    }
                    break;
                }
            }
        }

        $imgs = $el->getElementsByTagName('img');
        $img = $imgs->length > 0 ? $imgs->item(0) : null;

        if ($chosenSrc !== '') {
            $rNode->src = $chosenSrc;
            if ($chosenWidth !== '' && $rNode->style->width <= 0) {
                $ctx = UnitContext::default($termWidth);
                $rNode->style->width = Units::resolveLength($chosenWidth, Axis::X, $ctx);
            }
        } elseif ($img !== null) {
            $rNode->src = $img->getAttribute('src') ?? '';
            if ($rNode->src === '' && $img->hasAttribute('srcset')) {
                $rNode->src = self::parseSrcset($img->getAttribute('srcset'));
            }
        }

        if ($img !== null) {
            $rNode->alt = $img->getAttribute('alt') ?? '';
            if ($rNode->style->width <= 0 && $img->hasAttribute('width')) {
                $w = $img->getAttribute('width');
                if ($w !== '') {
                    $ctx = UnitContext::default($termWidth);
                    $rNode->style->width = Units::resolveLength($w, Axis::X, $ctx);
                }
            }
        }
    }

    public static function parseSrcset(string $srcset): string
    {
        $srcset = trim($srcset);
        if ($srcset === '') {
            return '';
        }
        $candidates = explode(',', $srcset);
        foreach ($candidates as $cand) {
            $cand = trim($cand);
            if ($cand === '') {
                continue;
            }
            $fields = preg_split('/\s+/', $cand, -1, PREG_SPLIT_NO_EMPTY) ?: [];
            if (!empty($fields)) {
                return $fields[0];
            }
        }
        return '';
    }

    public static function applyTextTransform(string $text, string $transform): string
    {
        return match ($transform) {
            'uppercase' => mb_strtoupper($text, 'UTF-8'),
            'lowercase' => mb_strtolower($text, 'UTF-8'),
            'capitalize' => mb_convert_case($text, MB_CASE_TITLE, 'UTF-8'),
            default => $text,
        };
    }

    public static function resolveLang(string ...$classAttrs): string
    {
        foreach ($classAttrs as $raw) {
            $classes = preg_split('/\s+/', trim($raw), -1, PREG_SPLIT_NO_EMPTY) ?: [];
            foreach ($classes as $cls) {
                $lower = strtolower($cls);
                foreach (['language-', 'lang-', 'highlight-', 'sourcecode-'] as $prefix) {
                    if (str_starts_with($lower, $prefix)) {
                        $l = substr($lower, strlen($prefix));
                        if ($l !== '') {
                            return $l;
                        }
                    }
                }
            }
        }
        return '';
    }
}
