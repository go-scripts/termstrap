package image

import (
	"fmt"
	"image"
	"strings"

	"github.com/stroborobo/aimg"
)

// halfBlockRenderer renders images using Unicode half-block characters (▄/▀)
// with ANSI TrueColor escape sequences. This is the universal fallback that
// works on virtually all modern terminals supporting UTF-8 and 256/TrueColor.
type halfBlockRenderer struct{}

func (r *halfBlockRenderer) Protocol() Protocol { return HalfBlock }

// Render converts the image to colored half-block characters.
// Each character cell represents two vertical pixels, effectively doubling
// the vertical resolution. Width is specified in terminal columns.
func (r *halfBlockRenderer) Render(img image.Image, width int) (string, error) {
	a := aimg.NewImage(width)

	// aimg expects to parse from a reader, but we already have an image.Image.
	// We need to encode it to a temporary buffer so aimg can decode it.
	// Instead, use the lower-level approach: encode to PNG in-memory.
	pr, pw := pipePNG(img)
	defer pr.Close()

	if err := a.ParseReader(pr); err != nil {
		pw.Close()
		return "", fmt.Errorf("halfblock: render: %w", err)
	}

	result := a.String()
	if strings.TrimSpace(result) == "" {
		return "", fmt.Errorf("halfblock: rendered empty output")
	}
	return result, nil
}
