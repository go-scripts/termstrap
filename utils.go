package termstrap

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

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

// hexToRGB converts a hex color string (#RGB or #RRGGBB) to (r, g, b) components.
func hexToRGB(hex string) (int, int, int, error) {
	hex = strings.TrimPrefix(hex, "#")

	if len(hex) == 3 {
		hex = string([]byte{hex[0], hex[0], hex[1], hex[1], hex[2], hex[2]})
	}

	if len(hex) != 6 {
		return 0, 0, 0, fmt.Errorf("invalid hex color length: %s", hex)
	}

	r, err := strconv.ParseInt(hex[0:2], 16, 32)
	if err != nil {
		return 0, 0, 0, err
	}
	g, err := strconv.ParseInt(hex[2:4], 16, 32)
	if err != nil {
		return 0, 0, 0, err
	}
	b, err := strconv.ParseInt(hex[4:6], 16, 32)
	if err != nil {
		return 0, 0, 0, err
	}

	return int(r), int(g), int(b), nil
}
