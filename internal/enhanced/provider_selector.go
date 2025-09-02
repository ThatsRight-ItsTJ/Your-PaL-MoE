package enhanced

import (
	"fmt"
	"sort"
)

// Implementation moved to constructors.go to avoid duplicate method declarations
// This file now only contains helper functions and constants

// providerCapabilities maps capabilities to provider requirements
var providerCapabilities = map[string][]string{
	"reasoning": {"logical_reasoning", "chain_of_thought", "problem_solving"},
	"mathematical": {"mathematical_computation", "symbolic_math", "numerical_analysis"},
	"creative": {"creative_writing", "content_generation", "artistic_creation"},
	"factual": {"knowledge_retrieval", "fact_checking", "information_synthesis"},
}

// tierWeights defines scoring weights for different provider tiers
var tierWeights = map[ProviderTier]float64{
	OfficialTier:   1.0,
	CommunityTier:  0.8,
	UnofficialTier: 0.6,
}

// ProviderScore represents a scored provider option
type ProviderScore struct {
	Provider    *Provider
	Score       float64
	Reasoning   string
	Confidence  float64
}

// scoreProviderForTask calculates how well a provider matches a task
func scoreProviderForTask(provider *Provider, complexity TaskComplexity, requirements map[string]interface{}) ProviderScore {
	score := ProviderScore{
		Provider: provider,
		Score:    0.0,
		Reasoning: "",
		Confidence: 0.0,
	}
	
	// Base score from provider tier
	tierWeight, exists := tierWeights[provider.Tier]
	if !exists {
		tierWeight = 0.5 // Default for unknown tiers
	}
	score.Score += tierWeight * 0.3 // 30% weight for tier
	
	// Capability matching score
	capabilityScore := calculateCapabilityScore(provider, complexity)
	score.Score += capabilityScore * 0.4 // 40% weight for capabilities
	
	// Health score (if available)
	healthScore := 0.8 // Default health score
	if provider.HealthMetrics != nil {
		healthScore = provider.HealthMetrics.HealthScore
	}
	score.Score += healthScore * 0.2 // 20% weight for health
	
	// Cost efficiency score
	costScore := calculateCostScore(provider, complexity.TokenEstimate)
	score.Score += costScore * 0.1 // 10% weight for cost
	
	// Generate reasoning
	score.Reasoning = generateSelectionReasoning(provider, complexity, capabilityScore, healthScore, costScore)
	
	// Calculate confidence based on score distribution
	score.Confidence = calculateConfidence(score.Score)
	
	return score
}

// calculateCapabilityScore evaluates provider capabilities against task requirements
func calculateCapabilityScore(provider *Provider, complexity TaskComplexity) float64 {
	if len(provider.Capabilities) == 0 {
		return 0.5 // Default score for providers without capability info
	}
	
	requiredCapabilities := determineRequiredCapabilities(complexity)
	if len(requiredCapabilities) == 0 {
		return 0.8 // Good score if no specific requirements
	}
	
	matchedCapabilities := 0
	for _, required := range requiredCapabilities {
		for _, available := range provider.Capabilities {
			if available == required {
				matchedCapabilities++
				break
			}
		}
	}
	
	return float64(matchedCapabilities) / float64(len(requiredCapabilities))
}

// determineRequiredCapabilities maps task complexity to required capabilities
func determineRequiredCapabilities(complexity TaskComplexity) []string {
	var required []string
	
	if complexity.Reasoning >= Medium {
		required = append(required, providerCapabilities["reasoning"]...)
	}
	
	if complexity.Mathematical >= Medium {
		required = append(required, providerCapabilities["mathematical"]...)
	}
	
	if complexity.Creative >= Medium {
		required = append(required, providerCapabilities["creative"]...)
	}
	
	if complexity.Factual >= Medium {
		required = append(required, providerCapabilities["factual"]...)
	}
	
	// Remove duplicates
	seen := make(map[string]bool)
	var unique []string
	for _, cap := range required {
		if !seen[cap] {
			seen[cap] = true
			unique = append(unique, cap)
		}
	}
	
	return unique
}

// calculateCostScore evaluates cost efficiency
func calculateCostScore(provider *Provider, estimatedTokens int64) float64 {
	if provider.CostPerToken <= 0 {
		return 1.0 // Free providers get maximum cost score
	}
	
	// Normalize against a baseline cost (e.g., GPT-4 pricing)
	baselineCost := 0.00003 // $0.03 per 1K tokens
	
	if provider.CostPerToken <= baselineCost {
		return 1.0 // Better than baseline gets max score
	}
	
	// Linear decay for higher costs
	costRatio := provider.CostPerToken / baselineCost
	return 1.0 / costRatio
}

// generateSelectionReasoning creates human-readable reasoning for provider selection
func generateSelectionReasoning(provider *Provider, complexity TaskComplexity, capScore, healthScore, costScore float64) string {
	reasons := []string{}
	
	// Tier reasoning
	switch provider.Tier {
	case OfficialTier:
		reasons = append(reasons, "official tier provider with high reliability")
	case CommunityTier:
		reasons = append(reasons, "community tier provider with good balance")
	case UnofficialTier:
		reasons = append(reasons, "unofficial tier provider with cost advantages")
	}
	
	// Capability reasoning
	if capScore >= 0.8 {
		reasons = append(reasons, "excellent capability match for task requirements")
	} else if capScore >= 0.6 {
		reasons = append(reasons, "good capability match for task requirements")
	} else if capScore >= 0.4 {
		reasons = append(reasons, "moderate capability match for task requirements")
	} else {
		reasons = append(reasons, "limited capability match for task requirements")
	}
	
	// Health reasoning
	if healthScore >= 0.9 {
		reasons = append(reasons, "excellent health and performance metrics")
	} else if healthScore >= 0.7 {
		reasons = append(reasons, "good health and performance metrics")
	} else {
		reasons = append(reasons, "moderate health metrics")
	}
	
	// Cost reasoning
	if costScore >= 0.8 {
		reasons = append(reasons, "cost-effective pricing")
	} else if costScore >= 0.6 {
		reasons = append(reasons, "reasonable pricing")
	} else {
		reasons = append(reasons, "higher cost but potentially better quality")
	}
	
	return fmt.Sprintf("Selected due to: %s", joinReasons(reasons))
}

// joinReasons combines reasoning strings appropriately
func joinReasons(reasons []string) string {
	if len(reasons) == 0 {
		return "no specific reasons available"
	}
	if len(reasons) == 1 {
		return reasons[0]
	}
	if len(reasons) == 2 {
		return reasons[0] + " and " + reasons[1]
	}
	
	result := ""
	for i, reason := range reasons {
		if i == len(reasons)-1 {
			result += "and " + reason
		} else if i == 0 {
			result += reason
		} else {
			result += ", " + reason
		}
	}
	
	return result
}

// calculateConfidence determines confidence level based on score
func calculateConfidence(score float64) float64 {
	// Confidence increases with score but with diminishing returns
	if score >= 0.9 {
		return 0.95
	} else if score >= 0.8 {
		return 0.85
	} else if score >= 0.7 {
		return 0.75
	} else if score >= 0.6 {
		return 0.65
	} else if score >= 0.5 {
		return 0.55
	} else {
		return 0.45
	}
}

// sortProvidersByScore sorts providers by their scores in descending order
func sortProvidersByScore(scores []ProviderScore) []ProviderScore {
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].Score > scores[j].Score
	})
	return scores
}

// filterValidProviders removes providers that don't meet minimum requirements
func filterValidProviders(scores []ProviderScore, minScore float64) []ProviderScore {
	var valid []ProviderScore
	for _, score := range scores {
		if score.Score >= minScore {
			valid = append(valid, score)
		}
	}
	return valid
}