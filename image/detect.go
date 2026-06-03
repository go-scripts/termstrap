package image

import (
	"os"
	"strings"

	"github.com/muesli/termenv"
	"golang.org/x/term"
)

// Detect probes the terminal and returns its graphics capabilities.
// It uses environment variables for fast, safe detection (no escape sequences).
// This is safe to call even when stdout is piped or redirected.
func Detect() Capabilities {
	caps := Capabilities{}

	// Detect color profile via termenv
	output := termenv.NewOutput(os.Stdout)
	profile := output.EnvColorProfile()
	caps.TrueColor = profile == termenv.TrueColor

	// Detect terminal dimensions
	caps.ColCount, caps.RowCount = terminalSize()

	// Detect best graphics protocol from environment
	caps.Protocol = detectFromEnv()

	return caps
}

// detectFromEnv determines the best graphics protocol using only environment
// variables. This is the fast, safe path that never sends escape sequences.
//
// Detection hierarchy (highest fidelity first):
//  1. Kitty:  TERM=xterm-kitty or TERM_PROGRAM=kitty
//  2. iTerm2: TERM_PROGRAM=iTerm.app, LC_TERMINAL=iTerm2, or TERM_PROGRAM=WezTerm
//  3. Sixel:  TERM contains mlterm/yaft, or TERM_PROGRAM is foot/contour
//  4. HalfBlock: universal fallback
//
// Users can override detection by setting TERMSTRAP_IMAGE_PROTOCOL to one of:
// kitty, iterm2, sixel, halfblock.
func detectFromEnv() Protocol {
	// User override takes priority
	if override := os.Getenv("TERMSTRAP_IMAGE_PROTOCOL"); override != "" {
		switch strings.ToLower(override) {
		case "kitty":
			return Kitty
		case "iterm2", "iterm":
			return ITerm2
		case "sixel":
			return Sixel
		case "halfblock", "ansi", "block":
			return HalfBlock
		}
	}

	term := os.Getenv("TERM")
	termProgram := os.Getenv("TERM_PROGRAM")
	lcTerminal := os.Getenv("LC_TERMINAL")

	// Kitty detection
	if term == "xterm-kitty" || strings.EqualFold(termProgram, "kitty") {
		return Kitty
	}

	// WezTerm supports Kitty graphics protocol natively
	if strings.EqualFold(termProgram, "WezTerm") {
		return Kitty
	}

	// iTerm2 detection
	if strings.EqualFold(termProgram, "iTerm.app") || strings.EqualFold(lcTerminal, "iTerm2") {
		return ITerm2
	}

	// Sixel-capable terminals
	termLower := strings.ToLower(term)
	termProgLower := strings.ToLower(termProgram)
	if strings.Contains(termLower, "mlterm") ||
		strings.Contains(termLower, "yaft") ||
		termProgLower == "foot" ||
		termProgLower == "contour" ||
		termProgLower == "konsole" {
		return Sixel
	}

	return HalfBlock
}

// terminalSize returns the terminal width and height in columns/rows.
// Returns (80, 24) as defaults if detection fails.
func terminalSize() (cols, rows int) {
	w, h, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || w == 0 {
		return 80, 24
	}
	return w, h
}
