package enhanced

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// NewEnhancedSystem creates a new enhanced system instance
func NewEnhancedSystem(providers []*Provider) *EnhancedSystem {
	// Initialize metrics storage
	metricsStorage, _ := NewMetricsStorage("enhanced_metrics.db")
	
	// Initialize rate limit manager
	rateLimitManager := NewRateLimitManager(metricsStorage)
	
	// Initialize health calculator
	healthCalculator := NewHealthScoreCalculator()
	
	// Initialize cost-based selector
	costBasedSelector := NewCostBasedSelector(metricsStorage, rateLimitManager)
	
	// Initialize components
	taskReasoner := NewTaskReasoner()
	providerSelector := NewProviderSelector(providers, healthCalculator, costBasedSelector)
	spoOptimizer := NewSPOOptimizer()
	
	return &EnhancedSystem{
		taskReasoner:      taskReasoner,
		providerSelector:  providerSelector,
		spoOptimizer:      spoOptimizer,
		providers:         providers,
		metricsStorage:    metricsStorage,
		rateLimitManager:  rateLimitManager,
		healthCalculator:  healthCalculator,
	}
}

// NewTaskReasoner creates a new task reasoner
func NewTaskReasoner() *TaskReasoner {
	return &TaskReasoner{
		complexityWeights: map[string]float64{
			"reasoning":    0.3,
			"mathematical": 0.25,
			"creative":     0.2,
			"factual":      0.25,
		},
		tokenEstimator: NewTokenEstimator(),
	}
}

// AnalyzeComplexity analyzes the complexity of a task
func (tr *TaskReasoner) AnalyzeComplexity(content string) (*TaskComplexity, error) {
	if content == "" {
		return nil, fmt.Errorf("content cannot be empty")
	}
	
	// Analyze different aspects of complexity
	reasoning := detectReasoningComplexity(content)
	mathematical := detectMathematicalComplexity(content)
	creative := detectCreativeComplexity(content)
	factual := Low // Default factual complexity
	
	// Calculate overall complexity
	overall := determineOverallComplexity(reasoning, mathematical, creative, factual)
	
	// Estimate tokens
	tokenEstimate := estimateTokensFromText(content)
	
	return &TaskComplexity{
		Overall:      overall,
		Reasoning:    reasoning,
		Mathematical: mathematical,
		Creative:     creative,
		Factual:      factual,
		TokenEstimate: tokenEstimate,
		RequiredCapabilities: []string{}, // Will be populated based on complexity
		Metadata:     make(map[string]interface{}),
	}, nil
}

// NewProviderSelector creates a new provider selector
func NewProviderSelector(providers []*Provider, healthCalculator *HealthScoreCalculator, costOptimizer *CostBasedSelector) *ProviderSelector {
	return &ProviderSelector{
		providers:        providers,
		healthCalculator: healthCalculator,
		costOptimizer:    costOptimizer,
	}
}

// SelectProvider selects the optimal provider for a task
func (ps *ProviderSelector) SelectProvider(complexity TaskComplexity, requirements map[string]interface{}) (*ProviderAssignment, error) {
	if len(ps.providers) == 0 {
		return nil, fmt.Errorf("no providers available")
	}
	
	// Score all providers
	var scores []ProviderScore
	for _, provider := range ps.providers {
		score := scoreProviderForTask(provider, complexity, requirements)
		scores = append(scores, score)
	}
	
	// Sort by score
	scores = sortProvidersByScore(scores)
	
	// Filter valid providers
	validScores := filterValidProviders(scores, 0.3) // Minimum score threshold
	if len(validScores) == 0 {
		// If no providers meet threshold, use best available
		validScores = scores[:1]
	}
	
	bestScore := validScores[0]
	
	// Create assignment
	assignment := &ProviderAssignment{
		Provider:        bestScore.Provider,
		Model:          bestScore.Provider.Models[0], // Use first available model
		Confidence:     bestScore.Confidence,
		EstimatedCost:  float64(complexity.TokenEstimate) * bestScore.Provider.CostPerToken,
		EstimatedTokens: complexity.TokenEstimate,
		Reasoning:      bestScore.Reasoning,
		Alternatives:   []*Provider{}, // Could populate with other high-scoring providers
		Metadata:       make(map[string]interface{}),
	}
	
	return assignment, nil
}

// NewSPOOptimizer creates a new SPO optimizer
func NewSPOOptimizer() *SPOOptimizer {
	return &SPOOptimizer{
		optimizationRules: map[ComplexityLevel][]string{
			Low:      {"Keep responses concise", "Use simple language"},
			Medium:   {"Provide context", "Use examples"},
			High:     {"Break down problems", "Provide detailed reasoning"},
			VeryHigh: {"Use systematic analysis", "Employ chain-of-thought"},
		},
		templateCache: make(map[string]string),
	}
}

// OptimizePrompt optimizes a prompt based on task complexity
func (spo *SPOOptimizer) OptimizePrompt(prompt string, complexity TaskComplexity) (string, error) {
	if prompt == "" {
		return "", fmt.Errorf("prompt cannot be empty")
	}
	
	// Apply complexity-specific optimization
	optimized := applyOptimizationRules(prompt, complexity.Overall)
	
	// Enhance for specific complexity aspects
	optimized = enhancePromptForComplexity(optimized, complexity)
	
	// Validate the optimized prompt
	if !validateOptimizedPrompt(prompt, optimized) {
		return prompt, nil // Return original if validation fails
	}
	
	return optimized, nil
}

// NewTokenEstimator creates a new token estimator
func NewTokenEstimator() *TokenEstimator {
	return &TokenEstimator{
		baseTokensPerWord: 1.33, // Rough approximation
		complexityMultipliers: map[ComplexityLevel]float64{
			Low:      1.0,
			Medium:   1.2,
			High:     1.5,
			VeryHigh: 2.0,
		},
	}
}

// EstimateTokens estimates token usage for content
func (te *TokenEstimator) EstimateTokens(content string, complexity ComplexityLevel) int64 {
	words := strings.Fields(content)
	baseTokens := float64(len(words)) * te.baseTokensPerWord
	
	multiplier, exists := te.complexityMultipliers[complexity]
	if !exists {
		multiplier = 1.0
	}
	
	return int64(baseTokens * multiplier)
}

// ProcessRequest processes a request through the enhanced system
func (es *EnhancedSystem) ProcessRequest(ctx context.Context, input RequestInput) (*ProcessResponse, error) {
	if input.Content == "" {
		return nil, fmt.Errorf("request content cannot be empty")
	}
	
	startTime := time.Now()
	
	// Analyze task complexity
	complexity, err := es.taskReasoner.AnalyzeComplexity(input.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze complexity: %w", err)
	}
	
	// Select optimal provider
	requirements := make(map[string]interface{})
	if input.PreferredProvider != "" {
		requirements["preferred_provider"] = input.PreferredProvider
	}
	
	assignment, err := es.providerSelector.SelectProvider(*complexity, requirements)
	if err != nil {
		return nil, fmt.Errorf("failed to select provider: %w", err)
	}
	
	// Optimize prompt
	optimizedPrompt, err := es.spoOptimizer.OptimizePrompt(input.Content, *complexity)
	if err != nil {
		return nil, fmt.Errorf("failed to optimize prompt: %w", err)
	}
	
	// Simulate processing (in real implementation, this would call the provider's API)
	latency := time.Since(startTime)
	
	response := &ProcessResponse{
		Content:     fmt.Sprintf("Processed: %s", optimizedPrompt),
		TokensUsed:  complexity.TokenEstimate,
		Cost:        assignment.EstimatedCost,
		Latency:     latency,
		Provider:    assignment.Provider.Name,
		Model:       assignment.Model,
		Success:     true,
		Error:       nil,
		Metadata:    make(map[string]interface{}),
	}
	
	return response, nil
}

// GetRequest processes a single request (alias for ProcessRequest)
func (es *EnhancedSystem) GetRequest(ctx context.Context, input RequestInput) (*ProcessResponse, error) {
	return es.ProcessRequest(ctx, input)
}

// GetProviders returns all available providers
func (es *EnhancedSystem) GetProviders() []*Provider {
	return es.providers
}

// GenerateProviderYAML generates YAML configuration for a specific provider
func (es *EnhancedSystem) GenerateProviderYAML(providerName string) (string, error) {
	for _, provider := range es.providers {
		if provider.Name == providerName {
			yaml := fmt.Sprintf(`name: %s
base_url: %s
models:
%s
tier: %s
max_tokens: %d
cost_per_token: %f
capabilities:
%s
`, 
				provider.Name,
				provider.BaseURL,
				formatModelsYAML(provider.Models),
				provider.Tier.String(),
				provider.MaxTokens,
				provider.CostPerToken,
				formatCapabilitiesYAML(provider.Capabilities),
			)
			return yaml, nil
		}
	}
	return "", fmt.Errorf("provider %s not found", providerName)
}

// GenerateAllProviderYAMLs generates YAML configuration for all providers
func (es *EnhancedSystem) GenerateAllProviderYAMLs() (string, error) {
	var yamlBuilder strings.Builder
	yamlBuilder.WriteString("providers:\n")
	
	for _, provider := range es.providers {
		yaml, err := es.GenerateProviderYAML(provider.Name)
		if err != nil {
			continue
		}
		
		// Indent the provider YAML
		lines := strings.Split(yaml, "\n")
		for _, line := range lines {
			if line != "" {
				yamlBuilder.WriteString("  " + line + "\n")
			}
		}
		yamlBuilder.WriteString("\n")
	}
	
	return yamlBuilder.String(), nil
}

// Helper functions for YAML formatting
func formatModelsYAML(models []string) string {
	var result strings.Builder
	for _, model := range models {
		result.WriteString(fmt.Sprintf("  - %s\n", model))
	}
	return strings.TrimSuffix(result.String(), "\n")
}

func formatCapabilitiesYAML(capabilities []string) string {
	var result strings.Builder
	for _, capability := range capabilities {
		result.WriteString(fmt.Sprintf("  - %s\n", capability))
	}
	return strings.TrimSuffix(result.String(), "\n")
}