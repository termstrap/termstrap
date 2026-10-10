package termstrap

import (
	"fmt"
	"image"
	"image/color"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/webp"

	"github.com/alecthomas/chroma/v2"
	"github.com/PuerkitoBio/goquery"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
	termimage "github.com/go-scripts/termstrap/image"
	"github.com/muesli/reflow/wordwrap"
	"github.com/muesli/reflow/wrap"
)

// ImageOverlay represents an image escape sequence separated from the text.
type ImageOverlay struct {
	Row    int    // Line offset in rendered text
	Col    int    // Column position in terminal
	Width  int    // Width in character cells
	Height int    // Height in character cells
	Escape string // Terminal escape sequence
}

// Model represents the HTML/CSS rendering context.
type Model struct {
	HTML        string   // Raw HTML content
	Theme       Theme    // Visual color theme (example: ThemeBootstrap)
	Stylesheets []string // Additional CSS stylesheets
	Width       int      // Terminal width in characters
	RootPath    string   // Base path for local images
	// DisableHyperlinks disables emitting OSC 8 ANSI hyperlinks for <a> tags.
	DisableHyperlinks bool

	// ImageRenderer overrides the default image renderer.
	ImageRenderer termimage.Renderer

	// MaxImageWidth caps image width in terminal columns (0 = default).
	MaxImageWidth int

	// ColorMode specifies color depth.
	ColorMode termimage.ColorMode

	// OptimizeSequences deduplicates consecutive ANSI color codes.
	OptimizeSequences *bool

	// DisableSyntaxHighlight turns off chroma syntax colorization of <pre>
	// code blocks. Highlighting is enabled by default.
	DisableSyntaxHighlight bool

	// SyntaxTheme overrides the chroma highlighting style (e.g. "monokai",
	// "dracula", "nord", "github-dark"). Empty picks a style from Theme.
	SyntaxTheme string

	// DisableImages replaces images with [alt](src) text.
	DisableImages bool

	// Cache policy
	CachePolicy CachePolicy
	CacheTTL    time.Duration

	// codeStyle is the chroma style resolved from the theme's --ts-code-* CSS
	// variables (or SyntaxTheme override), built once per render.
	codeStyle *chroma.Style

	// Step specifies the animation frame step for <spinner>, <shimmer>, .spinner, .shimmer.
	Step int
}

// Render compiles the HTML/CSS tree and returns the rendered ANSI string.
func (m Model) Render() (string, error) {
	rendered, _, err := m.renderWithOverlays(false)
	return rendered, err
}

// RenderWithOverlays returns rendered text along with separated image overlays.
func (m Model) RenderWithOverlays() (string, []ImageOverlay, error) {
	return m.renderWithOverlays(true)
}

func (m Model) renderWithOverlays(withOverlays bool) (string, []ImageOverlay, error) {
	if strings.TrimSpace(m.HTML) == "" {
		return "", nil, nil
	}

	// Configure Lipgloss color profile based on ColorMode and environment
	switch m.ColorMode {
	case termimage.ColorMode256:
		lipgloss.Writer.Profile = colorprofile.ANSI256
	case termimage.ColorMode16:
		lipgloss.Writer.Profile = colorprofile.ANSI
	case termimage.ColorModeTrueColor:
		lipgloss.Writer.Profile = colorprofile.TrueColor
	default:
		colorTerm := strings.ToLower(os.Getenv("COLORTERM"))
		if colorTerm == "truecolor" || colorTerm == "24bit" || os.Getenv("CLICOLOR_FORCE") == "1" {
			lipgloss.Writer.Profile = colorprofile.TrueColor
		} else if lipgloss.Writer.Profile == colorprofile.Ascii || lipgloss.Writer.Profile == colorprofile.NoTTY {
			lipgloss.Writer.Profile = colorprofile.TrueColor
		}
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(m.HTML))
	if err != nil {
		return "", nil, err
	}

	// Extract embedded <style> tags
	doc.Find("style").Each(func(i int, s *goquery.Selection) {
		text := s.Text()
		if strings.TrimSpace(text) != "" {
			m.Stylesheets = append(m.Stylesheets, text)
		}
	})

	// Extract linked <link rel="stylesheet"> tags
	doc.Find("link[rel='stylesheet']").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if !exists || strings.TrimSpace(href) == "" {
			return
		}
		href = strings.TrimSpace(href)
		if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") {
			client := &http.Client{Timeout: 5 * time.Second}
			resp, err := client.Get(href)
			if err == nil {
				defer resp.Body.Close()
				if resp.StatusCode >= 200 && resp.StatusCode < 300 {
					if data, err := io.ReadAll(resp.Body); err == nil {
						m.Stylesheets = append(m.Stylesheets, string(data))
					}
				}
			}
		} else {
			path := href
			if m.RootPath != "" && !filepath.IsAbs(path) {
				path = filepath.Join(m.RootPath, path)
			}
			if data, err := os.ReadFile(path); err == nil {
				m.Stylesheets = append(m.Stylesheets, string(data))
			}
		}
	})

	matcher, err := NewCSSMatcher(m.Theme, m.Stylesheets...)
	if err != nil {
		return "", nil, err
	}

	// Resolve the syntax-highlighting palette from the theme's --ts-code-* CSS
	// variables (overridable per theme) once for the whole render.
	if !m.DisableSyntaxHighlight {
		var vars map[string]string
		if matcher.Preprocessor != nil {
			vars = matcher.Preprocessor.GlobalVars
		}
		m.codeStyle = resolveCodeStyle(m.SyntaxTheme, vars)
	}

	body := doc.Find("body")
	var rootSel *goquery.Selection
	if body.Length() > 0 {
		rootSel = body
	} else {
		rootSel = doc.Selection
	}

	termWidth := m.Width
	if termWidth <= 0 {
		termWidth = 80
	}

	renderTree := BuildRenderTree(rootSel, matcher, termWidth, nil)
	if renderTree == nil {
		return "", nil, nil
	}

	ComputeLayout(renderTree, termWidth)

	var overlays []ImageOverlay
	var links []string
	output := m.renderNode(renderTree, 0, 0, &overlays, 0, &links)

	// Expand link placeholders into real OSC 8 sequences without breaking line wrapping
	for id, url := range links {
		startPlaceholder := fmt.Sprintf("\x1b[999000;%dm", id)
		endPlaceholder := fmt.Sprintf("\x1b[999001;%dm", id)
		osc8Start := fmt.Sprintf("\x1b]8;id=link_%d;%s\x1b\\", id, url)
		osc8End := "\x1b]8;;\x1b\\"

		// If wrapping broke the link across lines, ensure every line starts OSC 8 and ends OSC 8
		var lines []string
		inLink := false
		for _, line := range strings.Split(output, "\n") {
			if strings.Contains(line, startPlaceholder) {
				inLink = true
			}
			if inLink {
				// Replace end placeholder if present on this line
				hasEnd := strings.Contains(line, endPlaceholder)
				line = strings.ReplaceAll(line, startPlaceholder, osc8Start)
				if !strings.HasPrefix(line, osc8Start) && !strings.Contains(line, osc8Start) {
					line = osc8Start + line
				}
				if hasEnd {
					line = strings.ReplaceAll(line, endPlaceholder, osc8End)
					inLink = false
				} else {
					line = line + osc8End
				}
			}
			lines = append(lines, line)
		}
		output = strings.Join(lines, "\n")
	}

	return output, overlays, nil
}

func applyInlineStyle(s ComputedStyle, text string) string {
	if text == "" {
		return ""
	}
	ls := lipgloss.NewStyle()
	hasStyle := false

	if s.Bold {
		ls = ls.Bold(true)
		hasStyle = true
	}
	if s.Italic {
		ls = ls.Italic(true)
		hasStyle = true
	}
	if s.Underline {
		ls = ls.Underline(true)
		if s.UnderlineStyleSet {
			ls = ls.UnderlineStyle(s.UnderlineStyle)
		}
		if s.UnderlineColor != "" {
			ls = ls.UnderlineColor(lipgloss.Color(s.UnderlineColor))
		}
		hasStyle = true
	}
	if s.Strikethrough {
		ls = ls.Strikethrough(true)
		hasStyle = true
	}
	if s.Faint {
		ls = ls.Faint(true)
		hasStyle = true
	}
	if s.Blink {
		ls = ls.Blink(true)
		hasStyle = true
	}
	if s.FgColor != "" {
		ls = ls.Foreground(lipgloss.Color(s.FgColor))
		hasStyle = true
	}
	if s.BgColor != "" {
		ls = ls.Background(lipgloss.Color(s.BgColor))
		hasStyle = true
	}

	if hasStyle {
		return ls.Render(text)
	}
	return text
}

func (m Model) renderInline(node *RenderNode, currentX, currentY int, overlays *[]ImageOverlay, links *[]string) string {
	if node == nil {
		return ""
	}
	if node.Tag == "spinner" || node.Style.Spinner {
		return Spinner(m.Step)
	}

	if node.Type == NodeText {
		if node.Style.Shimmer {
			base := node.Style.FgColor
			if base == "" {
				base = "#44475a"
			}
			highlight := node.Style.ShimmerHighlight
			if highlight == "" {
				highlight = "#ffffff"
			}
			return Shimmer(node.Text, m.Step, ShimmerOptions{
				BaseColor:      base,
				HighlightColor: highlight,
				Bold:           node.Style.Bold,
			})
		}
		return applyInlineStyle(node.Style, node.Text)
	}

	if node.Tag == "br" {
		return "\n"
	}

	if node.Type == NodeImage {
		return m.renderImageNode(node, currentX, currentY, overlays)
	}

	var sb strings.Builder
	for _, child := range node.Children {
		sb.WriteString(m.renderInline(child, currentX, currentY, overlays, links))
	}
	content := applyInlineStyle(node.Style, sb.String())
	if node.Tag == "a" && node.Href != "" && !m.DisableHyperlinks && links != nil {
		id := len(*links)
		*links = append(*links, node.Href)
		startMarker := fmt.Sprintf("\x1b[999000;%dm", id)
		endMarker := fmt.Sprintf("\x1b[999001;%dm", id)
		content = startMarker + content + endMarker
	}
	return content
}

func (m Model) renderNode(node *RenderNode, currentX, currentY int, overlays *[]ImageOverlay, targetHeight int, links *[]string) string {
	if node == nil {
		return ""
	}

	if node.Tag == "hr" {
		w := node.ContentWidth
		if w <= 0 {
			w = 80
		}
		rule := strings.Repeat("─", w)
		if node.Style.BorderColor != "" {
			rule = lipgloss.NewStyle().Foreground(lipgloss.Color(node.Style.BorderColor)).Render(rule)
		}
		return applyLipglossStyle(node.Style, node.LipglossWidth, targetHeight, rule)
	}

	if node.Tag == "pre" {
		text := node.Text
		if !m.DisableSyntaxHighlight {
			text = highlightCode(text, node.Lang, m.ColorMode, m.codeStyle)
		}
		return applyLipglossStyle(node.Style, node.LipglossWidth, targetHeight, text)
	}

	if node.Type == NodeText {
		return node.Text
	}

	if node.Type == NodeImage {
		return m.renderImageNode(node, currentX, currentY, overlays)
	}

	if node.Tag == "ul" || node.Tag == "ol" {
		return m.renderListNode(node, targetHeight, overlays, links)
	}

	if node.Tag == "table" {
		return m.renderTable(node, targetHeight, overlays, links)
	}

	var content string
	if node.Style.Display == DisplayFlex {
		type flexSubRow struct {
			children []*RenderNode
			spans    []int
			spanSum  int
		}

		var subRows []flexSubRow
		currentSubRow := flexSubRow{}

		for _, child := range node.Children {
			span := child.Style.ColSpan
			if span <= 0 {
				span = 12
			}
			if len(currentSubRow.children) > 0 && currentSubRow.spanSum+span > 12 {
				subRows = append(subRows, currentSubRow)
				currentSubRow = flexSubRow{}
			}
			currentSubRow.children = append(currentSubRow.children, child)
			currentSubRow.spans = append(currentSubRow.spans, span)
			currentSubRow.spanSum += span
		}
		if len(currentSubRow.children) > 0 {
			subRows = append(subRows, currentSubRow)
		}

		var rowOutputs []string
		for _, row := range subRows {
			// Pass 1: measure child heights
			renderedItems := make([]string, len(row.children))
			maxHeight := 0
			for i, child := range row.children {
				out := m.renderNode(child, currentX, currentY, overlays, 0, links)
				renderedItems[i] = out
				if out != "" {
					h := strings.Count(out, "\n") + 1
					if h > maxHeight {
						maxHeight = h
					}
				}
			}

			// align-items: stretch (default) equalises heights so columns fill the
			// row; center/end instead keep natural heights and let JoinHorizontal
			// place each column on the cross axis.
			crossPos := lipgloss.Top
			stretch := true
			if node.Style.AlignItemsSet && node.Style.AlignItems != lipgloss.Top {
				crossPos = node.Style.AlignItems
				stretch = false
			}

			var itemOutputs []string
			if stretch && len(row.children) > 1 && maxHeight > 0 {
				for _, child := range row.children {
					// Re-render with target height
					out := m.renderNode(child, currentX, currentY, overlays, maxHeight, links)
					if out != "" {
						itemOutputs = append(itemOutputs, out)
					}
				}
			} else {
				for _, out := range renderedItems {
					if out != "" {
						itemOutputs = append(itemOutputs, out)
					}
				}
			}

			var rowOut string
			if len(itemOutputs) == 1 {
				rowOut = itemOutputs[0]
			} else if len(itemOutputs) > 1 {
				rowOut = lipgloss.JoinHorizontal(crossPos, itemOutputs...)
			} else {
				continue
			}
			// justify-content positions a row narrower than the content box.
			if node.Style.JustifyContentSet && node.Style.JustifyContent != lipgloss.Left && node.ContentWidth > lipgloss.Width(rowOut) {
				rowOut = lipgloss.PlaceHorizontal(node.ContentWidth, node.Style.JustifyContent, rowOut)
			}
			rowOutputs = append(rowOutputs, rowOut)
		}
		content = lipgloss.JoinVertical(lipgloss.Left, rowOutputs...)
	} else {
		// Block
		blockX := currentX + node.Style.MarginLeft + node.Style.PaddingLeft
		blockY := currentY + node.Style.MarginTop + node.Style.PaddingTop
		var childOutputs []string
		var inlineRun []*RenderNode

		flushInlineRun := func() {
			if len(inlineRun) == 0 {
				return
			}
			var sb strings.Builder
			for _, inlineNode := range inlineRun {
				sb.WriteString(m.renderInline(inlineNode, blockX, blockY, overlays, links))
			}
			inlineRun = nil

			rawText := sb.String()
			trimmed := strings.TrimSpace(rawText)
			if trimmed == "" {
				return
			}

			wrapped := wrapText(trimmed, node.ContentWidth)
			if node.Style.TextAlign == lipgloss.Center {
				lines := strings.Split(wrapped, "\n")
				for i, l := range lines {
					lines[i] = lipgloss.NewStyle().Width(node.ContentWidth).Align(lipgloss.Center).Render(l)
				}
				wrapped = strings.Join(lines, "\n")
			} else if node.Style.TextAlign == lipgloss.Right {
				lines := strings.Split(wrapped, "\n")
				for i, l := range lines {
					lines[i] = lipgloss.NewStyle().Width(node.ContentWidth).Align(lipgloss.Right).Render(l)
				}
				wrapped = strings.Join(lines, "\n")
			}

			if node.Tag == "li" {
				lines := strings.Split(wrapped, "\n")
				for i, l := range lines {
					if i == 0 {
						lines[i] = "• " + l
					} else {
						lines[i] = "  " + l
					}
				}
				wrapped = strings.Join(lines, "\n")
			}

			childOutputs = append(childOutputs, wrapped)
			blockY += strings.Count(wrapped, "\n") + 1
		}

		for _, child := range node.Children {
			if child.Style.Display == DisplayInline || child.Type == NodeText {
				inlineRun = append(inlineRun, child)
			} else {
				flushInlineRun()
				childTargetHeight := 0
				if len(node.Children) == 1 && targetHeight > 0 {
					childTargetHeight = targetHeight - node.Style.MarginTop - node.Style.MarginBottom - node.Style.PaddingTop - node.Style.PaddingBottom
					if node.Style.Border {
						childTargetHeight -= 2
					} else {
						if node.Style.BorderTop {
							childTargetHeight--
						}
						if node.Style.BorderBottom {
							childTargetHeight--
						}
					}
					if childTargetHeight < 0 {
						childTargetHeight = 0
					}
				}
				out := m.renderNode(child, blockX, blockY, overlays, childTargetHeight, links)
				if out != "" {
					childOutputs = append(childOutputs, out)
					blockY += strings.Count(out, "\n") + 1
				}
			}
		}
		flushInlineRun()
		content = lipgloss.JoinVertical(lipgloss.Left, childOutputs...)
	}
	out := applyLipglossStyle(node.Style, node.LipglossWidth, targetHeight, content)
	if node.Tag == "a" && node.Href != "" && !m.DisableHyperlinks && links != nil {
		id := len(*links)
		*links = append(*links, node.Href)
		startMarker := fmt.Sprintf("\x1b[999000;%dm", id)
		endMarker := fmt.Sprintf("\x1b[999001;%dm", id)
		out = startMarker + out + endMarker
	}
	return out
}

func (m Model) renderImageNode(node *RenderNode, x, y int, overlays *[]ImageOverlay) string {
	if m.DisableImages {
		if node.Alt != "" {
			return "[" + node.Alt + "](" + node.Src + ")"
		}
		return "[Image](" + node.Src + ")"
	}

	img, err := m.loadImage(node.Src)
	if err != nil {
		if node.Alt != "" {
			return "[" + node.Alt + "]"
		}
		return "[Image]"
	}

	renderer := m.imageRenderer()
	imgWidth := node.ContentWidth
	if m.MaxImageWidth > 0 && imgWidth > m.MaxImageWidth {
		imgWidth = m.MaxImageWidth
	}
	if imgWidth < 1 {
		imgWidth = 1
	}

	rendered, err := renderer.Render(img, imgWidth)
	if err != nil {
		return "[Image error: " + node.Alt + "]"
	}

	lines := strings.Split(rendered, "\n")
	imgHeight := len(lines)

	if renderer.Protocol().IsGraphic() && overlays != nil {
		*overlays = append(*overlays, ImageOverlay{
			Row:    y,
			Col:    x,
			Width:  imgWidth,
			Height: imgHeight,
			Escape: rendered,
		})
		// Reserve space with blank lines for overlay
		blankLine := strings.Repeat(" ", imgWidth)
		var blanks []string
		for range imgHeight {
			blanks = append(blanks, blankLine)
		}
		return strings.Join(blanks, "\n")
	}

	return rendered
}

func (m Model) loadImage(src string) (image.Image, error) {
	if strings.HasPrefix(src, "http://") || strings.HasPrefix(src, "https://") {
		resp, err := http.Get(src)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		img, _, err := image.Decode(resp.Body)
		return img, err
	}

	path := src
	if m.RootPath != "" && !filepath.IsAbs(path) {
		path = filepath.Join(m.RootPath, path)
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	return img, err
}

func (m Model) imageRenderer() termimage.Renderer {
	if m.ImageRenderer != nil {
		return m.ImageRenderer
	}
	opt := true
	if m.OptimizeSequences != nil {
		opt = *m.OptimizeSequences
	}
	return termimage.NewRenderer(
		termimage.WithColorMode(m.ColorMode),
		termimage.WithOptimizeSequences(opt),
	)
}

// borderForStyle resolves a lipgloss border from the computed border-style and
// border-radius. Dashed/dotted borders are built from box-drawing glyphs since
// lipgloss has no built-in for them.
func borderForStyle(s ComputedStyle) lipgloss.Border {
	switch s.BorderStyle {
	case "double":
		return lipgloss.DoubleBorder()
	case "thick":
		return lipgloss.ThickBorder()
	case "block":
		return lipgloss.BlockBorder()
	case "outer-half-block":
		return lipgloss.OuterHalfBlockBorder()
	case "inner-half-block":
		return lipgloss.InnerHalfBlockBorder()
	case "markdown":
		return lipgloss.MarkdownBorder()
	case "ascii":
		return lipgloss.ASCIIBorder()
	case "hidden", "none":
		return lipgloss.HiddenBorder()
	case "dashed":
		b := lipgloss.NormalBorder()
		b.Top, b.Bottom, b.Left, b.Right = "╌", "╌", "╎", "╎"
		return b
	case "dotted":
		b := lipgloss.NormalBorder()
		b.Top, b.Bottom, b.Left, b.Right = "┄", "┄", "┆", "┆"
		return b
	}
	if s.Rounded {
		return lipgloss.RoundedBorder()
	}
	return lipgloss.NormalBorder()
}

func applyLipglossStyle(s ComputedStyle, width, height int, content string) string {
	ls := lipgloss.NewStyle()

	if width > 0 && s.Display != DisplayInline {
		ls = ls.Width(width)
	}
	if height > 0 && s.Display != DisplayInline {
		borderV := 0
		if s.Border {
			borderV = 2
		} else {
			if s.BorderTop {
				borderV++
			}
			if s.BorderBottom {
				borderV++
			}
		}
		paddingV := s.PaddingTop + s.PaddingBottom
		marginV := s.MarginTop + s.MarginBottom
		shadowV := shadowBottomRows(s.Shadow)
		innerH := height - borderV - paddingV - marginV - shadowV
		if innerH > 0 {
			ls = ls.Height(innerH)
		}
	}

	// Margins
	ls = ls.Margin(s.MarginTop, s.MarginRight, s.MarginBottom, s.MarginLeft)

	// Paddings
	ls = ls.Padding(s.PaddingTop, s.PaddingRight, s.PaddingBottom, s.PaddingLeft)

	// Borders
	if s.Border || s.BorderTop || s.BorderBottom || s.BorderLeft || s.BorderRight {
		ls = ls.BorderStyle(borderForStyle(s))
		if !s.Border {
			ls = ls.BorderTop(s.BorderTop).
				BorderBottom(s.BorderBottom).
				BorderLeft(s.BorderLeft).
				BorderRight(s.BorderRight)
		}
	}

	if len(s.BorderBlend) >= 2 {
		colors := make([]color.Color, len(s.BorderBlend))
		for i, c := range s.BorderBlend {
			colors[i] = lipgloss.Color(c)
		}
		ls = ls.BorderForegroundBlend(colors...)
	} else if s.BorderColor != "" {
		ls = ls.BorderForeground(lipgloss.Color(s.BorderColor))
	}
	// Per-side border colors override the shared border color.
	if s.BorderTopColor != "" {
		ls = ls.BorderTopForeground(lipgloss.Color(s.BorderTopColor))
	}
	if s.BorderRightColor != "" {
		ls = ls.BorderRightForeground(lipgloss.Color(s.BorderRightColor))
	}
	if s.BorderBottomColor != "" {
		ls = ls.BorderBottomForeground(lipgloss.Color(s.BorderBottomColor))
	}
	if s.BorderLeftColor != "" {
		ls = ls.BorderLeftForeground(lipgloss.Color(s.BorderLeftColor))
	}

	// Colors
	if s.BgColor != "" {
		ls = ls.Background(lipgloss.Color(s.BgColor))
		// CSS border-box: the element background extends under the border, so the
		// border cells must share the fill color. Without this the border glyphs
		// render on the terminal's default background, leaving dark seams at every
		// boundary between colored boxes (most visible where nested column borders
		// stack up in a gutter).
		ls = ls.BorderBackground(lipgloss.Color(s.BgColor))
	}
	if s.FgColor != "" {
		ls = ls.Foreground(lipgloss.Color(s.FgColor))
	}

	// Typography
	if s.Bold {
		ls = ls.Bold(true)
	}
	if s.Italic {
		ls = ls.Italic(true)
	}
	if s.Underline {
		ls = ls.Underline(true)
		if s.UnderlineStyleSet {
			ls = ls.UnderlineStyle(s.UnderlineStyle)
		}
		if s.UnderlineColor != "" {
			ls = ls.UnderlineColor(lipgloss.Color(s.UnderlineColor))
		}
	}
	if s.Strikethrough {
		ls = ls.Strikethrough(true)
	}
	if s.Faint {
		ls = ls.Faint(true)
	}
	if s.Blink {
		ls = ls.Blink(true)
	}
	if s.MaxHeight > 0 && s.Display != DisplayInline {
		ls = ls.MaxHeight(s.MaxHeight)
	}

	out := ls.Render(content)
	if len(s.BgGradient) >= 2 {
		out = applyBackgroundGradient(out, s.BgGradient)
	} else if s.BgColor != "" {
		out = persistBackgroundColor(out, s.BgColor)
	}
	if s.Shadow > 0 {
		out = applyShadowWithWidth(out, s.Shadow, 0)
	}
	return out
}

func getBackgroundSequence(colorStr string) string {
	if colorStr == "" {
		return ""
	}
	rendered := lipgloss.NewStyle().Background(lipgloss.Color(colorStr)).Render(" ")
	idx := strings.Index(rendered, " ")
	if idx > 0 {
		return rendered[:idx]
	}
	return ""
}

func persistBackgroundColor(out string, bgColor string) string {
	if bgColor == "" || out == "" {
		return out
	}
	bgSeq := getBackgroundSequence(bgColor)
	if bgSeq == "" {
		return out
	}

	lines := strings.Split(out, "\n")
	bgCode := strings.TrimSuffix(strings.TrimPrefix(bgSeq, "\x1b["), "m")
	for i, line := range lines {
		if !strings.Contains(line, bgSeq) && !strings.Contains(line, bgCode) {
			continue
		}
		line = strings.ReplaceAll(line, "\x1b[49m", bgSeq)

		lastReset := strings.LastIndex(line, "\x1b[0m")
		if lastReset == -1 {
			lastReset = strings.LastIndex(line, "\x1b[m")
			if lastReset == -1 {
				lines[i] = line
				continue
			}
		}

		prefix := line[:lastReset]
		suffix := line[lastReset:]

		prefix = strings.ReplaceAll(prefix, "\x1b[0m", "\x1b[0m"+bgSeq)
		prefix = strings.ReplaceAll(prefix, "\x1b[m", "\x1b[m"+bgSeq)
		prefix = strings.ReplaceAll(prefix, bgSeq+bgSeq, bgSeq)

		lines[i] = prefix + suffix
	}
	return strings.Join(lines, "\n")
}
func applyBackgroundGradient(out string, stopHexes []string) string {
	if out == "" || len(stopHexes) < 2 {
		return out
	}
	stops := make([]color.Color, len(stopHexes))
	for i, h := range stopHexes {
		stops[i] = lipgloss.Color(h)
	}

	lines := strings.Split(out, "\n")
	w := 0
	for _, l := range lines {
		if lw := lipgloss.Width(l); lw > w {
			w = lw
		}
	}
	if w <= 0 {
		return out
	}

	blended := lipgloss.Blend1D(w, stops...)
	bgSeqs := make([]string, len(blended))
	for i, c := range blended {
		r, g, b, _ := c.RGBA()
		bgSeqs[i] = fmt.Sprintf("\x1b[48;2;%d;%d;%dm", r>>8, g>>8, b>>8)
	}

	for lineIdx, line := range lines {
		var sb strings.Builder
		col := 0
		i := 0
		for i < len(line) {
			if loc := ansiRegex.FindStringIndex(line[i:]); loc != nil && loc[0] == 0 {
				sb.WriteString(line[i : i+loc[1]])
				i += loc[1]
				continue
			}
			r := []rune(line[i:])
			ch := string(r[0])
			if col < len(bgSeqs) {
				sb.WriteString(bgSeqs[col])
			}
			sb.WriteString(ch)
			col += lipgloss.Width(ch)
			i += len(ch)
		}
		sb.WriteString("\x1b[0m")
		lines[lineIdx] = sb.String()
	}
	return strings.Join(lines, "\n")
}

func applyShadowWithWidth(content string, shadowSize, maxWidth int) string {
	if shadowSize <= 0 {
		return content
	}

	lines := strings.Split(content, "\n")
	start := 0
	for start < len(lines) && strings.TrimSpace(lines[start]) == "" {
		start++
	}
	end := len(lines)
	for end > start && strings.TrimSpace(lines[end-1]) == "" {
		end--
	}
	leading := lines[:start]
	visible := lines[start:end]
	trailing := lines[end:]

	if len(visible) == 0 {
		return content
	}

	contentWidth := lipgloss.Width(visible[0])
	effectiveShadow := shadowSize
	if maxWidth > 0 && contentWidth+shadowSize > maxWidth {
		effectiveShadow = maxWidth - contentWidth
		if effectiveShadow < 1 {
			effectiveShadow = 1
		}
	}

	shadowChar := "░"
	if effectiveShadow >= 3 {
		shadowChar = "▒"
	}

	shadowStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	result := make([]string, 0, len(lines)+effectiveShadow)
	result = append(result, leading...)

	for _, line := range visible {
		shadow := shadowStyle.Render(strings.Repeat(shadowChar, effectiveShadow))
		result = append(result, line+shadow)
	}

	bottomShadow := shadowStyle.Render(strings.Repeat(shadowChar, contentWidth))
	bottomRows := shadowBottomRows(effectiveShadow)
	for range bottomRows {
		result = append(result, strings.Repeat(" ", effectiveShadow)+bottomShadow)
	}

	result = append(result, trailing...)
	return strings.Join(result, "\n")
}

// shadowBottomRows returns the number of bottom rows a drop shadow of the given
// horizontal width occupies. Terminal cells are roughly twice as tall as they
// are wide, so a bottom shadow of N rows reads about twice as thick as a right
// shadow of N columns; halving the row count keeps the shadow visually
// balanced. Single source of truth for both the layout height reservation and
// the rendered output so they never disagree.
func shadowBottomRows(shadowSize int) int {
	if shadowSize <= 0 {
		return 0
	}
	return (shadowSize + 1) / 2
}

func boolStr(b bool, str string) string {
	if b {
		return str
	}
	return ""
}
func wrapText(text string, width int) string {
	if width <= 0 {
		return text
	}
	w := wordwrap.NewWriter(width)
	w.Breakpoints = []rune{' '}
	w.KeepNewlines = true
	_, _ = w.Write([]byte(text))
	_ = w.Close()
	out := w.String()

	var lines []string
	for _, line := range strings.Split(out, "\n") {
		if lipgloss.Width(line) > width {
			// wrap.String wraps without spaces. We must ensure that each wrapped line starts with the active ANSI sequences.
			wrappedChunk := wrap.String(line, width)
			chunkLines := strings.Split(wrappedChunk, "\n")
			var activeAnsi string
			for ci, cl := range chunkLines {
				if ci > 0 && activeAnsi != "" {
					cl = activeAnsi + cl
				}
				// Extract all ANSI escape sequences from cl to track active styles
				for _, seq := range ansiRegex.FindAllString(cl, -1) {
					if seq == "\x1b[0m" || seq == "\x1b[m" {
						activeAnsi = ""
					} else if strings.HasPrefix(seq, "\x1b[") && !strings.HasPrefix(seq, "\x1b[999") {
						activeAnsi = seq
					}
				}
				lines = append(lines, cl)
			}
		} else {
			lines = append(lines, line)
		}
	}
	return strings.Join(lines, "\n")
}
