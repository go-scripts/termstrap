package termstrap

import (
	"strings"

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
	Style    ComputedStyle
	Children []*RenderNode

	// Layout Geometry (calculated in character cells)
	AllocatedWidth int
	ContentWidth   int
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
		text := strings.TrimSpace(node.Data)
		if text == "" {
			return nil
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

	style := matcher.ComputeStyleForSelection(sel, termWidth)
	if style.Display == DisplayNone {
		return nil
	}

	if tag == "tr" {
		style.Display = DisplayFlex
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

	// Recursively build children
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		childSel := sel.FindNodes(c)
		childRNode := BuildRenderTree(childSel, matcher, termWidth)
		if childRNode != nil {
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

	// Reduce available width by margins and borders
	avail := parentWidth - node.Style.MarginLeft - node.Style.MarginRight
	if node.Style.Border || node.Style.BorderLeft {
		avail--
	}
	if node.Style.Border || node.Style.BorderRight {
		avail--
	}
	if node.Style.Shadow > 0 {
		avail -= node.Style.Shadow
	}
	if avail < 1 {
		avail = 1
	}

	node.AllocatedWidth = parentWidth
	node.ContentWidth = avail - node.Style.PaddingLeft - node.Style.PaddingRight
	if node.ContentWidth < 1 {
		node.ContentWidth = 1
	}

	if node.Style.Display == DisplayFlex {
		// Calculate columns total span
		totalSpan := 0
		autoCount := 0
		for _, child := range node.Children {
			if child.Style.ColSpan > 0 {
				totalSpan += child.Style.ColSpan
			} else {
				autoCount++
			}
		}

		if totalSpan > 12 {
			// Wrapping / Stacking condition: columns exceed 12
			for _, child := range node.Children {
				computeWidths(child, node.ContentWidth)
			}
		} else {
			remainingWidth := node.ContentWidth
			for _, child := range node.Children {
				var colWidth int
				if child.Style.ColSpan > 0 {
					colWidth = (node.ContentWidth * child.Style.ColSpan) / 12
				} else if autoCount > 0 {
					colWidth = remainingWidth / autoCount
				} else {
					colWidth = remainingWidth
				}
				if colWidth < 1 {
					colWidth = 1
				}
				computeWidths(child, colWidth)
			}
		}
	} else {
		// Block distribution
		for _, child := range node.Children {
			computeWidths(child, node.ContentWidth)
		}
	}
}
