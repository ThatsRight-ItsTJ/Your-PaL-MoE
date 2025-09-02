package enhanced

import (
	"sync"
	"time"
)

// ProviderHealthMetrics represents health metrics for a provider
type ProviderHealthMetrics struct {
	SuccessRate      float64   `json:"success_rate"`
	AverageLatency   float64   `json:"average_latency"`
	ErrorRate        float64   `json:"error_rate"`
	TotalRequests    int64     `json:"total_requests"`
	SuccessfulRequests int64   `json:"successful_requests"`
	FailedRequests   int64     `json:"failed_requests"`
	LastUpdated      time.Time `json:"last_updated"`
	Status           string    `json:"status"`
}

// UsageRecord represents a usage record for rate limiting
type UsageRecord struct {
	Timestamp time.Time `json:"timestamp"`
	Count     int64     `json:"count"`
	Cost      float64   `json:"cost"`
}

// RateLimitStatus represents the rate limit status for a provider
type RateLimitStatus struct {
	RequestsPerMinute int64     `json:"requests_per_minute"`
	CurrentUsage      int64     `json:"current_usage"`
	ResetTime         time.Time `json:"reset_time"`
	IsLimited         bool      `json:"is_limited"`
}

// ProviderHealthMonitor monitors provider health and performance
type ProviderHealthMonitor struct {
	providers map[string]*ProviderHealthMetrics
	mutex     sync.RWMutex
}

// NewProviderHealthMonitor creates a new provider health monitor
func NewProviderHealthMonitor() *ProviderHealthMonitor {
	return &ProviderHealthMonitor{
		providers: make(map[string]*ProviderHealthMetrics),
	}
}

// UpdateMetrics updates health metrics for a provider
func (phm *ProviderHealthMonitor) UpdateMetrics(providerName string, success bool, latency time.Duration) {
	phm.mutex.Lock()
	defer phm.mutex.Unlock()

	metrics, exists := phm.providers[providerName]
	if !exists {
		metrics = &ProviderHealthMetrics{
			LastUpdated: time.Now(),
			Status:      "active",
		}
		phm.providers[providerName] = metrics
	}

	// Update counters
	metrics.TotalRequests++
	if success {
		metrics.SuccessfulRequests++
	} else {
		metrics.FailedRequests++
	}

	// Calculate rates
	if metrics.TotalRequests > 0 {
		metrics.SuccessRate = float64(metrics.SuccessfulRequests) / float64(metrics.TotalRequests)
		metrics.ErrorRate = float64(metrics.FailedRequests) / float64(metrics.TotalRequests)
	}

	// Update average latency (simple moving average)
	if metrics.AverageLatency == 0 {
		metrics.AverageLatency = float64(latency.Milliseconds())
	} else {
		metrics.AverageLatency = (metrics.AverageLatency + float64(latency.Milliseconds())) / 2
	}

	metrics.LastUpdated = time.Now()

	// Update status based on metrics
	if metrics.ErrorRate > 0.5 {
		metrics.Status = "degraded"
	} else if metrics.ErrorRate > 0.2 {
		metrics.Status = "warning"
	} else {
		metrics.Status = "healthy"
	}
}

// GetMetrics returns health metrics for a provider
func (phm *ProviderHealthMonitor) GetMetrics(providerName string) *ProviderHealthMetrics {
	phm.mutex.RLock()
	defer phm.mutex.RUnlock()

	if metrics, exists := phm.providers[providerName]; exists {
		// Return a copy to avoid race conditions
		return &ProviderHealthMetrics{
			SuccessRate:        metrics.SuccessRate,
			AverageLatency:     metrics.AverageLatency,
			ErrorRate:          metrics.ErrorRate,
			TotalRequests:      metrics.TotalRequests,
			SuccessfulRequests: metrics.SuccessfulRequests,
			FailedRequests:     metrics.FailedRequests,
			LastUpdated:        metrics.LastUpdated,
			Status:             metrics.Status,
		}
	}

	return nil
}

// GetAllMetrics returns health metrics for all providers
func (phm *ProviderHealthMonitor) GetAllMetrics() map[string]*ProviderHealthMetrics {
	phm.mutex.RLock()
	defer phm.mutex.RUnlock()

	result := make(map[string]*ProviderHealthMetrics)
	for name, metrics := range phm.providers {
		result[name] = &ProviderHealthMetrics{
			SuccessRate:        metrics.SuccessRate,
			AverageLatency:     metrics.AverageLatency,
			ErrorRate:          metrics.ErrorRate,
			TotalRequests:      metrics.TotalRequests,
			SuccessfulRequests: metrics.SuccessfulRequests,
			FailedRequests:     metrics.FailedRequests,
			LastUpdated:        metrics.LastUpdated,
			Status:             metrics.Status,
		}
	}

	return result
}

// IsHealthy checks if a provider is healthy
func (phm *ProviderHealthMonitor) IsHealthy(providerName string) bool {
	metrics := phm.GetMetrics(providerName)
	if metrics == nil {
		return true // Assume healthy if no data
	}

	return metrics.Status == "healthy" || metrics.Status == "active"
}

// ResetMetrics resets metrics for a provider
func (phm *ProviderHealthMonitor) ResetMetrics(providerName string) {
	phm.mutex.Lock()
	defer phm.mutex.Unlock()

	delete(phm.providers, providerName)
}