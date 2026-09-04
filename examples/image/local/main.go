// Example: local — Loads and renders local image files from disk.
//
// Usage:
//
//	go run ./examples/image/local/
package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"

	"github.com/go-scripts/termstrap"
	"golang.org/x/term"
)

func main() {
	width := 80
	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
		width = w
	}

	tmpDir, err := os.MkdirTemp("", "termstrap-local-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create temp dir: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)

	createSolidImage(filepath.Join(tmpDir, "red.png"), 100, 60, color.RGBA{R: 220, G: 50, B: 50, A: 255})
	createSolidImage(filepath.Join(tmpDir, "green.png"), 100, 60, color.RGBA{R: 50, G: 180, B: 50, A: 255})
	createSolidImage(filepath.Join(tmpDir, "blue.png"), 100, 60, color.RGBA{R: 50, G: 100, B: 220, A: 255})

	content := `<h1>Local Image Files</h1>

<h2>Gradient (generated PNG)</h2>

<div><img src="red.png" alt="red" /></div>

<hr />

<h2>Color Swatches in Grid</h2>

<div class="row">
  <div class="col-md-4 border rounded p-1 text-center">
    <div><img src="red.png" alt="red" /></div>
    <div><b>Red</b></div>
  </div>
  <div class="col-md-4 border rounded p-1 text-center">
    <div><img src="green.png" alt="green" /></div>
    <div><b>Green</b></div>
  </div>
  <div class="col-md-4 border rounded p-1 text-center">
    <div><img src="blue.png" alt="blue" /></div>
    <div><b>Blue</b></div>
  </div>
</div>

<hr />

<p>These images were generated on the fly and loaded from disk using <code>RootPath</code>.</p>
`

	m := termstrap.Model{
		HTML:     content,
		Width:    width,
		RootPath: tmpDir,
	}
	output, err := m.Render()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Print(output)
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
