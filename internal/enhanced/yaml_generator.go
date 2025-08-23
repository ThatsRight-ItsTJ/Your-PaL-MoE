package enhanced

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// YAMLGenerator handles CSV-to-YAML conversion using AI
type YAMLGenerator struct {
	logger      *logrus.Logger
	pollinationsURL string
	httpClient  *http.Client
}

// NewYAMLGenerator creates a new YAML generator
func NewYAMLGenerator(logger *logrus.Logger) *YAMLGenerator {
	return &YAMLGenerator{
		logger:          logger,
		pollinationsURL: "https://text.pollinations.ai",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GenerateYAMLFromProvider converts a Provider to YAML using AI
func (y *YAMLGenerator) GenerateYAMLFromProvider(ctx context.Context, provider *Provider) (string, error) {
	y.logger.Infof("Generating YAML for provider: %s", provider.ID)
	
	prompt := y.buildPrompt(provider)
	
	response, err := y.callPollinationsAPI(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to call Pollinations API: %w", err)
	}
	
	// Extract YAML from response
	yaml := y.extractYAML(response)
	if yaml == "" {
		return "", fmt.Errorf("no valid YAML generated from AI response")
	}
	
	y.logger.Infof("Successfully generated YAML for provider %s", provider.ID)
	return yaml, nil
}

// buildPrompt creates the AI prompt for YAML generation
func (y *YAMLGenerator) buildPrompt(provider *Provider) string {
	var sb strings.Builder
	
	sb.WriteString("Generate a YAML configuration file for an AI provider with the following details:\n\n")
	sb.WriteString(fmt.Sprintf("Provider ID: %s\n", provider.ID))
	sb.WriteString(fmt.Sprintf("Provider Name: %s\n", provider.Name))
	sb.WriteString(fmt.Sprintf("Tier: %s\n", provider.Tier))
	sb.WriteString(fmt.Sprintf("Endpoint: %s\n", provider.Endpoint))
	
	// Handle multiple models
	if len(provider.Models) > 0 {
		if provider.ModelsSource == "list" {
			sb.WriteString(fmt.Sprintf("Models: %s\n", strings.Join(provider.Models, ", ")))
		} else {
			sb.WriteString(fmt.Sprintf("Models Endpoint: %s\n", provider.ModelsSource))
			if len(provider.Models) > 0 {
				sb.WriteString(fmt.Sprintf("Available Models: %s\n", strings.Join(provider.Models, ", ")))
			}
		}
	}
	
	sb.WriteString(fmt.Sprintf("Cost Per Token: %.6f\n", provider.CostPerToken))
	sb.WriteString(fmt.Sprintf("Max Tokens: %d\n", provider.MaxTokens))
	
	if len(provider.Capabilities) > 0 {
		sb.WriteString(fmt.Sprintf("Capabilities: %s\n", strings.Join(provider.Capabilities, ", ")))
	}
	
	// Add metadata/additional info
	if additionalInfo, exists := provider.Metadata["additional_info"]; exists {
		sb.WriteString(fmt.Sprintf("Additional Info: %s\n", additionalInfo))
	}
	
	// Add rate limits and other metadata
	for key, value := range provider.Metadata {
		if key != "additional_info" {
			sb.WriteString(fmt.Sprintf("%s: %v\n", strings.Title(strings.ReplaceAll(key, "_", " ")), value))
		}
	}
	
	sb.WriteString("\nGenerate a complete YAML configuration file that includes:\n")
	sb.WriteString("1. Provider configuration section\n")
	sb.WriteString("2. API settings (endpoint, authentication)\n")
	sb.WriteString("3. Model parameters (temperature, max_tokens, etc.)\n")
	sb.WriteString("4. Rate limiting configuration if applicable\n")
	sb.WriteString("5. Cost tracking settings\n")
	sb.WriteString("6. Retry and timeout configurations\n")
	sb.WriteString("7. Any provider-specific settings based on the additional info\n\n")
	sb.WriteString("Format the output as valid YAML. Include comments explaining each section.\n")
	sb.WriteString("Only return the YAML content, no additional text or explanation.")
	
	return sb.String()
}

// callPollinationsAPI calls the Pollinations text generation API (simple GET)
func (y *YAMLGenerator) callPollinationsAPI(ctx context.Context, prompt string) (string, error) {
	// Use simple text API with proper URL encoding
	encodedPrompt := url.QueryEscape(prompt)
	apiURL := fmt.Sprintf("%s/%s", y.pollinationsURL, encodedPrompt)
	
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	
	resp, err := y.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returned status %d", resp.StatusCode)
	}
	
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(resp.Body); err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}
	
	return buf.String(), nil
}

// extractYAML extracts YAML content from the AI response
func (y *YAMLGenerator) extractYAML(response string) string {
	// Remove any markdown code blocks
	response = strings.TrimSpace(response)
	
	// Look for YAML content between code blocks
	if strings.Contains(response, "```yaml") {
		parts := strings.Split(response, "```yaml")
		if len(parts) > 1 {
			yamlPart := strings.Split(parts[1], "```")[0]
			return strings.TrimSpace(yamlPart)
		}
	} else if strings.Contains(response, "```") {
		parts := strings.Split(response, "```")
		if len(parts) >= 2 {
			return strings.TrimSpace(parts[1])
		}
	}
	
	// If no code blocks, assume entire response is YAML
	return response
}

// GenerateYAMLBatch generates YAML for multiple providers
func (y *YAMLGenerator) GenerateYAMLBatch(ctx context.Context, providers []*Provider) (map[string]string, error) {
	results := make(map[string]string)
	
	for _, provider := range providers {
		yaml, err := y.GenerateYAMLFromProvider(ctx, provider)
		if err != nil {
			y.logger.Errorf("Failed to generate YAML for provider %s: %v", provider.ID, err)
			continue
		}
		results[provider.ID] = yaml
	}
	
	y.logger.Infof("Generated YAML for %d/%d providers", len(results), len(providers))
	return results, nil
}

// SaveYAMLToFile saves generated YAML to a file
func (y *YAMLGenerator) SaveYAMLToFile(providerID, yaml string) error {
	filename := fmt.Sprintf("configs/%s.yaml", providerID)
	
	// Create configs directory if it doesn't exist
	if err := ensureDir("configs"); err != nil {
		return fmt.Errorf("failed to create configs directory: %w", err)
	}
	
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filename, err)
	}
	defer file.Close()
	
	_, err = file.WriteString(yaml)
	if err != nil {
		return fmt.Errorf("failed to write YAML to file: %w", err)
	}
	
	y.logger.Infof("Saved YAML configuration to %s", filename)
	return nil
}

// Helper function to ensure directory exists
func ensureDir(dirName string) error {
	if _, err := os.Stat(dirName); os.IsNotExist(err) {
		return os.MkdirAll(dirName, 0755)
	}
	return nil
}