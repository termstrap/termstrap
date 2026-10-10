package termstrap

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/list"
	"github.com/charmbracelet/lipgloss/tree"
)

// splitListItem separates an <li>'s inline/block content from any nested
// <ul>/<ol> sublists, so the content renders as the item label and the nested
// lists attach as sublists.
func splitListItem(li *RenderNode) (content, nested []*RenderNode) {
	for _, c := range li.Children {
		if c.Tag == "ul" || c.Tag == "ol" {
			nested = append(nested, c)
		} else {
			content = append(content, c)
		}
	}
	return content, nested
}

// renderItemLabel renders an <li>'s own content (excluding nested lists) to a
// string, reusing the block renderer but without the "li" bullet prefix.
func (m Model) renderItemLabel(li *RenderNode, content []*RenderNode, overlays *[]ImageOverlay, links *[]string) string {
	cp := *li
	cp.Tag = "div" // suppress the built-in "• " bullet; the list owns enumeration
	cp.Children = content
	return strings.TrimRight(m.renderNode(&cp, 0, 0, overlays, 0, links), "\n")
}

// enumeratorFor resolves the list enumerator from the tag and CSS list-style.
func enumeratorFor(node *RenderNode) list.Enumerator {
	switch node.Style.ListStyle {
	case "none":
		return func(list.Items, int) string { return " " }
	case "decimal":
		return list.Arabic
	case "lower-alpha", "lower-latin", "upper-alpha", "upper-latin":
		return list.Alphabet
	case "lower-roman", "upper-roman":
		return list.Roman
	case "circle":
		return func(list.Items, int) string { return "◦" }
	case "square":
		return func(list.Items, int) string { return "▪" }
	case "disc":
		return list.Bullet
	}
	if node.Tag == "ol" {
		return list.Arabic
	}
	return list.Bullet
}

// buildList converts a <ul>/<ol> RenderNode into a lipgloss list, recursing into
// nested lists as sublists.
func (m Model) buildList(node *RenderNode, overlays *[]ImageOverlay, links *[]string) *list.List {
	l := list.New().Enumerator(enumeratorFor(node))
	if node.Style.FgColor != "" {
		l = l.EnumeratorStyle(lipgloss.NewStyle().Foreground(lipgloss.Color(node.Style.FgColor)))
	}
	for _, li := range node.Children {
		if li.Tag != "li" {
			continue
		}
		content, nested := splitListItem(li)
		l.Item(m.renderItemLabel(li, content, overlays, links))
		for _, sub := range nested {
			l.Item(m.buildList(sub, overlays, links))
		}
	}
	return l
}

// appendTreeItems adds a <ul>/<ol>'s items to an existing tree, recursing into
// nested lists so sublists share the parent branch instead of an empty root.
func (m Model) appendTreeItems(t *tree.Tree, node *RenderNode, overlays *[]ImageOverlay, links *[]string) {
	for _, li := range node.Children {
		if li.Tag != "li" {
			continue
		}
		content, nested := splitListItem(li)
		label := m.renderItemLabel(li, content, overlays, links)
		if len(nested) == 0 {
			t.Child(label)
			continue
		}
		branch := tree.Root(label)
		for _, sub := range nested {
			m.appendTreeItems(branch, sub, overlays, links)
		}
		t.Child(branch)
	}
}

// buildTree converts a <ul>/<ol> RenderNode into a lipgloss tree.
func (m Model) buildTree(node *RenderNode, overlays *[]ImageOverlay, links *[]string) *tree.Tree {
	t := tree.New()
	if node.Style.FgColor != "" {
		t = t.EnumeratorStyle(lipgloss.NewStyle().Foreground(lipgloss.Color(node.Style.FgColor)))
	}
	m.appendTreeItems(t, node, overlays, links)
	return t
}

// renderListNode renders a <ul>/<ol> node (as a list or tree) and applies the
// node's own box styling (margins, padding, border, colors).
func (m Model) renderListNode(node *RenderNode, targetHeight int, overlays *[]ImageOverlay, links *[]string) string {
	var content string
	if node.IsTree {
		content = m.buildTree(node, overlays, links).String()
	} else {
		content = m.buildList(node, overlays, links).String()
	}
	return applyLipglossStyle(node.Style, node.LipglossWidth, targetHeight, content)
}
