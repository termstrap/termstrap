<?php

declare(strict_types=1);

namespace Termstrap\Css;

final readonly class CssRule
{
 public const ORIGIN_TAG_DEFAULTS = 0;
 public const ORIGIN_FRAMEWORK = 1;
 public const ORIGIN_THEME = 2;
 public const ORIGIN_CUSTOM = 3;

 /**
  * @param list<array{property: string, value: string, important: bool}> $decls
  */
 public function __construct(
  public string $selector,
  public array $decls,
  public ?MediaCondition $mediaCond,
  public Specificity $specificity,
  public int $origin,
  public int $index,
  public string $pseudo = '',
 ) {}
}
