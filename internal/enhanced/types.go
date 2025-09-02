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

// ProviderTier represents the tier of a provider
type ProviderTier string

const (
	OfficialTier   ProviderTier = "official"
	CommunityTier  ProviderTier = "community"
	UnofficialTier ProviderTier = "unofficial"
)

// TaskComplexity represents the complexity analysis of a task
type TaskComplexity struct {
	Reasoning    float64         `json:"reasoning"`
	Knowledge    float64         `json:"knowledge"`
	Computation  float64         `json:"computation"`
	Coordination float64         `json:"coordination"`
	Overall      ComplexityLevel `json:"overall"`
	Score        float64         `json:"score"`
}

// RequestInput represents input for a request
type RequestInput struct {
	ID          string                 `json:"id"`
	Content     string                 `json:"content"`
	Query       string                 `json:"query"`
	Context     map[string]interface{} `json:"context,omitempty"`
	Constraints map[string]interface{} `json:"constraints,omitempty"`
	Mode        string                 `json:"mode,omitempty"`
}

// OptimizedPrompt represents an optimized prompt result
type OptimizedPrompt struct {
	Original        string   `json:"original"`
	Optimized       string   `json:"optimized"`
	OriginalPrompt  string   `json:"original_prompt"`
	OptimizedText   string   `json:"optimized_text"`
	Iterations      int      `json:"iterations"`
	Improvements    []string `json:"improvements"`
	Confidence      float64  `json:"confidence"`
	CostSavings     float64  `json:"cost_savings"`
}

// Provider represents a provider configuration
type Provider struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Models   []string `json:"models"`
	Endpoint string   `json:"endpoint"`
	BaseURL  string   `json:"base_url"`
	APIKey   string   `json:"api_key"`
	Tier     ProviderTier `json:"tier"`
	Priority int      `json:"priority"`
	Other    string   `json:"other"`

	// Provider metrics and health
	Metrics       ProviderMetrics        `json:"metrics,omitempty"`
	HealthMetrics *ProviderHealthMetrics `json:"health_metrics,omitempty"`
	Pricing       *ProviderPricing       `json:"pricing,omitempty"`
}

// ProviderMetrics represents provider performance metrics
type ProviderMetrics struct {
	RequestCount       int64         `json:"request_count"`
	SuccessCount       int64         `json:"success_count"`
	SuccessfulRequests int64         `json:"successful_requests"`
	FailureCount       int64         `json:"failure_count"`
	ErrorCount         int64         `json:"error_count"`
	AverageLatency     float64       `json:"average_latency"`
	TotalTokens        int64         `json:"total_tokens"`
	TotalCost          float64       `json:"total_cost"`
	AverageCost        float64       `json:"average_cost"`
	LastUpdated        time.Time     `json:"last_updated"`
	SuccessRate        float64       `json:"success_rate"`
	QualityScore       float64       `json:"quality_score"`
	CostEfficiency     float64       `json:"cost_efficiency"`
	ReliabilityScore   float64       `json:"reliability_score"`
}

// ProviderHealthMetrics extends existing ProviderMetrics with cost tracking
type ProviderHealthMetrics struct {
	Status              string                       `json:"status"`
	Uptime              float64                      `json:"uptime"`
	ResponseTime        time.Duration                `json:"response_time"`
	ErrorCount          int64                        `json:"error_count"`
	UsageHistory        map[string][]UsageRecord     `json:"usage_history"` // model -> records
	LastHealthCheck     time.Time                    `json:"last_health_check"`
	HealthScore         float64                      `json:"health_score"`
	CostEfficiencyScore float64                      `json:"cost_efficiency_score"`
	RateLimitStatus     *RateLimitStatus             `json:"rate_limit_status"`
	ConsecutiveFails    int                          `json:"consecutive_fails"`
	IsHealthy           bool                         `json:"is_healthy"`
}

// UsageRecord represents a single usage entry with cost tracking
type UsageRecord struct {
	Timestamp   time.Time `json:"timestamp"`
	Count       int64     `json:"count"`
	Failures    int64     `json:"failures"`
	Latency     float64   `json:"latency_ms"`
	TokensUsed  int64     `json:"tokens_used"`
	Cost        float64   `json:"cost"`
	RateLimited bool      `json:"rate_limited"`
}

// RateLimitStatus tracks current rate limit state
type RateLimitStatus struct {
	RequestsPerMinute int64     `json:"requests_per_minute"`
	RequestsRemaining int64     `json:"requests_remaining"`
	TokensPerMinute   int64     `json:"tokens_per_minute"`
	TokensRemaining   int64     `json:"tokens_remaining"`
	ResetTime         time.Time `json:"reset_time"`
	LastRateLimitHit  time.Time `json:"last_rate_limit_hit"`
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
	TaskID         string                `json:"task_id"`
	Provider       string                `json:"provider"`
	ProviderID     string                `json:"provider_id"`
	ProviderName   string                `json:"provider_name"`
	ProviderTier   ProviderTier          `json:"provider_tier"`
	Model          string                `json:"model"`
	Tier           string                `json:"tier"`
	Score          float64               `json:"score"`
	Confidence     float64               `json:"confidence"`
	Reasoning      string                `json:"reasoning"`
	EstimatedCost  float64               `json:"estimated_cost"`
	EstimatedTime  int64                 `json:"estimated_time"`
	Alternatives   []AlternativeProvider `json:"alternatives,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// AlternativeProvider represents an alternative provider option
type AlternativeProvider struct {
	ProviderID    string        `json:"provider_id"`
	ProviderName  string        `json:"provider_name"`
	Model         string        `json:"model"`
	Score         float64       `json:"score"`
	Confidence    float64       `json:"confidence"`
	EstimatedCost float64       `json:"estimated_cost"`
	EstimatedTime int64         `json:"estimated_time"`
	Reasoning     string        `json:"reasoning"`
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
	Response        string              `json:"response"`
	Success         bool                `json:"success"`
	Output          string              `json:"output"`
	Error           string              `json:"error,omitempty"`
	Provider        *ProviderAssignment `json:"provider"`
	Model           string              `json:"model"`
	OptimizedPrompt *OptimizedPrompt    `json:"optimized_prompt,omitempty"`
	Complexity      TaskComplexity      `json:"complexity"`
	Duration        time.Duration       `json:"duration"`
	TokensUsed      int                 `json:"tokens_used"`
	Cost            float64             `json:"cost"`
	Timestamp       time.Time           `json:"timestamp"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// SystemMetrics represents overall system performance metrics
type SystemMetrics struct {
	TotalRequests      int64         `json:"total_requests"`
	SuccessfulRequests int64         `json:"successful_requests"`
	FailedRequests     int64         `json:"failed_requests"`
	AverageLatency     time.Duration `json:"average_latency"`
	TotalCost          float64       `json:"total_cost"`
	TotalTokens        int64         `json:"total_tokens"`
	ActiveProviders    int           `json:"active_providers"`
	LastUpdated        time.Time     `json:"last_updated"`
}

// Config represents configuration for the enhanced system
type Config struct {
	SPO struct {
		MaxIterations   int     `json:"max_iterations"`
		SamplesPerRound int     `json:"samples_per_round"`
		ConvergenceRate float64 `json:"convergence_rate"`
		CacheTTL        int     `json:"cache_ttl"`
		CacheSize       int     `json:"cache_size"`
	} `json:"spo"`

	ProviderSelection struct {
		ReliabilityWeight float64 `json:"reliability_weight"`
		CostWeight        float64 `json:"cost_weight"`
		RateLimitWeight   float64 `json:"rate_limit_weight"`
	} `json:"provider_selection"`
}

// EnhancedSystem represents the main enhanced system
type EnhancedSystem struct {
	providers         []*Provider
	taskReasoner      *TaskReasoner
	taskReasoning     *TaskReasoningEngine
	spoOptimizer      *SPOOptimizer
	providerSelector  *ProviderSelector
	healthMonitor     *ProviderHealthMonitor
	config            *Config
	metrics           SystemMetrics
}

// TaskReasoningEngine analyzes task complexity and requirements
type TaskReasoningEngine struct {
	config *Config
}

// TaskReasoner analyzes task complexity and requirements
type TaskReasoner struct {
	config *Config
}

// ProviderSelector selects optimal providers
type ProviderSelector struct {
	providers []*Provider
}

// ProviderHealthMonitor monitors provider health
type ProviderHealthMonitor struct {
	providers []*Provider
}

// SPOOptimizer implements self-supervised prompt optimization
type SPOOptimizer struct {
	config *Config
}

// CostSavingsReport represents cost optimization analytics
type CostSavingsReport struct {
	TotalSavings     float64                        `json:"total_savings"`
	AverageSavings   float64                        `json:"average_savings"`
	RequestCount     int64                          `json:"request_count"`
	TopProviders     []string                       `json:"top_providers"`
	SavingsByTier    map[string]float64             `json:"savings_by_tier"`
	DailySavings     map[string]float64             `json:"daily_savings"`
	OptimizationRate float64                        `json:"optimization_rate"`
	Metadata         map[string]interface{}         `json:"metadata"`
}

// String method for ComplexityLevel to satisfy fmt.Stringer interface
func (c ComplexityLevel) GoString() string {
	return fmt.Sprintf("ComplexityLevel(%d)", int(c))
}