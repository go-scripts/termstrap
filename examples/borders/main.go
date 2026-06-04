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

	content := `# Border Variants Demo

## 1. Full Borders

<div class="row">
  <div class="col-md-6 border p-1">

**border** — Normal border

All four sides with square corners.

  </div>
  <div class="col-md-6 border rounded p-1">

**border rounded** — Rounded corners

All four sides with rounded corners.

  </div>
</div>

---

## 2. Partial Borders — Single Side

<div class="row">
  <div class="col-md-3 border-top p-1">

**border-top**

Top only

  </div>
  <div class="col-md-3 border-bottom p-1">

**border-bottom**

Bottom only

  </div>
  <div class="col-md-3 border-left p-1">

**border-left**

Left only

  </div>
  <div class="col-md-3 border-right p-1">

**border-right**

Right only

  </div>
</div>

---

## 3. Partial Borders — Combinations

<div class="row">
  <div class="col-md-4 border-top border-bottom p-1">

**Top + Bottom**

Horizontal lines only.

  </div>
  <div class="col-md-4 border-left border-right p-1">

**Left + Right**

Vertical sides only.

  </div>
  <div class="col-md-4 border-top border-left border-right p-1">

**Top + Left + Right**

Open bottom.

  </div>
</div>

---

## 4. Borders with Colors

<div class="row">
  <div class="col-md-4 border rounded bg-dark text-white p-2">

**Dark with border**

Border visible on dark background.

  </div>
  <div class="col-md-4 border-left bg-primary text-white p-2">

**Primary with left border**

Accent line on the left.

  </div>
  <div class="col-md-4 border-bottom bg-success text-white p-2">

**Success with bottom border**

Underline accent.

  </div>
</div>

---

## 5. Borders with Shadows

<div class="row">
  <div class="col-md-6 border rounded shadow-sm p-2">

**border + shadow-sm**

Light shadow beneath the border.

  </div>
  <div class="col-md-6 border rounded shadow-lg p-2">

**border + shadow-lg**

Strong shadow beneath the border.

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
