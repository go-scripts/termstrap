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

// layoutDivRegex detects opening <div> tags whose class contains "row" or "container".
var layoutDivRegex = regexp.MustCompile(`<div\s+[^>]*class="[^"]*(?:row|container)[^"]*"`)

// openDivRegex matches any opening <div...> tag.
var openDivRegex = regexp.MustCompile(`<div[\s>]`)

// closeDivRegex matches any </div> tag.
var closeDivRegex = regexp.MustCompile(`</div>`)

// extractSegments splits content into alternating markdown and HTML layout segments.
// HTML layout blocks are identified by the presence of Bootstrap grid classes
// (row, container). A depth counter tracks nested <div> tags so that inner
// rows do not truncate the outer block.
func extractSegments(content string) []segment {
	lines := strings.Split(content, "\n")
	segments := []segment{}
	var mdBuf []string
	var htmlBuf []string
	depth := 0

	flushMD := func() {
		if len(mdBuf) > 0 {
			text := strings.TrimSpace(strings.Join(mdBuf, "\n"))
			if text != "" {
				segments = append(segments, segment{Type: segmentMarkdown, Content: text})
			}
			mdBuf = nil
		}
	}

	for _, line := range lines {
		if depth == 0 {
			// Not inside an HTML layout block — check if this line starts one
			if layoutDivRegex.MatchString(line) {
				flushMD()
				htmlBuf = append(htmlBuf, line)
				depth += len(openDivRegex.FindAllString(line, -1))
				depth -= len(closeDivRegex.FindAllString(line, -1))
			} else {
				mdBuf = append(mdBuf, line)
			}
		} else {
			// Inside an HTML layout block — collect lines until depth returns to 0
			htmlBuf = append(htmlBuf, line)
			depth += len(openDivRegex.FindAllString(line, -1))
			depth -= len(closeDivRegex.FindAllString(line, -1))
			if depth <= 0 {
				depth = 0
				text := strings.TrimSpace(strings.Join(htmlBuf, "\n"))
				if text != "" {
					segments = append(segments, segment{Type: segmentHTML, Content: text})
				}
				htmlBuf = nil
			}
		}
	}

	// Flush remaining buffers
	if len(htmlBuf) > 0 {
		text := strings.TrimSpace(strings.Join(htmlBuf, "\n"))
		if text != "" {
			segments = append(segments, segment{Type: segmentHTML, Content: text})
		}
	}
	flushMD()

	if len(segments) == 0 {
		return []segment{{Type: segmentMarkdown, Content: content}}
	}

	return segments
}
