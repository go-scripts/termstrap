package bubbletea

import (
	"strings"
	"testing"

	"github.com/go-scripts/termstrap"
	termimage "github.com/go-scripts/termstrap/image"
)

func TestRender_HalfBlock_NoOverlays(t *testing.T) {
	m := termstrap.Model{
		Content:       "# Hello\n\nSome text.",
		Width:         80,
		ImageRenderer: termimage.NewRenderer(termimage.WithProtocol(termimage.HalfBlock)),
	}
	result, err := Render(m)
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}
	if result.HasOverlays() {
		t.Error("expected no overlays for HalfBlock")
	}
	if !strings.Contains(result.Text, "Hello") {
		t.Error("expected text to contain 'Hello'")
	}
}

func TestRender_NoImages_EmptyOverlays(t *testing.T) {
	m := termstrap.Model{
		Content:       "# Title\n\nJust plain markdown.",
		Width:         80,
		ImageRenderer: termimage.NewRenderer(termimage.WithProtocol(termimage.ITerm2)),
	}
	result, err := Render(m)
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}
	if result.HasOverlays() {
		t.Error("expected no overlays without images")
	}
	if result.OverlaySequence() != "" {
		t.Error("expected empty overlay sequence without images")
	}
}

func TestRenderResult_OverlaySequence(t *testing.T) {
	result := RenderResult{
		Text: "line1\nline2\nline3\n",
		Overlays: []termstrap.ImageOverlay{
			{Row: 1, Col: 5, Width: 10, Height: 3, Escape: "\x1b]1337;File=inline=1:AAAA\x07"},
		},
	}

	seq := result.OverlaySequence()
	// Should contain cursor save, CUP positioning, the escape, cursor restore
	if !strings.Contains(seq, "\x1b7") {
		t.Error("expected cursor save (ESC 7)")
	}
	if !strings.Contains(seq, "\x1b[2;6H") {
		t.Error("expected CUP sequence for row=1,col=5 (1-based: 2;6)")
	}
	if !strings.Contains(seq, "\x1b]1337;File=inline=1:AAAA\x07") {
		t.Error("expected image escape in overlay sequence")
	}
	if !strings.Contains(seq, "\x1b8") {
		t.Error("expected cursor restore (ESC 8)")
	}
}
