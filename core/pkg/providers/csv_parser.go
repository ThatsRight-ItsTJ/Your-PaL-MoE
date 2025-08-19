package providers

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

type ProviderConfig struct {
	Name           string            `json:"name"`
	Tier           string            `json:"tier"` // official, community, unofficial
	Endpoint       string            `json:"endpoint"`
	ModelsSource   ModelsSource      `json:"models_source"`
	LastUpdated    time.Time         `json:"last_updated"`
	Health         HealthStatus      `json:"health"`
	Capabilities   []string          `json:"capabilities"`
	Authentication AuthConfig        `json:"authentication"`
}

type ModelsSource struct {
	Type  string      `json:"type"`  // "url" or "list"
	Value interface{} `json:"value"` // URL string or []string
}

type HealthStatus struct {
	Status       string    `json:"status"` // healthy, degraded, down, unknown
	LastCheck    time.Time `json:"last_check"`
	ResponseTime int64     `json:"response_time_ms"`
	ErrorMessage string    `json:"error_message,omitempty"`
}

type AuthConfig struct {
	Type     string `json:"type"`     // bearer_token, api_key, cookie, none
	EnvVar   string `json:"env_var"`  // Environment variable name
	Header   string `json:"header"`   // Header name for auth
	Required bool   `json:"required"`
}

type CSVParser struct {
	filePath  string
	providers map[string]*ProviderConfig
}

func NewCSVParser(filePath string) *CSVParser {
	return &CSVParser{
		filePath:  filePath,
		providers: make(map[string]*ProviderConfig),
	}
}

func (p *CSVParser) LoadProviders() (map[string]*ProviderConfig, error) {
	file, err := os.Open(p.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comment = '#'
	reader.TrimLeadingSpace = true

	// Read header
	headers, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV headers: %w", err)
	}

	// Validate headers
	expectedHeaders := []string{"Name", "Tier", "Endpoint", "Model(s)"}
	if len(headers) != len(expectedHeaders) {
		return nil, fmt.Errorf("invalid CSV headers. Expected: %v, Got: %v", expectedHeaders, headers)
	}

	// Process rows
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read CSV record: %w", err)
		}

		// Skip empty rows
		if len(record) == 0 || strings.TrimSpace(record[0]) == "" {
			continue
		}

		provider, err := p.parseProviderRecord(record)
		if err != nil {
			return nil, fmt.Errorf("failed to parse provider record %v: %w", record, err)
		}

		p.providers[provider.Name] = provider
	}

	return p.providers, nil
}

func (p *CSVParser) parseProviderRecord(record []string) (*ProviderConfig, error) {
	if len(record) != 4 {
		return nil, fmt.Errorf("invalid record length: expected 4, got %d", len(record))
	}

	name := strings.TrimSpace(record[0])
	tier := strings.TrimSpace(strings.ToLower(record[1]))
	endpoint := strings.TrimSpace(record[2])
	modelsField := strings.TrimSpace(record[3])

	// Validate tier
	validTiers := []string{"official", "community", "unofficial"}
	if !contains(validTiers, tier) {
		return nil, fmt.Errorf("invalid tier '%s' for provider '%s'. Valid tiers: %v", tier, name, validTiers)
	}

	// Parse models source
	modelsSource := p.parseModelsSource(modelsField)

	// Determine authentication requirements
	authConfig := p.determineAuthConfig(name, tier, endpoint)

	provider := &ProviderConfig{
		Name:           name,
		Tier:           tier,
		Endpoint:       endpoint,
		ModelsSource:   modelsSource,
		LastUpdated:    time.Now(),
		Health:         HealthStatus{Status: "unknown"},
		Authentication: authConfig,
	}

	return provider, nil
}

func (p *CSVParser) parseModelsSource(modelsField string) ModelsSource {
	// Check if it's a URL
	if strings.HasPrefix(modelsField, "http://") || strings.HasPrefix(modelsField, "https://") {
		return ModelsSource{
			Type:  "url",
			Value: modelsField,
		}
	}

	// Parse as pipe-delimited list
	models := []string{}
	if modelsField != "" {
		rawModels := strings.Split(modelsField, "|")
		for _, model := range rawModels {
			model = strings.TrimSpace(model)
			if model != "" {
				models = append(models, model)
			}
		}
	}

	return ModelsSource{
		Type:  "list",
		Value: models,
	}
}

func (p *CSVParser) determineAuthConfig(name, tier, endpoint string) AuthConfig {
	// Default configurations based on provider patterns
	lowerName := strings.ToLower(name)

	switch {
	case tier == "official":
		switch {
		case strings.Contains(lowerName, "openai"):
			return AuthConfig{Type: "bearer_token", EnvVar: "OPENAI_API_KEY", Header: "Authorization", Required: true}
		case strings.Contains(lowerName, "anthropic"):
			return AuthConfig{Type: "bearer_token", EnvVar: "ANTHROPIC_API_KEY", Header: "Authorization", Required: true}
		case strings.Contains(lowerName, "azure"):
			return AuthConfig{Type: "api_key", EnvVar: "AZURE_OPENAI_KEY", Header: "api-key", Required: true}
		case strings.Contains(lowerName, "google"):
			return AuthConfig{Type: "api_key", EnvVar: "GOOGLE_API_KEY", Header: "x-goog-api-key", Required: true}
		default:
			return AuthConfig{Type: "bearer_token", EnvVar: strings.ToUpper(lowerName) + "_API_KEY", Header: "Authorization", Required: true}
		}
	case tier == "community":
		switch {
		case strings.Contains(lowerName, "pollinations"):
			return AuthConfig{Type: "none", Required: false}
		case strings.Contains(lowerName, "huggingface"):
			return AuthConfig{Type: "bearer_token", EnvVar: "HUGGINGFACE_API_KEY", Header: "Authorization", Required: false}
		case strings.Contains(lowerName, "replicate"):
			return AuthConfig{Type: "bearer_token", EnvVar: "REPLICATE_API_TOKEN", Header: "Authorization", Required: true}
		default:
			return AuthConfig{Type: "bearer_token", EnvVar: strings.ToUpper(lowerName) + "_API_KEY", Header: "Authorization", Required: false}
		}
	case tier == "unofficial":
		if strings.Contains(lowerName, "bing") {
			return AuthConfig{Type: "cookie", EnvVar: "BING_COOKIE_U", Header: "Cookie", Required: true}
		}
		return AuthConfig{Type: "custom", EnvVar: strings.ToUpper(lowerName) + "_AUTH", Required: false}
	}

	return AuthConfig{Type: "none", Required: false}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// ValidateCSV validates the CSV file format and content
func (p *CSVParser) ValidateCSV() error {
	providers, err := p.LoadProviders()
	if err != nil {
		return err
	}

	// Check for duplicate names
	names := make(map[string]bool)
	for name := range providers {
		if names[name] {
			return fmt.Errorf("duplicate provider name: %s", name)
		}
		names[name] = true
	}

	// Validate endpoints
	for name, provider := range providers {
		if err := p.validateEndpoint(provider); err != nil {
			return fmt.Errorf("invalid endpoint for provider %s: %w", name, err)
		}
	}

	return nil
}

func (p *CSVParser) validateEndpoint(provider *ProviderConfig) error {
	endpoint := provider.Endpoint

	if endpoint == "" {
		return fmt.Errorf("endpoint cannot be empty")
	}

	// Validate URL endpoints
	if strings.HasPrefix(endpoint, "http://") || strings.HasPrefix(endpoint, "https://") {
		// Basic URL validation - could be enhanced with proper URL parsing
		return nil
	}

	// Validate script paths
	if strings.HasPrefix(endpoint, "./scripts/") {
		// Check if it's a reasonable script path
		if !strings.Contains(endpoint, ".py") && !strings.Contains(endpoint, ".js") && !strings.Contains(endpoint, ".sh") {
			return fmt.Errorf("script endpoint must have a valid extension (.py, .js, .sh)")
		}
		return nil
	}

	// Check for localhost URLs (for local services like Ollama)
	if strings.Contains(endpoint, "localhost") || strings.Contains(endpoint, "127.0.0.1") {
		return nil
	}

	return fmt.Errorf("endpoint must be a valid HTTP(S) URL or script path")
}