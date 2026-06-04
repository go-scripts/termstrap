// Example: shadows — Demonstrates all shadow sizes (sm, normal, lg)
// with various column configurations and the intelligent overflow system.
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

	content := `# Shadow Rendering Demo

## 1. Shadow Sizes Comparison

<div class="row">
  <div class="col-md-4 border rounded shadow-sm p-2 m-1">

### shadow-sm

Light shadow (size 1). Subtle elevation effect using ░ character.

  </div>
  <div class="col-md-4 border rounded shadow p-2 m-1">

### shadow

Standard shadow (size 2). Medium elevation using ░ characters.

  </div>
  <div class="col-md-4 border rounded shadow-lg p-2 m-1">

### shadow-lg

Large shadow (size 3). Strong elevation using ▒ characters.

  </div>
</div>

---

## 2. Full-Width Shadow

<div class="row">
  <div class="col-md-12 border rounded shadow p-3">

### Full-Width Card with Shadow

The shadow system automatically adjusts its size to prevent overflow beyond the terminal width. When content is too wide for the requested shadow size, the shadow shrinks intelligently.

  </div>
</div>

---

## 3. Shadow with Colors

<div class="row">
  <div class="col-md-6 bg-dark text-white border rounded shadow-lg p-2 m-1">

### Dark Card

Shadow adds depth to dark backgrounds.

  </div>
  <div class="col-md-6 bg-primary text-white border rounded shadow p-2 m-1">

### Primary Card

Colors persist through the shadow rendering.

  </div>
</div>

---

## 4. Row-Level Shadow

<div class="row bg-light text-dark p-2 rounded shadow-lg">
  <div class="col-md-6">

**Left Column**

Row-level shadow wraps the entire row.

  </div>
  <div class="col-md-6">

**Right Column**

Both columns share the same shadow.

  </div>
</div>

---

## 5. Shadow None (Reset)

<div class="row">
  <div class="col-md-6 border rounded shadow-lg p-2">

**shadow-lg** — Has shadow

  </div>
  <div class="col-md-6 border rounded shadow-none p-2">

**shadow-none** — No shadow

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
