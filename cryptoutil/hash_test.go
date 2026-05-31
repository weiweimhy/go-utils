package cryptoutil

import "testing"

func TestSHA256HexHelpers(t *testing.T) {
	full := SHA256HexFromString("hello")
	if full != "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824" {
		t.Fatalf("SHA256HexFromString() = %q", full)
	}
	if got := SHA256Hex16FromString("hello"); got != "2cf24dba5fb0a30e" {
		t.Fatalf("SHA256Hex16FromString() = %q", got)
	}
	if got := SHA256Hex16FromBytes([]byte("hello")); got != "2cf24dba5fb0a30e" {
		t.Fatalf("SHA256Hex16FromBytes() = %q", got)
	}
}
