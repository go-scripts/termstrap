// Example: nested-no-md — Demonstrates nested grid rows with plain col-* classes
// (no breakpoint suffix like -md) to ensure content displays at all screen sizes.
// Columns stack vertically on narrow screens, showing all content.
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
	caps := termimage.Detect()
	width := 80
	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
		width = w
	}

	fmt.Printf("Protocol: %s | Width: %d\n\n", caps.Protocol, width)

	content := `# Nested Rows with Plain col-* Classes

## Layout: row → col-5 (contains nested row with col-6 / col-6) → col-3 → col-4

All content displays regardless of breakpoint, stacking vertically when needed.

<div class="row">
  <div class="col-5 border p-1 bg-light">

### Container Left (col-5)

<div class="row">
  <div class="col-6 border p-1">

**Nested Left**

Content in nested col-6

  </div>
  <div class="col-6 border p-1">

**Nested Right**

Another nested col-6

  </div>
</div>

  </div>
  <div class="col-3 border p-1 bg-secondary text-white">

### Middle (col-3)

Right-side content area.

  </div>
  <div class="col-4 border p-1">

### Right (col-4)

Final column content.

  </div>
</div>

---

## Deeper Nesting (3 levels)

<div class="row">
  <div class="col-12 border rounded p-1 bg-primary text-white">

### Level 1 Full Width

<div class="row">
  <div class="col-6 border p-1 bg-dark">

**Level 2 Left (col-6)**

<div class="row">
  <div class="col-6 border p-1 bg-info">

Level 3a

  </div>
  <div class="col-6 border p-1 bg-warning">

Level 3b

  </div>
</div>

  </div>
  <div class="col-6 border p-1 bg-success text-white">

**Level 2 Right (col-6)**

No further nesting here.

  </div>
</div>

  </div>
</div>

---

Text after nested layout.
`

	m := termstrap.Model{
		Content:  content,
		Width:    width,
		RootPath: ".",
	}
	out, err := m.Render()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Render error: %v\n", err)
		os.Exit(1)
	}
	fmt.Print(out)
}
