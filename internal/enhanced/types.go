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

// String returns the string representation of ComplexityLevel
func (cl ComplexityLevel) String() string {
	switch cl {
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

// TaskType represents different types of tasks
type TaskType string

const (
	TaskTypeText        TaskType = "text"
	TaskTypeCode        TaskType = "code"
	TaskTypeImage       TaskType = "image"
	TaskTypeAudio       TaskType = "audio"
	TaskTypeVideo       TaskType = "video"
	TaskTypeMultimodal  TaskType = "multimodal"
)

// TaskComplexity represents the complexity analysis of a task
type TaskComplexity struct {
	Overall              ComplexityLevel        `json:"overall"`
	Reasoning            ComplexityLevel        `json:"reasoning"`
	Mathematical         ComplexityLevel        `json:"mathematical"`
	Creative             ComplexityLevel        `json:"creative"`
	Factual              ComplexityLevel        `json:"factual"`
	TokenEstimate        int64                  `json:"token_estimate"`
	RequiredCapabilities []string               `json:"required_capabilities"`
	Metadata             map[string]interface{} `json:"metadata"`
}

// ProviderTier represents the tier of a provider
type ProviderTier string

const (
	OfficialTier   ProviderTier = "official"
	CommunityTier  ProviderTier = "community"
	UnofficialTier ProviderTier = "unofficial"
)

// Provider represents an AI provider
type Provider struct {
	Name           string                 `json:"name"`
	BaseURL        string                 `json:"base_url"`
	Models         []string               `json:"models"`
	Tier           ProviderTier           `json:"tier"`
	MaxTokens      int64                  `json:"max_tokens"`
	CostPerToken   float64                `json:"cost_per_token"`
	Capabilities   []string               `json:"capabilities"`
	RateLimits     map[string]int64       `json:"rate_limits"`
	HealthMetrics  *ProviderHealthMetrics `json:"health_metrics,omitempty"`
	Metadata       map[string]interface{} `json:"metadata"`
	LastUpdated    time.Time              `json:"last_updated"`
}

// ProviderAssignment represents a provider assignment for a task
type ProviderAssignment struct {
	Provider        *Provider              `json:"provider"`
	Model           string                 `json:"model"`
	Confidence      float64                `json:"confidence"`
	EstimatedCost   float64                `json:"estimated_cost"`
	EstimatedTokens int64                  `json:"estimated_tokens"`
	Reasoning       string                 `json:"reasoning"`
	Alternatives    []*Provider            `json:"alternatives"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// ProviderScore represents a provider's score for a task
type ProviderScore struct {
	Provider   *Provider `json:"provider"`
	Score      float64   `json:"score"`
	Confidence float64   `json:"confidence"`
	Reasoning  string    `json:"reasoning"`
}

// RequestInput represents input for processing
type RequestInput struct {
	Content     string                 `json:"content"`
	TaskType    TaskType               `json:"task_type"`
	MaxTokens   int64                  `json:"max_tokens,omitempty"`
	Temperature float64                `json:"temperature,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ProcessResponse represents the response from processing
type ProcessResponse struct {
	Content        string                 `json:"content"`
	Provider       *Provider              `json:"provider"`
	Model          string                 `json:"model"`
	Complexity     TaskComplexity         `json:"complexity"`
	ProcessingTime time.Duration          `json:"processing_time"`
	TokensUsed     int64                  `json:"tokens_used"`
	Cost           float64                `json:"cost"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// OptimizedPrompt represents an optimized prompt
type OptimizedPrompt struct {
	OriginalPrompt    string                 `json:"original_prompt"`
	OptimizedPrompt   string                 `json:"optimized_prompt"`
	Complexity        ComplexityLevel        `json:"complexity"`
	OptimizationRules []string               `json:"optimization_rules"`
	Metadata          map[string]interface{} `json:"metadata"`
	ProcessingTime    time.Duration          `json:"processing_time"`
}

// SystemMetrics represents system-wide metrics
type SystemMetrics struct {
	TotalRequests     int64                        `json:"total_requests"`
	SuccessfulRequests int64                       `json:"successful_requests"`
	FailedRequests    int64                        `json:"failed_requests"`
	AverageLatency    time.Duration                `json:"average_latency"`
	ComplexityDistribution map[ComplexityLevel]int64 `json:"complexity_distribution"`
	ProviderUsage     map[string]int64             `json:"provider_usage"`
	TotalCost         float64                      `json:"total_cost"`
	TotalTokens       int64                        `json:"total_tokens"`
	StartTime         time.Time                    `json:"start_time"`
}

// NewSystemMetrics creates a new SystemMetrics instance
func NewSystemMetrics() *SystemMetrics {
	return &SystemMetrics{
		ComplexityDistribution: make(map[ComplexityLevel]int64),
		ProviderUsage:         make(map[string]int64),
		StartTime:             time.Now(),
	}
}

// IncrementTotalRequests increments the total request count
func (sm *SystemMetrics) IncrementTotalRequests() {
	sm.TotalRequests++
}

// IncrementSuccessfulRequests increments the successful request count
func (sm *SystemMetrics) IncrementSuccessfulRequests() {
	sm.SuccessfulRequests++
}

// IncrementFailedRequests increments the failed request count
func (sm *SystemMetrics) IncrementFailedRequests() {
	sm.FailedRequests++
}

// RecordComplexity records complexity distribution
func (sm *SystemMetrics) RecordComplexity(level ComplexityLevel) {
	sm.ComplexityDistribution[level]++
}

// RecordProviderUsage records provider usage
func (sm *SystemMetrics) RecordProviderUsage(providerName string) {
	sm.ProviderUsage[providerName]++
}

// AddCost adds to the total cost
func (sm *SystemMetrics) AddCost(cost float64) {
	sm.TotalCost += cost
}

// AddTokens adds to the total token count
func (sm *SystemMetrics) AddTokens(tokens int64) {
	sm.TotalTokens += tokens
}

// EnhancedSystem represents the main enhanced system
type EnhancedSystem struct {
	selector      *EnhancedProviderSelector
	reasoner      *TaskReasoner
	optimizer     *SPOOptimizer
	healthMonitor *ProviderHealthMonitor
	providers     []*Provider
	metrics       *SystemMetrics
}

// TaskReasoner represents a task complexity analyzer
type TaskReasoner struct {
	config *TaskReasonerConfig
}

// TaskReasonerConfig represents configuration for task reasoner
type TaskReasonerConfig struct {
	ComplexityWeights map[string]float64                   `json:"complexity_weights"`
	TokenMultipliers  map[ComplexityLevel]float64          `json:"token_multipliers"`
}

// SPOOptimizer represents a self-prompt optimizer
type SPOOptimizer struct {
	templates         map[string]string
	optimizationRules map[ComplexityLevel][]string
	cache            map[string]CachedOptimization
	maxCacheSize     int
	cacheHitRate     float64
	totalOptimizations int64
}

// CachedOptimization represents a cached optimization result
type CachedOptimization struct {
	OptimizedPrompt string
	Complexity      ComplexityLevel
	Timestamp       time.Time
	HitCount        int
}

// ProviderHealthMetrics represents health metrics for a provider (defined in provider_health.go)
type ProviderHealthMetrics struct {
	SuccessRate        float64   `json:"success_rate"`
	AverageLatency     float64   `json:"average_latency"`
	ErrorRate          float64   `json:"error_rate"`
	TotalRequests      int64     `json:"total_requests"`
	SuccessfulRequests int64     `json:"successful_requests"`
	FailedRequests     int64     `json:"failed_requests"`
	LastUpdated        time.Time `json:"last_updated"`
	Status             string    `json:"status"`
}

// ProviderHealthMonitor monitors provider health
type ProviderHealthMonitor struct {
	providers map[string]*ProviderHealthMetrics
}

// EnhancedProviderSelector provides advanced provider selection
type EnhancedProviderSelector struct {
	providers         []*Provider
	capabilityFilters map[string][]string
	healthCalculator  *HealthScoreCalculator
	costOptimizer     *CostBasedSelector
}

// CostBasedSelector provides cost-based provider selection
type CostBasedSelector struct {
	costThresholds map[ComplexityLevel]float64
	budgetLimits   map[string]float64
}

// NewCostBasedSelector creates a new cost-based selector
func NewCostBasedSelector(costThresholds map[ComplexityLevel]float64, budgetLimits map[string]float64) *CostBasedSelector {
	if costThresholds == nil {
		costThresholds = map[ComplexityLevel]float64{
			Low:      0.001,
			Medium:   0.01,
			High:     0.05,
			VeryHigh: 0.1,
		}
	}
	return &CostBasedSelector{
		costThresholds: costThresholds,
		budgetLimits:   budgetLimits,
	}
}

// sortProvidersByScore sorts providers by score (highest first)
func sortProvidersByScore(scores []ProviderScore) []ProviderScore {
	// Simple bubble sort for now
	for i := 0; i < len(scores)-1; i++ {
		for j := 0; j < len(scores)-i-1; j++ {
			if scores[j].Score < scores[j+1].Score {
				scores[j], scores[j+1] = scores[j+1], scores[j]
			}
		}
	}
	return scores
}