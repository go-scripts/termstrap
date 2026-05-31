package termstrap

import (
	"fmt"
)

// ShadowExamples demonstrates the shadow pre-calculation and overflow detection system.
//
// Example usage:
//
//	metrics := calculateShadowMetrics(40, 3, 80)
//	fmt.Printf("Content width: %d\n", metrics.ContentWidth)
//	fmt.Printf("Shadow will overflow: %v\n", metrics.WillOverflow)
//	fmt.Printf("Adjusted shadow size: %d\n", metrics.AdjustedShadow)
//	fmt.Printf("Total width with shadow: %d\n", metrics.TotalWidth)
func ExampleShadowMetrics() {
	// Scenario 1: Shadow fits perfectly
	metrics1 := calculateShadowMetrics(70, 2, 80)
	fmt.Println("=== Scenario 1: Shadow fits ===")
	fmt.Printf("Content width: %d\n", metrics1.ContentWidth)
	fmt.Printf("Requested shadow: 2\n")
	fmt.Printf("Max width: 80\n")
	fmt.Printf("Will overflow: %v\n", metrics1.WillOverflow)
	fmt.Printf("Adjusted shadow: %d\n", metrics1.AdjustedShadow)
	fmt.Printf("Total width: %d\n", metrics1.TotalWidth)
	fmt.Println()

	// Scenario 2: Shadow would overflow - auto-adjusted
	metrics2 := calculateShadowMetrics(76, 3, 80)
	fmt.Println("=== Scenario 2: Shadow overflows - auto-adjusted ===")
	fmt.Printf("Content width: %d\n", metrics2.ContentWidth)
	fmt.Printf("Requested shadow: 3\n")
	fmt.Printf("Max width: 80\n")
	fmt.Printf("Will overflow: %v\n", metrics2.WillOverflow)
	fmt.Printf("Original shadow would be: %d, total: %d\n", 3, 76+3)
	fmt.Printf("Adjusted shadow: %d\n", metrics2.AdjustedShadow)
	fmt.Printf("Total width with adjusted shadow: %d\n", metrics2.TotalWidth)
	fmt.Println()

	// Scenario 3: Very narrow space - minimum shadow of 1
	metrics3 := calculateShadowMetrics(78, 3, 80)
	fmt.Println("=== Scenario 3: Narrow space - minimum shadow ===")
	fmt.Printf("Content width: %d\n", metrics3.ContentWidth)
	fmt.Printf("Requested shadow: 3\n")
	fmt.Printf("Max width: 80\n")
	fmt.Printf("Will overflow: %v\n", metrics3.WillOverflow)
	fmt.Printf("Adjusted shadow (minimum): %d\n", metrics3.AdjustedShadow)
	fmt.Printf("Total width: %d\n", metrics3.TotalWidth)
}

// How to use in your code:
//
// When rendering a column with shadow:
//
//	colStyle := resolveStyle(col.Classes)
//	rendered, err := renderMarkdown(col.Content)
//	ls := applyStyle(colStyle, styleWidth)
//	output := ls.Render(rendered)
//
//	// Pass the available width for smart shadow rendering
//	if colStyle.Shadow > 0 {
//		// applyShadowWithWidth automatically detects and adjusts for overflow
//		output = applyShadowWithWidth(output, colStyle.Shadow, totalWidth)
//	}
//
// Or for row-level shadows:
//
//	if rowStyle.Shadow > 0 {
//		// Use terminal width as the constraint
//		output = applyShadowWithWidth(output, rowStyle.Shadow, m.Width)
//	}
