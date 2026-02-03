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

// OllamaSummarizer implements the Summarizer interface for Ollama local API
type OllamaSummarizer struct {
	URL   string
	Model string
}

type ollamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type ollamaResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
	Error    string `json:"error,omitempty"`
}

// Summarize generates a summary using Ollama local API
func (o *OllamaSummarizer) Summarize(ctx context.Context, title, content string) (string, error) {
	// Truncate content if too long
	truncatedContent := truncateContent(content, maxContentLength)

	// Build the prompt
	prompt := buildPrompt(title, truncatedContent)

	// Create request payload
	reqBody := ollamaRequest{
		Model:  o.Model,
		Prompt: prompt,
		Stream: false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/api/generate", o.URL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Send request with longer timeout for local models
	client := &http.Client{
		Timeout: 120 * time.Second, // Longer timeout for local inference
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Ollama service unavailable: %w", err)
	}
	defer resp.Body.Close()

	// Check for HTTP errors
	if resp.StatusCode == http.StatusNotFound {
		return "", fmt.Errorf("Model not found: %s", o.Model)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%w: HTTP %d", ErrAPICallFailed, resp.StatusCode)
	}

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	var ollamaResp ollamaResponse
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for API errors
	if ollamaResp.Error != "" {
		return "", fmt.Errorf("%w: %s", ErrAPICallFailed, ollamaResp.Error)
	}

	// Extract summary text
	if ollamaResp.Response == "" {
		return "", fmt.Errorf("%w: empty summary returned", ErrAPICallFailed)
	}

	return ollamaResp.Response, nil
}
