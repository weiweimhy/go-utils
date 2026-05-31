package pathutil

import (
	"fmt"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

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
