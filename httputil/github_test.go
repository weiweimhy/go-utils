package httputil

import "testing"

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
