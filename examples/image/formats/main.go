// Example: formats — Tests rendering of different image formats
// (PNG, JPEG, WebP, GIF) through the image pipeline.
//
// Verifies that all registered decoders work correctly with each protocol.
//
// Usage:
//
//	go run ./examples/image/formats/
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

	fmt.Printf("Protocol: %s | Terminal: %dx%d\n\n", caps.Protocol, caps.ColCount, caps.RowCount)

	content := `# Image Format Test

## PNG (lossless, transparency support)

![png](https://go.dev/doc/gopher/frontpage.png =40)

---

## JPEG (lossy, photographs)

![jpeg](https://picsum.photos/id/237/200/300.jpg =40)

---

## WebP (modern format, lossy & lossless)

![webp](https://www.gstatic.com/webp/gallery/1.webp =40)

---

## GIF (animated / legacy)

![gif](https://go.dev/doc/gopher/pkg.png =30)

---

All four formats above should render correctly. If a format fails,
a warning is logged to stderr and the image is silently skipped.
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
