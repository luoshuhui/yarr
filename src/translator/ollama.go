package translator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type OllamaTranslator struct {
	URL        string
	Model      string
	TargetLang string
}

type ollamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type ollamaResponse struct {
	Response string `json:"response"`
}

func (o *OllamaTranslator) Translate(ctx context.Context, content, targetLang string) (string, error) {
	if targetLang == "" {
		targetLang = o.TargetLang
	}

	// Truncate if too long
	content = truncateContent(content, 30000)

	prompt := buildPrompt(content, targetLang)

	reqBody := ollamaRequest{
		Model:  o.Model,
		Prompt: prompt,
		Stream: false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrAPICallFailed, err)
	}

	url := fmt.Sprintf("%s/api/generate", o.URL)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrAPICallFailed, err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrAPICallFailed, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrAPICallFailed, err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%w: API returned status %d: %s", ErrAPICallFailed, resp.StatusCode, string(body))
	}

	var ollamaResp ollamaResponse
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		return "", fmt.Errorf("%w: %v", ErrAPICallFailed, err)
	}

	return ollamaResp.Response, nil
}
