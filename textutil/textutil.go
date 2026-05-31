package textutil

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// TruncateRunes truncates s to max runes, including suffix when truncation happens.
func TruncateRunes(s string, max int, suffix string) string {
	if max <= 0 {
		return ""
	}
	if utf8.RuneCountInString(s) <= max {
		return s
	}

	suffixRunes := []rune(suffix)
	if len(suffixRunes) >= max {
		return string(suffixRunes[:max])
	}

	limit := max - len(suffixRunes)
	out := make([]rune, 0, max)
	for _, r := range s {
		if len(out) >= limit {
			break
		}
		out = append(out, r)
	}
	return string(out) + suffix
}

// RemoveControlChars removes Unicode control characters except runes listed in keep.
func RemoveControlChars(s string, keep ...rune) string {
	if s == "" {
		return ""
	}
	allowed := make(map[rune]struct{}, len(keep))
	for _, r := range keep {
		allowed[r] = struct{}{}
	}

	return strings.Map(func(r rune) rune {
		if _, ok := allowed[r]; ok {
			return r
		}
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, s)
}

// SingleLine removes control characters and collapses all whitespace to single spaces.
func SingleLine(s string) string {
	if s == "" {
		return ""
	}
	cleaned := strings.Map(func(r rune) rune {
		if unicode.IsControl(r) || unicode.IsSpace(r) {
			return ' '
		}
		return r
	}, s)
	return strings.Join(strings.Fields(cleaned), " ")
}

// LimitBytesUTF8 truncates s to maxBytes without splitting UTF-8 runes.
func LimitBytesUTF8(s string, maxBytes int, suffix string) string {
	if maxBytes <= 0 {
		return ""
	}
	if len(s) <= maxBytes {
		return s
	}
	if len(suffix) >= maxBytes {
		return trimUTF8Bytes(suffix, maxBytes)
	}

	limit := maxBytes - len(suffix)
	var b strings.Builder
	b.Grow(maxBytes)
	for _, r := range s {
		if b.Len()+utf8.RuneLen(r) > limit {
			break
		}
		b.WriteRune(r)
	}
	b.WriteString(suffix)
	return b.String()
}

func trimUTF8Bytes(s string, maxBytes int) string {
	if maxBytes <= 0 {
		return ""
	}
	var b strings.Builder
	b.Grow(maxBytes)
	for _, r := range s {
		if b.Len()+utf8.RuneLen(r) > maxBytes {
			break
		}
		b.WriteRune(r)
	}
	return b.String()
}
