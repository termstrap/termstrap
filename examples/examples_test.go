package main_test

import (
	"context"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"testing"
	"time"
)

// networkExamples require fetching remote images.
var networkExamples = []string{
	".",
	"image/detect",
	"image/formats",
	"image/grid",
	"image/markdown",
	"image/protocols",
}

// discoverLocalExamples scans the examples/ directory and automatically detects
// all subdirectories containing a main.go file (excluding network examples).
func discoverLocalExamples(root string) []string {
	netMap := make(map[string]bool)
	for _, n := range networkExamples {
		netMap[n] = true
	}

	examplesDir := filepath.Join(root, "examples")
	var discovered []string

	_ = filepath.WalkDir(examplesDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || !d.IsDir() {
			return nil
		}

		mainFile := filepath.Join(path, "main.go")
		if _, err := os.Stat(mainFile); err == nil {
			rel, err := filepath.Rel(examplesDir, path)
			if err == nil && rel != "." && !netMap[rel] {
				discovered = append(discovered, filepath.ToSlash(rel))
			}
		}
		return nil
	})

	sort.Strings(discovered)
	return discovered
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
	localList := discoverLocalExamples(root)

	for _, ex := range localList {
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
