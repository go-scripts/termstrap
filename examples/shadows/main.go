// Example: shadows — Demonstrates shadow rendering with different sizes,
// colors, borders, and auto-clamping to prevent terminal overflow.
//
// Usage:
//
//	go run ./examples/shadows/
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

	content := `<h1>Shadow Rendering Demo</h1>

<h2>1. Shadow Sizes Comparison</h2>

<div class="row">
  <div class="col-md-4 border rounded shadow-sm p-2 m-1">
    <h3>shadow-sm</h3>
    <p>Light shadow (size 1). Subtle elevation effect.</p>
  </div>
  <div class="col-md-4 border rounded shadow p-2 m-1">
    <h3>shadow</h3>
    <p>Standard shadow (size 2). Medium elevation.</p>
  </div>
  <div class="col-md-4 border rounded shadow-lg p-2 m-1">
    <h3>shadow-lg</h3>
    <p>Large shadow (size 3). Strong elevation.</p>
  </div>
</div>

<hr />

<h2>2. Full-Width Shadow</h2>

<div class="row">
  <div class="col-12 border rounded shadow p-3">
    <h3>Full-Width Card with Shadow</h3>
    <p>The shadow system automatically adjusts its size to prevent overflow beyond the terminal width.</p>
  </div>
</div>

<hr />

<h2>3. Shadow with Colors</h2>

<div class="row">
  <div class="col-md-6 bg-dark text-white border rounded shadow-lg p-2 m-1">
    <h3>Dark Card</h3>
    <p>Shadow adds depth to dark backgrounds.</p>
  </div>
  <div class="col-md-6 bg-primary text-white border rounded shadow p-2 m-1">
    <h3>Primary Card</h3>
    <p>Colors persist through the shadow rendering.</p>
  </div>
</div>

<hr />

<h2>4. Row-Level Shadow</h2>

<div class="row bg-light text-dark p-2 rounded shadow-lg">
  <div class="col-md-6">
    <div><b>Left Column</b></div>
    <p>Row-level shadow wraps the entire row.</p>
  </div>
  <div class="col-md-6">
    <div><b>Right Column</b></div>
    <p>Both columns share the same shadow.</p>
  </div>
</div>

<hr />

<h2>5. Shadow None (Reset)</h2>

<div class="row">
  <div class="col-md-6 border rounded shadow-lg p-2">
    <div><b>shadow-lg</b> — Has shadow</div>
  </div>
  <div class="col-md-6 border rounded shadow-none p-2">
    <div><b>shadow-none</b> — No shadow</div>
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
