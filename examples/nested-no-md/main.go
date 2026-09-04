// Example: nested-no-md — Demonstrates deeply nested grid rows using
// plain col-* classes.
//
// Usage:
//
//	go run ./examples/nested-no-md/
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

	content := `<h1>Nested Rows with Plain col-* Classes</h1>

<h2>Layout: row → col-5 (contains nested row with col-6 / col-6) → col-3 → col-4</h2>

<p>All content displays regardless of breakpoint, stacking vertically when needed.</p>

<div class="row">
  <div class="col-5 border p-1 bg-light">
    <h3>Container Left (col-5)</h3>
    <div class="row">
      <div class="col-6 border p-1">
        <div><b>Nested Left</b></div>
        <p>Content in nested col-6</p>
      </div>
      <div class="col-6 border p-1">
        <div><b>Nested Right</b></div>
        <p>Another nested col-6</p>
      </div>
    </div>
  </div>
  <div class="col-3 border p-1 bg-secondary text-white">
    <h3>Middle (col-3)</h3>
    <p>Right-side content area.</p>
  </div>
  <div class="col-4 border p-1">
    <h3>Right (col-4)</h3>
    <p>Final column content.</p>
  </div>
</div>

<hr />

<h2>Deeper Nesting (3 levels)</h2>

<div class="row">
  <div class="col-12 border rounded p-1 bg-primary text-white">
    <h3>Level 1 Full Width</h3>
    <div class="row">
      <div class="col-6 border p-1 bg-dark">
        <div><b>Level 2 Left (col-6)</b></div>
        <div class="row">
          <div class="col-6 border p-1 bg-info">
            <div>Level 3a</div>
          </div>
          <div class="col-6 border p-1 bg-warning">
            <div>Level 3b</div>
          </div>
        </div>
      </div>
      <div class="col-6 border p-1 bg-success text-white">
        <div><b>Level 2 Right (col-6)</b></div>
        <p>No further nesting here.</p>
      </div>
    </div>
  </div>
</div>

<hr />

<p>Text after nested layout.</p>
`

	m := termstrap.Model{
		HTML:          content,
		Width:         width,
		RootPath:      ".",
		ImageRenderer: termimage.NewRenderer(termimage.WithProtocol(termimage.HalfBlock)),
	}
	output, err := m.Render()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Print(output)
}
