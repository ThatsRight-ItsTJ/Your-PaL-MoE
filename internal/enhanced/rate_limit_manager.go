package enhanced

import (
	"fmt"
	"strconv"
	"sync"
	"time"
)

// RateLimitManager handles provider rate limit tracking and management
type RateLimitManager struct {
	mu                sync.RWMutex
	providerLimits    map[string]*ProviderRateLimits
	metricsStorage    *MetricsStorage
}

// ProviderRateLimits tracks rate limits for a specific provider
type ProviderRateLimits struct {
	ProviderName        string
	Models              map[string]*ModelRateLimit
	LastUpdated         time.Time
}

// ModelRateLimit tracks rate limits for a specific model
type ModelRateLimit struct {
	Model               string
	RequestsPerMinute   int64
	RequestsRemaining   int64
	TokensPerMinute     int64
	TokensRemaining     int64
	ResetTime           time.Time
	LastRateLimitHit    time.Time
	ConsecutiveHits     int
}

// NewRateLimitManager creates a new rate limit manager
func NewRateLimitManager(storage *MetricsStorage) *RateLimitManager {
	return &RateLimitManager{
		providerLimits: make(map[string]*ProviderRateLimits),
		metricsStorage: storage,
	}
}

// UpdateRateLimitStatus updates rate limit information from API response headers
func (r *RateLimitManager) UpdateRateLimitStatus(
	providerName, model string,
	headers map[string]string,
) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	// Parse common rate limit headers
	requestsPerMin := r.parseHeader(headers, []string{"x-ratelimit-limit-requests", "x-ratelimit-requests"})
	requestsRemaining := r.parseHeader(headers, []string{"x-ratelimit-remaining-requests", "x-remaining-requests"})
	tokensPerMin := r.parseHeader(headers, []string{"x-ratelimit-limit-tokens", "x-ratelimit-tokens"})
	tokensRemaining := r.parseHeader(headers, []string{"x-ratelimit-remaining-tokens", "x-remaining-tokens"})
	
	resetTime := r.parseResetTime(headers)
	
	// Get or create provider limits
	providerLimits, exists := r.providerLimits[providerName]
	if !exists {
		providerLimits = &ProviderRateLimits{
			ProviderName: providerName,
			Models:       make(map[string]*ModelRateLimit),
		}
		r.providerLimits[providerName] = providerLimits
	}
	
	// Update model-specific limits
	modelLimit := &ModelRateLimit{
		Model:               model,
		RequestsPerMinute:   requestsPerMin,
		RequestsRemaining:   requestsRemaining,
		TokensPerMinute:     tokensPerMin,
		TokensRemaining:     tokensRemaining,
		ResetTime:           resetTime,
	}
	
	providerLimits.Models[model] = modelLimit
	providerLimits.LastUpdated = time.Now()
	
	// Persist to database if available
	if r.metricsStorage != nil {
		return r.metricsStorage.UpdateRateLimitStatus(
			providerName, model,
			requestsPerMin, requestsRemaining,
			tokensPerMin, tokensRemaining,
			resetTime,
		)
	}
	
	return nil
}

// RecordRateLimitHit records when a provider hits rate limits
func (r *RateLimitManager) RecordRateLimitHit(providerName, model string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	providerLimits, exists := r.providerLimits[providerName]
	if !exists {
		return
	}
	
	modelLimit, exists := providerLimits.Models[model]
	if !exists {
		return
	}
	
	modelLimit.LastRateLimitHit = time.Now()
	modelLimit.ConsecutiveHits++
	
	// Zero out remaining limits
	modelLimit.RequestsRemaining = 0
	modelLimit.TokensRemaining = 0
}

// CanHandleRequest checks if a provider can handle a request without hitting rate limits
func (r *RateLimitManager) CanHandleRequest(
	providerName, model string,
	estimatedTokens int64,
) (bool, string) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	providerLimits, exists := r.providerLimits[providerName]
	if !exists {
		return true, "no_rate_limit_data" // Assume OK if no data
	}
	
	modelLimit, exists := providerLimits.Models[model]
	if !exists {
		return true, "no_model_rate_limit_data"
	}
	
	now := time.Now()
	
	// Check if limits have reset
	if now.After(modelLimit.ResetTime) {
		// Reset the limits (assume full capacity restored)
		modelLimit.RequestsRemaining = modelLimit.RequestsPerMinute
		modelLimit.TokensRemaining = modelLimit.TokensPerMinute
		modelLimit.ConsecutiveHits = 0
	}
	
	// Check if we have enough request capacity
	if modelLimit.RequestsRemaining < 1 {
		return false, "request_limit_exceeded"
	}
	
	// Check if we have enough token capacity (with buffer)
	tokensNeeded := estimatedTokens * 2 // Buffer for input + output tokens
	if modelLimit.TokensRemaining < tokensNeeded {
		return false, "token_limit_exceeded"
	}
	
	// Check for recent consecutive rate limit hits (backoff)
	if modelLimit.ConsecutiveHits > 3 {
		timeSinceLastHit := now.Sub(modelLimit.LastRateLimitHit)
		backoffDuration := time.Duration(modelLimit.ConsecutiveHits) * time.Minute
		if timeSinceLastHit < backoffDuration {
			return false, "consecutive_rate_limits_backoff"
		}
	}
	
	return true, "available"
}

// GetRateLimitStatus returns current rate limit status for a provider/model
func (r *RateLimitManager) GetRateLimitStatus(providerName, model string) *RateLimitStatus {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	providerLimits, exists := r.providerLimits[providerName]
	if !exists {
		return nil
	}
	
	modelLimit, exists := providerLimits.Models[model]
	if !exists {
		return nil
	}
	
	return &RateLimitStatus{
		RequestsPerMinute: modelLimit.RequestsPerMinute,
		RequestsRemaining: modelLimit.RequestsRemaining,
		TokensPerMinute:   modelLimit.TokensPerMinute,
		TokensRemaining:   modelLimit.TokensRemaining,
		ResetTime:         modelLimit.ResetTime,
		LastRateLimitHit:  modelLimit.LastRateLimitHit,
	}
}

// parseHeader extracts rate limit values from response headers
func (r *RateLimitManager) parseHeader(headers map[string]string, keys []string) int64 {
	for _, key := range keys {
		if value, exists := headers[key]; exists {
			if parsed := parseInt64(value); parsed > 0 {
				return parsed
			}
		}
	}
	return 0
}

// parseResetTime extracts reset time from headers
func (r *RateLimitManager) parseResetTime(headers map[string]string) time.Time {
	resetHeaders := []string{"x-ratelimit-reset", "x-reset-time", "retry-after"}
	
	for _, header := range resetHeaders {
		if value, exists := headers[header]; exists {
			if timestamp := parseInt64(value); timestamp > 0 {
				return time.Unix(timestamp, 0)
			}
		}
	}
	
	// Default to 1 minute from now if no reset time found
	return time.Now().Add(time.Minute)
}

// Helper function to parse int64 from string
func parseInt64(s string) int64 {
	if val, err := strconv.ParseInt(s, 10, 64); err == nil {
		return val
	}
	return 0
}