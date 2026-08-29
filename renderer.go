package termstrap

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
	termimage "github.com/go-scripts/termstrap/image"
)

// colOverlay tracks deferred images from a column for native protocol overlay.
type colOverlay struct {
	deferred []deferredImage
	colIndex int
}

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
	proto := m.imageRenderer().Protocol()
	rendered, deferred, err := m.renderRowDeferred(row, proto)
	if err != nil {
		return "", err
	}

	bp := detectBreakpoint(m.Width)
	shouldStack := shouldStackRow(row, bp)
	isMultiCol := !shouldStack && len(row.Columns) > 1
	useOverlay := isMultiCol && proto != termimage.HalfBlock

	if useOverlay && len(deferred) > 0 {
		rendered = m.appendImageOverlays(rendered, deferred)
	}

	return rendered + "\n", nil
}

// renderRowDeferred renders a grid row without emitting cursor escape sequences,
// returning clean text and all deferred images with their relative (row, col) within the row.
func (m Model) renderRowDeferred(row gridRow, proto termimage.Protocol) (string, []deferredImage, error) {
	bp := detectBreakpoint(m.Width)
	shouldStack := shouldStackRow(row, bp)
	rowStyle := resolveStyle(row.Classes)

	innerWidth := m.Width
	if hasStyling(rowStyle) {
		innerWidth -= rowStyle.MarginLeft + rowStyle.MarginRight
		innerWidth -= rowStyle.PaddingLeft + rowStyle.PaddingRight
		if rowStyle.Border || rowStyle.BorderLeft {
			innerWidth -= 1
		}
		if rowStyle.Border || rowStyle.BorderRight {
			innerWidth -= 1
		}
		if rowStyle.Shadow > 0 {
			innerWidth -= rowStyle.Shadow
		}
		if innerWidth < 10 {
			innerWidth = 10
		}
		resolveColumnWidths(&row, innerWidth, bp)
	}

	innerModel := m
	innerModel.Width = innerWidth

	type colRenderResult struct {
		innerContent       string
		colStyle           style
		styleWidth         int
		contentWidth       int
		partialBorderWidth int
		deferred           []deferredImage
	}

	colResults := make([]colRenderResult, len(row.Columns))
	maxInnerLines := 0

	isMultiCol := !shouldStack && len(row.Columns) > 1
	useOverlay := isMultiCol && proto != termimage.HalfBlock

	for i, col := range row.Columns {
		colStyle := resolveStyle(col.Classes)
		totalWidth := col.Width
		if shouldStack {
			totalWidth = m.Width
		}

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

		partialBorderWidth := styleWidth

		if colStyle.Shadow > 0 {
			styleWidth -= colStyle.Shadow
			if styleWidth < 10 {
				styleWidth = 10
			}
		}

		contentWidth := styleWidth - colStyle.PaddingLeft - colStyle.PaddingRight
		if contentWidth < 10 {
			contentWidth = 10
		}

		colModel := innerModel
		colModel.Content = col.Content
		colModel.Width = contentWidth

		var innerRendered string
		var deferred []deferredImage

		if useOverlay {
			rendered, defs, err := colModel.renderContentDeferred(col.Content, proto)
			if err != nil {
				return "", nil, err
			}
			innerRendered = rendered
			deferred = defs
		} else {
			if m.ImageRenderer != nil {
				colModel.ImageRenderer = m.ImageRenderer
			}
			rendered, err := colModel.Render()
			if err != nil {
				return "", nil, err
			}
			innerRendered = trimBlankLines(rendered)
		}

		lines := 0
		if innerRendered != "" {
			lines = strings.Count(innerRendered, "\n") + 1
		}
		if lines > maxInnerLines {
			maxInnerLines = lines
		}

		colResults[i] = colRenderResult{
			innerContent:       innerRendered,
			colStyle:           colStyle,
			styleWidth:         styleWidth,
			contentWidth:       contentWidth,
			partialBorderWidth: partialBorderWidth,
			deferred:           deferred,
		}
	}

	renderedCols := make([]string, 0, len(row.Columns))
	var rowDeferred []deferredImage
	currentXOffset := 0
	currentYOffset := 0

	for i, res := range colResults {
		rendered := res.innerContent
		if !shouldStack && len(row.Columns) > 1 && maxInnerLines > 0 {
			lines := strings.Split(rendered, "\n")
			for len(lines) < maxInnerLines {
				lines = append(lines, "")
			}
			rendered = strings.Join(lines, "\n")
		}

		if res.colStyle.TextAlign == lipgloss.Center || res.colStyle.TextAlign == lipgloss.Right {
			rendered = alignContent(rendered, res.colStyle.TextAlign)
		}

		if res.colStyle.BgColor != "" || res.colStyle.FgColor != "" {
			rendered = persistColors(rendered, res.colStyle.BgColor, res.colStyle.FgColor)
		}

		ls := applyStyle(res.colStyle, res.styleWidth)
		output := ls.Render(rendered)

		topOverhead := res.colStyle.MarginTop + res.colStyle.PaddingTop
		if res.colStyle.Border || res.colStyle.BorderTop {
			topOverhead++
		}

		leftOverhead := res.colStyle.MarginLeft + res.colStyle.PaddingLeft
		if res.colStyle.Border || res.colStyle.BorderLeft {
			leftOverhead++
		}

		for _, di := range res.deferred {
			innerColOffset := 0
			if res.colStyle.TextAlign == lipgloss.Center {
				innerColOffset = (res.contentWidth - di.width) / 2
				if innerColOffset < 0 {
					innerColOffset = 0
				}
			} else if res.colStyle.TextAlign == lipgloss.Right {
				innerColOffset = res.contentWidth - di.width
				if innerColOffset < 0 {
					innerColOffset = 0
				}
			}

			if shouldStack {
				di.row += currentYOffset + topOverhead
				di.col += leftOverhead + innerColOffset
			} else {
				di.row += topOverhead
				di.col += currentXOffset + leftOverhead + innerColOffset
			}
			rowDeferred = append(rowDeferred, di)
		}

		if !res.colStyle.Border && (res.colStyle.BorderTop || res.colStyle.BorderBottom || res.colStyle.BorderLeft || res.colStyle.BorderRight) {
			output = applyPartialBorders(output, res.partialBorderWidth, res.colStyle.BorderTop, res.colStyle.BorderBottom, res.colStyle.BorderLeft, res.colStyle.BorderRight)
		}

		if res.colStyle.Shadow > 0 {
			shadowMaxWidth := res.styleWidth
			output = applyShadowIntelligent(output, res.colStyle.Shadow, shadowMaxWidth)
		}

		renderedCols = append(renderedCols, output)
		currentXOffset += row.Columns[i].Width
		currentYOffset += strings.Count(output, "\n") + 1
	}

	var output string
	if shouldStack {
		output = lipgloss.JoinVertical(lipgloss.Left, renderedCols...)
	} else {
		output = lipgloss.JoinHorizontal(lipgloss.Top, renderedCols...)
	}

	// Apply row-level styling
	if hasStyling(rowStyle) {
		rowStyleWidth := innerWidth
		if rowStyle.BgColor != "" || rowStyle.FgColor != "" {
			output = persistColors(output, rowStyle.BgColor, rowStyle.FgColor)
		}
		rowLipgloss := applyStyle(rowStyle, rowStyleWidth)
		output = rowLipgloss.Render(output)
	}

	// Apply shadow AFTER lipgloss styling (so it's outside the border)
	if rowStyle.Shadow > 0 {
		shadowMaxWidth := innerWidth
		output = applyShadowIntelligent(output, rowStyle.Shadow, shadowMaxWidth)
	}

	return output, rowDeferred, nil
}

// renderHTMLLayoutDeferred parses an HTML layout block and renders it to ANSI text
// without inline cursor overlays, collecting all deferred images.
func (m Model) renderHTMLLayoutDeferred(htmlContent string, proto termimage.Protocol) (string, []deferredImage, error) {
	rows, err := parseGrid(htmlContent, m.Width)
	if err != nil {
		return "", nil, err
	}

	var result string
	var allDeferred []deferredImage
	currentRow := 0

	for _, row := range rows {
		rendered, defs, err := m.renderRowDeferred(row, proto)
		if err != nil {
			return "", nil, err
		}
		for _, di := range defs {
			di.row += currentRow
			allDeferred = append(allDeferred, di)
		}
		result += rendered + "\n"
		currentRow += strings.Count(rendered, "\n") + 1
	}

	return result, allDeferred, nil
}

// renderContentDeferred renders markdown or HTML layout blocks, collecting
// deferred images for later cursor overlay.
func (m Model) renderContentDeferred(content string, proto termimage.Protocol) (string, []deferredImage, error) {
	segments := extractSegments(content)
	var renderedParts []string
	var allDeferred []deferredImage
	currentRow := 0

	for _, seg := range segments {
		switch seg.Type {
		case segmentMarkdown:
			rendered, deferred, err := m.renderMarkdownDeferred(seg.Content, proto)
			if err != nil {
				return "", nil, err
			}
			for _, di := range deferred {
				di.row += currentRow
				allDeferred = append(allDeferred, di)
			}
			renderedParts = append(renderedParts, rendered)
			currentRow += strings.Count(rendered, "\n")

		case segmentHTML:
			rendered, deferred, err := m.renderHTMLLayoutDeferred(seg.Content, proto)
			if err != nil {
				return "", nil, err
			}
			for _, di := range deferred {
				di.row += currentRow
				allDeferred = append(allDeferred, di)
			}
			renderedParts = append(renderedParts, rendered)
			currentRow += strings.Count(rendered, "\n")
		}
	}

	return strings.Join(renderedParts, ""), allDeferred, nil
}

// renderColumn renders a single column with its content and styling.
func (m Model) renderColumn(col gridColumn, stacked bool, forceHalfBlock bool) (string, error) {
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

	// Render the column content using the full pipeline (extractSegments → renderHTMLLayout
	// for nested rows, renderMarkdown for plain markdown). This enables recursive grid rendering.
	colModel := m
	colModel.Content = col.Content
	colModel.Width = contentWidth

	// Force halfblock for multi-column horizontal layouts to ensure
	// borders align correctly (graphic protocols break lipgloss alignment)
	if forceHalfBlock {
		opt := true
		if m.OptimizeSequences != nil {
			opt = *m.OptimizeSequences
		}
		colModel.ImageRenderer = termimage.NewRenderer(
			termimage.WithProtocol(termimage.HalfBlock),
			termimage.WithColorMode(m.ColorMode),
			termimage.WithOptimizeSequences(opt),
		)
	} else if m.ImageRenderer != nil {
		colModel.ImageRenderer = m.ImageRenderer
	}
	rendered, err := colModel.Render()
	if err != nil {
		return "", err
	}

	// Trim leading and trailing blank lines for tighter column layout
	rendered = trimBlankLines(rendered)

	// Inject background/foreground colors into glamour output to prevent
	// ANSI resets from clearing lipgloss background styling
	if colStyle.BgColor != "" || colStyle.FgColor != "" {
		rendered = persistColors(rendered, colStyle.BgColor, colStyle.FgColor)
	}

	if colStyle.TextAlign == lipgloss.Center || colStyle.TextAlign == lipgloss.Right {
		rendered = alignContent(rendered, colStyle.TextAlign)
	}

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

// renderColumnDeferred renders a column with deferred image overlay support.
// Images are replaced with blank space for layout; actual rendering happens
// via cursor-based overlay after the row is composed.
func (m Model) renderColumnDeferred(col gridColumn, stacked bool, colIndex int) (string, []deferredImage, error) {
	colStyle := resolveStyle(col.Classes)

	totalWidth := col.Width
	if stacked {
		totalWidth = m.Width
	}

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

	partialBorderWidth := styleWidth

	if colStyle.Shadow > 0 {
		styleWidth -= colStyle.Shadow
		if styleWidth < 10 {
			styleWidth = 10
		}
	}

	contentWidth := styleWidth - colStyle.PaddingLeft - colStyle.PaddingRight
	if contentWidth < 10 {
		contentWidth = 10
	}

	colModel := m
	colModel.Content = col.Content
	colModel.Width = contentWidth

	proto := m.imageRenderer().Protocol()
	rendered, deferred, err := colModel.renderContentDeferred(col.Content, proto)
	if err != nil {
		return "", nil, err
	}

	if colStyle.BgColor != "" || colStyle.FgColor != "" {
		rendered = persistColors(rendered, colStyle.BgColor, colStyle.FgColor)
	}

	if colStyle.TextAlign == lipgloss.Center || colStyle.TextAlign == lipgloss.Right {
		rendered = alignContent(rendered, colStyle.TextAlign)
	}

	ls := applyStyle(colStyle, styleWidth)
	output := ls.Render(rendered)

	// Calculate the vertical overhead that lipgloss added above the content.
	// This adjusts each deferred image's row to its absolute position in the column output.
	topOverhead := colStyle.MarginTop + colStyle.PaddingTop
	if colStyle.Border || colStyle.BorderTop {
		topOverhead++
	}
	for i := range deferred {
		deferred[i].row += topOverhead
	}

	if !colStyle.Border && (colStyle.BorderTop || colStyle.BorderBottom || colStyle.BorderLeft || colStyle.BorderRight) {
		output = applyPartialBorders(output, partialBorderWidth, colStyle.BorderTop, colStyle.BorderBottom, colStyle.BorderLeft, colStyle.BorderRight)
	}

	if colStyle.Shadow > 0 {
		shadowMaxWidth := styleWidth
		output = applyShadowIntelligent(output, colStyle.Shadow, shadowMaxWidth)
	}

	return output, deferred, nil
}

var trailingANSIWhitespace = regexp.MustCompile(`(?:\s|\x1b\[[0-9;]*m)+$`)

// trimLineForAlign strips leading/trailing whitespace and trailing ANSI fill sequences
// from a line so Lipgloss can center/right-align text properly.
// It ensures that ANSI reset codes (\x1b[0m) are preserved so colors never bleed.
func trimLineForAlign(l string) string {
	plain := stripANSI(l)
	if plain == "" {
		if strings.Contains(l, "\x1b") {
			return ""
		}
		return l
	}

	// Check if line contains color escapes or ended with reset
	hadReset := strings.HasSuffix(l, "\x1b[0m")
	hasColor := strings.Contains(l, "\x1b[38;") || strings.Contains(l, "\x1b[48;")

	// Strip trailing whitespace and reset/color codes
	l = trailingANSIWhitespace.ReplaceAllString(l, "")

	// Strip leading whitespace while preserving formatting ANSI codes
	var buf strings.Builder
	inLeading := true
	for i := 0; i < len(l); i++ {
		if inLeading && (l[i] == ' ' || l[i] == '\t') {
			continue
		}
		if inLeading && l[i] == '\x1b' {
			end := i + 1
			for end < len(l) && l[end] != 'm' && l[end] != 'G' && l[end] != 'K' {
				end++
			}
			if end < len(l) {
				end++
			}
			seq := l[i:end]
			if seq != "\x1b[0m" {
				buf.WriteString(seq)
			}
			i = end - 1
			continue
		}
		inLeading = false
		buf.WriteByte(l[i])
	}

	result := buf.String()
	if (hadReset || hasColor) && !strings.HasSuffix(result, "\x1b[0m") {
		result += "\x1b[0m"
	}

	return result
}

// alignContent trims leading and trailing spaces from non-blank lines for Center/Right alignment
// so Lipgloss can center/right-align properly without Glamour's default margins interfering.
func alignContent(content string, align lipgloss.Position) string {
	if align != lipgloss.Center && align != lipgloss.Right {
		return content
	}
	lines := strings.Split(content, "\n")
	for i, l := range lines {
		lines[i] = trimLineForAlign(l)
	}
	return strings.Join(lines, "\n")
}

// appendImageOverlays appends cursor-based native protocol image renders
// after the composed row output. Each image is painted at its calculated
// position using ANSI cursor movement (CUU/CHA) and the native protocol.
func (m Model) appendImageOverlays(output string, deferred []deferredImage) string {
	renderer := m.imageRenderer()

	totalLines := strings.Count(output, "\n")

	var buf strings.Builder
	buf.WriteString(output)

	for _, di := range deferred {
		absCol := di.col
		absRow := di.row

		// Move cursor up from the end of the output
		linesUp := totalLines - absRow
		if linesUp > 0 {
			fmt.Fprintf(&buf, "\x1b[%dA", linesUp)
		}

		// Move cursor to the image column (1-based CHA)
		fmt.Fprintf(&buf, "\x1b[%dG", absCol+1)

		// Render image with height constraint to prevent overflow.
		// ConstrainedRenderer tells the terminal to fit the image within
		// the exact width×height cell box, preventing border overflow.
		var rendered string
		var err error
		if cr, ok := renderer.(termimage.ConstrainedRenderer); ok {
			rendered, err = cr.RenderConstrained(di.img, di.width, di.height)
		} else {
			rendered, err = renderer.Render(di.img, di.width)
		}
		if err == nil {
			buf.WriteString(rendered)
		}

		// Move cursor back down to the end of the output.
		// After native image render, cursor is at (absRow + height).
		linesDown := totalLines - absRow - di.height
		if linesDown > 0 {
			fmt.Fprintf(&buf, "\x1b[%dB", linesDown)
		}

		// Reset to column 1
		buf.WriteString("\x1b[1G")
	}

	return buf.String()
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
	// Bottom shadow width = maxContentWidth (not + effectiveShadow) because
	// the leading spaces (effectiveShadow) already bring total to maxContentWidth + effectiveShadow
	bottomShadow := shadowStyle.Render(strings.Repeat(shadowChar, maxContentWidth))
	for i := 0; i < shadowSize; i++ {
		result = append(result, strings.Repeat(" ", effectiveShadow)+bottomShadow)
	}

	result = append(result, trailing...)
	return strings.Join(result, "\n")
}

// shadowMetrics holds the calculations for shadow rendering
type shadowMetrics struct {
	ContentWidth      int  // width of content (longest line)
	ShadowWidth       int  // requested shadow size
	AdjustedShadow    int  // actual shadow size to render (reduced if overflow)
	WillOverflow      bool // true if requested shadow exceeds maxWidth
	BottomShadowWidth int  // width of bottom shadow line
	TotalWidth        int  // final line width with shadow
}

// calculateShadowMetrics computes how a shadow should be rendered without overflowing maxWidth
func calculateShadowMetrics(contentWidth, shadowSize, maxWidth int) shadowMetrics {
	var metrics shadowMetrics
	metrics.ContentWidth = contentWidth
	metrics.ShadowWidth = shadowSize

	// If no maxWidth constraint, use full shadow size
	if maxWidth <= 0 {
		metrics.WillOverflow = false
		metrics.AdjustedShadow = shadowSize
		metrics.TotalWidth = contentWidth + shadowSize
		metrics.BottomShadowWidth = contentWidth
		return metrics
	}

	// Calculate total width if shadow is rendered
	totalWithShadow := contentWidth + shadowSize

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

	// Bottom shadow width matches content width; the leading offset spaces
	// bring the total line width to contentWidth + AdjustedShadow
	metrics.BottomShadowWidth = contentWidth

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
	// Bottom shadow width = content width; offset spaces bring total to content + shadow
	bottomShadow := shadowStyle.Render(strings.Repeat(shadowChar, metrics.BottomShadowWidth))
	for i := 0; i < metrics.AdjustedShadow; i++ {
		result = append(result, strings.Repeat(" ", metrics.AdjustedShadow)+bottomShadow)
	}

	// Preserve trailing empty lines as-is
	result = append(result, trailing...)

	return strings.Join(result, "\n")
}

// deferredToOverlay converts a deferredImage to an ImageOverlay by rendering
// the image with the native protocol renderer.
func (m Model) deferredToOverlay(di deferredImage, rowOffset, colOffset int, renderer termimage.Renderer) *ImageOverlay {
	var rendered string
	var err error
	if cr, ok := renderer.(termimage.ConstrainedRenderer); ok {
		rendered, err = cr.RenderConstrained(di.img, di.width, di.height)
	} else {
		rendered, err = renderer.Render(di.img, di.width)
	}
	if err != nil || rendered == "" {
		return nil
	}
	return &ImageOverlay{
		Row:    rowOffset + di.row,
		Col:    colOffset,
		Width:  di.width,
		Height: di.height,
		Escape: rendered,
	}
}

// renderHTMLLayoutWithOverlays parses an HTML layout block and renders it,
// collecting image overlays instead of appending cursor-based sequences inline.
func (m Model) renderHTMLLayoutWithOverlays(htmlContent string, baseRow int) (string, []ImageOverlay, error) {
	rows, err := parseGrid(htmlContent, m.Width)
	if err != nil {
		return "", nil, err
	}

	renderer := m.imageRenderer()
	var result string
	var allOverlays []ImageOverlay
	currentRow := baseRow

	for _, row := range rows {
		rendered, overlays, err := m.renderRowWithOverlays(row, currentRow, renderer)
		if err != nil {
			return "", nil, err
		}
		result += rendered
		allOverlays = append(allOverlays, overlays...)
		currentRow += strings.Count(rendered, "\n")
	}

	return result, allOverlays, nil
}

// renderRowWithOverlays renders a grid row and collects image overlays
// instead of appending cursor-based overlay sequences inline.
func (m Model) renderRowWithOverlays(row gridRow, baseRow int, renderer termimage.Renderer) (string, []ImageOverlay, error) {
	proto := renderer.Protocol()
	rendered, defs, err := m.renderRowDeferred(row, proto)
	if err != nil {
		return "", nil, err
	}

	var overlays []ImageOverlay
	for _, di := range defs {
		overlay := m.deferredToOverlay(di, baseRow, di.col, renderer)
		if overlay != nil {
			overlays = append(overlays, *overlay)
		}
	}

	return rendered, overlays, nil
}
