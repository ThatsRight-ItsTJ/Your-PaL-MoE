package selection

import (
	"fmt"
	"log"
	"strings"

	"github.com/ThatsRight-ItsTJ/Your-PaL-MoE/pkg/config"
)

type IntegratedProviderSystem struct {
	csvLoader  *EnhancedProviderLoader
	yamlLoader *YAMLProviderLoader
	providers  []config.ProviderConfig
}

// NewIntegratedProviderSystem creates a new integrated provider system
func NewIntegratedProviderSystem() *IntegratedProviderSystem {
	return &IntegratedProviderSystem{
		csvLoader:  NewEnhancedProviderLoader(),
		yamlLoader: NewYAMLProviderLoader(),
		providers:  make([]config.ProviderConfig, 0),
	}
}

// LoadAllProviders loads providers from both CSV and YAML sources
func (ips *IntegratedProviderSystem) LoadAllProviders(csvFile, yamlDir string) error {
	var allProviders []config.ProviderConfig

	// Load CSV providers
	if csvFile != "" {
		csvProviders, err := ips.csvLoader.LoadProviders(csvFile)
		if err != nil {
			log.Printf("Warning: Failed to load CSV providers: %v", err)
		} else {
			allProviders = append(allProviders, csvProviders...)
			log.Printf("Loaded %d providers from CSV", len(csvProviders))
		}
	}

	// Load YAML providers
	if yamlDir != "" {
		yamlProviders, err := ips.yamlLoader.LoadProvidersFromDirectory(yamlDir)
		if err != nil {
			log.Printf("Warning: Failed to load YAML providers: %v", err)
		} else {
			allProviders = append(allProviders, yamlProviders...)
			log.Printf("Loaded %d providers from YAML", len(yamlProviders))
		}
	}

	ips.providers = allProviders
	return nil
}

// GetProviders returns all loaded providers
func (ips *IntegratedProviderSystem) GetProviders() []config.ProviderConfig {
	return ips.providers
}

// GetEnabledProviders returns only enabled providers
func (ips *IntegratedProviderSystem) GetEnabledProviders() []config.ProviderConfig {
	// Phase 1: treat all as enabled
	return ips.providers
}

// GetProviderByName finds a provider by name
func (ips *IntegratedProviderSystem) GetProviderByName(name string) (*config.ProviderConfig, error) {
	for i := range ips.providers {
		if strings.EqualFold(ips.providers[i].Name, name) {
			return &ips.providers[i], nil
		}
	}
	return nil, fmt.Errorf("provider %s not found", name)
}

// RefreshProviders refreshes all provider data
func (ips *IntegratedProviderSystem) RefreshProviders(csvFile, yamlDir string) error {
	return ips.LoadAllProviders(csvFile, yamlDir)
}

// GetStats returns statistics about the integrated system
func (ips *IntegratedProviderSystem) GetStats() map[string]interface{} {
	stats := make(map[string]interface{})

	total := len(ips.providers)
	enabled := len(ips.GetEnabledProviders())

	stats["total_providers"] = total
	stats["enabled_providers"] = enabled
	stats["disabled_providers"] = total - enabled

	// Get CSV stats
	csvStats := ips.csvLoader.GetProviderStats()
	stats["csv_stats"] = csvStats

	return stats
}

// ValidateProviders checks if all providers have required fields
func (ips *IntegratedProviderSystem) ValidateProviders() []string {
	var issues []string

	for i, provider := range ips.providers {
		if provider.Name == "" {
			issues = append(issues, fmt.Sprintf("Provider %d: missing name", i))
		}
		if provider.Endpoint == "" {
			issues = append(issues, fmt.Sprintf("Provider %s: missing endpoint", provider.Name))
		}
	}

	return issues
}

// GetProvidersByPriority returns providers sorted by priority
func (ips *IntegratedProviderSystem) GetProvidersByPriority() []config.ProviderConfig {
	// Create a copy to avoid modifying original slice
	sortedProviders := make([]config.ProviderConfig, len(ips.providers))
	copy(sortedProviders, ips.providers)

	// Simple bubble sort by priority (higher priority first)
	for i := 0; i < len(sortedProviders)-1; i++ {
		for j := 0; j < len(sortedProviders)-i-1; j++ {
			if sortedProviders[j].Priority < sortedProviders[j+1].Priority {
				sortedProviders[j], sortedProviders[j+1] = sortedProviders[j+1], sortedProviders[j]
			}
		}
	}

	return sortedProviders
}
