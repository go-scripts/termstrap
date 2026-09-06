// Example: colormodes — Compares TrueColor (24-bit RGB) and ANSI 256 color modes.
//
// Usage:
//
//	go run ./examples/image/colormodes/
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
	width := 80
	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
		width = w
	}

	tmpDir, err := os.MkdirTemp("", "termstrap-colormode-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create temp dir: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)

	imgPath := filepath.Join(tmpDir, "poster.png")
	createGradientImage(imgPath, 120, 80)

	content := `<h1>Color Modes Comparison</h1>

<div class="row">
  <div class="col-md-6 border rounded p-1 text-center">
    <div><img src="poster.png" alt="poster" /></div>
    <div><b>TrueColor Mode</b></div>
  </div>
  <div class="col-md-6 border rounded p-1 text-center">
    <div><img src="poster.png" alt="poster" /></div>
    <div><b>256 Colors Mode</b></div>
  </div>
</div>
`

	m := termstrap.New(
		content,
		termstrap.WithWidth(width),
		termstrap.WithRootPath(tmpDir),
		termstrap.WithColorMode(termimage.ColorModeTrueColor),
	)

	out, err := m.Render()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Print(out)
}

func createGradientImage(path string, width, height int) {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			c := color.RGBA{
				R: uint8(x * 255 / width),
				G: uint8(y * 255 / height),
				B: uint8((x + y) * 255 / (width + height)),
				A: 255,
			}
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
