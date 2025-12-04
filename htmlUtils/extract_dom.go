package htmlUtils

import (
	"strings"

	"golang.org/x/net/html"
)

// ExtractTextByTagDOM 使用 DOM 解析器提取指定标签的文字内容
// htmlContent: HTML 内容
// tagName: 标签名称（不包含尖括号，如 "p", "h1", "div"）
// 返回所有匹配标签的文字内容列表
func ExtractTextByTagDOM(htmlContent, tagName string) []string {
	if htmlContent == "" || tagName == "" {
		return nil
	}

	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return nil
	}

	var results []string
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && strings.EqualFold(n.Data, tagName) {
			text := extractTextFromNode(n)
			if text != "" {
				results = append(results, text)
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(doc)

	return results
}

// ExtractTextByTagWithAttrDOM 使用 DOM 解析器提取指定标签且包含特定属性的文字内容
// htmlContent: HTML 内容
// tagName: 标签名称
// attrName: 属性名称
// attrValue: 属性值（可选，如果为空则只检查属性是否存在）
// 返回所有匹配标签的文字内容列表
func ExtractTextByTagWithAttrDOM(htmlContent, tagName, attrName, attrValue string) []string {
	if htmlContent == "" || tagName == "" || attrName == "" {
		return nil
	}

	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return nil
	}

	var results []string
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && strings.EqualFold(n.Data, tagName) {
			// 检查属性
			attrValueFound := ""
			for _, attr := range n.Attr {
				if strings.EqualFold(attr.Key, attrName) {
					attrValueFound = attr.Val
					break
				}
			}

			if attrValueFound != "" {
				// 如果指定了属性值，检查是否匹配
				if attrValue == "" || attrValueFound == attrValue {
					text := extractTextFromNode(n)
					if text != "" {
						results = append(results, text)
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(doc)

	return results
}

// ExtractTextByClassDOM 使用 DOM 解析器提取指定 class 的标签文字内容
// htmlContent: HTML 内容
// className: CSS class 名称
// 返回所有匹配标签的文字内容列表
func ExtractTextByClassDOM(htmlContent, className string) []string {
	if htmlContent == "" || className == "" {
		return nil
	}

	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return nil
	}

	var results []string
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode {
			// 检查 class 属性
			for _, attr := range n.Attr {
				if strings.EqualFold(attr.Key, "class") {
					classes := strings.Fields(attr.Val)
					for _, cls := range classes {
						if cls == className {
							text := extractTextFromNode(n)
							if text != "" {
								results = append(results, text)
							}
							break
						}
					}
					break
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(doc)

	return results
}

// ExtractTextByIDDOM 使用 DOM 解析器提取指定 id 的标签文字内容
// htmlContent: HTML 内容
// id: 元素 ID
// 返回匹配标签的文字内容（如果找到），否则返回空字符串
func ExtractTextByIDDOM(htmlContent, id string) string {
	if htmlContent == "" || id == "" {
		return ""
	}

	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return ""
	}

	var result string
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode {
			// 检查 id 属性
			for _, attr := range n.Attr {
				if strings.EqualFold(attr.Key, "id") && attr.Val == id {
					result = extractTextFromNode(n)
					return
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
			if result != "" {
				return
			}
		}
	}
	traverse(doc)

	return result
}

// ExtractAllTextDOM 使用 DOM 解析器提取 HTML 中所有文字内容
// htmlContent: HTML 内容
// 返回纯文本内容
func ExtractAllTextDOM(htmlContent string) string {
	if htmlContent == "" {
		return ""
	}

	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return ""
	}

	var result strings.Builder
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.TextNode {
			text := strings.TrimSpace(n.Data)
			if text != "" {
				if result.Len() > 0 {
					result.WriteString(" ")
				}
				result.WriteString(text)
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(doc)

	return result.String()
}

// extractTextFromNode 从节点及其子节点中提取所有文字内容
func extractTextFromNode(n *html.Node) string {
	var result strings.Builder
	var traverse func(*html.Node)
	traverse = func(node *html.Node) {
		if node.Type == html.TextNode {
			text := strings.TrimSpace(node.Data)
			if text != "" {
				if result.Len() > 0 {
					result.WriteString(" ")
				}
				result.WriteString(text)
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(n)
	return strings.TrimSpace(result.String())
}
