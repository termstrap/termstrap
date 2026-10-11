package termstrap

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/andybalholm/cascadia"
	"github.com/aymerick/douceur/css"
	"github.com/aymerick/douceur/parser"
	"charm.land/lipgloss/v2"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
)

var (
	builtInMatcherMu    sync.RWMutex
	builtInMatcherCache = make(map[Theme]*CSSMatcher)
)

// DisplayType specifies the display behavior of a node.
type DisplayType string

const (
	DisplayBlock  DisplayType = "block"
	DisplayInline DisplayType = "inline"
	DisplayFlex   DisplayType = "flex"
	DisplayNone   DisplayType = "none"
)

// Breakpoint represents a screen size category.
type Breakpoint int

const (
	bpXS  Breakpoint = iota // < 70
	bpSM                    // >= 70
	bpMD                    // >= 100
	bpLG                    // >= 130
	bpXL                    // >= 160
	bpXXL                   // >= 190
)

func detectBreakpoint(width int) Breakpoint {
	switch {
	case width >= 190:
		return bpXXL
	case width >= 160:
		return bpXL
	case width >= 130:
		return bpLG
	case width >= 100:
		return bpMD
	case width >= 70:
		return bpSM
	default:
		return bpXS
	}
}

// ComputedStyle holds the resolved visual properties for a DOM node.
type ComputedStyle struct {
	Display DisplayType

	// Margins (terminal cells/lines)
	MarginTop    int
	MarginBottom int
	MarginLeft   int
	MarginRight  int

	// Paddings (terminal cells/lines)
	PaddingTop    int
	PaddingBottom int
	PaddingLeft   int
	PaddingRight  int

	// Borders
	Border       bool
	BorderTop    bool
	BorderBottom bool
	BorderLeft   bool
	BorderRight  bool
	BorderColor  string
	BorderStyle  string // "", "solid", "double", "dashed", "dotted", "hidden", "thick"
	Rounded      bool
	// Per-side border colors (fall back to BorderColor when empty)
	BorderTopColor    string
	BorderRightColor  string
	BorderBottomColor string
	BorderLeftColor   string
	BorderShimmer     bool
	// Border color blending / gradients
	BorderBlend []string
	// Shadow (0 = none, 1 = sm, 2 = md, 3 = lg)
	Shadow int
	// Colors
	FgColor    string
	FgColorSet bool
	BgColor     string
	BgGradient  []string

	// Typography
	Bold              bool
	BoldSet           bool
	Italic            bool
	ItalicSet         bool
	Underline         bool
	UnderlineSet      bool
	UnderlineStyle    lipgloss.Underline // single, double, curly/wavy, dotted, dashed
	UnderlineStyleSet bool
	UnderlineColor    string
	UnderlineColorSet bool
	Strikethrough     bool
	StrikethroughSet  bool
	Faint             bool
	FaintSet          bool
	Blink             bool
	BlinkSet          bool
	Shimmer           bool
	ShimmerSet        bool
	ShimmerHighlight  string
	Spinner           bool
	SpinnerSet        bool
	TextAlign         lipgloss.Position
	TextAlignSet      bool
	TextTransform     string // "", "uppercase", "lowercase", "capitalize"
	TextTransformSet  bool
	// Flex alignment
	AlignItems        lipgloss.Position // cross-axis (align-items / vertical-align)
	AlignItemsSet     bool
	JustifyContent    lipgloss.Position // main-axis (justify-content)
	JustifyContentSet bool
	VerticalAlign     lipgloss.Position // vertical-align on inline/flex children
	VerticalAlignSet  bool
	// List rendering
	ListStyle string // "", "disc", "circle", "square", "decimal", "lower-alpha", "upper-roman", "none"
	// Grid / Sizing
	ColSpan   int  // 1-12 for flex items
	ColAuto   bool // true for .col without explicit span number
	RowCols   int  // 1-12 items per row from .row-cols-[bp-]N
	Width     int  // explicit width if specified
	MaxWidth  int  // max-width clamp (0 = none)
	MaxHeight int  // max-height clamp (0 = none)
}

// CSSOrigin defines the cascade layer of a stylesheet.
type CSSOrigin int

const (
	OriginTagDefaults CSSOrigin = iota // 0: User-agent default tag styling
	OriginFramework                    // 1: Built-in Bootstrap framework
	OriginTheme                        // 2: Color theme (Bootstrap, Dracula, TokyoNight)
	OriginCustom                       // 3: User custom stylesheets
)

// CSSRule represents a compiled selector and declaration block with specificity.
type CSSRule struct {
	Selector    string
	Compiled    cascadia.Selector
	Decls       []*css.Declaration
	MediaCond   *MediaCondition
	Specificity Specificity
	Origin      CSSOrigin
	Index       int    // Appearance order in source
	Pseudo      string // "", "before", or "after" (pseudo-element target)
}

// CSSMatcher compiles stylesheets into fast matchable rules.
type CSSMatcher struct {
	Rules        []CSSRule
	Preprocessor *CSSPreprocessor
	Theme        Theme
}

// NewCSSMatcher parses one or more CSS stylesheets into a CSSMatcher using the specified theme.
func NewCSSMatcher(theme Theme, stylesheets ...string) (*CSSMatcher, error) {
	// Fast-path: cache built-in framework & theme matcher if no custom stylesheets are supplied
	if len(stylesheets) == 0 {
		builtInMatcherMu.RLock()
		if cached, ok := builtInMatcherCache[theme]; ok {
			builtInMatcherMu.RUnlock()
			return cached, nil
		}
		builtInMatcherMu.RUnlock()
	}

	matcher := &CSSMatcher{Rules: []CSSRule{}, Theme: theme}

	type sheetInput struct {
		content string
		origin  CSSOrigin
	}
	var allSheets []sheetInput

	// 1. Base Framework: complete official Bootstrap CSS
	if bsCSS, err := cssFS.ReadFile("css/bootstrap.css"); err == nil {
		allSheets = append(allSheets, sheetInput{content: string(bsCSS), origin: OriginFramework})
	}

	// 2. TermStrap Custom Terminal Overwrites (replaces tags.css)
	if tsCSS, err := cssFS.ReadFile("css/termstrap.css"); err == nil {
		allSheets = append(allSheets, sheetInput{content: string(tsCSS), origin: OriginFramework})
	}

	// 3. Theme CSS variables (nullable: only load if theme is set and not empty/none/bootstrap)
	themeStr := strings.ToLower(strings.TrimSpace(string(theme)))
	if themeStr != "" && themeStr != "none" && themeStr != "bootstrap" {
		themeFile := "css/theme-" + themeStr + ".css"
		if themeCSS, err := cssFS.ReadFile(themeFile); err == nil && len(themeCSS) > 0 {
			allSheets = append(allSheets, sheetInput{content: string(themeCSS), origin: OriginTheme})
		}
	}
	// 4. Custom stylesheets
	for _, sheetContent := range stylesheets {
		allSheets = append(allSheets, sheetInput{content: sheetContent, origin: OriginCustom})
	}

	preproc := NewCSSPreprocessor(theme)

	type parsedSheet struct {
		sheet  *css.Stylesheet
		origin CSSOrigin
	}
	var parsed []parsedSheet

	for _, item := range allSheets {
		if strings.TrimSpace(item.content) == "" {
			continue
		}
		sheet, err := parser.Parse(item.content)
		if err != nil {
			continue
		}
		parsed = append(parsed, parsedSheet{sheet: sheet, origin: item.origin})
		preproc.ExtractVariables(sheet)
	}

	matcher.Preprocessor = preproc
	globalIndex := 0

	var addRule func(rule *css.Rule, parentMedia *MediaCondition, origin CSSOrigin)
	addRule = func(rule *css.Rule, parentMedia *MediaCondition, origin CSSOrigin) {
		currentMedia := parentMedia

		// If this rule is an @media rule, parse its prelude and merge with parent
		if strings.EqualFold(rule.Name, "@media") || strings.HasPrefix(strings.ToLower(rule.Name), "@media") {
			mediaCond := ParseMediaQuery(rule.Prelude)
			if parentMedia != nil {
				currentMedia = parentMedia.Merge(mediaCond)
			} else {
				currentMedia = mediaCond
			}
		}

		for _, sel := range rule.Selectors {
			base, pseudo := splitPseudoElement(sel)
			if base == "" {
				continue
			}
			compiled, err := cascadia.Compile(base)
			if err != nil {
				continue
			}
			globalIndex++
			matcher.Rules = append(matcher.Rules, CSSRule{
				Selector:    base,
				Compiled:    compiled,
				Decls:       rule.Declarations,
				MediaCond:   currentMedia,
				Specificity: ComputeSpecificity(sel),
				Origin:      origin,
				Index:       globalIndex,
				Pseudo:      pseudo,
			})
		}
		for _, child := range rule.Rules {
			addRule(child, currentMedia, origin)
		}
	}

	for _, ps := range parsed {
		preproc.PreprocessSheet(ps.sheet)
		for _, rule := range ps.sheet.Rules {
			addRule(rule, nil, ps.origin)
		}
	}

	// Sort rules in ascending cascade priority order (lowest priority first, highest priority last)
	sort.SliceStable(matcher.Rules, func(i, j int) bool {
		r1 := matcher.Rules[i]
		r2 := matcher.Rules[j]

		// 1. Origin (Tag < Framework < Theme < Custom)
		if r1.Origin != r2.Origin {
			return r1.Origin < r2.Origin
		}

		// 2. Specificity (A, B, C)
		specCmp := r1.Specificity.Compare(r2.Specificity)
		if specCmp != 0 {
			return specCmp < 0
		}

		// 3. Document Order (earlier < later)
		return r1.Index < r2.Index
	})
	if len(stylesheets) == 0 {
		builtInMatcherMu.Lock()
		builtInMatcherCache[theme] = matcher
		builtInMatcherMu.Unlock()
	}

	return matcher, nil
}

// ComputeStyleForSelection resolves styling for a specific DOM selection given a terminal width.
func (m *CSSMatcher) ComputeStyleForSelection(sel *goquery.Selection, termWidth int, parentStyle *ComputedStyle) ComputedStyle {
	s := ComputedStyle{
		Display:   DisplayBlock,
		TextAlign: lipgloss.Left,
	}
	node := sel.Get(0)
	if node == nil {
		return s
	}

	// Match tag defaults
	tagName := goquery.NodeName(sel)
	switch tagName {
	case "span", "a", "strong", "b", "em", "i", "code", "img":
		s.Display = DisplayInline
	case "blink":
		s.Display = DisplayInline
		s.Blink = true
		s.BlinkSet = true
	case "spinner":
		s.Display = DisplayInline
		s.Spinner = true
		s.SpinnerSet = true
	case "shimmer":
		s.Display = DisplayInline
		s.Shimmer = true
		s.ShimmerSet = true
	case "s", "del", "strike":
		s.Display = DisplayInline
		s.Strikethrough = true
		s.StrikethroughSet = true
	case "th", "td":
		s.Display = DisplayBlock
	}
	ctx := DefaultUnitContext(termWidth)
	if parentStyle != nil && parentStyle.Width > 0 {
		ctx.ParentWidth = parentStyle.Width
	}

	// Filter matched rules respecting media conditions
	var matchedRules []*CSSRule
	for i := range m.Rules {
		rule := &m.Rules[i]
		if rule.Pseudo != "" {
			continue
		}
		if rule.MediaCond != nil && !rule.MediaCond.Matches(termWidth, 24, m.Theme) {
			continue
		}
		if rule.Compiled.Match(node) {
			matchedRules = append(matchedRules, rule)
		}
	}

	// Pass 1: Apply normal stylesheet declarations in ascending specificity order
	for _, rule := range matchedRules {
		for _, decl := range rule.Decls {
			if !decl.Important {
				applyDeclaration(&s, decl, ctx)
			}
		}
	}

	// Parse inline styles if present
	// Apply HTML attributes as base inline declarations if present (e.g. width="50" color="#fff")
	if colorAttr, ok := sel.Attr("color"); ok && colorAttr != "" {
		applyDeclaration(&s, &css.Declaration{Property: "color", Value: colorAttr}, ctx)
	}
	if widthAttr, ok := sel.Attr("width"); ok && widthAttr != "" {
		applyDeclaration(&s, &css.Declaration{Property: "width", Value: widthAttr}, ctx)
	}

	var inlineDecls []*css.Declaration
	if inlineStyle, ok := sel.Attr("style"); ok && inlineStyle != "" {
		if decls, err := parser.ParseDeclarations(inlineStyle); err == nil {
			if m.Preprocessor != nil {
				m.Preprocessor.PreprocessDeclarations(decls)
			}
			inlineDecls = decls
		}
	}

	// Pass 2: Apply normal inline style declarations (overrides normal stylesheet rules)
	for _, decl := range inlineDecls {
		if !decl.Important {
			applyDeclaration(&s, decl, ctx)
		}
	}

	// Pass 3: Apply !important stylesheet declarations (overrides normal inline and stylesheet rules)
	for _, rule := range matchedRules {
		for _, decl := range rule.Decls {
			if decl.Important {
				applyDeclaration(&s, decl, ctx)
			}
		}
	}

	// Pass 4: Apply !important inline style declarations (highest priority)
	for _, decl := range inlineDecls {
		if decl.Important {
			applyDeclaration(&s, decl, ctx)
		}
	}

	// Extract responsive column spans and row-cols based on current breakpoint
	bp := detectBreakpoint(termWidth)
	classAttr, _ := sel.Attr("class")
	classes := strings.Fields(classAttr)
	for _, c := range classes {
		if c == "spinner" {
			s.Spinner = true
			s.SpinnerSet = true
		} else if c == "shimmer" {
			s.Shimmer = true
			s.ShimmerSet = true
		} else if c == "border-shimmer" {
			s.BorderShimmer = true
		}
	}
	span, isAuto := resolveColSpanResponsive(classes, bp)
	if span > 0 {
		s.ColSpan = span
	}
	s.ColAuto = isAuto
	s.RowCols = resolveRowColsResponsive(classes, bp)

	// Inherit missing properties from parentStyle if present
	if parentStyle != nil {
		if !s.FgColorSet {
			s.FgColor = parentStyle.FgColor
			s.FgColorSet = parentStyle.FgColorSet
		}
		if !s.TextAlignSet {
			s.TextAlign = parentStyle.TextAlign
			s.TextAlignSet = parentStyle.TextAlignSet
		}
		if !s.BoldSet {
			s.Bold = parentStyle.Bold
			s.BoldSet = parentStyle.BoldSet
		}
		if !s.ItalicSet {
			s.Italic = parentStyle.Italic
			s.ItalicSet = parentStyle.ItalicSet
		}
		if !s.UnderlineSet {
			s.Underline = parentStyle.Underline
			s.UnderlineSet = parentStyle.UnderlineSet
		}
		if !s.UnderlineStyleSet {
			s.UnderlineStyle = parentStyle.UnderlineStyle
			s.UnderlineStyleSet = parentStyle.UnderlineStyleSet
		}
		if !s.UnderlineColorSet {
			s.UnderlineColor = parentStyle.UnderlineColor
			s.UnderlineColorSet = parentStyle.UnderlineColorSet
		}
		if !s.StrikethroughSet {
			s.Strikethrough = parentStyle.Strikethrough
			s.StrikethroughSet = parentStyle.StrikethroughSet
		}
		if !s.FaintSet {
			s.Faint = parentStyle.Faint
			s.FaintSet = parentStyle.FaintSet
		}
		if !s.BlinkSet {
			s.Blink = parentStyle.Blink
			s.BlinkSet = parentStyle.BlinkSet
		}
		if !s.ShimmerSet {
			s.Shimmer = parentStyle.Shimmer
			s.ShimmerSet = parentStyle.ShimmerSet
			s.ShimmerHighlight = parentStyle.ShimmerHighlight
		}
		if !s.SpinnerSet {
			s.Spinner = parentStyle.Spinner
			s.SpinnerSet = parentStyle.SpinnerSet
		}
		if !s.TextTransformSet {
			s.TextTransform = parentStyle.TextTransform
			s.TextTransformSet = parentStyle.TextTransformSet
		}
	}

	return s
}

// splitPseudoElement separates a trailing ::before/::after (or legacy :before/:after)
// pseudo-element from a selector. Returns the base selector and the pseudo name
// ("before"/"after"), or an empty pseudo when none is present.
func splitPseudoElement(sel string) (string, string) {
	low := strings.ToLower(sel)
	suffixes := []struct{ suf, name string }{
		{"::before", "before"},
		{"::after", "after"},
		{":before", "before"},
		{":after", "after"},
	}
	for _, p := range suffixes {
		if strings.HasSuffix(low, p.suf) {
			return strings.TrimSpace(sel[:len(sel)-len(p.suf)]), p.name
		}
	}
	return sel, ""
}

// parseCSSContent resolves a CSS `content` value to literal text. It unquotes
// string tokens and decodes CSS unicode escapes (e.g. "\2014", "\00a0"). It
// returns ok=false for `none`/`normal`, empty strings, and unsupported
// functional values (counters(), attr(), unresolved var()).
func parseCSSContent(val string) (string, bool) {
	v := strings.TrimSpace(val)
	if i := strings.Index(v, "/*"); i >= 0 {
		v = strings.TrimSpace(v[:i])
	}
	if v == "" || strings.EqualFold(v, "none") || strings.EqualFold(v, "normal") {
		return "", false
	}
	if v[0] != '"' && v[0] != '\'' {
		return "", false
	}
	quote := v[0]
	var sb strings.Builder
	i := 1
	for i < len(v) {
		c := v[i]
		if c == '\\' {
			j := i + 1
			hex := ""
			for j < len(v) && len(hex) < 6 && isHexDigit(v[j]) {
				hex += string(v[j])
				j++
			}
			if hex != "" {
				if cp, err := strconv.ParseInt(hex, 16, 32); err == nil {
					sb.WriteRune(rune(cp))
				}
				if j < len(v) && (v[j] == ' ' || v[j] == '\t') {
					j++
				}
				i = j
				continue
			}
			if i+1 < len(v) {
				sb.WriteByte(v[i+1])
				i += 2
				continue
			}
			i++
			continue
		}
		if c == quote {
			break
		}
		sb.WriteByte(c)
		i++
	}
	return sb.String(), true
}

func isHexDigit(b byte) bool {
	return (b >= '0' && b <= '9') || (b >= 'a' && b <= 'f') || (b >= 'A' && b <= 'F')
}

// PseudoText resolves the generated text and style for a ::before/::after
// pseudo-element of sel. The pseudo-element inherits typographic properties from
// elementStyle; its own declarations (e.g. color) override. Returns ok=false
// when no pseudo rule matches or the resolved content is empty.
func (m *CSSMatcher) PseudoText(sel *goquery.Selection, termWidth int, elementStyle *ComputedStyle, pseudo string) (string, ComputedStyle, bool) {
	node := sel.Get(0)
	if node == nil {
		return "", ComputedStyle{}, false
	}
	ctx := DefaultUnitContext(termWidth)
	var matched []*CSSRule
	for i := range m.Rules {
		r := &m.Rules[i]
		if r.Pseudo != pseudo {
			continue
		}
		if r.MediaCond != nil && !r.MediaCond.Matches(termWidth, 24, m.Theme) {
			continue
		}
		if r.Compiled.Match(node) {
			matched = append(matched, r)
		}
	}
	if len(matched) == 0 {
		return "", ComputedStyle{}, false
	}

	s := ComputedStyle{Display: DisplayInline}
	if elementStyle != nil {
		s.Bold = elementStyle.Bold
		s.Italic = elementStyle.Italic
		s.Underline = elementStyle.Underline
		s.Strikethrough = elementStyle.Strikethrough
		s.Faint = elementStyle.Faint
		s.Blink = elementStyle.Blink
		s.FgColor = elementStyle.FgColor
		s.FgColorSet = elementStyle.FgColorSet
		s.TextTransform = elementStyle.TextTransform
		s.TextTransformSet = elementStyle.TextTransformSet
	}

	content := ""
	hasContent := false
	apply := func(important bool) {
		for _, r := range matched {
			for _, decl := range r.Decls {
				if decl.Important != important {
					continue
				}
				if strings.EqualFold(strings.TrimSpace(decl.Property), "content") {
					if txt, ok := parseCSSContent(decl.Value); ok {
						content = txt
						hasContent = true
					} else {
						hasContent = false
					}
					continue
				}
				applyDeclaration(&s, decl, ctx)
			}
		}
	}
	apply(false)
	apply(true)

	if !hasContent || content == "" {
		return "", ComputedStyle{}, false
	}
	return content, s, true
}

// resolveColSpanResponsive extracts the appropriate column span (1-12) for the current breakpoint.
func resolveColSpanResponsive(classes []string, currentBP Breakpoint) (int, bool) {
	type colSpec struct {
		bp     Breakpoint
		span   int
		isAuto bool
	}

	var specs []colSpec
	for _, cls := range classes {
		if cls == "col" {
			specs = append(specs, colSpec{bp: bpXS, span: 0, isAuto: true})
			continue
		}
		if !strings.HasPrefix(cls, "col-") {
			continue
		}
		parts := strings.Split(cls, "-")
		if len(parts) == 2 { // col-6, col-auto
			if parts[1] == "auto" {
				specs = append(specs, colSpec{bp: bpXS, span: 0, isAuto: true})
			} else if span, err := strconv.Atoi(parts[1]); err == nil && span >= 1 && span <= 12 {
				specs = append(specs, colSpec{bp: bpXS, span: span})
			}
		} else if len(parts) == 3 { // col-md-6, col-md-auto
			bpPrefix := parts[1]
			var bp Breakpoint
			switch bpPrefix {
			case "xs":
				bp = bpXS
			case "sm":
				bp = bpSM
			case "md":
				bp = bpMD
			case "lg":
				bp = bpLG
			case "xl":
				bp = bpXL
			case "xxl":
				bp = bpXXL
			default:
				continue
			}
			if parts[2] == "auto" {
				specs = append(specs, colSpec{bp: bp, span: 0, isAuto: true})
			} else if span, err := strconv.Atoi(parts[2]); err == nil && span >= 1 && span <= 12 {
				specs = append(specs, colSpec{bp: bp, span: span})
			}
		}
	}

	// Find the best match: highest breakpoint <= currentBP
	bestSpan := 0
	bestAuto := false
	bestBP := Breakpoint(-1)
	for _, spec := range specs {
		if spec.bp <= currentBP && spec.bp > bestBP {
			bestBP = spec.bp
			bestSpan = spec.span
			bestAuto = spec.isAuto
		}
	}

	// If classes specify column breakpoints but none match the current breakpoint (e.g. col-md-6 on xs),
	// the column stacks (span = 12).
	if len(specs) > 0 && bestSpan == 0 && !bestAuto {
		return 12, false
	}

	return bestSpan, bestAuto
}

// resolveRowColsResponsive extracts the number of columns per row (from row-cols-[bp-]N)
func resolveRowColsResponsive(classes []string, currentBP Breakpoint) int {
	type rowColSpec struct {
		bp    Breakpoint
		count int
	}
	var specs []rowColSpec
	for _, cls := range classes {
		if !strings.HasPrefix(cls, "row-cols-") {
			continue
		}
		parts := strings.Split(cls, "-")
		if len(parts) == 3 { // row-cols-2
			if n, err := strconv.Atoi(parts[2]); err == nil && n >= 1 && n <= 12 {
				specs = append(specs, rowColSpec{bp: bpXS, count: n})
			}
		} else if len(parts) == 4 { // row-cols-md-2
			bpPrefix := parts[2]
			var bp Breakpoint
			switch bpPrefix {
			case "xs":
				bp = bpXS
			case "sm":
				bp = bpSM
			case "md":
				bp = bpMD
			case "lg":
				bp = bpLG
			case "xl":
				bp = bpXL
			case "xxl":
				bp = bpXXL
			default:
				continue
			}
			if n, err := strconv.Atoi(parts[3]); err == nil && n >= 1 && n <= 12 {
				specs = append(specs, rowColSpec{bp: bp, count: n})
			}
		}
	}

	bestCount := 0
	bestBP := Breakpoint(-1)
	for _, spec := range specs {
		if spec.bp <= currentBP && spec.bp > bestBP {
			bestBP = spec.bp
			bestCount = spec.count
		}
	}
	return bestCount
}

func applyDeclaration(s *ComputedStyle, decl *css.Declaration, ctx UnitContext) {
	prop := strings.TrimSpace(strings.ToLower(decl.Property))
	val := strings.TrimSpace(strings.ToLower(decl.Value))

	switch prop {
	case "display":
		switch val {
		case "block":
			s.Display = DisplayBlock
		case "inline":
			s.Display = DisplayInline
		case "flex":
			s.Display = DisplayFlex
		case "none":
			s.Display = DisplayNone
		}

	case "margin":
		s.MarginTop, s.MarginRight, s.MarginBottom, s.MarginLeft = ResolveBox4D(val, ctx)
	case "margin-top":
		s.MarginTop = ResolveLength(val, AxisY, ctx)
	case "margin-bottom":
		s.MarginBottom = ResolveLength(val, AxisY, ctx)
	case "margin-left":
		s.MarginLeft = ResolveLength(val, AxisX, ctx)
	case "margin-right":
		s.MarginRight = ResolveLength(val, AxisX, ctx)

	case "padding":
		s.PaddingTop, s.PaddingRight, s.PaddingBottom, s.PaddingLeft = ResolveBox4D(val, ctx)
	case "padding-top":
		s.PaddingTop = ResolveLength(val, AxisY, ctx)
	case "padding-bottom":
		s.PaddingBottom = ResolveLength(val, AxisY, ctx)
	case "padding-left":
		s.PaddingLeft = ResolveLength(val, AxisX, ctx)
	case "padding-right":
		s.PaddingRight = ResolveLength(val, AxisX, ctx)

	case "width":
		s.Width = ResolveLength(val, AxisX, ctx)

	case "font-size":
		bold, _, _ := ResolveFontSize(val, ctx.RootFontSize)
		if bold && !s.BoldSet {
			s.Bold = true
			s.BoldSet = true
		}
	case "border":
		if val == "true" || val == "1" || strings.Contains(val, "solid") || strings.Contains(val, "1px") ||
			strings.Contains(val, "double") || strings.Contains(val, "dashed") || strings.Contains(val, "dotted") {
			s.Border = true
		}
		bs := borderStyleKeyword(val)
		if bs != "" {
			s.BorderStyle = bs
		}
		// Width only upgrades to thick for plain solid borders; an explicit
		// dashed/dotted/double keyword takes precedence.
		if (bs == "" || bs == "solid") && borderWidthIsThick(val) {
			s.BorderStyle = "thick"
		}
		for _, part := range strings.Fields(val) {
			if strings.HasPrefix(part, "#") && len(part) >= 4 {
				s.BorderColor = part
			}
		}
	case "border-top":
		s.BorderTop = val == "true" || val == "1" || strings.Contains(val, "solid")
	case "border-bottom":
		s.BorderBottom = val == "true" || val == "1" || strings.Contains(val, "solid")
	case "border-left":
		s.BorderLeft = val == "true" || val == "1" || strings.Contains(val, "solid")
	case "border-right":
		s.BorderRight = val == "true" || val == "1" || strings.Contains(val, "solid")
	case "border-radius":
		s.Rounded = val == "true" || val == "1" || val == "rounded" || (val != "0" && val != "0px" && val != "0rem" && val != "none" && val != "")
	case "color":
		if val == "inherit" || val == "initial" {
			s.FgColor = ""
			s.FgColorSet = false
		} else {
			s.FgColor = resolveColor(decl.Value)
			s.FgColorSet = true
		}
	case "background-color", "background":
		if strings.Contains(val, "gradient") {
			s.BgGradient = parseColorStops(val)
		} else if val == "transparent" || val == "inherit" || val == "initial" {
			s.BgColor = ""
			s.BgGradient = nil
		} else {
			s.BgColor = resolveColor(decl.Value)
		}
	case "border-color":
		s.BorderColor = resolveColor(decl.Value)
	case "font-weight":
		s.BoldSet = true
		if val == "bold" || val == "700" || val == "800" || val == "900" {
			s.Bold = true
		} else if val == "normal" || val == "400" {
			s.Bold = false
		}
	case "font-style":
		s.ItalicSet = true
		if val == "italic" || val == "oblique" {
			s.Italic = true
		} else if val == "normal" {
			s.Italic = false
		}
	case "text-decoration", "text-decoration-line":
		if strings.Contains(val, "underline") {
			s.UnderlineSet = true
			s.Underline = true
		}
		if strings.Contains(val, "wavy") || strings.Contains(val, "curly") {
			s.UnderlineStyle = lipgloss.UnderlineCurly
			s.UnderlineStyleSet = true
		} else if strings.Contains(val, "double") {
			s.UnderlineStyle = lipgloss.UnderlineDouble
			s.UnderlineStyleSet = true
		} else if strings.Contains(val, "dotted") {
			s.UnderlineStyle = lipgloss.UnderlineDotted
			s.UnderlineStyleSet = true
		} else if strings.Contains(val, "dashed") {
			s.UnderlineStyle = lipgloss.UnderlineDashed
			s.UnderlineStyleSet = true
		} else if strings.Contains(val, "solid") {
			s.UnderlineStyle = lipgloss.UnderlineSingle
			s.UnderlineStyleSet = true
		}
		if strings.Contains(val, "line-through") {
			s.StrikethroughSet = true
			s.Strikethrough = true
		}
		if strings.Contains(val, "blink") {
			s.BlinkSet = true
			s.Blink = true
		}
		if strings.Contains(val, "shimmer") {
			s.ShimmerSet = true
			s.Shimmer = true
		}
		if strings.Contains(val, "spinner") {
			s.SpinnerSet = true
			s.Spinner = true
		}
		if strings.Contains(val, "none") {
			s.UnderlineSet = true
			s.Underline = false
			s.UnderlineStyleSet = true
			s.UnderlineStyle = lipgloss.UnderlineNone
			s.StrikethroughSet = true
			s.Strikethrough = false
		}
	case "animation", "animation-name":
		if strings.Contains(val, "shimmer") {
			s.ShimmerSet = true
			s.Shimmer = true
		}
		if strings.Contains(val, "spinner") {
			s.SpinnerSet = true
			s.Spinner = true
		}
		if strings.Contains(val, "blink") {
			s.BlinkSet = true
			s.Blink = true
		}
	case "text-decoration-style":
		s.UnderlineStyleSet = true
		switch strings.TrimSpace(val) {
		case "wavy", "curly":
			s.UnderlineStyle = lipgloss.UnderlineCurly
		case "double":
			s.UnderlineStyle = lipgloss.UnderlineDouble
		case "dotted":
			s.UnderlineStyle = lipgloss.UnderlineDotted
		case "dashed":
			s.UnderlineStyle = lipgloss.UnderlineDashed
		default:
			s.UnderlineStyle = lipgloss.UnderlineSingle
		}
	case "text-decoration-color":
		s.UnderlineColor = decl.Value
		s.UnderlineColorSet = true
	case "text-align":
		s.TextAlignSet = true
		switch val {
		case "center":
			s.TextAlign = lipgloss.Center
		case "right", "end":
			s.TextAlign = lipgloss.Right
		case "left", "start":
			s.TextAlign = lipgloss.Left
		}
	case "text-transform":
		s.TextTransformSet = true
		switch val {
		case "uppercase", "lowercase", "capitalize":
			s.TextTransform = val
		case "none":
			s.TextTransform = ""
		}
	case "box-shadow":
		s.Shadow = shadowSizeFromBoxShadow(val)
	case "opacity":
		if f, err := strconv.ParseFloat(strings.TrimSpace(val), 64); err == nil {
			s.FaintSet = true
			s.Faint = f > 0 && f < 1
		}
	case "border-style":
		if bs := borderStyleKeyword(val); bs != "" {
			s.BorderStyle = bs
			if bs == "hidden" || bs == "none" {
				s.Border = false
			}
		}
	case "border-width":
		if borderWidthIsThick(val) {
			s.BorderStyle = "thick"
		}
	case "border-top-color":
		s.BorderTopColor = decl.Value
	case "border-right-color":
		s.BorderRightColor = decl.Value
	case "border-bottom-color":
		s.BorderBottomColor = decl.Value
	case "border-left-color":
		s.BorderLeftColor = decl.Value
	case "border-image", "border-color-blend":
		s.BorderBlend = parseColorStops(val)
	case "max-width":
		if val != "none" {
			s.MaxWidth = ResolveLength(val, AxisX, ctx)
		}
	case "max-height":
		if val != "none" {
			s.MaxHeight = ResolveLength(val, AxisY, ctx)
		}
	case "align-items":
		s.AlignItemsSet = true
		s.AlignItems = crossAxisPosition(val)
	case "justify-content":
		s.JustifyContentSet = true
		s.JustifyContent = justifyPosition(val)
	case "vertical-align":
		s.VerticalAlignSet = true
		s.VerticalAlign = verticalAlignPosition(val)
	case "list-style", "list-style-type":
		s.ListStyle = listStyleKeyword(val)
	}
}

// borderStyleKeyword extracts a supported border-style keyword from a CSS value.
func borderStyleKeyword(val string) string {
	for _, kw := range []string{"outer-half-block", "inner-half-block", "block", "markdown", "ascii", "double", "dashed", "dotted", "hidden", "none", "solid"} {
		if strings.Contains(val, kw) {
			return kw
		}
	}
	return ""
}

// borderWidthIsThick reports whether a border-width value reads as thick/medium
// or more than one pixel, mapping to lipgloss's ThickBorder.
func borderWidthIsThick(val string) bool {
	if strings.Contains(val, "thick") || strings.Contains(val, "medium") {
		return true
	}
	for _, m := range shadowBlurRe.FindAllStringSubmatch(val, -1) {
		if f, err := strconv.ParseFloat(m[1], 64); err == nil {
			if (m[2] == "px" && f > 1) || (m[2] == "rem" && f > 0.0625) {
				return true
			}
		}
	}
	return false
}

// crossAxisPosition maps a CSS align-items value to a lipgloss vertical Position.
func crossAxisPosition(val string) lipgloss.Position {
	switch val {
	case "center":
		return lipgloss.Center
	case "flex-end", "end", "bottom":
		return lipgloss.Bottom
	default: // flex-start, start, stretch, baseline, top
		return lipgloss.Top
	}
}

// justifyPosition maps a CSS justify-content value to a lipgloss horizontal Position.
func justifyPosition(val string) lipgloss.Position {
	switch val {
	case "center":
		return lipgloss.Center
	case "flex-end", "end", "right":
		return lipgloss.Right
	default: // flex-start, start, left, space-between, space-around
		return lipgloss.Left
	}
}

// verticalAlignPosition maps a CSS vertical-align value to a lipgloss vertical Position.
func verticalAlignPosition(val string) lipgloss.Position {
	switch val {
	case "middle":
		return lipgloss.Center
	case "bottom", "text-bottom":
		return lipgloss.Bottom
	default: // top, text-top, baseline
		return lipgloss.Top
	}
}

// listStyleKeyword normalises a CSS list-style/list-style-type value.
func listStyleKeyword(val string) string {
	for _, kw := range []string{"none", "disc", "circle", "square", "decimal", "lower-alpha", "lower-latin", "upper-alpha", "upper-latin", "lower-roman", "upper-roman"} {
		if strings.Contains(val, kw) {
			return kw
		}
	}
	return ""
}
// parseColorStops extracts hex/rgb/named colors from a CSS linear-gradient or space/comma separated list.
func parseColorStops(val string) []string {
	var stops []string
	// Match hex colors #fff, #ffffff, or rgb(...)
	hexRe := regexp.MustCompile(`#[0-9a-fA-F]{3,8}`)
	for _, m := range hexRe.FindAllString(val, -1) {
		stops = append(stops, m)
	}
	return stops
}
// resolveColor parses and applies color transformations like lighten(), darken(), complementary(), etc.
func resolveColor(val string) string {
	v := strings.TrimSpace(val)
	if strings.HasPrefix(v, "lighten(") && strings.HasSuffix(v, ")") {
		inner := strings.TrimSuffix(strings.TrimPrefix(v, "lighten("), ")")
		parts := strings.Split(inner, ",")
		if len(parts) >= 2 {
			c := lipgloss.Color(strings.TrimSpace(parts[0]))
			pctStr := strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(parts[1]), "%"))
			if pct, err := strconv.ParseFloat(pctStr, 64); err == nil {
				lightened := lipgloss.Lighten(c, pct/100.0)
				r, g, b, _ := lightened.RGBA()
				return fmt.Sprintf("#%02x%02x%02x", r>>8, g>>8, b>>8)
			}
		}
	} else if strings.HasPrefix(v, "darken(") && strings.HasSuffix(v, ")") {
		inner := strings.TrimSuffix(strings.TrimPrefix(v, "darken("), ")")
		parts := strings.Split(inner, ",")
		if len(parts) >= 2 {
			c := lipgloss.Color(strings.TrimSpace(parts[0]))
			pctStr := strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(parts[1]), "%"))
			if pct, err := strconv.ParseFloat(pctStr, 64); err == nil {
				darkened := lipgloss.Darken(c, pct/100.0)
				r, g, b, _ := darkened.RGBA()
				return fmt.Sprintf("#%02x%02x%02x", r>>8, g>>8, b>>8)
			}
		}
	} else if strings.HasPrefix(v, "complementary(") && strings.HasSuffix(v, ")") {
		inner := strings.TrimSuffix(strings.TrimPrefix(v, "complementary("), ")")
		c := lipgloss.Color(strings.TrimSpace(inner))
		comp := lipgloss.Complementary(c)
		r, g, b, _ := comp.RGBA()
		return fmt.Sprintf("#%02x%02x%02x", r>>8, g>>8, b>>8)
	}
	return val
}

// shadowBlurRe captures CSS length tokens (rem/px) from a box-shadow value,
// e.g. the offsets and blur radius. Unitless numbers (such as rgba components)
// are intentionally excluded.
var shadowBlurRe = regexp.MustCompile(`(-?\d*\.?\d+)(rem|px)`)

// shadowSizeFromBoxShadow maps a CSS box-shadow value to a terminal shadow
// size (0-3), derived from the largest length in the value (the blur radius).
// Bootstrap's shadow-sm (0.25rem), shadow (1rem) and shadow-lg (3rem) thus
// resolve to 1, 2 and 3 instead of collapsing to a single size.
func shadowSizeFromBoxShadow(val string) int {
	val = strings.TrimSpace(val)
	if val == "" || val == "none" || val == "0" {
		return 0
	}
	// Inset shadows paint inside the box (Bootstrap uses them for table/stripe
	// backgrounds); they are not drop shadows and must not reserve layout space.
	if strings.Contains(val, "inset") {
		return 0
	}
	matches := shadowBlurRe.FindAllStringSubmatch(val, -1)
	if len(matches) == 0 {
		// No CSS length units: treat as an explicit integer size (e.g. "box-shadow: 2").
		return parseInt(val)
	}
	maxRem := 0.0
	for _, m := range matches {
		f, err := strconv.ParseFloat(m[1], 64)
		if err != nil {
			continue
		}
		if f < 0 {
			f = -f
		}
		if m[2] == "px" {
			f /= 16.0
		}
		if f > maxRem {
			maxRem = f
		}
	}
	switch {
	case maxRem <= 0.3:
		return 1
	case maxRem <= 1.5:
		return 2
	default:
		return 3
	}
}

func parseInt(val string) int {
	return ResolveLength(val, AxisX, DefaultUnitContext(80))
}

func parseBox4D(val string) (int, int, int, int) {
	return ResolveBox4D(val, DefaultUnitContext(80))
}

func parseBox2D(val string) (int, int) {
	top, right, _, _ := parseBox4D(val)
	return top, right
}
