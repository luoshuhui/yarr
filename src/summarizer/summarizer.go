package summarizer

import (
	"context"
	"errors"
	"fmt"
)

var (
	ErrProviderNotConfigured = errors.New("AI provider not configured")
	ErrProviderNotSupported  = errors.New("AI provider not supported")
	ErrContentTooLong        = errors.New("content too long to summarize")
	ErrAPICallFailed         = errors.New("API call failed")
)

// Summarizer defines the interface for AI summarization services
type Summarizer interface {
	Summarize(ctx context.Context, title, content string) (string, error)
}

// Config holds the configuration for creating a summarizer
type Config struct {
	Provider      string
	GeminiAPIKey  string
	OllamaURL     string
	OllamaModel   string
}

// New creates a new Summarizer based on the provider configuration
func New(config Config) (Summarizer, error) {
	switch config.Provider {
	case "gemini":
		if config.GeminiAPIKey == "" {
			return nil, fmt.Errorf("%w: Gemini API key not set", ErrProviderNotConfigured)
		}
		return &GeminiSummarizer{
			APIKey: config.GeminiAPIKey,
		}, nil
	case "ollama":
		if config.OllamaURL == "" {
			return nil, fmt.Errorf("%w: Ollama URL not set", ErrProviderNotConfigured)
		}
		if config.OllamaModel == "" {
			config.OllamaModel = "qwen2:4b"
		}
		return &OllamaSummarizer{
			URL:   config.OllamaURL,
			Model: config.OllamaModel,
		}, nil
	case "disabled", "":
		return nil, fmt.Errorf("%w: AI provider is disabled", ErrProviderNotConfigured)
	default:
		return nil, fmt.Errorf("%w: %s", ErrProviderNotSupported, config.Provider)
	}
}

// buildPrompt constructs the summarization prompt
func buildPrompt(title, content string) string {
	return fmt.Sprintf(`请用中文总结以下文章的主要内容，生成3-5个要点。无论文章是什么语言，总结必须使用中文输出。

文章标题: %s

文章内容: %s

要求:
- 必须使用中文输出总结
- 提取3-5个主要观点
- 每个要点不超过50字
- 重点关注事实，而非观点
- 使用简洁清晰的语言`, title, content)
}

// truncateContent truncates content to a maximum length to avoid API limits
func truncateContent(content string, maxLen int) string {
	if len(content) <= maxLen {
		return content
	}
	return content[:maxLen] + "..."
}
