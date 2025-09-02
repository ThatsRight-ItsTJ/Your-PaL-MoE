package enhanced

import (
	"context"
	"fmt"
	"log"
	"time"
)

// NewEnhancedSystem creates a new enhanced system instance
func NewEnhancedSystem(providers []*Provider, config *Config) *EnhancedSystem {
	return &EnhancedSystem{
		providers:         providers,
		taskReasoner:      NewTaskReasoner(config),
		providerSelector:  NewProviderSelector(providers),
		spoOptimizer:      NewSPOOptimizer(config),
		healthMonitor:     NewProviderHealthMonitor(providers),
		metrics:           SystemMetrics{},
	}
}

// ProcessRequest processes a request through the enhanced system
func (es *EnhancedSystem) ProcessRequest(ctx context.Context, input RequestInput) (*RequestResult, error) {
	startTime := time.Now()

	// Analyze task complexity
	complexity, err := es.taskReasoner.AnalyzeComplexity(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze complexity: %w", err)
	}

	// Convert complexity to the expected format
	taskComplexity := TaskComplexity{
		Overall: complexity.Overall,
		Score:   complexity.Overall,
	}

	// Optimize prompt if needed
	optimizedPrompt, err := es.spoOptimizer.OptimizePrompt(ctx, input, complexity)
	if err != nil {
		log.Printf("Warning: prompt optimization failed: %v", err)
		optimizedPrompt = &OptimizedPrompt{
			OriginalPrompt: input.Query,
			OptimizedText:  input.Query,
			CostSavings:    0,
		}
	}

	// Select best provider
	assignment, err := es.providerSelector.SelectProvider(ctx, input, complexity)
	if err != nil {
		return nil, fmt.Errorf("failed to select provider: %w", err)
	}

	// Update system metrics
	es.updateMetrics(startTime, true)

	// Execute the request (placeholder implementation)
	result := &RequestResult{
		Response:   "Processed successfully", // This would be the actual LLM response
		Provider:   assignment.Provider,
		Model:      assignment.Model,
		Complexity: taskComplexity,
		Cost:       assignment.EstimatedCost,
		Duration:   time.Since(startTime),
		Metadata: map[string]interface{}{
			"optimized_prompt": optimizedPrompt.OptimizedText,
			"reasoning":        assignment.Reasoning,
		},
	}

	return result, nil
}

// GetMetrics returns current system metrics
func (es *EnhancedSystem) GetMetrics() SystemMetrics {
	return es.metrics
}

// GetProviderHealth returns health status of all providers
func (es *EnhancedSystem) GetProviderHealth() map[string]interface{} {
	health := make(map[string]interface{})
	for _, provider := range es.providers {
		health[provider.ID] = provider.HealthMetrics
	}
	return health
}

// updateMetrics updates system metrics
func (es *EnhancedSystem) updateMetrics(startTime time.Time, success bool) {
	es.metrics.TotalRequests++
	if success {
		es.metrics.SuccessfulRequests++
	} else {
		es.metrics.FailedRequests++
	}
	
	duration := time.Since(startTime)
	if es.metrics.TotalRequests == 1 {
		es.metrics.AverageLatency = duration
	} else {
		// Simple moving average
		es.metrics.AverageLatency = time.Duration(
			(int64(es.metrics.AverageLatency)*int64(es.metrics.TotalRequests-1) + int64(duration)) / int64(es.metrics.TotalRequests),
		)
	}
	
	es.metrics.LastUpdated = time.Now()
}

// MonitorHealth starts health monitoring for all providers
func (es *EnhancedSystem) MonitorHealth(ctx context.Context) error {
	return es.healthMonitor.StartMonitoring(ctx)
}

// GetProviderMetrics returns metrics for all providers
func (es *EnhancedSystem) GetProviderMetrics() map[string]ProviderMetrics {
	metrics := make(map[string]ProviderMetrics)
	for _, provider := range es.providers {
		metrics[provider.ID] = provider.Metrics
	}
	return metrics
}

// OptimizePrompt optimizes a prompt for better performance
func (es *EnhancedSystem) OptimizePrompt(ctx context.Context, prompt string, complexity TaskComplexity) (*OptimizedPrompt, error) {
	input := RequestInput{Query: prompt}
	return es.spoOptimizer.OptimizePrompt(ctx, input, &complexity)
}

// SelectProvider selects the best provider for a given task
func (es *EnhancedSystem) SelectProvider(ctx context.Context, input RequestInput, complexity TaskComplexity) (*ProviderAssignment, error) {
	return es.providerSelector.SelectProvider(ctx, input, &complexity)
}

// AnalyzeComplexity analyzes the complexity of a task
func (es *EnhancedSystem) AnalyzeComplexity(ctx context.Context, input RequestInput) (*TaskComplexity, error) {
	return es.taskReasoner.AnalyzeComplexity(ctx, input)
}

// UpdateProviderMetrics updates metrics for a specific provider
func (es *EnhancedSystem) UpdateProviderMetrics(providerID string, metrics ProviderMetrics) error {
	for _, provider := range es.providers {
		if provider.ID == providerID {
			provider.Metrics = metrics
			return nil
		}
	}
	return fmt.Errorf("provider %s not found", providerID)
}

// GetProvider returns a provider by ID
func (es *EnhancedSystem) GetProvider(providerID string) (*Provider, error) {
	for _, provider := range es.providers {
		if provider.ID == providerID {
			return provider, nil
		}
	}
	return nil, fmt.Errorf("provider %s not found", providerID)
}

// ListProviders returns all available providers
func (es *EnhancedSystem) ListProviders() []*Provider {
	return es.providers
}

// AddProvider adds a new provider to the system
func (es *EnhancedSystem) AddProvider(provider *Provider) {
	es.providers = append(es.providers, provider)
}

// RemoveProvider removes a provider from the system
func (es *EnhancedSystem) RemoveProvider(providerID string) error {
	for i, provider := range es.providers {
		if provider.ID == providerID {
			es.providers = append(es.providers[:i], es.providers[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("provider %s not found", providerID)
}

// Shutdown gracefully shuts down the enhanced system
func (es *EnhancedSystem) Shutdown(ctx context.Context) error {
	log.Println("Shutting down enhanced system...")
	
	// Stop health monitoring
	if es.healthMonitor != nil {
		// Add shutdown logic for health monitor if needed
	}
	
	log.Println("Enhanced system shutdown complete")
	return nil
}