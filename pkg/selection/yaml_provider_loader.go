package selection

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"

	"github.com/ThatsRight-ItsTJ/Your-PaL-MoE/pkg/providers"
	"gopkg.in/yaml.v2"
)

// YAMLProviderConfig represents the YAML structure for provider configuration
type YAMLProviderConfig struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	URL         string   `yaml:"url"`
	APIKey      string   `yaml:"api_key"`
	Source      string   `yaml:"source"`
	Models      []string `yaml:"models"`
	Priority    int      `yaml:"priority"`
	Enabled     bool     `yaml:"enabled"`
	Type        string   `yaml:"type"`
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
func (ypl *YAMLProviderLoader) LoadProviderFromYAML(filename string) (*providers.ProviderConfig, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read YAML file %s: %w", filename, err)
	}

	var yamlConfig YAMLProviderConfig
	if err := yaml.Unmarshal(data, &yamlConfig); err != nil {
		return nil, fmt.Errorf("failed to parse YAML file %s: %w", filename, err)
	}

	// Convert YAML config to ProviderConfig
	provider := &providers.ProviderConfig{
		Name:        yamlConfig.Name,
		Description: yamlConfig.Description,
		URL:         yamlConfig.URL,
		APIKey:      yamlConfig.APIKey,
		Priority:    yamlConfig.Priority,
		Enabled:     yamlConfig.Enabled,
		Type:        yamlConfig.Type,
		Headers:     yamlConfig.Headers,
	}

	// If source URL is provided, fetch dynamic models
	if yamlConfig.Source != "" {
		log.Printf("Fetching dynamic models from: %s", yamlConfig.Source)
		dynamicModels, err := ypl.dynamicLoader.FetchModels(yamlConfig.Source)
		if err != nil {
			log.Printf("Warning: Failed to fetch dynamic models from %s: %v", yamlConfig.Source, err)
			// Fall back to static models if dynamic fetch fails
			provider.Models = yamlConfig.Models
		} else {
			provider.Models = dynamicModels
			log.Printf("Successfully loaded %d dynamic models for provider %s", len(dynamicModels), yamlConfig.Name)
		}
	} else {
		// Use static models from YAML
		provider.Models = yamlConfig.Models
	}

	return provider, nil
}

// LoadProvidersFromDirectory loads all YAML providers from a directory
func (ypl *YAMLProviderLoader) LoadProvidersFromDirectory(directory string) ([]providers.ProviderConfig, error) {
	var providersList []providers.ProviderConfig

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
func (ypl *YAMLProviderLoader) RefreshDynamicModels(provider *providers.ProviderConfig, sourceURL string) error {
	if sourceURL == "" {
		return fmt.Errorf("no source URL provided for dynamic model refresh")
	}

	models, err := ypl.dynamicLoader.FetchModels(sourceURL)
	if err != nil {
		return fmt.Errorf("failed to refresh dynamic models: %w", err)
	}

	provider.Models = models
	log.Printf("Refreshed %d models for provider %s", len(models), provider.Name)
	return nil
}

// ValidateYAMLProvider validates a YAML provider configuration
func (ypl *YAMLProviderLoader) ValidateYAMLProvider(provider *YAMLProviderConfig) []string {
	var issues []string

	if provider.Name == "" {
		issues = append(issues, "provider name is required")
	}

	if provider.URL == "" && provider.Source == "" {
		issues = append(issues, "either URL or source is required")
	}

	if len(provider.Models) == 0 && provider.Source == "" {
		issues = append(issues, "either models list or source URL is required")
	}

	return issues
}