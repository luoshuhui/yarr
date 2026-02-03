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
	return fmt.Sprintf(`Please provide a concise summary of the following article in 3-5 bullet points. Focus on the main ideas and key takeaways.

Title: %s

Content: %s

Requirements:
- Use the same language as the article
- Keep each point under 50 words
- Focus on facts, not opinions`, title, content)
}

// truncateContent truncates content to a maximum length to avoid API limits
func truncateContent(content string, maxLen int) string {
	if len(content) <= maxLen {
		return content
	}
	return content[:maxLen] + "..."
}
