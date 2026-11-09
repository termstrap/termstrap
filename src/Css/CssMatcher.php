<?php

declare(strict_types=1);

namespace Termstrap\Css;

use Dom\Element;
use Dom\Node;
use Sabberworm\CSS\CSSList\AtRuleBlockList;
use Sabberworm\CSS\CSSList\Document;
use Sabberworm\CSS\OutputFormat;
use Sabberworm\CSS\Parser;
use Sabberworm\CSS\RuleSet\DeclarationBlock;
use Sabberworm\CSS\Settings;
use Termstrap\Util;
use Throwable;

final class CssMatcher
{
 /** @var array<string, self> */
 private static array $builtInMatcherCache = [];

 /** @var list<CssRule> */
 public array $rules = [];

 public ?Preprocessor $preprocessor = null;

 public function __construct(
  public string $theme = 'bootstrap',
  array $stylesheets = [],
  ?string $rootPath = null,
 ) {
  $this->theme = strtolower(trim($this->theme));
  if ($this->theme === '') {
   $this->theme = 'bootstrap';
  }

  $assetsDir = dirname(__DIR__, 2) . '/assets/css';

  /** @var list<array{content: string, origin: int}> $allSheets */
  $allSheets = [];

  // 1. Base Framework: bootstrap.css
  $bsPath = $assetsDir . '/bootstrap.css';
  if (file_exists($bsPath)) {
   $allSheets[] = ['content' => (string) file_get_contents($bsPath), 'origin' => CssRule::ORIGIN_FRAMEWORK];
  }

  // 2. TermStrap Custom Terminal Overwrites
  $tsPath = $assetsDir . '/termstrap.css';
  if (file_exists($tsPath)) {
   $allSheets[] = ['content' => (string) file_get_contents($tsPath), 'origin' => CssRule::ORIGIN_FRAMEWORK];
  }

  // 3. Theme CSS variables
  if ($this->theme !== '' && $this->theme !== 'none' && $this->theme !== 'bootstrap') {
   $themePath = $assetsDir . '/theme-' . $this->theme . '.css';
   if (file_exists($themePath)) {
    $allSheets[] = ['content' => (string) file_get_contents($themePath), 'origin' => CssRule::ORIGIN_THEME];
   }
  }

  // 4. Custom stylesheets
  foreach ($stylesheets as $sheetContent) {
   $allSheets[] = ['content' => (string) $sheetContent, 'origin' => CssRule::ORIGIN_CUSTOM];
  }

  $this->preprocessor = new Preprocessor($this->theme);

  /** @var list<array{sheet: Document, origin: int}> $parsed */
  $parsed = [];

  $settings = Settings::create()->withLenientParsing(true);
  foreach ($allSheets as $item) {
   $content = trim($item['content']);
   if ($content === '') {
    continue;
   }
   try {
    $parser = new Parser($content, $settings);
    $sheet = $parser->parse();
    $parsed[] = ['sheet' => $sheet, 'origin' => $item['origin']];
    $this->preprocessor->extractVariables($sheet);
   } catch (Throwable) {
    // Ignore malformed stylesheet
   }
  }

  $format = new OutputFormat();
  $globalIndex = 0;

  $addBlock = function (DeclarationBlock $block, ?MediaCondition $parentMedia, int $origin) use (&$globalIndex, $format): void {
   $localVars = [];
   foreach ($block->getRules() as $rule) {
    $prop = trim($rule->getRule());
    if (str_starts_with($prop, '--')) {
     $val = $rule->getValue();
     $localVars[$prop] = trim(is_string($val) ? $val : $val->render($format));
    }
   }

   $decls = [];
   foreach ($block->getRules() as $rule) {
    $prop = trim($rule->getRule());
    $val = $rule->getValue();
    $strVal = is_string($val) ? $val : $val->render($format);
    $decls[] = [
     'property' => $prop,
     'value' => $this->preprocessor->preprocessValue($strVal, $localVars),
     'important' => $rule->getIsImportant(),
    ];
   }
   foreach ($block->getSelectors() as $selObj) {
    $rawSel = trim((string) $selObj->getSelector());
    [$base, $pseudo] = self::splitPseudoElement($rawSel);
    if ($base === '') {
     continue;
    }

    $globalIndex++;
    $this->rules[] = new CssRule(
     selector: $base,
     decls: $decls,
     mediaCond: $parentMedia,
     specificity: Specificity::compute($rawSel),
     origin: $origin,
     index: $globalIndex,
     pseudo: $pseudo,
    );
   }
  };

  $processContents = function (array $contents, ?MediaCondition $parentMedia, int $origin) use (&$processContents, $addBlock): void {
   foreach ($contents as $item) {
    if ($item instanceof DeclarationBlock) {
     $addBlock($item, $parentMedia, $origin);
    } elseif ($item instanceof AtRuleBlockList) {
     $ruleName = strtolower(trim($item->atRuleName()));
     $currentMedia = $parentMedia;
     if ($ruleName === 'media') {
      $mediaCond = MediaCondition::parse($item->atRuleArgs());
      $currentMedia = $parentMedia !== null ? $parentMedia->merge($mediaCond) : $mediaCond;
     }
     $processContents($item->getContents(), $currentMedia, $origin);
    }
   }
  };

  foreach ($parsed as $ps) {
   $processContents($ps['sheet']->getContents(), null, $ps['origin']);
  }

  // Sort rules in ascending cascade priority order (lowest priority first, highest last)
  usort($this->rules, function (CssRule $r1, CssRule $r2): int {
   // 1. Origin (Tag < Framework < Theme < Custom)
   if ($r1->origin !== $r2->origin) {
    return $r1->origin <=> $r2->origin;
   }

   // 2. Specificity (A, B, C)
   $specCmp = $r1->specificity->compare($r2->specificity);
   if ($specCmp !== 0) {
    return $specCmp;
   }

   // 3. Document order (earlier < later)
   return $r1->index <=> $r2->index;
  });
 }

 public static function create(string $theme = 'bootstrap', array $stylesheets = [], ?string $rootPath = null): self
 {
  $themeStr = strtolower(trim($theme));
  if ($themeStr === '') {
   $themeStr = 'bootstrap';
  }

  if (empty($stylesheets)) {
   if (isset(self::$builtInMatcherCache[$themeStr])) {
    return self::$builtInMatcherCache[$themeStr];
   }
   $matcher = new self($themeStr, [], $rootPath);
   self::$builtInMatcherCache[$themeStr] = $matcher;
   return $matcher;
  }

  return new self($themeStr, $stylesheets, $rootPath);
 }

 /**
  * @return array{0: string, 1: string} (base, pseudo)
  */
 public static function splitPseudoElement(string $sel): array
 {
  $low = strtolower($sel);
  $suffixes = [
   ['::before', 'before'],
   [':before', 'before'],
   ['::after', 'after'],
   [':after', 'after'],
  ];

  foreach ($suffixes as [$suf, $name]) {
   if (str_ends_with($low, $suf)) {
    $base = trim(substr($sel, 0, -strlen($suf)));
    return [$base, $name];
   }
  }

  return [$sel, ''];
 }

 public function computeStyleForNode(Element $node, int $termWidth, ?ComputedStyle $parentStyle = null): ComputedStyle
 {
  $s = new ComputedStyle();
  $s->display = 'block';
  $s->textAlign = 'left';

  $tagName = strtolower($node->tagName);
  switch ($tagName) {
   case 'span':
   case 'a':
   case 'strong':
   case 'b':
   case 'em':
   case 'i':
   case 'code':
   case 'img':
    $s->display = 'inline';
    break;
   case 'blink':
    $s->display = 'inline';
    $s->blink = true;
    $s->blinkSet = true;
    break;
   case 'spinner':
    $s->display = 'inline';
    $s->spinner = true;
    $s->spinnerSet = true;
    break;
   case 'shimmer':
    $s->display = 'inline';
    $s->shimmer = true;
    $s->shimmerSet = true;
    break;
   case 's':
   case 'del':
   case 'strike':
    $s->display = 'inline';
    $s->strikethrough = true;
    $s->strikethroughSet = true;
    break;
   case 'u':
   case 'ins':
    $s->display = 'inline';
    $s->underline = true;
    $s->underlineSet = true;
    break;
   case 'th':
   case 'td':
    $s->display = 'block';
    break;
  }

  $ctx = UnitContext::default($termWidth);
  if ($parentStyle !== null && $parentStyle->width > 0) {
   $ctx = new UnitContext(
    parentWidth: $parentStyle->width,
    parentHeight: $ctx->parentHeight,
    terminalWidth: $ctx->terminalWidth,
    terminalHeight: $ctx->terminalHeight,
    rootFontSize: $ctx->rootFontSize,
   );
  }

        // Filter matched rules respecting media conditions
  /** @var list<CssRule> $matchedRules */
  $matchedRules = [];
  foreach ($this->rules as $rule) {
   if ($rule->pseudo !== '') {
    continue;
   }
   if ($rule->mediaCond !== null && !$rule->mediaCond->matches($termWidth, 24, $this->theme)) {
    continue;
   }
   try {
    if ($node->matches($rule->selector)) {
     $matchedRules[] = $rule;
    }
   } catch (Throwable) {
    // Selector not supported by DOM query
   }
  }

  // Pass 1: Apply normal stylesheet declarations in ascending specificity order
  foreach ($matchedRules as $rule) {
   foreach ($rule->decls as $decl) {
    if (!$decl['important']) {
     $this->applyDeclaration($s, $decl['property'], $decl['value'], $ctx);
    }
   }
  }

  // Apply HTML attributes as base inline declarations if present
  if ($node->hasAttribute('color')) {
   $colorAttr = $node->getAttribute('color');
   if ($colorAttr !== '') {
    $this->applyDeclaration($s, 'color', $colorAttr, $ctx);
   }
  }
  if ($node->hasAttribute('width')) {
   $widthAttr = $node->getAttribute('width');
   if ($widthAttr !== '') {
    $this->applyDeclaration($s, 'width', $widthAttr, $ctx);
   }
  }

  // Parse inline styles
  $inlineDecls = [];
  if ($node->hasAttribute('style')) {
   $inlineStyle = $node->getAttribute('style');
   if ($inlineStyle !== '') {
    $inlineDecls = $this->parseInlineDeclarations($inlineStyle);
   }
  }

  // Pass 2: Apply normal inline style declarations
  foreach ($inlineDecls as $decl) {
   if (!$decl['important']) {
    $this->applyDeclaration($s, $decl['property'], $decl['value'], $ctx);
   }
  }

  // Pass 3: Apply !important stylesheet declarations
  foreach ($matchedRules as $rule) {
   foreach ($rule->decls as $decl) {
    if ($decl['important']) {
     $this->applyDeclaration($s, $decl['property'], $decl['value'], $ctx);
    }
   }
  }

  // Pass 4: Apply !important inline style declarations
  foreach ($inlineDecls as $decl) {
   if ($decl['important']) {
    $this->applyDeclaration($s, $decl['property'], $decl['value'], $ctx);
   }
  }

  // Extract responsive column spans and row-cols based on current breakpoint
  $bp = self::detectBreakpoint($termWidth);
  $classAttr = $node->getAttribute('class') ?? '';
  $classes = preg_split('/\s+/', trim($classAttr), -1, PREG_SPLIT_NO_EMPTY) ?: [];

  foreach ($classes as $c) {
   if ($c === 'spinner') {
    $s->spinner = true;
    $s->spinnerSet = true;
   } elseif ($c === 'shimmer') {
    $s->shimmer = true;
    $s->shimmerSet = true;
   } elseif ($c === 'border-shimmer') {
    $s->borderShimmer = true;
   }
  }

  [$span, $isAuto] = self::resolveColSpanResponsive($classes, $bp);
  if ($span > 0) {
   $s->colSpan = $span;
  }
  $s->colAuto = $isAuto;
  $s->rowCols = self::resolveRowColsResponsive($classes, $bp);

  // Inherit missing properties from parentStyle if present
  if ($parentStyle !== null) {
   if (!$s->fgColorSet) {
    $s->fgColor = $parentStyle->fgColor;
    $s->fgColorSet = $parentStyle->fgColorSet;
   }
   if (!$s->textAlignSet) {
    $s->textAlign = $parentStyle->textAlign;
    $s->textAlignSet = $parentStyle->textAlignSet;
   }
   if (!$s->boldSet) {
    $s->bold = $parentStyle->bold;
    $s->boldSet = $parentStyle->boldSet;
   }
   if (!$s->italicSet) {
    $s->italic = $parentStyle->italic;
    $s->italicSet = $parentStyle->italicSet;
   }
   if (!$s->underlineSet) {
    $s->underline = $parentStyle->underline;
    $s->underlineSet = $parentStyle->underlineSet;
   }
   if (!$s->underlineStyleSet) {
    $s->underlineStyle = $parentStyle->underlineStyle;
    $s->underlineStyleSet = $parentStyle->underlineStyleSet;
   }
   if (!$s->underlineColorSet) {
    $s->underlineColor = $parentStyle->underlineColor;
    $s->underlineColorSet = $parentStyle->underlineColorSet;
   }
   if (!$s->strikethroughSet) {
    $s->strikethrough = $parentStyle->strikethrough;
    $s->strikethroughSet = $parentStyle->strikethroughSet;
   }
   if (!$s->faintSet) {
    $s->faint = $parentStyle->faint;
    $s->faintSet = $parentStyle->faintSet;
   }
   if (!$s->blinkSet) {
    $s->blink = $parentStyle->blink;
    $s->blinkSet = $parentStyle->blinkSet;
   }
   if (!$s->shimmerSet) {
    $s->shimmer = $parentStyle->shimmer;
    $s->shimmerSet = $parentStyle->shimmerSet;
    $s->shimmerHighlight = $parentStyle->shimmerHighlight;
   }
   if (!$s->spinnerSet) {
    $s->spinner = $parentStyle->spinner;
    $s->spinnerSet = $parentStyle->spinnerSet;
   }
   if (!$s->textTransformSet) {
    $s->textTransform = $parentStyle->textTransform;
    $s->textTransformSet = $parentStyle->textTransformSet;
   }
  }

  return $s;
 }

 /**
  * @return array{0: string, 1: ComputedStyle}|null
  */
 public function pseudoText(Element $node, int $termWidth, ?ComputedStyle $elementStyle, string $pseudo): ?array
 {
  $ctx = UnitContext::default($termWidth);
  /** @var list<CssRule> $matched */
  $matched = [];

  foreach ($this->rules as $r) {
   if ($r->pseudo !== $pseudo) {
    continue;
   }
   if ($r->mediaCond !== null && !$r->mediaCond->matches($termWidth, 24, $this->theme)) {
    continue;
   }
   try {
    if ($node->matches($r->selector)) {
     $matched[] = $r;
    }
   } catch (Throwable) {
   }
  }

  if (empty($matched)) {
   return null;
  }

  $s = new ComputedStyle();
  $s->display = 'inline';

  if ($elementStyle !== null) {
   $s->bold = $elementStyle->bold;
   $s->italic = $elementStyle->italic;
   $s->underline = $elementStyle->underline;
   $s->strikethrough = $elementStyle->strikethrough;
   $s->faint = $elementStyle->faint;
   $s->blink = $elementStyle->blink;
   $s->fgColor = $elementStyle->fgColor;
   $s->fgColorSet = $elementStyle->fgColorSet;
   $s->textTransform = $elementStyle->textTransform;
   $s->textTransformSet = $elementStyle->textTransformSet;
  }

  $content = '';
  $hasContent = false;

  $apply = function (bool $important) use ($matched, $ctx, &$content, &$hasContent, $s): void {
   foreach ($matched as $r) {
    foreach ($r->decls as $decl) {
     if ($decl['important'] !== $important) {
      continue;
     }
     if (strcasecmp(trim($decl['property']), 'content') === 0) {
      $txt = self::parseCssContent($decl['value']);
      if ($txt !== null) {
       $content = $txt;
       $hasContent = true;
      } else {
       $hasContent = false;
      }
      continue;
     }
     $this->applyDeclaration($s, $decl['property'], $decl['value'], $ctx);
    }
   }
  };

  $apply(false);
  $apply(true);

  if (!$hasContent || $content === '') {
   return null;
  }

  return [$content, $s];
 }

 public static function parseCssContent(string $val): ?string
 {
  $v = trim($val);
  if ($v === '' || $v === 'none' || $v === 'normal') {
   return null;
  }

  $quote = $v[0];
  if ($quote !== '"' && $quote !== "'") {
   return null;
  }

  $len = strlen($v);
  $sb = '';
  $i = 1;

  while ($i < $len) {
   $c = $v[$i];
   if ($c === '\\') {
    $j = $i + 1;
    $hex = '';
    while ($j < $len && strlen($hex) < 6 && ctype_xdigit($v[$j])) {
     $hex .= $v[$j];
     $j++;
    }
    if ($hex !== '') {
     $cp = hexdec($hex);
     $sb .= mb_chr((int) $cp, 'UTF-8');
     if ($j < $len && ($v[$j] === ' ' || $v[$j] === "\t")) {
      $j++;
     }
     $i = $j;
     continue;
    }
    if ($i + 1 < $len) {
     $sb .= $v[$i + 1];
     $i += 2;
     continue;
    }
    $i++;
    continue;
   }
   if ($c === $quote) {
    break;
   }
   $sb .= $c;
   $i++;
  }

  return $sb;
 }

 /**
  * @return list<array{property: string, value: string, important: bool}>
  */
 private function parseInlineDeclarations(string $inlineStyle): array
 {
  $decls = [];
  $parts = explode(';', $inlineStyle);
  $localVars = [];

  foreach ($parts as $part) {
   $part = trim($part);
   if ($part === '') {
    continue;
   }
   $colon = strpos($part, ':');
   if ($colon === false) {
    continue;
   }
   $prop = trim(substr($part, 0, $colon));
   $val = trim(substr($part, $colon + 1));
   if (str_starts_with($prop, '--')) {
    $localVars[$prop] = $val;
   }
  }

  foreach ($parts as $part) {
   $part = trim($part);
   if ($part === '') {
    continue;
   }
   $colon = strpos($part, ':');
   if ($colon === false) {
    continue;
   }
   $prop = trim(substr($part, 0, $colon));
   $val = trim(substr($part, $colon + 1));

   $important = false;
   if (preg_match('/\s+!important\s*$/i', $val)) {
    $important = true;
    $val = (string) preg_replace('/\s+!important\s*$/i', '', $val);
   }

   $preprocessedVal = $this->preprocessor !== null
    ? $this->preprocessor->preprocessValue($val, $localVars)
    : $val;

   $decls[] = [
    'property' => $prop,
    'value' => $preprocessedVal,
    'important' => $important,
   ];
  }

  return $decls;
 }

 public function applyDeclaration(ComputedStyle $s, string $property, string $rawVal, UnitContext $ctx): void
 {
  $prop = strtolower(trim($property));
  $val = strtolower(trim($rawVal));

  switch ($prop) {
   case 'display':
    match ($val) {
     'block' => $s->display = 'block',
     'inline' => $s->display = 'inline',
     'flex' => $s->display = 'flex',
     'none' => $s->display = 'none',
     default => null,
    };
    break;

   case 'margin':
    [$s->marginTop, $s->marginRight, $s->marginBottom, $s->marginLeft] = Units::resolveBox4D($val, $ctx);
    break;
   case 'margin-top':
    $s->marginTop = Units::resolveLength($val, Axis::Y, $ctx);
    break;
   case 'margin-bottom':
    $s->marginBottom = Units::resolveLength($val, Axis::Y, $ctx);
    break;
   case 'margin-left':
    $s->marginLeft = Units::resolveLength($val, Axis::X, $ctx);
    break;
   case 'margin-right':
    $s->marginRight = Units::resolveLength($val, Axis::X, $ctx);
    break;

   case 'padding':
    [$s->paddingTop, $s->paddingRight, $s->paddingBottom, $s->paddingLeft] = Units::resolveBox4D($val, $ctx);
    break;
   case 'padding-top':
    $s->paddingTop = Units::resolveLength($val, Axis::Y, $ctx);
    break;
   case 'padding-bottom':
    $s->paddingBottom = Units::resolveLength($val, Axis::Y, $ctx);
    break;
   case 'padding-left':
    $s->paddingLeft = Units::resolveLength($val, Axis::X, $ctx);
    break;
   case 'padding-right':
    $s->paddingRight = Units::resolveLength($val, Axis::X, $ctx);
    break;

   case 'width':
    $s->width = Units::resolveLength($val, Axis::X, $ctx);
    break;

   case 'font-size':
    [$bold] = Units::resolveFontSize($val, $ctx->rootFontSize);
    if ($bold && !$s->boldSet) {
     $s->bold = true;
     $s->boldSet = true;
    }
    break;

   case 'border':
    if (
     $val === 'true' || $val === '1' || str_contains($val, 'solid') || str_contains($val, '1px')
     || str_contains($val, 'double') || str_contains($val, 'dashed') || str_contains($val, 'dotted')
    ) {
     $s->border = true;
    }
    $bs = self::borderStyleKeyword($val);
    if ($bs !== '') {
     $s->borderStyle = $bs;
    }
    if (($bs === '' || $bs === 'solid') && self::borderWidthIsThick($val)) {
     $s->borderStyle = 'thick';
    }
    $fields = preg_split('/\s+/', $val, -1, PREG_SPLIT_NO_EMPTY) ?: [];
    foreach ($fields as $part) {
     if (str_starts_with($part, '#') && strlen($part) >= 4) {
      $s->borderColor = $part;
     }
    }
    break;

   case 'border-top':
    $s->borderTop = $val === 'true' || $val === '1' || str_contains($val, 'solid');
    break;
   case 'border-bottom':
    $s->borderBottom = $val === 'true' || $val === '1' || str_contains($val, 'solid');
    break;
   case 'border-left':
    $s->borderLeft = $val === 'true' || $val === '1' || str_contains($val, 'solid');
    break;
   case 'border-right':
    $s->borderRight = $val === 'true' || $val === '1' || str_contains($val, 'solid');
    break;

   case 'border-radius':
    $s->rounded = $val === 'true' || $val === '1' || $val === 'rounded'
     || ($val !== '0' && $val !== '0px' && $val !== '0rem' && $val !== 'none' && $val !== '');
    break;

   case 'color':
    if ($val === 'inherit' || $val === 'initial') {
     $s->fgColor = '';
     $s->fgColorSet = false;
    } else {
     $s->fgColor = self::resolveColor($rawVal);
     $s->fgColorSet = true;
    }
    break;

   case 'background-color':
   case 'background':
    if (str_contains($val, 'gradient')) {
     $s->bgGradient = self::parseColorStops($rawVal);
    } elseif ($val === 'transparent' || $val === 'inherit' || $val === 'initial') {
     $s->bgColor = '';
     $s->bgGradient = [];
    } else {
     $s->bgColor = self::resolveColor($rawVal);
    }
    break;

   case 'border-color':
    $s->borderColor = self::resolveColor($rawVal);
    break;

   case 'font-weight':
    $s->boldSet = true;
    if ($val === 'bold' || $val === '700' || $val === '800' || $val === '900') {
     $s->bold = true;
    } elseif ($val === 'normal' || $val === '400') {
     $s->bold = false;
    }
    break;

   case 'font-style':
    $s->italicSet = true;
    if ($val === 'italic' || $val === 'oblique') {
     $s->italic = true;
    } elseif ($val === 'normal') {
     $s->italic = false;
    }
    break;

   case 'text-decoration':
   case 'text-decoration-line':
    if (str_contains($val, 'underline')) {
     $s->underlineSet = true;
     $s->underline = true;
    }
    if (str_contains($val, 'wavy') || str_contains($val, 'curly')) {
     $s->underlineStyle = 'curly';
     $s->underlineStyleSet = true;
    } elseif (str_contains($val, 'double')) {
     $s->underlineStyle = 'double';
     $s->underlineStyleSet = true;
    } elseif (str_contains($val, 'dotted')) {
     $s->underlineStyle = 'dotted';
     $s->underlineStyleSet = true;
    } elseif (str_contains($val, 'dashed')) {
     $s->underlineStyle = 'dashed';
     $s->underlineStyleSet = true;
    } elseif (str_contains($val, 'solid')) {
     $s->underlineStyle = 'single';
     $s->underlineStyleSet = true;
    }
    if (str_contains($val, 'line-through')) {
     $s->strikethroughSet = true;
     $s->strikethrough = true;
    }
    if (str_contains($val, 'blink')) {
     $s->blinkSet = true;
     $s->blink = true;
    }
    if (str_contains($val, 'shimmer')) {
     $s->shimmerSet = true;
     $s->shimmer = true;
    }
    if (str_contains($val, 'spinner')) {
     $s->spinnerSet = true;
     $s->spinner = true;
    }
    if (str_contains($val, 'none')) {
     $s->underlineSet = true;
     $s->underline = false;
     $s->underlineStyleSet = true;
     $s->underlineStyle = 'none';
     $s->strikethroughSet = true;
     $s->strikethrough = false;
    }
    break;

   case 'animation':
   case 'animation-name':
    if (str_contains($val, 'shimmer')) {
     $s->shimmerSet = true;
     $s->shimmer = true;
    }
    if (str_contains($val, 'spinner')) {
     $s->spinnerSet = true;
     $s->spinner = true;
    }
    if (str_contains($val, 'blink')) {
     $s->blinkSet = true;
     $s->blink = true;
    }
    break;

   case 'text-decoration-style':
    $s->underlineStyleSet = true;
    $s->underlineStyle = match (trim($val)) {
     'wavy', 'curly' => 'curly',
     'double' => 'double',
     'dotted' => 'dotted',
     'dashed' => 'dashed',
     default => 'single',
    };
    break;

   case 'text-decoration-color':
    $s->underlineColor = $rawVal;
    $s->underlineColorSet = true;
    break;

   case 'text-align':
    $s->textAlignSet = true;
    $s->textAlign = match ($val) {
     'center' => 'center',
     'right', 'end' => 'right',
     default => 'left',
    };
    break;

   case 'text-transform':
    $s->textTransformSet = true;
    $s->textTransform = match ($val) {
     'uppercase', 'lowercase', 'capitalize' => $val,
     default => '',
    };
    break;

   case 'box-shadow':
    $s->shadow = self::shadowSizeFromBoxShadow($val);
    break;

   case 'opacity':
    if (is_numeric(trim($val))) {
     $f = (float) trim($val);
     $s->faintSet = true;
     $s->faint = $f > 0 && $f < 1;
    }
    break;

   case 'border-style':
    $bs = self::borderStyleKeyword($val);
    if ($bs !== '') {
     $s->borderStyle = $bs;
     if ($bs === 'hidden' || $bs === 'none') {
      $s->border = false;
     }
    }
    break;

   case 'border-width':
    if (self::borderWidthIsThick($val)) {
     $s->borderStyle = 'thick';
    }
    break;

   case 'border-top-color':
    $s->borderTopColor = $rawVal;
    break;
   case 'border-right-color':
    $s->borderRightColor = $rawVal;
    break;
   case 'border-bottom-color':
    $s->borderBottomColor = $rawVal;
    break;
   case 'border-left-color':
    $s->borderLeftColor = $rawVal;
    break;

   case 'border-image':
   case 'border-color-blend':
    $s->borderBlend = self::parseColorStops($rawVal);
    break;

   case 'max-width':
    if ($val !== 'none') {
     $s->maxWidth = Units::resolveLength($val, Axis::X, $ctx);
    }
    break;

   case 'max-height':
    if ($val !== 'none') {
     $s->maxHeight = Units::resolveLength($val, Axis::Y, $ctx);
    }
    break;

   case 'align-items':
    $s->alignItemsSet = true;
    $s->alignItems = match ($val) {
     'center' => 'center',
     'flex-end', 'end', 'bottom' => 'bottom',
     default => 'top',
    };
    break;

   case 'justify-content':
    $s->justifyContentSet = true;
    $s->justifyContent = match ($val) {
     'center' => 'center',
     'flex-end', 'end', 'right' => 'right',
     default => 'left',
    };
    break;

   case 'vertical-align':
    $s->verticalAlignSet = true;
    $s->verticalAlign = match ($val) {
     'middle' => 'center',
     'bottom', 'text-bottom' => 'bottom',
     default => 'top',
    };
    break;

   case 'list-style':
   case 'list-style-type':
    $s->listStyle = self::listStyleKeyword($val);
    break;
  }
 }

 public static function borderStyleKeyword(string $val): string
 {
  foreach (['outer-half-block', 'inner-half-block', 'block', 'markdown', 'ascii', 'double', 'dashed', 'dotted', 'hidden', 'none', 'solid'] as $kw) {
   if (str_contains($val, $kw)) {
    return $kw;
   }
  }
  return '';
 }

 public static function borderWidthIsThick(string $val): bool
 {
  if (str_contains($val, 'thick') || str_contains($val, 'medium')) {
   return true;
  }
  if (preg_match_all('/(-?\d*\.?\d+)(rem|px)/', $val, $matches, PREG_SET_ORDER)) {
   foreach ($matches as $m) {
    $f = (float) $m[1];
    if (($m[2] === 'px' && $f > 1) || ($m[2] === 'rem' && $f > 0.0625)) {
     return true;
    }
   }
  }
  return false;
 }

 public static function listStyleKeyword(string $val): string
 {
  foreach (['none', 'disc', 'circle', 'square', 'decimal', 'lower-alpha', 'lower-latin', 'upper-alpha', 'upper-latin', 'lower-roman', 'upper-roman'] as $kw) {
   if (str_contains($val, $kw)) {
    return $kw;
   }
  }
  return '';
 }

 /**
  * @return list<string>
  */
 public static function parseColorStops(string $val): array
 {
  $stops = [];
  if (preg_match_all('/#[0-9a-fA-F]{3,8}/', $val, $matches)) {
   foreach ($matches[0] as $m) {
    $stops[] = $m;
   }
  }
  return $stops;
 }

 public static function resolveColor(string $val): string
 {
  $v = trim($val);
  if (str_starts_with($v, 'lighten(') && str_ends_with($v, ')')) {
   $inner = substr($v, 8, -1);
   $parts = explode(',', $inner);
   if (count($parts) >= 2) {
    $hex = trim($parts[0]);
    $pctStr = rtrim(trim($parts[1]), '%');
    if (is_numeric($pctStr)) {
     $pct = max(0.0, min(1.0, ((float) $pctStr) / 100.0));
     [$r, $g, $b] = Util::hexToRgb($hex);
     $add = 255.0 * $pct;
     $r = (int) min(255, round($r + $add));
     $g = (int) min(255, round($g + $add));
     $b = (int) min(255, round($b + $add));
     return sprintf('#%02x%02x%02x', $r, $g, $b);
    }
   }
  } elseif (str_starts_with($v, 'darken(') && str_ends_with($v, ')')) {
   $inner = substr($v, 7, -1);
   $parts = explode(',', $inner);
   if (count($parts) >= 2) {
    $hex = trim($parts[0]);
    $pctStr = rtrim(trim($parts[1]), '%');
    if (is_numeric($pctStr)) {
     $pct = max(0.0, min(1.0, ((float) $pctStr) / 100.0));
     $mult = 1.0 - $pct;
     [$r, $g, $b] = Util::hexToRgb($hex);
     $r = (int) round($r * $mult);
     $g = (int) round($g * $mult);
     $b = (int) round($b * $mult);
     return sprintf('#%02x%02x%02x', $r, $g, $b);
    }
   }
  } elseif (str_starts_with($v, 'complementary(') && str_ends_with($v, ')')) {
   $inner = trim(substr($v, 14, -1));
   [$r, $g, $b] = Util::hexToRgb($inner);
   [$compR, $compG, $compB] = self::calcComplementary($r, $g, $b);
   return sprintf('#%02x%02x%02x', $compR, $compG, $compB);
  }

  return $val;
 }

 /**
  * @return array{0: int, 1: int, 2: int}
  */
 private static function calcComplementary(int $r, int $g, int $b): array
 {
  $rf = $r / 255.0;
  $gf = $g / 255.0;
  $bf = $b / 255.0;

  $max = max($rf, $gf, $bf);
  $min = min($rf, $gf, $bf);
  $d = $max - $min;
  $v = $max;
  $s = $max == 0.0 ? 0.0 : $d / $max;

  if ($max == $min) {
   $h = 0.0;
  } else {
   if ($max == $rf) {
    $h = ($gf - $bf) / $d + ($gf < $bf ? 6.0 : 0.0);
   } elseif ($max == $gf) {
    $h = ($bf - $rf) / $d + 2.0;
   } else {
    $h = ($rf - $gf) / $d + 4.0;
   }
   $h *= 60.0;
  }

  $h = fmod($h + 180.0, 360.0);
  if ($h < 0.0) {
   $h += 360.0;
  }

  $i = (int) (floor($h / 60.0)) % 6;
  $f = $h / 60.0 - floor($h / 60.0);
  $p = $v * (1.0 - $s);
  $q = $v * (1.0 - $f * $s);
  $t = $v * (1.0 - (1.0 - $f) * $s);

  [$outR, $outG, $outB] = match ($i) {
   0 => [$v, $t, $p],
   1 => [$q, $v, $p],
   2 => [$p, $v, $t],
   3 => [$p, $q, $v],
   4 => [$t, $p, $v],
   default => [$v, $p, $q],
  };

  return [(int) round($outR * 255), (int) round($outG * 255), (int) round($outB * 255)];
 }

 public static function shadowSizeFromBoxShadow(string $val): int
 {
  $val = trim($val);
  if ($val === '' || $val === 'none' || $val === '0') {
   return 0;
  }
  if (str_contains($val, 'inset')) {
   return 0;
  }

  if (!preg_match_all('/(-?\d*\.?\d+)(rem|px)/', $val, $matches, PREG_SET_ORDER) || empty($matches)) {
   return Units::resolveLength($val, Axis::X, UnitContext::default(80));
  }

  $maxRem = 0.0;
  foreach ($matches as $m) {
   $f = (float) $m[1];
   if ($f < 0) {
    $f = -$f;
   }
   if ($m[2] === 'px') {
    $f /= 16.0;
   }
   if ($f > $maxRem) {
    $maxRem = $f;
   }
  }

  if ($maxRem <= 0.3) {
   return 1;
  }
  if ($maxRem <= 1.5) {
   return 2;
  }
  return 3;
 }

 public static function detectBreakpoint(int $width): int
 {
  return match (true) {
   $width >= 190 => 5, // XXL
   $width >= 160 => 4, // XL
   $width >= 130 => 3, // LG
   $width >= 100 => 2, // MD
   $width >= 70 => 1,  // SM
   default => 0,       // XS
  };
 }

 /**
  * @param list<string> $classes
  * @return array{0: int, 1: bool} (span, isAuto)
  */
 public static function resolveColSpanResponsive(array $classes, int $currentBP): array
 {
  /** @var list<array{bp: int, span: int, isAuto: bool}> $specs */
  $specs = [];

  foreach ($classes as $cls) {
   if ($cls === 'col') {
    $specs[] = ['bp' => 0, 'span' => 0, 'isAuto' => true];
    continue;
   }
   if (!str_starts_with($cls, 'col-')) {
    continue;
   }
   $parts = explode('-', $cls);
   if (count($parts) === 2) {
    if ($parts[1] === 'auto') {
     $specs[] = ['bp' => 0, 'span' => 0, 'isAuto' => true];
    } elseif (ctype_digit($parts[1])) {
     $span = (int) $parts[1];
     if ($span >= 1 && $span <= 12) {
      $specs[] = ['bp' => 0, 'span' => $span, 'isAuto' => false];
     }
    }
   } elseif (count($parts) === 3) {
    $bp = match ($parts[1]) {
     'xs' => 0,
     'sm' => 1,
     'md' => 2,
     'lg' => 3,
     'xl' => 4,
     'xxl' => 5,
     default => -1,
    };
    if ($bp < 0) {
     continue;
    }
    if ($parts[2] === 'auto') {
     $specs[] = ['bp' => $bp, 'span' => 0, 'isAuto' => true];
    } elseif (ctype_digit($parts[2])) {
     $span = (int) $parts[2];
     if ($span >= 1 && $span <= 12) {
      $specs[] = ['bp' => $bp, 'span' => $span, 'isAuto' => false];
     }
    }
   }
  }

  $bestSpan = 0;
  $bestAuto = false;
  $bestBP = -1;

  foreach ($specs as $spec) {
   if ($spec['bp'] <= $currentBP && $spec['bp'] > $bestBP) {
    $bestBP = $spec['bp'];
    $bestSpan = $spec['span'];
    $bestAuto = $spec['isAuto'];
   }
  }

  // If classes specify column breakpoints but none match the current breakpoint (e.g. col-md-6 on xs/sm),
  // the column stacks (span = 12).
  if (!empty($specs) && $bestSpan === 0 && !$bestAuto) {
   return [12, false];
  }

  return [$bestSpan, $bestAuto];
 }

 /**
  * @param list<string> $classes
  */
 public static function resolveRowColsResponsive(array $classes, int $currentBP): int
 {
  /** @var list<array{bp: int, count: int}> $specs */
  $specs = [];

  foreach ($classes as $cls) {
   if (!str_starts_with($cls, 'row-cols-')) {
    continue;
   }
   $parts = explode('-', $cls);
   if (count($parts) === 3) {
    if (ctype_digit($parts[2])) {
     $n = (int) $parts[2];
     if ($n >= 1 && $n <= 12) {
      $specs[] = ['bp' => 0, 'count' => $n];
     }
    }
   } elseif (count($parts) === 4) {
    $bp = match ($parts[2]) {
     'xs' => 0,
     'sm' => 1,
     'md' => 2,
     'lg' => 3,
     'xl' => 4,
     'xxl' => 5,
     default => -1,
    };
    if ($bp < 0) {
     continue;
    }
    if (ctype_digit($parts[3])) {
     $n = (int) $parts[3];
     if ($n >= 1 && $n <= 12) {
      $specs[] = ['bp' => $bp, 'count' => $n];
     }
    }
   }
  }

  $bestCount = 0;
  $bestBP = -1;
  foreach ($specs as $spec) {
   if ($spec['bp'] <= $currentBP && $spec['bp'] > $bestBP) {
    $bestBP = $spec['bp'];
    $bestCount = $spec['count'];
   }
  }

  return $bestCount;
 }
}
