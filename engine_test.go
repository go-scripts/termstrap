package termstrap

import (
	"fmt"
	"strings"
	"testing"
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

	for _, th := range []Theme{ThemeBootstrap, ThemeTokyoNight, ThemeDracula, "unknown_fallback"} {
		m := New(html, WithWidth(80), WithTheme(th))
		out, err := m.Render()
		if err != nil {
			t.Fatalf("Render failed for theme %q: %v", th, err)
		}
		if !strings.Contains(out, "Theme Test") {
			t.Errorf("expected 'Theme Test' in output for theme %q, got:\n%s", th, out)
		}
	}
}
