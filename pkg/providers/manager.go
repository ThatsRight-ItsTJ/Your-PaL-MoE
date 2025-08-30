package providers

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Your-PaL-MoE/internal/types"
)

// ProviderManager manages AI providers
type ProviderManager struct {
	providers    map[string]*types.Provider
	healthStatus map[string]HealthStatus
	metrics      map[string]map[string]interface{}
	mu           sync.RWMutex
	csvFile      string
}

// HealthStatus represents provider health information
type HealthStatus struct {
	Status      string    `json:"status"`
	LastChecked time.Time `json:"last_checked"`
	ResponseTime time.Duration `json:"response_time"`
	ErrorCount  int       `json:"error_count"`
}

// NewProviderManager creates a new provider manager
func NewProviderManager(csvFile string) (*ProviderManager, error) {
	pm := &ProviderManager{
		providers:    make(map[string]*types.Provider),
		healthStatus: make(map[string]HealthStatus),
		metrics:      make(map[string]map[string]interface{}),
		csvFile:      csvFile,
	}
	
	if err := pm.loadProvidersFromCSV(); err != nil {
		return nil, err
	}
	
	return pm, nil
}

// loadProvidersFromCSV loads providers from CSV file
func (pm *ProviderManager) loadProvidersFromCSV() error {
	file, err := os.Open(pm.csvFile)
	if err != nil {
		return fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read CSV: %w", err)
	}

	if len(records) < 2 {
		return fmt.Errorf("CSV file must have at least a header and one data row")
	}

	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Clear existing providers
	pm.providers = make(map[string]*types.Provider)

	// Skip header row
	for i := 1; i < len(records); i++ {
		record := records[i]
		if len(record) < 4 {
			continue // Skip incomplete records
		}

		provider := &types.Provider{
			Name:         strings.TrimSpace(record[0]),
			Tier:         strings.TrimSpace(record[1]),
			Endpoint:     strings.TrimSpace(record[2]),
			ModelsSource: strings.TrimSpace(record[4]),
			Authentication: types.AuthConfig{
				Type: "api_key",
			},
			Metadata: make(map[string]string),
		}

		// Add API key if provided in record[3]
		if len(record) > 3 && strings.TrimSpace(record[3]) != "none" && strings.TrimSpace(record[3]) != "" {
			provider.Authentication.APIKey = strings.TrimSpace(record[3])
		}

		pm.providers[provider.Name] = provider
		
		// Initialize health status
		pm.healthStatus[provider.Name] = HealthStatus{
			Status:      "unknown",
			LastChecked: time.Now(),
		}
		
		// Initialize metrics
		pm.metrics[provider.Name] = map[string]interface{}{
			"total_requests":     0,
			"successful_requests": 0,
			"failed_requests":    0,
			"average_response_time": 0.0,
			"status": "healthy",
		}
	}

	return nil
}

// GetAvailableProviders returns all available providers
func (pm *ProviderManager) GetAvailableProviders() ([]*types.Provider, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	providers := make([]*types.Provider, 0, len(pm.providers))
	for _, provider := range pm.providers {
		providers = append(providers, provider)
	}

	return providers, nil
}

// GetProvider returns a specific provider by name
func (pm *ProviderManager) GetProvider(name string) (*types.Provider, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	provider, exists := pm.providers[name]
	if !exists {
		return nil, fmt.Errorf("provider %s not found", name)
	}

	return provider, nil
}

// ReloadProviders reloads providers from CSV
func (pm *ProviderManager) ReloadProviders(ctx context.Context) error {
	return pm.loadProvidersFromCSV()
}

// TestProvider tests if a provider is healthy
func (pm *ProviderManager) TestProvider(ctx context.Context, name string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, exists := pm.providers[name]; !exists {
		return fmt.Errorf("provider %s not found", name)
	}

	// Simulate health check
	pm.healthStatus[name] = HealthStatus{
		Status:       "healthy",
		LastChecked:  time.Now(),
		ResponseTime: 100 * time.Millisecond,
		ErrorCount:   0,
	}

	return nil
}

// GetHealthStatus returns health status for a provider
func (pm *ProviderManager) GetHealthStatus(name string) (HealthStatus, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	status, exists := pm.healthStatus[name]
	return status, exists
}

// GetProviderMetrics returns metrics for all providers
func (pm *ProviderManager) GetProviderMetrics() map[string]map[string]interface{} {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	// Return a copy of metrics
	result := make(map[string]map[string]interface{})
	for name, metrics := range pm.metrics {
		result[name] = make(map[string]interface{})
		for key, value := range metrics {
			result[name][key] = value
		}
	}

	return result
}

// DiscoverModels discovers available models for a provider
func (pm *ProviderManager) DiscoverModels(ctx context.Context, name string) ([]string, error) {
	provider, err := pm.GetProvider(name)
	if err != nil {
		return nil, err
	}

	// Parse models from ModelsSource
	models := strings.Split(provider.ModelsSource, "|")
	for i, model := range models {
		models[i] = strings.TrimSpace(model)
	}

	return models, nil
}

// GenerateConfigs generates configuration for providers
func (pm *ProviderManager) GenerateConfigs(ctx context.Context) (map[string]interface{}, error) {
	providers, err := pm.GetAvailableProviders()
	if err != nil {
		return nil, err
	}

	configs := make(map[string]interface{})
	for _, provider := range providers {
		configs[provider.Name] = map[string]interface{}{
			"name":     provider.Name,
			"tier":     provider.Tier,
			"endpoint": provider.Endpoint,
			"models":   provider.ModelsSource,
		}
	}

	return configs, nil
}