// Example: detect — Tests terminal capability detection and auto-configuration.
//
// Usage:
//
//	go run ./examples/image/detect/
package main

import (
	"fmt"
	"os"

	"github.com/go-scripts/termstrap"
	termimage "github.com/go-scripts/termstrap/image"
	"golang.org/x/term"
)

func main() {
	width := 80
	height := 24
	if w, h, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
		width = w
		height = h
	}

	caps := termimage.Detect()

	fmt.Println("╔══════════════════════════════════════════╗")
	fmt.Println("║   Termstrap — Terminal Detection Report  ║")
	fmt.Println("╠══════════════════════════════════════════╣")
	fmt.Printf("║  Protocol:   %-28s║\n", caps.Protocol)
	fmt.Printf("║  TrueColor:  %-28v║\n", caps.TrueColor)
	fmt.Printf("║  Columns:    %-28d║\n", width)
	fmt.Printf("║  Rows:       %-28d║\n", height)
	fmt.Printf("║  TERM:       %-28s║\n", os.Getenv("TERM"))
	fmt.Printf("║  TERM_PROGRAM: %-26s║\n", os.Getenv("TERM_PROGRAM"))
	fmt.Println("╚══════════════════════════════════════════╝")
	fmt.Println()

	content := `<h1>Image Protocol Demo</h1>
<p>Rendering a remote PNG image with auto-detected protocol:</p>
<div><img src="https://go.dev/doc/gopher/frontpage.png" alt="gopher" /></div>
`

	m := termstrap.New(content, termstrap.WithWidth(width))
	output, err := m.Render()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Print(output)
}
