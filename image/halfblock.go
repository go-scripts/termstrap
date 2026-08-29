package image

import (
	"fmt"
	"image"
	"strings"
)

// halfBlockRenderer renders images using Unicode half-block characters (▄/▀)
// with configurable color depth (TrueColor, 256 colors, 16 colors) and ANSI sequence optimization.
type halfBlockRenderer struct {
	colorMode ColorMode
	optimize  bool
}

// NewHalfBlockRenderer creates a new half-block renderer with specified options.
func NewHalfBlockRenderer(colorMode ColorMode, optimize bool) Renderer {
	return &halfBlockRenderer{
		colorMode: colorMode,
		optimize:  optimize,
	}
}

func (r *halfBlockRenderer) Protocol() Protocol { return HalfBlock }

// Render converts the image to colored half-block characters at the given column width.
func (r *halfBlockRenderer) Render(img image.Image, width int) (string, error) {
	if img == nil {
		return "", fmt.Errorf("halfblock: nil image")
	}
	if width <= 0 {
		return "", fmt.Errorf("halfblock: invalid width %d", width)
	}

	scaled := ResizeToWidth(img, width)
	return r.renderScaled(scaled)
}

// RenderConstrained fits the image within both column width and row height.
func (r *halfBlockRenderer) RenderConstrained(img image.Image, width, height int) (string, error) {
	if img == nil {
		return "", fmt.Errorf("halfblock: nil image")
	}
	if width <= 0 || height <= 0 {
		return "", fmt.Errorf("halfblock: invalid dimensions %dx%d", width, height)
	}

	// Each character row holds 2 vertical pixels
	scaled := ResizeToFit(img, width, height*2)
	return r.renderScaled(scaled)
}

func (r *halfBlockRenderer) renderScaled(scaled image.Image) (string, error) {
	bounds := scaled.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()
	if w == 0 || h == 0 {
		return "", fmt.Errorf("halfblock: rendered empty output")
	}

	var buf strings.Builder
	// Pre-allocate approximate buffer size
	buf.Grow(w * h * 12)

	for y := bounds.Min.Y; y < bounds.Max.Y; y += 2 {
		var lastFG, lastBG string
		hasFG, hasBG := false, false

		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			// Top pixel
			r1, g1, b1, a1 := scaled.At(x, y).RGBA()
			topAlpha := uint8(a1 >> 8)
			topR, topG, topB := uint8(r1>>8), uint8(g1>>8), uint8(b1>>8)

			// Bottom pixel
			var btmAlpha, btmR, btmG, btmB uint8
			if y+1 < bounds.Max.Y {
				r2, g2, b2, a2 := scaled.At(x, y+1).RGBA()
				btmAlpha = uint8(a2 >> 8)
				btmR, btmG, btmB = uint8(r2>>8), uint8(g2>>8), uint8(b2>>8)
			}

			topSolid := topAlpha >= 128
			btmSolid := btmAlpha >= 128

			switch {
			case topSolid && btmSolid:
				fgCode := r.fgEscape(topR, topG, topB)
				bgCode := r.bgEscape(btmR, btmG, btmB)

				if r.optimize {
					if !hasFG || fgCode != lastFG {
						buf.WriteString(fgCode)
						lastFG = fgCode
						hasFG = true
					}
					if !hasBG || bgCode != lastBG {
						buf.WriteString(bgCode)
						lastBG = bgCode
						hasBG = true
					}
				} else {
					buf.WriteString(fgCode)
					buf.WriteString(bgCode)
				}
				buf.WriteString("▀")

			case topSolid && !btmSolid:
				fgCode := r.fgEscape(topR, topG, topB)
				if r.optimize {
					if !hasFG || fgCode != lastFG {
						buf.WriteString(fgCode)
						lastFG = fgCode
						hasFG = true
					}
					if hasBG {
						buf.WriteString("\x1b[49m")
						hasBG = false
						lastBG = ""
					}
				} else {
					buf.WriteString(fgCode)
					buf.WriteString("\x1b[49m")
				}
				buf.WriteString("▀")

			case !topSolid && btmSolid:
				fgCode := r.fgEscape(btmR, btmG, btmB)
				if r.optimize {
					if !hasFG || fgCode != lastFG {
						buf.WriteString(fgCode)
						lastFG = fgCode
						hasFG = true
					}
					if hasBG {
						buf.WriteString("\x1b[49m")
						hasBG = false
						lastBG = ""
					}
				} else {
					buf.WriteString(fgCode)
					buf.WriteString("\x1b[49m")
				}
				buf.WriteString("▄")

			default: // Both transparent
				if r.optimize && hasBG {
					buf.WriteString("\x1b[49m")
					hasBG = false
					lastBG = ""
				}
				buf.WriteString(" ")
			}
		}

		// Reset formatting at end of each line
		buf.WriteString("\x1b[0m")
		if y+2 < bounds.Max.Y {
			buf.WriteString("\n")
		}
	}

	result := buf.String()
	if strings.TrimSpace(result) == "" {
		return "", fmt.Errorf("halfblock: rendered empty output")
	}
	return result, nil
}

func (r *halfBlockRenderer) fgEscape(cr, cg, cb uint8) string {
	switch r.colorMode {
	case ColorMode256:
		idx := RGBTo256(cr, cg, cb)
		return fmt.Sprintf("\x1b[38;5;%dm", idx)
	case ColorMode16:
		idx := RGBTo16(cr, cg, cb)
		if idx < 8 {
			return fmt.Sprintf("\x1b[%dm", 30+idx)
		}
		return fmt.Sprintf("\x1b[%dm", 90+(idx-8))
	default: // TrueColor
		return fmt.Sprintf("\x1b[38;2;%d;%d;%dm", cr, cg, cb)
	}
}

func (r *halfBlockRenderer) bgEscape(cr, cg, cb uint8) string {
	switch r.colorMode {
	case ColorMode256:
		idx := RGBTo256(cr, cg, cb)
		return fmt.Sprintf("\x1b[48;5;%dm", idx)
	case ColorMode16:
		idx := RGBTo16(cr, cg, cb)
		if idx < 8 {
			return fmt.Sprintf("\x1b[%dm", 40+idx)
		}
		return fmt.Sprintf("\x1b[%dm", 100+(idx-8))
	default: // TrueColor
		return fmt.Sprintf("\x1b[48;2;%d;%d;%dm", cr, cg, cb)
	}
}
