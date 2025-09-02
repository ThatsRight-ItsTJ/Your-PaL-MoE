package selection

import (
	"fmt"
	"log"
	"math"
	"strings"
)

// EnhancedAdaptiveSelector uses the model database for accurate provider selection
type EnhancedAdaptiveSelector struct {
	providers  []Provider
	detector   *EnhancedCapabilityDetector
	weights    SelectionWeights
}

// SelectionWeights defines the importance of different factors in provider selection
type SelectionWeights struct {
	CapabilityMatch float64 // How well provider matches task requirements
	QualityScore    float64 // Model quality (reasoning, knowledge, computation)
	Availability    float64 // Provider availability and reliability
	Cost           float64 // Cost considerations (if available)
	Speed          float64 // Response speed (if available)
}

// DefaultSelectionWeights provides balanced weights for provider selection
func DefaultSelectionWeights() SelectionWeights {
	return SelectionWeights{
		CapabilityMatch: 0.4, // 40% - Most important
		QualityScore:    0.3, // 30% - Quality matters
		Availability:    0.2, // 20% - Must be available
		Cost:           0.05, // 5% - Minor consideration
		Speed:          0.05, // 5% - Minor consideration
	}
}

// NewEnhancedAdaptiveSelector creates a new enhanced adaptive selector
func NewEnhancedAdaptiveSelector(providers []Provider) *EnhancedAdaptiveSelector {
	return &EnhancedAdaptiveSelector{
		providers: providers,
		detector:  NewEnhancedCapabilityDetector(),
		weights:   DefaultSelectionWeights(),
	}
}

// SelectProvider chooses the best provider using enhanced model database analysis
func (eas *EnhancedAdaptiveSelector) SelectProvider(request string) (*Provider, float64, error) {
	if len(eas.providers) == 0 {
		return nil, 0, fmt.Errorf("no providers available")
	}

	// Detect task type from request
	taskType := eas.detector.DetectTaskType(request)
	log.Printf("🎯 Detected task type: %s for request: %.50s...", taskType, request)

	// Score all providers
	var bestProvider *Provider
	var bestScore float64 = -1
	var compatibleProviders []string

	for i := range eas.providers {
		provider := &eas.providers[i]
		
		// Get provider capabilities using enhanced detection
		capabilities := eas.detector.DetectCapabilities(provider.Models, provider.Name)
		
		// Check compatibility first
		if !eas.detector.IsProviderCompatible(capabilities, taskType) {
			log.Printf("❌ Provider %s incompatible with task type %s", provider.Name, taskType)
			continue
		}
		
		compatibleProviders = append(compatibleProviders, provider.Name)
		
		// Calculate comprehensive score
		score := eas.calculateProviderScore(provider, capabilities, taskType, request)
		
		log.Printf("📊 Provider %s: score=%.3f, capabilities=%+v", 
			provider.Name, score, capabilities)
		
		if score > bestScore {
			bestScore = score
			bestProvider = provider
		}
	}

	if bestProvider == nil {
		return nil, 0, fmt.Errorf("no compatible providers found for task type %s. Available providers: %v", 
			taskType, eas.getProviderNames())
	}

	log.Printf("🏆 Selected provider: %s (score: %.3f) from compatible: %v", 
		bestProvider.Name, bestScore, compatibleProviders)

	return bestProvider, bestScore, nil
}

// calculateProviderScore computes a comprehensive score for a provider
func (eas *EnhancedAdaptiveSelector) calculateProviderScore(
	provider *Provider, 
	capabilities ProviderCapabilities, 
	taskType TaskType, 
	request string) float64 {
	
	// 1. Capability Match Score (0-1)
	capabilityScore := eas.calculateCapabilityScore(capabilities, taskType)
	
	// 2. Quality Score (0-1) - based on reasoning, knowledge, computation
	qualityScore := eas.calculateQualityScore(capabilities, taskType)
	
	// 3. Availability Score (0-1) - currently simplified
	availabilityScore := eas.calculateAvailabilityScore(provider)
	
	// 4. Cost Score (0-1) - currently simplified
	costScore := eas.calculateCostScore(provider)
	
	// 5. Speed Score (0-1) - currently simplified
	speedScore := eas.calculateSpeedScore(provider)
	
	// Weighted combination
	totalScore := (capabilityScore * eas.weights.CapabilityMatch) +
		(qualityScore * eas.weights.QualityScore) +
		(availabilityScore * eas.weights.Availability) +
		(costScore * eas.weights.Cost) +
		(speedScore * eas.weights.Speed)
	
	log.Printf("🔍 %s scores: capability=%.3f, quality=%.3f, availability=%.3f, cost=%.3f, speed=%.3f, total=%.3f",
		provider.Name, capabilityScore, qualityScore, availabilityScore, costScore, speedScore, totalScore)
	
	return totalScore
}

// calculateCapabilityScore evaluates how well the provider matches the task requirements
func (eas *EnhancedAdaptiveSelector) calculateCapabilityScore(capabilities ProviderCapabilities, taskType TaskType) float64 {
	switch taskType {
	case TaskTypeText:
		if capabilities.Text {
			return 1.0
		}
		return 0.0
		
	case TaskTypeImage:
		if capabilities.Image {
			return 1.0
		}
		return 0.0
		
	case TaskTypeCode:
		if capabilities.Code {
			return 1.0 // Perfect match
		}
		if capabilities.Text && capabilities.Reasoning >= 7 {
			return 0.8 // Good text model can handle code
		}
		if capabilities.Text {
			return 0.6 // Basic text model, limited code capability
		}
		return 0.0
		
	case TaskTypeAudio:
		if capabilities.Audio {
			return 1.0
		}
		return 0.0
		
	case TaskTypeVideo:
		if capabilities.Video {
			return 1.0
		}
		return 0.0
		
	case TaskTypeMultimodal:
		if capabilities.Multimodal {
			return 1.0
		}
		if capabilities.Text && capabilities.Image {
			return 0.9 // Good combination
		}
		return 0.0
		
	default:
		if capabilities.Text {
			return 0.8 // Default to text capability
		}
		return 0.0
	}
}

// calculateQualityScore evaluates the overall quality of the provider's models
func (eas *EnhancedAdaptiveSelector) calculateQualityScore(capabilities ProviderCapabilities, taskType TaskType) float64 {
	// Normalize scores from 1-10 scale to 0-1 scale
	reasoningScore := float64(capabilities.Reasoning) / 10.0
	knowledgeScore := float64(capabilities.Knowledge) / 10.0
	computationScore := float64(capabilities.Computation) / 10.0
	
	// Weight different aspects based on task type
	switch taskType {
	case TaskTypeCode:
		// Code tasks prioritize reasoning and computation
		return (reasoningScore*0.4 + computationScore*0.4 + knowledgeScore*0.2)
		
	case TaskTypeImage:
		// Image tasks prioritize computation and some reasoning
		return (computationScore*0.5 + reasoningScore*0.3 + knowledgeScore*0.2)
		
	case TaskTypeMultimodal:
		// Multimodal tasks need balanced capabilities
		return (reasoningScore*0.4 + knowledgeScore*0.3 + computationScore*0.3)
		
	default:
		// Text and other tasks: balanced with slight reasoning preference
		return (reasoningScore*0.4 + knowledgeScore*0.4 + computationScore*0.2)
	}
}

// calculateAvailabilityScore evaluates provider availability (simplified for now)
func (eas *EnhancedAdaptiveSelector) calculateAvailabilityScore(provider *Provider) float64 {
	// TODO: Implement real availability checking
	// For now, assume all providers are equally available
	
	// Could check:
	// - API endpoint health
	// - Recent response times
	// - Error rates
	// - Rate limit status
	
	return 1.0 // Placeholder
}

// calculateCostScore evaluates cost considerations (simplified for now)
func (eas *EnhancedAdaptiveSelector) calculateCostScore(provider *Provider) float64 {
	// TODO: Implement real cost evaluation
	// For now, provide simple heuristics based on provider type
	
	providerLower := strings.ToLower(provider.Name)
	switch {
	case strings.Contains(providerLower, "local") || strings.Contains(providerLower, "ollama"):
		return 1.0 // Local models are "free"
	case strings.Contains(providerLower, "openai"):
		return 0.6 // Premium pricing
	case strings.Contains(providerLower, "anthropic"):
		return 0.7 // Premium pricing
	case strings.Contains(providerLower, "together"):
		return 0.8 // Competitive pricing
	default:
		return 0.8 // Assume reasonable pricing
	}
}

// calculateSpeedScore evaluates expected response speed (simplified for now)
func (eas *EnhancedAdaptiveSelector) calculateSpeedScore(provider *Provider) float64 {
	// TODO: Implement real speed evaluation
	// For now, provide simple heuristics
	
	providerLower := strings.ToLower(provider.Name)
	switch {
	case strings.Contains(providerLower, "local") || strings.Contains(providerLower, "ollama"):
		return 0.9 // Local can be fast but depends on hardware
	case strings.Contains(providerLower, "openai"):
		return 0.8 // Generally fast
	case strings.Contains(providerLower, "anthropic"):
		return 0.7 // Can be slower
	case strings.Contains(providerLower, "together"):
		return 0.8 // Generally good speed
	default:
		return 0.7 // Conservative estimate
	}
}

// GetProviderCapabilities returns detailed capabilities for all providers
func (eas *EnhancedAdaptiveSelector) GetProviderCapabilities() map[string]ProviderCapabilities {
	result := make(map[string]ProviderCapabilities)
	
	for _, provider := range eas.providers {
		capabilities := eas.detector.DetectCapabilities(provider.Models, provider.Name)
		result[provider.Name] = capabilities
	}
	
	return result
}

// GetDetailedModelCapabilities returns per-model capabilities for all providers
func (eas *EnhancedAdaptiveSelector) GetDetailedModelCapabilities() map[string]map[string]ModelCapabilities {
	result := make(map[string]map[string]ModelCapabilities)
	
	for _, provider := range eas.providers {
		result[provider.Name] = eas.detector.GetDetailedCapabilities(provider.Models, provider.Name)
	}
	
	return result
}

// SetSelectionWeights allows customizing the selection criteria weights
func (eas *EnhancedAdaptiveSelector) SetSelectionWeights(weights SelectionWeights) {
	// Normalize weights to sum to 1.0
	total := weights.CapabilityMatch + weights.QualityScore + weights.Availability + weights.Cost + weights.Speed
	if total > 0 {
		weights.CapabilityMatch /= total
		weights.QualityScore /= total
		weights.Availability /= total
		weights.Cost /= total
		weights.Speed /= total
	}
	
	eas.weights = weights
}

// getProviderNames returns a list of provider names for error messages
func (eas *EnhancedAdaptiveSelector) getProviderNames() []string {
	names := make([]string, len(eas.providers))
	for i, provider := range eas.providers {
		names[i] = provider.Name
	}
	return names
}

// AnalyzeRequest provides detailed analysis of how providers match a request
func (eas *EnhancedAdaptiveSelector) AnalyzeRequest(request string) map[string]interface{} {
	taskType := eas.detector.DetectTaskType(request)
	
	analysis := map[string]interface{}{
		"request":           request,
		"detected_task":     taskType,
		"provider_analysis": make(map[string]interface{}),
	}
	
	providerAnalysis := analysis["provider_analysis"].(map[string]interface{})
	
	for _, provider := range eas.providers {
		capabilities := eas.detector.DetectCapabilities(provider.Models, provider.Name)
		compatible := eas.detector.IsProviderCompatible(capabilities, taskType)
		
		var score float64 = 0
		if compatible {
			score = eas.calculateProviderScore(&provider, capabilities, taskType, request)
		}
		
		providerAnalysis[provider.Name] = map[string]interface{}{
			"compatible":     compatible,
			"score":         score,
			"capabilities":  capabilities,
			"models":        provider.Models,
		}
	}
	
	return analysis
}

// GetModelDatabaseStats returns statistics about the model database
func (eas *EnhancedAdaptiveSelector) GetModelDatabaseStats() map[string]interface{} {
	return eas.detector.GetCacheStats()
}