package enhanced

import (
	"context"
	"fmt"
	"log"
	"time"
)

// NewEnhancedSystem creates a new enhanced system with default configuration
func NewEnhancedSystem(providers []*Provider) *EnhancedSystem {
	return &EnhancedSystem{
		selector:    NewEnhancedProviderSelector(providers),
		reasoner:    NewTaskReasoner(nil),
		optimizer:   NewSPOOptimizer(nil),
		healthMonitor: NewProviderHealthMonitor(),
		providers:   providers,
		metrics:     NewSystemMetrics(),
	}
}

// ProcessRequest processes a request using the enhanced system
func (es *EnhancedSystem) ProcessRequest(ctx context.Context, input RequestInput) (*ProcessResponse, error) {
	startTime := time.Now()

	// Analyze task complexity
	complexity, err := es.reasoner.AnalyzeComplexity(ctx, input.Content, string(input.TaskType))
	if err != nil {
		return nil, fmt.Errorf("failed to analyze complexity: %w", err)
	}

	// Optimize prompt
	optimizedPrompt, err := es.optimizer.OptimizePrompt(ctx, input.Content, *complexity)
	if err != nil {
		log.Printf("Failed to optimize prompt: %v", err)
		// Continue with original prompt
	}

	// Select provider
	assignment, err := es.selector.SelectProviderWithCapabilities(ctx, *complexity, complexity.RequiredCapabilities)
	if err != nil {
		return nil, fmt.Errorf("failed to select provider: %w", err)
	}

	// Update metrics
	es.metrics.IncrementTotalRequests()
	es.metrics.RecordComplexity(complexity.Overall)
	es.metrics.RecordProviderUsage(assignment.Provider.Name)

	// Process with selected provider (placeholder)
	response := &ProcessResponse{
		Content:     fmt.Sprintf("Processed by %s using model %s", assignment.Provider.Name, assignment.Model),
		Provider:    assignment.Provider,
		Model:       assignment.Model,
		Complexity:  *complexity,
		ProcessingTime: time.Since(startTime),
		TokensUsed:  complexity.TokenEstimate,
		Cost:        assignment.EstimatedCost,
		Metadata:    make(map[string]interface{}),
	}

	if optimizedPrompt != nil {
		response.Metadata["optimized_prompt"] = optimizedPrompt.OptimizedPrompt
		response.Metadata["optimization_rules"] = optimizedPrompt.OptimizationRules
	}

	// Update provider health metrics
	es.healthMonitor.UpdateMetrics(assignment.Provider.Name, true, time.Since(startTime))

	return response, nil
}

// GetSystemMetrics returns system metrics
func (es *EnhancedSystem) GetSystemMetrics() *SystemMetrics {
	return es.metrics
}

// GetProviderMetrics returns metrics for all providers
func (es *EnhancedSystem) GetProviderMetrics() map[string]*ProviderHealthMetrics {
	return es.healthMonitor.GetAllMetrics()
}

// GetProviderMetricsByName returns metrics for a specific provider
func (es *EnhancedSystem) GetProviderMetricsByName(providerName string) *ProviderHealthMetrics {
	return es.healthMonitor.GetMetrics(providerName)
}

// UpdateProviderHealth updates health metrics for a provider
func (es *EnhancedSystem) UpdateProviderHealth(providerName string, success bool, latency time.Duration) {
	es.healthMonitor.UpdateMetrics(providerName, success, latency)
}

// GetHealthyProviders returns a list of healthy providers
func (es *EnhancedSystem) GetHealthyProviders() []*Provider {
	var healthyProviders []*Provider
	for _, provider := range es.providers {
		if es.healthMonitor.IsHealthy(provider.Name) {
			healthyProviders = append(healthyProviders, provider)
		}
	}
	return healthyProviders
}

// OptimizePromptOnly optimizes a prompt without full processing
func (es *EnhancedSystem) OptimizePromptOnly(ctx context.Context, input RequestInput) (*OptimizedPrompt, error) {
	// Analyze complexity first
	complexity, err := es.reasoner.AnalyzeComplexity(ctx, input.Content, string(input.TaskType))
	if err != nil {
		return nil, fmt.Errorf("failed to analyze complexity: %w", err)
	}

	// Optimize prompt
	return es.optimizer.OptimizePrompt(ctx, input.Content, *complexity)
}

// AnalyzeComplexityOnly analyzes task complexity without full processing
func (es *EnhancedSystem) AnalyzeComplexityOnly(ctx context.Context, input RequestInput) (*TaskComplexity, error) {
	return es.reasoner.AnalyzeComplexity(ctx, input.Content, string(input.TaskType))
}

// SelectProviderOnly selects a provider without full processing
func (es *EnhancedSystem) SelectProviderOnly(ctx context.Context, complexity TaskComplexity, requiredCapabilities []string) (*ProviderAssignment, error) {
	return es.selector.SelectProviderWithCapabilities(ctx, complexity, requiredCapabilities)
}

// GetProviderStats returns statistics about providers
func (es *EnhancedSystem) GetProviderStats() map[string]interface{} {
	return es.selector.GetProviderStats()
}

// ResetProviderMetrics resets metrics for a specific provider
func (es *EnhancedSystem) ResetProviderMetrics(providerName string) {
	es.healthMonitor.ResetMetrics(providerName)
}

// Shutdown gracefully shuts down the enhanced system
func (es *EnhancedSystem) Shutdown(ctx context.Context) error {
	log.Println("Shutting down enhanced system...")
	// Add any cleanup logic here
	return nil
}