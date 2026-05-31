package termstrap

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// renderHTMLLayout parses an HTML layout block and renders it to ANSI terminal output.
func (m Model) renderHTMLLayout(htmlContent string) (string, error) {
	rows, err := parseGrid(htmlContent, m.Width)
	if err != nil {
		return "", err
	}

	var result string
	for _, row := range rows {
		rendered, err := m.renderRow(row)
		if err != nil {
			return "", err
		}
		result += rendered
	}

	return result, nil
}

// renderRow renders a single grid row with its columns.
func (m Model) renderRow(row gridRow) (string, error) {
	bp := detectBreakpoint(m.Width)

	// Check if columns should stack (responsive collapse)
	shouldStack := shouldStackRow(row, bp)

	rowStyle := resolveStyle(row.Classes)
	renderedCols := make([]string, 0, len(row.Columns))

	for _, col := range row.Columns {
		rendered, err := m.renderColumn(col, shouldStack)
		if err != nil {
			return "", err
		}
		renderedCols = append(renderedCols, rendered)
	}

	var output string
	if shouldStack {
		// Stack columns vertically
		output = lipgloss.JoinVertical(lipgloss.Left, renderedCols...)
	} else {
		// Join columns horizontally
		output = lipgloss.JoinHorizontal(lipgloss.Top, renderedCols...)
	}

	// Apply row-level styling
	if hasStyling(rowStyle) {
		rowStyleWidth := m.Width - rowStyle.MarginLeft - rowStyle.MarginRight
		if rowStyle.Border || rowStyle.BorderLeft {
			rowStyleWidth -= 1
		}
		if rowStyle.Border || rowStyle.BorderRight {
			rowStyleWidth -= 1
		}
		rowLipgloss := applyStyle(rowStyle, rowStyleWidth)
		output = rowLipgloss.Render(output)
	}

	// Apply shadow AFTER lipgloss styling (so it's outside the border)
	// Use intelligent shadow that measures and adjusts to prevent overflow
	if rowStyle.Shadow > 0 {
		// Calculate available width accounting for all elements (margins + borders)
		// This ensures shadow respects the same boundaries as the content
		shadowMaxWidth := m.Width - rowStyle.MarginLeft - rowStyle.MarginRight
		if rowStyle.Border || rowStyle.BorderLeft {
			shadowMaxWidth -= 1
		}
		if rowStyle.Border || rowStyle.BorderRight {
			shadowMaxWidth -= 1
		}
		output = applyShadowIntelligent(output, rowStyle.Shadow, shadowMaxWidth)
	}

	return output + "\n", nil
}

// renderColumn renders a single column with its content and styling.
func (m Model) renderColumn(col gridColumn, stacked bool) (string, error) {
	colStyle := resolveStyle(col.Classes)

	// Determine column total allocated width
	totalWidth := col.Width
	if stacked {
		totalWidth = m.Width
	}

	// Width for lipgloss: includes padding but NOT border or margin
	styleWidth := totalWidth - colStyle.MarginLeft - colStyle.MarginRight
	if colStyle.Border || colStyle.BorderLeft {
		styleWidth -= 1
	}
	if colStyle.Border || colStyle.BorderRight {
		styleWidth -= 1
	}
	if styleWidth < 10 {
		styleWidth = 10
	}

	// Width for partial borders (before any shadow reduction)
	partialBorderWidth := styleWidth
	
	// Reserve space for shadow: reduce styleWidth BEFORE applying lipgloss styling
	// so the box itself is smaller and the shadow has room
	if colStyle.Shadow > 0 {
		styleWidth -= colStyle.Shadow
		if styleWidth < 10 {
			styleWidth = 10
		}
	}

	// Content width for markdown rendering (text area = styleWidth minus padding)
	contentWidth := styleWidth - colStyle.PaddingLeft - colStyle.PaddingRight
	if contentWidth < 10 {
		contentWidth = 10
	}

	// Render the column content (markdown + images)
	colModel := Model{
		Content:  col.Content,
		Width:    contentWidth,
		RootPath: m.RootPath,
	}
	rendered, err := colModel.renderMarkdown(col.Content)
	if err != nil {
		return "", err
	}

	// Trim leading and trailing blank lines for tighter column layout
	rendered = trimBlankLines(rendered)

	// Apply column styling via lipgloss (padding, border, colors)
	// Note: lipgloss borders (colStyle.Border) are applied here
	// Partial borders are applied separately below
	ls := applyStyle(colStyle, styleWidth)
	output := ls.Render(rendered)

	// Apply partial borders manually (border-top, border-left, etc.)
	// Do this AFTER lipgloss styling to avoid conflicts with column joining
	// Use partialBorderWidth which is styleWidth but doesn't account for shadow reduction
	if !colStyle.Border && (colStyle.BorderTop || colStyle.BorderBottom || colStyle.BorderLeft || colStyle.BorderRight) {
		output = applyPartialBorders(output, partialBorderWidth, colStyle.BorderTop, colStyle.BorderBottom, colStyle.BorderLeft, colStyle.BorderRight)
	}

	// Apply shadow AFTER border styling (so it's outside the border)
	// Use intelligent shadow that measures and adjusts to prevent overflow
	if colStyle.Shadow > 0 {
		// Available space for shadow must account for ALL width-reducing elements:
		// margins, borders (which were already subtracted from styleWidth)
		// The shadow max width should match styleWidth to not overflow
		shadowMaxWidth := styleWidth
		output = applyShadowIntelligent(output, colStyle.Shadow, shadowMaxWidth)
	}

	return output, nil
}

// shouldStackRow determines if a row's columns should stack vertically
// based on responsive breakpoints.
func shouldStackRow(row gridRow, bp breakpoint) bool {
	for _, col := range row.Columns {
		if isStacked(col.Classes, bp) {
			return true
		}
	}
	return false
}

// hasStyling returns true if any non-default styling is applied.
func hasStyling(s style) bool {
	return s.PaddingTop > 0 || s.PaddingBottom > 0 || s.PaddingLeft > 0 || s.PaddingRight > 0 ||
		s.MarginTop > 0 || s.MarginBottom > 0 || s.MarginLeft > 0 || s.MarginRight > 0 ||
		s.Border || s.BorderTop || s.BorderBottom || s.BorderLeft || s.BorderRight ||
		s.BgColor != "" || s.FgColor != "" || s.Bold
}

// applyShadowIntelligent applies a shadow with post-render overflow detection and adjustment.
// Measures actual rendered line widths and reduces shadow if it exceeds maxWidth.
// shadowSize: 1=small, 2=medium, 3=large
// maxWidth: maximum available width in terminal
func applyShadowIntelligent(content string, shadowSize, maxWidth int) string {
	if shadowSize <= 0 || maxWidth <= 0 {
		return content
	}

	lines := strings.Split(content, "\n")

	// Trim leading and trailing empty lines
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

	// Find the maximum width of visible lines
	maxContentWidth := 0
	for _, line := range visible {
		width := lipgloss.Width(line)
		if width > maxContentWidth {
			maxContentWidth = width
		}
	}

	// Determine effective shadow size (reduce if it would overflow)
	effectiveShadow := shadowSize
	if maxContentWidth+shadowSize > maxWidth {
		effectiveShadow = maxWidth - maxContentWidth
		if effectiveShadow < 1 {
			effectiveShadow = 1
		}
	}

	shadowChar := "░"
	if effectiveShadow >= 3 {
		shadowChar = "▒"
	}

	shadowStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	result := make([]string, 0, len(lines)+effectiveShadow)
	result = append(result, leading...)

	// Apply shadow to visible lines
	for _, line := range visible {
		shadow := shadowStyle.Render(strings.Repeat(shadowChar, effectiveShadow))
		result = append(result, line+shadow)
	}

	// Add bottom shadow rows - keep original shadowSize for line count
	// but use effectiveShadow for the actual width to prevent overflow
	bottomShadow := shadowStyle.Render(strings.Repeat(shadowChar, maxContentWidth+effectiveShadow))
	for i := 0; i < shadowSize; i++ {
		result = append(result, strings.Repeat(" ", effectiveShadow)+bottomShadow)
	}

	result = append(result, trailing...)
	return strings.Join(result, "\n")
}

// shadowMetrics pre-calculates shadow rendering information.
type shadowMetrics struct {
	ContentWidth      int  // Width of the actual visible content
	ShadowWidth       int  // Width of shadow on the right side
	BottomShadowWidth int  // Width of bottom shadow
	WillOverflow      bool // True if shadow would exceed maxWidth
	AdjustedShadow    int  // Adjusted shadow size if it would overflow
	TotalWidth        int  // Total width including shadow (content + shadow width)
}

// calculateShadowMetrics pre-calculates shadow dimensions for a given content and available width.
// contentWidth: width of the actual content (without shadow)
// shadowSize: requested shadow size (1=small, 2=medium, 3=large)
// maxWidth: maximum available width in the terminal
// Returns shadowMetrics with overflow detection and adjusted size if needed.
func calculateShadowMetrics(contentWidth, shadowSize, maxWidth int) shadowMetrics {
	metrics := shadowMetrics{
		ContentWidth: contentWidth,
		ShadowWidth:  shadowSize,
	}
	fmt.Printf("shadowSize: %d, contentWidth: %d, maxWidth: %d\n", shadowSize, contentWidth, maxWidth)
	// Calculate total width if shadow is rendered
	totalWithShadow := contentWidth + shadowSize
	fmt.Printf("totalWithShadow: %d\n", totalWithShadow)

	// Check if shadow would overflow
	if totalWithShadow > maxWidth {
		metrics.WillOverflow = true
		// Adjust shadow size to fit
		metrics.AdjustedShadow = maxWidth - contentWidth
		if metrics.AdjustedShadow < 1 {
			metrics.AdjustedShadow = 1
		}
		metrics.TotalWidth = contentWidth + metrics.AdjustedShadow
	} else {
		metrics.WillOverflow = false
		metrics.AdjustedShadow = shadowSize
		metrics.TotalWidth = totalWithShadow
	}

	// Bottom shadow width includes the adjusted shadow size
	metrics.BottomShadowWidth = contentWidth + metrics.AdjustedShadow

	return metrics
}

// applyShadow adds a terminal shadow effect to a rendered block with overflow detection.
// shadowSize: 1=small, 2=medium, 3=large
// maxWidth: maximum available width (optional, 0 means no overflow check)
func applyShadow(content string, shadowSize int) string {
	return applyShadowWithWidth(content, shadowSize, 0)
}

// applyShadowWithWidth adds a terminal shadow effect with awareness of max terminal width.
// shadowSize: 1=small, 2=medium, 3=large
// maxWidth: maximum available width in terminal (0 means no overflow check)
func applyShadowWithWidth(content string, shadowSize, maxWidth int) string {
	lines := strings.Split(content, "\n")

	// Trim leading and trailing empty lines so shadow wraps the visible block only
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

	// Pre-calculate shadow metrics
	contentWidth := lipgloss.Width(visible[0])
	var metrics shadowMetrics
	if maxWidth > 0 {
		metrics = calculateShadowMetrics(contentWidth, shadowSize, maxWidth) // reserve space for shadow
	} else {
		// No max width constraint, use requested shadow size
		metrics = shadowMetrics{
			ContentWidth:      contentWidth,
			ShadowWidth:       shadowSize,
			AdjustedShadow:    shadowSize,
			WillOverflow:      false,
			BottomShadowWidth: contentWidth + shadowSize,
			TotalWidth:        contentWidth + shadowSize,
		}
	}

	shadowChar := "░"
	if metrics.AdjustedShadow >= 3 {
		shadowChar = "▒"
	}

	shadowStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	result := make([]string, 0, len(lines)+metrics.AdjustedShadow)

	// Preserve leading empty lines as-is
	result = append(result, leading...)

	// Render each visible line with right-side shadow
	for _, line := range visible {
		shadow := shadowStyle.Render(strings.Repeat(shadowChar, metrics.AdjustedShadow))
		result = append(result, line+shadow)
	}

	// Add shadow at the bottom
	bottomShadow := shadowStyle.Render(strings.Repeat(shadowChar, metrics.BottomShadowWidth))
	for i := 0; i < metrics.AdjustedShadow; i++ {
		result = append(result, strings.Repeat(" ", metrics.AdjustedShadow)+bottomShadow)
	}

	// Preserve trailing empty lines as-is
	result = append(result, trailing...)

	return strings.Join(result, "\n")
}
