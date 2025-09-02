package enhanced

import (
	"context"
	"fmt"
	"log"
	"math"
	"sort"
	"strings"
	"time"
)

// SelectProvider selects the best provider for a given task
func (ps *ProviderSelector) SelectProvider(ctx context.Context, input RequestInput, complexity *TaskComplexity) (*ProviderAssignment, error) {
	if len(ps.providers) == 0 {
		return nil, fmt.Errorf("no providers available")
	}

	// Score each provider based on the task complexity and provider capabilities
	scores := make(map[string]float64)
	
	for _, provider := range ps.providers {
		if !provider.Enabled {
			continue
		}
		
		score := ps.calculateProviderScore(provider, complexity)
		scores[provider.ID] = score
	}
	
	if len(scores) == 0 {
		return nil, fmt.Errorf("no enabled providers available")
	}
	
	// Find the best provider
	var bestProvider *Provider
	var bestScore float64
	
	for _, provider := range ps.providers {
		if score, exists := scores[provider.ID]; exists && score > bestScore {
			bestProvider = provider
			bestScore = score
		}
	}
	
	if bestProvider == nil {
		return nil, fmt.Errorf("failed to select provider")
	}
	
	// Select the best model for this provider
	model := ps.selectBestModel(bestProvider, complexity)
	
	// Calculate estimated cost and time
	estimatedCost := ps.calculateEstimatedCost(bestProvider, input.Query)
	estimatedTime := ps.calculateEstimatedTime(bestProvider, complexity)
	
	// Generate alternatives
	alternatives := ps.generateAlternatives(bestProvider, complexity, scores)
	
	assignment := &ProviderAssignment{
		Provider:      bestProvider.ID,
		Model:         model,
		Reasoning:     ps.generateReasoning(bestProvider, complexity, bestScore),
		Confidence:    bestScore / 4.0, // Normalize to 0-1
		EstimatedCost: estimatedCost,
		EstimatedTime: estimatedTime,
		Alternatives:  alternatives,
		ProviderName:  bestProvider.Name,
		ProviderTier:  bestProvider.Tier,
		Tier:          bestProvider.Tier,
	}
	
	log.Printf("Selected provider: %s (score: %.2f) for task with complexity: %.2f", 
		bestProvider.Name, bestScore, complexity.Overall)
	
	return assignment, nil
}

// calculateProviderScore calculates a score for a provider based on task complexity
func (ps *ProviderSelector) calculateProviderScore(provider *Provider, complexity *TaskComplexity) float64 {
	score := 0.0
	
	// Base score from provider priority
	score += float64(provider.Priority) * 0.1
	
	// Tier-based scoring
	switch strings.ToLower(provider.Tier) {
	case "premium", "tier1":
		score += 4.0
	case "standard", "tier2":
		score += 3.0
	case "basic", "tier3":
		score += 2.0
	default:
		score += 1.0
	}
	
	// Complexity matching - prefer higher tier providers for complex tasks
	complexityLevel := FloatToComplexityLevel(complexity.Overall)
	switch complexityLevel {
	case VeryHigh:
		if strings.ToLower(provider.Tier) == "premium" {
			score += 2.0
		}
	case High:
		if strings.ToLower(provider.Tier) == "premium" || strings.ToLower(provider.Tier) == "standard" {
			score += 1.5
		}
	case Medium:
		if strings.ToLower(provider.Tier) == "standard" || strings.ToLower(provider.Tier) == "basic" {
			score += 1.0
		}
	case Low:
		score += 0.5 // Any provider can handle low complexity
	}
	
	// Performance metrics (if available)
	if provider.Metrics.SuccessRate > 0 {
		score += provider.Metrics.SuccessRate * 2.0
	}
	
	// Penalize high error rates
	if provider.Metrics.ErrorRate > 0 {
		score -= provider.Metrics.ErrorRate * 3.0
	}
	
	// Consider latency (lower is better)
	if provider.Metrics.AverageLatency > 0 {
		latencyPenalty := float64(provider.Metrics.AverageLatency.Milliseconds()) / 1000.0
		score -= latencyPenalty * 0.1
	}
	
	// Health status bonus
	if provider.HealthMetrics.Status == "healthy" {
		score += 1.0
	}
	
	return math.Max(0, score) // Ensure non-negative score
}

// selectBestModel selects the best model for a provider based on complexity
func (ps *ProviderSelector) selectBestModel(provider *Provider, complexity *TaskComplexity) string {
	if len(provider.Models) == 0 {
		return "default"
	}
	
	// For now, just return the first model
	// In a real implementation, this would consider model capabilities
	return provider.Models[0]
}

// calculateEstimatedCost estimates the cost for processing the request
func (ps *ProviderSelector) calculateEstimatedCost(provider *Provider, query string) float64 {
	// Simple token estimation (rough approximation)
	tokenCount := len(strings.Fields(query)) * 1.3 // Rough token estimate
	
	inputCost := tokenCount * provider.Pricing.InputTokenCost
	outputCost := tokenCount * 0.5 * provider.Pricing.OutputTokenCost // Assume output is 50% of input
	
	return inputCost + outputCost
}

// calculateEstimatedTime estimates processing time
func (ps *ProviderSelector) calculateEstimatedTime(provider *Provider, complexity *TaskComplexity) time.Duration {
	baseTime := 1 * time.Second
	
	// Adjust based on complexity
	complexityMultiplier := complexity.Overall / 2.0
	
	// Adjust based on provider performance
	if provider.Metrics.AverageLatency > 0 {
		return time.Duration(float64(provider.Metrics.AverageLatency) * complexityMultiplier)
	}
	
	return time.Duration(float64(baseTime) * complexityMultiplier)
}

// generateAlternatives generates alternative provider options
func (ps *ProviderSelector) generateAlternatives(selectedProvider *Provider, complexity *TaskComplexity, scores map[string]float64) []AlternativeProvider {
	var alternatives []AlternativeProvider
	
	// Sort providers by score
	type providerScore struct {
		provider *Provider
		score    float64
	}
	
	var sortedProviders []providerScore
	for _, provider := range ps.providers {
		if provider.ID != selectedProvider.ID && provider.Enabled {
			if score, exists := scores[provider.ID]; exists {
				sortedProviders = append(sortedProviders, providerScore{provider, score})
			}
		}
	}
	
	sort.Slice(sortedProviders, func(i, j int) bool {
		return sortedProviders[i].score > sortedProviders[j].score
	})
	
	// Take top 3 alternatives
	maxAlternatives := 3
	if len(sortedProviders) < maxAlternatives {
		maxAlternatives = len(sortedProviders)
	}
	
	for i := 0; i < maxAlternatives; i++ {
		provider := sortedProviders[i].provider
		alternatives = append(alternatives, AlternativeProvider{
			Provider:      provider.ID,
			Model:         ps.selectBestModel(provider, complexity),
			Confidence:    sortedProviders[i].score / 4.0,
			EstimatedCost: ps.calculateEstimatedCost(provider, ""),
			EstimatedTime: ps.calculateEstimatedTime(provider, complexity),
			Reasoning:     fmt.Sprintf("Alternative option with score %.2f", sortedProviders[i].score),
			ProviderID:    provider.ID,
			ProviderName:  provider.Name,
		})
	}
	
	return alternatives
}

// generateReasoning generates reasoning for provider selection
func (ps *ProviderSelector) generateReasoning(provider *Provider, complexity *TaskComplexity, score float64) string {
	complexityLevel := FloatToComplexityLevel(complexity.Overall)
	
	reasoning := fmt.Sprintf("Selected %s (tier: %s) with score %.2f for %s complexity task. ", 
		provider.Name, provider.Tier, score, complexityLevel)
	
	if provider.Metrics.SuccessRate > 0.9 {
		reasoning += "High success rate. "
	}
	
	if provider.HealthMetrics.Status == "healthy" {
		reasoning += "Provider is healthy. "
	}
	
	return reasoning
}