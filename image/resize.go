package image

import (
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	// Register image decoders.
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"

	xdraw "golang.org/x/image/draw"
)

// httpClient is used for fetching remote images with a timeout.
var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

// LoadFromURL fetches an image from a URL and decodes it.
// It validates the Content-Type header and rejects SVG images.
func LoadFromURL(imgURL string) (image.Image, error) {
	req, err := http.NewRequest(http.MethodGet, imgURL, nil)
	if err != nil {
		return nil, fmt.Errorf("fetch image %s: %w", imgURL, err)
	}
	req.Header.Set("User-Agent", "termstrap/1.0")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch image %s: %w", imgURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch image %s: HTTP %d", imgURL, resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		return nil, fmt.Errorf("unexpected content type for %s: %s", imgURL, contentType)
	}
	if strings.Contains(contentType, "svg") {
		return nil, fmt.Errorf("SVG images are not supported: %s", imgURL)
	}

	img, _, err := image.Decode(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("decode image %s: %w", imgURL, err)
	}
	return img, nil
}

// LoadFromPath loads an image from a local file path.
// If path is relative and rootPath is set, it resolves relative to rootPath.
func LoadFromPath(path, rootPath string) (image.Image, error) {
	if rootPath != "" && !IsURL(path) && !strings.HasPrefix(path, "/") {
		path = rootPath + "/" + path
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open image %s: %w", path, err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("decode image %s: %w", path, err)
	}
	return img, nil
}

// ResizeToWidth scales an image to fit the given width in pixels,
// preserving aspect ratio. Uses Catmull-Rom (bicubic) resampling.
// If the image is already narrower than targetWidth, it is returned unchanged.
func ResizeToWidth(img image.Image, targetWidth int) image.Image {
	bounds := img.Bounds()
	origW := bounds.Dx()
	origH := bounds.Dy()

	if origW <= targetWidth {
		return img
	}

	ratio := float64(targetWidth) / float64(origW)
	newH := int(float64(origH) * ratio)
	if newH < 1 {
		newH = 1
	}

	dst := image.NewRGBA(image.Rect(0, 0, targetWidth, newH))
	xdraw.CatmullRom.Scale(dst, dst.Bounds(), img, bounds, draw.Over, nil)
	return dst
}

// ColsToPixels converts terminal column width to approximate pixel width.
// If cellWidth is unknown (0), assumes 8 pixels per cell (common default).
func ColsToPixels(cols, cellWidthPx int) int {
	if cellWidthPx <= 0 {
		cellWidthPx = 8
	}
	return cols * cellWidthPx
}

// defaultCellHeight is the assumed cell height in pixels when not detectable.
const defaultCellHeight = 16

// EstimateVisualHeight estimates how many terminal rows an image will occupy
// when rendered at the given column width using the specified protocol.
func EstimateVisualHeight(img image.Image, widthCols int, proto Protocol) int {
	pxWidth := ColsToPixels(widthCols, 0)
	origW := img.Bounds().Dx()
	origH := img.Bounds().Dy()

	var pxHeight int
	if origW > pxWidth {
		pxHeight = int(float64(origH) * float64(pxWidth) / float64(origW))
	} else {
		pxHeight = origH
	}
	if pxHeight < 1 {
		pxHeight = 1
	}

	switch proto {
	case HalfBlock:
		// Each character cell represents 2 vertical pixels (top ▀ + bottom ▄)
		return (pxHeight + 1) / 2
	default:
		// Native protocols render at pixel level; terminal maps to cell rows
		return (pxHeight + defaultCellHeight - 1) / defaultCellHeight
	}
}

// IsURL checks if a string is a valid absolute URL.
func IsURL(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}

// pipePNG encodes an image as PNG into an io.Pipe. The write side runs in a
// goroutine. Callers must close the returned ReadCloser when done.
func pipePNG(img image.Image) (io.ReadCloser, io.WriteCloser) {
	pr, pw := io.Pipe()
	go func() {
		err := png.Encode(pw, img)
		pw.CloseWithError(err)
	}()
	return pr, pw
}
