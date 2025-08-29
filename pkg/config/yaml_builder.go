package config

import (
	"crypto/sha256"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// YAMLBuilder handles CSV to YAML configuration enhancement
type YAMLBuilder struct {
	csvPath      string
	outputDir    string
	httpClient   *http.Client
	capabilities map[string]CapabilityProfile
}

// CapabilityProfile defines AI model capabilities
type CapabilityProfile struct {
	Reasoning    int `yaml:"reasoning"`
	Knowledge    int `yaml:"knowledge"`
	Computation  int `yaml:"computation"`
	Coordination int `yaml:"coordination"`
}

// ProviderConfig represents enhanced provider configuration
type ProviderConfig struct {
	ID           string            `yaml:"id"`
	Name         string            `yaml:"name"`
	Tier         string            `yaml:"tier"`
	API          APIConfig         `yaml:"api"`
	Model        ModelConfig       `yaml:"model"`
	Capabilities CapabilityProfile `yaml:"capabilities"`
	RateLimit    RateLimitConfig   `yaml:"rate_limit"`
	CostTracking CostConfig        `yaml:"cost_tracking"`
	RetryConfig  RetryConfig       `yaml:"retry_config"`
}

// APIConfig defines API connection settings
type APIConfig struct {
	Endpoint       string     `yaml:"endpoint"`
	Authentication AuthConfig `yaml:"authentication"`
}

// AuthConfig defines authentication settings
type AuthConfig struct {
	Type string `yaml:"type"`
	Key  string `yaml:"key"`
}

// ModelConfig defines model-specific settings
type ModelConfig struct {
	Name        string   `yaml:"name"`
	Variants    []string `yaml:"variants"`
	Temperature float64  `yaml:"temperature"`
	MaxTokens   int      `yaml:"max_tokens"`
}

// RateLimitConfig defines rate limiting settings
type RateLimitConfig struct {
	RequestsPerMinute int    `yaml:"requests_per_minute"`
	Tier              string `yaml:"tier"`
}

// CostConfig defines cost tracking settings
type CostConfig struct {
	CostPerToken   float64 `yaml:"cost_per_token"`
	BillingModel   string  `yaml:"billing_model"`
}

// RetryConfig defines retry behavior
type RetryConfig struct {
	MaxRetries        int `yaml:"max_retries"`
	BackoffMultiplier int `yaml:"backoff_multiplier"`
	TimeoutSeconds    int `yaml:"timeout_seconds"`
}

// CSVProvider represents a provider from CSV
type CSVProvider struct {
	Name     string
	Tier     string
	Endpoint string
	APIKey   string
	Model    string
	Other    string
}

// Known capability profiles for common models
var defaultCapabilities = map[string]CapabilityProfile{
	"gpt-4":           {Reasoning: 9, Knowledge: 9, Computation: 7, Coordination: 8},
	"gpt-4-turbo":     {Reasoning: 9, Knowledge: 9, Computation: 8, Coordination: 8},
	"gpt-3.5-turbo":   {Reasoning: 7, Knowledge: 7, Computation: 6, Coordination: 6},
	"claude-3-opus":   {Reasoning: 10, Knowledge: 9, Computation: 8, Coordination: 9},
	"claude-3-sonnet": {Reasoning: 8, Knowledge: 8, Computation: 7, Coordination: 7},
	"llama3":          {Reasoning: 7, Knowledge: 7, Computation: 6, Coordination: 5},
	"mistral":         {Reasoning: 7, Knowledge: 6, Computation: 7, Coordination: 5},
}

// NewYAMLBuilder creates a new YAML builder
func NewYAMLBuilder(csvPath, outputDir string) *YAMLBuilder {
	return &YAMLBuilder{
		csvPath:   csvPath,
		outputDir: outputDir,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		capabilities: defaultCapabilities,
	}
}

// BuildFromCSV processes CSV and generates enhanced YAML configurations
func (yb *YAMLBuilder) BuildFromCSV() error {
	providers, err := yb.ReadCSV()
	if err != nil {
		return fmt.Errorf("failed to read CSV: %w", err)
	}

	// Ensure output directory exists
	if err := os.MkdirAll(yb.outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	for _, provider := range providers {
		config, err := yb.buildProviderConfig(provider)
		if err != nil {
			fmt.Printf("Warning: Failed to build config for %s: %v\n", provider.Name, err)
			continue
		}

		if err := yb.saveYAML(config); err != nil {
			fmt.Printf("Warning: Failed to save YAML for %s: %v\n", provider.Name, err)
		}
	}

	return nil
}

// ReadCSV reads providers from CSV file - public method for external use
func (yb *YAMLBuilder) ReadCSV() ([]CSVProvider, error) {
	file, err := os.Open(yb.csvPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var providers []CSVProvider
	for i, record := range records {
		if i == 0 { // Skip header
			continue
		}

		if len(record) >= 6 {
			providers = append(providers, CSVProvider{
				Name:     record[0],
				Tier:     record[1],
				Endpoint: record[2],
				APIKey:   record[3],
				Model:    record[4],
				Other:    record[5],
			})
		}
	}

	return providers, nil
}

// buildProviderConfig creates enhanced configuration from CSV provider
func (yb *YAMLBuilder) buildProviderConfig(provider CSVProvider) (*ProviderConfig, error) {
	config := &ProviderConfig{
		ID:   yb.generateID(provider.Name),
		Name: provider.Name,
		Tier: provider.Tier,
	}

	// Parse models from CSV (pipe-delimited)
	models := strings.Split(provider.Model, "|")
	primaryModel := strings.TrimSpace(models[0])

	// Set up API configuration
	config.API = APIConfig{
		Endpoint: provider.Endpoint,
		Authentication: AuthConfig{
			Type: yb.detectAuthType(provider.APIKey),
			Key:  provider.APIKey,
		},
	}

	// Query endpoint for model information if possible
	modelInfo := yb.queryModelInfo(provider.Endpoint, provider.APIKey, primaryModel)

	// Set model configuration
	config.Model = ModelConfig{
		Name:        primaryModel,
		Variants:    models,
		Temperature: 0.7, // Default
		MaxTokens:   yb.extractMaxTokens(modelInfo, provider.Other),
	}

	// Determine capabilities
	config.Capabilities = yb.determineCapabilities(primaryModel, modelInfo)

	// Parse "Other" field for additional configuration
	config.RateLimit = yb.parseRateLimit(provider.Other)
	config.CostTracking = yb.parseCostInfo(provider.Other, primaryModel)

	// Set retry configuration
	config.RetryConfig = RetryConfig{
		MaxRetries:        3,
		BackoffMultiplier: 2,
		TimeoutSeconds:    30,
	}

	return config, nil
}

// generateID creates a unique ID from provider name
func (yb *YAMLBuilder) generateID(name string) string {
	// Convert to lowercase and replace spaces with underscores
	id := strings.ToLower(strings.ReplaceAll(name, " ", "_"))
	// Remove special characters
	reg := regexp.MustCompile(`[^a-z0-9_]`)
	return reg.ReplaceAllString(id, "")
}

// detectAuthType determines authentication type from API key
func (yb *YAMLBuilder) detectAuthType(apiKey string) string {
	if apiKey == "none" || apiKey == "" {
		return "none"
	}
	if strings.HasPrefix(apiKey, "sk-") {
		return "bearer_token"
	}
	if strings.HasPrefix(apiKey, "sk-ant-") {
		return "x-api-key"
	}
	return "bearer_token" // Default
}

// queryModelInfo attempts to query model information from API
func (yb *YAMLBuilder) queryModelInfo(endpoint, apiKey, model string) map[string]interface{} {
	modelInfo := make(map[string]interface{})

	// Detect provider type and query accordingly
	if strings.Contains(endpoint, "openai.com") {
		modelInfo = yb.queryOpenAIModel(endpoint, apiKey, model)
	} else if strings.Contains(endpoint, "anthropic.com") {
		modelInfo = yb.queryAnthropicModel(endpoint, apiKey, model)
	} else if strings.Contains(endpoint, "localhost") || strings.Contains(endpoint, "ollama") {
		modelInfo = yb.queryOllamaModel(endpoint, model)
	}

	return modelInfo
}

// queryOpenAIModel queries OpenAI API for model information
func (yb *YAMLBuilder) queryOpenAIModel(endpoint, apiKey, model string) map[string]interface{} {
	info := make(map[string]interface{})

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/models/%s", endpoint, model), nil)
	if err != nil {
		return info
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	resp, err := yb.httpClient.Do(req)
	if err != nil {
		return info
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err == nil {
			if data, ok := result["data"].(map[string]interface{}); ok {
				info["max_tokens"] = data["context_length"]
				info["created"] = data["created"]
			}
		}
	}

	return info
}

// queryOllamaModel queries Ollama API for model information
func (yb *YAMLBuilder) queryOllamaModel(endpoint, model string) map[string]interface{} {
	info := make(map[string]interface{})

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/show", endpoint),
		strings.NewReader(fmt.Sprintf(`{"name": "%s"}`, model)))
	if err != nil {
		return info
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := yb.httpClient.Do(req)
	if err != nil {
		return info
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err == nil {
			if details, ok := result["details"].(map[string]interface{}); ok {
				info["parameter_size"] = details["parameter_size"]
				info["quantization_level"] = details["quantization_level"]
			}
		}
	}

	return info
}

// queryAnthropicModel queries Anthropic API for model information
func (yb *YAMLBuilder) queryAnthropicModel(endpoint, apiKey, model string) map[string]interface{} {
	info := make(map[string]interface{})

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/models", endpoint), nil)
	if err != nil {
		return info
	}

	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := yb.httpClient.Do(req)
	if err != nil {
		return info
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err == nil {
			if models, ok := result["models"].([]interface{}); ok {
				for _, m := range models {
					if modelData, ok := m.(map[string]interface{}); ok {
						if modelData["id"] == model {
							info["max_tokens"] = modelData["max_tokens"]
							info["training_data_cutoff"] = modelData["training_data_cutoff"]
						}
					}
				}
			}
		}
	}

	return info
}

// determineCapabilities infers capabilities from model name and API info
func (yb *YAMLBuilder) determineCapabilities(model string, apiInfo map[string]interface{}) CapabilityProfile {
	// Check if we have a known profile
	if cap, exists := yb.capabilities[model]; exists {
		return cap
	}

	// Attempt to infer capabilities from model name
	cap := CapabilityProfile{
		Reasoning:    5, // Default middle values
		Knowledge:    5,
		Computation:  5,
		Coordination: 5,
	}

	modelLower := strings.ToLower(model)

	if strings.Contains(modelLower, "gpt-4") || strings.Contains(modelLower, "opus") {
		cap.Reasoning = 9
		cap.Knowledge = 9
	} else if strings.Contains(modelLower, "gpt-3") || strings.Contains(modelLower, "sonnet") {
		cap.Reasoning = 7
		cap.Knowledge = 7
	}

	if strings.Contains(modelLower, "turbo") {
		cap.Computation = 8
	}

	if strings.Contains(modelLower, "instruct") {
		cap.Coordination = 7
	}

	return cap
}

// parseRateLimit extracts rate limit information from Other field
func (yb *YAMLBuilder) parseRateLimit(otherField string) RateLimitConfig {
	config := RateLimitConfig{
		RequestsPerMinute: 1000, // Default
		Tier:              "standard",
	}

	rateLimitPattern := regexp.MustCompile(`Rate Limit:\s*(\d+)/min`)
	if matches := rateLimitPattern.FindStringSubmatch(otherField); len(matches) > 1 {
		if rpm, err := strconv.Atoi(matches[1]); err == nil {
			config.RequestsPerMinute = rpm
		}
	}

	// Determine tier based on rate limit
	if config.RequestsPerMinute >= 10000 {
		config.Tier = "premium"
	} else if config.RequestsPerMinute >= 5000 {
		config.Tier = "standard"
	} else {
		config.Tier = "basic"
	}

	return config
}

// parseCostInfo extracts cost information
func (yb *YAMLBuilder) parseCostInfo(otherField, model string) CostConfig {
	config := CostConfig{
		BillingModel: "per_token",
	}

	// Default costs based on model type
	modelLower := strings.ToLower(model)
	if strings.Contains(modelLower, "gpt-4") {
		config.CostPerToken = 0.00003 // $0.03 per 1k tokens
	} else if strings.Contains(modelLower, "gpt-3.5") {
		config.CostPerToken = 0.000002 // $0.002 per 1k tokens
	} else if strings.Contains(modelLower, "claude") {
		config.CostPerToken = 0.000015 // $0.015 per 1k tokens
	} else {
		config.CostPerToken = 0.000001 // $0.001 per 1k tokens (local/free)
	}

	return config
}

// extractMaxTokens extracts maximum token information
func (yb *YAMLBuilder) extractMaxTokens(apiInfo map[string]interface{}, otherField string) int {
	// First check API response
	if maxTokens, ok := apiInfo["max_tokens"].(float64); ok {
		return int(maxTokens)
	}

	if contextLength, ok := apiInfo["context_length"].(float64); ok {
		return int(contextLength)
	}

	// Then parse from Other field
	contextPattern := regexp.MustCompile(`Context:\s*(\d+)k`)
	if matches := contextPattern.FindStringSubmatch(otherField); len(matches) > 1 {
		if ctx, err := strconv.Atoi(matches[1]); err == nil {
			return ctx * 1000
		}
	}

	// Default based on common models
	return 8192 // Conservative default
}

// saveYAML saves configuration as YAML file
func (yb *YAMLBuilder) saveYAML(config *ProviderConfig) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	filename := fmt.Sprintf("%s/%s.yaml", yb.outputDir, config.ID)
	return ioutil.WriteFile(filename, data, 0644)
}