package regexputil

import (
	"regexp"
)

func compileRegexp(pattern string) (*regexp.Regexp, error) {
	return regexp.Compile(pattern)
}

// FindMatches returns regexp matches or a compile error for invalid patterns.
func FindMatches(str, pattern string) ([]string, error) {
	reg, err := compileRegexp(pattern)
	if err != nil {
		return nil, err
	}
	matches := reg.FindAllStringSubmatch(str, -1)
	rets := make([]string, len(matches))
	for i, match := range matches {
		if len(match) > 1 {
			rets[i] = match[1]
		} else {
			rets[i] = match[0]
		}
	}
	return rets, nil
}

// MustFindMatches returns regexp matches and hides invalid-pattern errors by returning nil.
func MustFindMatches(str, pattern string) []string {
	rets, err := FindMatches(str, pattern)
	if err != nil {
		return nil
	}
	return rets
}

// ReplaceAll replaces all regexp matches or returns a compile error for invalid patterns.
func ReplaceAll(str, pattern, newStr string) (string, error) {
	reg, err := compileRegexp(pattern)
	if err != nil {
		return "", err
	}
	return reg.ReplaceAllString(str, newStr), nil
}

// MustReplaceAll replaces all regexp matches and hides invalid-pattern errors by returning the input.
func MustReplaceAll(str, pattern, newStr string) string {
	ret, err := ReplaceAll(str, pattern, newStr)
	if err != nil {
		return str
	}
	return ret
}
