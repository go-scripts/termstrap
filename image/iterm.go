package image

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
)

// itermRenderer renders images using the iTerm2 inline image protocol.
// See: https://iterm2.com/documentation-images.html
type itermRenderer struct{}

func (r *itermRenderer) Protocol() Protocol { return ITerm2 }

// Render encodes the image as PNG, base64-encodes it, and emits an iTerm2
// OSC 1337 inline image sequence. The width parameter controls the display
// width in terminal columns.
func (r *itermRenderer) Render(img image.Image, width int) (string, error) {
	// Resize to approximate pixel width for the given columns
	pxWidth := ColsToPixels(width, 0)
	img = ResizeToWidth(img, pxWidth)

	// Encode as PNG
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", fmt.Errorf("iterm2: encode PNG: %w", err)
	}

	payload := base64.StdEncoding.EncodeToString(buf.Bytes())
	size := buf.Len()

	// Build OSC 1337 sequence:
	// \x1b]1337;File=inline=1;size=N;width=Ncols;preserveAspectRatio=1:BASE64\a
	var out bytes.Buffer
	fmt.Fprintf(&out, "\x1b]1337;File=inline=1;size=%d;width=%d;preserveAspectRatio=1:%s\a\n",
		size, width, payload)

	return out.String(), nil
}
