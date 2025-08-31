package enhanced

import (
	"fmt"
	"math"
	"sort"
	"time"
)

// CostBasedSelector implements cost-optimized provider selection
type CostBasedSelector struct {
	metricsStorage       *MetricsStorage
	rateLimitManager     *RateLimitManager
	healthCalculator     *HealthScoreCalculator
	costThreshold        float64 // Maximum acceptable cost per token
	reliabilityThreshold float64 // Minimum acceptable reliability
}

// NewCostBasedSelector creates a new cost-focused selector
func NewCostBasedSelector(storage *MetricsStorage, rateLimitManager *RateLimitManager) *CostBasedSelector {
	return &CostBasedSelector{
		metricsStorage:       storage,
		rateLimitManager:     rateLimitManager,
		healthCalculator:     NewHealthScoreCalculator(),
		costThreshold:        0.00005, // $0.05 per 1K tokens maximum
		reliabilityThreshold: 0.95,    // 95% minimum success rate
	}
}

// ProviderCostScore represents provider selection criteria
type ProviderCostScore struct {
	Provider            *Provider
	CostPerToken        float64
	ReliabilityScore    float64
	RateLimitAvailable  bool
	HealthScore         float64
	TotalScore          float64
	Reason              string
}

// SelectOptimalProvider implements cost-first provider selection
func (c *CostBasedSelector) SelectOptimalProvider(
	providers []*Provider,
	model string,
	estimatedTokens int64,
	taskComplexity TaskComplexity,
) (*Provider, error) {
	if len(providers) == 0 {
		return nil, fmt.Errorf("no providers available")
	}
	
	var scores []ProviderCostScore
	
	// Calculate scores for each provider
	for _, provider := range providers {
		score := c.calculateProviderCostScore(provider, model, estimatedTokens, taskComplexity)
		scores = append(scores, score)
	}
	
	// Filter out providers that don't meet minimum requirements
	validScores := c.filterValidProviders(scores)
	if len(validScores) == 0 {
		// If no providers meet strict criteria, use best available with warning
		return c.selectBestAvailable(scores), nil
	}
	
	// Sort by total score (cost-optimized)
	sort.Slice(validScores, func(i, j int) bool {
		return validScores[i].TotalScore > validScores[j].TotalScore
	})
	
	return validScores[0].Provider, nil
}

// calculateProviderCostScore computes cost-focused scoring
func (c *CostBasedSelector) calculateProviderCostScore(
	provider *Provider,
	model string,
	estimatedTokens int64,
	complexity TaskComplexity,
) ProviderCostScore {
	score := ProviderCostScore{
		Provider: provider,
	}
	
	// Get cost analysis from historical data
	var analysis *CostAnalysis
	var err error
	if c.metricsStorage != nil {
		analysis, err = c.metricsStorage.GetCostAnalysis(provider.Name, model, 24*time.Hour)
	}
	
	if err != nil || analysis == nil || analysis.TotalRecords == 0 {
		// Use provider's default pricing if no historical data
		score.CostPerToken = c.getDefaultCostPerToken(provider)
		score.ReliabilityScore = 0.9 // Conservative default
		score.Reason = "no_historical_data"
	} else {
		score.CostPerToken = analysis.AvgCostPerToken
		score.ReliabilityScore = analysis.AvgSuccessRate
	}
	
	// Check rate limit availability
	score.RateLimitAvailable = c.checkRateLimitAvailability(provider, model, estimatedTokens)
	
	// Calculate health score
	score.HealthScore = c.healthCalculator.CalculateHealthScore(provider, model, 0)
	
	// Cost score (lower cost = higher score)
	costScore := c.calculateCostScore(score.CostPerToken, estimatedTokens)
	
	// Reliability score
	reliabilityScore := score.ReliabilityScore
	
	// Rate limit score
	rateLimitScore := 0.0
	if score.RateLimitAvailable {
		rateLimitScore = 1.0
	}
	
	// Complexity alignment score
	complexityScore := c.getComplexityAlignmentScore(provider, complexity)
	
	// Weighted total score (cost-focused)
	score.TotalScore = 0.4*costScore +           // Primary: cost optimization
					  0.25*reliabilityScore +    // Secondary: reliability
					  0.2*rateLimitScore +       // Important: availability
					  0.1*score.HealthScore +    // Minor: overall health
					  0.05*complexityScore       // Minor: complexity match
	
	return score
}

// calculateCostScore converts cost per token to a score (0-1, higher is better)
func (c *CostBasedSelector) calculateCostScore(costPerToken float64, estimatedTokens int64) float64 {
	if costPerToken <= 0 {
		return 1.0 // Free providers get maximum cost score
	}
	
	// Use exponential decay for cost scoring to heavily favor cheaper options
	baselineCost := 0.00003 // GPT-4 baseline: $0.03 per 1K tokens
	
	// Exponential preference for lower costs
	costRatio := costPerToken / baselineCost
	costScore := math.Exp(-costRatio) // Exponential decay
	
	return math.Max(0, math.Min(1, costScore))
}

// checkRateLimitAvailability verifies if provider can handle the request
func (c *CostBasedSelector) checkRateLimitAvailability(
	provider *Provider,
	model string,
	estimatedTokens int64,
) bool {
	if c.rateLimitManager == nil {
		return true // Assume available if no rate limit manager
	}
	
	canHandle, _ := c.rateLimitManager.CanHandleRequest(provider.Name, model, estimatedTokens)
	return canHandle
}

// filterValidProviders removes providers that don't meet minimum standards
func (c *CostBasedSelector) filterValidProviders(scores []ProviderCostScore) []ProviderCostScore {
	var valid []ProviderCostScore
	
	for _, score := range scores {
		// Must meet reliability threshold
		if score.ReliabilityScore < c.reliabilityThreshold {
			continue
		}
		
		// Must be within cost threshold
		if score.CostPerToken > c.costThreshold {
			continue
		}
		
		// Must have rate limit availability
		if !score.RateLimitAvailable {
			continue
		}
		
		valid = append(valid, score)
	}
	
	return valid
}

// selectBestAvailable picks the best option when no providers meet strict criteria
func (c *CostBasedSelector) selectBestAvailable(scores []ProviderCostScore) *Provider {
	if len(scores) == 0 {
		return nil
	}
	
	// Sort by total score
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].TotalScore > scores[j].TotalScore
	})
	
	return scores[0].Provider
}

// getComplexityAlignmentScore evaluates if provider is suitable for task complexity
func (c *CostBasedSelector) getComplexityAlignmentScore(provider *Provider, complexity TaskComplexity) float64 {
	// For cost optimization, prefer using cheaper models when possible
	switch complexity.Overall {
	case Low:
		// For simple tasks, heavily favor tier-3 (cheapest) providers
		switch provider.Tier {
		case UnofficialTier:
			return 1.0
		case CommunityTier:
			return 0.7
		case OfficialTier:
			return 0.3
		}
	case Medium:
		// For medium tasks, prefer tier-2 providers
		switch provider.Tier {
		case CommunityTier:
			return 1.0
		case UnofficialTier:
			return 0.8
		case OfficialTier:
			return 0.6
		}
	case High, VeryHigh:
		// For complex tasks, use tier-1 providers but still consider cost
		switch provider.Tier {
		case OfficialTier:
			return 1.0
		case CommunityTier:
			return 0.4
		case UnofficialTier:
			return 0.1
		}
	}
	
	return 0.5 // Default score
}

// getDefaultCostPerToken returns default cost per token for a provider
func (c *CostBasedSelector) getDefaultCostPerToken(provider *Provider) float64 {
	// Default costs based on provider tier
	switch provider.Tier {
	case OfficialTier:
		return 0.00003 // $0.03 per 1K tokens (GPT-4 level)
	case CommunityTier:
		return 0.00001 // $0.01 per 1K tokens
	case UnofficialTier:
		return 0.0 // Often free
	default:
		return 0.00002 // $0.02 per 1K tokens default
	}
}