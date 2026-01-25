package htmlUtils

import (
	"strings"

	"golang.org/x/net/html"
)

// ExtractTextByTag 使用 DOM 解析器从 HTML 中提取指定标签的文字内容。
// tagName 为标签名称，如 "p", "h1", "div"（不含尖括号）。
func ExtractTextByTag(htmlContent, tagName string) []string {
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

// ExtractTextByTagWithAttr 使用 DOM 解析器提取指定标签且包含特定属性的文字内容。
// tagName 为标签名，attrName 为属性名，attrValue 为属性值（为空则只检查属性是否存在）。
func ExtractTextByTagWithAttr(htmlContent, tagName, attrName, attrValue string) []string {
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
			attrValueFound := ""
			for _, attr := range n.Attr {
				if strings.EqualFold(attr.Key, attrName) {
					attrValueFound = attr.Val
					break
				}
			}

			if attrValueFound != "" || attrValue == "" {
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

// ExtractTextByClass 使用 DOM 解析器提取指定 class 的标签文字内容。
func ExtractTextByClass(htmlContent, className string) []string {
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

// ExtractTextByID 使用 DOM 解析器提取指定 id 的标签文字内容。
func ExtractTextByID(htmlContent, id string) string {
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

// ExtractAllText 提取 HTML 中所有文字内容（去除所有 HTML 标签）。
func ExtractAllText(htmlContent string) string {
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
