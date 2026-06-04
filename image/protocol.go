// Package image provides multi-protocol terminal image rendering.
//
// It detects terminal graphics capabilities and renders images using the
// highest-fidelity protocol available, with the following fallback hierarchy:
//
//  1. Kitty Graphics Protocol (GPU-accelerated, pixel-perfect)
//  2. iTerm2 Inline Images (Base64 PNG via OSC 1337)
//  3. Sixel Graphics (for compatible terminals like foot, mlterm, xterm)
//  4. ANSI Half-Block (Unicode ▄/▀ with TrueColor — universal fallback)
package image

import (
	"image"
	"io"
	"os"
)

// Protocol represents a terminal graphics protocol.
type Protocol int

const (
	// HalfBlock renders images using Unicode half-block characters (▄/▀)
	// with ANSI TrueColor. Works on virtually all modern terminals.
	HalfBlock Protocol = iota
	// Sixel renders images using the Sixel graphics protocol (DEC standard).
	Sixel
	// ITerm2 renders images using iTerm2's inline image protocol (OSC 1337).
	ITerm2
	// Kitty renders images using the Kitty graphics protocol (APC sequences).
	Kitty
)

// String returns the human-readable name of the protocol.
func (p Protocol) String() string {
	switch p {
	case Kitty:
		return "kitty"
	case ITerm2:
		return "iterm2"
	case Sixel:
		return "sixel"
	case HalfBlock:
		return "halfblock"
	default:
		return "unknown"
	}
}

// Capabilities holds the detected terminal graphics capabilities.
type Capabilities struct {
	Protocol     Protocol // Best available graphics protocol
	TrueColor    bool     // Whether the terminal supports 24-bit color
	TermWidthPx  int      // Terminal width in pixels (0 if unknown)
	TermHeightPx int      // Terminal height in pixels (0 if unknown)
	CellWidthPx  int      // Single cell width in pixels (0 if unknown)
	CellHeightPx int      // Single cell height in pixels (0 if unknown)
	ColCount     int      // Terminal width in columns
	RowCount     int      // Terminal height in rows
}

// Renderer renders an image.Image as a terminal-ready string using
// a specific graphics protocol.
type Renderer interface {
	// Render converts an image to a terminal-displayable string.
	// width is the desired display width in terminal columns.
	Render(img image.Image, width int) (string, error)

	// Protocol returns which graphics protocol this renderer uses.
	Protocol() Protocol
}

// ConstrainedRenderer is optionally implemented by renderers that support
// both width and height constraints. This prevents images from overflowing
// their allocated space in bordered/padded containers.
type ConstrainedRenderer interface {
	Renderer
	// RenderConstrained renders an image constrained to fit within the given
	// width (columns) and height (rows). The image must not exceed these dimensions.
	RenderConstrained(img image.Image, width, height int) (string, error)
}

// Option configures a Renderer created by NewRenderer.
type Option func(*options)

type options struct {
	protocol     *Protocol // nil = auto-detect
	output       io.Writer // terminal output (for termenv detection)
	capabilities *Capabilities
}

// WithProtocol forces a specific graphics protocol instead of auto-detecting.
func WithProtocol(p Protocol) Option {
	return func(o *options) {
		o.protocol = &p
	}
}

// WithOutput sets the terminal output writer used for capability detection.
// Defaults to os.Stdout.
func WithOutput(w io.Writer) Option {
	return func(o *options) {
		o.output = w
	}
}

// WithCapabilities provides pre-detected capabilities, skipping detection.
func WithCapabilities(caps Capabilities) Option {
	return func(o *options) {
		o.capabilities = &caps
	}
}

// NewRenderer creates a Renderer using the best available graphics protocol.
// It auto-detects terminal capabilities unless overridden via options.
func NewRenderer(opts ...Option) Renderer {
	o := &options{
		output: os.Stdout,
	}
	for _, opt := range opts {
		opt(o)
	}

	var proto Protocol
	if o.protocol != nil {
		proto = *o.protocol
	} else if o.capabilities != nil {
		proto = o.capabilities.Protocol
	} else {
		caps := Detect()
		proto = caps.Protocol
	}

	switch proto {
	case Kitty:
		return &kittyRenderer{}
	case ITerm2:
		return &itermRenderer{}
	case Sixel:
		return &sixelRenderer{}
	default:
		return &halfBlockRenderer{}
	}
}
