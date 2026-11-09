<?php

declare(strict_types=1);

namespace Termstrap\Renderer;

use Tempest\Highlight\Highlighter;
use Tempest\Highlight\TerminalTheme;
use Tempest\Highlight\Tokens\TokenType;
use Tempest\Highlight\Tokens\TokenTypeEnum;
use Throwable;

final class TermstrapTerminalTheme implements TerminalTheme
{
 /**
  * @param array<string, string> $palette
  */
 public function __construct(
  private array $palette = [],
 ) {}

 public function before(TokenType $tokenType): string
 {
  $val = $tokenType->getValue();
  $varName = match ($val) {
   'keyword' => '--ts-code-keyword',
   'property' => '--ts-code-function',
   'type' => '--ts-code-type',
   'value', 'literal' => '--ts-code-string',
   'number' => '--ts-code-number',
   'comment' => '--ts-code-comment',
   'operator' => '--ts-code-operator',
   'variable', 'injection' => '--ts-code-variable',
   'attribute' => '--ts-code-attribute',
   'generic' => '--ts-code-name',
   default => '--ts-code-name',
  };

  $color = $this->palette[$varName] ?? CodeHighlighter::DEFAULT_CODE_PALETTE[$varName] ?? '#f8f9fa';
  return Color::fgEscape($color);
 }

 public function after(TokenType $tokenType): string
 {
  return "\x1b[m";
 }

 public function escape(string $content): string
 {
  return $content;
 }
}

final class CodeHighlighter
{
 public const DEFAULT_CODE_PALETTE = [
  '--ts-code-name'      => '#f8f9fa',
  '--ts-code-keyword'   => '#e685b5',
  '--ts-code-function'  => '#6ea8fe',
  '--ts-code-type'      => '#6edff6',
  '--ts-code-builtin'   => '#c29ffa',
  '--ts-code-string'    => '#75b798',
  '--ts-code-number'    => '#ffda6a',
  '--ts-code-comment'   => '#9aa0a6',
  '--ts-code-operator'  => '#dee2e6',
  '--ts-code-variable'  => '#f8f9fa',
  '--ts-code-constant'  => '#c29ffa',
  '--ts-code-tag'       => '#e685b5',
  '--ts-code-attribute' => '#6ea8fe',
  '--ts-code-regex'     => '#63e6be',
  '--ts-code-escape'    => '#ffbf69',
  '--ts-code-deleted'   => '#ea868f',
  '--ts-code-inserted'  => '#75b798',
  '--ts-code-heading'   => '#6ea8fe',
 ];

 /**
  * @param array<string, string> $vars
  */
 public static function highlight(string $code, string $lang, array $vars = []): string
 {
  $lang = strtolower(trim($lang));
  if ($lang === 'text' || $lang === 'plaintext' || $lang === 'none') {
   return $code;
  }

  try {
   $theme = new TermstrapTerminalTheme($vars);
   $hl = new Highlighter($theme);
   if ($lang === '' || $lang === 'auto') {
    return $hl->parse($code, 'php'); // default fallback or parse
   }
   return $hl->parse($code, $lang);
  } catch (Throwable) {
   return $code;
  }
 }
}
