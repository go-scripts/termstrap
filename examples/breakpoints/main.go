// Example: breakpoints — Demonstrates the responsive 12-column grid
// system across different terminal widths (xs, sm, md, lg, xl).
//
// Shows how columns stack vertically on narrow terminals and arrange
// horizontally on wider ones using Bootstrap-style breakpoint classes.
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
	content := `# Responsive Breakpoints Demo

## Breakpoint Thresholds

| Breakpoint | Prefix     | Terminal Width |
|------------|------------|----------------|
| xs         | col-       | < 60 cols      |
| sm         | col-sm-    | >= 60 cols     |
| md         | col-md-    | >= 80 cols     |
| lg         | col-lg-    | >= 120 cols    |
| xl         | col-xl-    | >= 160 cols    |

---

## col-md-6 — Stacks below 80 cols

<div class="row">
  <div class="col-md-6 border p-1">

**Left (col-md-6)**

Visible side-by-side at 80+ cols.

  </div>
  <div class="col-md-6 border p-1">

**Right (col-md-6)**

Stacks vertically below 80 cols.

  </div>
</div>

---

## col-lg-4 — Stacks below 120 cols

<div class="row">
  <div class="col-lg-4 border p-1">

**A (col-lg-4)**

  </div>
  <div class="col-lg-4 border p-1">

**B (col-lg-4)**

  </div>
  <div class="col-lg-4 border p-1">

**C (col-lg-4)**

  </div>
</div>

---

## col-sm-6 — Side-by-side from 60 cols

<div class="row">
  <div class="col-sm-6 border p-1">

**Left (col-sm-6)**

Side-by-side starting at 60 cols.

  </div>
  <div class="col-sm-6 border p-1">

**Right (col-sm-6)**

  </div>
</div>

---

## Mixed breakpoints — col-sm-12 / col-md-6

<div class="row">
  <div class="col-sm-12 col-md-6 border p-1">

**Panel A**

Full width on small, half on medium+.

  </div>
  <div class="col-sm-12 col-md-6 border p-1">

**Panel B**

Full width on small, half on medium+.

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
		fmt.Printf("\n%s\n%s\n\n", w.name, "════════════════════════════════════════")

		m := termstrap.Model{
			Content:       content,
			Width:         w.width,
			ImageRenderer: termimage.NewRenderer(termimage.WithProtocol(termimage.HalfBlock)),
		}
		output, err := m.Render()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error at %s: %v\n", w.name, err)
			continue
		}
		fmt.Print(output)
	}
}
