package enhanced

import (
	"math"
	"time"
)

// HealthScoreCalculator implements cost-focused health scoring
type HealthScoreCalculator struct {
	WindowSize         time.Duration
	MaxUsageRatio      float64
	ReliabilityWeight  float64
	CostWeight         float64
	RateLimitWeight    float64
}

// UsageRecord represents a single usage entry with cost tracking
type UsageRecord struct {
	Timestamp     time.Time `json:"timestamp"`
	Count         int64     `json:"count"`
	Failures      int64     `json:"failures"`
	Latency       float64   `json:"latency_ms"`
	TokensUsed    int64     `json:"tokens_used"`
	Cost          float64   `json:"cost"`
	RateLimited   bool      `json:"rate_limited"`
}

// ProviderHealthMetrics extends existing ProviderMetrics with cost tracking
type ProviderHealthMetrics struct {
	UsageHistory         map[string][]UsageRecord `json:"usage_history"` // model -> records
	LastHealthCheck      time.Time                `json:"last_health_check"`
	HealthScore          float64                  `json:"health_score"`
	CostEfficiencyScore  float64                  `json:"cost_efficiency_score"`
	RateLimitStatus      *RateLimitStatus         `json:"rate_limit_status"`
}

// RateLimitStatus tracks current rate limit state
type RateLimitStatus struct {
	RequestsPerMinute    int64     `json:"requests_per_minute"`
	RequestsRemaining    int64     `json:"requests_remaining"`
	TokensPerMinute      int64     `json:"tokens_per_minute"`
	TokensRemaining      int64     `json:"tokens_remaining"`
	ResetTime            time.Time `json:"reset_time"`
	LastRateLimitHit     time.Time `json:"last_rate_limit_hit"`
}

// NewHealthScoreCalculator creates a cost-focused calculator
func NewHealthScoreCalculator() *HealthScoreCalculator {
	return &HealthScoreCalculator{
		WindowSize:        24 * time.Hour,
		MaxUsageRatio:     1.0,
		ReliabilityWeight: 0.3,  // Reduced to prioritize cost
		CostWeight:        0.5,  // Increased for cost optimization
		RateLimitWeight:   0.2,  // Added for rate limit awareness
	}
}

// CalculateHealthScore implements cost-aware health scoring
func (h *HealthScoreCalculator) CalculateHealthScore(
	provider *Provider,
	model string,
	totalRequests int64,
) float64 {
	currentTime := time.Now()
	
	// Get recent usage for the specific model
	if provider.HealthMetrics == nil {
		return 0.8 // Default score for providers without health metrics
	}
	
	modelHistory, exists := provider.HealthMetrics.UsageHistory[model]
	if !exists {
		return 0.8 // Default score for new providers
	}
	
	// Filter to recent usage within window
	var recentUsage []UsageRecord
	cutoffTime := currentTime.Add(-h.WindowSize)
	
	for _, record := range modelHistory {
		if record.Timestamp.After(cutoffTime) {
			recentUsage = append(recentUsage, record)
		}
	}
	
	if len(recentUsage) == 0 {
		return 0.8 // Default score when no recent usage
	}
	
	// Calculate metrics with cost focus
	var providerRequests, providerFailures, totalTokens int64
	var totalCost, weightedLatency float64
	rateLimitHits := 0
	
	for _, record := range recentUsage {
		providerRequests += record.Count
		providerFailures += record.Failures
		totalTokens += record.TokensUsed
		totalCost += record.Cost
		weightedLatency += record.Latency * float64(record.Count)
		if record.RateLimited {
			rateLimitHits++
		}
	}
	
	// Reliability score (reduced weight)
	reliability := 1.0
	if providerRequests > 0 {
		reliability = float64(providerRequests-providerFailures) / float64(providerRequests)
	}
	
	// Cost efficiency score (primary factor)
	costPerToken := totalCost / math.Max(1, float64(totalTokens))
	// Normalize against a baseline cost (e.g., GPT-4 pricing)
	baselineCostPerToken := 0.00003 // $0.03 per 1K tokens
	costEfficiency := math.Max(0, 1-(costPerToken/baselineCostPerToken))
	
	// Rate limit availability score
	rateLimitScore := 1.0
	if len(recentUsage) > 0 {
		rateLimitPenalty := float64(rateLimitHits) / float64(len(recentUsage))
		rateLimitScore = 1.0 - rateLimitPenalty
	}
	
	// Check current rate limit status
	if provider.HealthMetrics.RateLimitStatus != nil {
		status := provider.HealthMetrics.RateLimitStatus
		if status.RequestsRemaining < status.RequestsPerMinute/10 { // Less than 10% remaining
			rateLimitScore *= 0.5 // Significant penalty
		}
	}
	
	// Cost-focused weighted final score
	finalScore := reliability*h.ReliabilityWeight +
		costEfficiency*h.CostWeight +
		rateLimitScore*h.RateLimitWeight
	
	return finalScore
}