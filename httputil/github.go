package httputil

import (
	"fmt"
	"regexp"
)

// GetGitHubRawURL 将 GitHub 浏览 URL 转换为 raw 下载 URL。
func GetGitHubRawURL(browseURL string) (string, error) {
	// 匹配 GitHub URL 格式: https://github.com/{owner}/{repo}/tree/{branch}/
	re := regexp.MustCompile(`https://github\.com/([^/]+)/([^/]+)/tree/([^/]+)/?`)
	matches := re.FindStringSubmatch(browseURL)
	if len(matches) != 4 {
		return "", fmt.Errorf("failed to parse GitHub URL: %s", browseURL)
	}

	owner := matches[1]
	repo := matches[2]
	branch := matches[3]

	// 构建 raw URL
	rawUrl := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/", owner, repo, branch)
	return rawUrl, nil
}

// GetGitHubRwaUrl 保留旧名称以兼容历史调用。
//
// Deprecated: use GetGitHubRawURL instead.
func GetGitHubRwaUrl(browseURL string) (string, error) {
	return GetGitHubRawURL(browseURL)
}
