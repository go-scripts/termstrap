// Example: themes — Demonstrates rendering the same HTML structure
// across different built-in color themes (Bootstrap, Tokyo Night, Dracula).
//
// Usage:
//
//	go run ./examples/themes/
package main

import (
	"fmt"
	"os"

	"github.com/go-scripts/termstrap"
	"golang.org/x/term"
)

func main() {
	width := 80
	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
		width = w
	}

	htmlCard := `<div class="container p-2 border rounded">
  <div class="row">
    <div class="col-md-6 p-2">
      <div class="p-2 bg-primary rounded">
        <h2>Primary Card Header</h2>
        <p>This card demonstrates the primary background and text styling.</p>
        <p class="text-white">Highlighted status: <span class="fw-bold">Active</span></p>
      </div>
    </div>
    <div class="col-md-6 p-2">
      <div class="p-2 bg-dark rounded">
        <h3>Code & Details</h3>
        <p>Inline command: <code>go run ./main.go</code></p>
        <blockquote>Tokyo Night, Dracula, and Bootstrap themes adapt automatically.</blockquote>
        <div class="p-1 bg-secondary rounded mt-1">
          <span class="text-warning">Warning text</span> | <span class="text-success">Success text</span> | <span class="text-info">Info text</span>
        </div>
      </div>
    </div>
  </div>
</div>`

	themes := []struct {
		name  string
		theme termstrap.Theme
	}{
		{name: "Default / Bootstrap Theme", theme: termstrap.ThemeBootstrap},
		{name: "Tokyo Night Theme", theme: termstrap.ThemeTokyoNight},
		{name: "Dracula Theme", theme: termstrap.ThemeDracula},
	}

	for _, t := range themes {
		fmt.Printf("\n=== %s (Theme: %q) ===\n\n", t.name, t.theme)

		m := termstrap.New(
			htmlCard,
			termstrap.WithWidth(width),
			termstrap.WithTheme(t.theme),
		)

		output, err := m.Render()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error rendering %s: %v\n", t.name, err)
			continue
		}

		fmt.Println(output)
	}
}
