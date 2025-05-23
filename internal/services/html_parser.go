package services

import (
	"strings"

	"golang.org/x/net/html"
)

// ExtractFirstImageFromHTML finds the first image URL in HTML content
func ExtractFirstImageFromHTML(content string) string {
	if content == "" {
		return ""
	}

	reader := strings.NewReader(content)
	doc, err := html.Parse(reader)
	if err != nil {
		return ""
	}

	var imageURL string
	var findImage func(*html.Node)
	findImage = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "img" {
			// Found an image, try to get its source
			for _, attr := range n.Attr {
				if attr.Key == "src" {
					imageURL = attr.Val
					return
				}
			}
		}

		// Continue searching if no image found yet
		if imageURL == "" {
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				findImage(c)
				if imageURL != "" {
					break
				}
			}
		}
	}

	findImage(doc)
	return imageURL
}

// DecodeHTMLEntities replaces common HTML entities with their characters
func DecodeHTMLEntities(content string) string {
	replacements := map[string]string{
		"&amp;":     "&",
		"&lt;":      "<",
		"&gt;":      ">",
		"&quot;":    "\"",
		"&#39;":     "'",
		"&#8230;":   "…", // Ellipsis
		"[&#8230;]": "[…]",
	}

	result := content
	for entity, char := range replacements {
		result = strings.ReplaceAll(result, entity, char)
	}

	return result
}
