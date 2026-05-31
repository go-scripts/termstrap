package termstrap

// gridColumns is the total number of columns in the Bootstrap grid system.
const gridColumns = 12

// breakpoint represents a responsive breakpoint based on terminal width.
type breakpoint int

const (
	bpXS breakpoint = iota // < 60 cols
	bpSM                   // >= 60 cols
	bpMD                   // >= 80 cols
	bpLG                   // >= 120 cols
	bpXL                   // >= 160 cols
)

// Breakpoint thresholds in terminal character columns.
const (
	thresholdSM = 60
	thresholdMD = 80
	thresholdLG = 120
	thresholdXL = 160
)

// detectBreakpoint returns the current breakpoint based on terminal width.
func detectBreakpoint(termWidth int) breakpoint {
	switch {
	case termWidth >= thresholdXL:
		return bpXL
	case termWidth >= thresholdLG:
		return bpLG
	case termWidth >= thresholdMD:
		return bpMD
	case termWidth >= thresholdSM:
		return bpSM
	default:
		return bpXS
	}
}

// breakpointPrefix returns the CSS class prefix for a given breakpoint.
// E.g., bpMD → "col-md-"
func breakpointPrefix(bp breakpoint) string {
	switch bp {
	case bpXL:
		return "col-xl-"
	case bpLG:
		return "col-lg-"
	case bpMD:
		return "col-md-"
	case bpSM:
		return "col-sm-"
	default:
		return "col-"
	}
}

// isStacked returns true if the terminal is too narrow for the given breakpoint,
// meaning columns should stack vertically instead of side-by-side.
func isStacked(classes []string, bp breakpoint) bool {
	// If any col class is defined for a breakpoint ABOVE current,
	// but none at or below current, columns should stack.
	hasHigher := false
	hasCurrentOrLower := false

	for _, class := range classes {
		for b := bpXS; b <= bpXL; b++ {
			prefix := breakpointPrefix(b)
			if parseColClass(class, prefix) > 0 {
				if b > bp {
					hasHigher = true
				} else {
					hasCurrentOrLower = true
				}
			}
		}
	}

	return hasHigher && !hasCurrentOrLower
}
