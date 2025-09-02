package providers

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// ProviderConfig represents a provider configuration
type ProviderConfig struct {
	Name         string            `json:"name"`
	URL          string            `json:"url"`
	Models       []string          `json:"models"`
	Capabilities []string          `json:"capabilities"`
	Priority     int               `json:"priority"`
	Metadata     map[string]string `json:"metadata"`
}

// Provider represents a model provider
type Provider struct {
	Name         string            `json:"name"`
	URL          string            `json:"url"`
	Models       []string          `json:"models"`
	Capabilities []string          `json:"capabilities"`
	Priority     int               `json:"priority"`
	Status       string            `json:"status"`
	Metadata     map[string]string `json:"metadata"`
}

// Manager handles provider operations
type Manager struct {
	providers map[string]*Provider
	mutex     sync.RWMutex
	monitor   *HealthMonitor
}

// NewManager creates a new provider manager
func NewManager() *Manager {
	return &Manager{
		providers: make(map[string]*Provider),
		monitor:   NewHealthMonitor(nil, 30*time.Second),
	}
}

// LoadProvidersFromConfigs loads providers from configuration structs
func (m *Manager) LoadProvidersFromConfigs(configs []ProviderConfig) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for _, providerCfg := range configs {
		provider := &Provider{
			Name:         providerCfg.Name,
			URL:          providerCfg.URL,
			Models:       providerCfg.Models,
			Capabilities: providerCfg.Capabilities,
			Priority:     providerCfg.Priority,
			Status:       "unknown",
			Metadata:     providerCfg.Metadata,
		}
		m.providers[provider.Name] = provider
	}

	return nil
}

// GetProvider returns a provider by name
func (m *Manager) GetProvider(name string) (*Provider, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	provider, exists := m.providers[name]
	return provider, exists
}

// GetAllProviders returns all providers
func (m *Manager) GetAllProviders() map[string]*Provider {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	result := make(map[string]*Provider)
	for k, v := range m.providers {
		result[k] = v
	}
	return result
}

// UpdateProviderStatus updates the status of a provider
func (m *Manager) UpdateProviderStatus(name, status string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	if provider, exists := m.providers[name]; exists {
		provider.Status = status
	}
}

// StartHealthMonitoring starts health monitoring for all providers
func (m *Manager) StartHealthMonitoring() {
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				m.checkProviderHealth()
			}
		}
	}()
}

// checkProviderHealth checks the health of all providers
func (m *Manager) checkProviderHealth() {
	m.mutex.RLock()
	providers := make([]*Provider, 0, len(m.providers))
	for _, provider := range m.providers {
		providers = append(providers, provider)
	}
	m.mutex.RUnlock()

	for _, provider := range providers {
		go func(p *Provider) {
			// Simple health check - just check if we can reach the provider
			if m.isProviderHealthy(p.URL) {
				m.UpdateProviderStatus(p.Name, "healthy")
			} else {
				m.UpdateProviderStatus(p.Name, "unhealthy")
			}
		}(provider)
	}
}

// isProviderHealthy performs a simple health check
func (m *Manager) isProviderHealthy(url string) bool {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode < 500
}

// GetHealthyProviders returns only healthy providers
func (m *Manager) GetHealthyProviders() []*Provider {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var healthy []*Provider
	for _, provider := range m.providers {
		if provider.Status == "healthy" {
			healthy = append(healthy, provider)
		}
	}
	return healthy
}

// ServeHTTP implements http.Handler for provider management endpoints
func (m *Manager) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/providers":
		m.handleGetProviders(w, r)
	case "/providers/health":
		m.handleGetHealth(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (m *Manager) handleGetProviders(w http.ResponseWriter, r *http.Request) {
	providers := m.GetAllProviders()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(providers)
}

func (m *Manager) handleGetHealth(w http.ResponseWriter, r *http.Request) {
	healthy := m.GetHealthyProviders()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"healthy_count": len(healthy),
		"providers":     healthy,
	})
}