package enhanced

import (
	"fmt"
	"time"
)

// ComplexityLevel represents different levels of task complexity
type ComplexityLevel string

const (
	VeryHigh ComplexityLevel = "very_high"
	High     ComplexityLevel = "high"
	Medium   ComplexityLevel = "medium"
	Low      ComplexityLevel = "low"
)

// String returns the string representation of ComplexityLevel
func (c ComplexityLevel) String() string {
	return string(c)
}

// TaskComplexity represents the complexity analysis of a task
type TaskComplexity struct {
	Reasoning     float64 `json:"reasoning"`
	Knowledge     float64 `json:"knowledge"`
	Computation   float64 `json:"computation"`
	Coordination  float64 `json:"coordination"`
	Overall       float64 `json:"overall"`
	Score         float64 `json:"score"`
}

// RequestInput represents the input for a request
type RequestInput struct {
	Query       string                 `json:"query"`
	Context     map[string]interface{} `json:"context,omitempty"`
	Constraints map[string]interface{} `json:"constraints,omitempty"`
	Mode        string                 `json:"mode,omitempty"`
}

// RequestResult represents the result of a request
type RequestResult struct {
	Response    string                 `json:"response"`
	Provider    string                 `json:"provider"`
	Model       string                 `json:"model"`
	Complexity  TaskComplexity         `json:"complexity"`
	Cost        float64                `json:"cost"`
	Duration    time.Duration          `json:"duration"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// SystemMetrics represents system performance metrics
type SystemMetrics struct {
	TotalRequests     int64         `json:"total_requests"`
	SuccessfulRequests int64        `json:"successful_requests"`
	FailedRequests    int64         `json:"failed_requests"`
	AverageLatency    time.Duration `json:"average_latency"`
	TotalCost         float64       `json:"total_cost"`
	LastUpdated       time.Time     `json:"last_updated"`
}

// OptimizedPrompt represents an optimized prompt result
type OptimizedPrompt struct {
	OriginalPrompt string  `json:"original_prompt"`
	OptimizedText  string  `json:"optimized_text"`
	Reasoning      string  `json:"reasoning"`
	TokenReduction int     `json:"token_reduction"`
	CostSavings    float64 `json:"cost_savings"`
}

// ExecutionResult represents the result of task execution
type ExecutionResult struct {
	Success   bool                   `json:"success"`
	Result    interface{}            `json:"result"`
	Error     string                 `json:"error,omitempty"`
	Duration  time.Duration          `json:"duration"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// Provider represents a language model provider
type Provider struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Models      []string `json:"models"`
	Tier        string   `json:"tier"`
	Enabled     bool     `json:"enabled"`
	Priority    int      `json:"priority"`
	Pricing     ProviderPricing `json:"pricing"`
	Metrics     ProviderMetrics `json:"metrics"`
	HealthMetrics ProviderHealthMetrics `json:"health_metrics"`
	Other       map[string]interface{} `json:"other,omitempty"`
}

// ProviderPricing represents pricing information for a provider
type ProviderPricing struct {
	InputTokenCost  float64   `json:"input_token_cost"`
	OutputTokenCost float64   `json:"output_token_cost"`
	Currency        string    `json:"currency"`
	LastUpdated     time.Time `json:"last_updated"`
}

// ProviderMetrics represents performance metrics for a provider
type ProviderMetrics struct {
	RequestCount    int64         `json:"request_count"`
	SuccessRate     float64       `json:"success_rate"`
	AverageLatency  time.Duration `json:"average_latency"`
	ErrorRate       float64       `json:"error_rate"`
	TotalCost       float64       `json:"total_cost"`
	LastUpdated     time.Time     `json:"last_updated"`
}

// ProviderHealthMetrics represents health metrics for a provider
type ProviderHealthMetrics struct {
	Status          string    `json:"status"`
	Uptime          float64   `json:"uptime"`
	ResponseTime    time.Duration `json:"response_time"`
	ErrorCount      int64     `json:"error_count"`
	LastHealthCheck time.Time `json:"last_health_check"`
}

// ProviderAssignment represents the assignment of a provider to a task
type ProviderAssignment struct {
	Provider      string                `json:"provider"`
	Model         string                `json:"model"`
	Reasoning     string                `json:"reasoning"`
	Confidence    float64               `json:"confidence"`
	EstimatedCost float64               `json:"estimated_cost"`
	EstimatedTime time.Duration         `json:"estimated_time"`
	Alternatives  []AlternativeProvider `json:"alternatives,omitempty"`
	ProviderName  string                `json:"provider_name"`
	ProviderTier  string                `json:"provider_tier"`
	Tier          string                `json:"tier"`
}

// AlternativeProvider represents an alternative provider option
type AlternativeProvider struct {
	Provider      string        `json:"provider"`
	Model         string        `json:"model"`
	Confidence    float64       `json:"confidence"`
	EstimatedCost float64       `json:"estimated_cost"`
	EstimatedTime time.Duration `json:"estimated_time"`
	Reasoning     string        `json:"reasoning"`
	ProviderID    string        `json:"provider_id"`
	ProviderName  string        `json:"provider_name"`
}

// EnhancedSystem represents the main enhanced system
type EnhancedSystem struct {
	providers         []*Provider
	taskReasoner      *TaskReasoner
	providerSelector  *ProviderSelector
	spoOptimizer      *SPOOptimizer
	healthMonitor     *ProviderHealthMonitor
	metrics           SystemMetrics
}

// TaskReasoner handles task complexity analysis
type TaskReasoner struct {
	config *Config
}

// ProviderSelector handles provider selection logic
type ProviderSelector struct {
	providers []*Provider
}

// SPOOptimizer handles prompt optimization
type SPOOptimizer struct {
	config *Config
}

// ProviderHealthMonitor monitors provider health
type ProviderHealthMonitor struct {
	providers []*Provider
}

// Config represents system configuration
type Config struct {
	Providers []ProviderConfig `json:"providers"`
	Settings  map[string]interface{} `json:"settings"`
}

// ProviderConfig represents provider configuration
type ProviderConfig struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Enabled  bool   `json:"enabled"`
	Priority int    `json:"priority"`
}

func (c ComplexityLevel) ToFloat64() float64 {
	switch c {
	case VeryHigh:
		return 4.0
	case High:
		return 3.0
	case Medium:
		return 2.0
	case Low:
		return 1.0
	default:
		return 2.0
	}
}

func (c ComplexityLevel) ToInt() int {
	switch c {
	case VeryHigh:
		return 4
	case High:
		return 3
	case Medium:
		return 2
	case Low:
		return 1
	default:
		return 2
	}
}

func FloatToComplexityLevel(f float64) ComplexityLevel {
	switch {
	case f >= 3.5:
		return VeryHigh
	case f >= 2.5:
		return High
	case f >= 1.5:
		return Medium
	default:
		return Low
	}
}

func (p *Provider) String() string {
	return fmt.Sprintf("Provider{ID: %s, Name: %s, Tier: %s}", p.ID, p.Name, p.Tier)
}