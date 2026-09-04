# Termstrap 🚀

[![Go Reference](https://pkg.go.dev/badge/github.com/go-scripts/termstrap.svg)](https://pkg.go.dev/github.com/go-scripts/termstrap)
[![Go Report Card](https://goreportcard.com/badge/github.com/go-scripts/termstrap)](https://goreportcard.com/report/github.com/go-scripts/termstrap)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

**Termstrap** is a modern, pure **HTML/CSS terminal rendering engine** written in Go. It allows you to build rich, responsive command-line and TUI interfaces using standard HTML markup, a built-in Bootstrap CSS utility framework, custom stylesheets, and multi-protocol terminal graphics (**Kitty**, **iTerm2**, **Sixel**, and **Half-Block Unicode**).

> 📖 **Full Feature Reference**: See [FEATURES.md](FEATURES.md) for the complete and up-to-date list of all supported HTML tags, Bootstrap classes, breakpoints, box models, and CSS capabilities.

---

## ✨ Highlights

* 📐 **Responsive Bootstrap Grid**: Full 12-column grid (`.row`, `.col-*`, `.col-md-*`), auto-columns, nested multi-tier rows, and automatic responsive stacking based on terminal width.
* 🎨 **Universal Box Model & Utility Classes**: Margins (`m-0` to `m-5`), padding (`p-0` to `p-5`), borders (`border`, `rounded`), text alignment (`text-center`, `text-end`), box shadows (`shadow-sm`, `shadow-lg`), and TrueColor backgrounds/foregrounds (`bg-primary`, `text-success`) applicable to **any HTML tag**.
* 🌐 **Pure HTML Support & Custom CSS**: Native rendering for `<table>`, `<h1>`-`<h6>`, `<blockquote>`, `<pre>`, `<code>`, `<hr>`, and external stylesheets via `WithStylesheets(...)`.
* 🖼️ **Terminal Graphics Engine**:
  * Auto-detection & fallback: **Kitty Graphics**, **iTerm2 Inline Images**, **Sixel**, and universal **Half-Block Unicode (`▄`/`▀`)**.
  * **Color Modes**: TrueColor (24-bit), ANSI 256 colors, and ANSI 16 colors with palette quantization.
  * **ANSI Sequence Optimization**: Deduplicates contiguous escape sequences for compact output and maximum throughput.
  * **In-Memory Caching**: Thread-safe decoded image and ANSI render cache with configurable TTL and eviction policies.
* 🫧 **Bubble Tea TUI Ready**: Dedicated `bubbletea/` adapter providing out-of-band graphic sequence handling without cell-buffer corruption.
---

## 🧭 Philosophy & Architecture

Building multi-column, styled terminal layouts has traditionally been tedious—requiring manual line-splitting, manual cursor calculations, and complex column-joining algorithms. Adding terminal images often breaks layout calculations because ANSI cursor escape sequences corrupt cell-based text joiners like Lipgloss.

**Termstrap bridges this gap by decoupling layout composition from image rendering:**

1. **Declarative Layouts**: Write intuitive HTML grid markup (`<div class="row">...</div>`) and Markdown content.
2. **Compositional Pipeline**: Termstrap calculates column dimensions, equalizes column heights, formats Markdown via Glamour, applies Lipgloss styles (borders, padding, shadows), and aligns text without color bleeding.
3. **Deferred Graphics Overlays**: For native protocols (iTerm2, Kitty, Sixel), image placeholders preserve exact spatial dimensions in text buffers while actual graphics sequences are calculated compositionally and emitted as positioned overlays once horizontal assembly is complete.

---

## 📦 Installation

```bash
go get github.com/go-scripts/termstrap
```

Requires **Go 1.21+**.

---

## 🚀 Quick Start

### Basic Markdown & Grid Layout

```go
package main

import (
	"fmt"
	"os"

	"github.com/go-scripts/termstrap"
)

func main() {
	content := `
# Welcome to Termstrap

<div class="row">
  <div class="col-md-6 border rounded p-1 bg-dark text-light">
    ### 📦 Left Column
    This column is wrapped in a rounded border with padding.
    * Item 1
    * Item 2
  </div>
  <div class="col-md-6 border rounded p-1 text-center">
    ### 🎯 Right Column
    Content is **centered** using the ` + "`text-center`" + ` class!
  </div>
</div>
`

	m := termstrap.Model{
		Content: content,
		Width:   80, // Terminal width in columns
	}

	output, err := m.Render()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Render error: %v\n", err)
		return
	}
	fmt.Print(output)
}
```

---

## 🖼️ Working with Images

Termstrap supports both Markdown image syntax (`![alt](src)`) and HTML `<img>` tags, allowing local files and remote HTTP/HTTPS URLs.

### Markdown Image Sizing and Options

```markdown
<!-- Fit image to 20 columns width -->
![My Image](assets/logo.png =20)

<!-- Fit image within 30 columns width and 10 rows height -->
![My Image](assets/logo.png =30x10)

<!-- Custom attributes: color mode & sequence optimization -->
![My Image](https://example.com/photo.jpg =25 {color=ansi256 termstrap-no-optimize=true})
```

### HTML `<img>` Tag Sizing and Classes

```html
<img src="assets/banner.png" width="40" height="15" class="ansi-256 no-opt rounded" alt="Banner" />
```

### Supported Image Classes & Attributes

| Attribute / Class | Description |
| :--- | :--- |
| `=W` or `width="W"` | Constrains image to `W` terminal columns (maintains aspect ratio). |
| `=WxH` or `width="W" height="H"` | Constrains image within `W` columns and `H` terminal character rows. |
| `color=ansi256` / `class="ansi-256"` | Forces image rendering using the 256-color ANSI palette. |
| `color=ansi16` / `class="ansi-16"` | Forces image rendering using the 16-color ANSI palette. |
| `termstrap-no-optimize=true` / `class="no-opt"` | Disables consecutive ANSI escape sequence deduplication. |

---

## 🎨 Supported Bootstrap Classes

Termstrap implements a subset of Bootstrap 5 CSS classes tailored for terminal grids:

### Grid & Responsive Breakpoints
* `.row`: Defines a flex row container.
* `.col`: Auto-width column (distributes remaining space evenly).
* `.col-1` to `.col-12`: Fixed column span out of 12.
* `.col-xs-*`, `.col-sm-*`, `.col-md-*`, `.col-lg-*`, `.col-xl-*`, `.col-xxl-*`: Responsive column spans based on terminal width breakpoints.

| Breakpoint | Terminal Width (Cols) |
| :--- | :--- |
| `xs` | `< 576` |
| `sm` | `≥ 576` |
| `md` | `≥ 768` |
| `lg` | `≥ 992` |
| `xl` | `≥ 1200` |
| `xxl` | `≥ 1400` |

### Spacing (Padding & Margin)
* **Padding**: `p-0` through `p-5`, `px-*`, `py-*`, `pt-*`, `pb-*`, `ps-*`, `pe-*`
* **Margin**: `m-0` through `m-3`, `mx-*`, `my-*`, `mt-*`, `mb-*`

### Borders & Shadows
* **Borders**: `border`, `border-top`, `border-bottom`, `border-left`, `border-right`, `rounded` (rounded corners).
* **Shadows**: `shadow-sm`, `shadow`, `shadow-lg`, `shadow-none` (intelligent drop shadows that adjust to available width).

### Typography & Alignment
* **Text Alignment**: `text-center`, `text-start` (`text-left`), `text-end` (`text-right`).
* **Font Weight**: `fw-bold`, `text-bold`.

### Colors & Themes
* **Backgrounds**: `bg-primary`, `bg-secondary`, `bg-success`, `bg-danger`, `bg-warning`, `bg-info`, `bg-light`, `bg-dark`, `bg-white`, `bg-black`, `bg-muted`, `bg-body`.
* **Text Colors**: `text-primary`, `text-secondary`, `text-success`, `text-danger`, `text-warning`, `text-info`, `text-light`, `text-dark`, `text-muted`.

---

## 📖 API Reference

### `termstrap.Model`

The primary struct for configuring and rendering layouts:

```go
type Model struct {
    // Content is the input markdown and/or HTML grid content.
    Content string

    // Width is the target terminal width in character columns.
    Width int

    // RootPath is the base filesystem directory for resolving local relative image paths.
    RootPath string

    // ImageRenderer overrides the image renderer (protocol, color mode, etc.).
    // If nil, automatically detected from terminal capabilities.
    ImageRenderer termimage.Renderer

    // MaxImageWidth caps image width in terminal columns (0 = unbounded).
    MaxImageWidth int

    // ColorMode specifies default color depth (ColorModeTrueColor, ColorMode256, ColorMode16).
    ColorMode termimage.ColorMode

    // OptimizeSequences controls whether adjacent matching ANSI color sequences are deduplicated.
    OptimizeSequences *bool

    // DisableImages replaces images with text descriptions ([Image: alt](url)).
    DisableImages bool

    // CachePolicy configures image and ANSI caching (CacheDefault, CacheReload, CacheNoStore).
    CachePolicy CachePolicy

    // CacheTTL sets custom cache TTL (0 = default 2 hours).
    CacheTTL time.Duration

    // ImageCache allows injecting a custom ImageCache implementation.
    ImageCache ImageCache
}
```

### Rendering Methods

* **`m.Render() (string, error)`**: Returns complete ANSI-rendered terminal output.
* **`m.RenderWithOverlays() (string, []ImageOverlay, error)`**: Separates text layout from native graphics escape sequences for TUI architectures.

---

## 🫧 Bubble Tea Integration

When integrating Termstrap with [Bubble Tea](https://github.com/charmbracelet/bubbletea), native terminal graphics (iTerm2, Kitty, Sixel) can be fragmented by Bubble Tea's cell buffer (`ultraviolet`).

The `github.com/go-scripts/termstrap/bubbletea` adapter solves this by separating layout text from out-of-band graphics sequences:

```go
package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-scripts/termstrap"
	tsbubble "github.com/go-scripts/termstrap/bubbletea"
)

type model struct {
	content string
	width   int
	height  int
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() string {
	tm := termstrap.Model{
		Content: m.content,
		Width:   m.width,
	}

	result, err := tsbubble.Render(tm)
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}

	// For HalfBlock, result.Text contains everything.
	// For native protocols (iTerm2, Kitty, Sixel), emit overlays out-of-band if needed:
	if result.HasOverlays() {
		// OverlaySequence returns positioned CUP escape sequences
		_ = result.OverlaySequence()
	}

	return result.Text
}

func main() {
	m := model{
		content: "# Hello Bubble Tea\n\n<div class=\"row\"><div class=\"col-6 border p-1\">Left</div><div class=\"col-6 border p-1\">Right</div></div>",
		width:   80,
	}
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
```

---

## ⚡ Caching System

Termstrap includes a thread-safe in-memory caching system for downloaded images and computed ANSI half-block outputs.

```go
// Custom cache configuration
m := termstrap.Model{
    Content:     content,
    Width:       100,
    CachePolicy: termstrap.CacheDefault, // CacheDefault, CacheReload, CacheNoStore
    CacheTTL:    30 * time.Minute,       // Custom expiration duration
}

// Or clear cache programmatically
termstrap.DefaultCache().Clear()
```

---

## 📁 Examples

Run any of the comprehensive examples located in [`examples/`](./examples):

```bash
# Three-column layout with centered images & borders
go run ./examples/image/three-columns/

# Compare TrueColor vs 256 colors vs 16 colors
go run ./examples/image/colormodes/

# Image caching demonstration (hit / miss / reload)
go run ./examples/image/caching/

# Nested grid layouts without Markdown
go run ./examples/nested-no-md/

# Responsive breakpoints and column stacking
go run ./examples/breakpoints/

# Terminal graphics protocol detection
go run ./examples/image/protocols/
```

### Force a Specific Protocol

You can test any protocol via the `TERMSTRAP_IMAGE_PROTOCOL` environment variable:

```bash
# Force universal Half-Block
TERMSTRAP_IMAGE_PROTOCOL=halfblock go run ./examples/image/three-columns/

# Force iTerm2 inline images
TERMSTRAP_IMAGE_PROTOCOL=iterm2 go run ./examples/image/three-columns/

# Force Kitty graphics
TERMSTRAP_IMAGE_PROTOCOL=kitty go run ./examples/image/three-columns/

# Force Sixel graphics
TERMSTRAP_IMAGE_PROTOCOL=sixel go run ./examples/image/three-columns/
```

---

## 🧪 Testing

Run unit tests and integration tests:

```bash
# Run quick tests
make test-short

# Run full test suite
make test

# Run tests with race detection
go test -race ./...
```

---

## 📄 License

This project is licensed under the **MIT License**. See the [LICENSE](LICENSE) file for details.
