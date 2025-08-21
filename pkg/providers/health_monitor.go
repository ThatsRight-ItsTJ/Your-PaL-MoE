package providers

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// HealthStatus represents the health status of a provider
type HealthStatus struct {
	Status       StatusInfo    `json:"status"`
	LastChecked  time.Time     `json:"last_checked"`
	ResponseTime time.Duration `json:"response_time"`
	ErrorCount   int           `json:"error_count"`
	Uptime       float64       `json:"uptime"`
}

// StatusInfo contains detailed status information
type StatusInfo struct {
	Status  string `json:"status"`  // "healthy", "degraded", "unhealthy"
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// HealthMonitor monitors provider health and manages circuit breakers
type HealthMonitor struct {
	providers map[string]*ProviderConfig
	statuses  map[string]*HealthStatus
	interval  time.Duration
	mutex     sync.RWMutex
	client    *http.Client
}

// NewHealthMonitor creates a new health monitor instance
func NewHealthMonitor(config interface{}, interval time.Duration) *HealthMonitor {
	return &HealthMonitor{
		providers: make(map[string]*ProviderConfig),
		statuses:  make(map[string]*HealthStatus),
		interval:  interval,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// RegisterProvider adds a provider to health monitoring
func (h *HealthMonitor) RegisterProvider(provider *ProviderConfig) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	
	h.providers[provider.Name] = provider
	h.statuses[provider.Name] = &HealthStatus{
		Status: StatusInfo{
			Status:  "unknown",
			Message: "Not yet checked",
		},
		LastChecked: time.Time{},
	}
}

// Start begins the health monitoring process
func (h *HealthMonitor) Start(ctx context.Context) {
	ticker := time.NewTicker(h.interval)
	defer ticker.Stop()

	// Initial health check
	h.checkAllProviders()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			h.checkAllProviders()
		}
	}
}

// checkAllProviders performs health checks on all registered providers
func (h *HealthMonitor) checkAllProviders() {
	h.mutex.RLock()
	providers := make([]*ProviderConfig, 0, len(h.providers))
	for _, provider := range h.providers {
		providers = append(providers, provider)
	}
	h.mutex.RUnlock()

	// Check providers in parallel
	var wg sync.WaitGroup
	for _, provider := range providers {
		wg.Add(1)
		go func(p *ProviderConfig) {
			defer wg.Done()
			h.checkProviderHealth(p)
		}(provider)
	}
	wg.Wait()
}

// checkProviderHealth performs a health check on a single provider
func (h *HealthMonitor) checkProviderHealth(provider *ProviderConfig) {
	start := time.Now()
	status := &HealthStatus{
		LastChecked: start,
	}

	// Perform health check based on provider tier
	var err error
	switch provider.Tier {
	case "official":
		err = h.checkOfficialProvider(provider)
	case "community":
		err = h.checkCommunityProvider(provider)
	case "unofficial":
		err = h.checkUnofficialProvider(provider)
	default:
		err = fmt.Errorf("unknown provider tier: %s", provider.Tier)
	}

	status.ResponseTime = time.Since(start)

	// Update status based on check result
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if existingStatus, exists := h.statuses[provider.Name]; exists {
		status.ErrorCount = existingStatus.ErrorCount
		status.Uptime = existingStatus.Uptime
	}

	if err != nil {
		status.Status = StatusInfo{
			Status:  "unhealthy",
			Message: err.Error(),
			Code:    500,
		}
		status.ErrorCount++
		status.Uptime = status.Uptime * 0.95 // Decrease uptime
	} else {
		status.Status = StatusInfo{
			Status:  "healthy",
			Message: "OK",
			Code:    200,
		}
		status.Uptime = status.Uptime*0.99 + 0.01 // Increase uptime gradually
		if status.Uptime > 1.0 {
			status.Uptime = 1.0
		}
	}

	h.statuses[provider.Name] = status
}

// checkOfficialProvider performs health check for official API providers
func (h *HealthMonitor) checkOfficialProvider(provider *ProviderConfig) error {
	healthURL := provider.Endpoint
	if healthURL == "https://api.openai.com/v1" {
		healthURL = "https://api.openai.com/v1/models"
	}

	req, err := http.NewRequest("GET", healthURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("health check returned status %d", resp.StatusCode)
	}

	return nil
}

// checkCommunityProvider performs health check for community providers
func (h *HealthMonitor) checkCommunityProvider(provider *ProviderConfig) error {
	req, err := http.NewRequest("GET", provider.Endpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		return fmt.Errorf("health check returned status %d", resp.StatusCode)
	}

	return nil
}

// checkUnofficialProvider performs health check for unofficial providers
func (h *HealthMonitor) checkUnofficialProvider(provider *ProviderConfig) error {
	// For script-based providers, we can't easily health check
	// Just assume they're healthy if the script exists
	if provider.ModelsSource.Type == "script" {
		return nil // Scripts are checked during execution
	}

	// For endpoint-based unofficial providers
	req, err := http.NewRequest("HEAD", provider.Endpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		return fmt.Errorf("health check returned status %d", resp.StatusCode)
	}

	return nil
}

// GetHealthStatus returns the health status of a specific provider
func (h *HealthMonitor) GetHealthStatus(name string) (*HealthStatus, bool) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	
	status, exists := h.statuses[name]
	return status, exists
}

// GetAllHealthStatuses returns health statuses for all providers
func (h *HealthMonitor) GetAllHealthStatuses() map[string]*HealthStatus {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	
	result := make(map[string]*HealthStatus)
	for name, status := range h.statuses {
		result[name] = status
	}
	return result
}

// IsProviderHealthy checks if a provider is currently healthy
func (h *HealthMonitor) IsProviderHealthy(name string) bool {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	
	status, exists := h.statuses[name]
	if !exists {
		return false
	}
	
	return status.Status.Status == "healthy"
}

// GetHealthyProviders returns a list of currently healthy providers
func (h *HealthMonitor) GetHealthyProviders() []string {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	
	var healthy []string
	for name, status := range h.statuses {
		if status.Status.Status == "healthy" {
			healthy = append(healthy, name)
		}
	}
	return healthy
}