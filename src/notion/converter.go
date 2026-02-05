package notion

import (
	"strings"

	"golang.org/x/net/html"
)

const MaxBlockTextLength = 2000

// HTMLToNotionBlocks 将HTML转换为Notion blocks
func HTMLToNotionBlocks(htmlContent string) ([]Block, error) {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return nil, err
	}

	blocks := make([]Block, 0)
	var f func(*html.Node)

	f = func(n *html.Node) {
		if n.Type == html.ElementNode {
			nodeBlocks := convertNodeToBlocks(n)
			blocks = append(blocks, nodeBlocks...)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}

	f(doc)

	// 如果没有生成任何block，至少添加一个空段落
	if len(blocks) == 0 {
		blocks = append(blocks, Block{
			Object: "block",
			Type:   "paragraph",
			Paragraph: &ParagraphBlock{
				RichText: []RichText{
					{
						Text: Text{
							Content: "",
						},
					},
				},
			},
		})
	}

	return blocks, nil
}

// splitTextIntoChunks 将长文本拆分成多个不超过maxLen的片段
func splitTextIntoChunks(text string, maxLen int) []string {
	if len(text) <= maxLen {
		return []string{text}
	}

	chunks := make([]string, 0)
	for len(text) > maxLen {
		chunks = append(chunks, text[:maxLen])
		text = text[maxLen:]
	}
	if len(text) > 0 {
		chunks = append(chunks, text)
	}
	return chunks
}

// convertNodeToBlocks 将HTML节点转换为一个或多个Notion blocks
// 对于超长内容会自动拆分成多个blocks
func convertNodeToBlocks(n *html.Node) []Block {
	text := extractText(n)
	if text == "" {
		return nil
	}

	blocks := make([]Block, 0)

	// 对于标题，如果超长则截断（标题不适合拆分）
	isHeading := n.Data == "h1" || n.Data == "h2" || n.Data == "h3"
	if isHeading && len(text) > MaxBlockTextLength {
		text = text[:MaxBlockTextLength-3] + "..."
	}

	// 对于非标题的内容，如果超长则拆分成多个blocks
	textChunks := []string{text}
	if !isHeading && len(text) > MaxBlockTextLength {
		textChunks = splitTextIntoChunks(text, MaxBlockTextLength)
	}

	// 为每个文本片段创建block
	for _, chunk := range textChunks {
		richText := []RichText{
			{
				Text: Text{
					Content: chunk,
				},
			},
		}

		var block Block
		switch n.Data {
		case "h1":
			block = Block{
				Object: "block",
				Type:   "heading_1",
				Heading1: &HeadingBlock{
					RichText: richText,
				},
			}
		case "h2":
			block = Block{
				Object: "block",
				Type:   "heading_2",
				Heading2: &HeadingBlock{
					RichText: richText,
				},
			}
		case "h3":
			block = Block{
				Object: "block",
				Type:   "heading_3",
				Heading3: &HeadingBlock{
					RichText: richText,
				},
			}
		case "li":
			block = Block{
				Object: "block",
				Type:   "bulleted_list_item",
				BulletedListItem: &BulletedListItemBlock{
					RichText: richText,
				},
			}
		case "blockquote":
			block = Block{
				Object: "block",
				Type:   "quote",
				Quote: &QuoteBlock{
					RichText: richText,
				},
			}
		case "pre", "code":
			block = Block{
				Object: "block",
				Type:   "code",
				Code: &CodeBlock{
					RichText: richText,
					Language: "plain text",
				},
			}
		case "p":
			block = Block{
				Object: "block",
				Type:   "paragraph",
				Paragraph: &ParagraphBlock{
					RichText: richText,
				},
			}
		default:
			// 忽略ul, ol等容器标签，只处理其子元素
			if n.Data == "ul" || n.Data == "ol" || n.Data == "div" {
				return nil
			}
			// 其他文本节点作为段落
			if n.Type == html.TextNode && strings.TrimSpace(n.Data) != "" {
				block = Block{
					Object: "block",
					Type:   "paragraph",
					Paragraph: &ParagraphBlock{
						RichText: richText,
					},
				}
			} else {
				continue
			}
		}
		blocks = append(blocks, block)
	}

	return blocks
}

func extractText(n *html.Node) string {
	if n.Type == html.TextNode {
		return strings.TrimSpace(n.Data)
	}

	// 对于容器标签，不直接提取文本
	if n.Data == "ul" || n.Data == "ol" || n.Data == "div" {
		return ""
	}

	var text strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.TextNode {
			text.WriteString(c.Data)
		} else if c.Type == html.ElementNode {
			text.WriteString(extractText(c))
		}
	}
	return strings.TrimSpace(text.String())
}
