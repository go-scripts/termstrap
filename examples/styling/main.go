// Example: styling — Demonstrates utility classes for padding, margin,
// text alignment, colors, and font styles.
//
// Usage:
//
//	go run ./examples/styling/
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

	content := `<h1>Styling Classes Demo</h1>

<h2>1. Padding Variants</h2>

<div class="row">
  <div class="col-md-3 border p-0">
    <div><b>p-0</b> (none)</div>
  </div>
  <div class="col-md-3 border p-1">
    <div><b>p-1</b> (small)</div>
  </div>
  <div class="col-md-3 border p-2">
    <div><b>p-2</b> (medium)</div>
  </div>
  <div class="col-md-3 border p-3">
    <div><b>p-3</b> (large)</div>
  </div>
</div>

<hr />

<h2>2. Directional Padding</h2>

<div class="row">
  <div class="col-md-4 border px-3">
    <div><b>px-3</b> — horizontal padding only</div>
  </div>
  <div class="col-md-4 border py-2">
    <div><b>py-2</b> — vertical padding only</div>
  </div>
  <div class="col-md-4 border pt-2 pb-1 ps-2 pe-1">
    <div><b>pt-2 pb-1 ps-2 pe-1</b> — mixed</div>
  </div>
</div>

<hr />

<h2>3. Margin</h2>

<div class="row">
  <div class="col-md-4 border m-1 p-1">
    <div><b>m-1</b> — small margin</div>
  </div>
  <div class="col-md-4 border m-2 p-1">
    <div><b>m-2</b> — medium margin</div>
  </div>
  <div class="col-md-4 border mx-2 my-1 p-1">
    <div><b>mx-2 my-1</b> — directional</div>
  </div>
</div>

<hr />

<h2>4. Text Alignment</h2>

<div class="row">
  <div class="col-md-4 border p-1 text-start">
    <div><b>text-start</b></div>
    <p>Left aligned text content.</p>
  </div>
  <div class="col-md-4 border p-1 text-center">
    <div><b>text-center</b></div>
    <p>Centered text content.</p>
  </div>
  <div class="col-md-4 border p-1 text-end">
    <div><b>text-end</b></div>
    <p>Right aligned text content.</p>
  </div>
</div>

<hr />

<h2>5. Background &amp; Text Colors</h2>

<div class="row">
  <div class="col-md-3 bg-primary text-white p-2">
    <div><b>Primary</b></div>
  </div>
  <div class="col-md-3 bg-success text-white p-2">
    <div><b>Success</b></div>
  </div>
  <div class="col-md-3 bg-warning text-dark p-2">
    <div><b>Warning</b></div>
  </div>
  <div class="col-md-3 bg-danger text-white p-2">
    <div><b>Danger</b></div>
  </div>
</div>

<div class="row mt-1">
  <div class="col-md-3 bg-info text-dark p-2">
    <div><b>Info</b></div>
  </div>
  <div class="col-md-3 bg-secondary text-white p-2">
    <div><b>Secondary</b></div>
  </div>
  <div class="col-md-3 bg-dark text-white p-2">
    <div><b>Dark</b></div>
  </div>
  <div class="col-md-3 bg-light text-dark p-2">
    <div><b>Light</b></div>
  </div>
</div>

<hr />

<h2>6. Bold Text</h2>

<div class="row">
  <div class="col-md-6 border p-1">
    <div class="fw-bold"><b>fw-bold</b> — Bold text via class</div>
  </div>
  <div class="col-md-6 border p-1">
    <div class="text-bold"><b>text-bold</b> — Also bold text</div>
  </div>
</div>

<hr />

<h2>7. Combined Styling</h2>

<div class="row">
  <div class="col-12 bg-dark text-white p-3 border rounded fw-bold text-center">
    Full-width card with <b>bg-dark</b>, <b>text-white</b>, <b>p-3</b>, <b>border</b>, <b>rounded</b>, <b>fw-bold</b>, and <b>text-center</b> combined.
  </div>
</div>
`

	m := termstrap.Model{
		HTML:          content,
		Width:         width,
		ImageRenderer: termimage.NewRenderer(termimage.WithProtocol(termimage.HalfBlock)),
	}
	output, err := m.Render()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Print(output)
}
