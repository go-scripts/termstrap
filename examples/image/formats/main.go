// Example: formats — Demonstrates rendering of all supported image formats:
// PNG, JPEG, WebP, and GIF.
//
// Usage:
//
//	go run ./examples/image/formats/
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
	fmt.Printf("Protocol: %s | Terminal: %dx%d\n\n", caps.Protocol, width, height)

	content := `<h1>Image Format Test</h1>

<h2>PNG (lossless, transparency support)</h2>
<div><img src="https://go.dev/doc/gopher/frontpage.png" alt="png" /></div>

<hr />

<h2>JPEG (lossy, photographs)</h2>
<div><img src="https://picsum.photos/id/237/200/300.jpg" alt="jpeg" /></div>

<hr />

<h2>WebP (modern format, lossy &amp; lossless)</h2>
<div><img src="https://www.gstatic.com/webp/gallery/1.webp" alt="webp" /></div>

<hr />

<p>All formats above should render correctly.</p>
`

	m := termstrap.New(content, termstrap.WithWidth(width))
	output, err := m.Render()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Print(output)
}
