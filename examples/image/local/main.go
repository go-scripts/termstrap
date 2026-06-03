// Example: local — Render local image files from disk.
// Demonstrates the RootPath feature for resolving relative paths.
//
// Usage:
//
//	# First, create a test image:
//	curl -o /tmp/test-gopher.png https://go.dev/doc/gopher/frontpage.png
//
//	# Then run:
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
	termimage "github.com/go-scripts/termstrap/image"
	"golang.org/x/term"
)

func main() {
	caps := termimage.Detect()
	width := 80
	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
		width = w
	}

	fmt.Printf("Protocol: %s | Width: %d\n\n", caps.Protocol, width)

	// Create a temporary directory with test images
	tmpDir, err := os.MkdirTemp("", "termstrap-local-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating temp dir: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)

	// Generate a simple test PNG (color gradient)
	generateTestPNG(filepath.Join(tmpDir, "gradient.png"), 200, 100)
	// Generate a red square
	generateSolidPNG(filepath.Join(tmpDir, "red.png"), 80, 80, color.RGBA{220, 50, 50, 255})
	// Generate a green square
	generateSolidPNG(filepath.Join(tmpDir, "green.png"), 80, 80, color.RGBA{50, 180, 50, 255})
	// Generate a blue square
	generateSolidPNG(filepath.Join(tmpDir, "blue.png"), 80, 80, color.RGBA{50, 100, 220, 255})

	fmt.Printf("Test images generated in: %s\n\n", tmpDir)

	content := `# Local Image Files

## Gradient (generated PNG)

![gradient](gradient.png =50)

---

## Color Swatches in Grid

<div class="row">
  <div class="col-md-4 border rounded p-1 text-center">

![red](red.png =20)

**Red**

  </div>
  <div class="col-md-4 border rounded p-1 text-center">

![green](green.png =20)

**Green**

  </div>
  <div class="col-md-4 border rounded p-1 text-center">

![blue](blue.png =20)

**Blue**

  </div>
</div>

---

These images were generated on the fly and loaded from disk using ` + "`RootPath`" + `.
`

	m := termstrap.Model{
		Content:  content,
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

func generateTestPNG(path string, w, h int) {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r := uint8(float64(x) / float64(w) * 255)
			g := uint8(float64(y) / float64(h) * 255)
			b := uint8(128)
			img.Set(x, y, color.RGBA{r, g, b, 255})
		}
	}
	savePNG(path, img)
}

func generateSolidPNG(path string, w, h int, c color.RGBA) {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, c)
		}
	}
	savePNG(path, img)
}

func savePNG(path string, img image.Image) {
	f, err := os.Create(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating %s: %v\n", path, err)
		return
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding %s: %v\n", path, err)
	}
}
