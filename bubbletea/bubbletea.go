// Package bubbletea provides an adapter for using termstrap with Bubble Tea applications.
//
// When termstrap renders native images (iTerm2, Kitty, Sixel) inside a Bubble Tea
// program, the escape sequences are broken by the cell buffer (ultraviolet) that
// parses View() content into individual cells. This package separates text from
// image escape sequences so they can be emitted via tea.Raw().
//
// Usage:
//
//	func (m model) View() string {
//	    tm := termstrap.Model{Content: m.content, Width: m.width}
//	    result := bubbletea.Render(tm)
//	    // Queue raw image sequences for out-of-band emission
//	    for _, cmd := range result.ImageCommands() {
//	        m.cmds = append(m.cmds, cmd)
//	    }
//	    return result.Text
//	}
//
//	func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
//	    // ... batch m.cmds with other commands
//	}
//
// For HalfBlock protocol (the universal fallback), images are rendered as Unicode
// characters directly in the text and no raw sequences are needed.
package bubbletea

import (
	"fmt"
	"strings"

	"github.com/go-scripts/termstrap"
)

// RenderResult holds the output of a termstrap render separated into
// cell-buffer-safe text and raw image sequences for Bubble Tea.
type RenderResult struct {
	// Text is the rendered content safe for Bubble Tea's View() return.
	// Image areas are filled with blank space of the correct dimensions.
	Text string

	// Overlays contains native image escape sequences with their positions.
	// Empty when using HalfBlock protocol (images are inline in Text).
	Overlays []termstrap.ImageOverlay
}

// HasOverlays reports whether the render produced any native image overlays
// that need out-of-band emission.
func (r RenderResult) HasOverlays() bool {
	return len(r.Overlays) > 0
}

// OverlaySequence returns a single string containing all image overlays
// formatted with ANSI cursor positioning, ready to be passed to tea.Println
// or written directly to the terminal output.
//
// Each overlay is positioned using CUP (Cursor Position) sequences relative
// to the top of the rendered block. The caller must account for any vertical
// offset if the content is not rendered at terminal row 0.
func (r RenderResult) OverlaySequence() string {
	if len(r.Overlays) == 0 {
		return ""
	}

	var buf strings.Builder
	for _, ov := range r.Overlays {
		// Save cursor position
		buf.WriteString("\x1b7")
		// Move to overlay position (CUP is 1-based)
		fmt.Fprintf(&buf, "\x1b[%d;%dH", ov.Row+1, ov.Col+1)
		// Emit the image escape sequence
		buf.WriteString(ov.Escape)
		// Restore cursor position
		buf.WriteString("\x1b8")
	}
	return buf.String()
}

// Render processes a termstrap Model and returns a RenderResult with text and
// image overlays separated for Bubble Tea consumption.
//
// The returned Text is safe to return from a Bubble Tea View() method.
// Image overlays should be emitted via tea.Raw() in an Update() cycle.
func Render(model termstrap.Model) (RenderResult, error) {
	text, overlays, err := model.RenderWithOverlays()
	if err != nil {
		return RenderResult{}, err
	}
	return RenderResult{
		Text:     text,
		Overlays: overlays,
	}, nil
}
