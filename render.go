package termstrap

import (
	"image"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/webp"

	"github.com/PuerkitoBio/goquery"
	"github.com/charmbracelet/lipgloss"
	termimage "github.com/go-scripts/termstrap/image"
	"github.com/muesli/reflow/wordwrap"
	"github.com/muesli/reflow/wrap"
	"github.com/muesli/termenv"
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
	Theme       Theme    // Visual color theme (default: ThemeBootstrap)
	Stylesheets []string // Additional CSS stylesheets
	Width       int      // Terminal width in characters
	RootPath    string   // Base path for local images

	// ImageRenderer overrides the default image renderer.
	ImageRenderer termimage.Renderer

	// MaxImageWidth caps image width in terminal columns (0 = default).
	MaxImageWidth int

	// ColorMode specifies color depth.
	ColorMode termimage.ColorMode

	// OptimizeSequences deduplicates consecutive ANSI color codes.
	OptimizeSequences *bool

	// DisableImages replaces images with [alt](src) text.
	DisableImages bool

	// Cache policy
	CachePolicy CachePolicy
	CacheTTL    time.Duration
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

	// Ensure color output is not silently stripped in non-interactive/SSH environments
	if lipgloss.ColorProfile() == termenv.Ascii {
		if m.ColorMode == termimage.ColorMode256 {
			lipgloss.SetColorProfile(termenv.ANSI256)
		} else {
			lipgloss.SetColorProfile(termenv.TrueColor)
		}
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(m.HTML))
	if err != nil {
		return "", nil, err
	}

	matcher, err := NewCSSMatcher(m.Theme, m.Stylesheets...)
	if err != nil {
		return "", nil, err
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

	renderTree := BuildRenderTree(rootSel, matcher, termWidth)
	if renderTree == nil {
		return "", nil, nil
	}

	ComputeLayout(renderTree, termWidth)

	var overlays []ImageOverlay
	output := m.renderNode(renderTree, 0, 0, &overlays, 0)

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

func (m Model) renderInline(node *RenderNode, currentX, currentY int, overlays *[]ImageOverlay) string {
	if node == nil {
		return ""
	}

	if node.Type == NodeText {
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
		sb.WriteString(m.renderInline(child, currentX, currentY, overlays))
	}
	return applyInlineStyle(node.Style, sb.String())
}

func (m Model) renderNode(node *RenderNode, currentX, currentY int, overlays *[]ImageOverlay, targetHeight int) string {
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
		return applyLipglossStyle(node.Style, node.LipglossWidth, targetHeight, node.Text)
	}

	if node.Type == NodeText {
		return node.Text
	}

	if node.Type == NodeImage {
		return m.renderImageNode(node, currentX, currentY, overlays)
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
				out := m.renderNode(child, currentX, currentY, overlays, 0)
				renderedItems[i] = out
				if out != "" {
					h := strings.Count(out, "\n") + 1
					if h > maxHeight {
						maxHeight = h
					}
				}
			}

			// Pass 2: equalise height across columns if multiple columns
			var itemOutputs []string
			if len(row.children) > 1 && maxHeight > 0 {
				for _, child := range row.children {
					// Re-render with target height
					out := m.renderNode(child, currentX, currentY, overlays, maxHeight)
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

			if len(itemOutputs) == 1 {
				rowOutputs = append(rowOutputs, itemOutputs[0])
			} else if len(itemOutputs) > 1 {
				rowOutputs = append(rowOutputs, lipgloss.JoinHorizontal(lipgloss.Top, itemOutputs...))
			}
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
				sb.WriteString(m.renderInline(inlineNode, blockX, blockY, overlays))
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
				out := m.renderNode(child, blockX, blockY, overlays, childTargetHeight)
				if out != "" {
					childOutputs = append(childOutputs, out)
					blockY += strings.Count(out, "\n") + 1
				}
			}
		}
		flushInlineRun()
		content = lipgloss.JoinVertical(lipgloss.Left, childOutputs...)
	}
	return applyLipglossStyle(node.Style, node.LipglossWidth, targetHeight, content)
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

	if renderer.Protocol() != termimage.HalfBlock && overlays != nil {
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
		shadowV := 0
		if s.Shadow > 0 {
			shadowV = s.Shadow
		}
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
	if s.Border {
		if s.Rounded {
			ls = ls.BorderStyle(lipgloss.RoundedBorder())
		} else {
			ls = ls.BorderStyle(lipgloss.NormalBorder())
		}
	} else if s.BorderTop || s.BorderBottom || s.BorderLeft || s.BorderRight {
		if s.Rounded {
			ls = ls.BorderStyle(lipgloss.RoundedBorder())
		} else {
			ls = ls.BorderStyle(lipgloss.NormalBorder())
		}
		ls = ls.BorderTop(s.BorderTop).
			BorderBottom(s.BorderBottom).
			BorderLeft(s.BorderLeft).
			BorderRight(s.BorderRight)
	}

	if s.BorderColor != "" {
		ls = ls.BorderForeground(lipgloss.Color(s.BorderColor))
	}

	// Colors
	if s.BgColor != "" {
		ls = ls.Background(lipgloss.Color(s.BgColor))
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
	}

	out := ls.Render(content)
	if s.BgColor != "" {
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
	for i, line := range lines {
		if !strings.Contains(line, bgSeq) {
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
	for i := 0; i < effectiveShadow; i++ {
		result = append(result, strings.Repeat(" ", effectiveShadow)+bottomShadow)
	}

	result = append(result, trailing...)
	return strings.Join(result, "\n")
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
	_, _ = w.Write([]byte(text))
	_ = w.Close()
	out := w.String()

	var lines []string
	for _, line := range strings.Split(out, "\n") {
		if lipgloss.Width(line) > width {
			lines = append(lines, wrap.String(line, width))
		} else {
			lines = append(lines, line)
		}
	}
	return strings.Join(lines, "\n")
}
