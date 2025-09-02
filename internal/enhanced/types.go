package enhanced

import (
	"time"

	"github.com/ThatsRight-ItsTJ/Your-PaL-MoE/internal/components"
)

// ProviderTier represents the tier/quality level of a provider
type ProviderTier string

const (
	TierFree     ProviderTier = "free"
	TierBasic    ProviderTier = "basic"
	TierPremium  ProviderTier = "premium"
	TierEnterprise ProviderTier = "enterprise"
)

// Provider represents an AI provider configuration
type Provider struct {
	Name         string       `json:"name"`
	BaseURL      string       `json:"base_url"`
	Models       []string     `json:"models"`
	Tier         ProviderTier `json:"tier"`
	MaxTokens    int          `json:"max_tokens"`
	CostPerToken float64      `json:"cost_per_token"`
	Capabilities []string     `json:"capabilities"`
	HealthMetrics *ProviderHealthMetrics `json:"health_metrics,omitempty"`
}

// RequestInput represents input for processing a request
type RequestInput struct {
	Content           string            `json:"content"`
	PreferredProvider string            `json:"preferred_provider,omitempty"`
	MaxTokens         int               `json:"max_tokens,omitempty"`
	Temperature       float64           `json:"temperature,omitempty"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

// ProcessResponse represents the response from processing a request
type ProcessResponse struct {
	Content        string                     `json:"content"`
	Provider       *Provider                  `json:"provider"`
	Model          string                     `json:"model"`
	Complexity     components.TaskComplexity  `json:"complexity"`
	ProcessingTime time.Duration              `json:"processing_time"`
	TokensUsed     int64                      `json:"tokens_used"`
	Cost           float64                    `json:"cost"`
	Metadata       map[string]interface{}     `json:"metadata"`
}

// ProviderAssignment represents the result of provider selection
type ProviderAssignment struct {
	Provider        *Provider `json:"provider"`
	Model           string    `json:"model"`
	Confidence      float64   `json:"confidence"`
	EstimatedCost   float64   `json:"estimated_cost"`
	EstimatedTokens int64     `json:"estimated_tokens"`
	Reasoning       string    `json:"reasoning"`
	Alternatives    []*Provider `json:"alternatives,omitempty"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// ProviderScore represents a scored provider for selection
type ProviderScore struct {
	Provider   *Provider `json:"provider"`
	Score      float64   `json:"score"`
	Confidence float64   `json:"confidence"`
	Reasoning  string    `json:"reasoning"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// OptimizedPrompt represents an optimized prompt result
type OptimizedPrompt struct {
	OriginalPrompt    string                      `json:"original_prompt"`
	OptimizedPrompt   string                      `json:"optimized_prompt"`
	Complexity        components.ComplexityLevel  `json:"complexity"`
	OptimizationRules []string                    `json:"optimization_rules"`
	Metadata          map[string]interface{}      `json:"metadata"`
	ProcessingTime    time.Duration               `json:"processing_time"`
}

// SystemMetrics represents system-wide metrics
type SystemMetrics struct {
	TotalRequests     int64                              `json:"total_requests"`
	SuccessfulRequests int64                             `json:"successful_requests"`
	FailedRequests    int64                              `json:"failed_requests"`
	AverageLatency    time.Duration                      `json:"average_latency"`
	ComplexityDistribution map[components.ComplexityLevel]int64 `json:"complexity_distribution"`
	ProviderUsage     map[string]int64                   `json:"provider_usage"`
	TotalCost         float64                            `json:"total_cost"`
	TotalTokens       int64                              `json:"total_tokens"`
	StartTime         time.Time                          `json:"start_time"`
	LastUpdated       time.Time                          `json:"last_updated"`
}

// NewSystemMetrics creates a new system metrics instance
func NewSystemMetrics() *SystemMetrics {
	return &SystemMetrics{
		ComplexityDistribution: make(map[components.ComplexityLevel]int64),
		ProviderUsage:         make(map[string]int64),
		StartTime:             time.Now(),
		LastUpdated:           time.Now(),
	}
}

// IncrementTotalRequests increments the total request counter
func (sm *SystemMetrics) IncrementTotalRequests() {
	sm.TotalRequests++
	sm.LastUpdated = time.Now()
}

// IncrementSuccessfulRequests increments the successful request counter
func (sm *SystemMetrics) IncrementSuccessfulRequests() {
	sm.SuccessfulRequests++
	sm.LastUpdated = time.Now()
}

// IncrementFailedRequests increments the failed request counter
func (sm *SystemMetrics) IncrementFailedRequests() {
	sm.FailedRequests++
	sm.LastUpdated = time.Now()
}

// RecordComplexity records complexity distribution
func (sm *SystemMetrics) RecordComplexity(complexity components.ComplexityLevel) {
	sm.ComplexityDistribution[complexity]++
	sm.LastUpdated = time.Now()
}

// RecordProviderUsage records provider usage
func (sm *SystemMetrics) RecordProviderUsage(providerName string) {
	sm.ProviderUsage[providerName]++
	sm.LastUpdated = time.Now()
}

// UpdateLatency updates average latency
func (sm *SystemMetrics) UpdateLatency(latency time.Duration) {
	// Simple moving average - could be improved with more sophisticated calculation
	if sm.TotalRequests == 1 {
		sm.AverageLatency = latency
	} else {
		sm.AverageLatency = time.Duration((int64(sm.AverageLatency)*int64(sm.TotalRequests-1) + int64(latency)) / int64(sm.TotalRequests))
	}
	sm.LastUpdated = time.Now()
}

// AddCost adds to total cost
func (sm *SystemMetrics) AddCost(cost float64) {
	sm.TotalCost += cost
	sm.LastUpdated = time.Now()
}

// AddTokens adds to total tokens
func (sm *SystemMetrics) AddTokens(tokens int64) {
	sm.TotalTokens += tokens
	sm.LastUpdated = time.Now()
}

// ProviderHealthMetrics represents health metrics for a provider
type ProviderHealthMetrics struct {
	ProviderName      string        `json:"provider_name"`
	TotalRequests     int64         `json:"total_requests"`
	SuccessfulRequests int64        `json:"successful_requests"`
	FailedRequests    int64         `json:"failed_requests"`
	AverageLatency    time.Duration `json:"average_latency"`
	LastRequestTime   time.Time     `json:"last_request_time"`
	HealthScore       float64       `json:"health_score"`
	IsHealthy         bool          `json:"is_healthy"`
	ErrorRate         float64       `json:"error_rate"`
	LastError         string        `json:"last_error,omitempty"`
	LastErrorTime     time.Time     `json:"last_error_time,omitempty"`
}

// ProviderHealthMonitor monitors provider health
type ProviderHealthMonitor struct {
	metrics map[string]*ProviderHealthMetrics
}

// NewProviderHealthMonitor creates a new provider health monitor
func NewProviderHealthMonitor() *ProviderHealthMonitor {
	return &ProviderHealthMonitor{
		metrics: make(map[string]*ProviderHealthMetrics),
	}
}

// UpdateMetrics updates health metrics for a provider
func (phm *ProviderHealthMonitor) UpdateMetrics(providerName string, success bool, latency time.Duration) {
	if phm.metrics[providerName] == nil {
		phm.metrics[providerName] = &ProviderHealthMetrics{
			ProviderName: providerName,
		}
	}

	metrics := phm.metrics[providerName]
	metrics.TotalRequests++
	metrics.LastRequestTime = time.Now()

	if success {
		metrics.SuccessfulRequests++
	} else {
		metrics.FailedRequests++
	}

	// Update average latency
	if metrics.TotalRequests == 1 {
		metrics.AverageLatency = latency
	} else {
		metrics.AverageLatency = time.Duration((int64(metrics.AverageLatency)*int64(metrics.TotalRequests-1) + int64(latency)) / int64(metrics.TotalRequests))
	}

	// Calculate error rate
	metrics.ErrorRate = float64(metrics.FailedRequests) / float64(metrics.TotalRequests)

	// Calculate health score (simple implementation)
	metrics.HealthScore = 1.0 - metrics.ErrorRate
	if metrics.AverageLatency > 5*time.Second {
		metrics.HealthScore *= 0.8 // Penalize high latency
	}

	// Determine if healthy
	metrics.IsHealthy = metrics.HealthScore > 0.7 && metrics.ErrorRate < 0.3
}

// GetMetrics returns health metrics for a provider
func (phm *ProviderHealthMonitor) GetMetrics(providerName string) *ProviderHealthMetrics {
	return phm.metrics[providerName]
}

// GetAllMetrics returns all provider health metrics
func (phm *ProviderHealthMonitor) GetAllMetrics() map[string]*ProviderHealthMetrics {
	return phm.metrics
}

// IsHealthy checks if a provider is healthy
func (phm *ProviderHealthMonitor) IsHealthy(providerName string) bool {
	if metrics := phm.metrics[providerName]; metrics != nil {
		return metrics.IsHealthy
	}
	return true // Assume healthy if no metrics available
}

// ResetMetrics resets metrics for a provider
func (phm *ProviderHealthMonitor) ResetMetrics(providerName string) {
	delete(phm.metrics, providerName)
}

// EnhancedProviderSelector represents an enhanced provider selector
type EnhancedProviderSelector struct {
	providers     []*Provider
	healthMonitor *ProviderHealthMonitor
}

// NewEnhancedProviderSelector creates a new enhanced provider selector
func NewEnhancedProviderSelector(providers []*Provider) *EnhancedProviderSelector {
	return &EnhancedProviderSelector{
		providers:     providers,
		healthMonitor: NewProviderHealthMonitor(),
	}
}

// SelectProviderWithCapabilities selects a provider based on complexity and required capabilities
func (eps *EnhancedProviderSelector) SelectProviderWithCapabilities(ctx context.Context, complexity components.TaskComplexity, requiredCapabilities []string) (*ProviderAssignment, error) {
	if len(eps.providers) == 0 {
		return nil, fmt.Errorf("no providers available")
	}

	// Score providers based on capabilities and health
	var scores []ProviderScore
	for _, provider := range eps.providers {
		score := eps.scoreProvider(provider, complexity, requiredCapabilities)
		scores = append(scores, score)
	}

	// Sort by score (highest first)
	for i := 0; i < len(scores)-1; i++ {
		for j := i + 1; j < len(scores); j++ {
			if scores[j].Score > scores[i].Score {
				scores[i], scores[j] = scores[j], scores[i]
			}
		}
	}

	if len(scores) == 0 {
		return nil, fmt.Errorf("no suitable providers found")
	}

	bestScore := scores[0]
	
	// Select first available model
	model := "default"
	if len(bestScore.Provider.Models) > 0 {
		model = bestScore.Provider.Models[0]
	}

	return &ProviderAssignment{
		Provider:        bestScore.Provider,
		Model:           model,
		Confidence:      bestScore.Confidence,
		EstimatedCost:   float64(complexity.TokenEstimate) * bestScore.Provider.CostPerToken,
		EstimatedTokens: complexity.TokenEstimate,
		Reasoning:       bestScore.Reasoning,
		Alternatives:    []*Provider{}, // Could populate with other high-scoring providers
		Metadata:        make(map[string]interface{}),
	}, nil
}

// scoreProvider scores a provider based on various factors
func (eps *EnhancedProviderSelector) scoreProvider(provider *Provider, complexity components.TaskComplexity, requiredCapabilities []string) ProviderScore {
	score := 1.0
	reasoning := []string{}

	// Check capability match
	capabilityScore := eps.calculateCapabilityScore(provider.Capabilities, requiredCapabilities)
	score *= capabilityScore
	if capabilityScore > 0.8 {
		reasoning = append(reasoning, "excellent capability match")
	} else if capabilityScore > 0.5 {
		reasoning = append(reasoning, "good capability match")
	} else {
		reasoning = append(reasoning, "limited capability match")
	}

	// Factor in health metrics
	if metrics := eps.healthMonitor.GetMetrics(provider.Name); metrics != nil {
		score *= metrics.HealthScore
		if metrics.IsHealthy {
			reasoning = append(reasoning, "healthy provider")
		} else {
			reasoning = append(reasoning, "provider health concerns")
		}
	}

	// Factor in cost efficiency
	costScore := 1.0 / (1.0 + provider.CostPerToken*100) // Prefer lower cost
	score *= costScore

	// Factor in complexity appropriateness
	complexityScore := eps.calculateComplexityScore(provider, complexity)
	score *= complexityScore

	confidence := score
	if confidence > 1.0 {
		confidence = 1.0
	}

	return ProviderScore{
		Provider:   provider,
		Score:      score,
		Confidence: confidence,
		Reasoning:  strings.Join(reasoning, ", "),
		Metadata:   make(map[string]interface{}),
	}
}

// calculateCapabilityScore calculates how well provider capabilities match requirements
func (eps *EnhancedProviderSelector) calculateCapabilityScore(providerCaps, requiredCaps []string) float64 {
	if len(requiredCaps) == 0 {
		return 1.0 // No specific requirements
	}

	matches := 0
	for _, required := range requiredCaps {
		for _, provided := range providerCaps {
			if provided == required {
				matches++
				break
			}
		}
	}

	return float64(matches) / float64(len(requiredCaps))
}

// calculateComplexityScore calculates how well provider handles complexity
func (eps *EnhancedProviderSelector) calculateComplexityScore(provider *Provider, complexity components.TaskComplexity) float64 {
	// Simple implementation - could be more sophisticated
	switch provider.Tier {
	case TierEnterprise:
		return 1.0 // Can handle any complexity
	case TierPremium:
		if complexity.Overall <= components.High {
			return 1.0
		}
		return 0.8
	case TierBasic:
		if complexity.Overall <= components.Medium {
			return 1.0
		}
		return 0.6
	case TierFree:
		if complexity.Overall <= components.Low {
			return 1.0
		}
		return 0.4
	default:
		return 0.5
	}
}

// GetProviderStats returns statistics about providers
func (eps *EnhancedProviderSelector) GetProviderStats() map[string]interface{} {
	stats := make(map[string]interface{})
	
	stats["total_providers"] = len(eps.providers)
	
	tierCounts := make(map[ProviderTier]int)
	for _, provider := range eps.providers {
		tierCounts[provider.Tier]++
	}
	stats["tier_distribution"] = tierCounts
	
	healthyCount := 0
	for _, provider := range eps.providers {
		if eps.healthMonitor.IsHealthy(provider.Name) {
			healthyCount++
		}
	}
	stats["healthy_providers"] = healthyCount
	
	return stats
}

// EnhancedSystem represents the main enhanced system
type EnhancedSystem struct {
	selector      *EnhancedProviderSelector
	reasoner      *components.TaskReasoner
	optimizer     *components.SPOOptimizer
	healthMonitor *ProviderHealthMonitor
	providers     []*Provider
	metrics       *SystemMetrics
}

// RateLimitStatus represents rate limiting status
type RateLimitStatus struct {
	IsLimited         bool      `json:"is_limited"`
	RequestsPerMinute int       `json:"requests_per_minute"`
	TokensPerMinute   int       `json:"tokens_per_minute"`
	ResetTime         time.Time `json:"reset_time"`
}