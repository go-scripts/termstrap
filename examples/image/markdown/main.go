// Example: markdown (now HTML images) — Demonstrates image rendering
// at various sizes and contexts.
//
// Usage:
//
//	go run ./examples/image/markdown/
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
	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
		width = w
	}

	caps := termimage.Detect()
	fmt.Printf("Protocol: %s | Width: %d\n\n", caps.Protocol, width)

	content := `<h1>HTML Images Demo</h1>

<h2>Default Width</h2>
<div><img src="https://go.dev/doc/gopher/frontpage.png" alt="gopher" /></div>

<hr />

<h2>Image with Text</h2>
<p>Some text <b>before</b> the image.</p>
<div><img src="https://go.dev/doc/gopher/pkg.png" alt="inline" /></div>
<p>Some text <b>after</b> the image.</p>

<hr />

<h2>Multiple Images in Sequence</h2>
<div><img src="https://go.dev/doc/gopher/frontpage.png" alt="first" /></div>
<div><img src="https://www.gstatic.com/webp/gallery/1.webp" alt="second" /></div>
`

	m := termstrap.Model{
		HTML:  content,
		Width: width,
	}
	output, err := m.Render()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Print(output)
}
