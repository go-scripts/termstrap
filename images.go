package termstrap

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/stroborobo/aimg"

	// Register image decoders for image.Decode() used by aimg.
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
)

// imageInfo holds information about a markdown image to render.
type imageInfo struct {
	url   string
	width int // 0 = default (half terminal width)
}

// imageMarkdownRegex matches markdown image syntax with optional width.
// Supports: ![alt](url) and ![alt](url =width)
var imageMarkdownRegex = regexp.MustCompile(`!\[.*?\]\((.*?)(?:\s+=([0-9]+))?\)`)

// extractImages finds all markdown images, replaces them with placeholders,
// and returns the modified content along with a map of placeholder → image info.
func extractImages(content string) (string, map[string]imageInfo) {
	imageMap := make(map[string]imageInfo)
	counter := 0

	newContent := imageMarkdownRegex.ReplaceAllStringFunc(content, func(match string) string {
		submatches := imageMarkdownRegex.FindStringSubmatch(match)
		imgURL := submatches[1]
		var imgWidth int
		if submatches[2] != "" {
			var err error
			imgWidth, err = strconv.Atoi(submatches[2])
			if err != nil {
				log.Printf("Warning: invalid image width '%s', using default. Error: %v", submatches[2], err)
			}
		}
		counter++
		placeholder := fmt.Sprintf("TERMSTRAPIMG%dHOLDER", counter)
		imageMap[placeholder] = imageInfo{url: imgURL, width: imgWidth}
		return placeholder
	})

	return newContent, imageMap
}

// renderImages replaces image placeholders in the rendered content with ANSI art.
func (m Model) renderImages(content string, imageMap map[string]imageInfo) string {
	for placeholder, img := range imageMap {
		imgWidth := m.Width / 2
		if img.width > 0 {
			imgWidth = img.width
		}
		if imgWidth > m.Width {
			imgWidth = m.Width
		}

		var rendered string
		var ok bool

		if isURL(img.url) {
			rendered, ok = renderImageFromURL(img.url, imgWidth)
		} else {
			rendered, ok = renderImageLocal(img.url, m.RootPath, imgWidth)
		}

		content = replacePlaceholder(content, placeholder, rendered, ok)
	}
	return content
}

// replacePlaceholder replaces an image placeholder in rendered content,
// handling ANSI escape codes that glamour may insert within the placeholder text.
func replacePlaceholder(content, placeholder, replacement string, ok bool) string {
	if !ok {
		replacement = ""
	}
	// Try direct replacement first (fastest path)
	if strings.Contains(content, placeholder) {
		return strings.Replace(content, placeholder, replacement, 1)
	}
	// Build a regex that allows ANSI escape codes between each character
	// of the placeholder, since glamour can insert style codes mid-token.
	ansiOpt := `(?:\x1b\[[^m]*m)*`
	var pattern strings.Builder
	pattern.WriteString(ansiOpt)
	for _, ch := range placeholder {
		pattern.WriteString(regexp.QuoteMeta(string(ch)))
		pattern.WriteString(ansiOpt)
	}
	re := regexp.MustCompile(pattern.String())
	return re.ReplaceAllLiteralString(content, replacement)
}

// renderImageFromURL fetches an image from a URL and renders it as ANSI art.
func renderImageFromURL(imgURL string, width int) (string, bool) {
	resp, err := http.Get(imgURL)
	if err != nil {
		log.Printf("Warning: failed to fetch image %s: %v", imgURL, err)
		return "", false
	}
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		log.Printf("Warning: unexpected content type for image %s: %s", imgURL, contentType)
		return "", false
	}
	// SVG images cannot be rendered as ANSI art
	if strings.Contains(contentType, "svg") {
		log.Printf("Warning: SVG images are not supported for ANSI rendering: %s", imgURL)
		return "", false
	}

	img := aimg.NewImage(width)
	if err := img.ParseReader(resp.Body); err != nil {
		log.Printf("Warning: failed to decode image %s: %v", imgURL, err)
		return "", false
	}
	result := img.String()
	if strings.TrimSpace(result) == "" {
		log.Printf("Warning: image rendered as empty for %s", imgURL)
		return "", false
	}
	return addImagePadding(result, 2), true
}

// renderImageLocal loads a local image file and renders it as ANSI art.
func renderImageLocal(path string, rootPath string, width int) (string, bool) {
	if rootPath != "" && !isURL(path) && !strings.HasPrefix(path, "/") {
		path = rootPath + "/" + path
	}

	file, err := os.Open(path)
	if err != nil {
		log.Printf("Warning: failed to open image %s: %v", path, err)
		return "", false
	}
	defer file.Close()

	img := aimg.NewImage(width)
	if err := img.ParseReader(file); err != nil {
		log.Printf("Warning: failed to decode image %s: %v", path, err)
		return "", false
	}
	result := img.String()
	if strings.TrimSpace(result) == "" {
		log.Printf("Warning: image rendered as empty for %s", path)
		return "", false
	}
	return addImagePadding(result, 2), true
}

// addImagePadding adds left padding to each line except the first,
// since the first line inherits indentation from the placeholder position.
func addImagePadding(input string, spaces int) string {
	lines := strings.Split(input, "\n")
	pad := strings.Repeat(" ", spaces)
	for i, line := range lines {
		if i == 0 {
			lines[i] = strings.TrimSpace(line)
		} else {
			lines[i] = pad + strings.TrimSpace(line)
		}
	}
	return strings.Join(lines, "\n")
}

// isURL checks if a string is a valid absolute URL.
func isURL(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}
