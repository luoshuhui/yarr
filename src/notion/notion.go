package notion

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type NotionClient struct {
	APIKey     string
	DatabaseID string
	HTTPClient *http.Client
}

type Config struct {
	APIKey     string
	DatabaseID string
}

func New(config Config) (*NotionClient, error) {
	if config.APIKey == "" {
		return nil, ErrMissingAPIKey
	}
	if config.DatabaseID == "" {
		return nil, ErrMissingDatabaseID
	}

	return &NotionClient{
		APIKey:     config.APIKey,
		DatabaseID: config.DatabaseID,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// CreatePage 创建Notion页面
func (n *NotionClient) CreatePage(ctx context.Context, title, content string) (string, error) {
	// 将HTML内容转换为Notion blocks
	blocks, err := HTMLToNotionBlocks(content)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrConversionFailed, err)
	}

	// Notion API限制：单次请求最多100个blocks
	const maxBlocksPerRequest = 100

	// 分批处理blocks
	initialBlocks := blocks
	if len(blocks) > maxBlocksPerRequest {
		initialBlocks = blocks[:maxBlocksPerRequest]
	}

	// 构建请求体
	reqBody := PageCreateRequest{
		Parent: Parent{
			DatabaseID: n.DatabaseID,
		},
		Properties: Properties{
			Title: TitleProperty{
				Title: []RichText{
					{
						Text: Text{
							Content: title,
						},
					},
				},
			},
		},
		Children: initialBlocks,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrAPICallFailed, err)
	}

	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, "POST",
		"https://api.notion.com/v1/pages",
		bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrAPICallFailed, err)
	}

	req.Header.Set("Authorization", "Bearer "+n.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Notion-Version", "2022-06-28")

	// 发送请求
	resp, err := n.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrAPICallFailed, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrAPICallFailed, err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%w: API returned status %d: %s",
			ErrAPICallFailed, resp.StatusCode, string(body))
	}

	// 解析响应
	var pageResp PageCreateResponse
	if err := json.Unmarshal(body, &pageResp); err != nil {
		return "", fmt.Errorf("%w: %v", ErrAPICallFailed, err)
	}

	// 如果还有剩余的blocks，分批追加
	if len(blocks) > maxBlocksPerRequest {
		remainingBlocks := blocks[maxBlocksPerRequest:]
		if err := n.appendBlocks(ctx, pageResp.ID, remainingBlocks); err != nil {
			return pageResp.URL, fmt.Errorf("page created but failed to append remaining blocks: %w", err)
		}
	}

	return pageResp.URL, nil
}

// appendBlocks 分批追加blocks到指定页面
func (n *NotionClient) appendBlocks(ctx context.Context, pageID string, blocks []Block) error {
	const maxBlocksPerRequest = 100

	for i := 0; i < len(blocks); i += maxBlocksPerRequest {
		end := i + maxBlocksPerRequest
		if end > len(blocks) {
			end = len(blocks)
		}

		batch := blocks[i:end]
		if err := n.appendBlocksBatch(ctx, pageID, batch); err != nil {
			return err
		}
	}

	return nil
}

// appendBlocksBatch 追加一批blocks（最多100个）
func (n *NotionClient) appendBlocksBatch(ctx context.Context, pageID string, blocks []Block) error {
	reqBody := map[string]interface{}{
		"children": blocks,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrAPICallFailed, err)
	}

	url := fmt.Sprintf("https://api.notion.com/v1/blocks/%s/children", pageID)
	req, err := http.NewRequestWithContext(ctx, "PATCH", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("%w: %v", ErrAPICallFailed, err)
	}

	req.Header.Set("Authorization", "Bearer "+n.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Notion-Version", "2022-06-28")

	resp, err := n.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrAPICallFailed, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrAPICallFailed, err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: API returned status %d: %s",
			ErrAPICallFailed, resp.StatusCode, string(body))
	}

	return nil
}
