package providers

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type HealthMonitor struct {
	httpClient     *http.Client
	checkInterval  time.Duration
	providers      map[string]*ProviderConfig
	circuitBreakers map[string]*CircuitBreaker
	mu             sync.RWMutex
	stopChan       chan struct{}
	callbacks      []HealthCallback
}

type CircuitBreaker struct {
	FailureCount    int       `json:"failure_count"`
	FailureThreshold int      `json:"failure_threshold"`
	LastFailure     time.Time `json:"last_failure"`
	State           string    `json:"state"` // closed, open, half-open
	TimeoutDuration time.Duration `json:"timeout_duration"`
}

type HealthCallback func(providerName string, health HealthStatus)

type HealthCheckResult struct {
	Provider     string        `json:"provider"`
	Status       string        `json:"status"`
	ResponseTime time.Duration `json:"response_time"`
	Error        string        `json:"error,omitempty"`
	Timestamp    time.Time     `json:"timestamp"`
}

func NewHealthMonitor() *HealthMonitor {
	return &HealthMonitor{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		checkInterval:   5 * time.Minute,
		circuitBreakers: make(map[string]*CircuitBreaker),
		stopChan:        make(chan struct{}),
		callbacks:       []HealthCallback{},
	}
}

func (hm *HealthMonitor) StartMonitoring(ctx context.Context, providers map[string]*ProviderConfig) {
	hm.mu.Lock()
	hm.providers = providers

	// Initialize circuit breakers
	for name := range providers {
		hm.circuitBreakers[name] = &CircuitBreaker{
			FailureThreshold: 5,
			TimeoutDuration:  10 * time.Minute,
			State:            "closed",
		}
	}
	hm.mu.Unlock()

	// Start monitoring goroutine
	go hm.monitoringLoop(ctx)
}

func (hm *HealthMonitor) monitoringLoop(ctx context.Context) {
	ticker := time.NewTicker(hm.checkInterval)
	defer ticker.Stop()

	// Initial health check
	hm.checkAllProviders(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-hm.stopChan:
			return
		case <-ticker.C:
			hm.checkAllProviders(ctx)
		}
	}
}

func (hm *HealthMonitor) checkAllProviders(ctx context.Context) {
	hm.mu.RLock()
	providers := make(map[string]*ProviderConfig)
	for k, v := range hm.providers {
		providers[k] = v
	}
	hm.mu.RUnlock()

	var wg sync.WaitGroup
	for name, provider := range providers {
		wg.Add(1)
		go func(name string, provider *ProviderConfig) {
			defer wg.Done()
			hm.checkProviderHealth(ctx, name, provider)
		}(name, provider)
	}

	wg.Wait()
}

func (hm *HealthMonitor) checkProviderHealth(ctx context.Context, name string, provider *ProviderConfig) {
	result := hm.performHealthCheck(ctx, provider)

	hm.mu.Lock()
	defer hm.mu.Unlock()

	// Update circuit breaker
	breaker := hm.circuitBreakers[name]
	if result.Status == "healthy" {
		breaker.FailureCount = 0
		if breaker.State == "half-open" {
			breaker.State = "closed"
		}
	} else {
		breaker.FailureCount++
		breaker.LastFailure = time.Now()

		if breaker.FailureCount >= breaker.FailureThreshold && breaker.State == "closed" {
			breaker.State = "open"
		}
	}

	// Check if circuit should move to half-open
	if breaker.State == "open" && time.Since(breaker.LastFailure) > breaker.TimeoutDuration {
		breaker.State = "half-open"
	}

	// Update provider health
	health := HealthStatus{
		Status:       result.Status,
		LastCheck:    result.Timestamp,
		ResponseTime: result.ResponseTime.Milliseconds(),
		ErrorMessage: result.Error,
	}

	// Apply circuit breaker state
	if breaker.State == "open" {
		health.Status = "down"
		health.ErrorMessage = "Circuit breaker open"
	}

	if p, exists := hm.providers[name]; exists {
		p.Health = health
	}

	// Notify callbacks
	for _, callback := range hm.callbacks {
		callback(name, health)
	}
}

func (hm *HealthMonitor) performHealthCheck(ctx context.Context, provider *ProviderConfig) HealthCheckResult {
	start := time.Now()
	result := HealthCheckResult{
		Provider:  provider.Name,
		Timestamp: start,
	}

	// Skip health check for script-based providers
	if strings.HasPrefix(provider.Endpoint, "./scripts/") {
		result.Status = "unknown"
		result.ResponseTime = 0
		return result
	}

	// Create health check URL
	healthURL := hm.buildHealthCheckURL(provider)

	req, err := http.NewRequestWithContext(ctx, "GET", healthURL, nil)
	if err != nil {
		result.Status = "down"
		result.Error = fmt.Sprintf("Failed to create request: %v", err)
		result.ResponseTime = time.Since(start)
		return result
	}

	// Add authentication if required
	hm.addAuthentication(req, provider)

	resp, err := hm.httpClient.Do(req)
	if err != nil {
		result.Status = "down"
		result.Error = fmt.Sprintf("Request failed: %v", err)
		result.ResponseTime = time.Since(start)
		return result
	}
	defer resp.Body.Close()

	result.ResponseTime = time.Since(start)

	// Check response status
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		result.Status = "healthy"
	} else if resp.StatusCode >= 500 {
		result.Status = "down"
		result.Error = fmt.Sprintf("Server error: %d", resp.StatusCode)
	} else {
		result.Status = "degraded"
		result.Error = fmt.Sprintf("HTTP %d", resp.StatusCode)
	}

	return result
}

func (hm *HealthMonitor) buildHealthCheckURL(provider *ProviderConfig) string {
	baseURL := provider.Endpoint

	// Handle different provider types
	switch {
	case strings.Contains(provider.Name, "OpenAI"):
		return baseURL + "/models"
	case strings.Contains(provider.Name, "Anthropic"):
		return baseURL + "/v1/messages" // Will return 400 but indicates service is up
	case strings.Contains(provider.Name, "Pollinations"):
		return baseURL + "/models"
	case strings.Contains(provider.Name, "HuggingFace"):
		return "https://huggingface.co/api/status"
	default:
		// Try common health check endpoints
		if strings.HasSuffix(baseURL, "/") {
			return baseURL + "health"
		}
		return baseURL + "/health"
	}
}

func (hm *HealthMonitor) addAuthentication(req *http.Request, provider *ProviderConfig) {
	if !provider.Authentication.Required || provider.Authentication.Type == "none" {
		return
	}

	authValue := os.Getenv(provider.Authentication.EnvVar)
	if authValue == "" {
		return
	}

	switch provider.Authentication.Type {
	case "bearer_token":
		req.Header.Set("Authorization", "Bearer "+authValue)
	case "api_key":
		req.Header.Set(provider.Authentication.Header, authValue)
	case "cookie":
		req.Header.Set("Cookie", authValue)
	}
}

func (hm *HealthMonitor) GetProviderHealth(providerName string) (HealthStatus, bool) {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	if provider, exists := hm.providers[providerName]; exists {
		return provider.Health, true
	}

	return HealthStatus{}, false
}

func (hm *HealthMonitor) GetCircuitBreakerStatus(providerName string) (*CircuitBreaker, bool) {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	breaker, exists := hm.circuitBreakers[providerName]
	return breaker, exists
}

func (hm *HealthMonitor) IsProviderAvailable(providerName string) bool {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	if breaker, exists := hm.circuitBreakers[providerName]; exists {
		return breaker.State != "open"
	}

	return true
}

func (hm *HealthMonitor) AddHealthCallback(callback HealthCallback) {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	hm.callbacks = append(hm.callbacks, callback)
}

func (hm *HealthMonitor) Stop() {
	close(hm.stopChan)
}

func (hm *HealthMonitor) ForceHealthCheck(ctx context.Context, providerName string) (*HealthCheckResult, error) {
	hm.mu.RLock()
	provider, exists := hm.providers[providerName]
	hm.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("provider %s not found", providerName)
	}

	result := hm.performHealthCheck(ctx, provider)
	return &result, nil
}

func (hm *HealthMonitor) GetHealthSummary() map[string]HealthCheckResult {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	summary := make(map[string]HealthCheckResult)
	for name, provider := range hm.providers {
		summary[name] = HealthCheckResult{
			Provider:     name,
			Status:       provider.Health.Status,
			ResponseTime: time.Duration(provider.Health.ResponseTime) * time.Millisecond,
			Error:        provider.Health.ErrorMessage,
			Timestamp:    provider.Health.LastCheck,
		}
	}

	return summary
}

func (hm *HealthMonitor) ResetCircuitBreaker(providerName string) error {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	if breaker, exists := hm.circuitBreakers[providerName]; exists {
		breaker.FailureCount = 0
		breaker.State = "closed"
		breaker.LastFailure = time.Time{}
		return nil
	}

	return fmt.Errorf("circuit breaker for provider %s not found", providerName)
}