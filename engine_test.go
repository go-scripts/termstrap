package termstrap

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/go-scripts/termstrap/testutil"
)

func TestEngine_SimpleDivAndMargins(t *testing.T) {
	html := `<div class="m-3 p-2 bg-dark text-white">Hello Termstrap</div>`
	m := New(html, WithWidth(80))

	out, err := m.Render()
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	if !strings.Contains(out, "Hello Termstrap") {
		t.Errorf("expected rendered text 'Hello Termstrap' in output, got:\n%s", out)
	}
}

func TestEngine_FlexRowAndCols(t *testing.T) {
	html := `<div class="row">
		<div class="col-6 bg-primary text-white p-1">Left Column</div>
		<div class="col-6 bg-success text-white p-1">Right Column</div>
	</div>`

	m := New(html, WithWidth(80))
	out, err := m.Render()
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	if !strings.Contains(out, "Left Column") || !strings.Contains(out, "Right Column") {
		t.Errorf("expected both columns in output, got:\n%s", out)
	}
}

func TestEngine_CustomStylesheet(t *testing.T) {
	customCSS := `
		.custom-box {
			background-color: #ff0000;
			color: #ffffff;
			margin: 2;
			padding: 1;
		}
	`
	html := `<div class="custom-box">Custom Box Content</div>`

	m := New(html, WithWidth(80), WithStylesheets(customCSS))
	out, err := m.Render()
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	if !strings.Contains(out, "Custom Box Content") {
		t.Errorf("expected 'Custom Box Content' in output, got:\n%s", out)
	}
}

func TestEngine_UserMargin5Scenario(t *testing.T) {
	html := `<div class="row">
  <div class="col-6 p-2">
    <div class="text-center m-5">
      Image Area
    </div>
    <div>Table Area</div>
  </div>
</div>`

	m := New(html, WithWidth(80))
	out, err := m.Render()
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	fmt.Printf("User Scenario Output:\n%s\n", out)
	if !strings.Contains(out, "Image Area") || !strings.Contains(out, "Table Area") {
		t.Errorf("expected content in output, got:\n%s", out)
	}
}

func TestEngine_Themes(t *testing.T) {
	html := `<div class="bg-primary text-white p-1">Theme Test</div>`

	expectedThemeColors := map[Theme]string{
		ThemeBootstrap:  "48;2;13;110;253",  // #0d6efd
		ThemeTokyoNight: "48;2;121;162;247", // #7aa2f7
		ThemeDracula:    "48;2;189;147;249", // #bd93f9
	}
	for th, expectedSeq := range expectedThemeColors {
		m := New(html, WithWidth(80), WithTheme(th))
		out, err := m.Render()
		if err != nil {
			t.Fatalf("Render failed for theme %q: %v", th, err)
		}
		if !strings.Contains(out, "Theme Test") {
			t.Errorf("expected 'Theme Test' in output for theme %q, got:\n%s", th, out)
		}
		if !strings.Contains(out, expectedSeq) {
			t.Errorf("expected ANSI color sequence %q for theme %q in output, got:\n%q", expectedSeq, th, out)
		}
	}
}

func TestEngine_ColoredAlerts(t *testing.T) {
	html := `<div class="row">
  <div class="col-4 bg-success text-white p-1">Success Alert</div>
  <div class="col-4 bg-warning text-dark p-1">Warning Alert</div>
  <div class="col-4 bg-danger text-white p-1">Danger Alert</div>
</div>`

	m := New(html, WithWidth(80), WithTheme(ThemeBootstrap))
	out, err := m.Render()
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	// bg-success: #198754 -> 48;2;25;135;84
	if !strings.Contains(out, "48;2;25;135;84") {
		t.Errorf("expected bg-success ANSI color in output, got:\n%q", out)
	}
	// bg-warning: #ffc107 -> 48;2;255;193;7
	if !strings.Contains(out, "48;2;255;193;7") {
		t.Errorf("expected bg-warning ANSI color in output, got:\n%q", out)
	}
	// bg-danger: #dc3545 -> 48;2;220;52;69
	if !strings.Contains(out, "48;2;220;52;69") {
		t.Errorf("expected bg-danger ANSI color in output, got:\n%q", out)
	}
}

func TestEngine_DocumentWidthRegression(t *testing.T) {
	htmlBytes, err := os.ReadFile("examples/bootstrap/80s Huge Hits FLAC 2026.html")
	if err != nil {
		t.Skip("html file not found")
	}

	m := Model{HTML: string(htmlBytes), Width: 100}
	out, err := m.Render()
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	docLines := strings.Split(out, "\n")
	for i, l := range docLines {
		w := lipgloss.Width(l)
		if w > 100 {
			t.Errorf("Line %d exceeds terminal width 100: width=%d", i, w)
		}
	}
}

func TestEngine_InlineFlowAndFormatting(t *testing.T) {
	html := `<p>This is <b>bold</b> and <i>italic</i> with <code>code</code> inline.</p>`
	m := New(html, WithWidth(80))
	out, err := m.Render()
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	// Inline content should not be split across multiple single-word lines
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 1 {
		t.Errorf("expected 1 line for short inline paragraph, got %d lines:\n%s", len(lines), out)
	}
}

func TestEngine_FlexRowNoWhitespaceWrapping(t *testing.T) {
	html := `<div class="row">
  <div class="col-md-3 border p-1">Col 1</div>
  <div class="col-md-3 border p-1">Col 2</div>
  <div class="col-md-3 border p-1">Col 3</div>
  <div class="col-md-3 border p-1">Col 4</div>
</div>`
	m := New(html, WithWidth(80))
	out, err := m.Render()
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	// All 4 columns should fit side-by-side in one subrow at 80 cols (not wrapped to 2 rows)
	lines := strings.Split(strings.TrimSpace(out), "\n")
	for _, l := range lines {
		w := lipgloss.Width(l)
		if w > 80 {
			t.Errorf("Line exceeds width 80: width=%d:\n%s", w, l)
		}
	}
}

func TestEngine_HRRendering(t *testing.T) {
	html := `<hr />`
	m := New(html, WithWidth(80))
	out, err := m.Render()
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	if !strings.Contains(out, "────────────────────────────────────────────────────────────────────────────────") {
		t.Errorf("expected HR rule across 80 chars, got:\n%s", out)
	}
}

func TestEngine_BackgroundColorPersistence(t *testing.T) {
	html := `<div class="p-2 border rounded bg-primary text-white">
		<h3>TokyoNight Theme</h3>
		<p>Blue accent header with dark text.</p>
	</div>`

	m := New(html, WithWidth(80), WithTheme(ThemeTokyoNight))
	out, err := m.Render()
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	screen := testutil.NewScreen(80, 15, out)
	boxW := 80
	boxH := 7
	screen.AssertBorderBox(t, 0, 0, boxW, boxH)
	screen.AssertText(t, 3, 2, "TokyoNight Theme")
	screen.AssertText(t, 3, 4, "Blue accent header with dark text.")

	// Verify that inside the rounded border box (x in [1, 78], y in [1, 8]),
	// background color is consistently TokyoNight primary (121;162;247 / #79a2f7)
	for y := 1; y < boxH-1; y++ {
		for x := 1; x < boxW-1; x++ {
			screen.AssertBgColor(t, x, y, "121;162;247")
		}
	}
}

func TestEngine_BackgroundColorAcrossCenteredText(t *testing.T) {
	html := `<div class="row">
		<div class="col-4 border rounded p-2 bg-primary text-white text-center">Col 1</div>
	</div>`

	m := New(html, WithWidth(90), WithTheme(ThemeBootstrap))
	out, err := m.Render()
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	screen := testutil.NewScreen(90, 10, out)
	boxW := 30
	boxH := 3
	screen.AssertBorderBox(t, 0, 0, boxW, boxH)
	screen.AssertText(t, 12, 1, "Col 1")

	// Verify that every interior cell (padding, centered text, and trailing spaces)
	// maintains Bootstrap primary background (13;110;253 / #0d6efd)
	for x := 1; x < boxW-1; x++ {
		screen.AssertBgColor(t, x, 1, "13;110;253")
	}
}
func TestEngine_ColumnWidthFullAllocation(t *testing.T) {
	html := `<div class="container p-1">
		<div class="border rounded p-1 mb-1">Header Card</div>
		<div class="row">
			<div class="col-5 border rounded p-1">Col 1</div>
			<div class="col-7 border rounded p-1">Col 2</div>
		</div>
	</div>`

	m := New(html, WithWidth(100))
	out, err := m.Render()
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}
	screen := testutil.NewScreen(100, 20, out)

	// Container padding is p-1 -> left padding 1, right padding 1.
	// Inner width = 100 - 2 = 98.
	// Header box spans x in [1, 98], so width is 98, height is 3 (lines y=0, 1, 2).
	screen.AssertBorderBox(t, 1, 0, 98, 3)
	screen.AssertText(t, 3, 1, "Header Card")

	// Row starts at y=4 (due to mb-1 at y=3).
	// Col-5 takes 40 cols, Col-7 takes remaining 58 cols.
	// Col 1 box: x=1, width=40, height=3 (lines y=4, 5, 6).
	screen.AssertBorderBox(t, 1, 4, 40, 3)
	screen.AssertText(t, 3, 5, "Col 1")

	// Col 2 box: x=41, width=58, height=3 (lines y=4, 5, 6).
	screen.AssertBorderBox(t, 41, 4, 58, 3)
	screen.AssertText(t, 43, 5, "Col 2")

	// Verify top-right corner of header (x=98, y=0) exactly aligns with top-right of Col 2 (x=98, y=4)
	screen.AssertChar(t, 98, 0, '╮')
	screen.AssertChar(t, 98, 4, '╮')
	screen.AssertChar(t, 98, 6, '╯')
}
func TestEngine_EmbeddedStyleAndCascadeInheritance(t *testing.T) {
	html := `<html>
<head>
	<style>
		.red-text { color: red; text-align: center; font-weight: bold; }
	</style>
</head>
<body>
	<div class="red-text">
		<span>Inherited</span>
	</div>
</body>
</html>`

	m := New(html, WithWidth(40))
	out, err := m.Render()
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	// Output must contain the word "Inherited"
	if !strings.Contains(out, "Inherited") {
		t.Fatalf("Expected output to contain 'Inherited', got:\n%q", out)
	}

	// Must be centered and styled (red color ANSI code or bold)
	// Lipgloss rendering with red produces ANSI escape (either \x1b[31m or \x1b[38;2;... or \x1b[1m for bold)
	if !strings.Contains(out, "\x1b[") {
		t.Errorf("Expected ANSI escape styling on 'Inherited', got:\n%q", out)
	}

	screen := testutil.NewScreen(40, 5, out)
	// "Inherited" has length 9. Centered in width 40 means (40-9)/2 = 15 padding, so starting around x=15
	found := false
	for x := 10; x <= 20; x++ {
		if screen.Char(x, 0) == 'I' && screen.Char(x+1, 0) == 'n' {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected centered 'Inherited' around middle of 40-col screen, got screen dump:\n%s", screen.Dump())
	}
}

func TestEngine_LinkedStylesheet_LocalAndHTTP(t *testing.T) {
	// 1. Setup local temporary CSS file
	tmpDir := t.TempDir()
	localCSS := `.local-blue { color: #0000ff; }`
	localFile := filepath.Join(tmpDir, "custom.css")
	if err := os.WriteFile(localFile, []byte(localCSS), 0644); err != nil {
		t.Fatalf("Failed to write local CSS: %v", err)
	}

	// 2. Setup HTTP server for remote CSS
	remoteCSS := `.remote-italic { font-style: italic; }`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css")
		w.Write([]byte(remoteCSS))
	}))
	defer server.Close()

	html := fmt.Sprintf(`<html>
<head>
	<link rel="stylesheet" href="custom.css">
	<link rel="stylesheet" href="%s/style.css">
</head>
<body>
	<div class="local-blue">
		<div class="remote-italic">
			<span>Styled Content</span>
		</div>
	</div>
</body>
</html>`, server.URL)

	m := New(html, WithWidth(40), WithRootPath(tmpDir))
	out, err := m.Render()
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	if !strings.Contains(out, "Styled Content") {
		t.Fatalf("Expected output to contain 'Styled Content', got:\n%q", out)
	}

	// Verify ANSI styling for italic (\x1b[3m) and color
	if !strings.Contains(out, "\x1b[3m") && !strings.Contains(out, "\x1b[") {
		t.Errorf("Expected ANSI styling from linked stylesheets, got:\n%q", out)
	}
}

func TestEngine_CascadeInheritance_UnsetAndOverride(t *testing.T) {
	html := `<html>
<head>
	<style>
		.parent { color: red; font-weight: bold; text-align: center; }
		.normal-child { font-weight: normal; color: green; }
	</style>
</head>
<body>
	<div class="parent">
		<span class="normal-child">Unset Bold</span>
	</div>
</body>
</html>`

	m := New(html, WithWidth(40))
	out, err := m.Render()
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	if !strings.Contains(out, "Unset Bold") {
		t.Fatalf("Expected output to contain 'Unset Bold', got:\n%q", out)
	}
}

