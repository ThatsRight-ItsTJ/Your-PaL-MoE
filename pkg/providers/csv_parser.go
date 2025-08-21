package providers

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"
)

// CSVParser handles loading provider configurations from CSV files
type CSVParser struct {
	csvPath string
}

// ProviderConfig represents a provider configuration
type ProviderConfig struct {
	Name         string       `json:"name"`
	Tier         string       `json:"tier"`
	Endpoint     string       `json:"endpoint"`
	ModelsSource ModelsSource `json:"models_source"`
	ApiKey       string       `json:"api_key,omitempty"`
	Headers      map[string]string `json:"headers,omitempty"`
	Enabled      bool         `json:"enabled"`
}

// ModelsSource defines how models are retrieved for a provider
type ModelsSource struct {
	Type  string      `json:"type"`  // "list", "endpoint", "script"
	Value interface{} `json:"value"` // []string, string URL, or script path
}

// NewCSVParser creates a new CSV parser instance
func NewCSVParser(csvPath string) *CSVParser {
	return &CSVParser{
		csvPath: csvPath,
	}
}

// LoadProviders reads and parses the CSV file to load provider configurations
func (p *CSVParser) LoadProviders() (map[string]*ProviderConfig, error) {
	file, err := os.Open(p.csvPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV file: %w", err)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("CSV file is empty")
	}

	// Skip header row
	if len(records) < 2 {
		return nil, fmt.Errorf("CSV file must have at least one provider entry")
	}

	providers := make(map[string]*ProviderConfig)
	
	for i, record := range records[1:] { // Skip header
		if len(record) < 4 {
			return nil, fmt.Errorf("invalid CSV format at line %d: expected at least 4 columns", i+2)
		}

		name := strings.TrimSpace(record[0])
		tier := strings.TrimSpace(record[1])
		endpoint := strings.TrimSpace(record[2])
		modelsStr := strings.TrimSpace(record[3])

		if name == "" || tier == "" || endpoint == "" {
			continue // Skip empty rows
		}

		provider := &ProviderConfig{
			Name:     name,
			Tier:     tier,
			Endpoint: endpoint,
			Enabled:  true,
		}

		// Parse models
		if strings.HasPrefix(modelsStr, "http") {
			// Models from endpoint
			provider.ModelsSource = ModelsSource{
				Type:  "endpoint",
				Value: modelsStr,
			}
		} else if strings.Contains(modelsStr, "|") {
			// Pipe-separated list of models
			models := strings.Split(modelsStr, "|")
			for j, model := range models {
				models[j] = strings.TrimSpace(model)
			}
			provider.ModelsSource = ModelsSource{
				Type:  "list",
				Value: models,
			}
		} else {
			// Single model or script
			if strings.HasPrefix(modelsStr, "./") {
				provider.ModelsSource = ModelsSource{
					Type:  "script",
					Value: modelsStr,
				}
			} else {
				provider.ModelsSource = ModelsSource{
					Type:  "list",
					Value: []string{modelsStr},
				}
			}
		}

		providers[name] = provider
	}

	return providers, nil
}

// ValidateProvider checks if a provider configuration is valid
func (p *CSVParser) ValidateProvider(provider *ProviderConfig) error {
	if provider.Name == "" {
		return fmt.Errorf("provider name cannot be empty")
	}

	validTiers := map[string]bool{
		"official":    true,
		"community":   true,
		"unofficial":  true,
	}

	if !validTiers[provider.Tier] {
		return fmt.Errorf("invalid tier '%s': must be one of official, community, unofficial", provider.Tier)
	}

	if provider.Endpoint == "" {
		return fmt.Errorf("provider endpoint cannot be empty")
	}

	return nil
}

// SaveProviders writes provider configurations back to CSV
func (p *CSVParser) SaveProviders(providers map[string]*ProviderConfig) error {
	file, err := os.Create(p.csvPath)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{"Name", "Tier", "Endpoint", "Model(s)"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write provider records
	for _, provider := range providers {
		var modelsStr string
		
		switch provider.ModelsSource.Type {
		case "list":
			if models, ok := provider.ModelsSource.Value.([]string); ok {
				modelsStr = strings.Join(models, "|")
			}
		case "endpoint":
			if endpoint, ok := provider.ModelsSource.Value.(string); ok {
				modelsStr = endpoint
			}
		case "script":
			if script, ok := provider.ModelsSource.Value.(string); ok {
				modelsStr = script
			}
		}

		record := []string{
			provider.Name,
			provider.Tier,
			provider.Endpoint,
			modelsStr,
		}

		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write provider record: %w", err)
		}
	}

	return nil
}