package termstrap

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	termimage "github.com/go-scripts/termstrap/image"
)

func createTestImage(t *testing.T, dir, filename string, w, h int) string {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x * 20), G: uint8(y * 20), B: 150, A: 255})
		}
	}
	f, err := os.Create(filepath.Join(dir, filename))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
	return filename
}

func TestCache_ImageAndANSI(t *testing.T) {
	dir := t.TempDir()
	imgName := createTestImage(t, dir, "cached_test.png", 60, 40)

	ClearImageCache()

	m := Model{
		Content:  "![Test](" + imgName + " =30)",
		Width:    80,
		RootPath: dir,
	}

	// First render: loads from disk, renders, populates cache
	out1, err := m.Render()
	if err != nil {
		t.Fatalf("first render error: %v", err)
	}

	cache := DefaultCache()
	resolvedPath := filepath.Join(dir, imgName)

	if _, ok := cache.GetImage(resolvedPath); !ok {
		t.Errorf("expected image to be cached for key %s", resolvedPath)
	}

	// Second render: should use ANSI cache
	out2, err := m.Render()
	if err != nil {
		t.Fatalf("second render error: %v", err)
	}
	if out1 != out2 {
		t.Errorf("expected identical output from cache")
	}

	// Invalidation
	InvalidateImage(resolvedPath)
	if _, ok := cache.GetImage(resolvedPath); ok {
		t.Errorf("expected image to be invalidated from cache")
	}
}

func TestModel_MaxImageWidth(t *testing.T) {
	dir := t.TempDir()
	imgName := createTestImage(t, dir, "wide_test.png", 200, 100)

	m := Model{
		Content:       "![Test](" + imgName + " =100)",
		Width:         120,
		RootPath:      dir,
		MaxImageWidth: 20, // hard ceiling
	}

	out, err := m.Render()
	if err != nil {
		t.Fatalf("render error: %v", err)
	}

	// Each line in the rendered image should not exceed MaxImageWidth + padding
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		stripped := stripAllANSI(line)
		if len(strings.TrimSpace(stripped)) > 24 {
			t.Errorf("line exceeded MaxImageWidth: len=%d, text=%q", len(strings.TrimSpace(stripped)), stripped)
		}
	}
}

func TestModel_ColorMode256(t *testing.T) {
	dir := t.TempDir()
	imgName := createTestImage(t, dir, "colormode_test.png", 40, 40)

	mTC := Model{
		Content:       "![Test](" + imgName + " =20)",
		Width:         80,
		RootPath:      dir,
		ColorMode:     termimage.ColorModeTrueColor,
		ImageRenderer: termimage.NewRenderer(termimage.WithProtocol(termimage.HalfBlock), termimage.WithColorMode(termimage.ColorModeTrueColor)),
		CachePolicy:   CacheNoStore,
	}
	outTC, err := mTC.Render()
	if err != nil {
		t.Fatalf("TrueColor render error: %v", err)
	}

	m256 := Model{
		Content:       "![Test](" + imgName + " =20)",
		Width:         80,
		RootPath:      dir,
		ColorMode:     termimage.ColorMode256,
		ImageRenderer: termimage.NewRenderer(termimage.WithProtocol(termimage.HalfBlock), termimage.WithColorMode(termimage.ColorMode256)),
		CachePolicy:   CacheNoStore,
	}
	out256, err := m256.Render()
	if err != nil {
		t.Fatalf("256-color render error: %v", err)
	}

	if !strings.Contains(outTC, ";2;") {
		t.Errorf("expected TrueColor escape in TrueColor mode")
	}
	if !strings.Contains(out256, ";5;") {
		t.Errorf("expected 256-color escape in ColorMode256")
	}
	if len(out256) >= len(outTC) {
		t.Errorf("expected 256-color mode to produce fewer bytes (%d) than TrueColor (%d)", len(out256), len(outTC))
	}
}

func TestModel_DisableImages(t *testing.T) {
	m := Model{
		Content:       "![My Poster](https://example.com/poster.jpg)",
		Width:         80,
		DisableImages: true,
	}

	out, err := m.Render()
	if err != nil {
		t.Fatalf("render error: %v", err)
	}

	if !strings.Contains(out, "My Poster") || !strings.Contains(out, "https://example.com/poster.jpg") {
		t.Errorf("expected fallback link when DisableImages is true, got: %s", out)
	}
}

func TestMarkdown_Attributes_ColorAndMaxWidth(t *testing.T) {
	dir := t.TempDir()
	imgName := createTestImage(t, dir, "attr_test.png", 80, 80)

	content := "![Poster](" + imgName + ", color=256, termstrap-max-width=24)"
	m := Model{
		Content:       content,
		Width:         80,
		RootPath:      dir,
		CachePolicy:   CacheNoStore,
		ImageRenderer: termimage.NewRenderer(termimage.WithProtocol(termimage.HalfBlock)),
	}

	out, err := m.Render()
	if err != nil {
		t.Fatalf("render error: %v", err)
	}

	if !strings.Contains(out, ";5;") {
		t.Errorf("expected 256-color ANSI output from Markdown attribute color=256")
	}
}

func TestHTML_ImgTags(t *testing.T) {
	dir := t.TempDir()
	imgName := createTestImage(t, dir, "html_test.png", 80, 80)

	html := `<div class="row">
  <div class="col-md-6">
    <img src="` + imgName + `" class="ansi-256 max-w-20" />
  </div>
</div>`

	m := Model{
		Content:       html,
		Width:         80,
		RootPath:      dir,
		CachePolicy:   CacheNoStore,
		ImageRenderer: termimage.NewRenderer(termimage.WithProtocol(termimage.HalfBlock)),
	}

	out, err := m.Render()
	if err != nil {
		t.Fatalf("render error: %v", err)
	}

	if !strings.Contains(out, ";5;") {
		t.Errorf("expected 256-color ANSI output from HTML class ansi-256")
	}
}

func TestCache_Expiry(t *testing.T) {
	cache := NewMemoryImageCache(10, 10*time.Millisecond)
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))

	cache.SetImage("key1", img, 10*time.Millisecond)
	if _, ok := cache.GetImage("key1"); !ok {
		t.Fatal("expected key1 to be present immediately")
	}

	time.Sleep(20 * time.Millisecond)
	if _, ok := cache.GetImage("key1"); ok {
		t.Errorf("expected key1 to be expired")
	}
}
