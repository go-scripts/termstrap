package termstrap

import (
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
