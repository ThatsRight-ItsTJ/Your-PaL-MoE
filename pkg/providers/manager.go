package providers

import (
	"context"
	"sync"
	"time"
)

// Manager handles provider operations and health monitoring
type Manager struct {
	providers     map[string]*Provider
	healthMonitor *HealthMonitor
	mu            sync.RWMutex
}

// Provider represents a configured provider
type Provider struct {
	Name         string
	URL          string
	Models       []string
	Capabilities []string
	Priority     int
	Metadata     map[string]interface{}
	IsHealthy    bool
	LastChecked  time.Time
}

// NewManager creates a new provider manager
func NewManager() *Manager {
	return &Manager{
		providers:     make(map[string]*Provider),
		healthMonitor: NewHealthMonitor(30*time.Second, 5*time.Minute),
	}
}

// LoadFromConfig loads providers from configuration
func (m *Manager) LoadFromConfig(configs []ProviderConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, config := range configs {
		provider := &Provider{
			Name:         config.Name,
			URL:          config.URL,
			Models:       config.Models,
			Capabilities: config.Capabilities,
			Priority:     config.Priority,
			Metadata:     config.Metadata,
			IsHealthy:    true,
			LastChecked:  time.Now(),
		}

		m.providers[config.Name] = provider
		
		// Start health monitoring for this provider
		if config.URL != "" {
			m.healthMonitor.AddProvider(config.Name, config.URL)
		}
	}

	return nil
}

// GetProvider returns a provider by name
func (m *Manager) GetProvider(name string) (*Provider, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	provider, exists := m.providers[name]
	return provider, exists
}

// GetAllProviders returns all providers
func (m *Manager) GetAllProviders() map[string]*Provider {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	result := make(map[string]*Provider)
	for name, provider := range m.providers {
		result[name] = provider
	}
	return result
}

// GetHealthyProviders returns only healthy providers
func (m *Manager) GetHealthyProviders() map[string]*Provider {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	result := make(map[string]*Provider)
	for name, provider := range m.providers {
		if provider.IsHealthy {
			result[name] = provider
		}
	}
	return result
}

// UpdateProviderHealth updates the health status of a provider
func (m *Manager) UpdateProviderHealth(name string, isHealthy bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if provider, exists := m.providers[name]; exists {
		provider.IsHealthy = isHealthy
		provider.LastChecked = time.Now()
	}
}

// StartHealthMonitoring starts the health monitoring process
func (m *Manager) StartHealthMonitoring(ctx context.Context) {
	m.healthMonitor.Start(ctx, m.UpdateProviderHealth)
}

// StopHealthMonitoring stops the health monitoring process
func (m *Manager) StopHealthMonitoring() {
	m.healthMonitor.Stop()
}

// GetProviderModels returns models for a specific provider
func (m *Manager) GetProviderModels(name string) ([]string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	if provider, exists := m.providers[name]; exists {
		return provider.Models, true
	}
	return nil, false
}

// UpdateProviderModels updates the models for a provider
func (m *Manager) UpdateProviderModels(name string, models []string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if provider, exists := m.providers[name]; exists {
		provider.Models = models
	}
}