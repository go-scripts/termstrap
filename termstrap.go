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
	termimage "github.com/go-scripts/termstrap/image"
)

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
