package idutil

import (
	"strings"
	"testing"
)

func TestFormatAndParseSequence(t *testing.T) {
	id := FormatSequence("INV-", 42, 5)
	if id != "INV-00042" {
		t.Fatalf("FormatSequence() = %q", id)
	}
	got, ok := ParseSequence(id, "INV-")
	if !ok || got != 42 {
		t.Fatalf("ParseSequence() = %d, %v", got, ok)
	}
	if _, ok := ParseSequence("BAD-00042", "INV-"); ok {
		t.Fatal("expected prefix mismatch")
	}
}

func TestRandomHelpers(t *testing.T) {
	hexID, err := RandomHex("tok_", 8)
	if err != nil {
		t.Fatalf("RandomHex() error = %v", err)
	}
	if !strings.HasPrefix(hexID, "tok_") || len(hexID) != len("tok_")+16 {
		t.Fatalf("unexpected hex ID %q", hexID)
	}

	b64ID, err := RandomBase64URL("tok_", 8)
	if err != nil {
		t.Fatalf("RandomBase64URL() error = %v", err)
	}
	if !strings.HasPrefix(b64ID, "tok_") {
		t.Fatalf("unexpected base64 ID %q", b64ID)
	}

	digits, err := RandomDigits(12)
	if err != nil {
		t.Fatalf("RandomDigits() error = %v", err)
	}
	if len(digits) != 12 {
		t.Fatalf("RandomDigits() length = %d, want 12", len(digits))
	}
	for _, digit := range digits {
		if digit < '0' || digit > '9' {
			t.Fatalf("RandomDigits() returned non-digit %q", digit)
		}
	}
}
