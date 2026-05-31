package termstrap

import (
	"regexp"
	"strings"

	lineTooLongRender "github.com/MichaelMure/go-term-markdown"
)

// addPadding adds left padding (spaces) to each line of content.
func addPadding(input string, spaces int) string {
	lines := strings.Split(input, "\n")
	pad := strings.Repeat(" ", spaces)
	paddedLines := make([]string, len(lines))
	for i, line := range lines {
		paddedLines[i] = pad + strings.TrimSpace(line)
	}
	return strings.Join(paddedLines, "\n")
}

// wrapLongLines wraps lines that exceed the given width,
// preserving ANSI escape sequences.
func wrapLongLines(content string, width int) string {
	if width <= 0 {
		return content
	}
	lines := strings.Split(content, "\n")
	outputLines := make([]string, 0, len(lines))
	for _, line := range lines {
		visible := stripANSI(line)
		if len(visible) >= width {
			wrapped := string(lineTooLongRender.Render(line, width-5, 0))
			outputLines = append(outputLines, addPadding(wrapped, 2))
		} else {
			outputLines = append(outputLines, line)
		}
	}
	return strings.Join(outputLines, "\n")
}

// ansiRegex matches ANSI escape sequences.
var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// stripANSI removes all ANSI escape sequences from a string.
func stripANSI(s string) string {
	return strings.TrimSpace(ansiRegex.ReplaceAllString(s, ""))
}

// visualWidth returns the visible width of a string (excluding ANSI codes).
func visualWidth(s string) int {
	return len(stripANSI(s))
}

// maxVisualWidth returns the maximum visible width among all lines.
func maxVisualWidth(lines []string) int {
	maxW := 0
	for _, line := range lines {
		w := visualWidth(line)
		if w > maxW {
			maxW = w
		}
	}
	return maxW
}

// trimBlankLines removes leading and trailing blank lines from content,
// accounting for ANSI escape sequences in line content.
func trimBlankLines(s string) string {
	lines := strings.Split(s, "\n")

	// Trim leading blank lines
	start := 0
	for start < len(lines) && strings.TrimSpace(stripANSI(lines[start])) == "" {
		start++
	}

	// Trim trailing blank lines
	end := len(lines)
	for end > start && strings.TrimSpace(stripANSI(lines[end-1])) == "" {
		end--
	}

	if start >= end {
		return ""
	}

	return strings.Join(lines[start:end], "\n")
}
