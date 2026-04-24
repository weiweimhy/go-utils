package regexputil

import (
	"testing"
)

func TestFindMatches(t *testing.T) {
	str := "order123, order456"
	pattern := `order(\d+)`

	// First call - compiles
	matches := MustFindMatches(str, pattern)
	if len(matches) != 2 || matches[0] != "123" || matches[1] != "456" {
		t.Errorf("MustFindMatches failed: %v", matches)
	}

	matches2 := MustFindMatches(str, pattern)
	if len(matches2) != 2 || matches2[0] != "123" || matches2[1] != "456" {
		t.Errorf("MustFindMatches repeat failed: %v", matches)
	}
}

func TestReplaceAll(t *testing.T) {
	str := "apple123 orange456"
	pattern := `\d+`
	newStr := "NUM"

	result := MustReplaceAll(str, pattern, newStr)
	expected := "appleNUM orangeNUM"
	if result != expected {
		t.Errorf("MustReplaceAll failed, got %s, want %s", result, expected)
	}
}

func TestFindMatchesInvalidPattern(t *testing.T) {
	if matches := MustFindMatches("abc", "["); matches != nil {
		t.Fatalf("expected nil matches for invalid pattern, got %v", matches)
	}

	if _, err := FindMatches("abc", "["); err == nil {
		t.Fatal("expected compile error for invalid pattern")
	}
}

func TestReplaceAllInvalidPattern(t *testing.T) {
	input := "abc123"

	if got := MustReplaceAll(input, "[", "x"); got != input {
		t.Fatalf("expected original input on invalid pattern, got %q", got)
	}

	if _, err := ReplaceAll(input, "[", "x"); err == nil {
		t.Fatal("expected compile error for invalid pattern")
	}
}
