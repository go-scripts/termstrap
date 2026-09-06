package termstrap

import (
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/andybalholm/cascadia"
	"github.com/aymerick/douceur/css"
	"github.com/aymerick/douceur/parser"
	"github.com/charmbracelet/lipgloss"
)

// DisplayType specifies the display behavior of a node.
type DisplayType string

const (
	DisplayBlock  DisplayType = "block"
	DisplayInline DisplayType = "inline"
	DisplayFlex   DisplayType = "flex"
	DisplayNone   DisplayType = "none"
)

// Breakpoint represents a screen size category.
type Breakpoint int

const (
	bpXS Breakpoint = iota // < 60
	bpSM                   // >= 60
	bpMD                   // >= 80
	bpLG                   // >= 120
	bpXL                   // >= 160
)

func detectBreakpoint(width int) Breakpoint {
	switch {
	case width >= 160:
		return bpXL
	case width >= 120:
		return bpLG
	case width >= 80:
		return bpMD
	case width >= 60:
		return bpSM
	default:
		return bpXS
	}
}

// ComputedStyle holds the resolved visual properties for a DOM node.
type ComputedStyle struct {
	Display DisplayType

	// Margins (terminal cells/lines)
	MarginTop    int
	MarginBottom int
	MarginLeft   int
	MarginRight  int

	// Paddings (terminal cells/lines)
	PaddingTop    int
	PaddingBottom int
	PaddingLeft   int
	PaddingRight  int

	// Borders
	Border       bool
	BorderTop    bool
	BorderBottom bool
	BorderLeft   bool
	BorderRight  bool
	BorderColor  string
	Rounded      bool

	// Shadow (0 = none, 1 = sm, 2 = md, 3 = lg)
	Shadow int

	// Colors
	FgColor    string
	FgColorSet bool
	BgColor    string

	// Typography
	Bold         bool
	BoldSet      bool
	Italic       bool
	ItalicSet    bool
	Underline    bool
	UnderlineSet bool
	TextAlign    lipgloss.Position
	TextAlignSet bool
	// Grid / Sizing
	ColSpan int  // 1-12 for flex items
	ColAuto bool // true for .col without explicit span number
	RowCols int  // 1-12 items per row from .row-cols-[bp-]N
	Width   int  // explicit width if specified
}

// CSSRule represents a compiled selector and declaration block.
type CSSRule struct {
	Selector string
	Compiled cascadia.Selector
	Decls    []*css.Declaration
}

// CSSMatcher compiles stylesheets into fast matchable rules.
type CSSMatcher struct {
	Rules []CSSRule
}

// NewCSSMatcher parses one or more CSS stylesheets into a CSSMatcher using the specified theme.
func NewCSSMatcher(theme Theme, stylesheets ...string) (*CSSMatcher, error) {
	matcher := &CSSMatcher{Rules: []CSSRule{}}

	var allCSS []string

	// 1. Tag defaults
	if tagsCSS, err := cssFS.ReadFile("css/tags.css"); err == nil {
		allCSS = append(allCSS, string(tagsCSS))
	}

	// 2. Framework Bootstrap layout
	if fwCSS, err := cssFS.ReadFile("css/framework-bootstrap.css"); err == nil {
		allCSS = append(allCSS, string(fwCSS))
	}

	// 3. Theme colors (fallback to ThemeBootstrap if empty or invalid)
	themeFile := "css/theme-" + string(theme) + ".css"
	themeCSS, err := cssFS.ReadFile(themeFile)
	if err != nil {
		themeCSS, _ = cssFS.ReadFile("css/theme-bootstrap.css")
	}
	if len(themeCSS) > 0 {
		allCSS = append(allCSS, string(themeCSS))
	}

	// 4. Custom stylesheets
	allCSS = append(allCSS, stylesheets...)
	for _, sheetContent := range allCSS {
		if strings.TrimSpace(sheetContent) == "" {
			continue
		}
		sheet, err := parser.Parse(sheetContent)
		if err != nil {
			continue
		}

		for _, rule := range sheet.Rules {
			for _, sel := range rule.Selectors {
				compiled, err := cascadia.Compile(sel)
				if err != nil {
					continue
				}
				matcher.Rules = append(matcher.Rules, CSSRule{
					Selector: sel,
					Compiled: compiled,
					Decls:    rule.Declarations,
				})
			}
		}
	}

	return matcher, nil
}

// ComputeStyleForSelection resolves styling for a specific DOM selection given a terminal width.
func (m *CSSMatcher) ComputeStyleForSelection(sel *goquery.Selection, termWidth int, parentStyle *ComputedStyle) ComputedStyle {
	s := ComputedStyle{
		Display:   DisplayBlock,
		TextAlign: lipgloss.Left,
	}
	node := sel.Get(0)
	if node == nil {
		return s
	}

	// Match tag defaults
	tagName := goquery.NodeName(sel)
	switch tagName {
	case "span", "a", "strong", "b", "em", "i", "code", "img":
		s.Display = DisplayInline
	case "th", "td":
		s.Display = DisplayBlock
	}

	// Apply matched rules in document/cascade order
	for _, rule := range m.Rules {
		if rule.Compiled.Match(node) {
			for _, decl := range rule.Decls {
				applyDeclaration(&s, decl)
			}
		}
	}

	// Apply inline style attribute if present
	if inlineStyle, ok := sel.Attr("style"); ok && inlineStyle != "" {
		decls, err := parser.ParseDeclarations(inlineStyle)
		if err == nil {
			for _, decl := range decls {
				applyDeclaration(&s, decl)
			}
		}
	}

	// Extract responsive column spans and row-cols based on current breakpoint
	bp := detectBreakpoint(termWidth)
	classAttr, _ := sel.Attr("class")
	classes := strings.Fields(classAttr)
	span, isAuto := resolveColSpanResponsive(classes, bp)
	if span > 0 {
		s.ColSpan = span
	}
	s.ColAuto = isAuto
	s.RowCols = resolveRowColsResponsive(classes, bp)

	// Inherit missing properties from parentStyle if present
	if parentStyle != nil {
		if !s.FgColorSet {
			s.FgColor = parentStyle.FgColor
			s.FgColorSet = parentStyle.FgColorSet
		}
		if !s.TextAlignSet {
			s.TextAlign = parentStyle.TextAlign
			s.TextAlignSet = parentStyle.TextAlignSet
		}
		if !s.BoldSet {
			s.Bold = parentStyle.Bold
			s.BoldSet = parentStyle.BoldSet
		}
		if !s.ItalicSet {
			s.Italic = parentStyle.Italic
			s.ItalicSet = parentStyle.ItalicSet
		}
		if !s.UnderlineSet {
			s.Underline = parentStyle.Underline
			s.UnderlineSet = parentStyle.UnderlineSet
		}
	}

	return s
}

// resolveColSpanResponsive extracts the appropriate column span (1-12) for the current breakpoint.
func resolveColSpanResponsive(classes []string, currentBP Breakpoint) (int, bool) {
	type colSpec struct {
		bp     Breakpoint
		span   int
		isAuto bool
	}

	var specs []colSpec
	for _, cls := range classes {
		if cls == "col" {
			specs = append(specs, colSpec{bp: bpXS, span: 0, isAuto: true})
			continue
		}
		if !strings.HasPrefix(cls, "col-") {
			continue
		}
		parts := strings.Split(cls, "-")
		if len(parts) == 2 { // col-6, col-auto
			if parts[1] == "auto" {
				specs = append(specs, colSpec{bp: bpXS, span: 0, isAuto: true})
			} else if span, err := strconv.Atoi(parts[1]); err == nil && span >= 1 && span <= 12 {
				specs = append(specs, colSpec{bp: bpXS, span: span})
			}
		} else if len(parts) == 3 { // col-md-6, col-md-auto
			bpPrefix := parts[1]
			var bp Breakpoint
			switch bpPrefix {
			case "sm":
				bp = bpSM
			case "md":
				bp = bpMD
			case "lg":
				bp = bpLG
			case "xl":
				bp = bpXL
			default:
				continue
			}
			if parts[2] == "auto" {
				specs = append(specs, colSpec{bp: bp, span: 0, isAuto: true})
			} else if span, err := strconv.Atoi(parts[2]); err == nil && span >= 1 && span <= 12 {
				specs = append(specs, colSpec{bp: bp, span: span})
			}
		}
	}

	// Find the best match: highest breakpoint <= currentBP
	bestSpan := 0
	bestAuto := false
	bestBP := Breakpoint(-1)
	for _, spec := range specs {
		if spec.bp <= currentBP && spec.bp > bestBP {
			bestBP = spec.bp
			bestSpan = spec.span
			bestAuto = spec.isAuto
		}
	}

	// If classes specify column breakpoints but none match the current breakpoint (e.g. col-md-6 on xs),
	// the column stacks (span = 12).
	if len(specs) > 0 && bestSpan == 0 && !bestAuto {
		return 12, false
	}

	return bestSpan, bestAuto
}

// resolveRowColsResponsive extracts the number of columns per row (from row-cols-[bp-]N)
func resolveRowColsResponsive(classes []string, currentBP Breakpoint) int {
	type rowColSpec struct {
		bp    Breakpoint
		count int
	}
	var specs []rowColSpec
	for _, cls := range classes {
		if !strings.HasPrefix(cls, "row-cols-") {
			continue
		}
		parts := strings.Split(cls, "-")
		if len(parts) == 3 { // row-cols-2
			if n, err := strconv.Atoi(parts[2]); err == nil && n >= 1 && n <= 12 {
				specs = append(specs, rowColSpec{bp: bpXS, count: n})
			}
		} else if len(parts) == 4 { // row-cols-md-2
			bpPrefix := parts[2]
			var bp Breakpoint
			switch bpPrefix {
			case "sm":
				bp = bpSM
			case "md":
				bp = bpMD
			case "lg":
				bp = bpLG
			case "xl":
				bp = bpXL
			default:
				continue
			}
			if n, err := strconv.Atoi(parts[3]); err == nil && n >= 1 && n <= 12 {
				specs = append(specs, rowColSpec{bp: bp, count: n})
			}
		}
	}

	bestCount := 0
	bestBP := Breakpoint(-1)
	for _, spec := range specs {
		if spec.bp <= currentBP && spec.bp > bestBP {
			bestBP = spec.bp
			bestCount = spec.count
		}
	}
	return bestCount
}

func applyDeclaration(s *ComputedStyle, decl *css.Declaration) {
	prop := strings.TrimSpace(strings.ToLower(decl.Property))
	val := strings.TrimSpace(strings.ToLower(decl.Value))

	switch prop {
	case "display":
		switch val {
		case "block":
			s.Display = DisplayBlock
		case "inline":
			s.Display = DisplayInline
		case "flex":
			s.Display = DisplayFlex
		case "none":
			s.Display = DisplayNone
		}

	case "margin":
		s.MarginTop, s.MarginRight, s.MarginBottom, s.MarginLeft = parseBox4D(val)
	case "margin-top":
		s.MarginTop = parseInt(val)
	case "margin-bottom":
		s.MarginBottom = parseInt(val)
	case "margin-left":
		s.MarginLeft = parseInt(val)
	case "margin-right":
		s.MarginRight = parseInt(val)

	case "padding":
		s.PaddingTop, s.PaddingRight, s.PaddingBottom, s.PaddingLeft = parseBox4D(val)
	case "padding-top":
		s.PaddingTop = parseInt(val)
	case "padding-bottom":
		s.PaddingBottom = parseInt(val)
	case "padding-left":
		s.PaddingLeft = parseInt(val)
	case "padding-right":
		s.PaddingRight = parseInt(val)

	case "border":
		if val == "true" || val == "1" || strings.Contains(val, "solid") || strings.Contains(val, "1px") {
			s.Border = true
		}
	case "border-top":
		s.BorderTop = val == "true" || val == "1" || strings.Contains(val, "solid")
	case "border-bottom":
		s.BorderBottom = val == "true" || val == "1" || strings.Contains(val, "solid")
	case "border-left":
		s.BorderLeft = val == "true" || val == "1" || strings.Contains(val, "solid")
	case "border-right":
		s.BorderRight = val == "true" || val == "1" || strings.Contains(val, "solid")
	case "border-radius":
		s.Rounded = val == "true" || val == "1" || val == "rounded"

	case "color":
		s.FgColor = decl.Value
		s.FgColorSet = true
	case "background-color", "background":
		s.BgColor = decl.Value
	case "border-color":
		s.BorderColor = decl.Value

	case "font-weight":
		s.BoldSet = true
		if val == "bold" || val == "700" || val == "800" || val == "900" {
			s.Bold = true
		} else if val == "normal" || val == "400" {
			s.Bold = false
		}
	case "font-style":
		s.ItalicSet = true
		if val == "italic" || val == "oblique" {
			s.Italic = true
		} else if val == "normal" {
			s.Italic = false
		}
	case "text-decoration", "text-decoration-line":
		s.UnderlineSet = true
		if strings.Contains(val, "underline") {
			s.Underline = true
		} else if strings.Contains(val, "none") {
			s.Underline = false
		}
	case "text-align":
		s.TextAlignSet = true
		switch val {
		case "center":
			s.TextAlign = lipgloss.Center
		case "right", "end":
			s.TextAlign = lipgloss.Right
		case "left", "start":
			s.TextAlign = lipgloss.Left
		}
	case "box-shadow":
		s.Shadow = parseInt(val)
	}
}

func parseInt(val string) int {
	val = strings.TrimSpace(val)
	val = strings.TrimSuffix(val, "px")
	val = strings.TrimSuffix(val, "rem")
	val = strings.TrimSuffix(val, "em")
	var n int
	for _, ch := range val {
		if ch >= '0' && ch <= '9' {
			n = n*10 + int(ch-'0')
		} else {
			break
		}
	}
	return n
}

func parseBox4D(val string) (int, int, int, int) {
	fields := strings.Fields(val)
	if len(fields) == 0 {
		return 0, 0, 0, 0
	}
	if len(fields) == 1 {
		n := parseInt(fields[0])
		return n, n, n, n
	}
	if len(fields) == 2 {
		v, h := parseInt(fields[0]), parseInt(fields[1])
		return v, h, v, h
	}
	if len(fields) == 3 {
		top, h, bottom := parseInt(fields[0]), parseInt(fields[1]), parseInt(fields[2])
		return top, h, bottom, h
	}
	return parseInt(fields[0]), parseInt(fields[1]), parseInt(fields[2]), parseInt(fields[3])
}

func parseBox2D(val string) (int, int) {
	top, right, _, _ := parseBox4D(val)
	return top, right
}
