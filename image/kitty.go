package image

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
)

// kittyRenderer renders images using the Kitty graphics protocol.
// See: https://sw.kovidgoyal.net/kitty/graphics-protocol/
type kittyRenderer struct{}

func (r *kittyRenderer) Protocol() Protocol { return Kitty }

// Render encodes the image as PNG, base64-encodes it, and emits Kitty APC
// graphics sequences. The image is resized to fit the given column width
// (assuming 8px per cell if cell dimensions are unknown).
func (r *kittyRenderer) Render(img image.Image, width int) (string, error) {
	// Resize to approximate pixel width for the given columns
	pxWidth := ColsToPixels(width, 0)
	img = ResizeToWidth(img, pxWidth)

	// Encode as PNG
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", fmt.Errorf("kitty: encode PNG: %w", err)
	}

	// Base64-encode the PNG payload
	payload := base64.StdEncoding.EncodeToString(buf.Bytes())

	// Build chunked APC sequences (max 4096 bytes per chunk)
	return kittyBuildSequence(payload, img.Bounds().Dx(), img.Bounds().Dy(), 0, 0), nil
}

// RenderConstrained renders an image constrained to width columns and height rows.
// The c= and r= placement parameters tell Kitty to fit the image within the
// specified cell dimensions.
func (r *kittyRenderer) RenderConstrained(img image.Image, width, height int) (string, error) {
	pxWidth := ColsToPixels(width, 0)
	img = ResizeToWidth(img, pxWidth)

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", fmt.Errorf("kitty: encode PNG: %w", err)
	}

	payload := base64.StdEncoding.EncodeToString(buf.Bytes())
	return kittyBuildSequence(payload, img.Bounds().Dx(), img.Bounds().Dy(), width, height), nil
}

// kittyChunkSize is the max base64 payload per APC chunk.
const kittyChunkSize = 4096

// kittyBuildSequence constructs the full Kitty graphics protocol output.
// cols and rows are optional display size constraints (0 = unset).
// First chunk: \x1b_Ga=T,f=100,q=2,s=W,v=H[,c=C,r=R],m=1;<chunk>\x1b\\
// Middle chunks: \x1b_Gm=1;<chunk>\x1b\\
// Last chunk: \x1b_Gm=0;<chunk>\x1b\\
func kittyBuildSequence(payload string, w, h, cols, rows int) string {
	var out bytes.Buffer

	chunks := splitPayload(payload, kittyChunkSize)
	total := len(chunks)

	for i, chunk := range chunks {
		out.WriteString("\x1b_G")
		if i == 0 {
			// First chunk: include image metadata
			fmt.Fprintf(&out, "a=T,f=100,q=2,s=%d,v=%d", w, h)
			if cols > 0 {
				fmt.Fprintf(&out, ",c=%d", cols)
			}
			if rows > 0 {
				fmt.Fprintf(&out, ",r=%d", rows)
			}
		}
		if total == 1 {
			// Single chunk: m=0 (no more data)
			if i == 0 {
				out.WriteString(",m=0")
			} else {
				out.WriteString("m=0")
			}
		} else if i < total-1 {
			// More data follows
			if i == 0 {
				out.WriteString(",m=1")
			} else {
				out.WriteString("m=1")
			}
		} else {
			// Last chunk
			out.WriteString("m=0")
		}
		out.WriteByte(';')
		out.WriteString(chunk)
		out.WriteString("\x1b\\")
	}

	out.WriteByte('\n')
	return out.String()
}

// splitPayload splits a string into chunks of at most size bytes.
func splitPayload(s string, size int) []string {
	if len(s) == 0 {
		return []string{""}
	}
	var chunks []string
	for len(s) > 0 {
		end := size
		if end > len(s) {
			end = len(s)
		}
		chunks = append(chunks, s[:end])
		s = s[end:]
	}
	return chunks
}
