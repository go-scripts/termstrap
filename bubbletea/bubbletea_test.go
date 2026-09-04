package bubbletea

import (
	"testing"

	"github.com/go-scripts/termstrap"
	termimage "github.com/go-scripts/termstrap/image"
)

func TestRender_HalfBlock(t *testing.T) {
	m := termstrap.Model{
		HTML:          "<div>Hello Bubbletea</div>",
		Width:         80,
		ImageRenderer: termimage.NewRenderer(termimage.WithProtocol(termimage.HalfBlock)),
	}

	result, err := Render(m)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}
	if result.Text == "" {
		t.Error("expected non-empty text")
	}
	if len(result.Overlays) != 0 {
		t.Errorf("expected 0 overlays for halfblock, got %d", len(result.Overlays))
	}
}

func TestRender_NativeProtocol(t *testing.T) {
	m := termstrap.Model{
		HTML:          "<div>Plain text</div>",
		Width:         80,
		ImageRenderer: termimage.NewRenderer(termimage.WithProtocol(termimage.ITerm2)),
	}

	result, err := Render(m)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}
	if result.Text == "" {
		t.Error("expected non-empty text")
	}
}
