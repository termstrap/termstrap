<?php

declare(strict_types=1);

namespace Termstrap\Css;

final readonly class Specificity
{
    public function __construct(
        public int $a = 0,
        public int $b = 0,
        public int $c = 0,
    ) {}

    public function compare(self $other): int
    {
        if ($this->a !== $other->a) {
            return $this->a - $other->a;
        }
        if ($this->b !== $other->b) {
            return $this->b - $other->b;
        }
        return $this->c - $other->c;
    }

    public function less(self $other): bool
    {
        return $this->compare($other) < 0;
    }

    public static function compute(string $sel): self
    {
        $sel = trim($sel);
        if ($sel === '' || $sel === '*') {
            return new self(0, 0, 0);
        }

        $a = 0;
        $b = 0;
        $c = 0;

        $n = strlen($sel);
        $i = 0;
        $expectElement = true;

        while ($i < $n) {
            $ch = $sel[$i];

            // 1. Whitespace and combinators (>, +, ~, ||)
            if (ctype_space($ch) || $ch === '>' || $ch === '+' || $ch === '~' || $ch === '|') {
                $expectElement = true;
                $i++;
                continue;
            }

            // 2. ID selector: #id
            if ($ch === '#') {
                $a++;
                $expectElement = false;
                $i++;
                while ($i < $n && self::isIdentifierChar($sel[$i])) {
                    $i++;
                }
                continue;
            }

            // 3. Class selector: .class
            if ($ch === '.') {
                $b++;
                $expectElement = false;
                $i++;
                while ($i < $n && self::isIdentifierChar($sel[$i])) {
                    $i++;
                }
                continue;
            }

            // 4. Attribute selector: [attr] or [attr=val]
            if ($ch === '[') {
                $b++;
                $expectElement = false;
                $i++;
                while ($i < $n && $sel[$i] !== ']') {
                    if ($sel[$i] === '"' || $sel[$i] === "'") {
                        $quote = $sel[$i];
                        $i++;
                        while ($i < $n && $sel[$i] !== $quote) {
                            if ($sel[$i] === '\\' && $i + 1 < $n) {
                                $i += 2;
                                continue;
                            }
                            $i++;
                        }
                        if ($i < $n) {
                            $i++; // skip closing quote
                        }
                        continue;
                    }
                    $i++;
                }
                if ($i < $n && $sel[$i] === ']') {
                    $i++;
                }
                continue;
            }

            // 5. Pseudo-classes and pseudo-elements: :pseudo or ::pseudo
            if ($ch === ':') {
                $expectElement = false;
                $i++;
                if ($i < $n && $sel[$i] === ':') {
                    // Pseudo-element: ::before, ::after, etc.
                    $c++;
                    $i++;
                    while ($i < $n && self::isIdentifierChar($sel[$i])) {
                        $i++;
                    }
                    continue;
                }

                // Read pseudo identifier
                $start = $i;
                while ($i < $n && self::isIdentifierChar($sel[$i])) {
                    $i++;
                }
                $pseudoName = strtolower(substr($sel, $start, $i - $start));

                // Legacy pseudo-elements with single colon: :before, :after, :first-line, :first-letter
                if (in_array($pseudoName, ['before', 'after', 'first-line', 'first-letter'], true)) {
                    $c++;
                    continue;
                }

                // Handle functional pseudo-class arguments e.g. :not(div.alert)
                if ($i < $n && $sel[$i] === '(') {
                    $argStart = $i + 1;
                    $parenDepth = 1;
                    $i++;
                    while ($i < $n && $parenDepth > 0) {
                        if ($sel[$i] === '(') {
                            $parenDepth++;
                        } elseif ($sel[$i] === ')') {
                            $parenDepth--;
                        }
                        $i++;
                    }
                    $arg = substr($sel, $argStart, ($i - 1) - $argStart);

                    if (in_array($pseudoName, ['not', 'is', 'has'], true)) {
                        $innerSpec = self::compute($arg);
                        $a += $innerSpec->a;
                        $b += $innerSpec->b;
                        $c += $innerSpec->c;
                        continue;
                    }
                    if ($pseudoName === 'where') {
                        // :where() always has 0 specificity
                        continue;
                    }
                }

                // Normal pseudo-class: :hover, :focus, etc.
                $b++;
                continue;
            }

            // 6. Element / type selector
            if ($expectElement && (ctype_alpha($ch) || $ch === '_')) {
                $c++;
                $expectElement = false;
                while ($i < $n && self::isIdentifierChar($sel[$i])) {
                    $i++;
                }
                continue;
            }

            // Universal selector *
            if ($ch === '*') {
                $expectElement = false;
                $i++;
                continue;
            }

            $i++;
        }

        return new self($a, $b, $c);
    }

    private static function isIdentifierChar(string $ch): bool
    {
        $o = ord($ch);
        return ($o >= 97 && $o <= 122) // a-z
            || ($o >= 65 && $o <= 90)  // A-Z
            || ($o >= 48 && $o <= 57)  // 0-9
            || $ch === '-' || $ch === '_';
    }
}
