package termstrap

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	termimage "github.com/go-scripts/termstrap/image"
)

// createOverlayTestPNG writes a solid-color PNG to dir and returns the filename.
func createOverlayTestPNG(t *testing.T, dir, name string) string {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 50, 50))
	for y := 0; y < 50; y++ {
		for x := 0; x < 50; x++ {
			img.Set(x, y, color.RGBA{200, 100, 50, 255})
		}
	}
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

func TestRenderWithOverlays_HalfBlock(t *testing.T) {
	dir := t.TempDir()
	imgName := createOverlayTestPNG(t, dir, "test.png")

	m := Model{
		Content:       "# Hello\n\n![img](" + imgName + ")",
		Width:         80,
		RootPath:      dir,
		ImageRenderer: termimage.NewRenderer(termimage.WithProtocol(termimage.HalfBlock)),
	}
	textRender, err := m.Render()
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}

	textOverlay, overlays, err := m.RenderWithOverlays()
	if err != nil {
		t.Fatalf("RenderWithOverlays() error: %v", err)
	}
	if len(overlays) != 0 {
		t.Errorf("expected 0 overlays for HalfBlock, got %d", len(overlays))
	}
	if textOverlay != textRender {
		t.Errorf("RenderWithOverlays text should match Render() for HalfBlock")
	}
}

func TestRenderWithOverlays_NoImages(t *testing.T) {
	m := Model{
		Content:       "# Hello\n\nJust text, no images.",
		Width:         80,
		ImageRenderer: termimage.NewRenderer(termimage.WithProtocol(termimage.ITerm2)),
	}
	text, overlays, err := m.RenderWithOverlays()
	if err != nil {
		t.Fatalf("RenderWithOverlays() error: %v", err)
	}
	if len(overlays) != 0 {
		t.Errorf("expected 0 overlays without images, got %d", len(overlays))
	}
	if !strings.Contains(text, "Hello") {
		t.Error("expected rendered text to contain 'Hello'")
	}
}

func TestRenderWithOverlays_NativeProtocol(t *testing.T) {
	dir := t.TempDir()
	imgName := createOverlayTestPNG(t, dir, "test.png")

	m := Model{
		Content:       "# Title\n\n![img](" + imgName + ")\n\nAfter image.",
		Width:         80,
		RootPath:      dir,
		ImageRenderer: termimage.NewRenderer(termimage.WithProtocol(termimage.ITerm2)),
	}
	text, overlays, err := m.RenderWithOverlays()
	if err != nil {
		t.Fatalf("RenderWithOverlays() error: %v", err)
	}

	// Text should NOT contain OSC 1337 escape sequences
	if strings.Contains(text, "\x1b]1337") {
		t.Error("rendered text should not contain OSC 1337 escape sequences")
	}
	// Overlays should contain the escape sequence
	if len(overlays) == 0 {
		t.Fatal("expected at least one overlay for iTerm2 protocol")
	}
	for i, ov := range overlays {
		if !strings.Contains(ov.Escape, "\x1b]1337") {
			t.Errorf("overlay[%d] should contain OSC 1337 escape sequence", i)
		}
		if ov.Width <= 0 {
			t.Errorf("overlay[%d] Width should be > 0, got %d", i, ov.Width)
		}
		if ov.Height <= 0 {
			t.Errorf("overlay[%d] Height should be > 0, got %d", i, ov.Height)
		}
	}
}

func TestRenderWithOverlays_GridLayout(t *testing.T) {
	dir := t.TempDir()
	img1 := createOverlayTestPNG(t, dir, "img1.png")
	img2 := createOverlayTestPNG(t, dir, "img2.png")

	content := `<div class="row">
<div class="col-6">

![img1](` + img1 + `)

</div>
<div class="col-6">

![img2](` + img2 + `)

</div>
</div>`
	m := Model{
		Content:       content,
		Width:         80,
		RootPath:      dir,
		ImageRenderer: termimage.NewRenderer(termimage.WithProtocol(termimage.ITerm2)),
	}
	text, overlays, err := m.RenderWithOverlays()
	if err != nil {
		t.Fatalf("RenderWithOverlays() error: %v", err)
	}
	if text == "" {
		t.Error("expected non-empty rendered text")
	}

	// Should have 2 overlays (one per column)
	if len(overlays) != 2 {
		t.Fatalf("expected 2 overlays, got %d", len(overlays))
	}
	// Second overlay should have Col > first overlay Col
	if overlays[1].Col <= overlays[0].Col {
		t.Errorf("second overlay Col (%d) should be > first overlay Col (%d)", overlays[1].Col, overlays[0].Col)
	}
}
