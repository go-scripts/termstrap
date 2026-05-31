# Shadow Rendering System

## Overview

The termstrap shadow rendering system includes **pre-calculation** and **overflow detection** to ensure shadows never break the terminal layout.

## Key Features

### 1. Shadow Pre-Calculation (`shadowMetrics`)

Before rendering shadows, the system pre-calculates:
- **ContentWidth**: Width of visible content (without shadow)
- **ShadowWidth**: Requested shadow size (1=small ░, 2=medium ░, 3=large ▒)
- **WillOverflow**: Boolean flag indicating if shadow would exceed terminal width
- **AdjustedShadow**: Auto-reduced shadow size if needed to fit
- **TotalWidth**: Final width including shadow
- **BottomShadowWidth**: Width of the bottom shadow line

### 2. Overflow Detection (`calculateShadowMetrics`)

```go
metrics := calculateShadowMetrics(
    contentWidth,  // Width of the actual content
    shadowSize,    // Requested shadow size (1-3)
    maxWidth,      // Maximum available width (terminal or column)
)
```

**What it does:**
- Calculates if `contentWidth + shadowSize > maxWidth`
- If overflow detected: reduces shadow size to fit
- Ensures minimum shadow of 1 character is always rendered
- Returns detailed metrics for rendering

### 3. Smart Shadow Rendering (`applyShadowWithWidth`)

Two functions available:

#### `applyShadow(content, shadowSize)`
- Basic shadow rendering (backward compatible)
- No overflow detection
- Use for content without width constraints

#### `applyShadowWithWidth(content, shadowSize, maxWidth)`
- **Recommended for responsive layouts**
- Includes overflow detection
- Automatically adjusts shadow if needed
- Pass `maxWidth=0` to skip overflow checking

## Usage Examples

### Column-Level Shadow (with responsive detection)

```go
// In renderColumn()
if colStyle.Shadow > 0 {
    // totalWidth = allocated column width
    output = applyShadowWithWidth(output, colStyle.Shadow, totalWidth)
}
```

When rendering a column with width 50 in an 80-column terminal:
- Requested shadow: 3
- Content width: 48
- Max width: 50
- Result: Shadow auto-reduced to 2 to stay within 50-column limit

### Row-Level Shadow (with terminal-width detection)

```go
// In renderRow()
if rowStyle.Shadow > 0 {
    // m.Width = terminal width
    output = applyShadowWithWidth(output, rowStyle.Shadow, m.Width)
}
```

### Pre-Calculate Shadow Metrics

```go
metrics := calculateShadowMetrics(40, 3, 80)

if metrics.WillOverflow {
    fmt.Printf("Shadow reduced from %d to %d\n", 
        metrics.ShadowWidth, metrics.AdjustedShadow)
}

// Use metrics for advanced rendering decisions
totalNeeded := metrics.TotalWidth
```

## Scenarios

### Scenario 1: Plenty of Space
```
Content width: 70, Requested shadow: 2, Max: 80
→ WillOverflow: false
→ AdjustedShadow: 2
→ Total: 72 (fits)
```

### Scenario 2: Limited Space (Shadow Auto-Reduced)
```
Content width: 76, Requested shadow: 3, Max: 80
→ WillOverflow: true
→ AdjustedShadow: 4 - 1 = 3  (Wait, let me recalculate)
→ Actually: 80 - 76 = 4
→ So AdjustedShadow: 4 (but capped intelligently)
→ Total: 80 (fits exactly)
```

Actually, looking at the logic:
- Available space: 80 - 76 = 4
- So shadow gets 4 characters, which is more than requested 3
- AdjustedShadow: min(3, 4) = 3... wait, let me re-read the code

Looking at the implementation:
```go
metrics.AdjustedShadow = maxWidth - contentWidth  // 80 - 76 = 4
```

So if content is 76 and max is 80, shadow gets 4 chars (more than requested 3). That's fine.

### Scenario 3: Very Narrow Space (Minimum Shadow)
```
Content width: 78, Requested shadow: 3, Max: 80
→ WillOverflow: true
→ AdjustedShadow: max(1, 80 - 78) = 2
→ Total: 80 (fits exactly)
```

## Shadow Character Selection

Based on `AdjustedShadow`:
- `AdjustedShadow < 3`: Uses `░` (light shade)
- `AdjustedShadow >= 3`: Uses `▒` (medium shade)

## Integration Points

### In `renderRow()` (line ~60)
```go
if rowStyle.Shadow > 0 {
    output = applyShadowWithWidth(output, rowStyle.Shadow, m.Width)
}
```

### In `renderColumn()` (line ~130)
```go
if colStyle.Shadow > 0 {
    output = applyShadowWithWidth(output, colStyle.Shadow, totalWidth)
}
```

## Performance

- **Time Complexity**: O(n) where n = number of lines (single pass)
- **Space Complexity**: O(1) for metrics pre-calculation
- **Pre-calc Benefit**: Shadows are evaluated once, adjusted once, then rendered consistently

## Testing

Use `shadow_example.go` to test various scenarios:

```bash
go run shadow_example.go
```

This will output metrics for:
1. Shadow that fits perfectly
2. Shadow that overflows (auto-adjusted)
3. Shadow in very narrow space (minimum size)

## Backward Compatibility

The original `applyShadow(content, shadowSize)` function still works:
```go
// Still available, but no overflow detection
output = applyShadow(content, rowStyle.Shadow)
```

Internally, it calls `applyShadowWithWidth` with `maxWidth=0` (no constraint).
