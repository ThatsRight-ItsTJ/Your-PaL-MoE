package enhanced

import (
	"fmt"
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
		return "Low"
	case Medium:
		return "Medium"
	case High:
		return "High"
	case VeryHigh:
		return "VeryHigh"
	default:
		return "Unknown"
	}
}

// TaskComplexity represents the complexity analysis of a task
type TaskComplexity struct {
	Reasoning    float64 `json:"reasoning"`
	Knowledge    float64 `json:"knowledge"`
	Computation  float64 `json:"computation"`
	Coordination float64 `json:"coordination"`
	Overall      float64 `json:"overall"`
	Score        float64 `json:"score"`
}

// RequestInput represents input for a request
type RequestInput struct {
	ID          string                 `json:"id"`
	Content     string                 `json:"content"`
	Context     map[string]interface{} `json:"context,omitempty"`
	Constraints map[string]interface{} `json:"constraints,omitempty"`
	Mode        string                 `json:"mode,omitempty"`
}

// OptimizedPrompt represents an optimized prompt result
type OptimizedPrompt struct {
	Original     string   `json:"original"`
	Optimized    string   `json:"optimized"`
	Iterations   int      `json:"iterations"`
	Improvements []string `json:"improvements"`
	Confidence   float64  `json:"confidence"`
	CostSavings  float64  `json:"cost_savings"`
}

// Provider represents a provider configuration
type Provider struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Models   []string `json:"models"`
	Endpoint string   `json:"endpoint"`
	APIKey   string   `json:"api_key"`
	Tier     string   `json:"tier"`
	Priority int      `json:"priority"`
	
	// Provider metrics and health
	Metrics       *ProviderMetrics `json:"metrics,omitempty"`
	HealthMetrics *HealthMetrics   `json:"health_metrics,omitempty"`
	Pricing       *ProviderPricing `json:"pricing,omitempty"`
	Other         map[string]interface{} `json:"other,omitempty"`
}

// ProviderMetrics represents provider performance metrics
type ProviderMetrics struct {
	RequestCount    int64         `json:"request_count"`
	SuccessCount    int64         `json:"success_count"`
	FailureCount    int64         `json:"failure_count"`
	AverageLatency  time.Duration `json:"average_latency"`
	TotalTokens     int64         `json:"total_tokens"`
	TotalCost       float64       `json:"total_cost"`
	LastUpdated     time.Time     `json:"last_updated"`
}

// HealthMetrics represents provider health status
type HealthMetrics struct {
	IsHealthy        bool               `json:"is_healthy"`
	LastHealthCheck  time.Time          `json:"last_health_check"`
	ConsecutiveFails int                `json:"consecutive_fails"`
	RateLimitStatus  *RateLimitStatus   `json:"rate_limit_status,omitempty"`
}

// RateLimitStatus represents rate limiting information
type RateLimitStatus struct {
	RequestsPerMinute int       `json:"requests_per_minute"`
	RequestsRemaining int       `json:"requests_remaining"`
	ResetTime         time.Time `json:"reset_time"`
}

// ProviderPricing represents provider pricing information
type ProviderPricing struct {
	InputTokenCost  float64   `json:"input_token_cost"`
	OutputTokenCost float64   `json:"output_token_cost"`
	Currency        string    `json:"currency"`
	LastUpdated     time.Time `json:"last_updated"`
}

// ProviderAssignment represents a provider assignment result
type ProviderAssignment struct {
	ProviderID     string                `json:"provider_id"`
	ProviderName   string                `json:"provider_name"`
	ProviderTier   string                `json:"provider_tier"`
	Model          string                `json:"model"`
	Tier           string                `json:"tier"`
	Score          float64               `json:"score"`
	Reasoning      string                `json:"reasoning"`
	EstimatedCost  float64               `json:"estimated_cost"`
	EstimatedTime  time.Duration         `json:"estimated_time"`
	Alternatives   []AlternativeProvider `json:"alternatives,omitempty"`
}

// AlternativeProvider represents an alternative provider option
type AlternativeProvider struct {
	ProviderID     string        `json:"provider_id"`
	ProviderName   string        `json:"provider_name"`
	Model          string        `json:"model"`
	Score          float64       `json:"score"`
	EstimatedCost  float64       `json:"estimated_cost"`
	EstimatedTime  time.Duration `json:"estimated_time"`
	Reasoning      string        `json:"reasoning"`
}

// ExecutionResult represents the result of task execution
type ExecutionResult struct {
	TaskID    string         `json:"task_id"`
	Success   bool           `json:"success"`
	Output    string         `json:"output"`
	Error     string         `json:"error,omitempty"`
	Duration  time.Duration  `json:"duration"`
	Quality   QualityMetrics `json:"quality"`
	Timestamp time.Time      `json:"timestamp"`
}

// QualityMetrics represents quality assessment metrics
type QualityMetrics struct {
	OverallScore float64 `json:"overall_score"`
	Accuracy     float64 `json:"accuracy"`
	Completeness float64 `json:"completeness"`
	Relevance    float64 `json:"relevance"`
}

// RequestResult represents the result of processing a request
type RequestResult struct {
	RequestID       string              `json:"request_id"`
	Success         bool                `json:"success"`
	Output          string              `json:"output"`
	Error           string              `json:"error,omitempty"`
	Provider        *ProviderAssignment `json:"provider"`
	OptimizedPrompt *OptimizedPrompt    `json:"optimized_prompt,omitempty"`
	Complexity      *TaskComplexity     `json:"complexity"`
	Duration        time.Duration       `json:"duration"`
	TokensUsed      int                 `json:"tokens_used"`
	Cost            float64             `json:"cost"`
	Timestamp       time.Time           `json:"timestamp"`
}

// SystemMetrics represents overall system performance metrics
type SystemMetrics struct {
	TotalRequests     int64         `json:"total_requests"`
	SuccessfulRequests int64        `json:"successful_requests"`
	FailedRequests    int64         `json:"failed_requests"`
	AverageLatency    time.Duration `json:"average_latency"`
	TotalCost         float64       `json:"total_cost"`
	TotalTokens       int64         `json:"total_tokens"`
	ActiveProviders   int           `json:"active_providers"`
	LastUpdated       time.Time     `json:"last_updated"`
}

// Config represents configuration for the enhanced system
type Config struct {
	SPO struct {
		MaxIterations    int     `json:"max_iterations"`
		SamplesPerRound  int     `json:"samples_per_round"`
		ConvergenceRate  float64 `json:"convergence_rate"`
		CacheTTL         int     `json:"cache_ttl"`
		CacheSize        int     `json:"cache_size"`
	} `json:"spo"`
	
	ProviderSelection struct {
		ReliabilityWeight float64 `json:"reliability_weight"`
		CostWeight        float64 `json:"cost_weight"`
		RateLimitWeight   float64 `json:"rate_limit_weight"`
	} `json:"provider_selection"`
}

// EnhancedSystem represents the main enhanced system
type EnhancedSystem struct {
	providers        []*Provider
	taskReasoning    *TaskReasoningEngine
	spoOptimizer     *SPOOptimizer
	providerSelector *EnhancedProviderSelector
	config           *Config
	metrics          *SystemMetrics
}

// TaskReasoningEngine analyzes task complexity and requirements
type TaskReasoningEngine struct {
	config *Config
}

// SPOOptimizer implements self-supervised prompt optimization
type SPOOptimizer struct {
	config *Config
}

// EnhancedProviderSelector selects optimal providers
type EnhancedProviderSelector struct {
	providers []*Provider
	config    *Config
}

// NewEnhancedSystem creates a new enhanced system instance
func NewEnhancedSystem(config *Config) (*EnhancedSystem, error) {
	if config == nil {
		config = &Config{}
		// Set default values
		config.SPO.MaxIterations = 5
		config.SPO.SamplesPerRound = 3
		config.SPO.ConvergenceRate = 0.1
		config.SPO.CacheTTL = 3600
		config.SPO.CacheSize = 1000
		config.ProviderSelection.ReliabilityWeight = 0.4
		config.ProviderSelection.CostWeight = 0.4
		config.ProviderSelection.RateLimitWeight = 0.2
	}

	return &EnhancedSystem{
		providers:        make([]*Provider, 0),
		taskReasoning:    &TaskReasoningEngine{config: config},
		spoOptimizer:     &SPOOptimizer{config: config},
		providerSelector: &EnhancedProviderSelector{config: config},
		config:           config,
		metrics:          &SystemMetrics{LastUpdated: time.Now()},
	}, nil
}

// String method for ComplexityLevel to satisfy fmt.Stringer interface
func (c ComplexityLevel) GoString() string {
	return fmt.Sprintf("ComplexityLevel(%d)", int(c))
}