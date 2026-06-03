package termstrap

import (
	termimage "github.com/go-scripts/termstrap/image"

	"github.com/charmbracelet/glamour"
)

// renderMarkdown renders a markdown string to styled ANSI terminal output.
// It handles image extraction, glamour rendering, line wrapping, and image injection.
func (m Model) renderMarkdown(content string) (string, error) {
	// Extract images and replace with placeholders
	content, imageMap := extractImages(content)

	// Render markdown to ANSI via glamour
	glamourRender, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithEmoji(),
		glamour.WithPreservedNewLines(),
		glamour.WithWordWrap(m.Width),
	)
	if err != nil {
		return "", err
	}

	rendered, err := glamourRender.Render(content)
	if err != nil {
		return "", err
	}

	// Wrap lines that are still too long
	rendered = wrapLongLines(rendered, m.Width)

	// Replace image placeholders with ANSI-rendered images
	rendered = m.renderImages(rendered, imageMap)

	return rendered, nil
}

// renderMarkdownDeferred renders markdown but replaces images with blank
// placeholders instead of actual rendered images. Returns the rendered content
// and a list of deferred images for later cursor-based overlay rendering.
func (m Model) renderMarkdownDeferred(content string, proto termimage.Protocol) (string, []deferredImage, error) {
	content, imageMap := extractImages(content)

	glamourRender, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithEmoji(),
		glamour.WithPreservedNewLines(),
		glamour.WithWordWrap(m.Width),
	)
	if err != nil {
		return "", nil, err
	}

	rendered, err := glamourRender.Render(content)
	if err != nil {
		return "", nil, err
	}

	rendered = wrapLongLines(rendered, m.Width)

	// Trim glamour's extra blank lines BEFORE replacing placeholders
	// with blank space, so trimBlankLines doesn't collapse our placeholders.
	rendered = trimBlankLines(rendered)

	// Replace images with blank space and track for overlay
	rendered, deferred := m.renderImagesDeferred(rendered, imageMap, proto)

	return rendered, deferred, nil
}
