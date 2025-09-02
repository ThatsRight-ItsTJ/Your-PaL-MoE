package providers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/ThatsRight-ItsTJ/Your-PaL-MoE/pkg/config"
)

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
		monitor:   NewHealthMonitor(),
	}
}

// LoadProviders loads providers from configuration
func (m *Manager) LoadProviders(cfg *config.Config) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for _, providerCfg := range cfg.Providers {
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
			status := m.monitor.CheckHealth(p.Name, p.URL)
			m.UpdateProviderStatus(p.Name, string(status))
		}(provider)
	}
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