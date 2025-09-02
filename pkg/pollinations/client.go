package pollinations

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Client represents a Pollinations API client
type Client struct {
	baseURL    string
	httpClient *http.Client
	timeout    time.Duration
}

// NewClient creates a new Pollinations client
func NewClient() *Client {
	return &Client{
		baseURL: "https://text.pollinations.ai",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		timeout: 30 * time.Second,
	}
}

// GenerateText generates text using the Pollinations API
func (c *Client) GenerateText(ctx context.Context, prompt string) (string, error) {
	if prompt == "" {
		return "", fmt.Errorf("prompt cannot be empty")
	}

	// Encode the prompt for URL
	encodedPrompt := url.QueryEscape(prompt)
	requestURL := fmt.Sprintf("%s/%s", c.baseURL, encodedPrompt)

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, "GET", requestURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("User-Agent", "Your-PaL-MoE/1.0")
	req.Header.Set("Accept", "text/plain")

	// Make the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	return string(body), nil
}

// GenerateYAML generates YAML configuration using the Pollinations API
func (c *Client) GenerateYAML(ctx context.Context, prompt string) (string, error) {
	yamlPrompt := fmt.Sprintf("Generate a valid YAML configuration based on the following requirements:\n\n%s\n\nReturn only the YAML content without any additional text or explanation.", prompt)
	
	return c.GenerateText(ctx, yamlPrompt)
}

// SetTimeout sets the client timeout
func (c *Client) SetTimeout(timeout time.Duration) {
	c.timeout = timeout
	c.httpClient.Timeout = timeout
}

// Health checks if the Pollinations API is accessible
func (c *Client) Health(ctx context.Context) error {
	testPrompt := "test"
	_, err := c.GenerateText(ctx, testPrompt)
	return err
}