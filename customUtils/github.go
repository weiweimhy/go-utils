package customUtils

import (
	"fmt"
	"regexp"
)

// GetGitHubRwaUrl 将 GitHub 浏览 URL 转换为 raw 下载 URL
// 例如: https://github.com/owner/repo/tree/master/path/to/file.pdf
// 转换为: https://raw.githubusercontent.com/owner/repo/master/path/to/file.pdf
func GetGitHubRwaUrl(browseUrl string) (string, error) {
	// 匹配 GitHub URL 格式: https://github.com/{owner}/{repo}/tree/{branch}/
	re := regexp.MustCompile(`https://github\.com/([^/]+)/([^/]+)/tree/([^/]+)/?`)
	matches := re.FindStringSubmatch(browseUrl)
	if len(matches) != 4 {
		return "", fmt.Errorf("failed to parse GitHub URL: %s", browseUrl)
	}

	owner := matches[1]
	repo := matches[2]
	branch := matches[3]

	// 构建 raw URL
	rawUrl := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/", owner, repo, branch)
	return rawUrl, nil
}
