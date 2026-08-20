package pathutil

import (
	"errors"
	"os"
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

func TestResolveExistingDescendant(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "nested", "file.txt")
	if err := os.Mkdir(filepath.Dir(target), 0755); err != nil {
		t.Fatalf("Mkdir() error = %v", err)
	}
	if err := os.WriteFile(target, []byte("ok"), 0600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	got, err := ResolveExistingDescendant(root, filepath.Join("nested", "file.txt"), ResolveOptions{})
	if err != nil {
		t.Fatalf("ResolveExistingDescendant() error = %v", err)
	}
	if !SamePath(got, target) {
		t.Fatalf("resolved path = %q, want %q", got, target)
	}
	if _, err := ResolveExistingDescendant(root, root, ResolveOptions{}); !errors.Is(err, ErrNotDescendant) {
		t.Fatalf("root target error = %v, want ErrNotDescendant", err)
	}
	if _, err := ResolveExistingDescendant(root, root, ResolveOptions{AllowRoot: true}); err != nil {
		t.Fatalf("AllowRoot error = %v", err)
	}
}

func TestResolveExistingDescendantRejectsSymlink(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	link := filepath.Join(root, "link")
	if err := os.Symlink(outside, link); err != nil {
		t.Skipf("creating symlink is unavailable: %v", err)
	}

	_, err := ResolveExistingDescendant(root, link, ResolveOptions{})
	if !errors.Is(err, ErrSymlink) {
		t.Fatalf("ResolveExistingDescendant() error = %v, want ErrSymlink", err)
	}
}
