package main

// Example: protocols — Render the same image with each supported graphics
// protocol so you can visually compare output quality.
//
// By default, only the auto-detected protocol and the halfblock fallback
// are rendered (unsupported protocols like Kitty on iTerm2 can produce
// garbage escape sequences that pollute the shell).
//
// Usage:
//
//	go run ./examples/image/protocols/
//
// Force a specific protocol via env:
//
//	TERMSTRAP_IMAGE_PROTOCOL=kitty    go run ./examples/image/protocols/
//	TERMSTRAP_IMAGE_PROTOCOL=iterm2   go run ./examples/image/protocols/
//	TERMSTRAP_IMAGE_PROTOCOL=sixel    go run ./examples/image/protocols/
//	TERMSTRAP_IMAGE_PROTOCOL=halfblock go run ./examples/image/protocols/
//
// Render ALL protocols (including unsupported — may produce artifacts):
//
//	TERMSTRAP_SHOW_ALL=1 go run ./examples/image/protocols/

import (
	"fmt"
	"image"
	"os"

	termimage "github.com/go-scripts/termstrap/image"
	"golang.org/x/term"
)

const testImageURL = "https://go.dev/doc/gopher/frontpage.png"

func main() {
	width := 80
	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
		width = w
	}

	caps := termimage.Detect()
	detected := caps.Protocol
	fmt.Printf("Auto-detected: %s | TrueColor: %v | %dx%d\n\n",
		detected, caps.TrueColor, caps.ColCount, caps.RowCount)

	// Load the test image once
	fmt.Print("Fetching test image... ")
	img, err := termimage.LoadFromURL(testImageURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\nError loading image: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("OK (%dx%d)\n\n", img.Bounds().Dx(), img.Bounds().Dy())

	renderWidth := width / 2
	if renderWidth > 60 {
		renderWidth = 60
	}

	showAll := os.Getenv("TERMSTRAP_SHOW_ALL") == "1"

	protocols := []termimage.Protocol{
		termimage.HalfBlock,
		termimage.Sixel,
		termimage.ITerm2,
		termimage.Kitty,
	}

	for _, proto := range protocols {
		supported := proto == termimage.HalfBlock || proto == detected
		if !supported && !showAll {
			fmt.Printf("┌─── %s ─── (skipped: not supported by this terminal)\n", proto)
			fmt.Println("└───────────────")
			fmt.Println()
			continue
		}
		renderWithProtocol(proto, img, renderWidth, !supported)
	}
}

func renderWithProtocol(proto termimage.Protocol, img image.Image, width int, unsupported bool) {
	renderer := termimage.NewRenderer(termimage.WithProtocol(proto))

	label := proto.String()
	if unsupported {
		label += " ⚠ forced, may produce artifacts"
	}
	fmt.Printf("┌─── %s ───\n", label)

	result, err := renderer.Render(img, width)
	if err != nil {
		fmt.Printf("│ ERROR: %v\n", err)
		fmt.Println("└───────────────")
		fmt.Println()
		return
	}

	fmt.Print(result)
	fmt.Println("└───────────────")
	fmt.Println()
}
