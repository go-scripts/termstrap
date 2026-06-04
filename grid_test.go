package termstrap

import (
	"testing"
)

// ---------- splitClasses tests ----------

func TestSplitClasses(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{"single class", "row", []string{"row"}},
		{"multiple classes", "col-md-6 border rounded p-1", []string{"col-md-6", "border", "rounded", "p-1"}},
		{"extra spaces", "  col-md-6   border  ", []string{"col-md-6", "border"}},
		{"empty string", "", []string{}},
		{"only spaces", "   ", []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitClasses(tt.input)
			if len(got) != len(tt.expected) {
				t.Fatalf("splitClasses(%q) = %v (len %d), want %v (len %d)",
					tt.input, got, len(got), tt.expected, len(tt.expected))
			}
			for i, v := range got {
				if v != tt.expected[i] {
					t.Errorf("splitClasses(%q)[%d] = %q, want %q", tt.input, i, v, tt.expected[i])
				}
			}
		})
	}
}

// ---------- parseColClass tests ----------

func TestParseColClass(t *testing.T) {
	tests := []struct {
		name   string
		class  string
		prefix string
		want   int
	}{
		{"col-md-6", "col-md-6", "col-md-", 6},
		{"col-md-12", "col-md-12", "col-md-", 12},
		{"col-md-1", "col-md-1", "col-md-", 1},
		{"col-lg-4", "col-lg-4", "col-lg-", 4},
		{"col-sm-3", "col-sm-3", "col-sm-", 3},
		{"col-6 (xs)", "col-6", "col-", 6},
		{"wrong prefix", "col-lg-6", "col-md-", 0},
		{"out of range 0", "col-md-0", "col-md-", 0},
		{"out of range 13", "col-md-13", "col-md-", 0},
		{"non-numeric suffix", "col-md-abc", "col-md-", 0},
		{"empty suffix", "col-md-", "col-md-", 0},
		{"unrelated class", "border", "col-md-", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseColClass(tt.class, tt.prefix)
			if got != tt.want {
				t.Errorf("parseColClass(%q, %q) = %d, want %d", tt.class, tt.prefix, got, tt.want)
			}
		})
	}
}

// ---------- resolveColSpan tests ----------

func TestResolveColSpan(t *testing.T) {
	tests := []struct {
		name    string
		classes []string
		bp      breakpoint
		want    int
	}{
		{"exact match md at md", []string{"col-md-6"}, bpMD, 6},
		{"fallback to sm at md", []string{"col-sm-4"}, bpMD, 4},
		{"fallback to xs at md", []string{"col-6"}, bpMD, 6},
		{"col-lg at md (higher bp ignored)", []string{"col-lg-8"}, bpMD, 0},
		{"multiple breakpoints", []string{"col-sm-12", "col-md-6", "col-lg-4"}, bpMD, 6},
		{"plain col class", []string{"col"}, bpMD, 0},
		{"no col classes", []string{"border", "rounded"}, bpMD, 0},
		{"xl at xl", []string{"col-xl-3"}, bpXL, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveColSpan(tt.classes, tt.bp)
			if got != tt.want {
				t.Errorf("resolveColSpan(%v, %d) = %d, want %d", tt.classes, tt.bp, got, tt.want)
			}
		})
	}
}

// ---------- parseGrid tests ----------

func TestParseGrid_SingleRow(t *testing.T) {
	html := `<div class="row"><div class="col-md-6">A</div><div class="col-md-6">B</div></div>`

	rows, err := parseGrid(html, 80)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if len(rows[0].Columns) != 2 {
		t.Fatalf("expected 2 columns, got %d", len(rows[0].Columns))
	}
}

func TestParseGrid_MultipleRows(t *testing.T) {
	html := `<div class="row"><div class="col-md-12">A</div></div>
<div class="row"><div class="col-md-6">B</div><div class="col-md-6">C</div></div>`

	rows, err := parseGrid(html, 80)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
}

func TestParseGrid_ColumnWidths(t *testing.T) {
	html := `<div class="row"><div class="col-md-4">A</div><div class="col-md-8">B</div></div>`

	rows, err := parseGrid(html, 120)
	if err != nil {
		t.Fatal(err)
	}

	if len(rows) != 1 || len(rows[0].Columns) != 2 {
		t.Fatal("expected 1 row with 2 columns")
	}

	col1Width := rows[0].Columns[0].Width
	col2Width := rows[0].Columns[1].Width

	// col-md-4 = 4/12 * 120 = 40
	if col1Width != 40 {
		t.Errorf("col-md-4 width = %d, want 40", col1Width)
	}
	// col-md-8 = 8/12 * 120 = 80
	if col2Width != 80 {
		t.Errorf("col-md-8 width = %d, want 80", col2Width)
	}
}

func TestParseGrid_EqualColumns(t *testing.T) {
	html := `<div class="row">
<div class="col-md-4">A</div>
<div class="col-md-4">B</div>
<div class="col-md-4">C</div>
</div>`

	rows, err := parseGrid(html, 120)
	if err != nil {
		t.Fatal(err)
	}

	for i, col := range rows[0].Columns {
		if col.Width != 40 {
			t.Errorf("column[%d] width = %d, want 40", i, col.Width)
		}
	}
}

func TestParseGrid_FullWidth(t *testing.T) {
	html := `<div class="row"><div class="col-md-12">Full</div></div>`

	rows, err := parseGrid(html, 80)
	if err != nil {
		t.Fatal(err)
	}

	if rows[0].Columns[0].Width != 80 {
		t.Errorf("col-md-12 width = %d, want 80", rows[0].Columns[0].Width)
	}
}

func TestParseGrid_RowClasses(t *testing.T) {
	html := `<div class="row bg-dark p-2"><div class="col-md-12">A</div></div>`

	rows, err := parseGrid(html, 80)
	if err != nil {
		t.Fatal(err)
	}

	found := false
	for _, c := range rows[0].Classes {
		if c == "bg-dark" {
			found = true
		}
	}
	if !found {
		t.Errorf("row classes should include bg-dark: %v", rows[0].Classes)
	}
}

func TestParseGrid_ColumnClasses(t *testing.T) {
	html := `<div class="row"><div class="col-md-6 border rounded p-1">A</div></div>`

	rows, err := parseGrid(html, 80)
	if err != nil {
		t.Fatal(err)
	}

	classes := rows[0].Columns[0].Classes
	expected := map[string]bool{"col-md-6": true, "border": true, "rounded": true, "p-1": true}
	for _, c := range classes {
		delete(expected, c)
	}
	if len(expected) > 0 {
		t.Errorf("missing classes: %v", expected)
	}
}

func TestParseGrid_NestedRowsSkipped(t *testing.T) {
	html := `<div class="row">
  <div class="col-md-6">
    <div class="row">
      <div class="col-md-12">Nested</div>
    </div>
  </div>
  <div class="col-md-6">Right</div>
</div>`

	rows, err := parseGrid(html, 80)
	if err != nil {
		t.Fatal(err)
	}

	// Only the outer row should be parsed (nested rows are handled recursively during render)
	if len(rows) != 1 {
		t.Errorf("expected 1 top-level row, got %d", len(rows))
	}
}

// ---------- resolveColumnWidths tests ----------

func TestResolveColumnWidths_AutoColumns(t *testing.T) {
	row := gridRow{
		Columns: []gridColumn{
			{Classes: []string{"col"}},
			{Classes: []string{"col"}},
			{Classes: []string{"col"}},
		},
	}

	resolveColumnWidths(&row, 120, bpMD)

	// All auto columns should share the width equally
	for i, col := range row.Columns {
		expected := (120 * 12) / (12 * 3) // = 40
		if col.Width != expected {
			t.Errorf("col[%d].Width = %d, want %d", i, col.Width, expected)
		}
	}
}

func TestResolveColumnWidths_MixedFixedAndAuto(t *testing.T) {
	row := gridRow{
		Columns: []gridColumn{
			{Classes: []string{"col-md-6"}},
			{Classes: []string{"col"}},
		},
	}

	resolveColumnWidths(&row, 120, bpMD)

	// col-md-6 = 60
	if row.Columns[0].Width != 60 {
		t.Errorf("fixed col width = %d, want 60", row.Columns[0].Width)
	}
	// auto col = remaining 6/12 * 120 / 1 = 60
	if row.Columns[1].Width != 60 {
		t.Errorf("auto col width = %d, want 60", row.Columns[1].Width)
	}
}
