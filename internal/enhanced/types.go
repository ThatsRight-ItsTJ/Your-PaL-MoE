package enhanced

import (
	"time"
	"github.com/Your-PaL-MoE/internal/types"
)

// Re-export types from internal/types for convenience
type RequestInput = types.RequestInput
type RequestResult = types.RequestResult
type Provider = types.Provider
type AuthConfig = types.AuthConfig

// ComplexityLevel represents the complexity level of a task
type ComplexityLevel int

const (
	Low ComplexityLevel = iota
	Medium
	High
	Critical
)

// TaskComplexity represents the complexity analysis of a task
type TaskComplexity struct {
	Overall          ComplexityLevel            `json:"overall"`
	Reasoning        ComplexityLevel            `json:"reasoning"`
	Knowledge        ComplexityLevel            `json:"knowledge"`
	Creativity       ComplexityLevel            `json:"creativity"`
	Computation      ComplexityLevel            `json:"computation"`
	Factors          map[string]interface{}     `json:"factors"`
	EstimatedTokens  int                        `json:"estimated_tokens"`
	RequiredCapabilities []string               `json:"required_capabilities"`
}

// OptimizedPrompt represents an optimized prompt with SPO analysis
type OptimizedPrompt struct {
	Original        string                 `json:"original"`
	Optimized       string                 `json:"optimized"`
	Subject         string                 `json:"subject"`
	Predicate       string                 `json:"predicate"`
	Object          string                 `json:"object"`
	Improvements    []string               `json:"improvements"`
	TokenReduction  int                    `json:"token_reduction"`
	CostSavings     float64                `json:"cost_savings"`
	Confidence      float64                `json:"confidence"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// ProviderAssignment represents the assignment of a task to a provider
type ProviderAssignment struct {
	ProviderID      string                 `json:"provider_id"`
	ProviderName    string                 `json:"provider_name"`
	Model           string                 `json:"model"`
	Tier            string                 `json:"tier"`
	Confidence      float64                `json:"confidence"`
	EstimatedCost   float64                `json:"estimated_cost"`
	EstimatedTime   int64                  `json:"estimated_time"`
	Reasoning       string                 `json:"reasoning"`
	Alternatives    []AlternativeProvider  `json:"alternatives"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// AlternativeProvider represents an alternative provider option
type AlternativeProvider struct {
	ProviderID      string  `json:"provider_id"`
	ProviderName    string  `json:"provider_name"`
	Model           string  `json:"model"`
	Confidence      float64 `json:"confidence"`
	EstimatedCost   float64 `json:"estimated_cost"`
	EstimatedTime   int64   `json:"estimated_time"`
	Reasoning       string  `json:"reasoning"`
}

// TaskRequirements represents requirements extracted from a task
type TaskRequirements struct {
	Domain           string            `json:"domain"`
	OutputFormat     string            `json:"output_format"`
	Language         string            `json:"language"`
	Capabilities     []string          `json:"capabilities"`
	Constraints      map[string]string `json:"constraints"`
	QualityLevel     string            `json:"quality_level"`
	MaxTokens        int               `json:"max_tokens"`
	MaxCost          float64           `json:"max_cost"`
	MaxTime          time.Duration     `json:"max_time"`
}

// ProviderMetrics represents performance metrics for a provider
type ProviderMetrics struct {
	TotalRequests     int64     `json:"total_requests"`
	SuccessfulRequests int64    `json:"successful_requests"`
	FailedRequests    int64     `json:"failed_requests"`
	SuccessRate       float64   `json:"success_rate"`
	AverageLatency    float64   `json:"average_latency"`
	AverageCost       float64   `json:"average_cost"`
	QualityScore      float64   `json:"quality_score"`
	ReliabilityScore  float64   `json:"reliability_score"`
	LastUsed          time.Time `json:"last_used"`
	LastUpdated       time.Time `json:"last_updated"`
}

// Enhanced Provider with additional fields needed by the system
type EnhancedProvider struct {
	*Provider
	ID           string           `json:"id"`
	Models       []string         `json:"models"`
	Capabilities []string         `json:"capabilities"`
	Pricing      PricingInfo      `json:"pricing"`
	Limits       ProviderLimits   `json:"limits"`
	Health       HealthStatus     `json:"health"`
	Metrics      ProviderMetrics  `json:"metrics"`
}

// PricingInfo represents pricing information for a provider
type PricingInfo struct {
	InputTokenPrice  float64 `json:"input_token_price"`
	OutputTokenPrice float64 `json:"output_token_price"`
	Currency         string  `json:"currency"`
	BillingUnit      string  `json:"billing_unit"`
}

// ProviderLimits represents limits for a provider
type ProviderLimits struct {
	MaxTokensPerRequest   int           `json:"max_tokens_per_request"`
	MaxRequestsPerMinute  int           `json:"max_requests_per_minute"`
	MaxRequestsPerDay     int           `json:"max_requests_per_day"`
	MaxConcurrentRequests int           `json:"max_concurrent_requests"`
	Timeout               time.Duration `json:"timeout"`
}

// HealthStatus represents the health status of a provider
type HealthStatus struct {
	Status      string    `json:"status"`
	LastCheck   time.Time `json:"last_check"`
	ResponseTime float64  `json:"response_time"`
	ErrorRate   float64   `json:"error_rate"`
	Available   bool      `json:"available"`
}