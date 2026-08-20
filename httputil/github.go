package httputil

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/weiweimhy/go-utils/v6/securityutil"
)

// GetGitHubRawURL converts a GitHub tree URL to its raw-content base URL.
func GetGitHubRawURL(browseURL string) (string, error) {
	parsed, err := url.Parse(browseURL)
	if err != nil || parsed.Scheme != "https" || !strings.EqualFold(parsed.Host, "github.com") {
		return "", fmt.Errorf("httputil: invalid GitHub tree URL: %s", securityutil.RedactURL(browseURL))
	}
	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(parts) < 4 || parts[0] == "" || parts[1] == "" || parts[2] != "tree" || parts[3] == "" {
		return "", fmt.Errorf("httputil: invalid GitHub tree URL: %s", securityutil.RedactURL(browseURL))
	}
	owner := parts[0]
	repo := parts[1]
	branch := parts[3]

	return fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/", owner, repo, branch), nil
}
