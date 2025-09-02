package selection

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ThatsRight-ItsTJ/Your-PaL-MoE/pkg/providers"
)

// EnhancedProviderLoader handles loading providers with enhanced capabilities
type EnhancedProviderLoader struct {
	providers       []providers.ProviderConfig
	lastLoadTime    time.Time
	cacheExpiration time.Duration
}

// NewEnhancedProviderLoader creates a new enhanced provider loader
func NewEnhancedProviderLoader() *EnhancedProviderLoader {
	return &EnhancedProviderLoader{
		providers:       make([]providers.ProviderConfig, 0),
		cacheExpiration: 5 * time.Minute,
	}
}

// LoadProvidersFromCSV loads providers from CSV file
func LoadProvidersFromCSV(filename string) ([]providers.ProviderConfig, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	var providersList []providers.ProviderConfig

	// Read header
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %w", err)
	}

	// Create column index map
	columnIndex := make(map[string]int)
	for i, col := range header {
		columnIndex[strings.ToLower(strings.TrimSpace(col))] = i
	}

	// Read data rows
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Error reading CSV row: %v", err)
			continue
		}

		provider := providers.ProviderConfig{}

		// Map CSV columns to provider fields
		if idx, exists := columnIndex["name"]; exists && idx < len(record) {
			provider.Name = strings.TrimSpace(record[idx])
		}
		if idx, exists := columnIndex["url"]; exists && idx < len(record) {
			provider.URL = strings.TrimSpace(record[idx])
		}
		if idx, exists := columnIndex["api_key"]; exists && idx < len(record) {
			provider.APIKey = strings.TrimSpace(record[idx])
		}
		if idx, exists := columnIndex["priority"]; exists && idx < len(record) {
			if priority, err := strconv.Atoi(strings.TrimSpace(record[idx])); err == nil {
				provider.Priority = priority
			}
		}
		if idx, exists := columnIndex["enabled"]; exists && idx < len(record) {
			enabled := strings.ToLower(strings.TrimSpace(record[idx]))
			provider.Enabled = enabled == "true" || enabled == "1" || enabled == "yes"
		} else {
			provider.Enabled = true // Default to enabled
		}

		if provider.Name != "" {
			providersList = append(providersList, provider)
		}
	}

	return providersList, nil
}

// LoadProviders loads providers with caching
func (epl *EnhancedProviderLoader) LoadProviders(filename string) ([]providers.ProviderConfig, error) {
	// Check if cache is still valid
	if time.Since(epl.lastLoadTime) < epl.cacheExpiration && len(epl.providers) > 0 {
		return epl.providers, nil
	}

	// Load fresh data
	newProviders, err := LoadProvidersFromCSV(filename)
	if err != nil {
		return nil, err
	}

	epl.providers = newProviders
	epl.lastLoadTime = time.Now()

	return epl.providers, nil
}

// GetProviderByName finds a provider by name
func (epl *EnhancedProviderLoader) GetProviderByName(name string) (*providers.ProviderConfig, error) {
	for _, provider := range epl.providers {
		if strings.EqualFold(provider.Name, name) {
			return &provider, nil
		}
	}
	return nil, fmt.Errorf("provider %s not found", name)
}

// GetEnabledProviders returns only enabled providers
func (epl *EnhancedProviderLoader) GetEnabledProviders() []providers.ProviderConfig {
	var enabled []providers.ProviderConfig
	for _, provider := range epl.providers {
		if provider.Enabled {
			enabled = append(enabled, provider)
		}
	}
	return enabled
}

// RefreshProviders forces a refresh of provider data
func (epl *EnhancedProviderLoader) RefreshProviders(filename string) error {
	epl.lastLoadTime = time.Time{} // Reset cache
	_, err := epl.LoadProviders(filename)
	return err
}

// GetProviderStats returns statistics about loaded providers
func (epl *EnhancedProviderLoader) GetProviderStats() map[string]interface{} {
	stats := make(map[string]interface{})
	
	total := len(epl.providers)
	enabled := len(epl.GetEnabledProviders())
	
	stats["total_providers"] = total
	stats["enabled_providers"] = enabled
	stats["disabled_providers"] = total - enabled
	stats["last_load_time"] = epl.lastLoadTime
	stats["cache_expiration"] = epl.cacheExpiration
	
	return stats
}