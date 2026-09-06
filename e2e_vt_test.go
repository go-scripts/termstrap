package termstrap

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-scripts/termstrap/testutil"
)

func TestE2E_ResponsiveStackingVsSideBySide(t *testing.T) {
	htmlPath := filepath.Join("examples", "html", "80s Huge Hits FLAC 2026.html")
	content, err := os.ReadFile(htmlPath)
	if err != nil {
		t.Fatalf("Failed to read test HTML file %s: %v", htmlPath, err)
	}

	htmlStr := string(content)

	t.Run("Stacked at width 70 (below col-md 80 breakpoint)", func(t *testing.T) {
		m := New(htmlStr, WithWidth(70), WithDisableImages(true))
		out, err := m.Render()
		if err != nil {
			t.Fatalf("Render failed: %v", err)
		}
		testutil.ShowInteractive(t, "Stacked Torrent Details at width 70", out)
		screen := testutil.NewScreen(70, 0, out)

		// Title at top (container p-2 -> x=2, y=0)
		screen.AssertText(t, 2, 0, "80s Huge Hits FLAC 2026")

		// Below width 80 (md breakpoint), col-md-6 columns stack vertically at full container width (66 cols: [2..67]).
		// First column (Cover & metadata table) starts at y=2, height=13
		screen.AssertBorderBox(t, 2, 2, 66, 13)
		screen.AssertText(t, 6, 8, "Catégorie")
		screen.AssertText(t, 36, 8, "Musiques")

		var botY70 int
		for y := 16; y < 100; y++ {
			if screen.Char(2, y) == '╰' {
				botY70 = y
				break
			}
		}
		height70 := botY70 - 15 + 1
		screen.AssertBorderBox(t, 2, 15, 66, height70)
		screen.AssertText(t, 5, 17, "Description")
		screen.AssertText(t, 8, 19, "• A-Ha - Take On Me [03:48]")
	})
	t.Run("Side-by-Side at width 120 (above col-md 80 breakpoint)", func(t *testing.T) {
		m := New(htmlStr, WithWidth(120), WithDisableImages(true))
		out, err := m.Render()
		if err != nil {
			t.Fatalf("Render failed: %v", err)
		}
		testutil.ShowInteractive(t, "Side-by-Side Torrent Details at width 120", out)
		screen := testutil.NewScreen(120, 0, out)

		// Title at top
		screen.AssertText(t, 2, 0, "80s Huge Hits FLAC 2026")

		// At width 120, col-md-6 divides space into two 58-col columns side-by-side:
		// Left column: x in [2, 59] (width 58, height 13)
		// Right column: x in [60, 117] (width 58, height 68)
		// Both top corners must be on the exact same row (y=2)
		screen.AssertBorderBox(t, 2, 2, 58, 13)

		// Top corners aligned on row 2
		screen.AssertChar(t, 2, 2, '╭')
		screen.AssertChar(t, 59, 2, '╮')
		screen.AssertChar(t, 60, 2, '╭')
		screen.AssertChar(t, 117, 2, '╮')

		var botY120 int
		for y := 3; y < 100; y++ {
			if screen.Char(60, y) == '╰' {
				botY120 = y
				break
			}
		}
		height120 := botY120 - 2 + 1
		screen.AssertBorderBox(t, 60, 2, 58, height120)
		screen.AssertText(t, 63, 4, "Description")
		screen.AssertText(t, 66, 6, "• A-Ha - Take On Me [03:48]")
	})
}

func TestE2E_ThreeColumnGridAlignment(t *testing.T) {
	html := `<div class="row">
		<div class="col-4 border rounded p-1">Col 1</div>
		<div class="col-4 border rounded p-1">Col 2</div>
		<div class="col-4 border rounded p-1">Col 3</div>
	</div>`

	m := New(html, WithWidth(90))
	out, err := m.Render()
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}
	testutil.ShowInteractive(t, "3-Column Responsive Grid (col-4) at width 90", out)
	screen := testutil.NewScreen(90, 0, out)

	// 90 width / 3 cols = 30 cols each (width 30, height 3)
	// Col 1: x in [0, 29]
	// Col 2: x in [30, 59]
	// Col 3: x in [60, 89]
	screen.AssertBorderBox(t, 0, 0, 30, 3)
	screen.AssertBorderBox(t, 30, 0, 30, 3)
	screen.AssertBorderBox(t, 60, 0, 30, 3)

	screen.AssertText(t, 2, 1, "Col 1")
	screen.AssertText(t, 32, 1, "Col 2")
	screen.AssertText(t, 62, 1, "Col 3")

	// Top corners aligned on row 0
	screen.AssertChar(t, 0, 0, '╭')
	screen.AssertChar(t, 29, 0, '╮')
	screen.AssertChar(t, 30, 0, '╭')
	screen.AssertChar(t, 59, 0, '╮')
	screen.AssertChar(t, 60, 0, '╭')
	screen.AssertChar(t, 89, 0, '╮')

	// Bottom corners aligned on row 2
	screen.AssertChar(t, 0, 2, '╰')
	screen.AssertChar(t, 29, 2, '╯')
	screen.AssertChar(t, 30, 2, '╰')
	screen.AssertChar(t, 59, 2, '╯')
	screen.AssertChar(t, 60, 2, '╰')
	screen.AssertChar(t, 89, 2, '╯')
}

func TestE2E_ThemeDraculaAndTokyoNightColors(t *testing.T) {
	html := `<div class="row">
		<div class="col-6 bg-primary text-white p-1">Primary Box</div>
		<div class="col-6 bg-success text-white p-1">Success Box</div>
	</div>`

	t.Run("TokyoNight Theme Colors", func(t *testing.T) {
		m := New(html, WithWidth(80), WithTheme(ThemeTokyoNight))
		out, err := m.Render()
		if err != nil {
			t.Fatalf("Render failed: %v", err)
		}
		testutil.ShowInteractive(t, "TokyoNight Theme Colors", out)
		screen := testutil.NewScreen(80, 0, out)

		// TokyoNight Primary is rendered as 121;162;247 (#79a2f7)
		// Left col is 40 cols wide [0, 39]
		for x := 0; x < 40; x++ {
			screen.AssertBgColor(t, x, 0, "121;162;247")
		}

		// TokyoNight Success is rendered as 158;206;105 (#9ece69)
		// Right col is 40 cols wide [40, 79]
		for x := 40; x < 80; x++ {
			screen.AssertBgColor(t, x, 0, "158;206;105")
		}
	})

	t.Run("Dracula Theme Colors", func(t *testing.T) {
		m := New(html, WithWidth(80), WithTheme(ThemeDracula))
		out, err := m.Render()
		if err != nil {
			t.Fatalf("Render failed: %v", err)
		}
		testutil.ShowInteractive(t, "Dracula Theme Colors", out)
		screen := testutil.NewScreen(80, 0, out)

		// Dracula Primary is 189;147;249 (#bd93f9)
		for x := 0; x < 40; x++ {
			screen.AssertBgColor(t, x, 0, "189;147;249")
		}

		// Dracula Success is 80;250;123 (#50fa7b)
		for x := 40; x < 80; x++ {
			screen.AssertBgColor(t, x, 0, "80;250;123")
		}
	})
}

func TestE2E_ColoredAlerts(t *testing.T) {
	html := `<div class="row">
		<div class="col-6 bg-success text-white p-1">Success Alert</div>
		<div class="col-6 bg-danger text-white p-1">Danger Alert</div>
	</div>`

	m := New(html, WithWidth(80), WithTheme(ThemeBootstrap))
	out, err := m.Render()
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}
	testutil.ShowInteractive(t, "Bootstrap Colored Alerts (Success & Danger)", out)
	screen := testutil.NewScreen(80, 0, out)

	// Bootstrap bg-success is 25;135;84 (#198754)
	for x := 0; x < 40; x++ {
		screen.AssertBgColor(t, x, 0, "25;135;84")
	}

	// Bootstrap bg-danger is 220;52;69 (#dc3445)
	for x := 40; x < 80; x++ {
		screen.AssertBgColor(t, x, 0, "220;52;69")
	}
}

func TestE2E_NestedCardsAndBorders(t *testing.T) {
	html := `<div class="row">
		<div class="col-6 border p-1">
			<div>Outer Left</div>
			<div class="row">
				<div class="col-6 border rounded p-1">Inner L</div>
				<div class="col-6 border rounded p-1">Inner R</div>
			</div>
		</div>
		<div class="col-6 border p-1">Outer Right</div>
	</div>`

	m := New(html, WithWidth(80))
	out, err := m.Render()
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}
	testutil.ShowInteractive(t, "Nested Cards & Responsive Rows", out)
	screen := testutil.NewScreen(80, 0, out)

	// Outer left card: x in [0, 39], width 40
	// Outer right card: x in [40, 79], width 40
	screen.AssertChar(t, 0, 0, '┌')
	screen.AssertChar(t, 39, 0, '┐')
	screen.AssertChar(t, 40, 0, '┌')
	screen.AssertChar(t, 79, 0, '┐')

	screen.AssertText(t, 2, 1, "Outer Left")
	screen.AssertText(t, 42, 1, "Outer Right")

	// Nested inner columns inside left column (padding p-1 leaves inner width 38 cols [1..38])
	// Inner L: x in [2, 19] (width 18)
	// Inner R: x in [20, 37] (width 18)
	screen.AssertChar(t, 2, 2, '╭')
	screen.AssertChar(t, 19, 2, '╮')
	screen.AssertChar(t, 20, 2, '╭')
	screen.AssertChar(t, 37, 2, '╮')
}
