package htmlutil

import (
	"regexp"
	"strings"
)

// ExtractTextByTag 从 HTML 中提取指定标签的文字内容
// htmlContent: HTML 内容
// tagName: 标签名称（不包含尖括号，如 "p", "h1", "div"）
// 返回所有匹配标签的文字内容列表
func ExtractTextByTag(htmlContent, tagName string) []string {
	if htmlContent == "" || tagName == "" {
		return nil
	}

	var results []string
	pattern := `<` + regexp.QuoteMeta(tagName) + `[^>]*>(.*?)</` + regexp.QuoteMeta(tagName) + `>`
	re := regexp.MustCompile(`(?i)` + pattern) // 不区分大小写

	matches := re.FindAllStringSubmatch(htmlContent, -1)
	for _, match := range matches {
		if len(match) > 1 {
			text := stripHTMLTags(match[1])
			text = strings.TrimSpace(text)
			if text != "" {
				results = append(results, text)
			}
		}
	}

	return results
}

// ExtractTextByTagWithAttr 从 HTML 中提取指定标签且包含特定属性的文字内容
// htmlContent: HTML 内容
// tagName: 标签名称
// attrName: 属性名称
// attrValue: 属性值（可选，如果为空则只检查属性是否存在）
// 返回所有匹配标签的文字内容列表
func ExtractTextByTagWithAttr(htmlContent, tagName, attrName, attrValue string) []string {
	if htmlContent == "" || tagName == "" || attrName == "" {
		return nil
	}

	var results []string
	var pattern string

	if attrValue == "" {
		pattern = `<` + regexp.QuoteMeta(tagName) + `[^>]*` + regexp.QuoteMeta(attrName) + `\s*=\s*["'][^"']*["'][^>]*>(.*?)</` + regexp.QuoteMeta(tagName) + `>`
	} else {
		pattern = `<` + regexp.QuoteMeta(tagName) + `[^>]*` + regexp.QuoteMeta(attrName) + `\s*=\s*["']` + regexp.QuoteMeta(attrValue) + `["'][^>]*>(.*?)</` + regexp.QuoteMeta(tagName) + `>`
	}

	re := regexp.MustCompile(`(?i)` + pattern) // 不区分大小写

	matches := re.FindAllStringSubmatch(htmlContent, -1)
	for _, match := range matches {
		if len(match) > 1 {
			text := stripHTMLTags(match[1])
			text = strings.TrimSpace(text)
			if text != "" {
				results = append(results, text)
			}
		}
	}

	return results
}

// ExtractTextByClass 从 HTML 中提取指定 class 的标签文字内容
// htmlContent: HTML 内容
// className: CSS class 名称
// 返回所有匹配标签的文字内容列表
func ExtractTextByClass(htmlContent, className string) []string {
	if htmlContent == "" || className == "" {
		return nil
	}

	var results []string
	pattern := `<[^>]+class\s*=\s*["'][^"']*\b` + regexp.QuoteMeta(className) + `\b[^"']*["'][^>]*>(.*?)</[^>]+>`
	re := regexp.MustCompile(`(?i)` + pattern)

	matches := re.FindAllStringSubmatch(htmlContent, -1)
	for _, match := range matches {
		if len(match) > 1 {
			text := stripHTMLTags(match[1])
			text = strings.TrimSpace(text)
			if text != "" {
				results = append(results, text)
			}
		}
	}

	return results
}

// ExtractTextByID 从 HTML 中提取指定 id 的标签文字内容
// htmlContent: HTML 内容
// id: 元素 ID
// 返回匹配标签的文字内容（如果找到），否则返回空字符串
func ExtractTextByID(htmlContent, id string) string {
	if htmlContent == "" || id == "" {
		return ""
	}

	pattern := `<[^>]+id\s*=\s*["']` + regexp.QuoteMeta(id) + `["'][^>]*>(.*?)</[^>]+>`
	re := regexp.MustCompile(`(?i)` + pattern)

	match := re.FindStringSubmatch(htmlContent)
	if len(match) > 1 {
		text := stripHTMLTags(match[1])
		return strings.TrimSpace(text)
	}

	return ""
}

// ExtractAllText 提取 HTML 中所有文字内容（去除所有 HTML 标签）
// htmlContent: HTML 内容
// 返回纯文本内容
func ExtractAllText(htmlContent string) string {
	return stripHTMLTags(htmlContent)
}

// stripHTMLTags 去除 HTML 标签，只保留文字内容
func stripHTMLTags(html string) string {
	re := regexp.MustCompile(`<[^>]+>`)
	text := re.ReplaceAllString(html, " ")

	text = strings.ReplaceAll(text, "&nbsp;", " ")
	text = strings.ReplaceAll(text, "&amp;", "&")
	text = strings.ReplaceAll(text, "&lt;", "<")
	text = strings.ReplaceAll(text, "&gt;", ">")
	text = strings.ReplaceAll(text, "&quot;", "\"")
	text = strings.ReplaceAll(text, "&#39;", "'")

	re = regexp.MustCompile(`\s+`)
	text = re.ReplaceAllString(text, " ")

	return text
}
