package termstrap

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

// collectRows walks a <table> subtree (through thead/tbody/tfoot) and returns
// its <tr> rows in document order.
func collectRows(node *RenderNode, out *[]*RenderNode) {
	for _, c := range node.Children {
		switch c.Tag {
		case "tr":
			*out = append(*out, c)
		case "thead", "tbody", "tfoot":
			collectRows(c, out)
		}
	}
}

// rowCells returns the <th>/<td> cells of a <tr> and whether every cell is a <th>.
func rowCells(tr *RenderNode) (cells []*RenderNode, allHeader bool) {
	allHeader = true
	for _, c := range tr.Children {
		if c.Tag == "th" || c.Tag == "td" {
			cells = append(cells, c)
			if c.Tag != "th" {
				allHeader = false
			}
		}
	}
	if len(cells) == 0 {
		allHeader = false
	}
	return cells, allHeader
}

// renderCell renders a table cell's inner content to a (possibly multi-line)
// string without the cell's own padding; the table applies its own padding.
func (m Model) renderCell(cell *RenderNode, overlays *[]ImageOverlay, links *[]string) string {
	cp := *cell
	cp.Tag = "div"
	cp.Style.PaddingTop, cp.Style.PaddingBottom = 0, 0
	cp.Style.PaddingLeft, cp.Style.PaddingRight = 0, 0
	return strings.TrimRight(m.renderNode(&cp, 0, 0, overlays, 0, links), "\n")
}

// renderTable renders a <table> using the lipgloss table engine: a boxed layout
// with a header separator and no inner column rules, matching Bootstrap's clean
// table look. Per-row background colors (e.g. .table-primary) are preserved.
func (m Model) renderTable(node *RenderNode, targetHeight int, overlays *[]ImageOverlay, links *[]string) string {
	var trs []*RenderNode
	collectRows(node, &trs)

	var headers []string
	var rows [][]string
	var rowBg []string
	headerSet := false

	for _, tr := range trs {
		cells, allHeader := rowCells(tr)
		if len(cells) == 0 {
			continue
		}
		vals := make([]string, len(cells))
		for i, cell := range cells {
			vals[i] = m.renderCell(cell, overlays, links)
		}
		if allHeader && !headerSet {
			headers = vals
			headerSet = true
			continue
		}
		rows = append(rows, vals)
		rowBg = append(rowBg, tr.Style.BgColor)
	}

	t := table.New().
		Border(borderForStyle(node.Style)).
		BorderRow(false).
		BorderColumn(false).
		BorderHeader(true)

	if node.Style.BorderColor != "" {
		t = t.BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color(node.Style.BorderColor)))
	}
	if headerSet {
		t = t.Headers(headers...)
	}
	if len(rows) > 0 {
		t = t.Rows(rows...)
	}
	if node.LipglossWidth > 0 {
		t = t.Width(node.LipglossWidth)
	}

	headerFg := node.Style.FgColor
	t = t.StyleFunc(func(row, col int) lipgloss.Style {
		st := lipgloss.NewStyle().Padding(0, 1)
		if row == table.HeaderRow {
			st = st.Bold(true)
			if headerFg != "" {
				st = st.Foreground(lipgloss.Color(headerFg))
			}
			return st
		}
		if row >= 0 && row < len(rowBg) && rowBg[row] != "" {
			st = st.Background(lipgloss.Color(rowBg[row]))
		}
		return st
	})

	// Margins/colors from the table node itself; the table owns its border, so
	// pass width 0 to avoid double-constraining.
	st := node.Style
	st.Border, st.BorderTop, st.BorderBottom, st.BorderLeft, st.BorderRight = false, false, false, false, false
	return applyLipglossStyle(st, 0, targetHeight, t.String())
}
