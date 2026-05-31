package termstrap

import (
	"regexp"
	"strings"
)

// segmentType identifies the type of content segment.
type segmentType int

const (
	segmentMarkdown segmentType = iota
	segmentHTML
)

// segment represents a piece of content, either markdown or HTML layout.
type segment struct {
	Type    segmentType
	Content string
}

// htmlBlockRegex matches HTML blocks that contain Bootstrap layout classes.
// It detects <div class="row">, <div class="container">, etc.
var htmlBlockRegex = regexp.MustCompile(`(?s)(<div\s+class="[^"]*(?:row|container)[^"]*"[^>]*>.*?</div>\s*</div>)`)

// extractSegments splits content into alternating markdown and HTML layout segments.
// HTML layout blocks are identified by the presence of Bootstrap grid classes.
func extractSegments(content string) []segment {
	segments := []segment{}

	matches := htmlBlockRegex.FindAllStringIndex(content, -1)
	if len(matches) == 0 {
		return []segment{{Type: segmentMarkdown, Content: content}}
	}

	lastEnd := 0
	for _, match := range matches {
		// Add markdown segment before this HTML block
		if match[0] > lastEnd {
			md := strings.TrimSpace(content[lastEnd:match[0]])
			if md != "" {
				segments = append(segments, segment{Type: segmentMarkdown, Content: md})
			}
		}

		// Add the HTML layout segment
		htmlContent := strings.TrimSpace(content[match[0]:match[1]])
		segments = append(segments, segment{Type: segmentHTML, Content: htmlContent})

		lastEnd = match[1]
	}

	// Add remaining markdown after last HTML block
	if lastEnd < len(content) {
		md := strings.TrimSpace(content[lastEnd:])
		if md != "" {
			segments = append(segments, segment{Type: segmentMarkdown, Content: md})
		}
	}

	return segments
}
