package termstrap

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// style holds resolved terminal styling properties from Bootstrap classes.
type style struct {
	PaddingTop    int
	PaddingBottom int
	PaddingLeft   int
	PaddingRight  int
	MarginTop     int
	MarginBottom  int
	MarginLeft    int
	MarginRight   int
	TextAlign     lipgloss.Position
	Border        bool
	BorderTop     bool
	BorderBottom  bool
	BorderLeft    bool
	BorderRight   bool
	Rounded       bool
	BgColor       string
	FgColor       string
	Bold          bool
	Shadow        int // 0=none, 1=sm, 2=normal, 3=lg
}

// resolveStyle parses Bootstrap CSS classes and returns resolved terminal styles.
func resolveStyle(classes []string) style {
	s := style{
		TextAlign: lipgloss.Left,
	}

	for _, class := range classes {
		// Padding
		switch {
		case class == "p-0":
			s.PaddingTop, s.PaddingBottom, s.PaddingLeft, s.PaddingRight = 0, 0, 0, 0
		case class == "p-1":
			s.PaddingTop, s.PaddingBottom, s.PaddingLeft, s.PaddingRight = 1, 1, 1, 1
		case class == "p-2":
			s.PaddingTop, s.PaddingBottom, s.PaddingLeft, s.PaddingRight = 1, 1, 2, 2
		case class == "p-3":
			s.PaddingTop, s.PaddingBottom, s.PaddingLeft, s.PaddingRight = 2, 2, 3, 3
		case class == "p-4":
			s.PaddingTop, s.PaddingBottom, s.PaddingLeft, s.PaddingRight = 2, 2, 4, 4
		case class == "p-5":
			s.PaddingTop, s.PaddingBottom, s.PaddingLeft, s.PaddingRight = 3, 3, 5, 5
		case class == "px-1":
			s.PaddingLeft, s.PaddingRight = 1, 1
		case class == "px-2":
			s.PaddingLeft, s.PaddingRight = 2, 2
		case class == "px-3":
			s.PaddingLeft, s.PaddingRight = 3, 3
		case class == "py-1":
			s.PaddingTop, s.PaddingBottom = 1, 1
		case class == "py-2":
			s.PaddingTop, s.PaddingBottom = 2, 2
		case class == "py-3":
			s.PaddingTop, s.PaddingBottom = 3, 3
		case class == "pt-1":
			s.PaddingTop = 1
		case class == "pt-2":
			s.PaddingTop = 2
		case class == "pb-1":
			s.PaddingBottom = 1
		case class == "pb-2":
			s.PaddingBottom = 2
		case class == "ps-1":
			s.PaddingLeft = 1
		case class == "ps-2":
			s.PaddingLeft = 2
		case class == "pe-1":
			s.PaddingRight = 1
		case class == "pe-2":
			s.PaddingRight = 2
		}

		// Margin
		switch {
		case class == "m-0":
			s.MarginTop, s.MarginBottom, s.MarginLeft, s.MarginRight = 0, 0, 0, 0
		case class == "m-1":
			s.MarginTop, s.MarginBottom, s.MarginLeft, s.MarginRight = 1, 1, 1, 1
		case class == "m-2":
			s.MarginTop, s.MarginBottom, s.MarginLeft, s.MarginRight = 1, 1, 2, 2
		case class == "m-3":
			s.MarginTop, s.MarginBottom, s.MarginLeft, s.MarginRight = 2, 2, 3, 3
		case class == "mx-1":
			s.MarginLeft, s.MarginRight = 1, 1
		case class == "mx-2":
			s.MarginLeft, s.MarginRight = 2, 2
		case class == "my-1":
			s.MarginTop, s.MarginBottom = 1, 1
		case class == "my-2":
			s.MarginTop, s.MarginBottom = 2, 2
		case class == "mt-1":
			s.MarginTop = 1
		case class == "mt-2":
			s.MarginTop = 2
		case class == "mb-1":
			s.MarginBottom = 1
		case class == "mb-2":
			s.MarginBottom = 2
		}

		// Text alignment
		switch class {
		case "text-center":
			s.TextAlign = lipgloss.Center
		case "text-end", "text-right":
			s.TextAlign = lipgloss.Right
		case "text-start", "text-left":
			s.TextAlign = lipgloss.Left
		}

		// Bold
		if class == "fw-bold" || class == "text-bold" {
			s.Bold = true
		}

		// Borders
		switch class {
		case "border":
			s.Border = true
		case "border-top":
			s.BorderTop = true
		case "border-bottom":
			s.BorderBottom = true
		case "border-start", "border-left":
			s.BorderLeft = true
		case "border-end", "border-right":
			s.BorderRight = true
		case "rounded":
			s.Rounded = true
		}

		// Shadows
		switch class {
		case "shadow-sm":
			s.Shadow = 1
		case "shadow":
			s.Shadow = 2
		case "shadow-lg":
			s.Shadow = 3
		case "shadow-none":
			s.Shadow = 0
		}

		// Background colors (mapped to ANSI 256 colors)
		if strings.HasPrefix(class, "bg-") {
			s.BgColor = resolveColor(strings.TrimPrefix(class, "bg-"))
		}

		// Text colors
		if strings.HasPrefix(class, "text-") && !strings.HasPrefix(class, "text-center") &&
			!strings.HasPrefix(class, "text-end") && !strings.HasPrefix(class, "text-start") &&
			!strings.HasPrefix(class, "text-left") && !strings.HasPrefix(class, "text-right") &&
			!strings.HasPrefix(class, "text-bold") {
			s.FgColor = resolveColor(strings.TrimPrefix(class, "text-"))
		}
	}

	return s
}

// resolveColor maps Bootstrap color names to ANSI 256 color codes.
func resolveColor(name string) string {
	colors := map[string]string{
		"primary":   "#0d6efd",
		"secondary": "#6c757d",
		"success":   "#198754",
		"danger":    "#dc3545",
		"warning":   "#ffc107",
		"info":      "#0dcaf0",
		"light":     "#f8f9fa",
		"dark":      "#212529",
		"white":     "#ffffff",
		"black":     "#000000",
		"muted":     "#6c757d",
		"body":      "#212529",
	}

	if color, ok := colors[name]; ok {
		return color
	}
	return ""
}

// applyStyle creates a lipgloss.Style from the resolved style properties.
// Note: Partial borders (border-top, border-left, etc.) are NOT applied here
// and must be rendered separately to avoid conflicts with horizontal column joining.
func applyStyle(s style, width int) lipgloss.Style {
	ls := lipgloss.NewStyle().
		Width(width).
		Padding(s.PaddingTop, s.PaddingRight, s.PaddingBottom, s.PaddingLeft).
		MarginTop(s.MarginTop).
		MarginBottom(s.MarginBottom).
		MarginLeft(s.MarginLeft).
		MarginRight(s.MarginRight).
		Align(s.TextAlign)

	// Apply full border only (not partial variants - those are rendered manually)
	if s.Border {
		if s.Rounded {
			ls = ls.Border(lipgloss.RoundedBorder())
		} else {
			ls = ls.Border(lipgloss.NormalBorder())
		}
	}

	// Colors
	if s.BgColor != "" {
		ls = ls.Background(lipgloss.Color(s.BgColor))
	}
	if s.FgColor != "" {
		ls = ls.Foreground(lipgloss.Color(s.FgColor))
	}

	// Bold
	if s.Bold {
		ls = ls.Bold(true)
	}

	return ls
}

// applyPartialBorders manually renders partial borders (border-top, border-left, etc.)
// around already-rendered content by adding border lines and characters.
func applyPartialBorders(content string, width int, top, bottom, left, right bool) string {
	if !top && !bottom && !left && !right {
		return content
	}

	lines := strings.Split(content, "\n")
	result := []string{}

	// Top border
	if top {
		topLine := buildBorderLine(width, left, right, "┌", "┐", "─")
		result = append(result, topLine)
	}

	// Content lines with side borders
	for _, line := range lines {
		// Measure visible width (ignore ANSI codes)
		visibleWidth := lipgloss.Width(line)
		
		// Padding needed to reach target width
		padCount := width
		if left {
			padCount--
		}
		if right {
			padCount--
		}
		padCount -= visibleWidth

		if padCount < 0 {
			padCount = 0
		}

		pad := strings.Repeat(" ", padCount)

		// Build line with side borders
		switch {
		case left && right:
			result = append(result, "│"+line+pad+"│")
		case left:
			result = append(result, "│"+line+pad)
		case right:
			result = append(result, line+pad+"│")
		default:
			result = append(result, line+pad)
		}
	}

	// Bottom border
	if bottom {
		bottomLine := buildBorderLine(width, left, right, "└", "┘", "─")
		result = append(result, bottomLine)
	}

	return strings.Join(result, "\n")
}

// buildBorderLine constructs a single border line with optional corners and sides.
func buildBorderLine(width int, hasLeft, hasRight bool, cornerLeft, cornerRight, fillChar string) string {
	if width < 2 {
		return strings.Repeat(fillChar, width)
	}

	left := fillChar
	right := fillChar
	if hasLeft {
		left = cornerLeft
	}
	if hasRight {
		right = cornerRight
	}

	fillWidth := width - 2
	if fillWidth < 0 {
		fillWidth = 0
	}

	return left + strings.Repeat(fillChar, fillWidth) + right
}
