package termstrap

import (
	"strings"
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

func TestExtractSegments_DeeplyIndentedHTML(t *testing.T) {
	// Test that extractSegments correctly handles deeply indented HTML.
	// The depth counter must accurately track nested divs regardless of indentation.
	input := `# Title

    <div class="row">
      <div class="col-md-6">
        <div class="row">
          <div class="col-md-6">
            Nested content
          </div>
        </div>
      </div>
      <div class="col-md-6">Right</div>
    </div>

Markdown after.`

	result := extractSegments(input)

	// Should have 3 segments: markdown (title), HTML (layout), markdown (after)
	if len(result) != 3 {
		t.Fatalf("expected 3 segments, got %d", len(result))
	}

	// First: markdown
	if result[0].Type != segmentMarkdown {
		t.Errorf("segment[0]: expected markdown, got %d", result[0].Type)
	}
	if !strings.Contains(result[0].Content, "Title") {
		t.Errorf("segment[0]: expected 'Title', got %q", result[0].Content)
	}

	// Second: HTML
	if result[1].Type != segmentHTML {
		t.Errorf("segment[1]: expected HTML, got %d", result[1].Type)
	}
	if !strings.Contains(result[1].Content, "row") {
		t.Errorf("segment[1]: expected 'row' class, got %q", result[1].Content)
	}

	// Third: markdown
	if result[2].Type != segmentMarkdown {
		t.Errorf("segment[2]: expected markdown, got %d", result[2].Type)
	}
	if !strings.Contains(result[2].Content, "after") {
		t.Errorf("segment[2]: expected 'after', got %q", result[2].Content)
	}
}

func TestExtractSegments_ChaosHTML_ErraticIndentation(t *testing.T) {
	// Test with wildly inconsistent indentation: 0, 2, 10, 4 spaces
	input := `Some markdown before

<div class="row">
  <div class="col-md-6">
          <div class="row">
    <div class="col-md-3">
Content A
</div>
          <div class="col-md-3">
               Content B
              </div>
  </div>
</div>
<div class="col-md-6">Right</div>
</div>

Markdown after`

	result := extractSegments(input)

	// Should still correctly identify markdown → HTML → markdown
	if len(result) != 3 {
		t.Fatalf("expected 3 segments, got %d", len(result))
	}

	if result[0].Type != segmentMarkdown {
		t.Errorf("segment[0]: expected markdown, got %d", result[0].Type)
	}
	if result[1].Type != segmentHTML {
		t.Errorf("segment[1]: expected HTML, got %d", result[1].Type)
	}
	if result[2].Type != segmentMarkdown {
		t.Errorf("segment[2]: expected markdown, got %d", result[2].Type)
	}
}

func TestExtractSegments_MixedTabs_Spaces_ChaosFormatting(t *testing.T) {
	// HTML with tabs and spaces mixed, no consistent indentation
	input := `# Title

<div class="row">
	<div class="col-6">
  Column A
	</div>
    <div class="col-6">
		Column B
    </div>
</div>

Final markdown`

	result := extractSegments(input)

	if len(result) != 3 {
		t.Fatalf("expected 3 segments, got %d", len(result))
	}

	for i, seg := range result {
		expected := []segmentType{segmentMarkdown, segmentHTML, segmentMarkdown}
		if seg.Type != expected[i] {
			t.Errorf("segment[%d]: expected type %d, got %d", i, expected[i], seg.Type)
		}
	}
}

func TestExtractSegments_NestedHTMLWithMarkdownContent_Chaos(t *testing.T) {
	// Nested HTML with markdown content inside, all formatted messily
	input := `Start

  <div class="container">
       <div class="row">
<div class="col-4">
## Left Section

Some **bold** text

![image](test.png =20)
</div>
      <div class="col-8">
            ## Right Section
                 More content here
                      with weird spacing
</div>
     </div>
  </div>

End`

	result := extractSegments(input)

	if len(result) != 3 {
		t.Fatalf("expected 3 segments, got %d", len(result))
	}

	if result[0].Type != segmentMarkdown {
		t.Errorf("segment[0]: expected markdown, got %d", result[0].Type)
	}
	if result[1].Type != segmentHTML {
		t.Errorf("segment[1]: expected HTML, got %d", result[1].Type)
	}
	if !strings.Contains(result[1].Content, "col-4") {
		t.Errorf("segment[1]: HTML should contain col-4, got %q", result[1].Content)
	}
	if result[2].Type != segmentMarkdown {
		t.Errorf("segment[2]: expected markdown, got %d", result[2].Type)
	}
}

func TestExtractSegments_ExtremeNesting_RandomIndent(t *testing.T) {
	// 4 levels of nesting with completely random indentation
	input := `# Header

<div class="row">
  <div class="col-12">
<div class="row">
    <div class="col-6">
      <div class="row">
<div class="col-12">
Level 3 content
</div>
</div>
    </div>
  </div>
  </div>
</div>

After HTML`

	result := extractSegments(input)

	if len(result) != 3 {
		t.Fatalf("expected 3 segments, got %d", len(result))
	}

	// Verify the HTML segment contains all the nested content
	if result[1].Type != segmentHTML {
		t.Errorf("segment[1]: expected HTML, got %d", result[1].Type)
	}

	htmlContent := result[1].Content
	if !strings.Contains(htmlContent, "col-12") {
		t.Errorf("HTML should contain col-12, got %q", htmlContent)
	}
	if !strings.Contains(htmlContent, "Level 3 content") {
		t.Errorf("HTML should contain Level 3 content, got %q", htmlContent)
	}
}

func TestExtractSegments_AlternatingMarkdownHTML_WithChaos(t *testing.T) {
	// Multiple alternating markdown and HTML blocks, all chaotically formatted
	input := `# First markdown

<div class="row">
     <div class="col-md-6">
Column content
</div>
</div>

## Second markdown

<div class="container">
  <div class="col-12">
     Nested content
</div>
  </div>

### Final markdown`

	result := extractSegments(input)

	if len(result) != 5 {
		t.Fatalf("expected 5 segments, got %d", len(result))
	}

	// Verify alternating pattern: md, html, md, html, md
	expected := []segmentType{
		segmentMarkdown,
		segmentHTML,
		segmentMarkdown,
		segmentHTML,
		segmentMarkdown,
	}

	for i, seg := range result {
		if seg.Type != expected[i] {
			t.Errorf("segment[%d]: expected type %d, got %d", i, expected[i], seg.Type)
		}
	}
}

func TestParseGrid_ChaosFormatting_WithIndentation(t *testing.T) {
	// Test parseGrid with completely disorganized indentation
	html := `<div class="row">
    <div class="col-md-4">
Left
</div>
  <div class="col-md-4">
     Center
  </div>
       <div class="col-md-4">
Right
         </div>
</div>`

	rows, err := parseGrid(html, 120)
	if err != nil {
		t.Fatal(err)
	}

	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}

	if len(rows[0].Columns) != 3 {
		t.Fatalf("expected 3 columns, got %d", len(rows[0].Columns))
	}

	// All should resolve to col-md-4 (40 width each)
	for i, col := range rows[0].Columns {
		if col.Width != 40 {
			t.Errorf("col[%d].Width: expected 40, got %d", i, col.Width)
		}
	}
}

func TestExtractSegments_CompleteChaosMixedFormat(t *testing.T) {
	// Test with EVERYTHING chaotic:
	// - Multiple indentation levels mixed
	// - Markdown in HTML, HTML in markdown
	// - Tabs, spaces, random whitespace
	// - Multiple nesting levels
	// - Alternating content types
	input := `# Main Title

Some intro text with **bold** and *italic*.

   <div class="container">
  <div class="row">
<div class="col-md-6">
## Left Column

This is **bold markdown** inside HTML.

[Link to nowhere](https://example.com)

	- List item 1
   - List item 2
</div>
      <div class="col-md-6">
         ## Right Column

Another **section** with:

1. Numbered item
2. Second item

			<div class="row">
	<div class="col-6">
Nested deeply
	</div>
</div>
      </div>
  </div>
   </div>

# After HTML

Final markdown with code:

` + "`" + `go
func main() {
    fmt.Println("Hello")
}
` + "`" + ``

	result := extractSegments(input)

	// Should be: markdown title+intro, then HTML block, then markdown after
	if len(result) != 3 {
		t.Fatalf("expected 3 segments, got %d", len(result))
	}

	// All types should be correct
	if result[0].Type != segmentMarkdown {
		t.Errorf("segment[0]: expected markdown, got %d", result[0].Type)
	}
	if result[1].Type != segmentHTML {
		t.Errorf("segment[1]: expected HTML, got %d", result[1].Type)
	}
	if result[2].Type != segmentMarkdown {
		t.Errorf("segment[2]: expected markdown, got %d", result[2].Type)
	}

	// Verify HTML contains all the nested divs
	htmlContent := result[1].Content
	if !strings.Contains(htmlContent, "col-md-6") {
		t.Errorf("HTML should contain col-md-6")
	}
	if !strings.Contains(htmlContent, "col-6") {
		t.Errorf("HTML should contain nested col-6")
	}
	if !strings.Contains(htmlContent, "Left Column") {
		t.Errorf("HTML should contain 'Left Column' markdown content")
	}
	if !strings.Contains(htmlContent, "Nested deeply") {
		t.Errorf("HTML should contain 'Nested deeply' content")
	}

	// Verify markdown segments contain expected content
	if !strings.Contains(result[0].Content, "Main Title") {
		t.Errorf("First markdown should contain 'Main Title'")
	}
	if !strings.Contains(result[2].Content, "After HTML") {
		t.Errorf("Final markdown should contain 'After HTML'")
	}
}

func TestParseGrid_ExtremeMess_MultipleNestingLevels(t *testing.T) {
	// Test parseGrid with EXTREME indentation chaos and deep nesting
	html := `<div class="row">
    <div class="col-md-5">
     A
        <div class="row">
  <div class="col-6">
A1
  </div>
            <div class="col-6">
       A2
            </div>
        </div>
    </div>
  <div class="col-md-3">
B
  </div>
       <div class="col-md-4">
      C
</div>
</div>`

	rows, err := parseGrid(html, 100)
	if err != nil {
		t.Fatal(err)
	}

	// Should have exactly 1 top-level row (nested rows handled recursively)
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}

	// Should have exactly 3 top-level columns
	if len(rows[0].Columns) != 3 {
		t.Fatalf("expected 3 columns, got %d", len(rows[0].Columns))
	}

	// Col widths should be: col-md-5=41, col-md-3=25, col-md-4=33
	expectedWidths := []int{41, 25, 33}
	for i, col := range rows[0].Columns {
		if col.Width != expectedWidths[i] {
			t.Errorf("col[%d].Width: expected %d, got %d", i, expectedWidths[i], col.Width)
		}
	}
}


