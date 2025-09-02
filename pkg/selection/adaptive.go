package selection

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/ThatsRight-ItsTJ/Your-PaL-MoE/pkg/analysis"
	"github.com/ThatsRight-ItsTJ/Your-PaL-MoE/pkg/config"
)

// ProviderScore represents the scoring result for a provider
type ProviderScore struct {
	ProviderID       string  `json:"provider_id"`
	TotalScore       float64 `json:"total_score"`
	CostScore        float64 `json:"cost_score"`
	QualityScore     float64 `json:"quality_score"`
	LatencyScore     float64 `json:"latency_score"`
	ReliabilityScore float64 `json:"reliability_score"`
	Reasoning        string  `json:"reasoning"`
}

// SelectionWeights defines the importance of different factors
type SelectionWeights struct {
	Cost        float64 `json:"cost"`
	Quality     float64 `json:"quality"`
	Latency     float64 `json:"latency"`
	Reliability float64 `json:"reliability"`
}

// ProviderMetrics tracks provider performance over time
type ProviderMetrics struct {
	AverageLatency     time.Duration `json:"average_latency"`
	SuccessRate        float64       `json:"success_rate"`
	QualityScore       float64       `json:"quality_score"`
	CostEfficiency     float64       `json:"cost_efficiency"`
	TotalRequests      int           `json:"total_requests"`
	SuccessfulRequests int           `json:"successful_requests"`
	LastUpdated        time.Time     `json:"last_updated"`
}

// AdaptiveSelector implements intelligent provider selection
type AdaptiveSelector struct {
	csvProviders    []config.CSVProvider
	enhancedConfigs map[string]*config.ProviderConfig
	performanceData map[string]*ProviderMetrics
	weights         SelectionWeights
	mutex           sync.RWMutex
	yamlBuilder     *config.YAMLBuilder
}

// NewAdaptiveSelector creates a new adaptive provider selector
func NewAdaptiveSelector(csvPath string) (*AdaptiveSelector, error) {
	yamlBuilder := config.NewYAMLBuilder()
	yamlBuilder.SetCSVPath(csvPath)
	yamlBuilder.SetConfigDir("./configs")
	
	selector := &AdaptiveSelector{
		enhancedConfigs: make(map[string]*config.ProviderConfig),
		performanceData: make(map[string]*ProviderMetrics),
		yamlBuilder:     yamlBuilder,
		weights: SelectionWeights{
			Cost:        0.25,
			Quality:     0.40,
			Latency:     0.20,
			Reliability: 0.15,
		},
	}

	// Load CSV providers
	providers, err := yamlBuilder.ReadCSV()
	if err != nil {
		return nil, fmt.Errorf("failed to load CSV providers: %w", err)
	}
	selector.csvProviders = providers

	// Try to load enhanced configurations
	if err := yamlBuilder.BuildFromCSV(); err == nil {
		selector.loadEnhancedConfigs("./configs")
	}

	return selector, nil
}

// SelectProvider selects the best provider for a given task complexity
func (as *AdaptiveSelector) SelectProvider(complexity analysis.TaskComplexity, constraints map[string]interface{}) (ProviderScore, error) {
	as.mutex.RLock()
	defer as.mutex.RUnlock()

	// Try enhanced selection first if available
	if len(as.enhancedConfigs) > 0 {
		return as.selectFromEnhanced(complexity, constraints)
	}

	// Fallback to CSV-based selection
	return as.selectFromCSV(complexity, constraints)
}

// selectFromEnhanced performs selection using enhanced configurations
func (as *AdaptiveSelector) selectFromEnhanced(complexity analysis.TaskComplexity, constraints map[string]interface{}) (ProviderScore, error) {
	var candidates []ProviderScore

	for _, provider := range as.enhancedConfigs {
		score := as.calculateProviderScore(*provider, complexity, constraints)
		candidates = append(candidates, score)
	}

	if len(candidates) == 0 {
		return ProviderScore{}, fmt.Errorf("no suitable providers found")
	}

	// Sort by total score (descending)
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].TotalScore > candidates[j].TotalScore
	})

	best := candidates[0]
	best.Reasoning = as.generateSelectionReasoning(best, complexity, constraints)

	return best, nil
}

// selectFromCSV performs fallback selection using only CSV data
func (as *AdaptiveSelector) selectFromCSV(complexity analysis.TaskComplexity, constraints map[string]interface{}) (ProviderScore, error) {
	var candidates []ProviderScore

	for _, provider := range as.csvProviders {
		score := as.calculateCSVProviderScore(provider, complexity, constraints)
		candidates = append(candidates, score)
	}

	if len(candidates) == 0 {
		return ProviderScore{}, fmt.Errorf("no suitable providers found")
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].TotalScore > candidates[j].TotalScore
	})

	best := candidates[0]
	best.Reasoning = as.generateCSVSelectionReasoning(best, complexity)

	return best, nil
}

// calculateProviderScore calculates comprehensive provider score
func (as *AdaptiveSelector) calculateProviderScore(provider config.ProviderConfig, complexity analysis.TaskComplexity, constraints map[string]interface{}) ProviderScore {
	qualityScore := as.calculateQualityScore(provider, complexity)
	costScore := as.calculateCostScore(provider, constraints)
	latencyScore := as.calculateLatencyScore(provider.ID)
	reliabilityScore := as.calculateReliabilityScore(provider.ID)

	totalScore := (qualityScore * as.weights.Quality) +
		(costScore * as.weights.Cost) +
		(latencyScore * as.weights.Latency) +
		(reliabilityScore * as.weights.Reliability)

	return ProviderScore{
		ProviderID:       provider.ID,
		TotalScore:       totalScore,
		CostScore:        costScore,
		QualityScore:     qualityScore,
		LatencyScore:     latencyScore,
		ReliabilityScore: reliabilityScore,
	}
}

// calculateCSVProviderScore calculates score using only CSV data
func (as *AdaptiveSelector) calculateCSVProviderScore(provider config.CSVProvider, complexity analysis.TaskComplexity, constraints map[string]interface{}) ProviderScore {
	// Simple tier-based quality scoring
	qualityScore := as.getTierQualityScore(provider.Tier)

	// Adjust based on complexity requirements
	if complexity.Score > 0.7 && provider.Tier != "official" {
		qualityScore *= 0.7 // Penalty for non-official providers on complex tasks
	}

	// Simple cost scoring based on tier
	costScore := as.getTierCostScore(provider.Tier)

	// Default latency and reliability scores
	latencyScore := 0.7
	reliabilityScore := 0.8

	// Apply historical data if available
	providerID := as.generateProviderID(provider.Name)
	if metrics, exists := as.performanceData[providerID]; exists {
		latencyScore = as.calculateLatencyScoreFromMetrics(metrics)
		reliabilityScore = metrics.SuccessRate
	}

	totalScore := (qualityScore * as.weights.Quality) +
		(costScore * as.weights.Cost) +
		(latencyScore * as.weights.Latency) +
		(reliabilityScore * as.weights.Reliability)

	return ProviderScore{
		ProviderID:       providerID,
		TotalScore:       totalScore,
		CostScore:        costScore,
		QualityScore:     qualityScore,
		LatencyScore:     latencyScore,
		ReliabilityScore: reliabilityScore,
	}
}

// calculateQualityScore matches provider capabilities to task requirements
func (as *AdaptiveSelector) calculateQualityScore(provider config.ProviderConfig, complexity analysis.TaskComplexity) float64 {
	capabilities := provider.Capabilities

	// Match provider capabilities to task requirements
	reasoningMatch := math.Min(float64(capabilities.Reasoning)/10.0, complexity.Reasoning/3.0)
	knowledgeMatch := math.Min(float64(capabilities.Knowledge)/10.0, complexity.Knowledge/3.0)
	computationMatch := math.Min(float64(capabilities.Computation)/10.0, complexity.Computation/3.0)
	coordinationMatch := math.Min(float64(capabilities.Coordination)/10.0, complexity.Coordination/3.0)

	// Weight the matches based on task's dominant dimension
	dominantDim := complexity.GetDominantDimension()
	score := (reasoningMatch + knowledgeMatch + computationMatch + coordinationMatch) / 4.0

	// Boost score if provider excels in the dominant dimension
	switch dominantDim {
	case "reasoning":
		if capabilities.Reasoning >= 8 {
			score *= 1.2
		}
	case "knowledge":
		if capabilities.Knowledge >= 8 {
			score *= 1.2
		}
	case "computation":
		if capabilities.Computation >= 8 {
			score *= 1.2
		}
	case "coordination":
		if capabilities.Coordination >= 8 {
			score *= 1.2
		}
	}

	return math.Min(score, 1.0)
}

// calculateCostScore evaluates cost efficiency
func (as *AdaptiveSelector) calculateCostScore(provider config.ProviderConfig, constraints map[string]interface{}) float64 {
	baseCost := provider.CostTracking.CostPerToken

	// Check budget constraints
	if maxCost, exists := constraints["max_cost_per_token"]; exists {
		if cost, ok := maxCost.(float64); ok {
			if baseCost > cost {
				return 0.0 // Provider exceeds budget
			}
		}
	}

	// Inverse scoring - lower cost = higher score
	maxPossibleCost := 0.0001 // $0.10 per 1k tokens
	score := 1.0 - (baseCost / maxPossibleCost)

	return math.Max(0.0, math.Min(1.0, score))
}

// calculateLatencyScore evaluates response time performance
func (as *AdaptiveSelector) calculateLatencyScore(providerID string) float64 {
	metrics, exists := as.performanceData[providerID]
	if !exists {
		return 0.7 // Default score for unknown providers
	}

	return as.calculateLatencyScoreFromMetrics(metrics)
}

// calculateLatencyScoreFromMetrics converts latency metrics to score
func (as *AdaptiveSelector) calculateLatencyScoreFromMetrics(metrics *ProviderMetrics) float64 {
	targetLatency := 500 * time.Millisecond
	if metrics.AverageLatency <= targetLatency {
		return 1.0
	}

	// Linear decay up to 5 seconds
	maxLatency := 5 * time.Second
	if metrics.AverageLatency >= maxLatency {
		return 0.1
	}

	ratio := float64(metrics.AverageLatency-targetLatency) / float64(maxLatency-targetLatency)
	return 1.0 - (ratio * 0.9)
}

// calculateReliabilityScore evaluates success rate
func (as *AdaptiveSelector) calculateReliabilityScore(providerID string) float64 {
	metrics, exists := as.performanceData[providerID]
	if !exists {
		return 0.8 // Default high reliability for untested providers
	}

	return metrics.SuccessRate
}

// getTierQualityScore returns quality score based on provider tier
func (as *AdaptiveSelector) getTierQualityScore(tier string) float64 {
	switch tier {
	case "official":
		return 0.9
	case "community":
		return 0.7
	case "unofficial":
		return 0.5
	default:
		return 0.6
	}
}

// getTierCostScore returns cost score based on provider tier
func (as *AdaptiveSelector) getTierCostScore(tier string) float64 {
	switch tier {
	case "official":
		return 0.3 // Higher cost
	case "community":
		return 0.7 // Medium cost
	case "unofficial":
		return 1.0 // Lower/free cost
	default:
		return 0.5
	}
}

// UpdateProviderMetrics updates performance metrics for a provider
func (as *AdaptiveSelector) UpdateProviderMetrics(providerID string, latency time.Duration, success bool, quality float64) {
	as.mutex.Lock()
	defer as.mutex.Unlock()

	metrics, exists := as.performanceData[providerID]
	if !exists {
		metrics = &ProviderMetrics{
			AverageLatency:     latency,
			SuccessRate:        1.0,
			QualityScore:       quality,
			TotalRequests:      0,
			SuccessfulRequests: 0,
			LastUpdated:        time.Now(),
		}
		as.performanceData[providerID] = metrics
	}

	// Update metrics using exponential moving average
	alpha := 0.1 // Learning rate
	metrics.AverageLatency = time.Duration(float64(metrics.AverageLatency)*(1-alpha) + float64(latency)*alpha)

	metrics.TotalRequests++
	if success {
		metrics.SuccessfulRequests++
	}

	metrics.SuccessRate = float64(metrics.SuccessfulRequests) / float64(metrics.TotalRequests)
	metrics.QualityScore = metrics.QualityScore*(1-alpha) + quality*alpha
	metrics.LastUpdated = time.Now()
}

// generateSelectionReasoning creates human-readable reasoning for selection
func (as *AdaptiveSelector) generateSelectionReasoning(score ProviderScore, complexity analysis.TaskComplexity, constraints map[string]interface{}) string {
	reasons := []string{}

	if score.QualityScore > 0.8 {
		reasons = append(reasons, "high capability match for task complexity")
	}
	if score.CostScore > 0.7 {
		reasons = append(reasons, "cost-effective option")
	}
	if score.LatencyScore > 0.8 {
		reasons = append(reasons, "fast response times")
	}
	if score.ReliabilityScore > 0.9 {
		reasons = append(reasons, "excellent reliability")
	}

	complexityDesc := complexity.GetComplexityDescription()
	dominantDim := complexity.GetDominantDimension()

	reasoning := fmt.Sprintf("Selected for %s task requiring strong %s capabilities. Key factors: %s",
		complexityDesc, dominantDim, joinReasons(reasons))

	return reasoning
}

// generateCSVSelectionReasoning creates reasoning for CSV-based selection
func (as *AdaptiveSelector) generateCSVSelectionReasoning(score ProviderScore, complexity analysis.TaskComplexity) string {
	complexityDesc := complexity.GetComplexityDescription()
	return fmt.Sprintf("Selected based on %s and provider tier scoring (total score: %.2f)",
		complexityDesc, score.TotalScore)
}

// Helper functions

func (as *AdaptiveSelector) loadEnhancedConfigs(configDir string) error {
	// Implementation would load YAML files from configDir
	// For now, return nil (configs loaded via YAML builder)
	return nil
}

func (as *AdaptiveSelector) generateProviderID(name string) string {
	return strings.ToLower(strings.ReplaceAll(name, " ", "_"))
}

func joinReasons(reasons []string) string {
	if len(reasons) == 0 {
		return "general suitability"
	}
	if len(reasons) == 1 {
		return reasons[0]
	}
	if len(reasons) == 2 {
		return reasons[0] + " and " + reasons[1]
	}

	last := reasons[len(reasons)-1]
	others := strings.Join(reasons[:len(reasons)-1], ", ")
	return others + ", and " + last
}

// GetProviderMetrics returns current metrics for all providers
func (as *AdaptiveSelector) GetProviderMetrics() map[string]*ProviderMetrics {
	as.mutex.RLock()
	defer as.mutex.RUnlock()

	result := make(map[string]*ProviderMetrics)
	for k, v := range as.performanceData {
		result[k] = v
	}
	return result
}

// SetWeights updates the selection weights
func (as *AdaptiveSelector) SetWeights(weights SelectionWeights) {
	as.mutex.Lock()
	defer as.mutex.Unlock()
	as.weights = weights
}