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

// Client represents a Pollinations API client
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// TextRequest represents a text generation request
type TextRequest struct {
	Prompt      string            `json:"prompt"`
	Model       string            `json:"model,omitempty"`
	MaxTokens   int               `json:"max_tokens,omitempty"`
	Temperature float64           `json:"temperature,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// TextResponse represents a text generation response
type TextResponse struct {
	Text      string            `json:"text"`
	Model     string            `json:"model"`
	Tokens    int               `json:"tokens"`
	Cost      float64           `json:"cost"`
	Metadata  map[string]string `json:"metadata"`
	Timestamp time.Time         `json:"timestamp"`
}

// NewClient creates a new Pollinations client
func NewClient() *Client {
	return &Client{
		baseURL: "https://text.pollinations.ai",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// NewClientWithURL creates a new client with custom base URL
func NewClientWithURL(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GenerateText generates text using Pollinations API
func (c *Client) GenerateText(ctx context.Context, prompt string) (string, error) {
	request := TextRequest{
		Prompt:      prompt,
		MaxTokens:   1000,
		Temperature: 0.7,
	}

	return c.GenerateTextWithRequest(ctx, request)
}

// GenerateTextWithRequest generates text with a custom request
func (c *Client) GenerateTextWithRequest(ctx context.Context, request TextRequest) (string, error) {
	// For now, return a mock response since we're creating a stub implementation
	// In a real implementation, this would make an HTTP request to Pollinations API
	
	mockResponse := fmt.Sprintf(`Based on the prompt "%s", here is a generated response. This is a mock implementation that provides structured analysis and suggestions.

Key points:
1. The request has been analyzed for complexity and requirements
2. Appropriate provider selection would be based on task characteristics
3. Cost optimization opportunities have been identified
4. Performance metrics suggest efficient routing

This mock response demonstrates the expected format and structure for AI-generated content.`, request.Prompt)

	return mockResponse, nil
}

// GenerateYAML generates YAML configuration using AI
func (c *Client) GenerateYAML(ctx context.Context, prompt string) (string, error) {
	yamlPrompt := fmt.Sprintf(`Generate a YAML configuration based on this request: %s

Please return valid YAML format with appropriate structure for the requested configuration.`, prompt)

	response, err := c.GenerateText(ctx, yamlPrompt)
	if err != nil {
		return "", err
	}

	// Mock YAML response
	mockYAML := `# Generated YAML Configuration
name: generated-config
version: "1.0"
settings:
  enabled: true
  timeout: 30s
  retries: 3
providers:
  - name: primary
    tier: official
    endpoint: https://api.example.com
  - name: fallback
    tier: community
    endpoint: https://community.example.com
optimization:
  cost_threshold: 0.01
  performance_weight: 0.7
  quality_weight: 0.3`

	return mockYAML, nil
}

// GetModels returns available models (mock implementation)
func (c *Client) GetModels(ctx context.Context) ([]string, error) {
	models := []string{
		"pollinations-text-v1",
		"pollinations-creative-v1",
		"pollinations-code-v1",
		"pollinations-analysis-v1",
	}

	return models, nil
}

// HealthCheck checks if the Pollinations service is available
func (c *Client) HealthCheck(ctx context.Context) error {
	// Mock health check - always return healthy for stub implementation
	return nil
}

// SetTimeout sets the HTTP client timeout
func (c *Client) SetTimeout(timeout time.Duration) {
	c.httpClient.Timeout = timeout
}

// makeRequest makes an HTTP request (helper method for future real implementation)
func (c *Client) makeRequest(ctx context.Context, method, endpoint string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+endpoint, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return c.httpClient.Do(req)
}