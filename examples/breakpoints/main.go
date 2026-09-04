// Example: breakpoints — Demonstrates responsive grid behavior across
// different terminal widths using Bootstrap-like breakpoint classes.
//
// Usage:
//
//	go run ./examples/breakpoints/
package main

import (
	"fmt"
	"os"

	"github.com/go-scripts/termstrap"
	termimage "github.com/go-scripts/termstrap/image"
)

func main() {
	content := `<h1>Responsive Breakpoints Demo</h1>

<h2>Breakpoint Thresholds</h2>

<table>
  <tr><th>Breakpoint</th><th>Prefix</th><th>Terminal Width</th></tr>
  <tr><td>xs</td><td>col-</td><td>&lt; 60 cols</td></tr>
  <tr><td>sm</td><td>col-sm-</td><td>&gt;= 60 cols</td></tr>
  <tr><td>md</td><td>col-md-</td><td>&gt;= 80 cols</td></tr>
  <tr><td>lg</td><td>col-lg-</td><td>&gt;= 120 cols</td></tr>
  <tr><td>xl</td><td>col-xl-</td><td>&gt;= 160 cols</td></tr>
</table>

<hr />

<h2>col-md-6 — Stacks below 80 cols</h2>

<div class="row">
  <div class="col-md-6 border p-1">
    <div><b>Left (col-md-6)</b></div>
    <p>Visible side-by-side at 80+ cols.</p>
  </div>
  <div class="col-md-6 border p-1">
    <div><b>Right (col-md-6)</b></div>
    <p>Stacks vertically below 80 cols.</p>
  </div>
</div>

<hr />

<h2>col-lg-4 — Stacks below 120 cols</h2>

<div class="row">
  <div class="col-lg-4 border p-1">
    <div><b>A (col-lg-4)</b></div>
  </div>
  <div class="col-lg-4 border p-1">
    <div><b>B (col-lg-4)</b></div>
  </div>
  <div class="col-lg-4 border p-1">
    <div><b>C (col-lg-4)</b></div>
  </div>
</div>

<hr />

<h2>col-sm-6 — Side-by-side from 60 cols</h2>

<div class="row">
  <div class="col-sm-6 border p-1">
    <div><b>Left (col-sm-6)</b></div>
    <p>Side-by-side starting at 60 cols.</p>
  </div>
  <div class="col-sm-6 border p-1">
    <div><b>Right (col-sm-6)</b></div>
  </div>
</div>

<hr />

<h2>Mixed breakpoints — col-sm-12 / col-md-6</h2>

<div class="row">
  <div class="col-sm-12 col-md-6 border p-1">
    <div><b>Panel A</b></div>
    <p>Full width on small, half on medium+.</p>
  </div>
  <div class="col-sm-12 col-md-6 border p-1">
    <div><b>Panel B</b></div>
    <p>Full width on small, half on medium+.</p>
  </div>
</div>
`

	widths := []struct {
		name  string
		width int
	}{
		{"XS (50 cols)", 50},
		{"SM (70 cols)", 70},
		{"MD (100 cols)", 100},
		{"LG (130 cols)", 130},
	}

	for _, w := range widths {
		fmt.Printf("\n\033[1;36m%s\033[0m\n", w.name)
		fmt.Println("════════════════════════════════════════")
		m := termstrap.Model{
			HTML:          content,
			Width:         w.width,
			ImageRenderer: termimage.NewRenderer(termimage.WithProtocol(termimage.HalfBlock)),
		}
		output, err := m.Render()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error rendering at width %d: %v\n", w.width, err)
			continue
		}
		fmt.Print(output)
	}
}
