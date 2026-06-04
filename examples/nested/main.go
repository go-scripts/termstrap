// Example: nested — Demonstrates nested grid rows inside columns,
// showing recursive layout rendering.
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
	width := 120
	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
		width = w
	}

	content := `# Nested Grid Demo

## 1. Basic Nested Row

Outer row with 2 columns. The left column contains a nested row with 2 sub-columns:

<div class="row">
  <div class="col-md-6 border p-1">

### Outer Left (col-md-6)

<div class="row">
  <div class="col-md-6 border p-1">

**Inner Left**

Nested col-md-6

  </div>
  <div class="col-md-6 border p-1">

**Inner Right**

Nested col-md-6

  </div>
</div>

  </div>
  <div class="col-md-6 border p-1">

### Outer Right (col-md-6)

This column has no nested layout. Just plain markdown content.

- Item one
- Item two
- Item three

  </div>
</div>

---

## 2. Deep Nesting (3 Levels)

<div class="row">
  <div class="col-md-12 border rounded p-1">

### Level 1 (col-md-12)

<div class="row">
  <div class="col-md-8 border p-1">

**Level 2 (col-md-8)**

<div class="row">
  <div class="col-md-6 bg-dark text-white p-1">

**Level 3a**

  </div>
  <div class="col-md-6 bg-secondary text-white p-1">

**Level 3b**

  </div>
</div>

  </div>
  <div class="col-md-4 bg-light text-dark p-1">

**Level 2 (col-md-4)**

Sidebar content.

  </div>
</div>

  </div>
</div>

---

## 3. Nested with Styled Containers

<div class="row">
  <div class="col-md-4 bg-primary text-white p-2 rounded">

### Navigation

<div class="row">
  <div class="col-md-12 p-1">

- Home
- About
- Contact

  </div>
</div>

  </div>
  <div class="col-md-8 border p-1">

### Content Area

<div class="row">
  <div class="col-md-6 p-1">

**Article 1**

Lorem ipsum dolor sit amet.

  </div>
  <div class="col-md-6 p-1">

**Article 2**

Consectetur adipiscing elit.

  </div>
</div>

  </div>
</div>
`

	m := termstrap.Model{
		Content:       content,
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
