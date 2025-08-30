package config

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// YAMLBuilder provides functionality to build YAML configurations
type YAMLBuilder struct {
	config    map[string]interface{}
	csvPath   string
	configDir string
}

// NewYAMLBuilder creates a new YAML builder
func NewYAMLBuilder() *YAMLBuilder {
	return &YAMLBuilder{
		config: make(map[string]interface{}),
	}
}

// SetField sets a field in the YAML configuration
func (y *YAMLBuilder) SetField(key string, value interface{}) *YAMLBuilder {
	y.config[key] = value
	return y
}

// SetNested sets a nested field using dot notation
func (y *YAMLBuilder) SetNested(path string, value interface{}) *YAMLBuilder {
	keys := strings.Split(path, ".")
	current := y.config
	
	for _, key := range keys[:len(keys)-1] {
		if _, exists := current[key]; !exists {
			current[key] = make(map[string]interface{})
		}
		if nested, ok := current[key].(map[string]interface{}); ok {
			current = nested
		} else {
			// Create new nested map if type doesn't match
			current[key] = make(map[string]interface{})
			current = current[key].(map[string]interface{})
		}
	}
	
	current[keys[len(keys)-1]] = value
	return y
}

// Build generates the YAML string
func (y *YAMLBuilder) Build() (string, error) {
	data, err := yaml.Marshal(y.config)
	if err != nil {
		return "", fmt.Errorf("failed to marshal YAML: %w", err)
	}
	return string(data), nil
}

// Reset clears the configuration
func (y *YAMLBuilder) Reset() *YAMLBuilder {
	y.config = make(map[string]interface{})
	return y
}

// GetConfig returns the current configuration map
func (y *YAMLBuilder) GetConfig() map[string]interface{} {
	return y.config
}

// AddTimestamp adds a timestamp field
func (y *YAMLBuilder) AddTimestamp() *YAMLBuilder {
	y.config["timestamp"] = time.Now().Format(time.RFC3339)
	return y
}

// AddMetadata adds metadata section
func (y *YAMLBuilder) AddMetadata(metadata map[string]string) *YAMLBuilder {
	if len(metadata) > 0 {
		y.config["metadata"] = metadata
	}
	return y
}

// ReadCSV reads CSV providers from file
func (y *YAMLBuilder) ReadCSV() ([]CSVProvider, error) {
	if y.csvPath == "" {
		return nil, fmt.Errorf("CSV path not set")
	}

	file, err := os.Open(y.csvPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("empty CSV file")
	}

	// Skip header row
	var providers []CSVProvider
	for i := 1; i < len(records); i++ {
		record := records[i]
		if len(record) >= 5 {
			provider := CSVProvider{
				Name:         record[0],
				Tier:         record[1],
				Endpoint:     record[2],
				ModelsSource: record[3],
				APIKey:       record[4],
			}
			providers = append(providers, provider)
		}
	}

	return providers, nil
}

// BuildFromCSV builds YAML configurations from CSV data
func (y *YAMLBuilder) BuildFromCSV() error {
	providers, err := y.ReadCSV()
	if err != nil {
		return err
	}

	// Create basic configurations from CSV data
	for _, provider := range providers {
		config := ProviderConfig{
			ID:       strings.ToLower(strings.ReplaceAll(provider.Name, " ", "_")),
			Name:     provider.Name,
			Tier:     provider.Tier,
			Endpoint: provider.Endpoint,
			Capabilities: Capabilities{
				Reasoning:    7, // Default values
				Knowledge:    7,
				Computation:  6,
				Coordination: 5,
			},
			CostTracking: CostTracking{
				CostPerToken:   0.00001, // Default cost
				CostPerRequest: 0.001,
				LastUpdated:    time.Now(),
			},
			Metadata: map[string]string{
				"source": "csv",
				"tier":   provider.Tier,
			},
		}

		// Adjust capabilities based on tier
		switch provider.Tier {
		case "official":
			config.Capabilities.Reasoning = 9
			config.Capabilities.Knowledge = 9
			config.Capabilities.Computation = 8
			config.Capabilities.Coordination = 7
			config.CostTracking.CostPerToken = 0.00005
		case "community":
			config.Capabilities.Reasoning = 7
			config.Capabilities.Knowledge = 7
			config.Capabilities.Computation = 6
			config.Capabilities.Coordination = 5
			config.CostTracking.CostPerToken = 0.00002
		case "unofficial":
			config.Capabilities.Reasoning = 5
			config.Capabilities.Knowledge = 6
			config.Capabilities.Computation = 5
			config.Capabilities.Coordination = 4
			config.CostTracking.CostPerToken = 0.00001
		}

		// Save to YAML file if configDir is set
		if y.configDir != "" {
			yamlData, err := yaml.Marshal(config)
			if err != nil {
				continue
			}

			filename := fmt.Sprintf("%s/%s.yaml", y.configDir, config.ID)
			os.MkdirAll(y.configDir, 0755)
			os.WriteFile(filename, yamlData, 0644)
		}
	}

	return nil
}

// SetCSVPath sets the CSV file path
func (y *YAMLBuilder) SetCSVPath(path string) {
	y.csvPath = path
}

// SetConfigDir sets the configuration directory
func (y *YAMLBuilder) SetConfigDir(dir string) {
	y.configDir = dir
}