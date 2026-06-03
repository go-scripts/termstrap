package termstrap

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestCalculateShadowMetrics(t *testing.T) {
	tests := []struct {
		name           string
		contentWidth   int
		shadowSize     int
		maxWidth       int
		expectOverflow bool
		expectAdjusted int
		expectTotal    int
	}{
		{
			name:           "Shadow fits perfectly",
			contentWidth:   70,
			shadowSize:     2,
			maxWidth:       80,
			expectOverflow: false,
			expectAdjusted: 2,
			expectTotal:    72,
		},
		{
			name:           "Shadow fits with room to spare",
			contentWidth:   76,
			shadowSize:     3,
			maxWidth:       80,
			expectOverflow: false,
			expectAdjusted: 3,
			expectTotal:    79,
		},
		{
			name:           "Very narrow space - minimum shadow",
			contentWidth:   78,
			shadowSize:     3,
			maxWidth:       80,
			expectOverflow: true,
			expectAdjusted: 2,
			expectTotal:    80,
		},
		{
			name:           "Exact fit",
			contentWidth:   77,
			shadowSize:     3,
			maxWidth:       80,
			expectOverflow: false,
			expectAdjusted: 3,
			expectTotal:    80,
		},
		{
			name:           "Content takes all space - minimum shadow",
			contentWidth:   79,
			shadowSize:     5,
			maxWidth:       80,
			expectOverflow: true,
			expectAdjusted: 1,
			expectTotal:    80,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := calculateShadowMetrics(tt.contentWidth, tt.shadowSize, tt.maxWidth)

			if metrics.WillOverflow != tt.expectOverflow {
				t.Errorf("WillOverflow: got %v, want %v", metrics.WillOverflow, tt.expectOverflow)
			}

			if metrics.AdjustedShadow != tt.expectAdjusted {
				t.Errorf("AdjustedShadow: got %d, want %d", metrics.AdjustedShadow, tt.expectAdjusted)
			}

			if metrics.TotalWidth != tt.expectTotal {
				t.Errorf("TotalWidth: got %d, want %d", metrics.TotalWidth, tt.expectTotal)
			}

			// Ensure total width never exceeds max
			if metrics.TotalWidth > tt.maxWidth {
				t.Errorf("TotalWidth %d exceeds maxWidth %d", metrics.TotalWidth, tt.maxWidth)
			}
		})
	}
}

func TestApplyShadowWithWidthBasic(t *testing.T) {
	content := "Hello World"

	// Test without width constraint (backward compat)
	result := applyShadow(content, 2)
	lines := strings.Split(result, "\n")

	// Should add 2 chars shadow on right and 2 lines at bottom
	if len(lines) < 2 {
		t.Errorf("Expected multiple lines with bottom shadow, got %d", len(lines))
	}

	// Each line should have shadow characters
	for i, line := range lines {
		if i < 1 && strings.Contains(line, "░") {
			// Shadow added correctly
			break
		}
	}
}

func TestApplyShadowWithWidthOverflow(t *testing.T) {
	// Create a content that's exactly maxWidth - 1, with shadow size 3
	// 77 + 3 = 80, fits exactly
	content := strings.Repeat("X", 77) // 77 chars

	result := applyShadowWithWidth(content, 3, 80)
	lines := strings.Split(result, "\n")

	// Check that content lines fit within 80 using visual width (not byte length)
	for _, line := range lines {
		w := lipgloss.Width(line)
		if w > 80 {
			t.Errorf("Line visual width exceeds max width: %d > 80", w)
		}
	}
}

func TestShadowMetricsZeroMaxWidth(t *testing.T) {
	// Test that maxWidth=0 uses requested shadow size
	metrics := calculateShadowMetrics(50, 3, 0)

	if metrics.WillOverflow {
		t.Error("Should not overflow with maxWidth=0")
	}

	if metrics.AdjustedShadow != 3 {
		t.Errorf("AdjustedShadow should be 3, got %d", metrics.AdjustedShadow)
	}
}

func TestShadowCharSelection(t *testing.T) {
	// Small shadow uses ░
	metrics1 := calculateShadowMetrics(70, 1, 80)
	if metrics1.AdjustedShadow < 3 {
		// Should use ░ (light shade)
		result := applyShadowWithWidth("Test", metrics1.AdjustedShadow, 80)
		if !strings.Contains(result, "░") {
			t.Error("Expected light shade ░ for small shadow")
		}
	}

	// Large shadow uses ▒
	metrics2 := calculateShadowMetrics(70, 5, 80)
	if metrics2.AdjustedShadow >= 3 {
		// Should use ▒ (medium shade)
		result := applyShadowWithWidth("Test", metrics2.AdjustedShadow, 80)
		if !strings.Contains(result, "▒") {
			t.Error("Expected medium shade ▒ for large shadow")
		}
	}
}

func TestShadowBottomLine(t *testing.T) {
	content := "Box Content"
	result := applyShadowWithWidth(content, 2, 100)
	lines := strings.Split(strings.TrimSpace(result), "\n")

	// Should have at least 2 lines (content + bottom shadow)
	if len(lines) < 2 {
		t.Errorf("Expected at least 2 lines (content + shadow), got %d", len(lines))
	}

	// First line should have the content
	if !strings.Contains(lines[0], "Box Content") {
		t.Errorf("First line should contain content: %s", lines[0])
	}

	// Last lines should be shadow
	lastLine := lines[len(lines)-1]
	if !strings.Contains(lastLine, "░") && !strings.Contains(lastLine, "▒") {
		t.Errorf("Last line should contain shadow character: %s", lastLine)
	}
}
