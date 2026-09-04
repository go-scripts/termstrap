// Example: nested — Demonstrates nested grid layouts (rows inside columns)
// with multiple nesting levels and styled containers.
//
// Usage:
//
//	go run ./examples/nested/
package main

import (
	"fmt"
	"os"

	"github.com/go-scripts/termstrap"
	termimage "github.com/go-scripts/termstrap/image"
	"golang.org/x/term"
)

func main() {
	width := 100
	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
		width = w
	}

	content := `<h1>Nested Grid Demo</h1>

<h2>1. Basic Nested Row</h2>

<p>Outer row with 2 columns. The left column contains a nested row with 2 sub-columns:</p>

<div class="row">
  <div class="col-md-6 border p-1">
    <h3>Outer Left (col-md-6)</h3>
    <div class="row">
      <div class="col-md-6 border p-1">
        <div><b>Inner Left</b></div>
        <p>Nested col-md-6</p>
      </div>
      <div class="col-md-6 border p-1">
        <div><b>Inner Right</b></div>
        <p>Nested col-md-6</p>
      </div>
    </div>
  </div>
  <div class="col-md-6 border p-1">
    <h3>Outer Right (col-md-6)</h3>
    <p>This column has no nested layout. Just plain content.</p>
    <ul>
      <li>Item one</li>
      <li>Item two</li>
      <li>Item three</li>
    </ul>
  </div>
</div>

<hr />

<h2>2. Deep Nesting (3 Levels)</h2>

<div class="row">
  <div class="col-md-12 border rounded p-1">
    <h3>Level 1 (col-md-12)</h3>
    <div class="row">
      <div class="col-md-8 border p-1">
        <div><b>Level 2 (col-md-8)</b></div>
        <div class="row">
          <div class="col-md-6 bg-dark text-white p-1">
            <div><b>Level 3a</b></div>
          </div>
          <div class="col-md-6 bg-secondary text-white p-1">
            <div><b>Level 3b</b></div>
          </div>
        </div>
      </div>
      <div class="col-md-4 bg-light text-dark p-1">
        <div><b>Level 2 (col-md-4)</b></div>
        <p>Sidebar content.</p>
      </div>
    </div>
  </div>
</div>

<hr />

<h2>3. Nested with Styled Containers</h2>

<div class="row">
  <div class="col-md-4 bg-primary text-white p-2 rounded">
    <h3>Navigation</h3>
    <div class="row">
      <div class="col-md-12 p-1">
        <div>- Home</div>
        <div>- About</div>
        <div>- Contact</div>
      </div>
    </div>
  </div>
  <div class="col-md-8 border p-1">
    <h3>Content Area</h3>
    <div class="row">
      <div class="col-md-6 p-1">
        <div><b>Article 1</b></div>
        <p>Lorem ipsum dolor sit amet.</p>
      </div>
      <div class="col-md-6 p-1">
        <div><b>Article 2</b></div>
        <p>Consectetur adipiscing elit.</p>
      </div>
    </div>
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
