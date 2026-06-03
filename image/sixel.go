package image

import (
	"bytes"
	"fmt"
	"image"

	"github.com/mattn/go-sixel"
)

// sixelRenderer renders images using the Sixel graphics protocol.
// See: https://en.wikipedia.org/wiki/Sixel
type sixelRenderer struct{}

func (r *sixelRenderer) Protocol() Protocol { return Sixel }

// Render encodes the image as Sixel data. The image is resized to fit
// the given column width before encoding.
func (r *sixelRenderer) Render(img image.Image, width int) (string, error) {
	// Resize to approximate pixel width for the given columns
	pxWidth := ColsToPixels(width, 0)
	img = ResizeToWidth(img, pxWidth)

	var buf bytes.Buffer
	enc := sixel.NewEncoder(&buf)
	enc.Dither = true

	if err := enc.Encode(img); err != nil {
		return "", fmt.Errorf("sixel: encode: %w", err)
	}

	return buf.String() + "\n", nil
}
