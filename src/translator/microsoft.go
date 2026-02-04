package translator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type MicrosoftTranslator struct {
	APIKey     string
	TargetLang string
}

type microsoftRequest struct {
	Text string `json:"Text"`
}

type microsoftResponse struct {
	Translations []struct {
		Text string `json:"text"`
		To   string `json:"to"`
	} `json:"translations"`
}

func (m *MicrosoftTranslator) Translate(ctx context.Context, content, targetLang string) (string, error) {
	if targetLang == "" {
		targetLang = m.TargetLang
	}

	// Truncate if too long
	content = truncateContent(content, 50000)

	reqBody := []microsoftRequest{
		{Text: content},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrAPICallFailed, err)
	}

	url := fmt.Sprintf("https://api.cognitive.microsofttranslator.com/translate?api-version=3.0&to=%s", targetLang)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrAPICallFailed, err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Ocp-Apim-Subscription-Key", m.APIKey)

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

	var msResp []microsoftResponse
	if err := json.Unmarshal(body, &msResp); err != nil {
		return "", fmt.Errorf("%w: %v", ErrAPICallFailed, err)
	}

	if len(msResp) == 0 || len(msResp[0].Translations) == 0 {
		return "", fmt.Errorf("%w: no translation returned", ErrAPICallFailed)
	}

	return msResp[0].Translations[0].Text, nil
}
