package termstrap

import (
	"testing"
)

func TestExtractSegments_PureMarkdown(t *testing.T) {
	input := "# Hello\n\nSome text"
	result := extractSegments(input)

	if len(result) != 1 {
		t.Fatalf("expected 1 segment, got %d", len(result))
	}
	if result[0].Type != segmentMarkdown {
		t.Errorf("expected segmentMarkdown, got %d", result[0].Type)
	}
	if result[0].Content != "# Hello\n\nSome text" {
		t.Errorf("unexpected content: %q", result[0].Content)
	}
}

func TestExtractSegments_SingleRowDiv(t *testing.T) {
	input := "<div class=\"row\">\n<div class=\"col-md-6\">Content</div>\n</div>"
	result := extractSegments(input)

	if len(result) != 1 {
		t.Fatalf("expected 1 segment, got %d", len(result))
	}
	if result[0].Type != segmentHTML {
		t.Errorf("expected segmentHTML, got %d", result[0].Type)
	}
}

func TestExtractSegments_MarkdownBeforeAndAfterHTML(t *testing.T) {
	input := "# Title\n\n<div class=\"row\">\n<div class=\"col-md-6\">Col</div>\n</div>\n\nFooter text"
	result := extractSegments(input)

	if len(result) != 3 {
		t.Fatalf("expected 3 segments, got %d: %+v", len(result), result)
	}
	if result[0].Type != segmentMarkdown {
		t.Errorf("segment[0] expected markdown, got %d", result[0].Type)
	}
	if result[1].Type != segmentHTML {
		t.Errorf("segment[1] expected HTML, got %d", result[1].Type)
	}
	if result[2].Type != segmentMarkdown {
		t.Errorf("segment[2] expected markdown, got %d", result[2].Type)
	}
}

func TestExtractSegments_ContainerClass(t *testing.T) {
	input := "<div class=\"container\">\n<p>Hello</p>\n</div>"
	result := extractSegments(input)

	if len(result) != 1 {
		t.Fatalf("expected 1 segment, got %d", len(result))
	}
	if result[0].Type != segmentHTML {
		t.Errorf("expected segmentHTML, got %d", result[0].Type)
	}
}

func TestExtractSegments_NestedDivs(t *testing.T) {
	input := "<div class=\"row\">\n<div class=\"col-md-6\">\n<div class=\"inner\">Nested</div>\n</div>\n</div>"
	result := extractSegments(input)

	if len(result) != 1 {
		t.Fatalf("expected 1 segment, got %d", len(result))
	}
	if result[0].Type != segmentHTML {
		t.Errorf("expected segmentHTML, got %d", result[0].Type)
	}
}

func TestExtractSegments_EmptyInput(t *testing.T) {
	result := extractSegments("")

	if len(result) != 1 {
		t.Fatalf("expected 1 segment, got %d", len(result))
	}
	if result[0].Type != segmentMarkdown {
		t.Errorf("expected segmentMarkdown, got %d", result[0].Type)
	}
}

func TestExtractSegments_NonLayoutDiv(t *testing.T) {
	input := "<div class=\"something\">Not a layout</div>"
	result := extractSegments(input)

	if len(result) != 1 {
		t.Fatalf("expected 1 segment, got %d", len(result))
	}
	if result[0].Type != segmentMarkdown {
		t.Errorf("expected segmentMarkdown for non-layout div, got %d", result[0].Type)
	}
}

func TestExtractSegments_MultipleHTMLBlocks(t *testing.T) {
	input := "<div class=\"row\">\n<div class=\"col-6\">A</div>\n</div>\n\nMiddle\n\n" +
		"<div class=\"row\">\n<div class=\"col-6\">B</div>\n</div>"
	result := extractSegments(input)

	if len(result) != 3 {
		t.Fatalf("expected 3 segments, got %d: %+v", len(result), result)
	}
	if result[0].Type != segmentHTML {
		t.Errorf("segment[0] expected HTML, got %d", result[0].Type)
	}
	if result[1].Type != segmentMarkdown {
		t.Errorf("segment[1] expected markdown, got %d", result[1].Type)
	}
	if result[1].Content != "Middle" {
		t.Errorf("segment[1] content: got %q, want %q", result[1].Content, "Middle")
	}
	if result[2].Type != segmentHTML {
		t.Errorf("segment[2] expected HTML, got %d", result[2].Type)
	}
}

func TestExtractSegments_UnclosedHTMLBlock(t *testing.T) {
	input := "<div class=\"row\">\n<div class=\"col-6\">Unclosed"
	result := extractSegments(input)

	if len(result) != 1 {
		t.Fatalf("expected 1 segment, got %d", len(result))
	}
	if result[0].Type != segmentHTML {
		t.Errorf("expected segmentHTML for unclosed block, got %d", result[0].Type)
	}
}

func TestExtractSegments_OpenAndCloseOnSameLine(t *testing.T) {
	input := "<div class=\"row\"><div class=\"col-6\">Inline</div></div>"
	result := extractSegments(input)

	if len(result) != 1 {
		t.Fatalf("expected 1 segment, got %d", len(result))
	}
	if result[0].Type != segmentHTML {
		t.Errorf("expected segmentHTML, got %d", result[0].Type)
	}
}

func TestExtractSegments_CombinedContainerRow(t *testing.T) {
	input := "<div class=\"container row\">\n<div class=\"col-12\">X</div>\n</div>"
	result := extractSegments(input)

	if len(result) != 1 {
		t.Fatalf("expected 1 segment, got %d", len(result))
	}
	if result[0].Type != segmentHTML {
		t.Errorf("expected segmentHTML, got %d", result[0].Type)
	}
}

func TestExtractSegments_WhitespaceOnly(t *testing.T) {
	// Whitespace-only input should result in a single empty-ish markdown segment
	result := extractSegments("   \n  \n   ")

	// All whitespace gets trimmed, so we expect the fallback
	if len(result) != 1 {
		t.Fatalf("expected 1 segment, got %d", len(result))
	}
	if result[0].Type != segmentMarkdown {
		t.Errorf("expected segmentMarkdown, got %d", result[0].Type)
	}
}
