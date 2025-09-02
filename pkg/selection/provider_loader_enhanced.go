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

	"github.com/ThatsRight-ItsTJ/Your-PaL-MoE/pkg/config"
)

type EnhancedProviderLoader struct {
	providers       []config.ProviderConfig
	lastLoadTime    time.Time
	cacheExpiration time.Duration
}

// NewEnhancedProviderLoader creates a new enhanced provider loader
func NewEnhancedProviderLoader() *EnhancedProviderLoader {
	return &EnhancedProviderLoader{
		providers:       make([]config.ProviderConfig, 0),
		cacheExpiration: 5 * time.Minute,
	}
}

// LoadProvidersFromCSV loads providers from CSV file
func LoadProvidersFromCSV(filename string) ([]config.ProviderConfig, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	var providersList []config.ProviderConfig

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

		provider := config.ProviderConfig{}

		// Map CSV columns to provider fields
		if idx, exists := columnIndex["name"]; exists && idx < len(record) {
			provider.Name = strings.TrimSpace(record[idx])
		}
		if idx, exists := columnIndex["tier"]; exists && idx < len(record) {
			provider.Tier = strings.TrimSpace(record[idx])
		}
		if idx, exists := columnIndex["endpoint"]; exists && idx < len(record) {
			provider.Endpoint = strings.TrimSpace(record[idx])
		}
		if idx, exists := columnIndex["api_key"]; exists && idx < len(record) {
			provider.APIKey = strings.TrimSpace(record[idx])
		}
		if idx, exists := columnIndex["priority"]; exists && idx < len(record) {
			if priority, err := strconv.Atoi(strings.TrimSpace(record[idx])); err == nil {
				provider.Priority = priority
			}
		}

		// Phase 1: ignore Enabled/URL; only require Name for validity
		if provider.Name != "" {
			providersList = append(providersList, provider)
		}
	}

	return providersList, nil
}

// LoadProviders loads providers with caching
func (epl *EnhancedProviderLoader) LoadProviders(filename string) ([]config.ProviderConfig, error) {
	if time.Since(epl.lastLoadTime) < epl.cacheExpiration && len(epl.providers) > 0 {
		return epl.providers, nil
	}

	newProviders, err := LoadProvidersFromCSV(filename)
	if err != nil {
		return nil, err
	}

	epl.providers = newProviders
	epl.lastLoadTime = time.Now()

	return epl.providers, nil
}

// GetProviderByName finds a provider by name
func (epl *EnhancedProviderLoader) GetProviderByName(name string) (*config.ProviderConfig, error) {
	for i := range epl.providers {
		if strings.EqualFold(epl.providers[i].Name, name) {
			return &epl.providers[i], nil
		}
	}
	return nil, fmt.Errorf("provider %s not found", name)
}

// GetEnabledProviders returns only enabled providers
func (epl *EnhancedProviderLoader) GetEnabledProviders() []config.ProviderConfig {
	// Phase 1: treat all as enabled
	return epl.providers
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
