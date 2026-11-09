<?php

declare(strict_types=1);

namespace Termstrap\Animation;

final class Spinner
{
 public const SPINNER_FRAMES = ['⠋', '⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏'];

 /**
  * @param list<string> $frames
  */
 public static function spinner(int $step, array $frames = []): string
 {
  $f = !empty($frames) ? $frames : self::SPINNER_FRAMES;
  $count = count($f);
  $idx = $step % $count;
  if ($idx < 0) {
   $idx += $count;
  }
  return $f[$idx];
 }
}
