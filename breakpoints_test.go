package termstrap

import (
	"testing"
)

func TestDetectBreakpoint(t *testing.T) {
	tests := []struct {
		name     string
		width    int
		expected breakpoint
	}{
		{"extra small", 40, bpXS},
		{"small threshold", 60, bpSM},
		{"small", 70, bpSM},
		{"medium threshold", 80, bpMD},
		{"medium", 100, bpMD},
		{"large threshold", 120, bpLG},
		{"large", 140, bpLG},
		{"extra large threshold", 160, bpXL},
		{"extra large", 200, bpXL},
		{"zero width", 0, bpXS},
		{"just below sm", 59, bpXS},
		{"just below md", 79, bpSM},
		{"just below lg", 119, bpMD},
		{"just below xl", 159, bpLG},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectBreakpoint(tt.width)
			if got != tt.expected {
				t.Errorf("detectBreakpoint(%d) = %d, want %d", tt.width, got, tt.expected)
			}
		})
	}
}

func TestBreakpointPrefix(t *testing.T) {
	tests := []struct {
		bp       breakpoint
		expected string
	}{
		{bpXS, "col-"},
		{bpSM, "col-sm-"},
		{bpMD, "col-md-"},
		{bpLG, "col-lg-"},
		{bpXL, "col-xl-"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := breakpointPrefix(tt.bp)
			if got != tt.expected {
				t.Errorf("breakpointPrefix(%d) = %q, want %q", tt.bp, got, tt.expected)
			}
		})
	}
}

func TestIsStacked(t *testing.T) {
	tests := []struct {
		name     string
		classes  []string
		bp       breakpoint
		expected bool
	}{
		{
			name:     "col-md at sm breakpoint - should stack",
			classes:  []string{"col-md-6"},
			bp:       bpSM,
			expected: true,
		},
		{
			name:     "col-md at md breakpoint - should not stack",
			classes:  []string{"col-md-6"},
			bp:       bpMD,
			expected: false,
		},
		{
			name:     "col-md at lg breakpoint - should not stack",
			classes:  []string{"col-md-6"},
			bp:       bpLG,
			expected: false,
		},
		{
			name:     "col-lg at sm breakpoint - should stack",
			classes:  []string{"col-lg-4"},
			bp:       bpSM,
			expected: true,
		},
		{
			name:     "col-sm at xs breakpoint - should stack",
			classes:  []string{"col-sm-6"},
			bp:       bpXS,
			expected: true,
		},
		{
			name:     "col-sm at sm breakpoint - should not stack",
			classes:  []string{"col-sm-6"},
			bp:       bpSM,
			expected: false,
		},
		{
			name:     "col-xs at xs breakpoint - should not stack",
			classes:  []string{"col-6"},
			bp:       bpXS,
			expected: false,
		},
		{
			name:     "no col classes - should not stack",
			classes:  []string{"border", "rounded"},
			bp:       bpMD,
			expected: false,
		},
		{
			name:     "multiple breakpoints with current match",
			classes:  []string{"col-sm-12", "col-md-6"},
			bp:       bpMD,
			expected: false,
		},
		{
			name:     "col-xl at md breakpoint - should stack",
			classes:  []string{"col-xl-3"},
			bp:       bpMD,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isStacked(tt.classes, tt.bp)
			if got != tt.expected {
				t.Errorf("isStacked(%v, %d) = %v, want %v", tt.classes, tt.bp, got, tt.expected)
			}
		})
	}
}

func TestGridColumnsConstant(t *testing.T) {
	if gridColumns != 12 {
		t.Errorf("gridColumns = %d, want 12", gridColumns)
	}
}

func TestBreakpointThresholds(t *testing.T) {
	if thresholdSM != 60 {
		t.Errorf("thresholdSM = %d, want 60", thresholdSM)
	}
	if thresholdMD != 80 {
		t.Errorf("thresholdMD = %d, want 80", thresholdMD)
	}
	if thresholdLG != 120 {
		t.Errorf("thresholdLG = %d, want 120", thresholdLG)
	}
	if thresholdXL != 160 {
		t.Errorf("thresholdXL = %d, want 160", thresholdXL)
	}
}
