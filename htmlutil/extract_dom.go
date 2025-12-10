package htmlutil

import (
	"bytes"
	"fmt"
	"strings"

	"golang.org/x/net/html"
)

// ExtractTextByTagDOM 使用 DOM 解析器提取指定标签的文字内容
func ExtractTextByTagDOM(htmlContent, tagName string) ([]string, error) {
	if htmlContent == "" || tagName == "" {
		return nil, nil
	}

	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
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

	return results, nil
}

// ExtractTextByTagWithAttrDOM 使用 DOM 解析器提取指定标签且包含特定属性的文字内容
func ExtractTextByTagWithAttrDOM(htmlContent, tagName, attrName, attrValue string) ([]string, error) {
	if htmlContent == "" || tagName == "" || attrName == "" {
		return nil, nil
	}

	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
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

			if attrValueFound != "" {
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

	return results, nil
}

// ExtractTextByClassDOM 使用 DOM 解析器提取指定 class 的标签文字内容
func ExtractTextByClassDOM(htmlContent, className string) ([]string, error) {
	if htmlContent == "" || className == "" {
		return nil, nil
	}

	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
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

	return results, nil
}

// ExtractTextByIDDOM 使用 DOM 解析器提取指定 id 的标签文字内容
func ExtractTextByIDDOM(htmlContent, id string) (string, error) {
	if htmlContent == "" || id == "" {
		return "", nil
	}

	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return "", fmt.Errorf("failed to parse HTML: %w", err)
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

	return result, nil
}

// ExtractAllTextDOM 使用 DOM 解析器提取 HTML 中所有文字内容
func ExtractAllTextDOM(htmlContent string) (string, error) {
	if htmlContent == "" {
		return "", nil
	}

	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return "", fmt.Errorf("failed to parse HTML: %w", err)
	}

	var result bytes.Buffer
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

	return result.String(), nil
}

// extractTextFromNode 从节点及其子节点中提取所有文字内容
func extractTextFromNode(n *html.Node) string {
	var result bytes.Buffer
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
