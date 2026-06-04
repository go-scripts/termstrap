// Example: markdown — Images embedded in pure markdown content
// (no HTML grid). Tests inline image rendering with various sizes
// and mixed markdown features.
//
// Usage:
//
//	go run ./examples/image/markdown/
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

	content := `# Markdown Inline Images

## Default Width (half terminal)

Below is an image at default width (no size specified):

![gopher](https://go.dev/doc/gopher/frontpage.png)

---

## Explicit Width — Small (20 cols)

![small](https://go.dev/doc/gopher/frontpage.png =20)

## Explicit Width — Medium (40 cols)

![medium](https://go.dev/doc/gopher/frontpage.png =40)

## Explicit Width — Large (60 cols)

![large](https://go.dev/doc/gopher/frontpage.png =60)

---

## Extended Attribute Syntax

![extended](https://go.dev/doc/gopher/frontpage.png, width=40, class=rounded, title='Example')

This image uses the new comma-separated attribute syntax, where "width" is parsed
and other key/value pairs are preserved for later use.

---

## Image Between Text

Some text **before** the image. This paragraph demonstrates that images
can be inlined within flowing markdown content.

![inline](https://go.dev/doc/gopher/pkg.png =30)

Some text **after** the image. The rendering pipeline replaces the
image placeholder while preserving ANSI codes from glamour.

---

## Multiple Images in Sequence

![first](https://go.dev/doc/gopher/frontpage.png =25)

![second](https://www.gstatic.com/webp/gallery/1.webp =25)

---

## Image in a List

- Item with no image
- Item with image: ![list-img](https://go.dev/doc/gopher/frontpage.png =15)
- Another item after

## Image in a Blockquote

> Here is a quoted section with an image:
>
> ![quoted](https://go.dev/doc/gopher/frontpage.png =20)
>
> The image should render inside the quote block.
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
