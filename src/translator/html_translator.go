package translator

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

// HTMLTranslator handles paragraph-by-paragraph translation of HTML content
type HTMLTranslator struct {
	translator Translator
	ctx        context.Context
	targetLang string
}

// NewHTMLTranslator creates a new HTML translator
func NewHTMLTranslator(translator Translator, ctx context.Context, targetLang string) *HTMLTranslator {
	return &HTMLTranslator{
		translator: translator,
		ctx:        ctx,
		targetLang: targetLang,
	}
}

// TranslateHTML translates HTML content paragraph by paragraph
// Each original paragraph is followed by its translation
func (ht *HTMLTranslator) TranslateHTML(htmlContent string) (string, error) {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		// If HTML parsing fails, fall back to simple text translation
		return ht.translateSimpleText(htmlContent)
	}

	var result strings.Builder
	err = ht.traverseAndTranslate(doc, &result)
	if err != nil {
		return "", err
	}

	return result.String(), nil
}

func (ht *HTMLTranslator) traverseAndTranslate(n *html.Node, result *strings.Builder) error {
	if n.Type == html.ElementNode {
		// Check if this is a code block - don't translate
		if ht.isCodeBlock(n) {
			// Write the original code block without translation
			var buf strings.Builder
			html.Render(&buf, n)
			result.WriteString(buf.String())
			return nil
		}

		// Check if this is a translatable block element (p, div, h1-h6, li, blockquote, etc.)
		if ht.isTranslatableBlock(n) {
			// Extract text content
			text := ht.extractText(n)
			text = strings.TrimSpace(text)

			if text != "" {
				// Write original content
				var originalBuf strings.Builder
				html.Render(&originalBuf, n)
				result.WriteString(originalBuf.String())

				// Translate and write translation
				translated, err := ht.translator.Translate(ht.ctx, text, ht.targetLang)
				if err != nil {
					// If translation fails, skip this paragraph
					return nil
				}

				// Create translation paragraph with same tag but add a class
				translatedHTML := ht.createTranslationElement(n, translated)
				result.WriteString(translatedHTML)
			}
			return nil
		}
	}

	// For other nodes, traverse children
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		err := ht.traverseAndTranslate(c, result)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ht *HTMLTranslator) isCodeBlock(n *html.Node) bool {
	if n.Type != html.ElementNode {
		return false
	}
	codeTags := map[string]bool{
		"pre":  true,
		"code": true,
		"kbd":  true,
		"samp": true,
		"var":  true,
	}
	return codeTags[n.Data]
}

func (ht *HTMLTranslator) isTranslatableBlock(n *html.Node) bool {
	if n.Type != html.ElementNode {
		return false
	}
	blockTags := map[string]bool{
		"p":          true,
		"div":        true,
		"h1":         true,
		"h2":         true,
		"h3":         true,
		"h4":         true,
		"h5":         true,
		"h6":         true,
		"li":         true,
		"blockquote": true,
		"td":         true,
		"th":         true,
		"dd":         true,
		"dt":         true,
		"figcaption": true,
	}
	return blockTags[n.Data]
}

func (ht *HTMLTranslator) extractText(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}
	if ht.isCodeBlock(n) {
		return "" // Don't extract text from code blocks
	}

	var text strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		text.WriteString(ht.extractText(c))
	}
	return text.String()
}

func (ht *HTMLTranslator) createTranslationElement(original *html.Node, translatedText string) string {
	// Get the tag name
	tag := original.Data

	// Copy attributes
	var attrs strings.Builder
	for _, attr := range original.Attr {
		attrs.WriteString(fmt.Sprintf(` %s="%s"`, attr.Key, html.EscapeString(attr.Val)))
	}

	// Add translation class and style
	attrs.WriteString(` class="yarr-translation"`)
	attrs.WriteString(` style="color: #555; margin-top: 0.3em; margin-bottom: 1em; padding-left: 1.2em; border-left: 4px solid #4CAF50 !important; background-color: rgba(76, 175, 80, 0.05);"`)

	return fmt.Sprintf("<%s%s>%s</%s>", tag, attrs.String(), html.EscapeString(translatedText), tag)
}

func (ht *HTMLTranslator) translateSimpleText(text string) (string, error) {
	// Simple fallback: split by double newlines (paragraphs)
	paragraphs := strings.Split(text, "\n\n")
	var result strings.Builder

	for _, para := range paragraphs {
		para = strings.TrimSpace(para)
		if para == "" {
			continue
		}

		// Check if it looks like code (indented or has code markers)
		if ht.looksLikeCode(para) {
			result.WriteString(para)
			result.WriteString("\n\n")
			continue
		}

		// Write original
		result.WriteString(para)
		result.WriteString("\n\n")

		// Translate and write translation
		translated, err := ht.translator.Translate(ht.ctx, para, ht.targetLang)
		if err == nil {
			result.WriteString("[译] ")
			result.WriteString(translated)
			result.WriteString("\n\n")
		}
	}

	return result.String(), nil
}

func (ht *HTMLTranslator) looksLikeCode(text string) bool {
	// Simple heuristic: if line starts with 4 spaces or a tab, or contains many special chars
	lines := strings.Split(text, "\n")
	codeLineCount := 0
	for _, line := range lines {
		if strings.HasPrefix(line, "    ") || strings.HasPrefix(line, "\t") {
			codeLineCount++
		}
		// Check for code-like patterns
		if matched, _ := regexp.MatchString(`^[a-zA-Z_][a-zA-Z0-9_]*\s*[=({]`, line); matched {
			codeLineCount++
		}
	}
	return codeLineCount > len(lines)/2
}
