package pathutil

import (
	"path/filepath"
	"runtime"
	"testing"
)

func TestIsWithinIncludesRoot(t *testing.T) {
	root := t.TempDir()
	if !IsWithin(root, root) {
		t.Fatal("expected root to be within itself")
	}
	if !IsWithin(filepath.Join(root, "child", "file.txt"), root) {
		t.Fatal("expected child to be within root")
	}
	if IsWithin(filepath.Dir(root), root) {
		t.Fatal("expected parent to be outside root")
	}
}

func TestCleanRelative(t *testing.T) {
	got, err := CleanRelative(filepath.Join(".", "a", "b.txt"))
	if err != nil {
		t.Fatalf("CleanRelative() error = %v", err)
	}
	if got != filepath.Join("a", "b.txt") {
		t.Fatalf("CleanRelative() = %q", got)
	}
	if _, err := CleanRelative(filepath.Join("..", "secret")); err == nil {
		t.Fatal("expected escaping path to fail")
	}
	if _, err := CleanRelative(filepath.Join(string(filepath.Separator), "tmp")); err == nil {
		t.Fatal("expected absolute path to fail")
	}
	if runtime.GOOS == "windows" {
		if _, err := CleanRelative(`C:tmp`); err == nil {
			t.Fatal("expected volume path to fail")
		}
	}
}

func TestFirstMatchedPattern(t *testing.T) {
	pattern, ok := FirstMatchedPattern("src/app/main.go", []string{"*.md", "src/**/*.go"})
	if !ok || pattern != "src/**/*.go" {
		t.Fatalf("match = %q, %v", pattern, ok)
	}
}
