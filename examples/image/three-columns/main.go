// Example: three-columns — Render 3 local images in a single row with 3 columns.
// Demonstrates a compact gallery layout using col-md-4 + border styles.
//
// Usage:
//
//	go run ./examples/image/three-columns/
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
	width := 100
	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
		width = w
	}

	fmt.Printf("Protocol: %s | Width: %d\n\n", caps.Protocol, width)

	tmpDir, err := os.MkdirTemp("", "termstrap-three-cols-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating temp dir: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)

	writeSolidPNG(filepath.Join(tmpDir, "red.png"), 90, 90, color.RGBA{220, 70, 70, 255})
	writeSolidPNG(filepath.Join(tmpDir, "green.png"), 90, 90, color.RGBA{70, 180, 90, 255})
	writeSolidPNG(filepath.Join(tmpDir, "blue.png"), 90, 90, color.RGBA{70, 110, 220, 255})

	content := `# Images into columns
	
## Three Images In One Row

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


----

## Multi rows nested with 3 columns layout with images, tables, and text

<div class="row">
  <div class="col-md-5 border p-1">
    <div class="row">
      <div class="col-6 border p-1">

## Left

Local image:

![red](red.png =10)

      </div>
      <div class="col-6 border p-1">

## Left 2

Another image:

![green](green.png =10)

      </div>
    </div>
  </div>
  <div class="col-md-3 border p-1">

## Right

Blue image:

![blue](blue.png =10)

  </div>
  <div class="col-md-4 border p-1">

## Full right

Content here.

  </div>
</div>

----

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

func writeSolidPNG(path string, w, h int, c color.RGBA) {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, c)
		}
	}

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
