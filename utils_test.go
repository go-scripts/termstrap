package termstrap

import (
	"strings"
	"testing"
)

// ---------- stripANSI tests ----------

func TestStripANSI(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"no ANSI", "Hello World", "Hello World"},
		{"bold", "\x1b[1mBold\x1b[0m", "Bold"},
		{"color", "\x1b[38;5;252mColored\x1b[0m", "Colored"},
		{"truecolor", "\x1b[38;2;255;0;0mRed\x1b[0m", "Red"},
		{"multiple sequences", "\x1b[1m\x1b[38;5;252mBoldColor\x1b[0m", "BoldColor"},
		{"empty string", "", ""},
		{"only ANSI", "\x1b[0m", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripANSI(tt.input)
			if got != tt.expected {
				t.Errorf("stripANSI(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

// ---------- visualWidth tests ----------

func TestVisualWidth(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"plain text", "Hello", 5},
		{"with ANSI", "\x1b[1mHello\x1b[0m", 5},
		{"empty", "", 0},
		{"spaces", "   ", 0}, // TrimSpace in stripANSI removes spaces
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := visualWidth(tt.input)
			if got != tt.expected {
				t.Errorf("visualWidth(%q) = %d, want %d", tt.input, got, tt.expected)
			}
		})
	}
}

// ---------- maxVisualWidth tests ----------

func TestMaxVisualWidth(t *testing.T) {
	tests := []struct {
		name     string
		lines    []string
		expected int
	}{
		{"single line", []string{"Hello"}, 5},
		{"multiple lines", []string{"Hi", "Hello World", "Hey"}, 11},
		{"with ANSI", []string{"\x1b[1mBold\x1b[0m", "Normal"}, 6},
		{"empty", []string{}, 0},
		{"empty lines", []string{"", "", ""}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := maxVisualWidth(tt.lines)
			if got != tt.expected {
				t.Errorf("maxVisualWidth(%v) = %d, want %d", tt.lines, got, tt.expected)
			}
		})
	}
}

// ---------- trimBlankLines tests ----------

func TestTrimBlankLines(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"no blanks", "Hello\nWorld", "Hello\nWorld"},
		{"leading blanks", "\n\nHello\nWorld", "Hello\nWorld"},
		{"trailing blanks", "Hello\nWorld\n\n", "Hello\nWorld"},
		{"both blanks", "\n\nHello\nWorld\n\n", "Hello\nWorld"},
		{"all blank", "\n\n\n", ""},
		{"empty string", "", ""},
		{"preserves inner blanks", "Hello\n\nWorld", "Hello\n\nWorld"},
		{"ANSI blank lines", "\x1b[0m\n\nContent\n\n\x1b[0m", "Content"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := trimBlankLines(tt.input)
			if got != tt.expected {
				t.Errorf("trimBlankLines(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

// ---------- hexToRGB tests ----------

func TestHexToRGB(t *testing.T) {
	tests := []struct {
		hex     string
		r, g, b uint8
	}{
		{"#000000", 0, 0, 0},
		{"#ffffff", 255, 255, 255},
		{"#ff0000", 255, 0, 0},
		{"#00ff00", 0, 255, 0},
		{"#0000ff", 0, 0, 255},
		{"#0d6efd", 13, 110, 253},
		{"#198754", 25, 135, 84},
		// Without # prefix
		{"198754", 25, 135, 84},
	}

	for _, tt := range tests {
		t.Run(tt.hex, func(t *testing.T) {
			r, g, b := hexToRGB(tt.hex)
			if r != tt.r || g != tt.g || b != tt.b {
				t.Errorf("hexToRGB(%q) = (%d, %d, %d), want (%d, %d, %d)",
					tt.hex, r, g, b, tt.r, tt.g, tt.b)
			}
		})
	}
}

func TestHexToRGB_Invalid(t *testing.T) {
	tests := []struct {
		name string
		hex  string
	}{
		{"too short", "#fff"},
		{"too long", "#fffffff"},
		{"empty", ""},
		{"non-hex", "#gggggg"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, g, b := hexToRGB(tt.hex)
			// Invalid hex should return 0,0,0
			if r != 0 || g != 0 || b != 0 {
				t.Errorf("hexToRGB(%q) = (%d, %d, %d), want (0, 0, 0)", tt.hex, r, g, b)
			}
		})
	}
}

// ---------- persistColors tests ----------

func TestPersistColors_EmptyColors(t *testing.T) {
	content := "Hello\x1b[0mWorld"
	result := persistColors(content, "", "")
	if result != content {
		t.Errorf("persistColors with empty colors should return original")
	}
}

func TestPersistColors_BgOnly(t *testing.T) {
	content := "Hello\x1b[0mWorld"
	result := persistColors(content, "#ff0000", "")

	// Should contain bg restore after reset
	if !strings.Contains(result, "\x1b[48;2;255;0;0m") {
		t.Error("expected bg color sequence after reset")
	}
}

func TestPersistColors_FgOnly(t *testing.T) {
	content := "Hello\x1b[0mWorld"
	result := persistColors(content, "", "#00ff00")

	// Should contain fg restore after reset
	if !strings.Contains(result, "\x1b[38;2;0;255;0m") {
		t.Error("expected fg color sequence after reset")
	}
}

func TestPersistColors_BothColors(t *testing.T) {
	content := "Hello\x1b[0mWorld"
	result := persistColors(content, "#ff0000", "#00ff00")

	if !strings.Contains(result, "\x1b[48;2;255;0;0m") {
		t.Error("expected bg color sequence")
	}
	if !strings.Contains(result, "\x1b[38;2;0;255;0m") {
		t.Error("expected fg color sequence")
	}
}

func TestPersistColors_ReplacesGlamourDefault(t *testing.T) {
	// glamour uses palette 252 for default text color
	content := "Hello\x1b[38;5;252mWorld\x1b[0m"
	result := persistColors(content, "", "#ffffff")

	// The palette 252 should be replaced with our custom fg
	if strings.Contains(result, "\x1b[38;5;252m") {
		t.Error("glamour default color should be replaced")
	}
	if !strings.Contains(result, "\x1b[38;2;255;255;255m") {
		t.Error("expected custom fg color to replace glamour default")
	}
}

// ---------- addPadding tests ----------

func TestAddPadding(t *testing.T) {
	input := "Hello\nWorld"
	result := addPadding(input, 4)
	lines := strings.Split(result, "\n")

	for _, line := range lines {
		if !strings.HasPrefix(line, "    ") {
			t.Errorf("line should have 4-space prefix: %q", line)
		}
	}
}

func TestAddPadding_Zero(t *testing.T) {
	input := "Hello\nWorld"
	result := addPadding(input, 0)
	lines := strings.Split(result, "\n")

	if lines[0] != "Hello" {
		t.Errorf("zero padding should not add spaces: %q", lines[0])
	}
}

// ---------- wrapLongLines tests ----------

func TestWrapLongLines_ShortLines(t *testing.T) {
	input := "Short line\nAnother short"
	result := wrapLongLines(input, 80)

	// Short lines should not be modified
	if strings.Count(result, "\n") != strings.Count(input, "\n") {
		t.Error("short lines should not be wrapped")
	}
}

func TestWrapLongLines_ZeroWidth(t *testing.T) {
	input := "Hello"
	result := wrapLongLines(input, 0)
	if result != input {
		t.Errorf("zero width should return original: %q", result)
	}
}
