package enhanced

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// EnhancedProviderSelector implements cost-optimized provider selection with rate limit management
type EnhancedProviderSelector struct {
	logger *logrus.Logger
	
	// Provider management
	providers      []*Provider
	providersMutex sync.RWMutex
	providersFile  string
	
	// Cost-based selection components
	costBasedSelector *CostBasedSelector
	rateLimitManager  *RateLimitManager
	metricsStorage    *MetricsStorage
	healthCalculator  *HealthScoreCalculator
	
	// Selection configuration
	costWeight        float64
	performanceWeight float64
	latencyWeight     float64
	reliabilityWeight float64
	adaptationRate    float64
	
	// Performance tracking
	selectionHistory map[string][]SelectionRecord
	historyMutex     sync.RWMutex
}

// SelectionRecord tracks provider selection decisions with cost focus
type SelectionRecord struct {
	TaskComplexity   TaskComplexity `json:"task_complexity"`
	SelectedProvider string         `json:"selected_provider"`
	ActualCost       float64        `json:"actual_cost"`
	ActualLatency    float64        `json:"actual_latency"`
	QualityScore     float64        `json:"quality_score"`
	Success          bool           `json:"success"`
	CostSavings      float64        `json:"cost_savings"`
	Timestamp        time.Time      `json:"timestamp"`
}

// NewEnhancedProviderSelector creates a new cost-optimized provider selector
func NewEnhancedProviderSelector(logger *logrus.Logger, providersFile string) (*EnhancedProviderSelector, error) {
	// Initialize metrics storage
	metricsStorage, err := NewMetricsStorage("./metrics.db")
	if err != nil {
		logger.Warnf("Failed to initialize metrics storage: %v", err)
		metricsStorage = nil // Continue without persistent storage
	}
	
	// Initialize rate limit manager
	rateLimitManager := NewRateLimitManager(metricsStorage)
	
	// Initialize cost-based selector
	costBasedSelector := NewCostBasedSelector(metricsStorage, rateLimitManager)
	
	selector := &EnhancedProviderSelector{
		logger:            logger,
		providersFile:     providersFile,
		metricsStorage:    metricsStorage,
		rateLimitManager:  rateLimitManager,
		costBasedSelector: costBasedSelector,
		healthCalculator:  NewHealthScoreCalculator(),
		costWeight:        0.4,  // Increased for cost focus
		performanceWeight: 0.25, // Reduced
		latencyWeight:     0.2,  // Reduced
		reliabilityWeight: 0.15, // Slightly reduced
		adaptationRate:    0.05,
		selectionHistory:  make(map[string][]SelectionRecord),
	}
	
	// Load providers from CSV file
	if err := selector.loadProviders(); err != nil {
		return nil, fmt.Errorf("failed to load providers: %w", err)
	}
	
	return selector, nil
}

// loadProviders loads providers from CSV file with enhanced health metrics
func (e *EnhancedProviderSelector) loadProviders() error {
	file, err := os.Open(e.providersFile)
	if err != nil {
		return fmt.Errorf("failed to open providers file: %w", err)
	}
	defer file.Close()
	
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read CSV: %w", err)
	}
	
	if len(records) == 0 {
		return fmt.Errorf("empty providers file")
	}
	
	// Skip header row
	for i := 1; i < len(records); i++ {
		provider, err := e.parseProviderRecord(records[i])
		if err != nil {
			e.logger.Warnf("Failed to parse provider record %d: %v", i, err)
			continue
		}
		e.providers = append(e.providers, provider)
	}
	
	e.logger.Infof("Loaded %d providers from %s with cost optimization enabled", len(e.providers), e.providersFile)
	return nil
}

// parseProviderRecord parses a 6-column CSV record into a Provider struct with enhanced metrics
func (e *EnhancedProviderSelector) parseProviderRecord(record []string) (*Provider, error) {
	if len(record) < 6 {
		return nil, fmt.Errorf("insufficient columns: expected 6 columns (Name,Tier,Base_URL,APIKey,Models,Other), got %d", len(record))
	}
	
	provider := &Provider{
		Name:    strings.TrimSpace(record[0]), // Column 1: Name
		Tier:    ProviderTier(strings.ToLower(strings.TrimSpace(record[1]))), // Column 2: Tier
		BaseURL: strings.TrimSpace(record[2]), // Column 3: Base_URL
		APIKey:  strings.TrimSpace(record[3]), // Column 4: APIKey  
		Models:  strings.TrimSpace(record[4]), // Column 5: Model(s)
		Other:   strings.TrimSpace(record[5]), // Column 6: Other
		Metrics: ProviderMetrics{
			SuccessRate:      0.9, // Default values
			AverageLatency:   1000.0,
			QualityScore:     0.8,
			CostEfficiency:   0.7,
			LastUpdated:      time.Now(),
			RequestCount:     0,
			ErrorCount:       0,
			AverageCost:      0.001, // Default cost
			ReliabilityScore: 0.8,
		},
		HealthMetrics: &ProviderHealthMetrics{
			UsageHistory:         make(map[string][]UsageRecord),
			LastHealthCheck:      time.Now(),
			HealthScore:          0.8,
			CostEfficiencyScore:  0.7,
			RateLimitStatus:      nil,
		},
		Pricing: e.getDefaultPricing(ProviderTier(strings.ToLower(strings.TrimSpace(record[1])))),
	}
	
	return provider, nil
}

// getDefaultPricing returns default pricing based on provider tier
func (e *EnhancedProviderSelector) getDefaultPricing(tier ProviderTier) *ProviderPricing {
	switch tier {
	case OfficialTier:
		return &ProviderPricing{
			InputTokenCost:  0.00003, // $0.03 per 1K tokens
			OutputTokenCost: 0.00006, // $0.06 per 1K tokens
			Currency:        "USD",
			LastUpdated:     time.Now(),
		}
	case CommunityTier:
		return &ProviderPricing{
			InputTokenCost:  0.00001, // $0.01 per 1K tokens
			OutputTokenCost: 0.00002, // $0.02 per 1K tokens
			Currency:        "USD",
			LastUpdated:     time.Now(),
		}
	case UnofficialTier:
		return &ProviderPricing{
			InputTokenCost:  0.0, // Often free
			OutputTokenCost: 0.0,
			Currency:        "USD",
			LastUpdated:     time.Now(),
		}
	default:
		return &ProviderPricing{
			InputTokenCost:  0.00002, // $0.02 per 1K tokens default
			OutputTokenCost: 0.00004, // $0.04 per 1K tokens default
			Currency:        "USD",
			LastUpdated:     time.Now(),
		}
	}
}

// SelectOptimalProvider selects the optimal provider using cost-based optimization
func (e *EnhancedProviderSelector) SelectOptimalProvider(ctx context.Context, taskID string, complexity TaskComplexity, requirements map[string]interface{}) (ProviderAssignment, error) {
	e.logger.Infof("Selecting cost-optimal provider for task %s (complexity: %.2f)", taskID, complexity.Score)
	
	e.providersMutex.RLock()
	defer e.providersMutex.RUnlock()
	
	if len(e.providers) == 0 {
		return ProviderAssignment{}, fmt.Errorf("no providers available")
	}
	
	// Extract model from requirements if available
	model := "default"
	if modelReq, exists := requirements["model"]; exists {
		if modelStr, ok := modelReq.(string); ok {
			model = modelStr
		}
	}
	
	// Estimate tokens based on complexity
	estimatedTokens := e.estimateTokensFromComplexity(complexity)
	
	// Use cost-based selector for optimal provider selection
	selectedProvider, err := e.costBasedSelector.SelectOptimalProvider(
		e.providers,
		model,
		estimatedTokens,
		complexity,
	)
	
	if err != nil {
		return ProviderAssignment{}, fmt.Errorf("cost-based selection failed: %w", err)
	}
	
	if selectedProvider == nil {
		return ProviderAssignment{}, fmt.Errorf("no suitable provider found")
	}
	
	// Calculate estimated cost and time
	estimatedCost := e.calculateEstimatedCost(selectedProvider, estimatedTokens)
	estimatedTime := e.estimateTaskTime(selectedProvider, complexity)
	
	// Create alternatives list from other providers
	alternatives := e.generateAlternatives(selectedProvider, model, estimatedTokens, complexity)
	
	assignment := ProviderAssignment{
		TaskID:        taskID,
		ProviderID:    selectedProvider.Name,
		ProviderName:  selectedProvider.Name,
		ProviderTier:  selectedProvider.Tier,
		Model:         model,
		Tier:          string(selectedProvider.Tier),
		Confidence:    0.9, // High confidence with cost-based selection
		EstimatedCost: estimatedCost,
		EstimatedTime: int64(estimatedTime),
		Reasoning:     e.generateCostOptimizedReasoning(selectedProvider, complexity, estimatedCost),
		Alternatives:  alternatives,
		Metadata:      map[string]interface{}{
			"selection_method": "cost_optimized",
			"estimated_tokens": estimatedTokens,
			"cost_per_token":   selectedProvider.Pricing.InputTokenCost,
		},
	}
	
	e.logger.Infof("Selected cost-optimal provider %s for task %s (estimated cost: $%.6f)", 
		selectedProvider.Name, taskID, estimatedCost)
	
	return assignment, nil
}

// estimateTokensFromComplexity estimates token usage based on task complexity
func (e *EnhancedProviderSelector) estimateTokensFromComplexity(complexity TaskComplexity) int64 {
	baseTokens := int64(100)
	
	switch complexity.Overall {
	case VeryHigh:
		return baseTokens * 20 // 2000 tokens
	case High:
		return baseTokens * 10 // 1000 tokens
	case Medium:
		return baseTokens * 5  // 500 tokens
	case Low:
		return baseTokens * 2  // 200 tokens
	default:
		return baseTokens * 3  // 300 tokens
	}
}

// calculateEstimatedCost calculates estimated cost for a provider
func (e *EnhancedProviderSelector) calculateEstimatedCost(provider *Provider, estimatedTokens int64) float64 {
	if provider.Pricing == nil {
		return 0.001 * float64(estimatedTokens) // Default fallback
	}
	
	// Assume 70% input tokens, 30% output tokens
	inputTokens := float64(estimatedTokens) * 0.7
	outputTokens := float64(estimatedTokens) * 0.3
	
	return inputTokens*provider.Pricing.InputTokenCost + outputTokens*provider.Pricing.OutputTokenCost
}

// estimateTaskTime estimates task completion time
func (e *EnhancedProviderSelector) estimateTaskTime(provider *Provider, complexity TaskComplexity) int {
	baseTime := int(provider.Metrics.AverageLatency) // Base latency in ms
	
	// Adjust based on complexity
	complexityMultiplier := 1.0
	switch complexity.Overall {
	case VeryHigh:
		complexityMultiplier = 2.0
	case High:
		complexityMultiplier = 1.5
	case Medium:
		complexityMultiplier = 1.2
	case Low:
		complexityMultiplier = 1.0
	}
	
	return int(float64(baseTime) * complexityMultiplier)
}

// generateAlternatives creates alternative provider options
func (e *EnhancedProviderSelector) generateAlternatives(selectedProvider *Provider, model string, estimatedTokens int64, complexity TaskComplexity) []AlternativeProvider {
	var alternatives []AlternativeProvider
	
	for _, provider := range e.providers {
		if provider.Name == selectedProvider.Name {
			continue // Skip the selected provider
		}
		
		// Check if provider can handle the request
		if e.rateLimitManager != nil {
			if canHandle, _ := e.rateLimitManager.CanHandleRequest(provider.Name, model, estimatedTokens); !canHandle {
				continue
			}
		}
		
		estimatedCost := e.calculateEstimatedCost(provider, estimatedTokens)
		estimatedTime := e.estimateTaskTime(provider, complexity)
		
		alternative := AlternativeProvider{
			ProviderID:    provider.Name,
			ProviderName:  provider.Name,
			Model:         model,
			Confidence:    0.8, // Slightly lower confidence for alternatives
			EstimatedCost: estimatedCost,
			EstimatedTime: int64(estimatedTime),
			Reasoning:     fmt.Sprintf("Alternative %s provider with estimated cost $%.6f", provider.Tier, estimatedCost),
		}
		
		alternatives = append(alternatives, alternative)
		
		if len(alternatives) >= 3 {
			break // Limit to top 3 alternatives
		}
	}
	
	return alternatives
}

// generateCostOptimizedReasoning generates reasoning for cost-optimized selection
func (e *EnhancedProviderSelector) generateCostOptimizedReasoning(provider *Provider, complexity TaskComplexity, estimatedCost float64) string {
	return fmt.Sprintf("Selected %s (tier: %s) using cost-optimization algorithm. "+
		"Estimated cost: $%.6f for %s complexity task. "+
		"Provider offers optimal balance of cost efficiency (%.2f), reliability (%.2f), and capability alignment.",
		provider.Name, provider.Tier, estimatedCost, complexity.Overall,
		provider.Metrics.CostEfficiency, provider.Metrics.ReliabilityScore)
}

// RecordProviderPerformance records actual performance for learning
func (e *EnhancedProviderSelector) RecordProviderPerformance(taskID string, provider *Provider, actualCost, actualLatency, qualityScore float64, success bool) {
	if e.metricsStorage != nil {
		// Record metrics in persistent storage
		tokensUsed := int64(actualCost / provider.Pricing.InputTokenCost) // Rough estimate
		e.metricsStorage.RecordProviderMetrics(
			provider.Name, "default", 1, 0, tokensUsed,
			actualLatency, actualCost, false)
	}
	
	// Update provider metrics
	e.providersMutex.Lock()
	defer e.providersMutex.Unlock()
	
	provider.Metrics.RequestCount++
	if success {
		provider.Metrics.SuccessfulRequests++
	} else {
		provider.Metrics.ErrorCount++
	}
	
	// Update running averages
	provider.Metrics.SuccessRate = float64(provider.Metrics.SuccessfulRequests) / float64(provider.Metrics.RequestCount)
	provider.Metrics.AverageLatency = (provider.Metrics.AverageLatency + actualLatency) / 2
	provider.Metrics.AverageCost = (provider.Metrics.AverageCost + actualCost) / 2
	provider.Metrics.QualityScore = (provider.Metrics.QualityScore + qualityScore) / 2
	provider.Metrics.LastUpdated = time.Now()
	
	e.logger.Infof("Updated performance metrics for provider %s: success_rate=%.2f, avg_cost=$%.6f", 
		provider.Name, provider.Metrics.SuccessRate, provider.Metrics.AverageCost)
}

// GetProviders returns all available providers
func (e *EnhancedProviderSelector) GetProviders() []*Provider {
	e.providersMutex.RLock()
	defer e.providersMutex.RUnlock()
	
	// Return a copy to prevent external modification
	providers := make([]*Provider, len(e.providers))
	copy(providers, e.providers)
	return providers
}

// GetCostSavingsReport returns cost optimization analytics
func (e *EnhancedProviderSelector) GetCostSavingsReport(days int) (*CostSavingsReport, error) {
	if e.metricsStorage == nil {
		return nil, fmt.Errorf("metrics storage not available")
	}
	
	return e.metricsStorage.GetCostSavingsReport(days)
}

// Close closes database connections and cleans up resources
func (e *EnhancedProviderSelector) Close() error {
	if e.metricsStorage != nil {
		return e.metricsStorage.Close()
	}
	return nil
}