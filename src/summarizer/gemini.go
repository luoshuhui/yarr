package summarizer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	geminiAPIEndpoint = "https://generativelanguage.googleapis.com/v1beta/models/gemini-3-flash-preview:generateContent"
	maxContentLength  = 10000 // Maximum content length to avoid token limits
)

// GeminiSummarizer implements the Summarizer interface for Google Gemini API
type GeminiSummarizer struct {
	APIKey string
}

type geminiRequest struct {
	Contents []geminiContent `json:"contents"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
	} `json:"error,omitempty"`
}

// Summarize generates a summary using Google Gemini API
func (g *GeminiSummarizer) Summarize(ctx context.Context, title, content string) (string, error) {
	// Truncate content if too long
	truncatedContent := truncateContent(content, maxContentLength)

	// Build the prompt
	prompt := buildPrompt(title, truncatedContent)

	// Create request payload
	reqBody := geminiRequest{
		Contents: []geminiContent{
			{
				Parts: []geminiPart{
					{Text: prompt},
				},
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s?key=%s", geminiAPIEndpoint, g.APIKey)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Send request
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrAPICallFailed, err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	var geminiResp geminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for API errors
	if geminiResp.Error != nil {
		return "", fmt.Errorf("%w: %s (code: %d)", ErrAPICallFailed, geminiResp.Error.Message, geminiResp.Error.Code)
	}

	// Extract summary text
	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("%w: no summary generated", ErrAPICallFailed)
	}

	summary := geminiResp.Candidates[0].Content.Parts[0].Text
	if summary == "" {
		return "", fmt.Errorf("%w: empty summary returned", ErrAPICallFailed)
	}

	return summary, nil
}
