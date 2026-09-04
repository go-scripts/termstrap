// Example: grid — Demonstrates images inside Bootstrap-like grid layouts
// with borders, shadows, backgrounds, and responsive stacking.
//
// Usage:
//
//	go run ./examples/image/grid/
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

	content := `<h1>Grid + Images + Styling</h1>

<h2>Card Layout — Image + Metadata</h2>

<div class="row">
  <div class="col-12 border rounded shadow-sm p-1">
    <div><img src="https://go.dev/doc/gopher/frontpage.png" alt="poster" /></div>
    <h3>The Go Gopher</h3>
    <p><b>Rob Pike &amp; Renée French</b> — 2009</p>
    <table>
      <tr><th>Info</th><th>Value</th></tr>
      <tr><td>Language</td><td>Go</td></tr>
      <tr><td>Mascot</td><td>Gopher</td></tr>
      <tr><td>Created</td><td>2009</td></tr>
      <tr><td>License</td><td>Creative Commons</td></tr>
    </table>
    <blockquote>"Don't communicate by sharing memory; share memory by communicating."</blockquote>
  </div>
</div>

<hr />

<h2>Gallery — Three Images with Labels</h2>

<div class="row">
  <div class="col-md-4 border rounded p-1 text-center">
    <div><img src="https://go.dev/doc/gopher/pkg.png" alt="img1" /></div>
    <div><b>PNG Format</b></div>
  </div>
  <div class="col-md-4 border rounded p-1 text-center">
    <div><img src="https://go.dev/doc/gopher/frontpage.png" alt="img2" /></div>
    <div><b>Another PNG</b></div>
  </div>
  <div class="col-md-4 border rounded shadow-lg p-1 text-center">
    <div><img src="https://www.gstatic.com/webp/gallery/1.webp" alt="img3" /></div>
    <div><b>WebP + Shadow</b></div>
  </div>
</div>

<hr />

<h2>Styled Text Columns — Colors &amp; Backgrounds</h2>

<div class="row">
  <div class="col-md-4 bg-primary text-white p-2 border rounded">
    <div><b>Primary</b> — Blue background with white text for important content.</div>
  </div>
  <div class="col-md-4 bg-dark text-white p-2 border rounded">
    <div><b>Dark</b> — Dark background for contrast sections and emphasis.</div>
  </div>
  <div class="col-md-4 bg-success text-white p-2 border rounded">
    <div><b>Success</b> — Green background for positive messages and confirmations.</div>
  </div>
</div>

<hr />

<h2>Full Width Image Card</h2>

<div class="row">
  <div class="col-12 border rounded shadow p-2">
    <h3>Featured Image</h3>
    <div><img src="https://go.dev/doc/gopher/frontpage.png" alt="featured" /></div>
    <p>A full-width card with <b>border</b>, <b>rounded corners</b>, <b>shadow</b>, and a large image.</p>
  </div>
</div>
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
