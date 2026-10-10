package termstrap

import (
	"strings"
	"unicode"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

// NodeType identifies the kind of DOM node in the render tree.
type NodeType int

const (
	NodeElement NodeType = iota
	NodeText
	NodeImage
)

// RenderNode represents a box or text node in the terminal layout tree.
type RenderNode struct {
	Type     NodeType
	Tag      string
	Text     string
	Src      string
	Alt      string
	Href     string // URL for <a> tags
	Lang     string // Source language hint for <pre> code blocks (empty = auto/none)
	IsTree   bool   // ul/ol flagged as a tree (class="tree" or role="tree")
	Style    ComputedStyle
	Children []*RenderNode
	// Layout Geometry (calculated in character cells)
	AllocatedWidth int // Parent allocated width
	BoxWidth       int // Width between outer border edges (AllocatedWidth - Margins - Shadow)
	LipglossWidth  int // Width passed to lipgloss.Width() (BoxWidth - Borders)
	ContentWidth   int // Inner width for text and children (LipglossWidth - Padding)
	Height         int
	X              int // Relative offset in parent
	Y              int
}

// BuildRenderTree recursively converts a goquery DOM tree into a RenderTree with resolved styles.
func BuildRenderTree(sel *goquery.Selection, matcher *CSSMatcher, termWidth int, parentStyle *ComputedStyle) *RenderNode {
	node := sel.Get(0)
	if node == nil {
		return nil
	}

	if node.Type == html.TextNode {
		raw := node.Data
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" {
			if len(raw) > 0 {
				return &RenderNode{
					Type:  NodeText,
					Text:  " ",
					Style: ComputedStyle{Display: DisplayInline},
				}
			}
			return nil
		}

		words := strings.Fields(raw)
		hasLeading := raw[0] == ' ' || raw[0] == '\t' || raw[0] == '\n' || raw[0] == '\r'
		hasTrailing := raw[len(raw)-1] == ' ' || raw[len(raw)-1] == '\t' || raw[len(raw)-1] == '\n' || raw[len(raw)-1] == '\r'

		var sb strings.Builder
		if hasLeading {
			sb.WriteString(" ")
		}
		sb.WriteString(strings.Join(words, " "))
		if hasTrailing {
			sb.WriteString(" ")
		}

		text := sb.String()
		if parentStyle != nil {
			text = applyTextTransform(text, parentStyle.TextTransform)
		}
		return &RenderNode{
			Type:  NodeText,
			Text:  text,
			Style: ComputedStyle{Display: DisplayInline},
		}
	}

	if node.Type != html.ElementNode {
		return nil
	}

	tag := strings.ToLower(node.Data)
	if tag == "head" || tag == "style" || tag == "script" || tag == "meta" || tag == "title" {
		return nil
	}

	style := matcher.ComputeStyleForSelection(sel, termWidth, parentStyle)
	if style.Display == DisplayNone {
		return nil
	}

	if tag == "tr" {
		style.Display = DisplayFlex
	}
	if tag == "th" || tag == "td" {
		style.ColAuto = true
	}
	if tag == "br" {
		style.Display = DisplayInline
	}

	rNode := &RenderNode{
		Type:  NodeElement,
		Tag:   tag,
		Style: style,
	}

	if href, ok := sel.Attr("href"); ok {
		rNode.Href = href
	}

	if tag == "ul" || tag == "ol" {
		classAttr, _ := sel.Attr("class")
		role, _ := sel.Attr("role")
		if role == "tree" {
			rNode.IsTree = true
		} else {
			for _, c := range strings.Fields(classAttr) {
				if c == "tree" {
					rNode.IsTree = true
					break
				}
			}
		}
	}

	if tag == "img" {
		rNode.Type = NodeImage
		rNode.Src, _ = sel.Attr("src")
		rNode.Alt, _ = sel.Attr("alt")
		return rNode
	}

	if tag == "pre" {
		rNode.Text = sel.Text()
		// Resolve a language hint from the inner <code class="language-xxx">,
		// then the <pre>'s own class, following the HTML/Markdown convention.
		preClass, _ := sel.Attr("class")
		codeClass, _ := sel.Find("code").First().Attr("class")
		rNode.Lang = resolveLang(codeClass, preClass)
		return rNode
	}

	// ::before pseudo-element (generated content, prepended inline)
	if beforeText, beforeStyle, ok := matcher.PseudoText(sel, termWidth, &style, "before"); ok {
		beforeText = applyTextTransform(beforeText, beforeStyle.TextTransform)
		rNode.Children = append(rNode.Children, &RenderNode{
			Type:  NodeText,
			Text:  beforeText,
			Style: beforeStyle,
		})
	}

	// Recursively build children
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		childSel := sel.FindNodes(c)
		childRNode := BuildRenderTree(childSel, matcher, termWidth, &style)
		if childRNode != nil {
			if rNode.Style.Display == DisplayFlex && childRNode.Type == NodeText && strings.TrimSpace(childRNode.Text) == "" {
				continue
			}
			rNode.Children = append(rNode.Children, childRNode)
		}
	}

	// ::after pseudo-element (generated content, appended inline)
	if afterText, afterStyle, ok := matcher.PseudoText(sel, termWidth, &style, "after"); ok {
		afterText = applyTextTransform(afterText, afterStyle.TextTransform)
		rNode.Children = append(rNode.Children, &RenderNode{
			Type:  NodeText,
			Text:  afterText,
			Style: afterStyle,
		})
	}

	return rNode
}

// applyTextTransform applies a CSS text-transform to raw (ANSI-free) text.
func applyTextTransform(text, transform string) string {
	switch transform {
	case "uppercase":
		return strings.ToUpper(text)
	case "lowercase":
		return strings.ToLower(text)
	case "capitalize":
		var b strings.Builder
		b.Grow(len(text))
		prevSpace := true
		for _, r := range text {
			if prevSpace && unicode.IsLetter(r) {
				b.WriteRune(unicode.ToUpper(r))
			} else {
				b.WriteRune(r)
			}
			prevSpace = unicode.IsSpace(r)
		}
		return b.String()
	default:
		return text
	}
}

// ComputeLayout performs layout calculations on the RenderTree.
func ComputeLayout(root *RenderNode, availableWidth int) {
	if root == nil || availableWidth <= 0 {
		return
	}

	computeWidths(root, availableWidth)
}

func computeWidths(node *RenderNode, parentWidth int) {
	if node == nil {
		return
	}

	// Reduce available width by margins and shadow
	avail := parentWidth - node.Style.MarginLeft - node.Style.MarginRight - node.Style.Shadow
	if avail < 1 {
		avail = 1
	}

	if node.Style.Width > 0 {
		explicitW := node.Style.Width - node.Style.MarginLeft - node.Style.MarginRight - node.Style.Shadow
		if explicitW > 0 && explicitW < avail {
			avail = explicitW
		}
	}
	if node.Style.MaxWidth > 0 {
		maxW := node.Style.MaxWidth - node.Style.MarginLeft - node.Style.MarginRight - node.Style.Shadow
		if maxW > 0 && maxW < avail {
			avail = maxW
		}
	}
	borderH := 0
	if node.Style.Border || node.Style.BorderLeft {
		borderH++
	}
	if node.Style.Border || node.Style.BorderRight {
		borderH++
	}

	lipglossW := avail - borderH
	if lipglossW < 1 {
		lipglossW = 1
	}

	paddingH := node.Style.PaddingLeft + node.Style.PaddingRight
	contentW := lipglossW - paddingH
	if contentW < 1 {
		contentW = 1
	}

	node.AllocatedWidth = parentWidth
	node.BoxWidth = avail
	node.LipglossWidth = lipglossW
	node.ContentWidth = contentW
	if node.Style.Display == DisplayFlex {
		// Calculate count of auto columns
		autoCount := 0
		for _, child := range node.Children {
			if child.Style.ColSpan == 0 || child.Style.ColAuto {
				autoCount++
			}
		}

		// Resolve column spans for all children
		for _, child := range node.Children {
			if child.Style.ColSpan == 0 || child.Style.ColAuto {
				if node.Style.RowCols > 0 {
					child.Style.ColSpan = 12 / node.Style.RowCols
				} else if autoCount > 0 && autoCount <= 12 {
					child.Style.ColSpan = 12 / autoCount
				} else {
					child.Style.ColSpan = 12
				}
			}
			if child.Style.ColSpan < 1 {
				child.Style.ColSpan = 1
			}
			if child.Style.ColSpan > 12 {
				child.Style.ColSpan = 12
			}
		}

		type flexSubRow struct {
			children []*RenderNode
			spans    []int
			spanSum  int
		}

		var subRows []flexSubRow
		currentSubRow := flexSubRow{}

		for _, child := range node.Children {
			span := child.Style.ColSpan
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

		for _, row := range subRows {
			targetWidth := (node.ContentWidth * row.spanSum) / 12
			if row.spanSum >= 12 {
				targetWidth = node.ContentWidth
			}
			remWidth := targetWidth
			remSpan := row.spanSum
			for i, child := range row.children {
				span := row.spans[i]
				colWidth := (remWidth * span) / remSpan
				if colWidth < 1 {
					colWidth = 1
				}
				remWidth -= colWidth
				remSpan -= span
				computeWidths(child, colWidth)
			}
		}
	} else {
		for _, child := range node.Children {
			computeWidths(child, node.ContentWidth)
		}
	}
}
