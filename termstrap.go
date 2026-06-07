// Package termstrap provides Bootstrap-like HTML layout rendering for terminal applications.
// It combines markdown rendering (via glamour), image rendering (via aimg),
// and a responsive 12-column grid system inspired by Bootstrap CSS.
//
// Usage:
//
//	m := termstrap.Model{
//	    Content:  markdownWithHTML,
//	    Width:    terminalWidth,
//	    RootPath: "/path/to/assets",
//	}
//	output, err := m.Render()
package termstrap

import (
	"strings"

	termimage "github.com/go-scripts/termstrap/image"
)

// ImageOverlay represents a native image escape sequence separated from the
// rendered text. This allows consumers that use cell-based buffers (TUI frameworks,
// multiplexers) to emit image sequences through a raw/passthrough channel.
type ImageOverlay struct {
	Row    int    // Line number (0-based) in the rendered text where image starts
	Col    int    // Column position (0-based) in terminal cells
	Width  int    // Image width in terminal columns
	Height int    // Image height in terminal rows
	Escape string // Complete escape sequence (ready to emit to terminal)
}

// Model represents the rendering context for terminal content.
type Model struct {
	Content  string // Markdown content (may contain HTML layout blocks)
	Width    int    // Terminal width in characters
	RootPath string // Base path for resolving local images

	// ImageRenderer overrides the default image renderer.
	// If nil, a renderer is auto-detected on first use based on
	// terminal capabilities (Kitty → iTerm2 → Sixel → half-block).
	ImageRenderer termimage.Renderer
}

// imageRenderer returns the configured renderer, or auto-detects one
// based on terminal capabilities if none was explicitly set.
func (m Model) imageRenderer() termimage.Renderer {
	if m.ImageRenderer != nil {
		return m.ImageRenderer
	}
	return termimage.NewRenderer()
}

// Render processes the content through the full rendering pipeline:
// 1. Extract HTML layout blocks from markdown content
// 2. Render plain markdown sections with glamour
// 3. Parse and render HTML layout blocks with the grid system
// 4. Render images (URL and local)
// 5. Assemble the final ANSI output
func (m Model) Render() (string, error) {
	// Split content into segments: markdown parts and HTML layout blocks
	segments := extractSegments(m.Content)

	var result string
	for _, seg := range segments {
		switch seg.Type {
		case segmentMarkdown:
			rendered, err := m.renderMarkdown(seg.Content)
			if err != nil {
				return "", err
			}
			result += rendered

		case segmentHTML:
			rendered, err := m.renderHTMLLayout(seg.Content)
			if err != nil {
				return "", err
			}
			result += rendered
		}
	}

	return result, nil
}

// RenderWithOverlays processes content and returns text with image overlays separated.
//
// For HalfBlock protocol: overlays is nil, text is identical to Render().
// For native protocols (iTerm2, Kitty, Sixel): text contains blank placeholder
// space where images should appear; overlays contains the escape sequences with
// their absolute positions for out-of-band emission.
func (m Model) RenderWithOverlays() (string, []ImageOverlay, error) {
	proto := m.imageRenderer().Protocol()

	// HalfBlock: images are inline character art, no overlays needed
	if proto == termimage.HalfBlock {
		text, err := m.Render()
		return text, nil, err
	}

	// Native protocols: use deferred rendering and collect overlays
	return m.renderWithOverlaysNative()
}

// renderWithOverlaysNative renders content using deferred image mode, returning
// text with blank placeholders and separate overlay escape sequences.
func (m Model) renderWithOverlaysNative() (string, []ImageOverlay, error) {
	segments := extractSegments(m.Content)
	renderer := m.imageRenderer()
	proto := renderer.Protocol()

	var result string
	var allOverlays []ImageOverlay
	currentRow := 0

	for _, seg := range segments {
		switch seg.Type {
		case segmentMarkdown:
			rendered, deferred, err := m.renderMarkdownDeferred(seg.Content, proto)
			if err != nil {
				return "", nil, err
			}
			for _, di := range deferred {
				overlay := m.deferredToOverlay(di, currentRow, 0, renderer)
				if overlay != nil {
					allOverlays = append(allOverlays, *overlay)
				}
			}
			result += rendered
			currentRow += strings.Count(rendered, "\n")

		case segmentHTML:
			rendered, overlays, err := m.renderHTMLLayoutWithOverlays(seg.Content, currentRow)
			if err != nil {
				return "", nil, err
			}
			result += rendered
			allOverlays = append(allOverlays, overlays...)
			currentRow += strings.Count(rendered, "\n")
		}
	}

	return result, allOverlays, nil
}
