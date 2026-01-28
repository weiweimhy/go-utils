package regexputil

import (
	"testing"
)

func TestGetRegexpMatches(t *testing.T) {
	str := "order123, order456"
	pattern := `order(\d+)`

	// First call - compiles
	matches := GetRegexpMatches(str, pattern)
	if len(matches) != 2 || matches[0] != "123" || matches[1] != "456" {
		t.Errorf("GetRegexpMatches failed: %v", matches)
	}

	// Second call - should use cache
	matches2 := GetRegexpMatches(str, pattern)
	if len(matches2) != 2 || matches2[0] != "123" || matches2[1] != "456" {
		t.Errorf("GetRegexpMatches cache failed: %v", matches)
	}
}

func TestRegexpReplaceAll(t *testing.T) {
	str := "apple123 orange456"
	pattern := `\d+`
	newStr := "NUM"

	result := RegexpReplaceAll(str, pattern, newStr)
	expected := "appleNUM orangeNUM"
	if result != expected {
		t.Errorf("RegexpReplaceAll failed, got %s, want %s", result, expected)
	}
}
