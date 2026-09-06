package testutil

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/hinshun/vt10x"
)

// VirtualScreen wraps vt10x.Terminal to provide an in-memory 2D character and color grid.
type VirtualScreen struct {
	Term   vt10x.Terminal
	Width  int
	Height int
}

// NewScreen creates and initializes a VirtualScreen with rendered terminal output.
func NewScreen(width, height int, content string) *VirtualScreen {
	lines := strings.Split(content, "\n")
	if height <= 0 {
		height = len(lines)
		if height < 24 {
			height = 24
		}
	}
	if width <= 0 {
		width = 80
	}

	term := vt10x.New(vt10x.WithSize(width, height))
	// Standard VT terminal requires CRLF to advance cursor down and to column 0
	normalized := strings.ReplaceAll(strings.ReplaceAll(content, "\r\n", "\n"), "\n", "\r\n")
	_, _ = term.Write([]byte(normalized))

	return &VirtualScreen{
		Term:   term,
		Width:  width,
		Height: height,
	}
}

// ShowInteractive prints the rendered terminal output and pauses for user validation
// if the INTERACTIVE=1 or TERMSTRAP_INTERACTIVE=1 environment variable is set.
func ShowInteractive(t *testing.T, title string, content string) {
	t.Helper()
	if os.Getenv("INTERACTIVE") != "1" && os.Getenv("TERMSTRAP_INTERACTIVE") != "1" {
		return
	}

	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		fmt.Printf("\n--- [Visual Test: %s] ---\n%s\n-------------------------\n", title, content)
		return
	}
	defer tty.Close()

	fmt.Fprintf(tty, "\n\033[1;36m══════════════════════════════════════════════════════════════════\033[0m\n")
	fmt.Fprintf(tty, "  \033[1m[Visual Test: %s]\033[0m\n", title)
	fmt.Fprintf(tty, "\033[1;36m══════════════════════════════════════════════════════════════════\033[0m\n\n")
	fmt.Fprintf(tty, "%s\n\n", content)
	fmt.Fprintf(tty, "\033[1;33m[Appuyez sur Entrée pour continuer (ou 'q' pour arrêter)...]\033[0m ")

	var buf [1]byte
	_, _ = tty.Read(buf[:])
	if buf[0] == 'q' || buf[0] == 'Q' {
		t.Fatalf("Test interactif interrompu par l'utilisateur.")
	}
}

// Cell returns the vt10x.Glyph at the given (x, y) coordinate.
func (s *VirtualScreen) Cell(x, y int) vt10x.Glyph {
	if x < 0 || x >= s.Width || y < 0 || y >= s.Height {
		return vt10x.Glyph{Char: ' '}
	}
	return s.Term.Cell(x, y)
}

// Char returns the character rune at the given (x, y) coordinate.
func (s *VirtualScreen) Char(x, y int) rune {
	cell := s.Cell(x, y)
	if cell.Char == 0 {
		return ' '
	}
	return cell.Char
}

// Text extracts a string of length `length` starting at (x, y).
func (s *VirtualScreen) Text(x, y, length int) string {
	var sb strings.Builder
	for i := range length {
		sb.WriteRune(s.Char(x+i, y))
	}
	return sb.String()
}

// Line returns the raw text content of row y up to width.
func (s *VirtualScreen) Line(y int) string {
	return s.Text(0, y, s.Width)
}

// Dump returns the complete screen buffer as a multi-line string for debugging.
func (s *VirtualScreen) Dump() string {
	var sb strings.Builder
	for y := range s.Height {
		line := strings.TrimRight(s.Line(y), " ")
		sb.WriteString(fmt.Sprintf("%3d | %s\n", y, line))
	}
	return sb.String()
}

// AssertChar asserts that the character at (x, y) matches the expected rune.
func (s *VirtualScreen) AssertChar(t *testing.T, x, y int, expected rune) {
	t.Helper()
	got := s.Char(x, y)
	if got != expected {
		t.Errorf("AssertChar failed at (%d, %d): expected %q (U+%04X), got %q (U+%04X)\nScreen dump:\n%s",
			x, y, expected, expected, got, got, s.Dump())
	}
}

// AssertText asserts that contiguous horizontal text starting at (x, y) matches expected.
func (s *VirtualScreen) AssertText(t *testing.T, x, y int, expected string) {
	t.Helper()
	expectedRunes := []rune(expected)
	got := s.Text(x, y, len(expectedRunes))
	if got != expected {
		t.Errorf("AssertText failed at (%d, %d): expected %q, got %q\nScreen dump:\n%s",
			x, y, expected, got, s.Dump())
	}
}

// AssertBgColor asserts the background ANSI color of the cell at (x, y).
// expectedColor can be hex ("#79a2f7", "79a2f7"), RGB ("121;162;247", "121,162,247"), or "default".
func (s *VirtualScreen) AssertBgColor(t *testing.T, x, y int, expectedColor string) {
	t.Helper()
	cell := s.Cell(x, y)
	expectedRGB, err := parseColor(expectedColor)
	if err != nil {
		t.Fatalf("AssertBgColor: invalid expected color %q: %v", expectedColor, err)
	}

	actualBG := uint32(cell.BG)
	if expectedColor == "default" {
		if actualBG != uint32(vt10x.DefaultBG) {
			t.Errorf("AssertBgColor failed at (%d, %d): expected default BG, got %06x\nScreen dump:\n%s",
				x, y, actualBG, s.Dump())
		}
		return
	}

	// For truecolor in vt10x, cell.BG is stored directly as 0xRRGGBB
	actualRGB := actualBG & 0xFFFFFF
	if actualRGB != expectedRGB {
		t.Errorf("AssertBgColor failed at (%d, %d): expected #%06x (%s), got #%06x\nScreen dump:\n%s",
			x, y, expectedRGB, expectedColor, actualRGB, s.Dump())
	}
}

// AssertBorderBox verifies that a box with standard lipgloss rounded or normal borders
// exactly occupies the bounding box (x, y, width, height).
func (s *VirtualScreen) AssertBorderBox(t *testing.T, x, y, width, height int) {
	t.Helper()
	if width < 2 || height < 2 {
		t.Fatalf("AssertBorderBox: invalid dimensions (%d x %d)", width, height)
	}

	x2 := x + width - 1
	y2 := y + height - 1

	// Check top-left corner
	tl := s.Char(x, y)
	if tl != '╭' && tl != '┌' && tl != '+' {
		t.Errorf("AssertBorderBox top-left corner mismatch at (%d, %d): expected '╭' or '┌', got %q (U+%04X)\nScreen dump:\n%s",
			x, y, tl, tl, s.Dump())
	}

	// Check top-right corner
	tr := s.Char(x2, y)
	if tr != '╮' && tr != '┐' && tr != '+' {
		t.Errorf("AssertBorderBox top-right corner mismatch at (%d, %d): expected '╮' or '┐', got %q (U+%04X)\nScreen dump:\n%s",
			x2, y, tr, tr, s.Dump())
	}

	// Check bottom-left corner
	bl := s.Char(x, y2)
	if bl != '╰' && bl != '└' && bl != '+' {
		t.Errorf("AssertBorderBox bottom-left corner mismatch at (%d, %d): expected '╰' or '└', got %q (U+%04X)\nScreen dump:\n%s",
			x, y2, bl, bl, s.Dump())
	}

	// Check bottom-right corner
	br := s.Char(x2, y2)
	if br != '╯' && br != '┘' && br != '+' {
		t.Errorf("AssertBorderBox bottom-right corner mismatch at (%d, %d): expected '╯' or '┘', got %q (U+%04X)\nScreen dump:\n%s",
			x2, y2, br, br, s.Dump())
	}

	// Check top & bottom horizontal edges
	for col := x + 1; col < x2; col++ {
		topChar := s.Char(col, y)
		if topChar != '─' && topChar != '-' {
			t.Errorf("AssertBorderBox top edge mismatch at (%d, %d): expected '─', got %q\nScreen dump:\n%s",
				col, y, topChar, s.Dump())
			break
		}
		botChar := s.Char(col, y2)
		if botChar != '─' && botChar != '-' {
			t.Errorf("AssertBorderBox bottom edge mismatch at (%d, %d): expected '─', got %q\nScreen dump:\n%s",
				col, y2, botChar, s.Dump())
			break
		}
	}

	// Check left & right vertical edges
	for row := y + 1; row < y2; row++ {
		leftChar := s.Char(x, row)
		if leftChar != '│' && leftChar != '|' {
			t.Errorf("AssertBorderBox left edge mismatch at (%d, %d): expected '│', got %q\nScreen dump:\n%s",
				x, row, leftChar, s.Dump())
			break
		}
		rightChar := s.Char(x2, row)
		if rightChar != '│' && rightChar != '|' {
			t.Errorf("AssertBorderBox right edge mismatch at (%d, %d): expected '│', got %q\nScreen dump:\n%s",
				x2, row, rightChar, s.Dump())
			break
		}
	}
}

func parseColor(col string) (uint32, error) {
	col = strings.TrimSpace(strings.ToLower(col))
	if col == "" || col == "default" {
		return 0, nil
	}

	// Hex format: #RRGGBB, 0xRRGGBB, or RRGGBB
	if strings.HasPrefix(col, "#") {
		col = strings.TrimPrefix(col, "#")
	} else if strings.HasPrefix(col, "0x") {
		col = strings.TrimPrefix(col, "0x")
	}

	if len(col) == 6 {
		val, err := strconv.ParseUint(col, 16, 32)
		if err == nil {
			return uint32(val), nil
		}
	}

	// Semicolon or comma format: "121;162;247" or "121,162,247"
	var parts []string
	if strings.Contains(col, ";") {
		parts = strings.Split(col, ";")
	} else if strings.Contains(col, ",") {
		col = strings.TrimPrefix(col, "rgb(")
		col = strings.TrimSuffix(col, ")")
		parts = strings.Split(col, ",")
	}

	if len(parts) == 3 {
		r, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
		g, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
		b, err3 := strconv.Atoi(strings.TrimSpace(parts[2]))
		if err1 == nil && err2 == nil && err3 == nil {
			return uint32((r&0xFF)<<16 | (g&0xFF)<<8 | (b&0xFF)), nil
		}
	}

	return 0, fmt.Errorf("cannot parse color %q", col)
}
