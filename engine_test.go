package termstrap

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
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

	bgSeq := "48;2;121;162;247"
	lines := strings.Split(out, "\n")
	for i, line := range lines {
		// Content lines inside borders should preserve TokyoNight blue bg
		if strings.Contains(line, "TokyoNight Theme") || strings.Contains(line, "Blue accent") {
			if !strings.Contains(line, bgSeq) {
				t.Errorf("Line %d expected to contain bgSeq %s, got:\n%q", i, bgSeq, line)
			}
			// Ensure reset code is not left without background re-applied
			if strings.Contains(line, "\x1b[0m   ") {
				t.Errorf("Line %d has uncolored trailing spaces after reset: %q", i, line)
			}
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

	bgSeq := "48;2;13;110;253"
	lines := strings.Split(out, "\n")
	for i, line := range lines {
		if strings.Contains(line, "Col 1") {
			if !strings.Contains(line, bgSeq) {
				t.Errorf("Line %d expected to contain bgSeq %s, got:\n%q", i, bgSeq, line)
			}
			if strings.Contains(line, "\x1b[0m   ") {
				t.Errorf("Line %d has uncolored spaces after reset: %q", i, line)
			}
		}
	}
}
