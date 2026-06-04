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
