package termstrap

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode"

	"github.com/aymerick/douceur/css"
)

// CSSPreprocessor handles CSS Custom Properties (var(--*)), calc() arithmetic,
// and rgba/rgb color resolution before selectors and declarations are matched.
type CSSPreprocessor struct {
	Theme      Theme
	GlobalVars map[string]string
}

// NewCSSPreprocessor creates a new CSS preprocessor configured for the given theme.
func NewCSSPreprocessor(theme Theme) *CSSPreprocessor {
	return &CSSPreprocessor{
		Theme:      theme,
		GlobalVars: make(map[string]string),
	}
}

// ExtractVariables extracts root-level custom properties from stylesheet rules.
func (p *CSSPreprocessor) ExtractVariables(sheet *css.Stylesheet) {
	if sheet == nil {
		return
	}

	themeStr := string(p.Theme)
	if themeStr == "" {
		themeStr = "bootstrap"
	}

	for _, rule := range sheet.Rules {
		isRootScope := false
		for _, sel := range rule.Selectors {
			s := strings.TrimSpace(sel)
			if s == ":root" || s == "html" || s == "body" || s == "*" {
				isRootScope = true
				break
			}
			// Match [data-bs-theme=light] or [data-bs-theme=dark] depending on active theme
			if strings.HasPrefix(s, "[data-bs-theme=") {
				target := strings.TrimSuffix(strings.TrimPrefix(s, "[data-bs-theme="), "]")
				target = strings.Trim(target, `"'`)
				if strings.EqualFold(target, themeStr) {
					isRootScope = true
					break
				}
				if themeStr == "light" && target == "light" {
					isRootScope = true
					break
				} else if (themeStr == "bootstrap" || themeStr == "dark" || themeStr == "none" || themeStr == "") && target == "dark" {
					isRootScope = true
					break
				}
			}
		}

		if isRootScope {
			for _, decl := range rule.Declarations {
				prop := strings.TrimSpace(decl.Property)
				if strings.HasPrefix(prop, "--") {
					val := strings.TrimSpace(decl.Value)
					p.GlobalVars[prop] = val
				}
			}
		}
	}

	// Recursively resolve references within global variables (e.g. --a: var(--b))
	p.resolveGlobalVarDependencies()
}

// resolveGlobalVarDependencies resolves variables that reference other variables in GlobalVars.
func (p *CSSPreprocessor) resolveGlobalVarDependencies() {
	lookup := func(name string) (string, bool) {
		v, ok := p.GlobalVars[name]
		return v, ok
	}

	// Iterate multiple passes until stable or max passes reached (cycle guard)
	for range 10 {
		changed := false
		for k, v := range p.GlobalVars {
			if strings.Contains(v, "var(") {
				resolved := ResolveVariables(v, lookup)
				if resolved != v {
					p.GlobalVars[k] = resolved
					changed = true
				}
			}
		}
		if !changed {
			break
		}
	}
}

// PreprocessSheet resolves variables, calc(), and rgba() in all declarations of a stylesheet.
func (p *CSSPreprocessor) PreprocessSheet(sheet *css.Stylesheet) {
	if sheet == nil {
		return
	}
	for _, rule := range sheet.Rules {
		p.preprocessRule(rule)
	}
}
// PreprocessDeclarations resolves variables, calc(), and rgba() in a slice of declarations (e.g. from inline styles).
func (p *CSSPreprocessor) PreprocessDeclarations(decls []*css.Declaration) {
	if p == nil || len(decls) == 0 {
		return
	}

	localVars := make(map[string]string)
	for _, decl := range decls {
		prop := strings.TrimSpace(decl.Property)
		if strings.HasPrefix(prop, "--") {
			localVars[prop] = strings.TrimSpace(decl.Value)
		}
	}

	lookup := func(name string) (string, bool) {
		if v, ok := localVars[name]; ok {
			return v, true
		}
		if v, ok := p.GlobalVars[name]; ok {
			return v, true
		}
		return "", false
	}

	bgHex := "#ffffff"
	if p.Theme == ThemeDracula || p.Theme == ThemeTokyoNight || p.Theme == ThemeGlamour {
		bgHex = "#1a1b26"
	}

	for _, decl := range decls {
		val := strings.TrimSpace(decl.Value)

		// 1. Resolve var(--*)
		if strings.Contains(val, "var(") {
			val = ResolveVariables(val, lookup)
		}

		// 2. Evaluate calc(...)
		if strings.Contains(strings.ToLower(val), "calc(") {
			val = EvaluateCalc(val)
		}

		// 3. Resolve rgba(...) and rgb(...)
		valLower := strings.ToLower(val)
		if strings.Contains(valLower, "rgba(") || strings.Contains(valLower, "rgb(") {
			val = ResolveColors(val, bgHex)
		}

		decl.Value = val
	}
}


func (p *CSSPreprocessor) preprocessRule(rule *css.Rule) {
	if rule == nil {
		return
	}

	// Collect local custom properties defined inside this specific rule
	localVars := make(map[string]string)
	for _, decl := range rule.Declarations {
		prop := strings.TrimSpace(decl.Property)
		if strings.HasPrefix(prop, "--") {
			localVars[prop] = strings.TrimSpace(decl.Value)
		}
	}

	lookup := func(name string) (string, bool) {
		if v, ok := localVars[name]; ok {
			return v, true
		}
		if v, ok := p.GlobalVars[name]; ok {
			return v, true
		}
		return "", false
	}

	// Resolve local vars that reference others
	for k, v := range localVars {
		if strings.Contains(v, "var(") {
			localVars[k] = ResolveVariables(v, lookup)
		}
	}

	// Preprocess all declarations in this rule
	for _, decl := range rule.Declarations {
		val := strings.TrimSpace(decl.Value)

		// 1. Resolve var(--*)
		if strings.Contains(val, "var(") {
			val = ResolveVariables(val, lookup)
		}

		// 2. Evaluate calc(...)
		if strings.Contains(strings.ToLower(val), "calc(") {
			val = EvaluateCalc(val)
		}

		// 3. Resolve rgba(...) and rgb(...)
		valLower := strings.ToLower(val)
		if strings.Contains(valLower, "rgba(") || strings.Contains(valLower, "rgb(") {
			bgHex := "#ffffff"
			if p.Theme == ThemeDracula || p.Theme == ThemeTokyoNight || p.Theme == ThemeGlamour {
				bgHex = "#1a1b26"
			}
			val = ResolveColors(val, bgHex)
		}

		decl.Value = val
	}

	// Recursively process nested rules (such as in @media)
	for _, childRule := range rule.Rules {
		p.preprocessRule(childRule)
	}
}

// ResolveVariables replaces all var(--name, fallback) occurrences in value.
func ResolveVariables(value string, lookup func(string) (string, bool)) string {
	return resolveVariablesWithDepth(value, lookup, nil)
}

func resolveVariablesWithDepth(value string, lookup func(string) (string, bool), visited map[string]bool) string {
	if !strings.Contains(value, "var(") {
		return value
	}

	var sb strings.Builder
	idx := 0
	n := len(value)

	for idx < n {
		pos := strings.Index(value[idx:], "var(")
		if pos == -1 {
			sb.WriteString(value[idx:])
			break
		}

		varStart := idx + pos
		sb.WriteString(value[idx:varStart])

		// Find matching closing parenthesis
		contentStart := varStart + 4 // len("var(")
		parenCount := 1
		closeIdx := -1

		for i := contentStart; i < n; i++ {
			if value[i] == '(' {
				parenCount++
			} else if value[i] == ')' {
				parenCount--
				if parenCount == 0 {
					closeIdx = i
					break
				}
			}
		}

		if closeIdx == -1 {
			// Malformed var(), copy verbatim
			sb.WriteString(value[varStart:])
			break
		}

		inside := value[contentStart:closeIdx]
		varName, fallback := splitVarFallback(inside)
		varName = strings.TrimSpace(varName)

		resolved := ""
		found := false

		// Check cycle
		if visited == nil {
			visited = make(map[string]bool)
		}

		if !visited[varName] {
			if val, ok := lookup(varName); ok {
				visited[varName] = true
				resolved = resolveVariablesWithDepth(val, lookup, visited)
				delete(visited, varName)
				found = true
			}
		}

		if !found && fallback != "" {
			resolved = resolveVariablesWithDepth(fallback, lookup, visited)
		}

		sb.WriteString(resolved)
		idx = closeIdx + 1
	}

	return sb.String()
}

// splitVarFallback splits "name, fallback" taking nested parentheses into account.
func splitVarFallback(content string) (string, string) {
	parenCount := 0
	for i := range content {
		if content[i] == '(' {
			parenCount++
		} else if content[i] == ')' {
			parenCount--
		} else if content[i] == ',' && parenCount == 0 {
			return content[:i], strings.TrimSpace(content[i+1:])
		}
	}
	return content, ""
}

// EvaluateCalc resolves calc(...) mathematical expressions.
func EvaluateCalc(input string) string {
	lower := strings.ToLower(input)
	if !strings.Contains(lower, "calc(") {
		return input
	}

	var sb strings.Builder
	idx := 0
	n := len(input)

	for idx < n {
		pos := strings.Index(strings.ToLower(input[idx:]), "calc(")
		if pos == -1 {
			sb.WriteString(input[idx:])
			break
		}

		calcStart := idx + pos
		sb.WriteString(input[idx:calcStart])

		contentStart := calcStart + 5 // len("calc(")
		parenCount := 1
		closeIdx := -1

		for i := contentStart; i < n; i++ {
			if input[i] == '(' {
				parenCount++
			} else if input[i] == ')' {
				parenCount--
				if parenCount == 0 {
					closeIdx = i
					break
				}
			}
		}

		if closeIdx == -1 {
			sb.WriteString(input[calcStart:])
			break
		}

		inside := input[contentStart:closeIdx]
		res := evalCalcExpression(inside)
		sb.WriteString(res)
		idx = closeIdx + 1
	}

	return sb.String()
}

// evalCalcExpression evaluates simple arithmetic operations: e.g. "1.5rem * 0.5", "-1 * 2px".
func evalCalcExpression(expr string) string {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return "0"
	}

	// Tokenize numbers, units, and operators
	type token struct {
		isOp bool
		op   byte
		num  float64
		unit string
	}

	var tokens []token
	i := 0
	m := len(expr)

	for i < m {
		ch := expr[i]
		if unicode.IsSpace(rune(ch)) {
			i++
			continue
		}

		// Operators
		if ch == '*' || ch == '/' || ch == '+' {
			tokens = append(tokens, token{isOp: true, op: ch})
			i++
			continue
		}

		// '-' can be operator or negative number sign
		if ch == '-' {
			// If previous token was a number/unit, '-' is a subtraction operator
			if len(tokens) > 0 && !tokens[len(tokens)-1].isOp {
				tokens = append(tokens, token{isOp: true, op: '-'})
				i++
				continue
			}
		}

		// Number parsing
		start := i
		if ch == '-' || ch == '+' {
			i++
		}
		hasDigit := false
		for i < m && ((expr[i] >= '0' && expr[i] <= '9') || expr[i] == '.') {
			hasDigit = true
			i++
		}

		if hasDigit {
			numStr := expr[start:i]
			num, _ := strconv.ParseFloat(numStr, 64)

			// Extract unit
			unitStart := i
			for i < m && (unicode.IsLetter(rune(expr[i])) || expr[i] == '%') {
				i++
			}
			unit := expr[unitStart:i]

			tokens = append(tokens, token{isOp: false, num: num, unit: unit})
			continue
		}

		// Skip unhandled characters
		i++
	}

	if len(tokens) == 0 {
		return expr
	}

	// 1. First pass: Handle multiplication and division (*, /)
	var pass1 []token
	idx := 0
	for idx < len(tokens) {
		tok := tokens[idx]
		if tok.isOp && (tok.op == '*' || tok.op == '/') && len(pass1) > 0 && idx+1 < len(tokens) && !tokens[idx+1].isOp {
			left := pass1[len(pass1)-1]
			right := tokens[idx+1]
			unit := left.unit
			if unit == "" {
				unit = right.unit
			}

			var val float64
			if tok.op == '*' {
				val = left.num * right.num
			} else if right.num != 0 {
				val = left.num / right.num
			}

			pass1[len(pass1)-1] = token{isOp: false, num: val, unit: unit}
			idx += 2
			continue
		}
		pass1 = append(pass1, tok)
		idx++
	}

	// 2. Second pass: Handle addition and subtraction (+, -)
	if len(pass1) == 0 {
		return expr
	}

	acc := pass1[0]
	idx = 1
	for idx+1 < len(pass1) {
		opTok := pass1[idx]
		rightTok := pass1[idx+1]
		if opTok.isOp && !rightTok.isOp {
			unit := acc.unit
			if unit == "" {
				unit = rightTok.unit
			}
			if opTok.op == '+' {
				acc.num += rightTok.num
			} else if opTok.op == '-' {
				acc.num -= rightTok.num
			}
			acc.unit = unit
			idx += 2
		} else {
			idx++
		}
	}

	// Format result
	if math.Abs(acc.num) < 0.0001 {
		acc.num = 0
	}
	resStr := strconv.FormatFloat(acc.num, 'f', -1, 64)
	return resStr + acc.unit
}

// ResolveColors replaces rgba(...) and rgb(...) in text with hex colors (#rrggbb).
func ResolveColors(input string, bgHex string) string {
	lower := strings.ToLower(input)
	if !strings.Contains(lower, "rgb") {
		return input
	}

	var sb strings.Builder
	idx := 0
	n := len(input)

	for idx < n {
		sub := strings.ToLower(input[idx:])
		posRGBA := strings.Index(sub, "rgba(")
		posRGB := strings.Index(sub, "rgb(")

		pos := -1
		isRGBA := false
		if posRGBA != -1 && (posRGB == -1 || posRGBA <= posRGB) {
			pos = posRGBA
			isRGBA = true
		} else if posRGB != -1 {
			pos = posRGB
			isRGBA = false
		}

		if pos == -1 {
			sb.WriteString(input[idx:])
			break
		}

		callStart := idx + pos
		sb.WriteString(input[idx:callStart])

		fnLen := 4
		if isRGBA {
			fnLen = 5
		}

		contentStart := callStart + fnLen
		parenCount := 1
		closeIdx := -1

		for i := contentStart; i < n; i++ {
			if input[i] == '(' {
				parenCount++
			} else if input[i] == ')' {
				parenCount--
				if parenCount == 0 {
					closeIdx = i
					break
				}
			}
		}

		if closeIdx == -1 {
			sb.WriteString(input[callStart:])
			break
		}

		inside := input[contentStart:closeIdx]
		r, g, b, a, ok := ParseRGBA(inside)
		if ok {
			hex := BlendRGBA(r, g, b, a, bgHex)
			sb.WriteString(hex)
		} else {
			sb.WriteString(input[callStart : closeIdx+1])
		}

		idx = closeIdx + 1
	}

	return sb.String()
}

// ParseRGBA parses components from strings like "13, 110, 253, 1" or "13 110 253 / 0.5".
func ParseRGBA(str string) (r, g, b uint8, a float64, ok bool) {
	str = strings.ReplaceAll(str, "/", ",")
	rawParts := strings.Split(str, ",")
	var parts []string
	for _, p := range rawParts {
		f := strings.Fields(strings.TrimSpace(p))
		parts = append(parts, f...)
	}

	if len(parts) < 3 {
		return 0, 0, 0, 0, false
	}

	parseChan := func(s string) (uint8, bool) {
		s = strings.TrimSpace(s)
		s = strings.TrimSuffix(s, "%")
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return 0, false
		}
		if v < 0 {
			v = 0
		}
		if v > 255 {
			v = 255
		}
		return uint8(math.Round(v)), true
	}

	rVal, okR := parseChan(parts[0])
	gVal, okG := parseChan(parts[1])
	bVal, okB := parseChan(parts[2])
	if !okR || !okG || !okB {
		return 0, 0, 0, 0, false
	}

	alpha := 1.0
	if len(parts) >= 4 {
		s := strings.TrimSpace(parts[3])
		if strings.HasSuffix(s, "%") {
			s = strings.TrimSuffix(s, "%")
			if v, err := strconv.ParseFloat(s, 64); err == nil {
				alpha = v / 100.0
			}
		} else {
			if v, err := strconv.ParseFloat(s, 64); err == nil {
				alpha = v
			}
		}
		if alpha < 0 {
			alpha = 0
		}
		if alpha > 1 {
			alpha = 1
		}
	}

	return rVal, gVal, bVal, alpha, true
}

// BlendRGBA blends an RGBA color over a background hex color (#rrggbb).
func BlendRGBA(r, g, b uint8, a float64, bgHex string) string {
	if a >= 0.999 {
		return fmt.Sprintf("#%02x%02x%02x", r, g, b)
	}
	if a <= 0.001 {
		return "transparent"
	}

	bgR, bgG, bgB := parseHexRGB(bgHex)
	blendR := uint8(math.Round(float64(r)*a + float64(bgR)*(1.0-a)))
	blendG := uint8(math.Round(float64(g)*a + float64(bgG)*(1.0-a)))
	blendB := uint8(math.Round(float64(b)*a + float64(bgB)*(1.0-a)))

	return fmt.Sprintf("#%02x%02x%02x", blendR, blendG, blendB)
}

func parseHexRGB(hex string) (uint8, uint8, uint8) {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) == 3 {
		hex = string([]byte{hex[0], hex[0], hex[1], hex[1], hex[2], hex[2]})
	}
	if len(hex) == 6 {
		val, err := strconv.ParseUint(hex, 16, 32)
		if err == nil {
			return uint8(val >> 16), uint8((val >> 8) & 0xff), uint8(val & 0xff)
		}
	}
	// Default to white
	return 255, 255, 255
}
