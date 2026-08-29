// Example: colormodes — Demonstrates TrueColor, 256-color, and 16-color image rendering modes,
// along with sequence optimization and Markdown / HTML attribute overrides.
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
)

func main() {
	width := 80

	// Create a temporary directory with a rich test gradient
	tmpDir, err := os.MkdirTemp("", "termstrap-colormodes-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating temp dir: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)

	imgPath := filepath.Join(tmpDir, "poster.png")
	generateSamplePoster(imgPath, 120, 80)

	fmt.Println("================================================================================")
	fmt.Println(" termstrap: Color Modes & ANSI Optimization Demo")
	fmt.Println("================================================================================")

	// 1. TrueColor mode
	mTC := termstrap.Model{
		Content:       "# Mode TrueColor 24-bit\n\n![Poster](poster.png =30)",
		Width:         width,
		RootPath:      tmpDir,
		ColorMode:     termimage.ColorModeTrueColor,
		ImageRenderer: termimage.NewRenderer(termimage.WithProtocol(termimage.HalfBlock), termimage.WithColorMode(termimage.ColorModeTrueColor)),
		CachePolicy:   termstrap.CacheNoStore,
	}
	outTC, _ := mTC.Render()

	// 2. ANSI 256 Colors mode
	m256 := termstrap.Model{
		Content:       "# Mode ANSI 256 Couleurs (50-60% plus léger)\n\n![Poster](poster.png =30)",
		Width:         width,
		RootPath:      tmpDir,
		ColorMode:     termimage.ColorMode256,
		ImageRenderer: termimage.NewRenderer(termimage.WithProtocol(termimage.HalfBlock), termimage.WithColorMode(termimage.ColorMode256)),
		CachePolicy:   termstrap.CacheNoStore,
	}
	out256, _ := m256.Render()

	// 3. Grid comparison with in-content attributes
	gridContent := `# Comparaison Côte à Côte (Attributs Markdown)

<div class="row">
  <div class="col-md-6 border rounded p-1 text-center">

![TC](poster.png, width=24, color=truecolor)

**TrueColor (24-bit)**

  </div>
  <div class="col-md-6 border rounded p-1 text-center">

![256](poster.png, width=24, color=256)

**ANSI 256 (8-bit)**

  </div>
</div>
`
	mGrid := termstrap.Model{
		Content:     gridContent,
		Width:       width,
		RootPath:    tmpDir,
		CachePolicy: termstrap.CacheNoStore,
	}
	outGrid, _ := mGrid.Render()

	fmt.Printf("\n--- Rendu TrueColor (Taille ANSI: %d octets) ---\n", len(outTC))
	fmt.Print(outTC)

	fmt.Printf("\n--- Rendu ANSI 256 (Taille ANSI: %d octets - Réduction: %.1f%%) ---\n",
		len(out256), float64(len(outTC)-len(out256))/float64(len(outTC))*100)
	fmt.Print(out256)

	fmt.Println("\n--- Rendu Grille Comparatif ---")
	fmt.Print(outGrid)
}

func generateSamplePoster(path string, w, h int) {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r := uint8(float64(x) / float64(w) * 255)
			g := uint8(float64(y) / float64(h) * 200)
			b := uint8(180)
			img.Set(x, y, color.RGBA{r, g, b, 255})
		}
	}
	f, _ := os.Create(path)
	defer f.Close()
	_ = png.Encode(f, img)
}
