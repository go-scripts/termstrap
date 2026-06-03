package image

import (
	"image"
	"image/color"
	"strings"
	"testing"
)

// newTestImage creates a solid-color RGBA image for testing.
func newTestImage(w, h int, c color.Color) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, c)
		}
	}
	return img
}

func TestProtocolString(t *testing.T) {
	tests := []struct {
		proto Protocol
		want  string
	}{
		{Kitty, "kitty"},
		{ITerm2, "iterm2"},
		{Sixel, "sixel"},
		{HalfBlock, "halfblock"},
		{Protocol(99), "unknown"},
	}
	for _, tt := range tests {
		if got := tt.proto.String(); got != tt.want {
			t.Errorf("Protocol(%d).String() = %q, want %q", tt.proto, got, tt.want)
		}
	}
}

func TestDetectFromEnv(t *testing.T) {
	tests := []struct {
		name     string
		env      map[string]string
		expected Protocol
	}{
		{
			name:     "override kitty",
			env:      map[string]string{"TERMSTRAP_IMAGE_PROTOCOL": "kitty"},
			expected: Kitty,
		},
		{
			name:     "override iterm2",
			env:      map[string]string{"TERMSTRAP_IMAGE_PROTOCOL": "iterm2"},
			expected: ITerm2,
		},
		{
			name:     "override sixel",
			env:      map[string]string{"TERMSTRAP_IMAGE_PROTOCOL": "sixel"},
			expected: Sixel,
		},
		{
			name:     "override halfblock",
			env:      map[string]string{"TERMSTRAP_IMAGE_PROTOCOL": "halfblock"},
			expected: HalfBlock,
		},
		{
			name:     "kitty via TERM",
			env:      map[string]string{"TERM": "xterm-kitty"},
			expected: Kitty,
		},
		{
			name:     "kitty via TERM_PROGRAM",
			env:      map[string]string{"TERM_PROGRAM": "kitty"},
			expected: Kitty,
		},
		{
			name:     "wezterm uses kitty",
			env:      map[string]string{"TERM_PROGRAM": "WezTerm"},
			expected: Kitty,
		},
		{
			name:     "iterm via TERM_PROGRAM",
			env:      map[string]string{"TERM_PROGRAM": "iTerm.app"},
			expected: ITerm2,
		},
		{
			name:     "iterm via LC_TERMINAL",
			env:      map[string]string{"LC_TERMINAL": "iTerm2"},
			expected: ITerm2,
		},
		{
			name:     "sixel via mlterm",
			env:      map[string]string{"TERM": "mlterm"},
			expected: Sixel,
		},
		{
			name:     "sixel via foot",
			env:      map[string]string{"TERM_PROGRAM": "foot"},
			expected: Sixel,
		},
		{
			name:     "fallback to halfblock",
			env:      map[string]string{"TERM": "xterm-256color"},
			expected: HalfBlock,
		},
		{
			name:     "empty env fallback",
			env:      map[string]string{},
			expected: HalfBlock,
		},
	}

	envKeys := []string{"TERMSTRAP_IMAGE_PROTOCOL", "TERM", "TERM_PROGRAM", "LC_TERMINAL"}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all relevant env vars
			for _, k := range envKeys {
				t.Setenv(k, "")
			}
			// Set test env
			for k, v := range tt.env {
				t.Setenv(k, v)
			}

			got := detectFromEnv()
			if got != tt.expected {
				t.Errorf("detectFromEnv() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestResizeToWidth(t *testing.T) {
	img := newTestImage(200, 100, color.White)

	// Should resize
	resized := ResizeToWidth(img, 100)
	bounds := resized.Bounds()
	if bounds.Dx() != 100 {
		t.Errorf("ResizeToWidth(200px, 100) width = %d, want 100", bounds.Dx())
	}
	if bounds.Dy() != 50 {
		t.Errorf("ResizeToWidth(200px, 100) height = %d, want 50", bounds.Dy())
	}

	// Should not resize (already smaller)
	notResized := ResizeToWidth(img, 300)
	if notResized.Bounds().Dx() != 200 {
		t.Errorf("ResizeToWidth(200px, 300) should not resize, got width %d", notResized.Bounds().Dx())
	}
}

func TestColsToPixels(t *testing.T) {
	tests := []struct {
		cols        int
		cellWidth   int
		expectedPx  int
	}{
		{80, 8, 640},
		{40, 0, 320},  // 0 cellWidth defaults to 8
		{100, 10, 1000},
	}
	for _, tt := range tests {
		got := ColsToPixels(tt.cols, tt.cellWidth)
		if got != tt.expectedPx {
			t.Errorf("ColsToPixels(%d, %d) = %d, want %d", tt.cols, tt.cellWidth, got, tt.expectedPx)
		}
	}
}

func TestIsURL(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"https://example.com/img.png", true},
		{"http://localhost:8080/test.jpg", true},
		{"./local/path.png", false},
		{"/absolute/path.png", false},
		{"relative.png", false},
		{"", false},
	}
	for _, tt := range tests {
		if got := IsURL(tt.input); got != tt.want {
			t.Errorf("IsURL(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestKittyRender(t *testing.T) {
	img := newTestImage(4, 4, color.RGBA{255, 0, 0, 255})
	r := &kittyRenderer{}

	result, err := r.Render(img, 10)
	if err != nil {
		t.Fatalf("kittyRenderer.Render() error: %v", err)
	}

	// Verify it starts with Kitty APC sequence
	if !strings.HasPrefix(result, "\x1b_G") {
		t.Errorf("kitty output should start with APC sequence, got %q...", result[:min(20, len(result))])
	}

	// Should contain required parameters
	if !strings.Contains(result, "a=T") {
		t.Error("kitty output missing a=T (transmit action)")
	}
	if !strings.Contains(result, "f=100") {
		t.Error("kitty output missing f=100 (PNG format)")
	}

	// Should end with ST
	trimmed := strings.TrimRight(result, "\n")
	if !strings.HasSuffix(trimmed, "\x1b\\") {
		t.Errorf("kitty output should end with ST (\\x1b\\\\), got ...%q", trimmed[max(0, len(trimmed)-10):])
	}

	if r.Protocol() != Kitty {
		t.Errorf("Protocol() = %v, want Kitty", r.Protocol())
	}
}

func TestItermRender(t *testing.T) {
	img := newTestImage(4, 4, color.RGBA{0, 255, 0, 255})
	r := &itermRenderer{}

	result, err := r.Render(img, 10)
	if err != nil {
		t.Fatalf("itermRenderer.Render() error: %v", err)
	}

	// Verify OSC 1337 sequence
	if !strings.HasPrefix(result, "\x1b]1337;File=") {
		t.Errorf("iterm output should start with OSC 1337, got %q...", result[:min(30, len(result))])
	}

	// Should contain inline=1
	if !strings.Contains(result, "inline=1") {
		t.Error("iterm output missing inline=1")
	}

	// Should end with BEL
	trimmed := strings.TrimRight(result, "\n")
	if !strings.HasSuffix(trimmed, "\a") {
		t.Error("iterm output should end with BEL (\\a)")
	}

	if r.Protocol() != ITerm2 {
		t.Errorf("Protocol() = %v, want ITerm2", r.Protocol())
	}
}

func TestSixelRender(t *testing.T) {
	img := newTestImage(4, 4, color.RGBA{0, 0, 255, 255})
	r := &sixelRenderer{}

	result, err := r.Render(img, 10)
	if err != nil {
		t.Fatalf("sixelRenderer.Render() error: %v", err)
	}

	// Sixel output should start with DCS (Device Control String)
	if !strings.HasPrefix(result, "\x1bP") && !strings.HasPrefix(result, "\033P") {
		t.Errorf("sixel output should start with DCS, got %q...", result[:min(20, len(result))])
	}

	if r.Protocol() != Sixel {
		t.Errorf("Protocol() = %v, want Sixel", r.Protocol())
	}
}

func TestHalfBlockRender(t *testing.T) {
	img := newTestImage(8, 8, color.RGBA{128, 128, 128, 255})
	r := &halfBlockRenderer{}

	result, err := r.Render(img, 10)
	if err != nil {
		t.Fatalf("halfBlockRenderer.Render() error: %v", err)
	}

	// Should contain ANSI escape codes
	if !strings.Contains(result, "\x1b[") {
		t.Error("halfblock output should contain ANSI escape codes")
	}

	if r.Protocol() != HalfBlock {
		t.Errorf("Protocol() = %v, want HalfBlock", r.Protocol())
	}
}

func TestNewRendererWithProtocol(t *testing.T) {
	tests := []struct {
		proto Protocol
		want  Protocol
	}{
		{Kitty, Kitty},
		{ITerm2, ITerm2},
		{Sixel, Sixel},
		{HalfBlock, HalfBlock},
	}
	for _, tt := range tests {
		r := NewRenderer(WithProtocol(tt.proto))
		if got := r.Protocol(); got != tt.want {
			t.Errorf("NewRenderer(WithProtocol(%v)).Protocol() = %v, want %v", tt.proto, got, tt.want)
		}
	}
}

func TestEstimateVisualHeight(t *testing.T) {
	tests := []struct {
		name      string
		imgW, imgH int
		widthCols  int
		proto      Protocol
		want       int
	}{
		{
			name: "halfblock tall image",
			imgW: 400, imgH: 300, widthCols: 22,
			proto: HalfBlock,
			// pxWidth = 22*8=176, pxHeight = 300*176/400=132, rows = (132+1)/2=66
			want: 66,
		},
		{
			name: "iterm2 tall image",
			imgW: 400, imgH: 300, widthCols: 22,
			proto: ITerm2,
			// pxWidth = 176, pxHeight = 132, rows = (132+15)/16=9
			want: 9,
		},
		{
			name: "small image no resize",
			imgW: 10, imgH: 10, widthCols: 22,
			proto: ITerm2,
			// origW 10 < pxWidth 176 → pxHeight = 10, rows = (10+15)/16=1
			want: 1,
		},
		{
			name: "halfblock 1px image",
			imgW: 1, imgH: 1, widthCols: 10,
			proto: HalfBlock,
			want: 1,
		},
		{
			name: "kitty wide image",
			imgW: 800, imgH: 100, widthCols: 40,
			proto: Kitty,
			// pxWidth = 320, pxHeight = 100*320/800=40, rows = (40+15)/16=3
			want: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img := newTestImage(tt.imgW, tt.imgH, color.White)
			got := EstimateVisualHeight(img, tt.widthCols, tt.proto)
			if got != tt.want {
				t.Errorf("EstimateVisualHeight(%dx%d, cols=%d, %v) = %d, want %d",
					tt.imgW, tt.imgH, tt.widthCols, tt.proto, got, tt.want)
			}
		})
	}
}

func TestSplitPayload(t *testing.T) {
	tests := []struct {
		input string
		size  int
		want  int // expected number of chunks
	}{
		{"", 10, 1},
		{"hello", 10, 1},
		{"helloworld", 5, 2},
		{"123456789", 3, 3},
	}
	for _, tt := range tests {
		chunks := splitPayload(tt.input, tt.size)
		if len(chunks) != tt.want {
			t.Errorf("splitPayload(%q, %d) = %d chunks, want %d", tt.input, tt.size, len(chunks), tt.want)
		}
		// Verify all chunks rejoin to the original
		rejoined := strings.Join(chunks, "")
		if rejoined != tt.input {
			t.Errorf("splitPayload chunks don't rejoin: got %q, want %q", rejoined, tt.input)
		}
	}
}
