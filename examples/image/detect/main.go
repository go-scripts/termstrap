// Example: detect — Shows detected terminal capabilities and renders
// a test image using the auto-detected best protocol.
//
// Usage:
//
//	go run ./examples/image/detect/
//	TERMSTRAP_IMAGE_PROTOCOL=sixel go run ./examples/image/detect/
package main

import (
	"fmt"
	"os"

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

	fmt.Println("╔══════════════════════════════════════════╗")
	fmt.Println("║   Termstrap — Terminal Detection Report  ║")
	fmt.Println("╠══════════════════════════════════════════╣")
	fmt.Printf("║  Protocol:   %-27s ║\n", caps.Protocol)
	fmt.Printf("║  TrueColor:  %-27v ║\n", caps.TrueColor)
	fmt.Printf("║  Columns:    %-27d ║\n", caps.ColCount)
	fmt.Printf("║  Rows:       %-27d ║\n", caps.RowCount)
	fmt.Printf("║  TERM:       %-27s ║\n", os.Getenv("TERM"))
	fmt.Printf("║  TERM_PROGRAM: %-25s ║\n", os.Getenv("TERM_PROGRAM"))
	fmt.Println("╚══════════════════════════════════════════╝")
	fmt.Println()

	content := fmt.Sprintf(`# Image Protocol: %s

Rendering a remote PNG image with auto-detected protocol:

![gopher](https://go.dev/doc/gopher/frontpage.png =40)

The image above was rendered using the **%s** protocol.
`, caps.Protocol, caps.Protocol)

	m := termstrap.Model{
		Content: content,
		Width:   width,
	}
	output, err := m.Render()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Print(output)
}
