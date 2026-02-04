package translator

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type GoogleTranslator struct {
	TargetLang string
}

type googleResponse struct {
	Data struct {
		Translations []struct {
			TranslatedText string `json:"translatedText"`
		} `json:"translations"`
	} `json:"data"`
}

func (g *GoogleTranslator) Translate(ctx context.Context, content, targetLang string) (string, error) {
	if targetLang == "" {
		targetLang = g.TargetLang
	}

	// Truncate if too long
	content = truncateContent(content, 5000)

	// Use Google Translate free API endpoint
	baseURL := "https://translate.googleapis.com/translate_a/single"
	params := url.Values{}
	params.Add("client", "gtx")
	params.Add("sl", "auto")
	params.Add("tl", targetLang)
	params.Add("dt", "t")
	params.Add("q", content)

	reqURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrAPICallFailed, err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0")

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
		return "", fmt.Errorf("%w: API returned status %d", ErrAPICallFailed, resp.StatusCode)
	}

	// Parse the response - it's a JSON array
	var result []interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("%w: %v", ErrAPICallFailed, err)
	}

	if len(result) == 0 {
		return "", fmt.Errorf("%w: no translation returned", ErrAPICallFailed)
	}

	// Extract translated text
	translations, ok := result[0].([]interface{})
	if !ok || len(translations) == 0 {
		return "", fmt.Errorf("%w: unexpected response format", ErrAPICallFailed)
	}

	var translatedText string
	for _, t := range translations {
		if arr, ok := t.([]interface{}); ok && len(arr) > 0 {
			if text, ok := arr[0].(string); ok {
				translatedText += text
			}
		}
	}

	if translatedText == "" {
		return "", fmt.Errorf("%w: no translation returned", ErrAPICallFailed)
	}

	return translatedText, nil
}
