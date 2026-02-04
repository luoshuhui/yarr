package translator

import (
	"context"
	"errors"
	"fmt"
)

var (
	ErrProviderNotConfigured = errors.New("translation provider not configured")
	ErrProviderNotSupported  = errors.New("translation provider not supported")
	ErrContentTooLong        = errors.New("content too long to translate")
	ErrAPICallFailed         = errors.New("API call failed")
)

// Translator defines the interface for translation services
type Translator interface {
	Translate(ctx context.Context, content, targetLang string) (string, error)
}

// Provider represents an AI provider configuration
type Provider struct {
	Name     string `json:"name"`
	Type     string `json:"type"` // gemini, ollama, openai, claude, etc.
	APIKey   string `json:"api_key,omitempty"`
	URL      string `json:"url,omitempty"`      // For Ollama
	Model    string `json:"model,omitempty"`    // For Ollama and other custom models
	Enabled  bool   `json:"enabled"`
}

// Config holds the configuration for creating a translator
type Config struct {
	Provider   string
	APIKey     string
	URL        string
	Model      string
	TargetLang string
}

// New creates a new Translator based on the provider configuration
func New(config Config) (Translator, error) {
	if config.TargetLang == "" {
		config.TargetLang = "zh-CN" // Default to Chinese
	}

	switch config.Provider {
	case "gemini":
		if config.APIKey == "" {
			return nil, fmt.Errorf("%w: Gemini API key not set", ErrProviderNotConfigured)
		}
		return &GeminiTranslator{
			APIKey:     config.APIKey,
			TargetLang: config.TargetLang,
		}, nil
	case "ollama":
		if config.URL == "" {
			return nil, fmt.Errorf("%w: Ollama URL not set", ErrProviderNotConfigured)
		}
		if config.Model == "" {
			config.Model = "qwen2:4b"
		}
		return &OllamaTranslator{
			URL:        config.URL,
			Model:      config.Model,
			TargetLang: config.TargetLang,
		}, nil
	case "google":
		return &GoogleTranslator{
			TargetLang: config.TargetLang,
		}, nil
	case "microsoft":
		if config.APIKey == "" {
			return nil, fmt.Errorf("%w: Microsoft Translator API key not set", ErrProviderNotConfigured)
		}
		return &MicrosoftTranslator{
			APIKey:     config.APIKey,
			TargetLang: config.TargetLang,
		}, nil
	case "disabled", "":
		return nil, fmt.Errorf("%w: translation provider is disabled", ErrProviderNotConfigured)
	default:
		return nil, fmt.Errorf("%w: %s", ErrProviderNotSupported, config.Provider)
	}
}

// buildPrompt constructs the translation prompt for AI-based translation
func buildPrompt(content, targetLang string) string {
	langName := map[string]string{
		"zh-CN": "简体中文",
		"zh-TW": "繁體中文",
		"en":    "English",
		"ja":    "日本語",
		"ko":    "한국어",
		"fr":    "Français",
		"de":    "Deutsch",
		"es":    "Español",
	}

	lang := langName[targetLang]
	if lang == "" {
		lang = targetLang
	}

	return fmt.Sprintf(`请将以下内容翻译成%s。要求：
- 保持原文的格式和结构
- 准确传达原文的意思
- 使用流畅自然的%s表达
- 只返回翻译后的内容，不要添加任何解释或说明

内容：
%s`, lang, lang, content)
}

// truncateContent truncates content to a maximum length to avoid API limits
func truncateContent(content string, maxLen int) string {
	if len(content) <= maxLen {
		return content
	}
	return content[:maxLen] + "..."
}
