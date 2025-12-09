package httputil

import (
	"fmt"
	"regexp"
)

var githubURLRegex = regexp.MustCompile(`https://github\.com/([^/]+)/([^/]+)/tree/([^/]+)/?`)

// GetGitHubRawUrl 将 GitHub 浏览 URL 转换为 raw 下载 URL
func GetGitHubRawUrl(browseUrl string) (string, error) {
	matches := githubURLRegex.FindStringSubmatch(browseUrl)
	if len(matches) != 4 {
		return "", fmt.Errorf("invalid GitHub URL: %s", browseUrl)
	}
	return fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/", matches[1], matches[2], matches[3]), nil
}

// GetGitHubRwaUrl 已废弃，请使用 GetGitHubRawUrl
// Deprecated: Use GetGitHubRawUrl instead
func GetGitHubRwaUrl(browseUrl string) (string, error) {
	return GetGitHubRawUrl(browseUrl)
}

