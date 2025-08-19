package pollinations

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	httpClient *http.Client
	baseURL    string
}

type TextRequest struct {
	Prompt string `json:"prompt"`
	Model  string `json:"model,omitempty"`
	Seed   int    `json:"seed,omitempty"`
}

type TextResponse struct {
	Text string `json:"text"`
}

func NewPollinationsClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: "https://text.pollinations.ai",
	}
}

func (c *Client) GenerateText(ctx context.Context, prompt string) (string, error) {
	// Simple text generation using Pollinations
	req := TextRequest{
		Prompt: prompt,
		Model:  "openai",
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL, bytes.NewReader(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// For simplicity, if Pollinations returns plain text, return it directly
	// Otherwise, try to parse as JSON
	var textResp TextResponse
	if err := json.Unmarshal(body, &textResp); err != nil {
		// Return as plain text if JSON parsing fails
		return string(body), nil
	}

	return textResp.Text, nil
}

// GetModels fetches available models from Pollinations
func (c *Client) GetModels(ctx context.Context) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/models", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Return default models if endpoint doesn't exist
		return []string{"openai", "anthropic", "mistral", "llama"}, nil
	}

	var models []string
	if err := json.NewDecoder(resp.Body).Decode(&models); err != nil {
		// Return default models if parsing fails
		return []string{"openai", "anthropic", "mistral", "llama"}, nil
	}

	return models, nil
}