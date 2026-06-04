package main_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// localExamples can run without network access.
var localExamples = []string{
	"breakpoints",
	"styling",
	"nested",
	"borders",
	"shadows",
	"image/local",
}

// networkExamples require fetching remote images.
var networkExamples = []string{
	".",
	"image/detect",
	"image/formats",
	"image/grid",
	"image/markdown",
	"image/protocols",
}

func moduleRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Dir(wd)
}

func runExample(t *testing.T, root, example string, timeout time.Duration) {
	t.Helper()

	pkg := filepath.Join("examples", example)
	if example == "." {
		pkg = "examples"
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "run", "./"+pkg+"/")
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "TERMSTRAP_IMAGE_PROTOCOL=halfblock")

	output, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		t.Fatalf("example %s timed out after %s", example, timeout)
	}
	if err != nil {
		t.Fatalf("example %s failed: %v\nOutput:\n%s", example, err, string(output))
	}
	if len(output) == 0 {
		t.Errorf("example %s produced no output", example)
	}
}

func TestLocalExamples(t *testing.T) {
	root := moduleRoot(t)

	for _, ex := range localExamples {
		t.Run(ex, func(t *testing.T) {
			runExample(t, root, ex, 30*time.Second)
		})
	}
}

func TestNetworkExamples(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping network examples in short mode")
	}

	root := moduleRoot(t)

	for _, ex := range networkExamples {
		t.Run(ex, func(t *testing.T) {
			runExample(t, root, ex, 60*time.Second)
		})
	}
}
