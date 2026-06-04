package termstrap

import (
	"strings"
	"testing"
)

// ---------- extractImages tests ----------

func TestExtractImages_NoImages(t *testing.T) {
	content := "# Hello\n\nSome text without images"
	newContent, imageMap := extractImages(content)

	if len(imageMap) != 0 {
		t.Errorf("expected 0 images, got %d", len(imageMap))
	}
	if newContent != content {
		t.Errorf("content should be unchanged")
	}
}

func TestExtractImages_SingleImage(t *testing.T) {
	content := "Before\n\n![alt](https://example.com/img.png)\n\nAfter"
	newContent, imageMap := extractImages(content)

	if len(imageMap) != 1 {
		t.Fatalf("expected 1 image, got %d", len(imageMap))
	}
	if !strings.Contains(newContent, "TERMSTRAPIMG1HOLDER") {
		t.Error("expected placeholder TERMSTRAPIMG1HOLDER in content")
	}
	if strings.Contains(newContent, "![alt]") {
		t.Error("original image markdown should be replaced")
	}

	for _, img := range imageMap {
		if img.url != "https://example.com/img.png" {
			t.Errorf("url = %q, want https://example.com/img.png", img.url)
		}
		if img.width != 0 {
			t.Errorf("width = %d, want 0 (default)", img.width)
		}
	}
}

func TestExtractImages_WithWidth(t *testing.T) {
	content := "![alt](https://example.com/img.png =40)"
	_, imageMap := extractImages(content)

	if len(imageMap) != 1 {
		t.Fatalf("expected 1 image, got %d", len(imageMap))
	}

	for _, img := range imageMap {
		if img.width != 40 {
			t.Errorf("width = %d, want 40", img.width)
		}
	}
}

func TestExtractImages_WithAttributes(t *testing.T) {
	content := "![alt](https://example.com/img.png, width=40, class=rounded, title='Example')"
	_, imageMap := extractImages(content)

	if len(imageMap) != 1 {
		t.Fatalf("expected 1 image, got %d", len(imageMap))
	}

	for _, img := range imageMap {
		if img.url != "https://example.com/img.png" {
			t.Errorf("url = %q, want https://example.com/img.png", img.url)
		}
		if img.width != 40 {
			t.Errorf("width = %d, want 40", img.width)
		}
		if img.attrs["class"] != "rounded" {
			t.Errorf("class = %q, want rounded", img.attrs["class"])
		}
		if img.attrs["title"] != "Example" {
			t.Errorf("title = %q, want Example", img.attrs["title"])
		}
	}
}

func TestExtractImages_MultipleImages(t *testing.T) {
	content := "![a](url1)\n\n![b](url2 =20)\n\n![c](url3 =60)"
	newContent, imageMap := extractImages(content)

	if len(imageMap) != 3 {
		t.Fatalf("expected 3 images, got %d", len(imageMap))
	}
	if !strings.Contains(newContent, "TERMSTRAPIMG1HOLDER") {
		t.Error("missing placeholder 1")
	}
	if !strings.Contains(newContent, "TERMSTRAPIMG2HOLDER") {
		t.Error("missing placeholder 2")
	}
	if !strings.Contains(newContent, "TERMSTRAPIMG3HOLDER") {
		t.Error("missing placeholder 3")
	}
}

func TestExtractImages_LocalPath(t *testing.T) {
	content := "![local](./images/photo.png)"
	_, imageMap := extractImages(content)

	if len(imageMap) != 1 {
		t.Fatalf("expected 1 image, got %d", len(imageMap))
	}

	for _, img := range imageMap {
		if img.url != "./images/photo.png" {
			t.Errorf("url = %q, want ./images/photo.png", img.url)
		}
	}
}

// ---------- replacePlaceholder tests ----------

func TestReplacePlaceholder_DirectMatch(t *testing.T) {
	content := "Before TERMSTRAPIMG1HOLDER After"
	result := replacePlaceholder(content, "TERMSTRAPIMG1HOLDER", "[IMAGE]", true)

	if !strings.Contains(result, "[IMAGE]") {
		t.Error("replacement should be present")
	}
	if strings.Contains(result, "TERMSTRAPIMG1HOLDER") {
		t.Error("placeholder should be replaced")
	}
}

func TestReplacePlaceholder_NotOK(t *testing.T) {
	content := "Before TERMSTRAPIMG1HOLDER After"
	result := replacePlaceholder(content, "TERMSTRAPIMG1HOLDER", "[IMAGE]", false)

	if strings.Contains(result, "[IMAGE]") {
		t.Error("replacement should not be present when ok=false")
	}
	if strings.Contains(result, "TERMSTRAPIMG1HOLDER") {
		t.Error("placeholder should still be removed when ok=false")
	}
}

func TestReplacePlaceholder_WithANSICodes(t *testing.T) {
	// Simulate glamour injecting ANSI codes within the placeholder text
	content := "Before \x1b[38;5;252mTERMSTRAPIMG1HOLDER\x1b[0m After"
	result := replacePlaceholder(content, "TERMSTRAPIMG1HOLDER", "[IMAGE]", true)

	if !strings.Contains(result, "[IMAGE]") {
		t.Error("should handle ANSI-interspersed placeholder")
	}
}

// ---------- findPlaceholderLine tests ----------

func TestFindPlaceholderLine_DirectMatch(t *testing.T) {
	content := "line0\nline1\nTERMSTRAPIMG1HOLDER\nline3"
	line := findPlaceholderLine(content, "TERMSTRAPIMG1HOLDER")

	if line != 2 {
		t.Errorf("findPlaceholderLine = %d, want 2", line)
	}
}

func TestFindPlaceholderLine_FirstLine(t *testing.T) {
	content := "TERMSTRAPIMG1HOLDER\nline1"
	line := findPlaceholderLine(content, "TERMSTRAPIMG1HOLDER")

	if line != 0 {
		t.Errorf("findPlaceholderLine = %d, want 0", line)
	}
}

func TestFindPlaceholderLine_WithANSI(t *testing.T) {
	content := "line0\n\x1b[1mT\x1b[0mE\x1b[38;5;252mRMSTRAPIMG1HOLDER\nline2"
	line := findPlaceholderLine(content, "TERMSTRAPIMG1HOLDER")

	if line != 1 {
		t.Errorf("findPlaceholderLine with ANSI = %d, want 1", line)
	}
}

func TestFindPlaceholderLine_NotFound(t *testing.T) {
	content := "line0\nline1\nline2"
	line := findPlaceholderLine(content, "TERMSTRAPIMG1HOLDER")

	// Should return 0 when not found
	if line != 0 {
		t.Errorf("findPlaceholderLine not found = %d, want 0", line)
	}
}

// ---------- addImagePadding tests ----------

func TestAddImagePadding(t *testing.T) {
	input := "Line1\nLine2\nLine3"
	result := addImagePadding(input, 2)
	lines := strings.Split(result, "\n")

	// First line: no padding (inherits from placeholder position)
	if strings.HasPrefix(lines[0], "  ") {
		t.Errorf("first line should not have padding: %q", lines[0])
	}
	if lines[0] != "Line1" {
		t.Errorf("first line = %q, want %q", lines[0], "Line1")
	}

	// Subsequent lines: 2-space padding
	for i := 1; i < len(lines); i++ {
		if !strings.HasPrefix(lines[i], "  ") {
			t.Errorf("line[%d] should have 2-space padding: %q", i, lines[i])
		}
	}
}

func TestAddImagePadding_SingleLine(t *testing.T) {
	input := "OnlyLine"
	result := addImagePadding(input, 4)
	if result != "OnlyLine" {
		t.Errorf("single line should have no padding: %q", result)
	}
}
