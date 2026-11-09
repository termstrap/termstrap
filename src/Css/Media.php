<?php

declare(strict_types=1);

namespace Termstrap\Css;

final readonly class MediaClause
{
 public function __construct(
  public string $mediaType = 'all',
  public bool $negate = false,
  public int $minWidth = -1,
  public int $maxWidth = -1,
  public int $minHeight = -1,
  public int $maxHeight = -1,
  public string $orientation = '',
  public string $colorScheme = '',
 ) {}

 public function matches(int $termWidth, int $termHeight, bool $isDark): bool
 {
  $matchesType = match ($this->mediaType) {
   'print', 'speech' => false,
   default => true,
  };

  if ($this->negate) {
   $matchesType = !$matchesType;
  }
  if (!$matchesType) {
   return false;
  }

  // Width boundaries
  if ($this->minWidth >= 0 && $termWidth < $this->minWidth) {
   return false;
  }
  if ($this->maxWidth >= 0 && $termWidth > $this->maxWidth) {
   return false;
  }

  // Height boundaries
  if ($this->minHeight >= 0 && $termHeight < $this->minHeight) {
   return false;
  }
  if ($this->maxHeight >= 0 && $termHeight > $this->maxHeight) {
   return false;
  }

  // Orientation
  if ($this->orientation !== '') {
   if ($this->orientation === 'landscape' && $termWidth < $termHeight) {
    return false;
   }
   if ($this->orientation === 'portrait' && $termWidth >= $termHeight) {
    return false;
   }
  }

  // Color scheme
  if ($this->colorScheme !== '') {
   if ($this->colorScheme === 'dark' && !$isDark) {
    return false;
   }
   if ($this->colorScheme === 'light' && $isDark) {
    return false;
   }
  }

  return true;
 }
}

final readonly class MediaCondition
{
 /**
  * @param list<MediaClause> $clauses
  */
 public function __construct(
  public array $clauses = [],
 ) {}

 public function matches(int $termWidth, int $termHeight, string $theme): bool
 {
  if (empty($this->clauses)) {
   return true;
  }

  $tw = $termWidth > 0 ? $termWidth : 80;
  $th = $termHeight > 0 ? $termHeight : 24;

  $themeStr = strtolower(trim($theme));
  if ($themeStr === '') {
   $themeStr = 'bootstrap';
  }
  $isDark = in_array($themeStr, ['dracula', 'tokyonight', 'glamour', 'dark'], true);

  // OR condition across clauses
  foreach ($this->clauses as $clause) {
   if ($clause->matches($tw, $th, $isDark)) {
    return true;
   }
  }

  return false;
 }

 public function merge(?self $child): self
 {
  if ($child === null || empty($child->clauses)) {
   return $this;
  }
  if (empty($this->clauses)) {
   return $child;
  }

  $combined = [];
  foreach ($this->clauses as $p) {
   foreach ($child->clauses as $c) {
    $combined[] = new MediaClause(
     mediaType: $c->mediaType,
     negate: $c->negate,
     minWidth: max($p->minWidth, $c->minWidth),
     maxWidth: self::minNonNeg($p->maxWidth, $c->maxWidth),
     minHeight: max($p->minHeight, $c->minHeight),
     maxHeight: self::minNonNeg($p->maxHeight, $c->maxHeight),
     orientation: $c->orientation !== '' ? $c->orientation : $p->orientation,
     colorScheme: $c->colorScheme !== '' ? $c->colorScheme : $p->colorScheme,
    );
   }
  }

  return new self($combined);
 }

 private static function minNonNeg(int $a, int $b): int
 {
  if ($a < 0) {
   return $b;
  }
  if ($b < 0) {
   return $a;
  }
  return min($a, $b);
 }

 public static function parse(string $prelude): ?self
 {
  $prelude = trim($prelude);
  if ($prelude === '') {
   return null;
  }

  // Comma-separated acting as OR
  $rawParts = explode(',', $prelude);
  $clauses = [];

  foreach ($rawParts as $part) {
   $part = trim($part);
   if ($part === '') {
    continue;
   }
   $clauses[] = self::parseClause($part);
  }

  if (empty($clauses)) {
   return null;
  }

  return new self($clauses);
 }

 private static function parseClause(string $expr): MediaClause
 {
  $mediaType = 'all';
  $negate = false;
  $minWidth = -1;
  $maxWidth = -1;
  $minHeight = -1;
  $maxHeight = -1;
  $orientation = '';
  $colorScheme = '';

  $tokens = explode('and', $expr);
  foreach ($tokens as $rawTok) {
   $tok = trim($rawTok);
   if ($tok === '') {
    continue;
   }

   if (str_starts_with($tok, '(') && str_ends_with($tok, ')')) {
    $inner = trim(substr($tok, 1, -1));
    $parts = explode(':', $inner, 2);
    if (count($parts) === 2) {
     $key = strtolower(trim($parts[0]));
     $val = trim($parts[1]);
     $ctx = UnitContext::default(80);

     match ($key) {
      'min-width' => $minWidth = self::resolveMediaLengthFloorCeil($val, Axis::X, $ctx, true),
      'max-width' => $maxWidth = self::resolveMediaLengthFloorCeil($val, Axis::X, $ctx, false),
      'width' => $minWidth = $maxWidth = self::resolveMediaLengthFloorCeil($val, Axis::X, $ctx, true),
      'min-height' => $minHeight = self::resolveMediaLengthFloorCeil($val, Axis::Y, $ctx, true),
      'max-height' => $maxHeight = self::resolveMediaLengthFloorCeil($val, Axis::Y, $ctx, false),
      'height' => $minHeight = $maxHeight = self::resolveMediaLengthFloorCeil($val, Axis::Y, $ctx, true),
      'orientation' => $orientation = strtolower($val),
      'prefers-color-scheme' => $colorScheme = strtolower($val),
      default => null,
     };
    }
    continue;
   }

   $lower = strtolower($tok);
   if (str_starts_with($lower, 'not ')) {
    $negate = true;
    $mediaType = trim(substr($lower, 4));
   } elseif (str_starts_with($lower, 'only ')) {
    $mediaType = trim(substr($lower, 5));
   } else {
    $mediaType = $lower;
   }
  }

  return new MediaClause(
   mediaType: $mediaType,
   negate: $negate,
   minWidth: $minWidth,
   maxWidth: $maxWidth,
   minHeight: $minHeight,
   maxHeight: $maxHeight,
   orientation: $orientation,
   colorScheme: $colorScheme,
  );
 }

 private static function resolveMediaLengthFloorCeil(string $val, Axis $axis, UnitContext $ctx, bool $isMin): int
 {
  [$numPart, $unitPart] = Units::splitNumberAndUnit($val);
  if ($numPart === '' || !is_numeric($numPart)) {
   return Units::resolveLength($val, $axis, $ctx);
  }

  $num = (float) $numPart;
  $rootFS = $ctx->rootFontSize > 0 ? $ctx->rootFontSize : 16.0;

  $unitPart = strtolower($unitPart);
  $px = match ($unitPart) {
   'px' => $num,
   'rem', 'em' => $num * $rootFS,
   'ch' => $num * ($rootFS * 0.5),
   'pt' => $num * (96.0 / 72.0),
   default => null,
  };

  if ($px === null) {
   return Units::resolveLength($val, $axis, $ctx);
  }

  if ($axis === Axis::X) {
   $cols = $px / ($rootFS * 0.5);
   return $isMin ? (int) round($cols) : (int) floor($cols);
  }

  $lines = $px / $rootFS;
  return $isMin ? (int) round($lines) : (int) floor($lines);
 }
}
