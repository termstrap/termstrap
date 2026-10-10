package termstrap

import (
	"strings"
	"unicode"
)

// Specificity represents the W3C CSS specificity tuple (A, B, C).
// A = ID selectors count
// B = Class selectors, attribute selectors, and pseudo-classes count
// C = Element (type) selectors and pseudo-elements count
type Specificity struct {
	A int // IDs
	B int // Classes, attributes, pseudo-classes
	C int // Elements, pseudo-elements
}

// Compare compares two specificities.
// Returns a positive integer if s > other, negative if s < other, and 0 if equal.
func (s Specificity) Compare(other Specificity) int {
	if s.A != other.A {
		return s.A - other.A
	}
	if s.B != other.B {
		return s.B - other.B
	}
	return s.C - other.C
}

// Less reports whether s is strictly less specific than other.
func (s Specificity) Less(other Specificity) bool {
	return s.Compare(other) < 0
}

// ComputeSpecificity calculates the specificity tuple (A, B, C) for a given CSS selector.
func ComputeSpecificity(sel string) Specificity {
	sel = strings.TrimSpace(sel)
	if sel == "" || sel == "*" {
		return Specificity{}
	}

	var spec Specificity
	n := len(sel)
	i := 0

	// Expecting an element type selector when starting or immediately after a combinator/whitespace
	expectElement := true

	for i < n {
		ch := sel[i]

		// 1. Whitespace and combinators (>, +, ~, ||)
		if unicode.IsSpace(rune(ch)) || ch == '>' || ch == '+' || ch == '~' || ch == '|' {
			expectElement = true
			i++
			continue
		}

		// 2. ID selector: #id
		if ch == '#' {
			spec.A++
			expectElement = false
			i++
			for i < n && isIdentifierChar(sel[i]) {
				i++
			}
			continue
		}

		// 3. Class selector: .class
		if ch == '.' {
			spec.B++
			expectElement = false
			i++
			for i < n && isIdentifierChar(sel[i]) {
				i++
			}
			continue
		}

		// 4. Attribute selector: [attr] or [attr=val]
		if ch == '[' {
			spec.B++
			expectElement = false
			i++
			for i < n && sel[i] != ']' {
				if sel[i] == '"' || sel[i] == '\'' {
					quote := sel[i]
					i++
					for i < n && sel[i] != quote {
						if sel[i] == '\\' && i+1 < n {
							i += 2
							continue
						}
						i++
					}
					if i < n {
						i++ // skip closing quote
					}
					continue
				}
				i++
			}
			if i < n && sel[i] == ']' {
				i++
			}
			continue
		}

		// 5. Pseudo-classes and pseudo-elements: :pseudo or ::pseudo
		if ch == ':' {
			expectElement = false
			i++
			if i < n && sel[i] == ':' {
				// Pseudo-element: ::before, ::after, etc.
				spec.C++
				i++
				for i < n && isIdentifierChar(sel[i]) {
					i++
				}
				continue
			}

			// Read pseudo identifier
			start := i
			for i < n && isIdentifierChar(sel[i]) {
				i++
			}
			pseudoName := strings.ToLower(sel[start:i])

			// Legacy pseudo-elements with single colon: :before, :after, :first-line, :first-letter
			if pseudoName == "before" || pseudoName == "after" || pseudoName == "first-line" || pseudoName == "first-letter" {
				spec.C++
				continue
			}

			// Handle functional pseudo-class arguments e.g. :not(div.alert)
			if i < n && sel[i] == '(' {
				argStart := i + 1
				parenDepth := 1
				i++
				for i < n && parenDepth > 0 {
					if sel[i] == '(' {
						parenDepth++
					} else if sel[i] == ')' {
						parenDepth--
					}
					i++
				}
				arg := sel[argStart : i-1]

				// :not(X) does not count itself, but its argument counts
				if pseudoName == "not" || pseudoName == "is" || pseudoName == "has" {
					innerSpec := ComputeSpecificity(arg)
					spec.A += innerSpec.A
					spec.B += innerSpec.B
					spec.C += innerSpec.C
					continue
				}
				if pseudoName == "where" {
					// :where() always has 0 specificity
					continue
				}
			}

			// Normal pseudo-class: :hover, :focus, :first-child, :disabled, etc.
			spec.B++
			continue
		}

		// 6. Element / type selector (e.g. div, h1, span, p)
		if expectElement && (unicode.IsLetter(rune(ch)) || ch == '_') {
			spec.C++
			expectElement = false
			for i < n && isIdentifierChar(sel[i]) {
				i++
			}
			continue
		}

		// Universal selector *
		if ch == '*' {
			expectElement = false
			i++
			continue
		}

		i++
	}

	return spec
}

func isIdentifierChar(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') ||
		(ch >= 'A' && ch <= 'Z') ||
		(ch >= '0' && ch <= '9') ||
		ch == '-' || ch == '_'
}
