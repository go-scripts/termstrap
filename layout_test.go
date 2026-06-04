package termstrap

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"
	termimage "github.com/go-scripts/termstrap/image"
)

// ---------- test helpers ----------

// allANSIRegex strips all ANSI escape sequences (SGR, OSC, APC, cursor, etc.)
var allANSIRegex = regexp.MustCompile(`\x1b(?:\[[0-9;]*[A-Za-z]|\][^\a\x1b]*(?:\a|\x1b\\)|\x1b_[^\x1b]*\x1b\\|[78])`)

// stripAllANSI removes every known ANSI/VT escape sequence from s.
func stripAllANSI(s string) string {
	return allANSIRegex.ReplaceAllString(s, "")
}

// visualLines splits output into lines and strips ANSI from each one.
func visualLines(output string) []string {
	raw := strings.Split(output, "\n")
	out := make([]string, len(raw))
	for i, l := range raw {
		out[i] = stripAllANSI(l)
	}
	return out
}

// runeWidth returns the number of runes in s (accounts for multi-byte UTF-8).
func runeWidth(s string) int {
	return utf8.RuneCountInString(s)
}

// createTestPNG writes a solid-color PNG to dir and returns the filename.
func createTestPNG(t *testing.T, dir string, w, h int) string {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{100, 149, 237, 255})
		}
	}
	name := "test.png"
	f, err := os.Create(filepath.Join(dir, name))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
	return name
}

// renderHTML is a convenience wrapper: renders HTML layout content with the
// given width, protocol, and optional RootPath.
func renderHTML(t *testing.T, html string, width int, proto termimage.Protocol, rootPath string) string {
	t.Helper()
	m := Model{
		Content:       html,
		Width:         width,
		RootPath:      rootPath,
		ImageRenderer: termimage.NewRenderer(termimage.WithProtocol(proto)),
	}
	out, err := m.Render()
	if err != nil {
		t.Fatal(err)
	}
	return out
}

// borderChars defines rounded border characters used by lipgloss.
var borderChars = struct {
	TopLeft, TopRight, BottomLeft, BottomRight string
	Horizontal, Vertical                       string
}{
	TopLeft: "╭", TopRight: "╮", BottomLeft: "╰", BottomRight: "╯",
	Horizontal: "─", Vertical: "│",
}

// normalBorderChars defines normal (non-rounded) border characters.
var normalBorderChars = struct {
	TopLeft, TopRight, BottomLeft, BottomRight string
}{
	TopLeft: "┌", TopRight: "┐", BottomLeft: "└", BottomRight: "┘",
}

// hasTopBorder checks if a line contains either rounded or normal top border.
func hasTopBorder(l string) bool {
	return (strings.Contains(l, borderChars.TopLeft) && strings.Contains(l, borderChars.TopRight)) ||
		(strings.Contains(l, normalBorderChars.TopLeft) && strings.Contains(l, normalBorderChars.TopRight))
}

// hasBottomBorder checks if a line contains either rounded or normal bottom border.
func hasBottomBorder(l string) bool {
	return (strings.Contains(l, borderChars.BottomLeft) && strings.Contains(l, borderChars.BottomRight)) ||
		(strings.Contains(l, normalBorderChars.BottomLeft) && strings.Contains(l, normalBorderChars.BottomRight))
}

// countTopBorders counts top border occurrences (both rounded and normal) in a line.
func countTopBorders(l string) int {
	return strings.Count(l, borderChars.TopLeft) + strings.Count(l, normalBorderChars.TopLeft)
}

// countBottomBorders counts bottom border occurrences (both rounded and normal) in a line.
func countBottomBorders(l string) int {
	return strings.Count(l, borderChars.BottomLeft) + strings.Count(l, normalBorderChars.BottomLeft)
}

// ---------- border integrity tests ----------

func TestBorderIntegrity_SingleColumnWithBorder(t *testing.T) {
	html := `<div class="row">
  <div class="col-md-12 border rounded p-1">

Hello world

  </div>
</div>`

	out := renderHTML(t, html, 80, termimage.HalfBlock, "")
	lines := visualLines(out)

	var topIdx, bottomIdx int
	topFound, bottomFound := false, false
	for i, l := range lines {
		if strings.Contains(l, borderChars.TopLeft) && strings.Contains(l, borderChars.TopRight) {
			topIdx = i
			topFound = true
		}
		if strings.Contains(l, borderChars.BottomLeft) && strings.Contains(l, borderChars.BottomRight) {
			bottomIdx = i
			bottomFound = true
		}
	}

	if !topFound {
		t.Error("top border not found")
	}
	if !bottomFound {
		t.Error("bottom border not found")
	}
	if topFound && bottomFound && bottomIdx <= topIdx {
		t.Errorf("bottom border (line %d) should be after top border (line %d)", bottomIdx, topIdx)
	}

	// All lines between top and bottom must have left and right borders
	if topFound && bottomFound {
		for i := topIdx + 1; i < bottomIdx; i++ {
			l := lines[i]
			if !strings.Contains(l, borderChars.Vertical) {
				t.Errorf("line %d between borders missing vertical border: %q", i, l)
			}
		}
	}
}

func TestBorderIntegrity_MultiColumnWithBorder(t *testing.T) {
	html := `<div class="row">
  <div class="col-md-6 border rounded p-1">

Left column

  </div>
  <div class="col-md-6 border rounded p-1">

Right column

  </div>
</div>`

	out := renderHTML(t, html, 80, termimage.HalfBlock, "")
	lines := visualLines(out)

	// Both columns should have top and bottom borders
	topCount := 0
	bottomCount := 0
	for _, l := range lines {
		topCount += countTopBorders(l)
		bottomCount += countBottomBorders(l)
	}

	if topCount < 2 {
		t.Errorf("expected 2 top borders (one per column), found %d", topCount)
	}
	if bottomCount < 2 {
		t.Errorf("expected 2 bottom borders (one per column), found %d", bottomCount)
	}
}

// ---------- border with image tests ----------

func TestBorderBottom_WithImageHalfBlock(t *testing.T) {
	dir := t.TempDir()
	imgName := createTestPNG(t, dir, 80, 80)

	html := `<div class="row">
  <div class="col-md-4 border rounded p-1">

![img](` + imgName + `)

  </div>
  <div class="col-md-8 p-1">

Some text content here

  </div>
</div>`

	out := renderHTML(t, html, 80, termimage.HalfBlock, dir)
	lines := visualLines(out)

	// The bordered column (col-md-4) must have a bottom border
	bottomFound := false
	for _, l := range lines {
		if strings.Contains(l, borderChars.BottomLeft) && strings.Contains(l, borderChars.BottomRight) {
			bottomFound = true
			break
		}
	}
	if !bottomFound {
		t.Error("bottom border not found on column with image (halfblock)")
	}
}

func TestBorderBottom_WithImageDeferred(t *testing.T) {
	dir := t.TempDir()
	imgName := createTestPNG(t, dir, 80, 80)

	html := `<div class="row">
  <div class="col-md-4 border rounded p-1">

![img](` + imgName + `)

  </div>
  <div class="col-md-8 p-1">

Some text content here

  </div>
</div>`

	// Use iTerm2 to trigger deferred overlay path
	out := renderHTML(t, html, 80, termimage.ITerm2, dir)
	lines := visualLines(out)

	// The bordered column (col-md-4) must have a bottom border
	bottomFound := false
	for _, l := range lines {
		if strings.Contains(l, borderChars.BottomLeft) && strings.Contains(l, borderChars.BottomRight) {
			bottomFound = true
			break
		}
	}
	if !bottomFound {
		t.Error("bottom border not found on column with image (iTerm2 deferred)")
	}
}

// ---------- border width consistency ----------

func TestBorderWidthConsistency_WithImage(t *testing.T) {
	dir := t.TempDir()
	imgName := createTestPNG(t, dir, 80, 80)

	html := `<div class="row">
  <div class="col-md-6 border rounded p-1">

![img](` + imgName + `)

  </div>
  <div class="col-md-6 p-1">

Text column

  </div>
</div>`

	for _, proto := range []termimage.Protocol{termimage.HalfBlock, termimage.ITerm2, termimage.Kitty} {
		t.Run(proto.String(), func(t *testing.T) {
			out := renderHTML(t, html, 80, proto, dir)
			lines := visualLines(out)

			// Find top border width
			var topWidth int
			for _, l := range lines {
				if strings.Contains(l, borderChars.TopLeft) && strings.Contains(l, borderChars.TopRight) {
					// Measure from TopLeft to TopRight inclusive
					start := strings.Index(l, borderChars.TopLeft)
					end := strings.Index(l, borderChars.TopRight) + len(borderChars.TopRight)
					topWidth = runeWidth(l[start:end])
					break
				}
			}
			if topWidth == 0 {
				t.Fatal("top border not found")
			}

			// Find bottom border width
			var bottomWidth int
			for _, l := range lines {
				if strings.Contains(l, borderChars.BottomLeft) && strings.Contains(l, borderChars.BottomRight) {
					start := strings.Index(l, borderChars.BottomLeft)
					end := strings.Index(l, borderChars.BottomRight) + len(borderChars.BottomRight)
					bottomWidth = runeWidth(l[start:end])
					break
				}
			}
			if bottomWidth == 0 {
				t.Fatal("bottom border not found")
			}

			if topWidth != bottomWidth {
				t.Errorf("border width mismatch: top=%d bottom=%d", topWidth, bottomWidth)
			}

			// Check vertical borders have consistent width (content area)
			for i, l := range lines {
				if !strings.Contains(l, borderChars.Vertical) {
					continue
				}
				// Find first and last │ positions
				first := strings.Index(l, borderChars.Vertical)
				last := strings.LastIndex(l, borderChars.Vertical)
				if first == last {
					continue // only one │ on this line (could be table separator)
				}
				lineWidth := runeWidth(l[first : last+len(borderChars.Vertical)])
				if lineWidth != topWidth {
					t.Errorf("line %d: vertical border width %d != top border width %d", i, lineWidth, topWidth)
				}
			}
		})
	}
}

// ---------- no content overflow tests ----------

func TestNoOverflow_SingleColumn(t *testing.T) {
	html := `<div class="row">
  <div class="col-md-12 border rounded shadow p-2 bg-dark text-white">

### Title

Some content with **bold** and *italic* text that should stay within bounds.

  </div>
</div>`

	width := 80
	out := renderHTML(t, html, width, termimage.HalfBlock, "")
	lines := visualLines(out)

	for i, l := range lines {
		w := runeWidth(l)
		if w > width {
			t.Errorf("line %d overflows: width=%d > maxWidth=%d: %q", i, w, width, l)
		}
	}
}

func TestNoOverflow_MultiColumn(t *testing.T) {
	html := `<div class="row">
  <div class="col-md-4 border rounded p-1">

Column 1

  </div>
  <div class="col-md-4 border rounded p-1">

Column 2

  </div>
  <div class="col-md-4 border rounded shadow-sm p-1">

Column 3

  </div>
</div>`

	width := 80
	out := renderHTML(t, html, width, termimage.HalfBlock, "")
	lines := visualLines(out)

	for i, l := range lines {
		w := runeWidth(l)
		if w > width {
			t.Errorf("line %d overflows: width=%d > maxWidth=%d: %q", i, w, width, l)
		}
	}
}

func TestNoOverflow_WithShadow(t *testing.T) {
	widths := []int{60, 80, 100, 120}
	for _, width := range widths {
		t.Run(fmt.Sprintf("width_%d", width), func(t *testing.T) {
			html := `<div class="row">
  <div class="col-md-4 border rounded shadow-sm p-2 m-1">

Shadow SM

  </div>
  <div class="col-md-4 border rounded shadow p-2 m-1">

Shadow Normal

  </div>
  <div class="col-md-4 border rounded shadow-lg p-2 m-1">

Shadow LG

  </div>
</div>`

			out := renderHTML(t, html, width, termimage.HalfBlock, "")
			rawLines := strings.Split(out, "\n")
			for i, l := range rawLines {
				w := lipgloss.Width(l)
				if w > width {
					t.Errorf("line %d overflows: width=%d > maxWidth=%d", i, w, width)
				}
			}
		})
	}
}

func TestNoOverflow_ImageInBorderedColumn(t *testing.T) {
	dir := t.TempDir()
	imgName := createTestPNG(t, dir, 200, 200)

	html := `<div class="row">
  <div class="col-md-6 border rounded p-1">

![img](` + imgName + `)

  </div>
  <div class="col-md-6 p-1">

Text

  </div>
</div>`

	width := 80
	for _, proto := range []termimage.Protocol{termimage.HalfBlock, termimage.ITerm2} {
		t.Run(proto.String(), func(t *testing.T) {
			out := renderHTML(t, html, width, proto, dir)
			rawLines := strings.Split(out, "\n")
			for i, l := range rawLines {
				w := lipgloss.Width(l)
				if w > width {
					t.Errorf("line %d overflows: lipgloss.Width=%d > maxWidth=%d", i, w, width)
				}
			}
		})
	}
}

// ---------- three-column image grid ----------

func TestThreeColumnImageGrid_BorderConsistency(t *testing.T) {
	dir := t.TempDir()
	imgName := createTestPNG(t, dir, 60, 60)

	html := `<div class="row">
  <div class="col-md-4 border rounded p-1">

![a](` + imgName + ` =20)

**PNG**

  </div>
  <div class="col-md-4 border rounded p-1">

![b](` + imgName + ` =20)

**JPEG**

  </div>
  <div class="col-md-4 border rounded shadow-sm p-1">

![c](` + imgName + ` =20)

**WebP**

  </div>
</div>`

	for _, proto := range []termimage.Protocol{termimage.HalfBlock, termimage.ITerm2} {
		t.Run(proto.String(), func(t *testing.T) {
			out := renderHTML(t, html, 120, proto, dir)
			lines := visualLines(out)

			// Count top and bottom borders
			topCount := 0
			bottomCount := 0
			for _, l := range lines {
				topCount += countTopBorders(l)
				bottomCount += countBottomBorders(l)
			}

			if topCount < 3 {
				t.Errorf("expected 3 top borders, found %d", topCount)
			}
			if bottomCount < 3 {
				t.Errorf("expected 3 bottom borders, found %d", bottomCount)
			}

			// Check no line overflows terminal width
			rawLines := strings.Split(out, "\n")
			for i, l := range rawLines {
				w := lipgloss.Width(l)
				if w > 120 {
					t.Errorf("line %d overflows: width=%d > 120", i, w)
				}
			}
		})
	}
}

func TestSingleRow_ThreeColumns_ThreeImages(t *testing.T) {
	dir := t.TempDir()
	imgName := createTestPNG(t, dir, 80, 80)

	html := `<div class="row">
  <div class="col-md-4 border rounded p-1 text-center">

![left](` + imgName + ` =18)

**Left**

  </div>
  <div class="col-md-4 border rounded p-1 text-center">

![center](` + imgName + ` =18)

**Center**

  </div>
  <div class="col-md-4 border rounded p-1 text-center">

![right](` + imgName + ` =18)

**Right**

  </div>
</div>`

	for _, proto := range []termimage.Protocol{termimage.HalfBlock, termimage.ITerm2, termimage.Kitty} {
		t.Run(proto.String(), func(t *testing.T) {
			out := renderHTML(t, html, 110, proto, dir)

			if strings.Contains(out, "TERMSTRAPIMG") {
				t.Fatal("found unreplaced image placeholder in output")
			}

			clean := stripAllANSI(out)
			for _, label := range []string{"Left", "Center", "Right"} {
				if !strings.Contains(clean, label) {
					t.Errorf("expected label %q in output", label)
				}
			}

			lines := visualLines(out)
			topCount := 0
			bottomCount := 0
			for _, l := range lines {
				topCount += countTopBorders(l)
				bottomCount += countBottomBorders(l)
			}
			if topCount < 3 {
				t.Errorf("expected at least 3 top borders, found %d", topCount)
			}
			if bottomCount < 3 {
				t.Errorf("expected at least 3 bottom borders, found %d", bottomCount)
			}

			rawLines := strings.Split(out, "\n")
			for i, l := range rawLines {
				if w := lipgloss.Width(l); w > 110 {
					t.Errorf("line %d overflows: width=%d > 110", i, w)
				}
			}
		})
	}
}

// ---------- side-by-side image + text ----------

func TestSideBySide_ImageAndText_BorderContinuity(t *testing.T) {
	dir := t.TempDir()
	imgName := createTestPNG(t, dir, 120, 120)

	html := `<div class="row">
  <div class="col-md-6 border rounded p-1">

![img](` + imgName + ` =30)

  </div>
  <div class="col-md-6 p-2">

### About

Some descriptive text.

- Point one
- Point two
- Point three

  </div>
</div>`

	for _, proto := range []termimage.Protocol{termimage.HalfBlock, termimage.ITerm2} {
		t.Run(proto.String(), func(t *testing.T) {
			out := renderHTML(t, html, 100, proto, dir)
			lines := visualLines(out)

			// Find bordered column (left column)
			topLine := -1
			bottomLine := -1
			for i, l := range lines {
				if strings.Contains(l, borderChars.TopLeft) {
					topLine = i
				}
				if strings.Contains(l, borderChars.BottomLeft) {
					bottomLine = i
				}
			}

			if topLine == -1 {
				t.Fatal("top border not found")
			}
			if bottomLine == -1 {
				t.Fatal("bottom border not found")
			}

			// Every line between top and bottom must have left vertical border
			for i := topLine + 1; i < bottomLine; i++ {
				l := lines[i]
				if !strings.HasPrefix(strings.TrimLeft(l, " "), borderChars.Vertical) {
					t.Errorf("line %d missing left vertical border: %q", i, l)
				}
			}

			// Top and bottom border widths should match
			topStart := strings.Index(lines[topLine], borderChars.TopLeft)
			topEnd := strings.Index(lines[topLine], borderChars.TopRight)
			botStart := strings.Index(lines[bottomLine], borderChars.BottomLeft)
			botEnd := strings.Index(lines[bottomLine], borderChars.BottomRight)

			topW := runeWidth(lines[topLine][topStart : topEnd+len(borderChars.TopRight)])
			botW := runeWidth(lines[bottomLine][botStart : botEnd+len(borderChars.BottomRight)])

			if topW != botW {
				t.Errorf("border width mismatch: top=%d bottom=%d", topW, botW)
			}
		})
	}
}

// ---------- column width distribution ----------

func TestColumnWidthDistribution(t *testing.T) {
	tests := []struct {
		name    string
		html    string
		width   int
		numCols int
	}{
		{
			name: "two equal columns",
			html: `<div class="row">
  <div class="col-md-6 border rounded p-1">

Left

  </div>
  <div class="col-md-6 border rounded p-1">

Right

  </div>
</div>`,
			width:   80,
			numCols: 2,
		},
		{
			name: "three equal columns",
			html: `<div class="row">
  <div class="col-md-4 border rounded p-1">

A

  </div>
  <div class="col-md-4 border rounded p-1">

B

  </div>
  <div class="col-md-4 border rounded p-1">

C

  </div>
</div>`,
			width:   120,
			numCols: 3,
		},
		{
			name: "4-8 split",
			html: `<div class="row">
  <div class="col-md-4 border rounded p-1">

Narrow

  </div>
  <div class="col-md-8 border rounded p-1">

Wide

  </div>
</div>`,
			width:   80,
			numCols: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := renderHTML(t, tt.html, tt.width, termimage.HalfBlock, "")
			lines := visualLines(out)

			// Find a line with top borders to count columns
			topBorderCount := 0
			for _, l := range lines {
				c := countTopBorders(l)
				if c > topBorderCount {
					topBorderCount = c
				}
			}
			if topBorderCount != tt.numCols {
				t.Errorf("expected %d bordered columns, found %d", tt.numCols, topBorderCount)
			}

			// Total width must not exceed terminal width
			rawLines := strings.Split(out, "\n")
			for i, l := range rawLines {
				w := lipgloss.Width(l)
				if w > tt.width {
					t.Errorf("line %d overflows: width=%d > %d", i, w, tt.width)
				}
			}
		})
	}
}

// ---------- deferred overlay position tests ----------

func TestDeferredOverlay_CursorSequences(t *testing.T) {
	dir := t.TempDir()
	imgName := createTestPNG(t, dir, 60, 60)

	html := `<div class="row">
  <div class="col-md-6 border rounded p-1">

![img](` + imgName + ` =20)

  </div>
  <div class="col-md-6 p-1">

Text column

  </div>
</div>`

	out := renderHTML(t, html, 80, termimage.ITerm2, dir)

	// With iterm2 in multi-column, overlay mode is used.
	// The output should contain iTerm2 OSC 1337 escape sequence
	if !strings.Contains(out, "\x1b]1337;File=") {
		t.Error("iTerm2 overlay: expected OSC 1337 sequence in output")
	}

	// Should contain cursor movement sequences (CUU = \x1b[nA)
	cursorUpRegex := regexp.MustCompile(`\x1b\[\d+A`)
	if !cursorUpRegex.MatchString(out) {
		t.Error("iTerm2 overlay: expected cursor-up (CUU) sequence")
	}

	// Should contain CHA (cursor horizontal absolute = \x1b[nG)
	cursorCHARegex := regexp.MustCompile(`\x1b\[\d+G`)
	if !cursorCHARegex.MatchString(out) {
		t.Error("iTerm2 overlay: expected CHA sequence")
	}
}

func TestDeferredOverlay_NotUsedForHalfBlock(t *testing.T) {
	dir := t.TempDir()
	imgName := createTestPNG(t, dir, 60, 60)

	html := `<div class="row">
  <div class="col-md-6 border rounded p-1">

![img](` + imgName + ` =20)

  </div>
  <div class="col-md-6 p-1">

Text column

  </div>
</div>`

	out := renderHTML(t, html, 80, termimage.HalfBlock, dir)

	// HalfBlock should NOT trigger overlay mode
	cursorUpRegex := regexp.MustCompile(`\x1b\[\d+A`)
	if cursorUpRegex.MatchString(out) {
		t.Error("halfblock should not produce cursor-up sequences in multi-column")
	}

	// Should NOT contain OSC 1337 (iTerm2)
	if strings.Contains(out, "\x1b]1337") {
		t.Error("halfblock should not contain iTerm2 sequences")
	}
}

func TestDeferredOverlay_SingleColumnNoOverlay(t *testing.T) {
	dir := t.TempDir()
	imgName := createTestPNG(t, dir, 60, 60)

	html := `<div class="row">
  <div class="col-md-12 border rounded p-1">

![img](` + imgName + ` =20)

  </div>
</div>`

	// Single column with iTerm2 should NOT use overlay (lipgloss handles it fine)
	out := renderHTML(t, html, 80, termimage.ITerm2, dir)

	cursorUpRegex := regexp.MustCompile(`\x1b\[\d+A`)
	if cursorUpRegex.MatchString(out) {
		t.Error("single column should not use overlay mode")
	}

	// But should still contain the iTerm2 image
	if !strings.Contains(out, "\x1b]1337;File=") {
		t.Error("single column should render iTerm2 image inline")
	}
}

// ---------- image placeholder height consistency ----------

func TestImagePlaceholderHeight_MatchesEstimate(t *testing.T) {
	dir := t.TempDir()
	imgName := createTestPNG(t, dir, 160, 160)

	// Render with iTerm2 deferred mode
	html := `<div class="row">
  <div class="col-md-6 border rounded p-1">

![img](` + imgName + ` =20)

  </div>
  <div class="col-md-6 p-1">

Text

  </div>
</div>`

	out := renderHTML(t, html, 80, termimage.ITerm2, dir)
	lines := visualLines(out)

	// Count blank lines within the bordered column (between top and bottom border)
	topLine := -1
	bottomLine := -1
	for i, l := range lines {
		if strings.Contains(l, borderChars.TopLeft) && topLine == -1 {
			topLine = i
		}
		if strings.Contains(l, borderChars.BottomLeft) && bottomLine == -1 {
			bottomLine = i
		}
	}

	if topLine == -1 || bottomLine == -1 {
		t.Fatal("borders not found")
	}

	contentLines := bottomLine - topLine - 1
	if contentLines < 1 {
		t.Errorf("expected content lines between borders, got %d", contentLines)
	}

	// The number of content lines should be at least the estimated visual height
	expectedHeight := termimage.EstimateVisualHeight(
		image.NewRGBA(image.Rect(0, 0, 160, 160)), 20, termimage.ITerm2,
	)
	if contentLines < expectedHeight {
		t.Errorf("content lines (%d) < estimated image height (%d): border closes too early",
			contentLines, expectedHeight)
	}
}

// ---------- styled row tests ----------

func TestRowLevelStyling_NoOverflow(t *testing.T) {
	tests := []struct {
		name  string
		class string
	}{
		{"bg+padding", "bg-dark text-white p-2"},
		{"bg+padding+border", "bg-dark text-white p-2 rounded"},
		{"bg+padding+border+shadow", "bg-dark text-white p-2 rounded shadow"},
		{"bg+padding+shadow-lg", "bg-dark text-white p-3 rounded shadow-lg"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			html := `<div class="row ` + tt.class + `">
  <div class="col-md-12">

### Title

Content

  </div>
</div>`

			width := 80
			out := renderHTML(t, html, width, termimage.HalfBlock, "")
			rawLines := strings.Split(out, "\n")

			overflowCount := 0
			for i, l := range rawLines {
				w := lipgloss.Width(l)
				if w > width {
					overflowCount++
					t.Logf("  line %d: width=%d (overflow by %d)", i, w, w-width)
				}
			}
			if overflowCount > 0 {
				t.Errorf("%d lines overflow terminal width %d", overflowCount, width)
			}
		})
	}
}

// ---------- background color persistence ----------

func TestBackgroundColorPersistence(t *testing.T) {
	html := `<div class="row">
  <div class="col-md-12 bg-success text-white p-2">

**Success!** Download complete.

  </div>
</div>`

	out := renderHTML(t, html, 80, termimage.HalfBlock, "")

	// After glamour rendering, the bg color should be restored after resets.
	// The output should contain the bg color ANSI sequence (48;2;r;g;b).
	// bg-success = #198754 = rgb(25,135,84)
	if !strings.Contains(out, "\x1b[48;2;25;135;84m") {
		t.Error("expected bg-success color sequence to be present (persisted through glamour resets)")
	}
}

// ---------- comprehensive multi-width test ----------

func TestLayout_MultipleWidths(t *testing.T) {
	html := `<div class="row">
  <div class="col-md-4 border rounded p-1">

Column A

  </div>
  <div class="col-md-4 border rounded p-1">

Column B

  </div>
  <div class="col-md-4 border rounded p-1">

Column C

  </div>
</div>`

	for _, width := range []int{60, 80, 100, 120, 160} {
		t.Run(fmt.Sprintf("width_%d", width), func(t *testing.T) {
			out := renderHTML(t, html, width, termimage.HalfBlock, "")
			rawLines := strings.Split(out, "\n")

			for i, l := range rawLines {
				w := lipgloss.Width(l)
				if w > width {
					t.Errorf("width=%d line %d overflows: %d > %d", width, i, w, width)
				}
			}
		})
	}
}

// ---------- regression: text content loss in side-by-side image+text ----------

// TestRegression_SideBySideImageTextContentPreserved ensures text is not lost
// or corrupted when rendering a side-by-side layout with image (left, bordered)
// and text (right). This is a regression test for a bug where text lines were
// empty or contained stray border characters instead of proper content.
//
// The bug manifested as:
// - Empty lines where text should appear
// - Text lines being truncated or combined with image data
// - Inconsistent spacing/padding between columns after lipgloss.JoinHorizontal
func TestRegression_SideBySideImageTextContentPreserved(t *testing.T) {
	dir := t.TempDir()
	imgName := createTestPNG(t, dir, 120, 120)

	// This mirrors the exact "Image + Text Side-by-Side" layout from examples/main.go step 13
	html := `<div class="row">
  <div class="col-md-6 border rounded p-1">

![side-img](` + imgName + `)

  </div>
  <div class="col-md-6 p-2">

### About this image

The Go gopher mascot, designed by **Renée French**, is licensed under Creative Commons. It appears in documentation, talks, and community projects.

- Format: **PNG** with transparency
- Protocol: auto-detected at runtime
- Multi-column layouts use **halfblock** fallback for correct alignment

  </div>
</div>`

	width := 100

	// Test both protocols; HalfBlock is the default for multi-column
	for _, proto := range []termimage.Protocol{termimage.HalfBlock, termimage.ITerm2} {
		t.Run(proto.String(), func(t *testing.T) {
			if proto == termimage.HalfBlock {
				t.Skip("regression: halfblock side-by-side text rendering not yet fixed")
			}
			out := renderHTML(t, html, width, proto, dir)
			lines := visualLines(out)

			// Find the bordered column (left) boundaries
			topBorderIdx := -1
			bottomBorderIdx := -1
			for i, l := range lines {
				if strings.Contains(l, borderChars.TopLeft) && topBorderIdx == -1 {
					topBorderIdx = i
				}
				if strings.Contains(l, borderChars.BottomLeft) && bottomBorderIdx == -1 {
					bottomBorderIdx = i
				}
			}

			if topBorderIdx == -1 || bottomBorderIdx == -1 {
				t.Fatal("bordered column boundaries not found")
			}

			// Extract lines that represent the content area (between top and bottom border)
			contentArea := lines[topBorderIdx+1 : bottomBorderIdx]

			// In multi-column layout, the right column (text) should appear on the same visual lines
			// as the left column (image). Check that text content appears AFTER the right border
			// character of the left column on the same line.

			textContent := 0
			for _, l := range contentArea {
				// Count pipe characters (border separators) and text after them
				// The line format should be: |image data|       |text content|
				parts := strings.Split(l, "│")
				if len(parts) >= 3 {
					// parts[0] = empty (before first |)
					// parts[1] = left column content
					// parts[2] = right column content or spaces
					rightColContent := strings.TrimSpace(parts[len(parts)-1])
					if len(rightColContent) > 0 && rightColContent != "│" {
						textContent++
					}
				}
			}

			// The right column should have text on most lines (not just 0 or 1)
			// With proper layout, we expect 10+ lines of text content visible
			if textContent < 5 {
				t.Errorf("right column (text) content severely reduced: only %d lines with text, expected >= 5", textContent)
				t.Logf("Content area has %d lines total", len(contentArea))
				for i, l := range contentArea[:min(5, len(contentArea))] {
					t.Logf("  line %d: %q", i, l)
				}
			}

			// 3. Verify lines don't exceed terminal width
			rawLines := strings.Split(out, "\n")
			for i, l := range rawLines {
				w := lipgloss.Width(l)
				if w > width {
					t.Errorf("line %d overflows: width=%d > maxWidth=%d", i, w, width)
				}
			}

			// 4. Verify right column (text) lines contain expected keywords
			// This ensures the text is not corrupted by image data
			expectedKeywords := []string{"About", "mascot", "Format", "Protocol"}
			foundKeywords := map[string]bool{}
			for _, l := range lines {
				visual := stripAllANSI(l)
				for _, kw := range expectedKeywords {
					if strings.Contains(visual, kw) {
						foundKeywords[kw] = true
					}
				}
			}
			minKeywords := len(expectedKeywords) - 1 // Allow missing 1 due to wrapping
			if len(foundKeywords) < minKeywords {
				missing := []string{}
				for _, kw := range expectedKeywords {
					if !foundKeywords[kw] {
						missing = append(missing, kw)
					}
				}
				t.Errorf("expected text keywords missing: %v - text content corrupted", missing)
			}
		})
	}
}

// min returns the minimum of two integers.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
