package termstrap

import (
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	termimage "github.com/termstrap/termstrap/image"
)

// Syntax token CSS custom properties. Themes override these in their stylesheet
// to align code colorization with their graphic charter; the defaults below are
// tuned to Bootstrap's palette on the dark <pre> background. Every variable is
// also declared in css/termstrap.css so it is discoverable and overridable via
// WithStylesheets(...).
const (
	varCodeName      = "--ts-code-name"      // identifiers / default text
	varCodeKeyword   = "--ts-code-keyword"   // language keywords
	varCodeFunction  = "--ts-code-function"  // function / method names
	varCodeType      = "--ts-code-type"      // class / type names
	varCodeBuiltin   = "--ts-code-builtin"   // builtins / pseudo-builtins
	varCodeString    = "--ts-code-string"    // string literals
	varCodeNumber    = "--ts-code-number"    // numeric literals
	varCodeComment   = "--ts-code-comment"   // comments
	varCodeOperator  = "--ts-code-operator"  // operators / punctuation
	varCodeVariable  = "--ts-code-variable"  // variables / properties / labels
	varCodeConstant  = "--ts-code-constant"  // constants / enum members / symbols
	varCodeTag       = "--ts-code-tag"       // markup tags (HTML/XML)
	varCodeAttribute = "--ts-code-attribute" // markup / decorator attributes
	varCodeRegex     = "--ts-code-regex"     // regular expressions
	varCodeEscape    = "--ts-code-escape"    // escape sequences / interpolation
	varCodeDeleted   = "--ts-code-deleted"   // diff deletions / errors
	varCodeInserted  = "--ts-code-inserted"  // diff insertions
	varCodeHeading   = "--ts-code-heading"   // markup headings / generic emphasis
)

// defaultCodePalette is the Bootstrap-flavored fallback palette (dark <pre> bg
// #212529). It is used whenever a CSS variable is absent so highlighting still
// works without any theme CSS. Kept in sync with the :root block in
// css/termstrap.css.
var defaultCodePalette = map[string]string{
	varCodeName:      "#f8f9fa",
	varCodeKeyword:   "#e685b5",
	varCodeFunction:  "#6ea8fe",
	varCodeType:      "#6edff6",
	varCodeBuiltin:   "#c29ffa",
	varCodeString:    "#75b798",
	varCodeNumber:    "#ffda6a",
	varCodeComment:   "#9aa0a6",
	varCodeOperator:  "#dee2e6",
	varCodeVariable:  "#f8f9fa",
	varCodeConstant:  "#c29ffa",
	varCodeTag:       "#e685b5",
	varCodeAttribute: "#6ea8fe",
	varCodeRegex:     "#63e6be",
	varCodeEscape:    "#ffbf69",
	varCodeDeleted:   "#ea868f",
	varCodeInserted:  "#75b798",
	varCodeHeading:   "#6ea8fe",
}

// resolveLang extracts a source language hint from a <pre> selection following
// the common HTML/Markdown convention, in priority order:
//  1. a child <code class="language-xxx"> or "lang-xxx" (GitHub/CommonMark)
//  2. a class on the <pre> itself (language-xxx / lang-xxx / highlight-xxx)
//
// The returned value is the raw hint (e.g. "go", "python"); an empty string
// means "no explicit language" and leaves room for content auto-detection.
func resolveLang(classAttrs ...string) string {
	for _, raw := range classAttrs {
		for _, cls := range strings.Fields(raw) {
			lower := strings.ToLower(cls)
			for _, prefix := range []string{"language-", "lang-", "highlight-", "sourcecode-"} {
				if strings.HasPrefix(lower, prefix) {
					if l := strings.TrimPrefix(lower, prefix); l != "" {
						return l
					}
				}
			}
		}
	}
	return ""
}

// pickLexer selects a chroma lexer from an explicit language hint, falling back
// to content analysis when the hint is empty/unknown. Returns nil only when no
// lexer applies at all (caller then skips highlighting).
func pickLexer(lang, code string) chroma.Lexer {
	lang = strings.ToLower(strings.TrimSpace(lang))
	if lang != "" && lang != "auto" && lang != "text" && lang != "plaintext" && lang != "none" {
		if lx := lexers.Get(lang); lx != nil {
			return lx
		}
	}
	// Auto-detection by content when no usable hint was provided.
	if lang == "" || lang == "auto" {
		if lx := lexers.Analyse(code); lx != nil {
			return lx
		}
	}
	return nil
}

// pickFormatter maps the model color depth to a chroma terminal formatter.
func pickFormatter(mode termimage.ColorMode) chroma.Formatter {
	switch mode {
	case termimage.ColorMode16:
		return formatters.TTY16
	case termimage.ColorMode256:
		return formatters.TTY256
	default: // ColorModeTrueColor and anything else
		return formatters.TTY16m
	}
}

// codeColor resolves a syntax token color from the resolved CSS variables,
// falling back to the Bootstrap default palette.
func codeColor(vars map[string]string, name string) string {
	if vars != nil {
		if v, ok := vars[name]; ok {
			if v = strings.TrimSpace(v); v != "" {
				return v
			}
		}
	}
	return defaultCodePalette[name]
}

// buildCodeStyle constructs a chroma style from the resolved CSS variables so
// each theme colorizes code per its own charter (see css/theme-*.css). chroma's
// formatters fall back from a specific token to its parent category, so mapping
// the categories plus a few salient sub-tokens covers every lexer.
func buildCodeStyle(vars map[string]string) *chroma.Style {
	c := func(name string) string { return codeColor(vars, name) }
	entries := chroma.StyleEntries{
		// Base
		chroma.Background: c(varCodeName),
		chroma.Text:       c(varCodeName),

		// Keywords
		chroma.Keyword:            c(varCodeKeyword) + " bold",
		chroma.KeywordConstant:    c(varCodeConstant),
		chroma.KeywordType:        c(varCodeType),
		chroma.KeywordDeclaration: c(varCodeKeyword) + " bold",
		chroma.KeywordNamespace:   c(varCodeKeyword) + " bold",
		chroma.KeywordPseudo:      c(varCodeKeyword),

		// Names
		chroma.Name:               c(varCodeName),
		chroma.NameFunction:       c(varCodeFunction),
		chroma.NameFunctionMagic:  c(varCodeFunction),
		chroma.NameClass:          c(varCodeType),
		chroma.NameNamespace:      c(varCodeType),
		chroma.NameException:      c(varCodeType),
		chroma.NameBuiltin:        c(varCodeBuiltin),
		chroma.NameBuiltinPseudo:  c(varCodeBuiltin),
		chroma.NameConstant:       c(varCodeConstant),
		chroma.NameVariable:       c(varCodeVariable),
		chroma.NameVariableGlobal: c(varCodeVariable),
		chroma.NameProperty:       c(varCodeVariable),
		chroma.NameLabel:          c(varCodeVariable),
		chroma.NameTag:            c(varCodeTag),
		chroma.NameAttribute:      c(varCodeAttribute),
		chroma.NameDecorator:      c(varCodeAttribute),
		chroma.NameEntity:         c(varCodeEscape),

		// Literals
		chroma.LiteralString:         c(varCodeString),
		chroma.LiteralStringDoc:      c(varCodeComment) + " italic",
		chroma.LiteralStringEscape:   c(varCodeEscape),
		chroma.LiteralStringInterpol: c(varCodeEscape),
		chroma.LiteralStringRegex:    c(varCodeRegex),
		chroma.LiteralStringSymbol:   c(varCodeConstant),
		chroma.LiteralNumber:         c(varCodeNumber),

		// Operators & punctuation
		chroma.Operator:     c(varCodeOperator),
		chroma.OperatorWord: c(varCodeKeyword),
		chroma.Punctuation:  c(varCodeOperator),

		// Comments
		chroma.Comment:         c(varCodeComment) + " italic",
		chroma.CommentSpecial:  c(varCodeComment) + " bold italic",
		chroma.CommentPreproc:  c(varCodeTag),
		chroma.CommentHashbang: c(varCodeComment) + " italic",

		// Generic (diffs, markup, REPL output)
		chroma.GenericDeleted:    c(varCodeDeleted),
		chroma.GenericInserted:   c(varCodeInserted),
		chroma.GenericError:      c(varCodeDeleted) + " bold",
		chroma.GenericTraceback:  c(varCodeDeleted),
		chroma.GenericHeading:    c(varCodeHeading) + " bold",
		chroma.GenericSubheading: c(varCodeHeading),
		chroma.GenericEmph:       "italic",
		chroma.GenericStrong:     "bold",
		chroma.GenericPrompt:     c(varCodeComment),
		chroma.Error:             c(varCodeDeleted),
	}
	if st, err := chroma.NewStyle("termstrap", entries); err == nil {
		return st
	}
	return styles.Fallback
}

// resolveCodeStyle returns the chroma style used for highlighting. An explicit
// SyntaxTheme (a named chroma style such as "monokai" or "dracula") takes
// precedence as a manual escape hatch; otherwise colors are derived from the
// theme's --ts-code-* CSS variables.
func resolveCodeStyle(override string, vars map[string]string) *chroma.Style {
	if name := strings.TrimSpace(override); name != "" {
		if st := styles.Get(name); st != nil && strings.EqualFold(st.Name, name) {
			return st
		}
	}
	return buildCodeStyle(vars)
}

// highlightCode syntax-colorizes source into ANSI escape sequences using the
// pre-resolved style. The terminal formatters emit foreground colors and text
// attributes only (chroma clears token backgrounds), so the surrounding <pre>
// background applied by lipgloss — and repaired by persistBackgroundColor across
// the \x1b[0m resets chroma emits — remains intact.
//
// On any failure (unknown language with no detectable match, tokenise error)
// the original code is returned unchanged so the block still renders.
func highlightCode(code, lang string, mode termimage.ColorMode, style *chroma.Style) string {
	if strings.TrimSpace(code) == "" {
		return code
	}
	if style == nil {
		style = styles.Fallback
	}
	lexer := pickLexer(lang, code)
	if lexer == nil {
		return code
	}
	lexer = chroma.Coalesce(lexer)

	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		return code
	}

	var sb strings.Builder
	if err := pickFormatter(mode).Format(&sb, style, iterator); err != nil {
		return code
	}
	return sb.String()
}
