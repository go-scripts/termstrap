// Example: borders — Demonstrates all border variants:
// full border, partial borders (top, bottom, left, right),
// rounded corners, and combinations.
//
// Usage:
//
//	go run ./examples/borders/
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

	content := `<h1>Border Variants Demo</h1>

<h2>1. Full Borders</h2>

<div class="row">
  <div class="col-md-6 border p-1">
    <div><b>border</b> — Normal border</div>
    <p>All four sides with square corners.</p>
  </div>
  <div class="col-md-6 border rounded p-1">
    <div><b>border rounded</b> — Rounded corners</div>
    <p>All four sides with rounded corners.</p>
  </div>
</div>

<hr />

<h2>2. Partial Borders — Single Side</h2>

<div class="row">
  <div class="col-md-3 border-top p-1">
    <div><b>border-top</b></div>
    <p>Top only</p>
  </div>
  <div class="col-md-3 border-bottom p-1">
    <div><b>border-bottom</b></div>
    <p>Bottom only</p>
  </div>
  <div class="col-md-3 border-left p-1">
    <div><b>border-left</b></div>
    <p>Left only</p>
  </div>
  <div class="col-md-3 border-right p-1">
    <div><b>border-right</b></div>
    <p>Right only</p>
  </div>
</div>

<hr />

<h2>3. Partial Borders — Combinations</h2>

<div class="row">
  <div class="col-md-4 border-top border-bottom p-1">
    <div><b>Top + Bottom</b></div>
    <p>Horizontal lines only.</p>
  </div>
  <div class="col-md-4 border-left border-right p-1">
    <div><b>Left + Right</b></div>
    <p>Vertical sides only.</p>
  </div>
  <div class="col-md-4 border-top border-left border-right p-1">
    <div><b>Top + Left + Right</b></div>
    <p>Open bottom.</p>
  </div>
</div>

<hr />

<h2>4. Borders with Colors</h2>

<div class="row">
  <div class="col-md-4 border rounded bg-dark text-white p-2">
    <div><b>Dark with border</b></div>
    <p>Border visible on dark background.</p>
  </div>
  <div class="col-md-4 border-left bg-primary text-white p-2">
    <div><b>Primary with left border</b></div>
    <p>Accent line on the left.</p>
  </div>
  <div class="col-md-4 border-bottom bg-success text-white p-2">
    <div><b>Success with bottom border</b></div>
    <p>Underline accent.</p>
  </div>
</div>

<hr />

<h2>5. Borders with Shadows</h2>

<div class="row">
  <div class="col-md-6 border rounded shadow-sm p-2">
    <div><b>border + shadow-sm</b></div>
    <p>Light shadow beneath the border.</p>
  </div>
  <div class="col-md-6 border rounded shadow-lg p-2">
    <div><b>border + shadow-lg</b></div>
    <p>Strong shadow beneath the border.</p>
  </div>
</div>
`

	m := termstrap.New(
		content,
		termstrap.WithWidth(width),
		termstrap.WithImageRenderer(termimage.NewRenderer(termimage.WithProtocol(termimage.HalfBlock))),
	)
	output, err := m.Render()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Print(output)
}
