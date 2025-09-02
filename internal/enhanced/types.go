package enhanced

import (
	"context"
	"time"
)

// ComplexityLevel represents the complexity level of a task
type ComplexityLevel int

const (
	Low ComplexityLevel = iota
	Medium
	High
	VeryHigh
)

// String returns the string representation of ComplexityLevel
func (c ComplexityLevel) String() string {
	switch c {
	case Low:
		return "low"
	case Medium:
		return "medium"
	case High:
		return "high"
	case VeryHigh:
		return "very_high"
	default:
		return "unknown"
	}
}

// TaskComplexity represents the complexity analysis of a task
type TaskComplexity struct {
	Overall          ComplexityLevel            `json:"overall"`
	Reasoning        ComplexityLevel            `json:"reasoning"`
	Mathematical     ComplexityLevel            `json:"mathematical"`
	Creative         ComplexityLevel            `json:"creative"`
	Factual          ComplexityLevel            `json:"factual"`
	TokenEstimate    int64                      `json:"token_estimate"`
	RequiredCapabilities []string               `json:"required_capabilities"`
	Metadata         map[string]interface{}     `json:"metadata"`
}

// ProviderTier represents the tier/quality level of a provider
type ProviderTier int

const (
	OfficialTier ProviderTier = iota
	CommunityTier
	UnofficialTier
)

// String returns the string representation of ProviderTier
func (p ProviderTier) String() string {
	switch p {
	case OfficialTier:
		return "official"
	case CommunityTier:
		return "community"
	case UnofficialTier:
		return "unofficial"
	default:
		return "unknown"
	}
}

// Provider represents an AI provider with enhanced metadata
type Provider struct {
	Name             string                     `json:"name"`
	BaseURL          string                     `json:"base_url"`
	APIKey           string                     `json:"api_key"`
	Models           []string                   `json:"models"`
	Tier             ProviderTier               `json:"tier"`
	MaxTokens        int64                      `json:"max_tokens"`
	CostPerToken     float64                    `json:"cost_per_token"`
	Capabilities     []string                   `json:"capabilities"`
	RateLimits       map[string]int64           `json:"rate_limits"`
	HealthMetrics    *ProviderHealthMetrics     `json:"health_metrics,omitempty"`
	Metadata         map[string]interface{}     `json:"metadata"`
	LastUpdated      time.Time                  `json:"last_updated"`
}

// ProviderAssignment represents a provider assignment for a task
type ProviderAssignment struct {
	Provider         *Provider                  `json:"provider"`
	Model            string                     `json:"model"`
	Confidence       float64                    `json:"confidence"`
	EstimatedCost    float64                    `json:"estimated_cost"`
	EstimatedTokens  int64                      `json:"estimated_tokens"`
	Reasoning        string                     `json:"reasoning"`
	Alternatives     []*Provider                `json:"alternatives,omitempty"`
	Metadata         map[string]interface{}     `json:"metadata"`
}

// TaskType represents the type of task being processed
type TaskType string

const (
	TaskTypeReasoning    TaskType = "reasoning"
	TaskTypeMathematical TaskType = "mathematical"
	TaskTypeCreative     TaskType = "creative"
	TaskTypeFactual      TaskType = "factual"
	TaskTypeGeneral      TaskType = "general"
)

// ProcessResponse represents the response from processing a task
type ProcessResponse struct {
	Content          string                     `json:"content"`
	TokensUsed       int64                      `json:"tokens_used"`
	Cost             float64                    `json:"cost"`
	Latency          time.Duration              `json:"latency"`
	Provider         string                     `json:"provider"`
	Model            string                     `json:"model"`
	Success          bool                       `json:"success"`
	Error            error                      `json:"error,omitempty"`
	Metadata         map[string]interface{}     `json:"metadata"`
}

// TaskReasoner analyzes task complexity
type TaskReasoner struct {
	complexityWeights map[string]float64
	tokenEstimator    *TokenEstimator
}

// ProviderSelector selects optimal providers
type ProviderSelector struct {
	providers         []*Provider
	healthCalculator  *HealthScoreCalculator
	costOptimizer     *CostBasedSelector
}

// SPOOptimizer optimizes prompts
type SPOOptimizer struct {
	optimizationRules map[ComplexityLevel][]string
	templateCache     map[string]string
}

// EnhancedSystem orchestrates the enhanced system
type EnhancedSystem struct {
	taskReasoner      *TaskReasoner
	providerSelector  *ProviderSelector
	spoOptimizer      *SPOOptimizer
	providers         []*Provider
	metricsStorage    *MetricsStorage
	rateLimitManager  *RateLimitManager
	healthCalculator  *HealthScoreCalculator
}

// TokenEstimator estimates token usage
type TokenEstimator struct {
	baseTokensPerWord float64
	complexityMultipliers map[ComplexityLevel]float64
}

// HealthScoreCalculator calculates provider health scores
type HealthScoreCalculator struct {
	WindowSize         time.Duration
	MaxUsageRatio      float64
	ReliabilityWeight  float64
	CostWeight         float64
	RateLimitWeight    float64
}

// Request represents a processing request
type Request struct {
	ID               string                     `json:"id"`
	Content          string                     `json:"content"`
	TaskType         TaskType                   `json:"task_type"`
	Priority         int                        `json:"priority"`
	MaxTokens        int64                      `json:"max_tokens"`
	Temperature      float64                    `json:"temperature"`
	Context          context.Context            `json:"-"`
	Metadata         map[string]interface{}     `json:"metadata"`
	CreatedAt        time.Time                  `json:"created_at"`
}

// Response represents a processing response
type Response struct {
	RequestID        string                     `json:"request_id"`
	Content          string                     `json:"content"`
	TokensUsed       int64                      `json:"tokens_used"`
	Cost             float64                    `json:"cost"`
	Latency          time.Duration              `json:"latency"`
	Provider         string                     `json:"provider"`
	Model            string                     `json:"model"`
	Success          bool                       `json:"success"`
	Error            string                     `json:"error,omitempty"`
	Complexity       *TaskComplexity            `json:"complexity,omitempty"`
	Assignment       *ProviderAssignment        `json:"assignment,omitempty"`
	Metadata         map[string]interface{}     `json:"metadata"`
	ProcessedAt      time.Time                  `json:"processed_at"`
}

// SystemConfig represents system configuration
type SystemConfig struct {
	Providers        []*Provider                `json:"providers"`
	DefaultModel     string                     `json:"default_model"`
	MaxRetries       int                        `json:"max_retries"`
	Timeout          time.Duration              `json:"timeout"`
	CostThreshold    float64                    `json:"cost_threshold"`
	HealthThreshold  float64                    `json:"health_threshold"`
	MetricsEnabled   bool                       `json:"metrics_enabled"`
	DatabasePath     string                     `json:"database_path"`
	Metadata         map[string]interface{}     `json:"metadata"`
}