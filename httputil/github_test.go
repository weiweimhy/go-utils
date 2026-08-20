package httputil

import (
	"strings"
	"testing"
)

func TestGetGitHubRawURL(t *testing.T) {
	got, err := GetGitHubRawURL("https://github.com/owner/repo/tree/main/path/to/file.pdf")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	want := "https://raw.githubusercontent.com/owner/repo/main/"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestGetGitHubRawURLRedactsInvalidURL(t *testing.T) {
	_, err := GetGitHubRawURL("https://github.com/owner/repo/blob/main/file?access_token=secret")
	if err == nil {
		t.Fatal("expected invalid tree URL error")
	}
	if strings.Contains(err.Error(), "secret") {
		t.Fatalf("invalid URL error leaked secret: %v", err)
	}
}
