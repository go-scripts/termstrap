package termstrap

import (
	"fmt"
	"image"
	"log"
	"regexp"
	"strconv"
	"strings"

	termimage "github.com/go-scripts/termstrap/image"
)

// imageInfo holds information about a markdown image to render.
type imageInfo struct {
	url   string
	width int // 0 = default (half terminal width)
	attrs map[string]string
}

// imageMarkdownRegex matches markdown image syntax with optional width or
// comma-separated attributes.
// Supports:
//
//	![alt](url)
//	![alt](url =40)
//	![alt](url, width=40, class=rounded)
var imageMarkdownRegex = regexp.MustCompile(`!\[.*?\]\(([^)]*)\)`)
var legacyImageWidthRegex = regexp.MustCompile(`\s+=\s*([0-9]+)$`)

func parseMarkdownImageSpec(raw string) (string, map[string]string) {
	raw = strings.TrimSpace(raw)
	attrs := make(map[string]string)

	commaIndex := strings.Index(raw, ",")
	if commaIndex >= 0 {
		imgURL := strings.TrimSpace(raw[:commaIndex])
		attrs = parseImageAttributes(raw[commaIndex+1:])
		return imgURL, attrs
	}

	if matches := legacyImageWidthRegex.FindStringSubmatch(raw); len(matches) == 2 {
		imgURL := strings.TrimSpace(raw[:len(raw)-len(matches[0])])
		attrs["width"] = matches[1]
		return imgURL, attrs
	}

	return raw, attrs
}

func parseImageAttributes(raw string) map[string]string {
	attrs := make(map[string]string)
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}

		key := strings.ToLower(strings.TrimSpace(kv[0]))
		value := strings.TrimSpace(kv[1])
		if len(value) >= 2 {
			if (strings.HasPrefix(value, `"`) && strings.HasSuffix(value, `"`)) ||
				(strings.HasPrefix(value, `'`) && strings.HasSuffix(value, `'`)) {
				value = strings.Trim(value, `"'`)
			}
		}
		attrs[key] = value
	}
	return attrs
}

// extractImages finds all markdown images, replaces them with placeholders,
// and returns the modified content along with a map of placeholder → image info.
func extractImages(content string) (string, map[string]imageInfo) {
	imageMap := make(map[string]imageInfo)
	counter := 0

	newContent := imageMarkdownRegex.ReplaceAllStringFunc(content, func(match string) string {
		submatches := imageMarkdownRegex.FindStringSubmatch(match)
		imgURL, attrs := parseMarkdownImageSpec(submatches[1])
		var imgWidth int
		if widthStr, ok := attrs["width"]; ok && widthStr != "" {
			var err error
			imgWidth, err = strconv.Atoi(widthStr)
			if err != nil {
				log.Printf("Warning: invalid image width '%s', using default. Error: %v", widthStr, err)
			}
		}
		counter++
		placeholder := fmt.Sprintf("TERMSTRAPIMG%dHOLDER", counter)
		imageMap[placeholder] = imageInfo{url: imgURL, width: imgWidth, attrs: attrs}
		return placeholder
	})

	return newContent, imageMap
}

// renderImages replaces image placeholders in the rendered content
// using the auto-detected (or user-configured) graphics protocol.
func (m Model) renderImages(content string, imageMap map[string]imageInfo) string {
	renderer := m.imageRenderer()

	for placeholder, img := range imageMap {
		imgWidth := m.Width / 2
		if img.width > 0 {
			imgWidth = img.width
		}
		if imgWidth > m.Width {
			imgWidth = m.Width
		}

		rendered, ok := m.renderImage(renderer, img.url, imgWidth)
		content = replacePlaceholder(content, placeholder, rendered, ok)
	}
	return content
}

// deferredImage tracks an image that should be rendered with a native
// graphics protocol via cursor-based overlay after text layout is composed.
type deferredImage struct {
	img    image.Image // loaded image data
	width  int         // display width in terminal columns
	height int         // estimated visual height in terminal rows
	row    int         // line offset within the column content (post-glamour)
}

// renderImagesDeferred replaces image placeholders with blank space of the
// correct dimensions for layout, and collects image data for later native
// protocol overlay rendering.
func (m Model) renderImagesDeferred(content string, imageMap map[string]imageInfo, proto termimage.Protocol) (string, []deferredImage) {
	var deferred []deferredImage

	for placeholder, info := range imageMap {
		imgWidth := m.Width / 2
		if info.width > 0 {
			imgWidth = info.width
		}
		// Cap image width to content area minus glamour's paragraph indent (~2 chars).
		// The placeholder replaces text within a glamour-indented line, so the
		// total line width is glamourIndent + imgWidth. This must fit within m.Width.
		maxImgWidth := m.Width - 2
		if maxImgWidth < 1 {
			maxImgWidth = 1
		}
		if imgWidth > maxImgWidth {
			imgWidth = maxImgWidth
		}

		goImg, ok := m.loadImage(info.url)
		if !ok {
			content = replacePlaceholder(content, placeholder, "", false)
			continue
		}

		visualHeight := termimage.EstimateVisualHeight(goImg, imgWidth, proto)
		linesBefore := findPlaceholderLine(content, placeholder)

		// Build blank placeholder: imgWidth columns × visualHeight rows.
		// Do NOT use addImagePadding here — TrimSpace would destroy the blank lines.
		blankLine := strings.Repeat(" ", imgWidth)
		blankLines := make([]string, visualHeight)
		for i := range blankLines {
			blankLines[i] = blankLine
		}
		blank := strings.Join(blankLines, "\n")

		deferred = append(deferred, deferredImage{
			img:    goImg,
			width:  imgWidth,
			height: visualHeight,
			row:    linesBefore,
		})

		content = replacePlaceholder(content, placeholder, blank, true)
	}

	return content, deferred
}

// loadImage loads an image from a URL or local path.
func (m Model) loadImage(src string) (image.Image, bool) {
	var (
		goImg image.Image
		err   error
	)
	if termimage.IsURL(src) {
		goImg, err = termimage.LoadFromURL(src)
	} else {
		goImg, err = termimage.LoadFromPath(src, m.RootPath)
	}
	if err != nil {
		log.Printf("Warning: %v", err)
		return nil, false
	}
	return goImg, true
}

// findPlaceholderLine returns the 0-based line number where a placeholder
// appears in content. Handles ANSI codes injected by glamour.
func findPlaceholderLine(content, placeholder string) int {
	// Fast path: direct match
	if idx := strings.Index(content, placeholder); idx >= 0 {
		return strings.Count(content[:idx], "\n")
	}
	// Slow path: ANSI-tolerant regex
	ansiOpt := `(?:\x1b\[[^m]*m)*`
	var pattern strings.Builder
	pattern.WriteString(ansiOpt)
	for _, ch := range placeholder {
		pattern.WriteString(regexp.QuoteMeta(string(ch)))
		pattern.WriteString(ansiOpt)
	}
	re := regexp.MustCompile(pattern.String())
	loc := re.FindStringIndex(content)
	if loc != nil {
		return strings.Count(content[:loc[0]], "\n")
	}
	return 0
}

// renderImage loads an image from a URL or local path and renders it
// using the given renderer.
func (m Model) renderImage(r termimage.Renderer, src string, width int) (string, bool) {
	var (
		goImg image.Image
		err   error
	)

	if termimage.IsURL(src) {
		goImg, err = termimage.LoadFromURL(src)
	} else {
		goImg, err = termimage.LoadFromPath(src, m.RootPath)
	}
	if err != nil {
		log.Printf("Warning: %v", err)
		return "", false
	}

	rendered, err := r.Render(goImg, width)
	if err != nil {
		log.Printf("Warning: render image %s: %v", src, err)
		return "", false
	}
	if strings.TrimSpace(rendered) == "" {
		log.Printf("Warning: image rendered as empty for %s", src)
		return "", false
	}
	return addImagePadding(rendered, 2), true
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
