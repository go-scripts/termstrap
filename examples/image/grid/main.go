// Example: grid — Images inside Bootstrap-like grid layouts with
// borders, shadows, backgrounds, and text alignment.
//
// Usage:
//
//	go run ./examples/image/grid/
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

	content := `# Grid + Images + Styling

## Card Layout — Image + Metadata

Image in a standalone bordered card, metadata below:

<div class="row">
  <div class="col-md-12 border rounded shadow-sm p-1">

![poster](https://go.dev/doc/gopher/frontpage.png =40)

### The Go Gopher

**Rob Pike & Renée French** — 2009

| Info       | Value              |
|------------|--------------------|
| Language   | Go                 |
| Mascot     | Gopher             |
| Created    | 2009               |
| License    | Creative Commons   |

> *"Don't communicate by sharing memory; share memory by communicating."*

  </div>
</div>

---

## Gallery — Three Images with Labels

<div class="row">
  <div class="col-md-4 border rounded p-1 text-center">

![img1](https://go.dev/doc/gopher/pkg.png)

**PNG Format**

  </div>
  <div class="col-md-4 border rounded p-1 text-center">

![img2](https://go.dev/doc/gopher/frontpage.png)

**Another PNG**

  </div>
  <div class="col-md-4 border rounded shadow-lg p-1 text-center">

![img3](https://www.gstatic.com/webp/gallery/1.webp)

**WebP + Shadow**

  </div>
</div>

---

## Styled Text Columns — Colors & Backgrounds

<div class="row">
  <div class="col-md-4 bg-primary text-white p-2 border rounded">

**Primary** — Blue background with white text for important content.

  </div>
  <div class="col-md-4 bg-dark text-white p-2 border rounded">

**Dark** — Dark background for contrast sections and emphasis.

  </div>
  <div class="col-md-4 bg-success text-white p-2 border rounded">

**Success** — Green background for positive messages and confirmations.

  </div>
</div>

---

## Full Width Image Card

<div class="row">
  <div class="col-md-12 border rounded shadow p-2">

### Featured Image

![wide](https://go.dev/doc/gopher/frontpage.png =60)

A full-width card with **border**, **rounded corners**, **shadow**, and a wide image.

  </div>
</div>

---

## Stacked on Small Terminals

Resize your terminal below 80 columns to see the grid collapse to vertical stacking.

<div class="row">
  <div class="col-md-6 border rounded p-1">

![left](https://go.dev/doc/gopher/frontpage.png =30)

Left column with image.

  </div>
  <div class="col-md-6 border rounded p-1 bg-info text-white">

### Right Column

This column has an **info background**. On narrow terminals, it stacks below the image.

  </div>
</div>
`

	m := termstrap.Model{
		Content: content,
		Width:   width,
	}
	output, err := m.Render()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Print(output)
}
