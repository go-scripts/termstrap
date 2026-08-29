package image

import (
	"image"
	"image/color"
	"strings"
	"testing"
)

func TestParseColorMode(t *testing.T) {
	tests := []struct {
		input    string
		expected ColorMode
		ok       bool
	}{
		{"256", ColorMode256, true},
		{"ansi256", ColorMode256, true},
		{"16", ColorMode16, true},
		{"ansi16", ColorMode16, true},
		{"truecolor", ColorModeTrueColor, true},
		{"24bit", ColorModeTrueColor, true},
		{"invalid", ColorModeTrueColor, false},
	}

	for _, tt := range tests {
		got, ok := ParseColorMode(tt.input)
		if got != tt.expected || ok != tt.ok {
			t.Errorf("ParseColorMode(%q) = (%v, %v), expected (%v, %v)", tt.input, got, ok, tt.expected, tt.ok)
		}
	}
}

func TestRGBTo256(t *testing.T) {
	// Exact primary and grayscale matches
	if idx := RGBTo256(0, 0, 0); idx != 16 {
		t.Errorf("expected black to be 16, got %d", idx)
	}
	if idx := RGBTo256(255, 255, 255); idx != 231 {
		t.Errorf("expected white to be 231, got %d", idx)
	}
	if idx := RGBTo256(255, 0, 0); idx != 196 { // 16 + 36*5 = 196
		t.Errorf("expected pure red to be 196, got %d", idx)
	}
	if idx := RGBTo256(0, 255, 0); idx != 46 { // 16 + 6*5 = 46
		t.Errorf("expected pure green to be 46, got %d", idx)
	}
	if idx := RGBTo256(0, 0, 255); idx != 21 { // 16 + 5 = 21
		t.Errorf("expected pure blue to be 21, got %d", idx)
	}
}

func TestHalfBlockRenderer_ColorModes(t *testing.T) {
	// Create simple 4x4 test image
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x * 60), G: uint8(y * 60), B: 100, A: 255})
		}
	}

	// 1. TrueColor
	rTC := NewHalfBlockRenderer(ColorModeTrueColor, false)
	outTC, err := rTC.Render(img, 4)
	if err != nil {
		t.Fatalf("TrueColor render error: %v", err)
	}
	if !strings.Contains(outTC, ";2;") {
		t.Errorf("expected TrueColor ANSI code ';2;' in output")
	}

	// 2. 256 Colors
	r256 := NewHalfBlockRenderer(ColorMode256, false)
	out256, err := r256.Render(img, 4)
	if err != nil {
		t.Fatalf("256-color render error: %v", err)
	}
	if !strings.Contains(out256, ";5;") {
		t.Errorf("expected 256-color ANSI code ';5;' in output")
	}

	// 3. Optimized vs Unoptimized (byte reduction check)
	r256Opt := NewHalfBlockRenderer(ColorMode256, true)
	// Create uniform image to test deduplication
	uniform := image.NewRGBA(image.Rect(0, 0, 20, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 20; x++ {
			uniform.Set(x, y, color.RGBA{R: 200, G: 100, B: 50, A: 255})
		}
	}

	outUnopt, _ := r256.Render(uniform, 20)
	outOpt, _ := r256Opt.Render(uniform, 20)

	if len(outOpt) >= len(outUnopt) {
		t.Errorf("expected optimized output (%d bytes) to be smaller than unoptimized (%d bytes)", len(outOpt), len(outUnopt))
	}
}
