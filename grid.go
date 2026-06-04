package termstrap

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// gridColumn represents a single column in a Bootstrap row.
type gridColumn struct {
	Classes []string // All CSS classes on the col element
	Content string   // Inner HTML/markdown content
	Width   int      // Resolved width in characters
}

// gridRow represents a Bootstrap row with its columns.
type gridRow struct {
	Classes []string     // CSS classes on the row element
	Columns []gridColumn // Columns in this row
}

// parseGrid parses an HTML string and extracts the grid structure.
// Supports: <div class="row"><div class="col-*-*">content</div></div>
func parseGrid(htmlContent string, termWidth int) ([]gridRow, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return nil, err
	}

	rows := []gridRow{}

	doc.Find(".row").Each(func(_ int, rowSel *goquery.Selection) {
		// Skip nested rows — they will be rendered recursively by their parent column
		if rowSel.ParentsFiltered(".row").Length() > 0 {
			return
		}

		row := gridRow{}

		classAttr, _ := rowSel.Attr("class")
		row.Classes = splitClasses(classAttr)

		rowSel.Children().Each(func(_ int, colSel *goquery.Selection) {
			col := gridColumn{}

			colClassAttr, _ := colSel.Attr("class")
			col.Classes = splitClasses(colClassAttr)

			// Get inner HTML content
			innerHTML, _ := colSel.Html()
			col.Content = strings.TrimSpace(innerHTML)

			row.Columns = append(row.Columns, col)
		})

		rows = append(rows, row)
	})

	// Resolve column widths based on breakpoint
	bp := detectBreakpoint(termWidth)
	for i := range rows {
		resolveColumnWidths(&rows[i], termWidth, bp)
	}

	return rows, nil
}

// resolveColumnWidths calculates the actual character width for each column
// based on Bootstrap col-*-* classes and the current breakpoint.
func resolveColumnWidths(row *gridRow, termWidth int, bp breakpoint) {
	totalCols := 0
	autoCount := 0

	for i, col := range row.Columns {
		span := resolveColSpan(col.Classes, bp)
		if span > 0 {
			width := (termWidth * span) / gridColumns
			row.Columns[i].Width = width
			totalCols += span
		} else {
			autoCount++
		}
	}

	// Distribute remaining width to auto columns
	if autoCount > 0 {
		remaining := gridColumns - totalCols
		if remaining <= 0 {
			remaining = gridColumns
		}
		autoWidth := (termWidth * remaining) / (gridColumns * autoCount)
		for i, col := range row.Columns {
			if col.Width == 0 {
				row.Columns[i].Width = autoWidth
			}
		}
	}
}

// resolveColSpan determines the column span (1-12) for a column
// based on its classes and the current breakpoint.
// Falls back to smaller breakpoints if the current one isn't specified.
func resolveColSpan(classes []string, bp breakpoint) int {
	// Walk from the current breakpoint DOWN to xs, returning the first matching span.
	// This ensures col-md-6 used on a small screen falls back to matching col-* or col-sm-* or col-xs-* as available.
	for b := bp; b >= bpXS; b-- {
		prefix := breakpointPrefix(b)
		for _, class := range classes {
			if span := parseColClass(class, prefix); span > 0 {
				return span
			}
		}
	}

	// Check for plain "col" class (equal distribution)
	for _, class := range classes {
		if class == "col" {
			return 0 // signals auto-width
		}
	}

	return 0
}

// parseColClass extracts the column span from a class like "col-md-6".
func parseColClass(class string, prefix string) int {
	if !strings.HasPrefix(class, prefix) {
		return 0
	}

	numStr := strings.TrimPrefix(class, prefix)
	span := 0
	for _, c := range numStr {
		if c >= '0' && c <= '9' {
			span = span*10 + int(c-'0')
		} else {
			return 0
		}
	}

	if span >= 1 && span <= gridColumns {
		return span
	}
	return 0
}

// splitClasses splits a class attribute string into individual class names.
func splitClasses(classAttr string) []string {
	parts := strings.Fields(classAttr)
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}
