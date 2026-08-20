package pathutil

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

var (
	// ErrNotDescendant is returned when a path is not a descendant of root.
	ErrNotDescendant = errors.New("pathutil: path is not a descendant of root")
	// ErrSymlink is returned when a checked path segment is a symbolic link.
	ErrSymlink = errors.New("pathutil: symbolic links are not allowed")
	// ErrReparsePoint is returned for Windows reparse points, including junctions.
	ErrReparsePoint = errors.New("pathutil: reparse points are not allowed")
)

// ResolveOptions controls strict existing-path resolution.
type ResolveOptions struct {
	// AllowRoot permits target to resolve to root itself. The default requires a
	// proper descendant.
	AllowRoot bool
}

// SamePath reports whether a and b identify the same cleaned absolute path.
func SamePath(a, b string) bool {
	aa, err := filepath.Abs(filepath.Clean(a))
	if err != nil {
		return false
	}
	bb, err := filepath.Abs(filepath.Clean(b))
	if err != nil {
		return false
	}
	if runtime.GOOS == "windows" {
		return strings.EqualFold(aa, bb)
	}
	return aa == bb
}

// IsWithin reports whether p is root or is contained under root.
func IsWithin(p, root string) bool {
	absPath, err := filepath.Abs(filepath.Clean(p))
	if err != nil {
		return false
	}
	absRoot, err := filepath.Abs(filepath.Clean(root))
	if err != nil {
		return false
	}
	if sameCleanPath(absPath, absRoot) {
		return true
	}
	rel, err := filepath.Rel(absRoot, absPath)
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel)
}

// CleanRelative cleans rel and rejects absolute paths or paths that escape via "..".
func CleanRelative(rel string) (string, error) {
	if rel == "" {
		return "", fmt.Errorf("pathutil: relative path is required")
	}
	if filepath.IsAbs(rel) || filepath.VolumeName(rel) != "" || strings.HasPrefix(rel, `\`) || strings.HasPrefix(rel, `/`) {
		return "", fmt.Errorf("pathutil: absolute paths are not allowed: %q", rel)
	}
	cleaned := filepath.Clean(rel)
	if cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("pathutil: path escapes root: %q", rel)
	}
	return cleaned, nil
}

// ResolveExistingDescendant resolves an existing target beneath root while
// rejecting symbolic links and Windows reparse points in every existing path
// segment. target may be relative to root or absolute. The returned path is an
// absolute resolved path. This check does not eliminate time-of-check to
// time-of-use races; callers still need an appropriate filesystem boundary for
// adversarial concurrent writers.
func ResolveExistingDescendant(root, target string, opts ResolveOptions) (string, error) {
	if root == "" {
		return "", fmt.Errorf("pathutil: root is required")
	}
	if target == "" {
		return "", fmt.Errorf("pathutil: target is required")
	}

	rootAbs, err := filepath.Abs(filepath.Clean(root))
	if err != nil {
		return "", fmt.Errorf("pathutil: resolve root: %w", err)
	}
	if err := validateExistingPathSegment(rootAbs); err != nil {
		return "", fmt.Errorf("pathutil: validate root: %w", err)
	}
	rootInfo, err := os.Lstat(rootAbs)
	if err != nil {
		return "", fmt.Errorf("pathutil: stat root: %w", err)
	}
	if !rootInfo.IsDir() {
		return "", fmt.Errorf("pathutil: root is not a directory: %s", rootAbs)
	}

	candidate := target
	if !filepath.IsAbs(candidate) && filepath.VolumeName(candidate) == "" {
		candidate = filepath.Join(rootAbs, candidate)
	}
	candidateAbs, err := filepath.Abs(filepath.Clean(candidate))
	if err != nil {
		return "", fmt.Errorf("pathutil: resolve target: %w", err)
	}
	if !IsWithin(candidateAbs, rootAbs) {
		return "", fmt.Errorf("%w: %s", ErrNotDescendant, candidateAbs)
	}
	if SamePath(candidateAbs, rootAbs) && !opts.AllowRoot {
		return "", fmt.Errorf("%w: target must not equal root", ErrNotDescendant)
	}

	rel, err := filepath.Rel(rootAbs, candidateAbs)
	if err != nil {
		return "", fmt.Errorf("pathutil: make target relative: %w", err)
	}
	current := rootAbs
	if rel != "." {
		for _, segment := range strings.Split(rel, string(filepath.Separator)) {
			current = filepath.Join(current, segment)
			if err := validateExistingPathSegment(current); err != nil {
				return "", fmt.Errorf("pathutil: validate %s: %w", current, err)
			}
		}
	}

	resolvedRoot, err := filepath.EvalSymlinks(rootAbs)
	if err != nil {
		return "", fmt.Errorf("pathutil: resolve root links: %w", err)
	}
	resolvedTarget, err := filepath.EvalSymlinks(candidateAbs)
	if err != nil {
		return "", fmt.Errorf("pathutil: resolve target links: %w", err)
	}
	if !IsWithin(resolvedTarget, resolvedRoot) {
		return "", fmt.Errorf("%w after resolution: %s", ErrNotDescendant, resolvedTarget)
	}
	if SamePath(resolvedTarget, resolvedRoot) && !opts.AllowRoot {
		return "", fmt.Errorf("%w after resolution: target must not equal root", ErrNotDescendant)
	}
	return resolvedTarget, nil
}

func validateExistingPathSegment(path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return ErrSymlink
	}
	if isReparsePoint(path) {
		return ErrReparsePoint
	}
	return nil
}

// FirstMatchedPattern returns the first pattern matching p.
//
// Patterns use filepath.Match syntax per path segment. A segment equal to "**"
// matches zero or more path segments.
func FirstMatchedPattern(p string, patterns []string) (string, bool) {
	name := cleanSlashPath(p)
	for _, pattern := range patterns {
		if matchPattern(pattern, name) {
			return pattern, true
		}
	}
	return "", false
}

func sameCleanPath(a, b string) bool {
	if runtime.GOOS == "windows" {
		return strings.EqualFold(a, b)
	}
	return a == b
}

func cleanSlashPath(p string) string {
	return path.Clean(filepath.ToSlash(filepath.Clean(p)))
}

func matchPattern(pattern, name string) bool {
	pattern = cleanSlashPath(pattern)
	if !strings.Contains(pattern, "/") {
		ok, _ := path.Match(pattern, path.Base(name))
		return ok
	}
	return matchSegments(strings.Split(pattern, "/"), strings.Split(name, "/"))
}

func matchSegments(pattern, name []string) bool {
	if len(pattern) == 0 {
		return len(name) == 0
	}
	if pattern[0] == "**" {
		if matchSegments(pattern[1:], name) {
			return true
		}
		return len(name) > 0 && matchSegments(pattern, name[1:])
	}
	if len(name) == 0 {
		return false
	}
	ok, err := path.Match(pattern[0], name[0])
	if err != nil || !ok {
		return false
	}
	return matchSegments(pattern[1:], name[1:])
}
