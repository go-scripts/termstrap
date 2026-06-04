// Example: styling — Demonstrates all CSS-like styling classes supported
// by termstrap: padding, margin, text alignment, borders, colors, and bold.
//
// Usage:
//
//	go run ./examples/styling/
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

	content := `# Styling Classes Demo

## 1. Padding Variants

<div class="row">
  <div class="col-md-3 border p-0">

**p-0** (none)

  </div>
  <div class="col-md-3 border p-1">

**p-1** (small)

  </div>
  <div class="col-md-3 border p-2">

**p-2** (medium)

  </div>
  <div class="col-md-3 border p-3">

**p-3** (large)

  </div>
</div>

---

## 2. Directional Padding

<div class="row">
  <div class="col-md-4 border px-3">

**px-3** — horizontal padding only

  </div>
  <div class="col-md-4 border py-2">

**py-2** — vertical padding only

  </div>
  <div class="col-md-4 border pt-2 pb-1 ps-2 pe-1">

**pt-2 pb-1 ps-2 pe-1** — mixed

  </div>
</div>

---

## 3. Margin

<div class="row">
  <div class="col-md-4 border m-1 p-1">

**m-1** — small margin

  </div>
  <div class="col-md-4 border m-2 p-1">

**m-2** — medium margin

  </div>
  <div class="col-md-4 border mx-2 my-1 p-1">

**mx-2 my-1** — directional

  </div>
</div>

---

## 4. Text Alignment

<div class="row">
  <div class="col-md-4 border p-1 text-start">

**text-start**

Left aligned text content.

  </div>
  <div class="col-md-4 border p-1 text-center">

**text-center**

Centered text content.

  </div>
  <div class="col-md-4 border p-1 text-end">

**text-end**

Right aligned text content.

  </div>
</div>

---

## 5. Background & Text Colors

<div class="row">
  <div class="col-md-3 bg-primary text-white p-2">

**Primary**

  </div>
  <div class="col-md-3 bg-success text-white p-2">

**Success**

  </div>
  <div class="col-md-3 bg-warning text-dark p-2">

**Warning**

  </div>
  <div class="col-md-3 bg-danger text-white p-2">

**Danger**

  </div>
</div>

<div class="row">
  <div class="col-md-3 bg-info text-dark p-2">

**Info**

  </div>
  <div class="col-md-3 bg-secondary text-white p-2">

**Secondary**

  </div>
  <div class="col-md-3 bg-dark text-light p-2">

**Dark**

  </div>
  <div class="col-md-3 bg-light text-dark p-2">

**Light**

  </div>
</div>

---

## 6. Bold Text

<div class="row">
  <div class="col-md-6 border p-1 fw-bold">

**fw-bold** — Bold text via class

  </div>
  <div class="col-md-6 border p-1 text-bold">

**text-bold** — Also bold text

  </div>
</div>

---

## 7. Combined Styling

<div class="row">
  <div class="col-md-12 bg-dark text-white p-3 border rounded fw-bold text-center">

Full-width card with **bg-dark**, **text-white**, **p-3**, **border**, **rounded**, **fw-bold**, and **text-center** combined.

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
