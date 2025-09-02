package selection

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// EnhancedProviderLoader combines CSV and YAML loading with dynamic model discovery
type EnhancedProviderLoader struct {
	csvLoader  *CSVProviderLoader
	yamlLoader *YAMLProviderLoader
}

// CSVProviderLoader handles CSV-based provider loading (existing functionality)
type CSVProviderLoader struct{}

// NewEnhancedProviderLoader creates a new enhanced provider loader
func NewEnhancedProviderLoader() *EnhancedProviderLoader {
	return &EnhancedProviderLoader{
		csvLoader:  &CSVProviderLoader{},
		yamlLoader: NewYAMLProviderLoader(),
	}
}

// LoadProviders loads providers from multiple sources with priority:
// 1. YAML files (if directory exists)
// 2. CSV file (fallback)
func (epl *EnhancedProviderLoader) LoadProviders(csvPath string, yamlDir string) ([]Provider, error) {
	var providers []Provider
	var err error
	
	// Try loading from YAML directory first
	if yamlDir != "" {
		if _, err := os.Stat(yamlDir); err == nil {
			log.Printf("🔍 Attempting to load providers from YAML directory: %s", yamlDir)
			
			yamlProviders, yamlErr := epl.yamlLoader.LoadProvidersFromYAMLDir(yamlDir)
			if yamlErr == nil && len(yamlProviders) > 0 {
				log.Printf("✅ Successfully loaded %d providers from YAML files", len(yamlProviders))
				return yamlProviders, nil
			} else {
				log.Printf("⚠️  Failed to load from YAML directory: %v", yamlErr)
			}
		} else {
			log.Printf("📁 YAML directory not found: %s", yamlDir)
		}
	}
	
	// Fallback to CSV loading
	log.Printf("🔄 Falling back to CSV provider loading: %s", csvPath)
	providers, err = LoadProvidersFromCSV(csvPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load providers from CSV: %w", err)
	}
	
	log.Printf("✅ Successfully loaded %d providers from CSV", len(providers))
	return providers, nil
}

// LoadProvidersWithDynamicModels loads providers and refreshes their models dynamically
func (epl *EnhancedProviderLoader) LoadProvidersWithDynamicModels(csvPath string, yamlDir string) ([]Provider, error) {
	providers, err := epl.LoadProviders(csvPath, yamlDir)
	if err != nil {
		return nil, err
	}
	
	// For CSV-loaded providers, check if they have dynamic model sources
	// This would require extending the CSV format or having companion YAML files
	
	return providers, nil
}

// RefreshProviderModels refreshes models for all providers that support dynamic loading
func (epl *EnhancedProviderLoader) RefreshProviderModels(providers []Provider, yamlDir string) error {
	if yamlDir == "" {
		log.Println("📋 No YAML directory specified, skipping model refresh")
		return nil
	}
	
	log.Printf("🔄 Refreshing models for providers using YAML configs in: %s", yamlDir)
	
	// Clear the model cache
	epl.yamlLoader.RefreshAllModels()
	
	// Reload providers to get fresh models
	refreshedProviders, err := epl.yamlLoader.LoadProvidersFromYAMLDir(yamlDir)
	if err != nil {
		return fmt.Errorf("failed to refresh provider models: %w", err)
	}
	
	// Update the existing providers with fresh model lists
	providerMap := make(map[string][]string)
	for _, provider := range refreshedProviders {
		providerMap[provider.Name] = provider.Models
	}
	
	updated := 0
	for i, provider := range providers {
		if newModels, exists := providerMap[provider.Name]; exists {
			oldCount := len(provider.Models)
			providers[i].Models = newModels
			newCount := len(newModels)
			
			if oldCount != newCount {
				log.Printf("🔄 Updated %s: %d → %d models", provider.Name, oldCount, newCount)
				updated++
			}
		}
	}
	
	log.Printf("✅ Refreshed models for %d providers", updated)
	return nil
}

// GetProviderLoadingStats returns statistics about provider loading
func (epl *EnhancedProviderLoader) GetProviderLoadingStats() map[string]interface{} {
	stats := map[string]interface{}{
		"csv_loader":  "available",
		"yaml_loader": "available",
		"cache_stats": epl.yamlLoader.GetCacheStats(),
	}
	
	return stats
}

// ValidateProviderSources checks if provider sources are accessible
func (epl *EnhancedProviderLoader) ValidateProviderSources(providers []Provider, yamlDir string) map[string][]string {
	issues := make(map[string][]string)
	
	if yamlDir == "" {
		return issues
	}
	
	// Check each provider's YAML file for dynamic loading configuration
	for _, provider := range providers {
		yamlFile := filepath.Join(yamlDir, provider.Name+".yaml")
		if _, err := os.Stat(yamlFile); err == nil {
			// Try to load and validate the YAML configuration
			yamlProvider, err := epl.yamlLoader.LoadProviderFromYAML(yamlFile)
			if err != nil {
				issues[provider.Name] = append(issues[provider.Name], 
					fmt.Sprintf("Failed to load YAML config: %v", err))
			} else if len(yamlProvider.Models) == 0 {
				issues[provider.Name] = append(issues[provider.Name], 
					"No models found in YAML configuration")
			}
		}
	}
	
	return issues
}