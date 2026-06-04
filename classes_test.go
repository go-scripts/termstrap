package termstrap

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

// ---------- resolveStyle tests ----------

func TestResolveStyle_Padding(t *testing.T) {
	tests := []struct {
		name    string
		classes []string
		wantPT  int
		wantPB  int
		wantPL  int
		wantPR  int
	}{
		{"p-0", []string{"p-0"}, 0, 0, 0, 0},
		{"p-1", []string{"p-1"}, 1, 1, 1, 1},
		{"p-2", []string{"p-2"}, 1, 1, 2, 2},
		{"p-3", []string{"p-3"}, 2, 2, 3, 3},
		{"p-4", []string{"p-4"}, 2, 2, 4, 4},
		{"p-5", []string{"p-5"}, 3, 3, 5, 5},
		{"px-1", []string{"px-1"}, 0, 0, 1, 1},
		{"px-2", []string{"px-2"}, 0, 0, 2, 2},
		{"px-3", []string{"px-3"}, 0, 0, 3, 3},
		{"py-1", []string{"py-1"}, 1, 1, 0, 0},
		{"py-2", []string{"py-2"}, 2, 2, 0, 0},
		{"py-3", []string{"py-3"}, 3, 3, 0, 0},
		{"pt-1", []string{"pt-1"}, 1, 0, 0, 0},
		{"pt-2", []string{"pt-2"}, 2, 0, 0, 0},
		{"pb-1", []string{"pb-1"}, 0, 1, 0, 0},
		{"pb-2", []string{"pb-2"}, 0, 2, 0, 0},
		{"ps-1", []string{"ps-1"}, 0, 0, 1, 0},
		{"ps-2", []string{"ps-2"}, 0, 0, 2, 0},
		{"pe-1", []string{"pe-1"}, 0, 0, 0, 1},
		{"pe-2", []string{"pe-2"}, 0, 0, 0, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := resolveStyle(tt.classes)
			if s.PaddingTop != tt.wantPT {
				t.Errorf("PaddingTop = %d, want %d", s.PaddingTop, tt.wantPT)
			}
			if s.PaddingBottom != tt.wantPB {
				t.Errorf("PaddingBottom = %d, want %d", s.PaddingBottom, tt.wantPB)
			}
			if s.PaddingLeft != tt.wantPL {
				t.Errorf("PaddingLeft = %d, want %d", s.PaddingLeft, tt.wantPL)
			}
			if s.PaddingRight != tt.wantPR {
				t.Errorf("PaddingRight = %d, want %d", s.PaddingRight, tt.wantPR)
			}
		})
	}
}

func TestResolveStyle_Margin(t *testing.T) {
	tests := []struct {
		name    string
		classes []string
		wantMT  int
		wantMB  int
		wantML  int
		wantMR  int
	}{
		{"m-0", []string{"m-0"}, 0, 0, 0, 0},
		{"m-1", []string{"m-1"}, 1, 1, 1, 1},
		{"m-2", []string{"m-2"}, 1, 1, 2, 2},
		{"m-3", []string{"m-3"}, 2, 2, 3, 3},
		{"mx-1", []string{"mx-1"}, 0, 0, 1, 1},
		{"mx-2", []string{"mx-2"}, 0, 0, 2, 2},
		{"my-1", []string{"my-1"}, 1, 1, 0, 0},
		{"my-2", []string{"my-2"}, 2, 2, 0, 0},
		{"mt-1", []string{"mt-1"}, 1, 0, 0, 0},
		{"mt-2", []string{"mt-2"}, 2, 0, 0, 0},
		{"mb-1", []string{"mb-1"}, 0, 1, 0, 0},
		{"mb-2", []string{"mb-2"}, 0, 2, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := resolveStyle(tt.classes)
			if s.MarginTop != tt.wantMT {
				t.Errorf("MarginTop = %d, want %d", s.MarginTop, tt.wantMT)
			}
			if s.MarginBottom != tt.wantMB {
				t.Errorf("MarginBottom = %d, want %d", s.MarginBottom, tt.wantMB)
			}
			if s.MarginLeft != tt.wantML {
				t.Errorf("MarginLeft = %d, want %d", s.MarginLeft, tt.wantML)
			}
			if s.MarginRight != tt.wantMR {
				t.Errorf("MarginRight = %d, want %d", s.MarginRight, tt.wantMR)
			}
		})
	}
}

func TestResolveStyle_TextAlignment(t *testing.T) {
	tests := []struct {
		class    string
		expected lipgloss.Position
	}{
		{"text-center", lipgloss.Center},
		{"text-end", lipgloss.Right},
		{"text-right", lipgloss.Right},
		{"text-start", lipgloss.Left},
		{"text-left", lipgloss.Left},
	}

	for _, tt := range tests {
		t.Run(tt.class, func(t *testing.T) {
			s := resolveStyle([]string{tt.class})
			if s.TextAlign != tt.expected {
				t.Errorf("TextAlign = %v, want %v", s.TextAlign, tt.expected)
			}
		})
	}
}

func TestResolveStyle_DefaultAlignment(t *testing.T) {
	s := resolveStyle([]string{})
	if s.TextAlign != lipgloss.Left {
		t.Errorf("default TextAlign = %v, want Left", s.TextAlign)
	}
}

func TestResolveStyle_Bold(t *testing.T) {
	for _, class := range []string{"fw-bold", "text-bold"} {
		t.Run(class, func(t *testing.T) {
			s := resolveStyle([]string{class})
			if !s.Bold {
				t.Errorf("Bold should be true for class %q", class)
			}
		})
	}
}

func TestResolveStyle_Borders(t *testing.T) {
	tests := []struct {
		class string
		check func(s style) bool
	}{
		{"border", func(s style) bool { return s.Border }},
		{"border-top", func(s style) bool { return s.BorderTop }},
		{"border-bottom", func(s style) bool { return s.BorderBottom }},
		{"border-start", func(s style) bool { return s.BorderLeft }},
		{"border-left", func(s style) bool { return s.BorderLeft }},
		{"border-end", func(s style) bool { return s.BorderRight }},
		{"border-right", func(s style) bool { return s.BorderRight }},
		{"rounded", func(s style) bool { return s.Rounded }},
	}

	for _, tt := range tests {
		t.Run(tt.class, func(t *testing.T) {
			s := resolveStyle([]string{tt.class})
			if !tt.check(s) {
				t.Errorf("expected %q to set its corresponding border flag", tt.class)
			}
		})
	}
}

func TestResolveStyle_Shadows(t *testing.T) {
	tests := []struct {
		class  string
		expect int
	}{
		{"shadow-sm", 1},
		{"shadow", 2},
		{"shadow-lg", 3},
		{"shadow-none", 0},
	}

	for _, tt := range tests {
		t.Run(tt.class, func(t *testing.T) {
			s := resolveStyle([]string{tt.class})
			if s.Shadow != tt.expect {
				t.Errorf("Shadow = %d, want %d", s.Shadow, tt.expect)
			}
		})
	}
}

func TestResolveStyle_BackgroundColors(t *testing.T) {
	tests := []struct {
		class string
		color string
	}{
		{"bg-primary", "#0d6efd"},
		{"bg-secondary", "#6c757d"},
		{"bg-success", "#198754"},
		{"bg-danger", "#dc3545"},
		{"bg-warning", "#ffc107"},
		{"bg-info", "#0dcaf0"},
		{"bg-light", "#f8f9fa"},
		{"bg-dark", "#212529"},
		{"bg-white", "#ffffff"},
		{"bg-black", "#000000"},
	}

	for _, tt := range tests {
		t.Run(tt.class, func(t *testing.T) {
			s := resolveStyle([]string{tt.class})
			if s.BgColor != tt.color {
				t.Errorf("BgColor = %q, want %q", s.BgColor, tt.color)
			}
		})
	}
}

func TestResolveStyle_TextColors(t *testing.T) {
	tests := []struct {
		class string
		color string
	}{
		{"text-white", "#ffffff"},
		{"text-dark", "#212529"},
		{"text-muted", "#6c757d"},
		{"text-primary", "#0d6efd"},
	}

	for _, tt := range tests {
		t.Run(tt.class, func(t *testing.T) {
			s := resolveStyle([]string{tt.class})
			if s.FgColor != tt.color {
				t.Errorf("FgColor = %q, want %q", s.FgColor, tt.color)
			}
		})
	}
}

func TestResolveStyle_TextColorNotConfusedWithAlignment(t *testing.T) {
	// text-center should not set FgColor
	s := resolveStyle([]string{"text-center"})
	if s.FgColor != "" {
		t.Errorf("text-center should not set FgColor, got %q", s.FgColor)
	}
}

func TestResolveStyle_CombinedClasses(t *testing.T) {
	s := resolveStyle([]string{"p-2", "m-1", "border", "rounded", "shadow", "bg-dark", "text-white", "fw-bold"})

	if s.PaddingLeft != 2 || s.PaddingRight != 2 {
		t.Errorf("p-2 padding: L=%d R=%d", s.PaddingLeft, s.PaddingRight)
	}
	if s.MarginTop != 1 || s.MarginBottom != 1 {
		t.Errorf("m-1 margin: T=%d B=%d", s.MarginTop, s.MarginBottom)
	}
	if !s.Border {
		t.Error("Border should be true")
	}
	if !s.Rounded {
		t.Error("Rounded should be true")
	}
	if s.Shadow != 2 {
		t.Errorf("Shadow = %d, want 2", s.Shadow)
	}
	if s.BgColor != "#212529" {
		t.Errorf("BgColor = %q, want #212529", s.BgColor)
	}
	if s.FgColor != "#ffffff" {
		t.Errorf("FgColor = %q, want #ffffff", s.FgColor)
	}
	if !s.Bold {
		t.Error("Bold should be true")
	}
}

// ---------- resolveColor tests ----------

func TestResolveColor(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"primary", "#0d6efd"},
		{"secondary", "#6c757d"},
		{"success", "#198754"},
		{"danger", "#dc3545"},
		{"warning", "#ffc107"},
		{"info", "#0dcaf0"},
		{"light", "#f8f9fa"},
		{"dark", "#212529"},
		{"white", "#ffffff"},
		{"black", "#000000"},
		{"muted", "#6c757d"},
		{"body", "#212529"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveColor(tt.name)
			if got != tt.want {
				t.Errorf("resolveColor(%q) = %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}

func TestResolveColor_Unknown(t *testing.T) {
	got := resolveColor("nonexistent")
	if got != "" {
		t.Errorf("resolveColor(nonexistent) = %q, want empty", got)
	}
}

// ---------- applyPartialBorders tests ----------

func TestApplyPartialBorders_NoBorders(t *testing.T) {
	content := "Hello\nWorld"
	result := applyPartialBorders(content, 20, false, false, false, false)
	if result != content {
		t.Errorf("no borders should return original content, got %q", result)
	}
}

func TestApplyPartialBorders_TopOnly(t *testing.T) {
	content := "Hello"
	result := applyPartialBorders(content, 10, true, false, false, false)
	lines := strings.Split(result, "\n")

	if len(lines) < 2 {
		t.Fatalf("expected at least 2 lines, got %d", len(lines))
	}
	if !strings.Contains(lines[0], "─") {
		t.Errorf("first line should contain horizontal border: %q", lines[0])
	}
}

func TestApplyPartialBorders_BottomOnly(t *testing.T) {
	content := "Hello"
	result := applyPartialBorders(content, 10, false, true, false, false)
	lines := strings.Split(result, "\n")

	lastLine := lines[len(lines)-1]
	if !strings.Contains(lastLine, "─") {
		t.Errorf("last line should contain horizontal border: %q", lastLine)
	}
}

func TestApplyPartialBorders_LeftAndRight(t *testing.T) {
	content := "Hello"
	result := applyPartialBorders(content, 10, false, false, true, true)
	lines := strings.Split(result, "\n")

	for _, line := range lines {
		if !strings.HasPrefix(line, "│") {
			t.Errorf("line should start with │: %q", line)
		}
		if !strings.HasSuffix(line, "│") {
			t.Errorf("line should end with │: %q", line)
		}
	}
}

func TestApplyPartialBorders_AllSides(t *testing.T) {
	content := "Hello"
	result := applyPartialBorders(content, 10, true, true, true, true)
	lines := strings.Split(result, "\n")

	if len(lines) < 3 {
		t.Fatalf("expected at least 3 lines (top+content+bottom), got %d", len(lines))
	}
	// Top line should have corners
	if !strings.Contains(lines[0], "┌") || !strings.Contains(lines[0], "┐") {
		t.Errorf("top line missing corners: %q", lines[0])
	}
	// Bottom line should have corners
	lastLine := lines[len(lines)-1]
	if !strings.Contains(lastLine, "└") || !strings.Contains(lastLine, "┘") {
		t.Errorf("bottom line missing corners: %q", lastLine)
	}
}

// ---------- buildBorderLine tests ----------

func TestBuildBorderLine(t *testing.T) {
	tests := []struct {
		name     string
		width    int
		hasLeft  bool
		hasRight bool
		expected string
	}{
		{"no corners", 10, false, false, "──────────"},
		{"left corner", 10, true, false, "┌─────────"},
		{"right corner", 10, false, true, "─────────┐"},
		{"both corners", 10, true, true, "┌────────┐"},
		{"width 2 both corners", 2, true, true, "┌┐"},
		{"width 1", 1, false, false, "─"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildBorderLine(tt.width, tt.hasLeft, tt.hasRight, "┌", "┐", "─")
			if got != tt.expected {
				t.Errorf("buildBorderLine(%d, %v, %v) = %q, want %q",
					tt.width, tt.hasLeft, tt.hasRight, got, tt.expected)
			}
		})
	}
}
