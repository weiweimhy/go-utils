package textutil

import "testing"

func TestTruncateRunes(t *testing.T) {
	got := TruncateRunes("hello世界", 6, "...")
	if got != "hel..." {
		t.Fatalf("TruncateRunes() = %q", got)
	}
}

func TestRemoveControlChars(t *testing.T) {
	got := RemoveControlChars("a\nb\tc\x00", '\n')
	if got != "a\nbc" {
		t.Fatalf("RemoveControlChars() = %q", got)
	}
}

func TestSingleLine(t *testing.T) {
	got := SingleLine(" hello\n\tworld  ")
	if got != "hello world" {
		t.Fatalf("SingleLine() = %q", got)
	}
}

func TestLimitBytesUTF8(t *testing.T) {
	got := LimitBytesUTF8("你好abc", 7, "...")
	if got != "你..." {
		t.Fatalf("LimitBytesUTF8() = %q", got)
	}
}
