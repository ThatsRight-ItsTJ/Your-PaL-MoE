package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/labring/aiproxy/core/pkg/pollinations"
	"gopkg.in/yaml.v3"
)

type AutoConfigurator struct {
	csvParser          *CSVParser
	pollinationsClient *pollinations.Client
	outputDir          string
	httpClient         *http.Client
}

type GeneratedConfig struct {
	Provider     ProviderInfo     `yaml:"provider"`
	Endpoint     EndpointInfo     `yaml:"endpoint"`
	Capabilities []string         `yaml:"capabilities"`
	Models       ModelsConfig     `yaml:"models"`
	APIFormat    APIFormatInfo    `yaml:"api_format"`
	HealthCheck  HealthCheckInfo  `yaml:"health_check"`
	Metadata     MetadataInfo     `yaml:"metadata"`
}

type ProviderInfo struct {
	Name      string    `yaml:"name"`
	Tier      string    `yaml:"tier"`
	Type      string    `yaml:"type"`
	Source    string    `yaml:"source"`
	Generated time.Time `yaml:"generated"`
}

type EndpointInfo struct {
	BaseURL        string     `yaml:"base_url"`
	Authentication AuthConfig `yaml:"authentication"`
}

type ModelsConfig struct {
	TextGeneration  []ModelInfo `yaml:"text_generation,omitempty"`
	ImageGeneration []ModelInfo `yaml:"image_generation,omitempty"`
	AudioGeneration []ModelInfo `yaml:"audio_generation,omitempty"`
	Embeddings      []ModelInfo `yaml:"embeddings,omitempty"`
}

type ModelInfo struct {
	Name            string  `yaml:"name"`
	ContextWindow   int     `yaml:"context_window,omitempty"`
	CostPer1kTokens float64 `yaml:"cost_per_1k_tokens,omitempty"`
	CostPerImage    float64 `yaml:"cost_per_image,omitempty"`
	QualityScore    int     `yaml:"quality_score"`
	MaxResolution   string  `yaml:"max_resolution,omitempty"`
}

type APIFormatInfo struct {
	RequestFormat  string `yaml:"request_format"`
	ResponseFormat string `yaml:"response_format"`
}

type HealthCheckInfo struct {
	Endpoint string `yaml:"endpoint"`
	Method   string `yaml:"method"`
	Timeout  int    `yaml:"timeout"`
}

type MetadataInfo struct {
	CSVRow      int       `yaml:"csv_row"`
	LastUpdated time.Time `yaml:"last_updated"`
	Notes       []string  `yaml:"notes,omitempty"`
}

type ProviderAnalysis struct {
	Capabilities        []string           `json:"capabilities"`
	RequestFormat       string             `json:"request_format"`
	ResponseFormat      string             `json:"response_format"`
	HealthCheckEndpoint string             `json:"health_check_endpoint"`
	EstimatedCosts      map[string]float64 `json:"estimated_costs"`
	QualityScores       map[string]int     `json:"quality_scores"`
	Notes               []string           `json:"notes"`
}

func NewAutoConfigurator(csvPath, outputDir string) *AutoConfigurator {
	return &AutoConfigurator{
		csvParser:          NewCSVParser(csvPath),
		pollinationsClient: pollinations.NewPollinationsClient(),
		outputDir:          outputDir,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (ac *AutoConfigurator) GenerateConfigurations(ctx context.Context) error {
	providers, err := ac.csvParser.LoadProviders()
	if err != nil {
		return fmt.Errorf("failed to load providers from CSV: %w", err)
	}

	for name, provider := range providers {
		fmt.Printf("Configuring provider: %s\n", name)

		config, err := ac.generateProviderConfig(ctx, provider)
		if err != nil {
			fmt.Printf("Warning: Failed to generate config for %s: %v\n", name, err)
			continue
		}

		if err := ac.saveConfiguration(provider.Tier, name, config); err != nil {
			return fmt.Errorf("failed to save configuration for %s: %w", name, err)
		}

		// Generate script template for unofficial APIs
		if provider.Tier == "unofficial" && strings.HasPrefix(provider.Endpoint, "./scripts/") {
			if err := ac.generateScriptTemplate(provider); err != nil {
				fmt.Printf("Warning: Failed to generate script template for %s: %v\n", name, err)
			}
		}
	}

	return nil
}

func (ac *AutoConfigurator) generateProviderConfig(ctx context.Context, provider *ProviderConfig) (*GeneratedConfig, error) {
	// Discover models
	models, err := ac.discoverModels(ctx, provider)
	if err != nil {
		return nil, fmt.Errorf("failed to discover models: %w", err)
	}

	// Use Pollinations to analyze the provider
	analysis, err := ac.analyzeProviderWithPollinations(ctx, provider, models)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze provider with Pollinations: %w", err)
	}

	// Build configuration
	config := &GeneratedConfig{
		Provider: ProviderInfo{
			Name:      provider.Name,
			Tier:      provider.Tier,
			Type:      ac.determineProviderType(provider),
			Source:    "csv_auto_config",
			Generated: time.Now(),
		},
		Endpoint: EndpointInfo{
			BaseURL:        provider.Endpoint,
			Authentication: provider.Authentication,
		},
		Capabilities: analysis.Capabilities,
		Models:       ac.organizeModels(models, analysis),
		APIFormat: APIFormatInfo{
			RequestFormat:  analysis.RequestFormat,
			ResponseFormat: analysis.ResponseFormat,
		},
		HealthCheck: HealthCheckInfo{
			Endpoint: analysis.HealthCheckEndpoint,
			Method:   "GET",
			Timeout:  30,
		},
		Metadata: MetadataInfo{
			LastUpdated: time.Now(),
			Notes:       analysis.Notes,
		},
	}

	return config, nil
}

func (ac *AutoConfigurator) discoverModels(ctx context.Context, provider *ProviderConfig) ([]string, error) {
	switch provider.ModelsSource.Type {
	case "list":
		// Direct list provided
		if models, ok := provider.ModelsSource.Value.([]string); ok {
			return models, nil
		}
		return nil, fmt.Errorf("invalid models list format")

	case "url":
		// Fetch from URL
		url, ok := provider.ModelsSource.Value.(string)
		if !ok {
			return nil, fmt.Errorf("invalid models URL format")
		}
		return ac.fetchModelsFromURL(ctx, url, provider)

	default:
		return nil, fmt.Errorf("unknown models source type: %s", provider.ModelsSource.Type)
	}
}

func (ac *AutoConfigurator) fetchModelsFromURL(ctx context.Context, url string, provider *ProviderConfig) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Add authentication if required
	if provider.Authentication.Required && provider.Authentication.Type != "none" {
		authValue := os.Getenv(provider.Authentication.EnvVar)
		if authValue != "" {
			switch provider.Authentication.Type {
			case "bearer_token":
				req.Header.Set("Authorization", "Bearer "+authValue)
			case "api_key":
				req.Header.Set(provider.Authentication.Header, authValue)
			}
		}
	}

	resp, err := ac.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	var data interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	// Parse different response formats
	return ac.parseModelsResponse(data, provider)
}

func (ac *AutoConfigurator) parseModelsResponse(data interface{}, provider *ProviderConfig) ([]string, error) {
	var models []string

	switch v := data.(type) {
	case []interface{}:
		// Direct array of models
		for _, item := range v {
			if model := ac.extractModelName(item); model != "" {
				models = append(models, model)
			}
		}
	case map[string]interface{}:
		// Check common response patterns
		if dataField, ok := v["data"]; ok {
			if dataArray, ok := dataField.([]interface{}); ok {
				for _, item := range dataArray {
					if model := ac.extractModelName(item); model != "" {
						models = append(models, model)
					}
				}
			}
		} else if modelsField, ok := v["models"]; ok {
			if modelsArray, ok := modelsField.([]interface{}); ok {
				for _, item := range modelsArray {
					if model := ac.extractModelName(item); model != "" {
						models = append(models, model)
					}
				}
			}
		}
	}

	// If we couldn't parse automatically, use Pollinations to help
	if len(models) == 0 {
		return ac.analyzeModelsWithPollinations(data, provider)
	}

	return models, nil
}

func (ac *AutoConfigurator) extractModelName(item interface{}) string {
	switch v := item.(type) {
	case string:
		return v
	case map[string]interface{}:
		// Try common field names
		for _, field := range []string{"id", "name", "model", "model_name"} {
			if name, ok := v[field].(string); ok {
				return name
			}
		}
	}
	return ""
}

func (ac *AutoConfigurator) analyzeProviderWithPollinations(ctx context.Context, provider *ProviderConfig, models []string) (*ProviderAnalysis, error) {
	prompt := fmt.Sprintf(`
Analyze this AI provider and generate configuration details:

Provider Name: %s
Tier: %s (official=paid/premium, community=free/low-cost, unofficial=reverse-engineered)
Endpoint: %s
Available Models: %v

Please analyze and return JSON with:
{
    "capabilities": ["text_generation", "image_generation", "audio_generation", "embeddings"],
    "request_format": "openai|anthropic|huggingface|custom",
    "response_format": "openai|anthropic|huggingface|custom", 
    "health_check_endpoint": "/health or /models or /",
    "estimated_costs": {"text_generation": 0.001, "image_generation": 0.02},
    "quality_scores": {"model_name": 8},
    "notes": ["Additional configuration notes"]
}

Consider:
- Official tier providers typically use bearer token auth and have usage costs
- Community tier providers are often free or low-cost with simpler auth
- Unofficial tier providers may use custom auth (cookies, etc.) and are free but unstable
- Infer capabilities from model names and provider patterns
- Estimate costs based on tier (official=paid, community=low/free, unofficial=free)
- Quality scores 1-10 based on model reputation and tier
`, provider.Name, provider.Tier, provider.Endpoint, models)

	response, err := ac.pollinationsClient.GenerateText(ctx, prompt)
	if err != nil {
		return nil, err
	}

	var analysis ProviderAnalysis
	if err := json.Unmarshal([]byte(response), &analysis); err != nil {
		// If JSON parsing fails, create a reasonable default
		analysis = ac.createDefaultAnalysis(provider, models)
	}

	return &analysis, nil
}

func (ac *AutoConfigurator) createDefaultAnalysis(provider *ProviderConfig, models []string) ProviderAnalysis {
	capabilities := []string{"text_generation"}
	requestFormat := "openai"
	responseFormat := "openai"

	// Infer capabilities from provider name and models
	providerLower := strings.ToLower(provider.Name)
	if strings.Contains(providerLower, "dalle") || strings.Contains(providerLower, "image") || strings.Contains(providerLower, "pollinations") {
		capabilities = append(capabilities, "image_generation")
	}

	// Set costs based on tier
	estimatedCosts := make(map[string]float64)
	if provider.Tier == "official" {
		estimatedCosts["text_generation"] = 0.002
		estimatedCosts["image_generation"] = 0.04
	} else {
		estimatedCosts["text_generation"] = 0.0
		estimatedCosts["image_generation"] = 0.0
	}

	// Set quality scores based on tier
	qualityScores := make(map[string]int)
	defaultQuality := 7
	if provider.Tier == "official" {
		defaultQuality = 9
	} else if provider.Tier == "unofficial" {
		defaultQuality = 6
	}

	for _, model := range models {
		qualityScores[model] = defaultQuality
	}

	return ProviderAnalysis{
		Capabilities:        capabilities,
		RequestFormat:       requestFormat,
		ResponseFormat:      responseFormat,
		HealthCheckEndpoint: "/models",
		EstimatedCosts:      estimatedCosts,
		QualityScores:       qualityScores,
		Notes:               []string{"Auto-generated configuration - customize as needed"},
	}
}

func (ac *AutoConfigurator) organizeModels(models []string, analysis *ProviderAnalysis) ModelsConfig {
	config := ModelsConfig{}

	for _, model := range models {
		modelInfo := ModelInfo{
			Name:         model,
			QualityScore: analysis.QualityScores[model],
		}

		// Set context window based on model patterns
		if strings.Contains(strings.ToLower(model), "gpt-4") {
			modelInfo.ContextWindow = 128000
		} else if strings.Contains(strings.ToLower(model), "gpt-3") {
			modelInfo.ContextWindow = 4096
		} else {
			modelInfo.ContextWindow = 8192 // Default
		}

		// Categorize by capability
		modelLower := strings.ToLower(model)
		if strings.Contains(modelLower, "dalle") || strings.Contains(modelLower, "image") || strings.Contains(modelLower, "vision") {
			modelInfo.CostPerImage = analysis.EstimatedCosts["image_generation"]
			modelInfo.MaxResolution = "1024x1024"
			config.ImageGeneration = append(config.ImageGeneration, modelInfo)
		} else if strings.Contains(modelLower, "embed") {
			modelInfo.CostPer1kTokens = analysis.EstimatedCosts["text_generation"] * 0.1 // Embeddings typically cheaper
			config.Embeddings = append(config.Embeddings, modelInfo)
		} else if strings.Contains(modelLower, "audio") || strings.Contains(modelLower, "speech") || strings.Contains(modelLower, "tts") {
			config.AudioGeneration = append(config.AudioGeneration, modelInfo)
		} else {
			// Default to text generation
			modelInfo.CostPer1kTokens = analysis.EstimatedCosts["text_generation"]
			config.TextGeneration = append(config.TextGeneration, modelInfo)
		}
	}

	return config
}

func (ac *AutoConfigurator) saveConfiguration(tier, name string, config *GeneratedConfig) error {
	// Create tier directory
	tierDir := filepath.Join(ac.outputDir, tier)
	if err := os.MkdirAll(tierDir, 0755); err != nil {
		return err
	}

	// Generate filename
	filename := strings.ToLower(strings.ReplaceAll(name, " ", "-")) + ".yaml"
	filePath := filepath.Join(tierDir, filename)

	// Marshal to YAML
	yamlData, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	// Add header comment
	header := fmt.Sprintf(`# Auto-generated configuration for %s
# Tier: %s | Generated: %s
# Source: CSV auto-configuration
# 
# To regenerate: task-master configure-providers
# To customize: Edit this file directly
#

`, name, tier, time.Now().Format(time.RFC3339))

	finalYAML := header + string(yamlData)

	return os.WriteFile(filePath, []byte(finalYAML), 0644)
}

func (ac *AutoConfigurator) determineProviderType(provider *ProviderConfig) string {
	if strings.HasPrefix(provider.Endpoint, "http://") || strings.HasPrefix(provider.Endpoint, "https://") {
		return "rest_api"
	}
	if strings.HasPrefix(provider.Endpoint, "./scripts/") {
		return "script"
	}
	return "unknown"
}

func (ac *AutoConfigurator) analyzeModelsWithPollinations(data interface{}, provider *ProviderConfig) ([]string, error) {
	prompt := fmt.Sprintf(`
Extract model names from this API response:

Provider: %s
Response Data: %v

Return a JSON array of model names found in the response.
Look for common patterns like "id", "name", "model", etc.
`, provider.Name, data)

	response, err := ac.pollinationsClient.GenerateText(context.Background(), prompt)
	if err != nil {
		return []string{}, nil // Return empty if analysis fails
	}

	var models []string
	if err := json.Unmarshal([]byte(response), &models); err != nil {
		return []string{}, nil // Return empty if parsing fails
	}

	return models, nil
}

func (ac *AutoConfigurator) generateScriptTemplate(provider *ProviderConfig) error {
	// Create scripts directory
	scriptsDir := "scripts"
	if err := os.MkdirAll(scriptsDir, 0755); err != nil {
		return err
	}

	// Extract script path
	scriptPath := strings.TrimPrefix(provider.Endpoint, "./")

	// Generate template based on file extension
	var template string
	switch {
	case strings.HasSuffix(scriptPath, ".py"):
		template = generatePythonTemplate(provider)
	case strings.HasSuffix(scriptPath, ".js"):
		template = generateJavaScriptTemplate(provider)
	case strings.HasSuffix(scriptPath, ".sh"):
		template = generateBashTemplate(provider)
	default:
		return fmt.Errorf("unsupported script type for %s", scriptPath)
	}

	return os.WriteFile(scriptPath, []byte(template), 0755)
}

func generatePythonTemplate(provider *ProviderConfig) string {
	return fmt.Sprintf(`#!/usr/bin/env python3
"""
Auto-generated wrapper for %s
Tier: %s
Generated: %s
"""

import sys
import json
import requests
import os

def main():
    if len(sys.argv) < 2:
        print("Usage: python %s '<request_json>'")
        sys.exit(1)
    
    request_data = json.loads(sys.argv[1])
    
    # TODO: Implement %s integration
    # Extract prompt, model, parameters from request_data
    # Make API calls to %s
    # Return OpenAI-compatible response
    
    response = {
        "choices": [
            {
                "message": {
                    "role": "assistant", 
                    "content": "TODO: Implement %s wrapper"
                }
            }
        ]
    }
    
    print(json.dumps(response))

if __name__ == "__main__":
    main()
`, provider.Name, provider.Tier, time.Now().Format(time.RFC3339), 
   provider.Name, provider.Name, provider.Endpoint, provider.Name)
}

func generateJavaScriptTemplate(provider *ProviderConfig) string {
	return fmt.Sprintf(`#!/usr/bin/env node
/**
 * Auto-generated wrapper for %s
 * Tier: %s
 * Generated: %s
 */

const axios = require('axios');

async function main() {
    if (process.argv.length < 3) {
        console.error('Usage: node %s \'<request_json>\'');
        process.exit(1);
    }
    
    const requestData = JSON.parse(process.argv[2]);
    
    // TODO: Implement %s integration
    // Extract prompt, model, parameters from requestData
    // Make API calls to %s
    // Return OpenAI-compatible response
    
    const response = {
        choices: [
            {
                message: {
                    role: 'assistant',
                    content: 'TODO: Implement %s wrapper'
                }
            }
        ]
    };
    
    console.log(JSON.stringify(response));
}

main().catch(console.error);
`, provider.Name, provider.Tier, time.Now().Format(time.RFC3339),
   provider.Name, provider.Name, provider.Endpoint, provider.Name)
}

func generateBashTemplate(provider *ProviderConfig) string {
	return fmt.Sprintf(`#!/bin/bash
# Auto-generated wrapper for %s
# Tier: %s
# Generated: %s

if [ $# -lt 1 ]; then
    echo "Usage: $0 '<request_json>'"
    exit 1
fi

REQUEST_JSON="$1"

# TODO: Implement %s integration
# Parse REQUEST_JSON to extract prompt, model, parameters
# Make API calls to %s
# Return OpenAI-compatible response

# Placeholder response
echo '{"choices":[{"message":{"role":"assistant","content":"TODO: Implement %s wrapper"}}]}'
`, provider.Name, provider.Tier, time.Now().Format(time.RFC3339),
   provider.Name, provider.Endpoint, provider.Name)
}