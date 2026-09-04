// Example: caching — Demonstrates image cache policies (Default, Reload, NoStore)
// and custom cache TTL.
//
// Usage:
//
//	go run ./examples/image/caching/
package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"time"

	"github.com/go-scripts/termstrap"
	"golang.org/x/term"
)

func main() {
	width := 80
	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
		width = w
	}

	tmpDir, err := os.MkdirTemp("", "termstrap-cache-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create temp dir: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)

	imgPath := filepath.Join(tmpDir, "test.png")
	createSolidImage(imgPath, 80, 40, color.RGBA{R: 200, G: 100, B: 50, A: 255})

	content := `<h1>Cache Policy Demo</h1>

<h2>Default Caching (CacheDefault)</h2>
<div><img src="test.png" alt="default" /></div>
`

	m := termstrap.Model{
		HTML:     content,
		Width:    width,
		RootPath: tmpDir,
		CacheTTL: 10 * time.Minute,
	}

	out, err := m.Render()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Print(out)
}

func createSolidImage(path string, width, height int, c color.Color) {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, c)
		}
	}
	f, err := os.Create(path)
	if err != nil {
		return
	}
	defer f.Close()
	_ = png.Encode(f, img)
}
