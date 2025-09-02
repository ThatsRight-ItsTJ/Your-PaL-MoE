package enhanced

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"time"
)

// EnhancedProviderSelector provides advanced provider selection with capability filtering
type EnhancedProviderSelector struct {
	providers         []*Provider
	capabilityFilters map[string][]string
	healthCalculator  *HealthScoreCalculator
	costOptimizer     *CostBasedSelector
}

// NewEnhancedProviderSelector creates a new enhanced provider selector
func NewEnhancedProviderSelector(providers []*Provider) *EnhancedProviderSelector {
	return &EnhancedProviderSelector{
		providers: providers,
		capabilityFilters: map[string][]string{
			"reasoning":    {"gpt-4", "claude-3", "gemini-pro"},
			"mathematical": {"gpt-4", "claude-3", "codex"},
			"creative":     {"gpt-4", "claude-3", "dall-e"},
			"factual":      {"gpt-3.5", "claude-instant", "gemini"},
		},
		healthCalculator: NewHealthScoreCalculator(),
		costOptimizer:    NewCostBasedSelector(nil, nil),
	}
}

// LoadProvidersFromCSV loads providers from a CSV file
func (eps *EnhancedProviderSelector) LoadProvidersFromCSV(reader io.Reader) ([]*Provider, error) {
	csvReader := csv.NewReader(reader)
	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	var providers []*Provider
	for i, record := range records {
		if i == 0 { // Skip header
			continue
		}

		if len(record) < 6 {
			log.Printf("Skipping invalid record at line %d: insufficient fields", i+1)
			continue
		}

		// Parse max tokens
		maxTokens, err := strconv.ParseInt(strings.TrimSpace(record[3]), 10, 64)
		if err != nil {
			log.Printf("Invalid max_tokens at line %d: %v", i+1, err)
			maxTokens = 4096 // Default value
		}

		// Parse cost per token
		costPerToken, err := strconv.ParseFloat(strings.TrimSpace(record[5]), 64)
		if err != nil {
			log.Printf("Invalid cost_per_token at line %d: %v", i+1, err)
			costPerToken = 0.00003 // Default value
		}

		// Parse tier
		tierStr := strings.ToLower(strings.TrimSpace(record[1]))
		var tier ProviderTier
		switch tierStr {
		case "official":
			tier = OfficialTier
		case "community":
			tier = CommunityTier
		case "unofficial":
			tier = UnofficialTier
		default:
			tier = CommunityTier // Default
		}

		// Parse models (comma-separated)
		modelsStr := strings.TrimSpace(record[4])
		var models []string
		if modelsStr != "" {
			models = strings.Split(modelsStr, ",")
			for j := range models {
				models[j] = strings.TrimSpace(models[j])
			}
		}

		provider := &Provider{
			Name:         strings.TrimSpace(record[0]),
			BaseURL:      strings.TrimSpace(record[2]),
			Models:       models,
			Tier:         tier,
			MaxTokens:    maxTokens,
			CostPerToken: costPerToken,
			Capabilities: []string{"reasoning", "creative", "factual"}, // Default capabilities
			RateLimits:   map[string]int64{"requests_per_minute": 60},
			Metadata:     make(map[string]interface{}),
			LastUpdated:  time.Now(),
		}

		providers = append(providers, provider)
	}

	return providers, nil
}

// SelectProviderWithCapabilities selects a provider based on task complexity and required capabilities
func (eps *EnhancedProviderSelector) SelectProviderWithCapabilities(ctx context.Context, complexity TaskComplexity, requiredCapabilities []string) (*ProviderAssignment, error) {
	if len(eps.providers) == 0 {
		return nil, fmt.Errorf("no providers available")
	}

	// Filter providers by capabilities
	compatibleProviders := eps.filterProvidersByCapabilities(requiredCapabilities)
	if len(compatibleProviders) == 0 {
		// Fallback to all providers if no exact matches
		compatibleProviders = eps.providers
	}

	// Score providers
	var scores []ProviderScore
	for _, provider := range compatibleProviders {
		score := eps.scoreProviderForComplexity(provider, complexity)
		scores = append(scores, score)
	}

	// Sort by score (highest first)
	scores = sortProvidersByScore(scores)

	// Select best provider
	bestScore := scores[0]
	
	assignment := &ProviderAssignment{
		Provider:        bestScore.Provider,
		Model:          eps.selectBestModel(bestScore.Provider, complexity),
		Confidence:     bestScore.Confidence,
		EstimatedCost:  float64(complexity.TokenEstimate) * bestScore.Provider.CostPerToken,
		EstimatedTokens: complexity.TokenEstimate,
		Reasoning:      bestScore.Reasoning,
		Alternatives:   []*Provider{}, // Could populate with other high-scoring providers
		Metadata:       make(map[string]interface{}),
	}

	return assignment, nil
}

// filterProvidersByCapabilities filters providers based on required capabilities
func (eps *EnhancedProviderSelector) filterProvidersByCapabilities(requiredCapabilities []string) []*Provider {
	var compatibleProviders []*Provider

	for _, provider := range eps.providers {
		isCompatible := true
		for _, requiredCap := range requiredCapabilities {
			if !eps.providerHasCapability(provider, requiredCap) {
				isCompatible = false
				break
			}
		}
		if isCompatible {
			compatibleProviders = append(compatibleProviders, provider)
		}
	}

	return compatibleProviders
}

// providerHasCapability checks if a provider has a specific capability
func (eps *EnhancedProviderSelector) providerHasCapability(provider *Provider, capability string) bool {
	// Check provider's declared capabilities
	for _, cap := range provider.Capabilities {
		if strings.EqualFold(cap, capability) {
			return true
		}
	}

	// Check capability filters
	if compatibleModels, exists := eps.capabilityFilters[capability]; exists {
		for _, model := range provider.Models {
			for _, compatibleModel := range compatibleModels {
				if strings.Contains(strings.ToLower(model), strings.ToLower(compatibleModel)) {
					return true
				}
			}
		}
	}

	return false
}

// scoreProviderForComplexity scores a provider based on task complexity
func (eps *EnhancedProviderSelector) scoreProviderForComplexity(provider *Provider, complexity TaskComplexity) ProviderScore {
	score := 0.0
	reasoning := "Provider scoring: "

	// Base score from tier
	switch provider.Tier {
	case OfficialTier:
		score += 0.4
		reasoning += "Official tier (+0.4), "
	case CommunityTier:
		score += 0.2
		reasoning += "Community tier (+0.2), "
	case UnofficialTier:
		score += 0.1
		reasoning += "Unofficial tier (+0.1), "
	}

	// Complexity-based scoring
	complexityScore := float64(complexity.Overall) / float64(VeryHigh)
	score += complexityScore * 0.3
	reasoning += fmt.Sprintf("Complexity match (+%.2f), ", complexityScore*0.3)

	// Cost efficiency (lower cost = higher score)
	if provider.CostPerToken > 0 {
		costScore := 1.0 / (provider.CostPerToken * 100000) // Normalize cost
		if costScore > 0.2 {
			costScore = 0.2 // Cap cost benefit
		}
		score += costScore
		reasoning += fmt.Sprintf("Cost efficiency (+%.2f), ", costScore)
	}

	// Token capacity
	if provider.MaxTokens >= complexity.TokenEstimate {
		score += 0.1
		reasoning += "Sufficient tokens (+0.1), "
	}

	// Health metrics (if available)
	if provider.HealthMetrics != nil {
		healthScore := eps.calculateHealthScore(provider)
		score += healthScore * 0.2
		reasoning += fmt.Sprintf("Health score (+%.2f), ", healthScore*0.2)
	}

	return ProviderScore{
		Provider:   provider,
		Score:      score,
		Confidence: score, // Use score as confidence for now
		Reasoning:  strings.TrimSuffix(reasoning, ", "),
	}
}

// calculateHealthScore calculates a health score for a provider
func (eps *EnhancedProviderSelector) calculateHealthScore(provider *Provider) float64 {
	if provider.HealthMetrics == nil {
		return 0.5 // Neutral score if no health data
	}

	score := 0.0
	
	// Success rate
	if provider.HealthMetrics.SuccessRate > 0.9 {
		score += 0.4
	} else if provider.HealthMetrics.SuccessRate > 0.8 {
		score += 0.3
	} else if provider.HealthMetrics.SuccessRate > 0.7 {
		score += 0.2
	}

	// Average latency (lower is better)
	if provider.HealthMetrics.AverageLatency < 1000 { // Less than 1 second
		score += 0.3
	} else if provider.HealthMetrics.AverageLatency < 3000 { // Less than 3 seconds
		score += 0.2
	} else if provider.HealthMetrics.AverageLatency < 5000 { // Less than 5 seconds
		score += 0.1
	}

	// Error rate (lower is better)
	if provider.HealthMetrics.ErrorRate < 0.05 { // Less than 5%
		score += 0.3
	} else if provider.HealthMetrics.ErrorRate < 0.1 { // Less than 10%
		score += 0.2
	} else if provider.HealthMetrics.ErrorRate < 0.2 { // Less than 20%
		score += 0.1
	}

	return score
}

// selectBestModel selects the best model from a provider for the given complexity
func (eps *EnhancedProviderSelector) selectBestModel(provider *Provider, complexity TaskComplexity) string {
	if len(provider.Models) == 0 {
		return "default"
	}

	// Simple model selection based on complexity
	switch complexity.Overall {
	case VeryHigh:
		// Prefer most capable models
		for _, model := range provider.Models {
			if strings.Contains(strings.ToLower(model), "gpt-4") ||
				strings.Contains(strings.ToLower(model), "claude-3") ||
				strings.Contains(strings.ToLower(model), "opus") {
				return model
			}
		}
	case High:
		// Prefer balanced models
		for _, model := range provider.Models {
			if strings.Contains(strings.ToLower(model), "gpt-4") ||
				strings.Contains(strings.ToLower(model), "claude") ||
				strings.Contains(strings.ToLower(model), "sonnet") {
				return model
			}
		}
	case Medium:
		// Prefer efficient models
		for _, model := range provider.Models {
			if strings.Contains(strings.ToLower(model), "gpt-3.5") ||
				strings.Contains(strings.ToLower(model), "claude-instant") ||
				strings.Contains(strings.ToLower(model), "haiku") {
				return model
			}
		}
	}

	// Default to first available model
	return provider.Models[0]
}

// GetCapabilityFilters returns the current capability filters
func (eps *EnhancedProviderSelector) GetCapabilityFilters() map[string][]string {
	return eps.capabilityFilters
}

// SetCapabilityFilters sets new capability filters
func (eps *EnhancedProviderSelector) SetCapabilityFilters(filters map[string][]string) {
	eps.capabilityFilters = filters
}

// GetProviderStats returns statistics about providers
func (eps *EnhancedProviderSelector) GetProviderStats() map[string]interface{} {
	stats := map[string]interface{}{
		"total_providers": len(eps.providers),
		"providers_by_tier": map[string]int{
			"official":   0,
			"community":  0,
			"unofficial": 0,
		},
		"total_models": 0,
		"capabilities": eps.capabilityFilters,
	}

	for _, provider := range eps.providers {
		stats["total_models"] = stats["total_models"].(int) + len(provider.Models)
		
		tierStats := stats["providers_by_tier"].(map[string]int)
		switch provider.Tier {
		case OfficialTier:
			tierStats["official"]++
		case CommunityTier:
			tierStats["community"]++
		case UnofficialTier:
			tierStats["unofficial"]++
		}
	}

	return stats
}