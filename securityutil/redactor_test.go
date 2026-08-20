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

func TestRedactURLRemovesUserInfoAndQuerySecrets(t *testing.T) {
	for _, rawURL := range []string{
		"https://client:password@example.com/cb?access_token=secret&ok=1",
		"https://client:password@example.com/%zz?access_token=secret&ok=1",
	} {
		got := RedactURL(rawURL)
		for _, secret := range []string{"client", "password", "secret"} {
			if strings.Contains(got, secret) {
				t.Fatalf("RedactURL(%q) leaked %q in %q", rawURL, secret, got)
			}
		}
		if !strings.Contains(got, "ok=1") {
			t.Fatalf("RedactURL(%q) removed non-secret query data: %q", rawURL, got)
		}
	}
}

func TestRedactURLQueryMalformedURL(t *testing.T) {
	got := RedactURLQuery("https://example.com/%zz?token=secret&ok=1")
	if strings.Contains(got, "secret") || !strings.Contains(got, "ok=1") {
		t.Fatalf("unexpected malformed URL redaction: %q", got)
	}

	got = RedactURLQuery("https://example.com/%zz?private_key=secret", "private_key")
	if strings.Contains(got, "secret") {
		t.Fatalf("custom key leaked from malformed URL: %q", got)
	}

	got = RedactURLQuery("https://example.com/cb?token=secret;state=ok")
	if strings.Contains(got, "secret") {
		t.Fatalf("query parse error leaked a secret: %q", got)
	}
}

func TestDefaultRedactorCoversOAuthKeys(t *testing.T) {
	redactor := DefaultRedactor()
	input := "access_token=access-value refresh_token=refresh-value client_secret=client-value code=authorization-value"
	got := redactor.Redact(input)
	for _, secret := range []string{"access-value", "refresh-value", "client-value", "authorization-value"} {
		if strings.Contains(got, secret) {
			t.Fatalf("Redact() leaked %q in %q", secret, got)
		}
	}
}

func TestRedactLiterals(t *testing.T) {
	got := RedactLiterals("token=abcdef; suffix=cdef", "cdef", "abcdef")
	if strings.Contains(got, "abcdef") || strings.Contains(got, "cdef") {
		t.Fatalf("RedactLiterals() = %q", got)
	}
}
