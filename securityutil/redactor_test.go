package securityutil

import (
	"strings"
	"testing"
)

func TestDefaultRedactor(t *testing.T) {
	input := `Authorization: Bearer abc.def token=secret password: "open"`
	got := DefaultRedactor().Redact(input)
	if strings.Contains(got, "abc.def") || strings.Contains(got, "secret") || strings.Contains(got, "open") {
		t.Fatalf("redaction leaked sensitive data: %q", got)
	}
}

func TestRedactURLQuery(t *testing.T) {
	got := RedactURLQuery("https://example.com/cb?token=abc&ok=1")
	if strings.Contains(got, "abc") || !strings.Contains(got, "ok=1") {
		t.Fatalf("unexpected redacted URL: %q", got)
	}
}
