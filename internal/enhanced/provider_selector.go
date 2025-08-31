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

// AdaptiveProviderSelector implements intelligent provider selection
type AdaptiveProviderSelector struct {
	logger *logrus.Logger
	
	// Provider management
	providers     []*Provider
	providersMutex sync.RWMutex
	providersFile  string
	
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

// NewAdaptiveProviderSelector creates a new adaptive provider selector
func NewAdaptiveProviderSelector(logger *logrus.Logger, providersFile string) (*AdaptiveProviderSelector, error) {
	selector := &AdaptiveProviderSelector{
		logger:            logger,
		providersFile:     providersFile,
		costWeight:        0.4,
		performanceWeight: 0.3,
		latencyWeight:     0.2,
		reliabilityWeight: 0.1,
		adaptationRate:    0.05,
		selectionHistory:  make(map[string][]SelectionRecord),
	}
	
	// Load providers from CSV file
	if err := selector.loadProviders(); err != nil {
		return nil, fmt.Errorf("failed to load providers: %w", err)
	}
	
	return selector, nil
}

// loadProviders loads providers from CSV file
func (a *AdaptiveProviderSelector) loadProviders() error {
	file, err := os.Open(a.providersFile)
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
		provider, err := a.parseProviderRecord(records[i])
		if err != nil {
			a.logger.Warnf("Failed to parse provider record %d: %v", i, err)
			continue
		}
		a.providers = append(a.providers, provider)
	}
	
	a.logger.Infof("Loaded %d providers from %s", len(a.providers), a.providersFile)
	return nil
}

// parseProviderRecord parses a 6-column CSV record into a Provider struct
func (a *AdaptiveProviderSelector) parseProviderRecord(record []string) (*Provider, error) {
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
	}
	
	return provider, nil
}

// SelectOptimalProvider selects the optimal provider for a task
func (a *AdaptiveProviderSelector) SelectOptimalProvider(ctx context.Context, taskID string, complexity TaskComplexity, requirements map[string]interface{}) (ProviderAssignment, error) {
	a.logger.Infof("Selecting optimal provider for task %s (complexity: %.2f)", taskID, complexity.Score)
	
	a.providersMutex.RLock()
	defer a.providersMutex.RUnlock()
	
	if len(a.providers) == 0 {
		return ProviderAssignment{}, fmt.Errorf("no providers available")
	}
	
	// Calculate scores for all providers
	scores := make([]ProviderScore, 0, len(a.providers))
	for _, provider := range a.providers {
		score := a.calculateProviderScore(provider, complexity, requirements)
		scores = append(scores, ProviderScore{
			Provider: provider,
			Score:    score,
		})
	}
	
	// Sort by score (highest first)
	for i := 0; i < len(scores)-1; i++ {
		for j := i + 1; j < len(scores); j++ {
			if scores[i].Score < scores[j].Score {
				scores[i], scores[j] = scores[j], scores[i]
			}
		}
	}
	
	// Select the best provider
	bestProvider := scores[0].Provider
	
	// Create alternatives list
	alternatives := make([]AlternativeProvider, 0, min(3, len(scores)-1))
	for i := 1; i < min(4, len(scores)); i++ {
		alternatives = append(alternatives, AlternativeProvider{
			ProviderID:    scores[i].Provider.Name, // Use Name as ID
			ProviderName:  scores[i].Provider.Name,
			Confidence:    scores[i].Score,
			EstimatedCost: a.estimateTaskCost(scores[i].Provider, complexity),
			Reasoning:     fmt.Sprintf("Alternative option with score %.2f", scores[i].Score),
		})
	}
	
	assignment := ProviderAssignment{
		TaskID:        taskID,
		ProviderID:    bestProvider.Name, // Use Name as ID
		ProviderName:  bestProvider.Name,
		ProviderTier:  bestProvider.Tier,
		Confidence:    scores[0].Score,
		EstimatedCost: a.estimateTaskCost(bestProvider, complexity),
		EstimatedTime: int64(a.estimateTaskTime(bestProvider, complexity)),
		Reasoning:     a.generateSelectionReasoning(bestProvider, complexity, scores[0].Score),
		Alternatives:  alternatives,
		Metadata:      make(map[string]interface{}),
	}
	
	a.logger.Infof("Selected provider %s for task %s with confidence %.2f", 
		bestProvider.Name, taskID, scores[0].Score)
	
	return assignment, nil
}

// UpdateProviderMetrics updates provider performance metrics
func (a *AdaptiveProviderSelector) UpdateProviderMetrics(ctx context.Context, providerID string, success bool, cost, latency, quality float64) {
	a.providersMutex.Lock()
	defer a.providersMutex.Unlock()
	
	for _, provider := range a.providers {
		if provider.Name == providerID {
			provider.Metrics.TotalRequests++
			if success {
				provider.Metrics.SuccessfulRequests++
			} else {
				provider.Metrics.FailedRequests++
			}
			
			// Update running averages
			if provider.Metrics.TotalRequests > 0 {
				provider.Metrics.SuccessRate = float64(provider.Metrics.SuccessfulRequests) / float64(provider.Metrics.TotalRequests)
			}
			provider.Metrics.AverageLatency = (provider.Metrics.AverageLatency + latency) / 2
			provider.Metrics.AverageCost = (provider.Metrics.AverageCost + cost) / 2
			provider.Metrics.QualityScore = (provider.Metrics.QualityScore + quality) / 2
			provider.Metrics.LastUpdated = time.Now()
			
			a.logger.Infof("Updated metrics for provider %s: success_rate=%.2f, avg_cost=%.6f", 
				provider.Name, provider.Metrics.SuccessRate, provider.Metrics.AverageCost)
			break
		}
	}
}

// ProviderScore represents a provider with its calculated score
type ProviderScore struct {
	Provider *Provider
	Score    float64
}

// calculateProviderScore calculates a score for a provider based on task requirements
func (a *AdaptiveProviderSelector) calculateProviderScore(provider *Provider, complexity TaskComplexity, requirements map[string]interface{}) float64 {
	// Base score from provider tier
	tierScore := a.getTierScore(provider.Tier)
	
	// Cost efficiency score (assume default cost per token)
	defaultCostPerToken := 0.001
	costScore := 1.0 - (defaultCostPerToken / 0.001) // Normalize against typical max cost
	if costScore < 0 {
		costScore = 0
	}
	
	// Performance score from metrics
	performanceScore := provider.Metrics.QualityScore
	
	// Latency score (lower latency is better)
	latencyScore := 1.0 - (provider.Metrics.AverageLatency / 5000.0) // Normalize against 5s max
	if latencyScore < 0 {
		latencyScore = 0
	}
	
	// Reliability score
	reliabilityScore := provider.Metrics.ReliabilityScore
	
	// Complexity alignment score
	complexityScore := a.getComplexityAlignmentScore(provider, complexity)
	
	// Weighted final score
	finalScore := a.costWeight*costScore +
		a.performanceWeight*performanceScore +
		a.latencyWeight*latencyScore +
		a.reliabilityWeight*reliabilityScore +
		0.2*tierScore +
		0.1*complexityScore
	
	return finalScore
}

// getTierScore returns a score based on provider tier
func (a *AdaptiveProviderSelector) getTierScore(tier ProviderTier) float64 {
	switch tier {
	case OfficialTier:
		return 1.0
	case CommunityTier:
		return 0.7
	case UnofficialTier:
		return 0.4
	default:
		return 0.5
	}
}

// getComplexityAlignmentScore returns a score based on how well the provider aligns with task complexity
func (a *AdaptiveProviderSelector) getComplexityAlignmentScore(provider *Provider, complexity TaskComplexity) float64 {
	// Higher complexity tasks should prefer higher-tier providers
	if complexity.Overall >= High && provider.Tier == OfficialTier {
		return 1.0
	} else if complexity.Overall >= Medium && provider.Tier == CommunityTier {
		return 0.8
	} else if complexity.Overall <= Medium && provider.Tier == UnofficialTier {
		return 0.6
	}
	return 0.5
}

// estimateTaskCost estimates the cost for a task with a given provider
func (a *AdaptiveProviderSelector) estimateTaskCost(provider *Provider, complexity TaskComplexity) float64 {
	// Estimate tokens based on complexity
	estimatedTokens := 100.0 // Base tokens
	
	switch complexity.Overall {
	case VeryHigh:
		estimatedTokens = 2000.0
	case High:
		estimatedTokens = 1000.0
	case Medium:
		estimatedTokens = 500.0
	case Low:
		estimatedTokens = 200.0
	}
	
	defaultCostPerToken := 0.001
	return defaultCostPerToken * estimatedTokens
}

// estimateTaskTime estimates the time for a task with a given provider
func (a *AdaptiveProviderSelector) estimateTaskTime(provider *Provider, complexity TaskComplexity) int {
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

// generateSelectionReasoning generates reasoning for provider selection
func (a *AdaptiveProviderSelector) generateSelectionReasoning(provider *Provider, complexity TaskComplexity, score float64) string {
	return fmt.Sprintf("Selected %s (tier: %s) based on optimal balance of cost, "+
		"quality score (%.2f), and complexity alignment for %s complexity task. Overall score: %.2f",
		provider.Name, provider.Tier, provider.Metrics.QualityScore, complexity.Overall, score)
}

// GetProviders returns all available providers
func (a *AdaptiveProviderSelector) GetProviders() []*Provider {
	a.providersMutex.RLock()
	defer a.providersMutex.RUnlock()
	
	// Return a copy to prevent external modification
	providers := make([]*Provider, len(a.providers))
	copy(providers, a.providers)
	return providers
}