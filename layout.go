package termstrap

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/charmbracelet/lipgloss"
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
func BuildRenderTree(sel *goquery.Selection, matcher *CSSMatcher, termWidth int) *RenderNode {
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

		return &RenderNode{
			Type:  NodeText,
			Text:  sb.String(),
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

	style := matcher.ComputeStyleForSelection(sel, termWidth)
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

	if tag == "img" {
		rNode.Type = NodeImage
		rNode.Src, _ = sel.Attr("src")
		rNode.Alt, _ = sel.Attr("alt")
		return rNode
	}

	if tag == "pre" {
		rNode.Text = sel.Text()
		return rNode
	}

	// Recursively build children
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		childSel := sel.FindNodes(c)
		childRNode := BuildRenderTree(childSel, matcher, termWidth)
		if childRNode != nil {
			if rNode.Style.Display == DisplayFlex && childRNode.Type == NodeText && strings.TrimSpace(childRNode.Text) == "" {
				continue
			}
			// Inherit CSS properties (color, text-align)
			if childRNode.Style.FgColor == "" && rNode.Style.FgColor != "" {
				childRNode.Style.FgColor = rNode.Style.FgColor
			}
			if childRNode.Style.TextAlign == lipgloss.Left && rNode.Style.TextAlign != lipgloss.Left {
				childRNode.Style.TextAlign = rNode.Style.TextAlign
			}
			rNode.Children = append(rNode.Children, childRNode)
		}
	}

	return rNode
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
