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
	output := m.renderNode(renderTree, 0, 0, &overlays)

	return output, overlays, nil
}

func (m Model) renderNode(node *RenderNode, currentX, currentY int, overlays *[]ImageOverlay) string {
	if node == nil {
		return ""
	}

	if node.Type == NodeText {
		text := wordwrap.String(node.Text, node.ContentWidth)
		if node.Style.TextAlign == lipgloss.Center {
			lines := strings.Split(text, "\n")
			for i, l := range lines {
				lines[i] = lipgloss.NewStyle().Width(node.ContentWidth).Align(lipgloss.Center).Render(l)
			}
			text = strings.Join(lines, "\n")
		} else if node.Style.TextAlign == lipgloss.Right {
			lines := strings.Split(text, "\n")
			for i, l := range lines {
				lines[i] = lipgloss.NewStyle().Width(node.ContentWidth).Align(lipgloss.Right).Render(l)
			}
			text = strings.Join(lines, "\n")
		}
		return text
	}

	if node.Type == NodeImage {
		return m.renderImageNode(node, currentX, currentY, overlays)
	}

	var childOutputs []string
	if node.Style.Display == DisplayFlex {
		// Calculate total span to know if we stack vertically
		totalSpan := 0
		for _, child := range node.Children {
			totalSpan += child.Style.ColSpan
		}

		if totalSpan > 12 {
			// Stack vertically
			stackX := currentX + node.Style.MarginLeft + node.Style.PaddingLeft
			stackY := currentY + node.Style.MarginTop + node.Style.PaddingTop
			for _, child := range node.Children {
				out := m.renderNode(child, stackX, stackY, overlays)
				if out != "" {
					childOutputs = append(childOutputs, out)
					stackY += strings.Count(out, "\n") + 1
				}
			}
		} else {
			// Side-by-side horizontal flex
			flexX := currentX + node.Style.MarginLeft + node.Style.PaddingLeft
			flexY := currentY + node.Style.MarginTop + node.Style.PaddingTop
			for _, child := range node.Children {
				out := m.renderNode(child, flexX, flexY, overlays)
				childOutputs = append(childOutputs, out)
				flexX += child.AllocatedWidth
			}
		}
	} else {
		// Block
		blockX := currentX + node.Style.MarginLeft + node.Style.PaddingLeft
		blockY := currentY + node.Style.MarginTop + node.Style.PaddingTop
		for _, child := range node.Children {
			out := m.renderNode(child, blockX, blockY, overlays)
			if out != "" {
				childOutputs = append(childOutputs, out)
				blockY += strings.Count(out, "\n") + 1
			}
		}
	}

	var content string
	if node.Style.Display == DisplayFlex {
		totalSpan := 0
		for _, child := range node.Children {
			totalSpan += child.Style.ColSpan
		}
		if totalSpan > 12 {
			content = lipgloss.JoinVertical(lipgloss.Left, childOutputs...)
		} else {
			content = lipgloss.JoinHorizontal(lipgloss.Top, childOutputs...)
		}
	} else {
		content = lipgloss.JoinVertical(lipgloss.Left, childOutputs...)
	}

	return applyLipglossStyle(node.Style, node.AllocatedWidth, content)
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

func applyLipglossStyle(s ComputedStyle, width int, content string) string {
	ls := lipgloss.NewStyle()

	// If explicit width or container
	if width > 0 && s.Display == DisplayBlock {
		// width constraint
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
		border := lipgloss.Border{
			Top:    boolStr(s.BorderTop, "─"),
			Bottom: boolStr(s.BorderBottom, "─"),
			Left:   boolStr(s.BorderLeft, "│"),
			Right:  boolStr(s.BorderRight, "│"),
		}
		ls = ls.BorderStyle(border)
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

	return ls.Render(content)
}

func boolStr(b bool, str string) string {
	if b {
		return str
	}
	return ""
}
