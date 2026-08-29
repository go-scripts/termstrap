package termstrap

import (
	"strings"
	"testing"

	termimage "github.com/go-scripts/termstrap/image"
)

func TestNestedRowInCol(t *testing.T) {
	html := `<div class="row">
  <div class="col-5">
    <div class="row">
      <div class="col-6">
## Left inner
Test
      </div>
      <div class="col-6">
## Right inner
      </div>
    </div>
  </div>
  <div class="col-3">
## Middle
  </div>
  <div class="col-4">
## Right
  </div>
</div>`

	m := Model{
		Content:       html,
		Width:         120,
		ImageRenderer: termimage.NewRenderer(termimage.WithProtocol(termimage.HalfBlock)),
	}

	out, err := m.Render()
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	plain := stripAllANSI(out)

	// Verify all expected headings appear in the output
	for _, heading := range []string{"Left inner", "Right inner", "Middle", "Right"} {
		if !strings.Contains(plain, heading) {
			t.Errorf("expected heading %q not found in output", heading)
		}
	}

	// Verify "Test" content from inner col appears
	if !strings.Contains(plain, "Test") {
		t.Error("expected 'Test' content not found in output")
	}

	t.Logf("Rendered output:\n%s", out)
}

func TestNestedRowInCol_NativeProtocols(t *testing.T) {
	html := `<div class="row">
  <div class="col-md-5 border p-1">
    <div class="row">
      <div class="col-6 border p-1">
## Inner 1
Content 1
      </div>
      <div class="col-6 border p-1">
## Inner 2
Content 2
      </div>
    </div>
  </div>
  <div class="col-md-7 border p-1">
## Outer Right
Content Right
  </div>
</div>`

	for _, proto := range []termimage.Protocol{termimage.HalfBlock, termimage.ITerm2, termimage.Kitty} {
		t.Run(proto.String(), func(t *testing.T) {
			m := Model{
				Content:       html,
				Width:         100,
				ImageRenderer: termimage.NewRenderer(termimage.WithProtocol(proto)),
			}

			out, err := m.Render()
			if err != nil {
				t.Fatalf("Render failed: %v", err)
			}

			plain := stripAllANSI(out)

			// Ensure HTML tags are parsed and not rendered as raw text
			if strings.Contains(plain, "<div") || strings.Contains(plain, "</div>") {
				t.Errorf("found raw HTML div tags in rendered output for proto %s:\n%s", proto, out)
			}

			for _, heading := range []string{"Inner 1", "Inner 2", "Outer Right"} {
				if !strings.Contains(plain, heading) {
					t.Errorf("expected heading %q not found in output for proto %s", heading, proto)
				}
			}
		})
	}
}

func TestHalfBlockImageIndentation(t *testing.T) {
	dir := t.TempDir()
	imgName := createTestPNG(t, dir, 90, 90)

	html := `<div class="row">
  <div class="col-md-4 border rounded p-1 text-center">

![red](` + imgName + ` =20)

**Red**

  </div>
</div>`

	m := Model{
		Content:       html,
		Width:         100,
		RootPath:      dir,
		ImageRenderer: termimage.NewRenderer(termimage.WithProtocol(termimage.HalfBlock)),
	}

	out, err := m.Render()
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	lines := strings.Split(out, "\n")
	for i, l := range lines {
		t.Logf("line %2d (w=%2d): %s", i, visualWidth(l), l)
	}
}

