// Example: caching — Demonstrates in-memory HTTP and ANSI caching in termstrap,
// cache policies (CacheDefault, CacheReload, CacheNoStore), and invalidation.
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
)

func main() {
	width := 80

	tmpDir, err := os.MkdirTemp("", "termstrap-caching-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)

	imgPath := filepath.Join(tmpDir, "avatar.png")
	generateSquarePNG(imgPath, 80, 80)

	termstrap.ClearImageCache()

	content := `# Démonstration du Cache d'Images

![Avatar](avatar.png =24)

Ce document teste la vitesse de rechargement avec et sans cache.
`

	m := termstrap.Model{
		Content:  content,
		Width:    width,
		RootPath: tmpDir,
	}

	fmt.Println("================================================================================")
	fmt.Println(" termstrap: Image & ANSI Caching Demo")
	fmt.Println("================================================================================")

	// 1. First render (Cache Miss -> loads image & computes ANSI)
	start1 := time.Now()
	out1, err := m.Render()
	duration1 := time.Since(start1)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Render 1 error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("[1] Premier rendu (Cache Miss)       : %v\n", duration1)

	// 2. Second render (Cache Hit -> 0 ms instant re-use from memory)
	start2 := time.Now()
	_, err = m.Render()
	duration2 := time.Since(start2)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Render 2 error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("[2] Deuxième rendu (Cache Hit ANSI)  : %v\n", duration2)

	// 3. Render with CacheReload (forces recalculation)
	mReload := m
	mReload.CachePolicy = termstrap.CacheReload
	start3 := time.Now()
	_, _ = mReload.Render()
	duration3 := time.Since(start3)
	fmt.Printf("[3] Rendu avec CacheReload           : %v\n", duration3)

	// 4. Invalidation
	termstrap.ClearImageCache()
	fmt.Println("\nCache global vidé via termstrap.ClearImageCache().")

	fmt.Println("\n--- Aperçu du document rendu ---")
	fmt.Print(out1)
}

func generateSquarePNG(path string, w, h int) {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{R: 70, G: 130, B: 240, A: 255})
		}
	}
	f, _ := os.Create(path)
	defer f.Close()
	_ = png.Encode(f, img)
}
