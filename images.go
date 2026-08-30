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
	url       string
	alt       string
	width     int // 0 = default (half terminal width)
	maxWidth  int // 0 = unbounded
	colorMode *termimage.ColorMode
	optimize  *bool
	attrs     map[string]string
}

// imageMarkdownRegex matches markdown image syntax with optional width or
// comma-separated attributes.
// Supports:
//
//	![alt](url)
//	![alt](url =40)
//	![alt](url, width=40, color=256, termstrap-max-width=30)
var imageMarkdownRegex = regexp.MustCompile(`!\[(.*?)\]\(([^)]*)\)`)
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

var htmlImgRegex = regexp.MustCompile(`(?i)<img\s+([^>]+)/?>`)

func parseHTMLImgTag(rawAttrs string) (imageInfo, bool) {
	// Parse HTML attributes
	attrRegex := regexp.MustCompile(`([a-zA-Z0-9_-]+)=(?:"([^"]*)"|'([^']*)'|([^\s>]+))`)
	matches := attrRegex.FindAllStringSubmatch(rawAttrs, -1)
	if len(matches) == 0 {
		return imageInfo{}, false
	}

	attrs := make(map[string]string)
	for _, m := range matches {
		k := strings.ToLower(m[1])
		v := m[2]
		if v == "" {
			v = m[3]
		}
		if v == "" {
			v = m[4]
		}
		attrs[k] = v
	}

	src, ok := attrs["src"]
	if !ok || src == "" {
		return imageInfo{}, false
	}

	info := imageInfo{
		url:   src,
		alt:   attrs["alt"],
		attrs: attrs,
	}

	// Width
	if widthStr, ok := getAttr(attrs, "width", "termstrap-width", "data-termstrap-width"); ok {
		if w, err := strconv.Atoi(widthStr); err == nil {
			info.width = w
		}
	}

	// MaxWidth
	if maxWStr, ok := getAttr(attrs, "max-width", "maxwidth", "termstrap-max-width", "data-termstrap-max-width"); ok {
		if mw, err := strconv.Atoi(maxWStr); err == nil {
			info.maxWidth = mw
		}
	}

	// ColorMode
	if colorStr, ok := getAttr(attrs, "color", "colormode", "termstrap-color", "data-termstrap-color"); ok {
		if cm, ok := termimage.ParseColorMode(colorStr); ok {
			info.colorMode = &cm
		}
	}

	// Optimize
	if optStr, ok := getAttr(attrs, "optimize", "termstrap-optimize", "data-termstrap-optimize"); ok {
		opt := strings.ToLower(optStr) == "true" || optStr == "1"
		info.optimize = &opt
	}

	// Classes
	if classStr, ok := attrs["class"]; ok {
		for _, cls := range strings.Fields(classStr) {
			switch cls {
			case "ansi-256", "ansi256":
				cm := termimage.ColorMode256
				info.colorMode = &cm
			case "ansi-16", "ansi16":
				cm := termimage.ColorMode16
				info.colorMode = &cm
			case "ansi-truecolor", "ansi-tc":
				cm := termimage.ColorModeTrueColor
				info.colorMode = &cm
			}
			if strings.HasPrefix(cls, "max-w-") {
				if mw, err := strconv.Atoi(strings.TrimPrefix(cls, "max-w-")); err == nil {
					info.maxWidth = mw
				}
			} else if strings.HasPrefix(cls, "w-") {
				if w, err := strconv.Atoi(strings.TrimPrefix(cls, "w-")); err == nil {
					info.width = w
				}
			}
		}
	}

	return info, true
}

// extractImages finds all markdown and HTML images, replaces them with placeholders,
// and returns the modified content along with a map of placeholder → image info.
func extractImages(content string) (string, map[string]imageInfo) {
	imageMap := make(map[string]imageInfo)
	counter := 0

	// 1. Process Markdown images
	newContent := imageMarkdownRegex.ReplaceAllStringFunc(content, func(match string) string {
		submatches := imageMarkdownRegex.FindStringSubmatch(match)
		alt := ""
		if len(submatches) >= 2 {
			alt = submatches[1]
		}
		imgURL, attrs := parseMarkdownImageSpec(submatches[2])

		info := imageInfo{
			url:   imgURL,
			alt:   alt,
			attrs: attrs,
		}

		// Width
		if widthStr, ok := getAttr(attrs, "width", "termstrap-width", "data-termstrap-width"); ok {
			if w, err := strconv.Atoi(widthStr); err == nil {
				info.width = w
			}
		}

		// MaxWidth
		if maxWStr, ok := getAttr(attrs, "max-width", "maxwidth", "termstrap-max-width", "data-termstrap-max-width"); ok {
			if mw, err := strconv.Atoi(maxWStr); err == nil {
				info.maxWidth = mw
			}
		}

		// ColorMode
		if colorStr, ok := getAttr(attrs, "color", "colormode", "termstrap-color", "data-termstrap-color"); ok {
			if cm, ok := termimage.ParseColorMode(colorStr); ok {
				info.colorMode = &cm
			}
		}

		// Optimize
		if optStr, ok := getAttr(attrs, "optimize", "termstrap-optimize", "data-termstrap-optimize"); ok {
			opt := strings.ToLower(optStr) == "true" || optStr == "1"
			info.optimize = &opt
		}

		counter++
		placeholder := fmt.Sprintf("TERMSTRAPIMG%dHOLDER", counter)
		imageMap[placeholder] = info
		return placeholder
	})

	// 2. Process HTML <img> tags
	newContent = htmlImgRegex.ReplaceAllStringFunc(newContent, func(match string) string {
		submatches := htmlImgRegex.FindStringSubmatch(match)
		if len(submatches) < 2 {
			return match
		}
		info, ok := parseHTMLImgTag(submatches[1])
		if !ok {
			return match
		}

		counter++
		placeholder := fmt.Sprintf("TERMSTRAPIMG%dHOLDER", counter)
		imageMap[placeholder] = info
		return placeholder
	})

	return newContent, imageMap
}

func getAttr(attrs map[string]string, keys ...string) (string, bool) {
	for _, k := range keys {
		if val, ok := attrs[k]; ok && val != "" {
			return val, true
		}
	}
	return "", false
}

// renderImages replaces image placeholders in the rendered content
// using the auto-detected (or user-configured) graphics protocol.
func (m Model) renderImages(content string, imageMap map[string]imageInfo) string {
	for placeholder, img := range imageMap {
		imgWidth := m.Width
		if m.Width > 60 {
			imgWidth = m.Width / 2
		}
		if img.width > 0 {
			imgWidth = img.width
		}
		if img.maxWidth > 0 && imgWidth > img.maxWidth {
			imgWidth = img.maxWidth
		}
		if m.MaxImageWidth > 0 && imgWidth > m.MaxImageWidth {
			imgWidth = m.MaxImageWidth
		}
		if imgWidth > m.Width {
			imgWidth = m.Width
		}
		if imgWidth < 1 {
			imgWidth = 1
		}

		rendered, ok := m.renderImageWithInfo(img, imgWidth)
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
	col    int         // column offset (terminal cells from left)
}

// renderImagesDeferred replaces image placeholders with blank space of the
// correct dimensions for layout, and collects image data for later native
// protocol overlay rendering.
func (m Model) renderImagesDeferred(content string, imageMap map[string]imageInfo, proto termimage.Protocol) (string, []deferredImage) {
	var deferred []deferredImage

	for placeholder, info := range imageMap {
		imgWidth := m.Width
		if m.Width > 60 {
			imgWidth = m.Width / 2
		}
		if info.width > 0 {
			imgWidth = info.width
		}
		if info.maxWidth > 0 && imgWidth > info.maxWidth {
			imgWidth = info.maxWidth
		}
		if m.MaxImageWidth > 0 && imgWidth > m.MaxImageWidth {
			imgWidth = m.MaxImageWidth
		}
		// Cap image width to content area minus glamour's paragraph indent (~2 chars).
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
			col:    2,
		})

		content = replacePlaceholder(content, placeholder, blank, true)
	}

	return content, deferred
}

func (m Model) resolveImageSource(src string) string {
	if !termimage.IsURL(src) && m.RootPath != "" && !strings.HasPrefix(src, "/") {
		return m.RootPath + "/" + src
	}
	return src
}

// loadImage loads an image from a URL or local path with caching support.
func (m Model) loadImage(src string) (image.Image, bool) {
	resolved := m.resolveImageSource(src)
	cache := m.effectiveCache()
	if m.CachePolicy != CacheReload {
		if img, ok := cache.GetImage(resolved); ok {
			return img, true
		}
	}

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

	if m.CachePolicy != CacheNoStore {
		cache.SetImage(resolved, goImg, m.CacheTTL)
	}
	return goImg, true
}

// findPlaceholderLine returns the 0-based line number where a placeholder
// appears in content. Handles ANSI codes and line wraps injected by glamour.
func findPlaceholderLine(content, placeholder string) int {
	// Fast path: direct match
	if idx := strings.Index(content, placeholder); idx >= 0 {
		return strings.Count(content[:idx], "\n")
	}
	// Slow path: ANSI and line-wrap tolerant regex
	ansiOnly := `(?:\x1b\[[0-9;]*[a-zA-Z])*`
	ansiOrWrap := `(?:\x1b\[[0-9;]*[a-zA-Z]|\r?\n[ \t]{0,8})*`
	var pattern strings.Builder
	pattern.WriteString(ansiOnly)
	for i, ch := range placeholder {
		if i > 0 {
			pattern.WriteString(ansiOrWrap)
		}
		pattern.WriteString(regexp.QuoteMeta(string(ch)))
	}
	pattern.WriteString(ansiOnly)

	re, err := regexp.Compile(pattern.String())
	if err != nil {
		return 0
	}
	loc := re.FindStringIndex(content)
	if loc != nil {
		return strings.Count(content[:loc[0]], "\n")
	}
	return 0
}

// renderImage loads an image from a URL or local path and renders it
// using the given renderer with caching.
func (m Model) renderImage(r termimage.Renderer, src string, width int) (string, bool) {
	info := imageInfo{url: src, width: width}
	return m.renderImageWithInfo(info, width)
}

// renderImageWithInfo renders an image using metadata and caching.
func (m Model) renderImageWithInfo(imgInfo imageInfo, width int) (string, bool) {
	src := imgInfo.url
	if m.DisableImages {
		label := imgInfo.alt
		if label == "" {
			label = "Image"
		}
		return fmt.Sprintf("[%s](%s)", label, src), true
	}

	imgWidth := width
	if imgInfo.width > 0 {
		imgWidth = imgInfo.width
	}
	if imgInfo.maxWidth > 0 && imgWidth > imgInfo.maxWidth {
		imgWidth = imgInfo.maxWidth
	}
	if m.MaxImageWidth > 0 && imgWidth > m.MaxImageWidth {
		imgWidth = m.MaxImageWidth
	}
	if imgWidth > m.Width {
		imgWidth = m.Width
	}
	if imgWidth < 1 {
		imgWidth = 1
	}

	colorMode := m.ColorMode
	if imgInfo.colorMode != nil {
		colorMode = *imgInfo.colorMode
	}

	opt := true
	if m.OptimizeSequences != nil {
		opt = *m.OptimizeSequences
	}
	if imgInfo.optimize != nil {
		opt = *imgInfo.optimize
	}

	resolved := m.resolveImageSource(src)
	cache := m.effectiveCache()
	ansiKey := fmt.Sprintf("%s|w:%d|c:%s|opt:%t", resolved, imgWidth, colorMode, opt)

	if m.CachePolicy != CacheReload {
		if cached, ok := cache.GetANSI(ansiKey); ok {
			return addImagePadding(cached, 2), true
		}
	}

	goImg, ok := m.loadImage(src)
	if !ok {
		return "", false
	}

	renderer := m.imageRenderer()
	if imgInfo.colorMode != nil || imgInfo.optimize != nil {
		renderer = termimage.NewRenderer(
			termimage.WithProtocol(renderer.Protocol()),
			termimage.WithColorMode(colorMode),
			termimage.WithOptimizeSequences(opt),
		)
	}

	rendered, err := renderer.Render(goImg, imgWidth)
	if err != nil {
		log.Printf("Warning: render image %s: %v", src, err)
		return "", false
	}
	if strings.TrimSpace(rendered) == "" {
		log.Printf("Warning: image rendered as empty for %s", src)
		return "", false
	}

	if m.CachePolicy != CacheNoStore {
		cache.SetANSI(ansiKey, rendered, m.CacheTTL)
	}

	return addImagePadding(rendered, 2), true
}

// replacePlaceholder replaces an image placeholder in rendered content,
// handling ANSI escape codes and word wraps that glamour may insert within the placeholder text.
func replacePlaceholder(content, placeholder, replacement string, ok bool) string {
	if !ok {
		replacement = ""
	}
	ansiOnly := `(?:\x1b\[[0-9;]*[a-zA-Z])*`
	ansiOrWrap := `(?:\x1b\[[0-9;]*[a-zA-Z]|\r?\n[ \t]{0,8})*`
	var pattern strings.Builder
	pattern.WriteString(ansiOnly)
	for i, ch := range placeholder {
		if i > 0 {
			pattern.WriteString(ansiOrWrap)
		}
		pattern.WriteString(regexp.QuoteMeta(string(ch)))
	}
	pattern.WriteString(ansiOnly)

	re, err := regexp.Compile(pattern.String())
	var loc []int
	if err == nil {
		loc = re.FindStringIndex(content)
	}
	if loc == nil {
		if idx := strings.Index(content, placeholder); idx >= 0 {
			loc = []int{idx, idx + len(placeholder)}
		} else {
			return content
		}
	}

	// Multi-line replacement (rendered image) on a line by itself
	if strings.Contains(replacement, "\n") {
		lineStart := strings.LastIndex(content[:loc[0]], "\n") + 1
		prefix := content[lineStart:loc[0]]
		// Check if prefix is only spaces / ANSI codes
		if stripANSI(prefix) == "" {
			// Find end of line (consume all trailing spaces and ANSI codes on this line if blank)
			lineEnd := loc[1]
			for lineEnd < len(content) && content[lineEnd] != '\n' && content[lineEnd] != '\r' {
				lineEnd++
			}
			tail := content[loc[1]:lineEnd]
			if stripANSI(tail) != "" {
				// There is non-whitespace text after placeholder on the same line; preserve it
				lineEnd = loc[1]
				for lineEnd < len(content) && (content[lineEnd] == ' ' || content[lineEnd] == '\t') {
					lineEnd++
				}
			}

			// Format each line of replacement with uniform clean indent
			lines := strings.Split(replacement, "\n")
			for i, line := range lines {
				lines[i] = strings.TrimLeft(line, " ")
			}
			return content[:lineStart] + strings.Join(lines, "\n") + content[lineEnd:]
		}
	}

	return content[:loc[0]] + replacement + content[loc[1]:]
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
