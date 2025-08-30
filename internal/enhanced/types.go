package enhanced

import (
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

// String returns string representation of ComplexityLevel
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

// RequestInput represents an incoming request
type RequestInput struct {
	ID          string                 `json:"id"`
	Content     string                 `json:"content"`
	Context     map[string]interface{} `json:"context,omitempty"`
	Priority    int                    `json:"priority,omitempty"`
	Constraints map[string]interface{} `json:"constraints,omitempty"`
	Mode        string                 `json:"mode,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
}

// RequestResult represents the result of processing a request
type RequestResult struct {
	ID               string                 `json:"id"`
	Status           string                 `json:"status"`
	Complexity       TaskComplexity         `json:"complexity"`
	OptimizedPrompt  OptimizedPrompt        `json:"optimized_prompt"`
	Assignment       ProviderAssignment     `json:"assignment"`
	Response         map[string]interface{} `json:"response,omitempty"`
	TotalCost        float64                `json:"total_cost"`
	TotalDuration    string                 `json:"total_duration"`
	Error            string                 `json:"error,omitempty"`
	CreatedAt        time.Time              `json:"created_at"`
	CompletedAt      *time.Time             `json:"completed_at,omitempty"`
}

// TaskComplexity represents the analyzed complexity of a task
type TaskComplexity struct {
	Overall              ComplexityLevel            `json:"overall"`
	Reasoning            ComplexityLevel            `json:"reasoning"`
	Knowledge            ComplexityLevel            `json:"knowledge"`
	Creativity           ComplexityLevel            `json:"creativity,omitempty"`
	Computation          ComplexityLevel            `json:"computation"`
	Coordination         ComplexityLevel            `json:"coordination,omitempty"`
	Score                float64                    `json:"score"`
	Factors              map[string]interface{}     `json:"factors,omitempty"`
	EstimatedTokens      int                        `json:"estimated_tokens,omitempty"`
	RequiredCapabilities []string                   `json:"required_capabilities,omitempty"`
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
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// ProviderAssignment represents the assignment of a task to a provider
type ProviderAssignment struct {
	TaskID          string                 `json:"task_id,omitempty"`
	ProviderID      string                 `json:"provider_id"`
	ProviderName    string                 `json:"provider_name,omitempty"`
	ProviderTier    ProviderTier           `json:"provider_tier,omitempty"`
	Model           string                 `json:"model,omitempty"`
	Tier            string                 `json:"tier,omitempty"`
	Confidence      float64                `json:"confidence"`
	EstimatedCost   float64                `json:"estimated_cost"`
	EstimatedTime   int64                  `json:"estimated_time"`
	Reasoning       string                 `json:"reasoning"`
	Alternatives    []AlternativeProvider  `json:"alternatives,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// AlternativeProvider represents an alternative provider option
type AlternativeProvider struct {
	ProviderID      string  `json:"provider_id"`
	ProviderName    string  `json:"provider_name,omitempty"`
	Model           string  `json:"model,omitempty"`
	Confidence      float64 `json:"confidence"`
	EstimatedCost   float64 `json:"estimated_cost"`
	EstimatedTime   int64   `json:"estimated_time,omitempty"`
	Reasoning       string  `json:"reasoning"`
}

// ProviderTier represents the tier of a provider
type ProviderTier string

const (
	OfficialTier   ProviderTier = "official"
	CommunityTier  ProviderTier = "community"
	UnofficialTier ProviderTier = "unofficial"
)

// Provider represents an AI provider with simplified 6-column structure
type Provider struct {
	ID       string          `json:"id,omitempty"`
	Name     string          `json:"name"`
	Tier     ProviderTier    `json:"tier"`
	BaseURL  string          `json:"base_url"`
	APIKey   string          `json:"api_key"`
	Models   string          `json:"models"`
	Other    string          `json:"other"`
	Metrics  ProviderMetrics `json:"metrics"`
}

// ProviderMetrics tracks performance metrics for a provider
type ProviderMetrics struct {
	TotalRequests     int64     `json:"total_requests"`
	SuccessfulRequests int64    `json:"successful_requests"`
	FailedRequests    int64     `json:"failed_requests"`
	SuccessRate       float64   `json:"success_rate"`
	AverageLatency    float64   `json:"average_latency"`
	AverageCost       float64   `json:"average_cost"`
	QualityScore      float64   `json:"quality_score"`
	CostEfficiency    float64   `json:"cost_efficiency"`
	ReliabilityScore  float64   `json:"reliability_score"`
	RequestCount      int64     `json:"request_count"`
	ErrorCount        int64     `json:"error_count"`
	LastUsed          time.Time `json:"last_used"`
	LastUpdated       time.Time `json:"last_updated"`
}

// SystemMetrics represents overall system performance metrics
type SystemMetrics struct {
	TotalRequests        int64              `json:"total_requests"`
	SuccessfulRequests   int64              `json:"successful_requests"`
	FailedRequests       int64              `json:"failed_requests"`
	AverageResponseTime  float64            `json:"average_response_time"`
	TotalCost            float64            `json:"total_cost"`
	CostSavings          float64            `json:"cost_savings"`
	ActiveRequests       int                `json:"active_requests"`
	ProviderHealthScores map[string]float64 `json:"provider_health_scores"`
	LastUpdated          time.Time          `json:"last_updated"`
	SystemUptime         time.Time          `json:"system_uptime"`
}

// ExecutionStatus represents the status of task execution
type ExecutionStatus string

const (
	StatusPending   ExecutionStatus = "pending"
	StatusRunning   ExecutionStatus = "running"
	StatusCompleted ExecutionStatus = "completed"
	StatusFailed    ExecutionStatus = "failed"
	StatusCancelled ExecutionStatus = "cancelled"
)

// ProcessingRequest represents a request being processed through the system
type ProcessingRequest struct {
	ID              string               `json:"id"`
	Input           RequestInput         `json:"input"`
	Complexity      TaskComplexity       `json:"complexity"`
	OptimizedPrompt OptimizedPrompt      `json:"optimized_prompt"`
	Assignment      ProviderAssignment   `json:"assignment"`
	Status          ExecutionStatus      `json:"status"`
	Result          string               `json:"result"`
	TotalCost       float64              `json:"total_cost"`
	TotalDuration   time.Duration        `json:"total_duration"`
	CreatedAt       time.Time            `json:"created_at"`
	CompletedAt     *time.Time           `json:"completed_at,omitempty"`
	Error           string               `json:"error,omitempty"`
}

// Helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}