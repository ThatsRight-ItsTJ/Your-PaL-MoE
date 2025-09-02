package selection

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"

	"github.com/ThatsRight-ItsTJ/Your-PaL-MoE/pkg/config"
	"gopkg.in/yaml.v2"
)

// YAMLProviderConfig represents the YAML structure for provider configuration
type YAMLProviderConfig struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description"`
	Endpoint    string            `yaml:"url"`
	APIKey      string            `yaml:"api_key"`
	Source      string            `yaml:"source"`
	Models      []string          `yaml:"models"`
	Priority    int               `yaml:"priority"`
	Enabled     bool              `yaml:"enabled"`
	Type        string            `yaml:"type"`
	Headers     map[string]string `yaml:"headers"`
}

// YAMLProviderLoader handles loading providers from YAML files
type YAMLProviderLoader struct {
	dynamicLoader *DynamicModelLoader
}

// NewYAMLProviderLoader creates a new YAML provider loader
func NewYAMLProviderLoader() *YAMLProviderLoader {
	return &YAMLProviderLoader{
		dynamicLoader: NewDynamicModelLoader(),
	}
}

// LoadProviderFromYAML loads a single provider from YAML file
func (ypl *YAMLProviderLoader) LoadProviderFromYAML(filename string) (*config.ProviderConfig, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read YAML file %s: %w", filename, err)
	}

	var yamlConfig YAMLProviderConfig
	if err := yaml.Unmarshal(data, &yamlConfig); err != nil {
		return nil, fmt.Errorf("failed to parse YAML file %s: %w", filename, err)
	}

	// Convert YAML config to canonical ProviderConfig (config.ProviderConfig)
	provider := &config.ProviderConfig{
		Name:     yamlConfig.Name,
		Endpoint: yamlConfig.Endpoint,
		APIKey:   yamlConfig.APIKey,
		Priority: yamlConfig.Priority,
		Metadata: map[string]string{
			"source": yamlConfig.Source,
		},
	}
	// Do not populate optional fields like Capabilities or CostTracking here.
	// Models are not stored on the canonical ProviderConfig in this pass.
	return provider, nil
}

// LoadProvidersFromDirectory loads all YAML providers from a directory
func (ypl *YAMLProviderLoader) LoadProvidersFromDirectory(directory string) ([]config.ProviderConfig, error) {
	var providersList []config.ProviderConfig

	files, err := ioutil.ReadDir(directory)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", directory, err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		// Check if file has YAML extension
		if !strings.HasSuffix(strings.ToLower(file.Name()), ".yaml") &&
			!strings.HasSuffix(strings.ToLower(file.Name()), ".yml") {
			continue
		}

		filePath := filepath.Join(directory, file.Name())
		provider, err := ypl.LoadProviderFromYAML(filePath)
		if err != nil {
			log.Printf("Warning: Failed to load provider from %s: %v", filePath, err)
			continue
		}

		if provider.Name != "" {
			providersList = append(providersList, *provider)
		}
	}

	return providersList, nil
}

// RefreshDynamicModels refreshes models for providers with dynamic sources
func (ypl *YAMLProviderLoader) RefreshDynamicModels(provider *config.ProviderConfig, sourceURL string) error {
	if sourceURL == "" {
		return fmt.Errorf("no source URL provided for dynamic model refresh")
	}

	// Phase 1: dynamic models loading is not implemented; skip
	log.Printf("Dynamic model refresh skipped for provider %s in Phase 1", provider.Name)
	return nil
}

// ValidateYAMLProvider validates a YAML provider configuration
func (ypl *YAMLProviderLoader) ValidateYAMLProvider(provider *YAMLProviderConfig) []string {
	var issues []string

	if provider.Name == "" {
		issues = append(issues, "provider name is required")
	}

	if provider.Endpoint == "" && provider.Source == "" {
		issues = append(issues, "either URL or source is required")
	}

	if len(provider.Models) == 0 && provider.Source == "" {
		issues = append(issues, "either models list or source URL is required")
	}

	return issues
}
