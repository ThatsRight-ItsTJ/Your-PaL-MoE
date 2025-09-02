package enhanced

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/ThatsRight-ItsTJ/Your-PaL-MoE/internal/components"
)

// NewEnhancedSystem creates a new enhanced system with default configuration
func NewEnhancedSystem(providers []*Provider) *EnhancedSystem {
	return &EnhancedSystem{
		selector:      NewEnhancedProviderSelector(providers),
		reasoner:      components.NewTaskReasoner(),
		optimizer:     components.NewSPOOptimizer(),
		healthMonitor: NewProviderHealthMonitor(),
		providers:     providers,
		metrics:       NewSystemMetrics(),
	}
}

// ProcessRequest processes a request using the enhanced system
func (es *EnhancedSystem) ProcessRequest(ctx context.Context, input RequestInput) (*ProcessResponse, error) {
	startTime := time.Now()

	// Analyze task complexity
	complexity, err := es.reasoner.AnalyzeComplexity(input.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze complexity: %w", err)
	}

	// Optimize prompt
	optimizedPrompt, err := es.optimizer.OptimizePrompt(input.Content, *complexity)
	if err != nil {
		log.Printf("Failed to optimize prompt: %v", err)
		// Continue with original prompt
		optimizedPrompt = input.Content
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
		Content:        fmt.Sprintf("Processed by %s using model %s: %s", assignment.Provider.Name, assignment.Model, optimizedPrompt),
		Provider:       assignment.Provider,
		Model:          assignment.Model,
		Complexity:     *complexity,
		ProcessingTime: time.Since(startTime),
		TokensUsed:     complexity.TokenEstimate,
		Cost:           assignment.EstimatedCost,
		Metadata:       make(map[string]interface{}),
	}

	if optimizedPrompt != input.Content {
		response.Metadata["optimized_prompt"] = optimizedPrompt
		response.Metadata["original_prompt"] = input.Content
	}

	// Update provider health metrics
	es.healthMonitor.UpdateMetrics(assignment.Provider.Name, true, time.Since(startTime))

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
				string(provider.Tier),
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
	complexity, err := es.reasoner.AnalyzeComplexity(input.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze complexity: %w", err)
	}

	// Optimize prompt
	optimizedText, err := es.optimizer.OptimizePrompt(input.Content, *complexity)
	if err != nil {
		return nil, err
	}

	return &OptimizedPrompt{
		OriginalPrompt:    input.Content,
		OptimizedPrompt:   optimizedText,
		Complexity:        complexity.Overall,
		OptimizationRules: []string{"complexity-based optimization"},
		Metadata:          make(map[string]interface{}),
		ProcessingTime:    time.Since(time.Now()),
	}, nil
}

// AnalyzeComplexityOnly analyzes task complexity without full processing
func (es *EnhancedSystem) AnalyzeComplexityOnly(ctx context.Context, input RequestInput) (*TaskComplexity, error) {
	return es.reasoner.AnalyzeComplexity(input.Content)
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