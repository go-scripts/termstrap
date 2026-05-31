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

// Model represents the rendering context for terminal content.
type Model struct {
	Content  string // Markdown content (may contain HTML layout blocks)
	Width    int    // Terminal width in characters
	RootPath string // Base path for resolving local images
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
