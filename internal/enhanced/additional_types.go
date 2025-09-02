package enhanced

import (
	"context"
)

// EnhancedAdaptiveSelector is an interface for adaptive provider selection
type EnhancedAdaptiveSelector interface {
	SelectOptimalProvider(ctx context.Context, taskID string, complexity TaskComplexity, requirements map[string]interface{}) (ProviderAssignment, error)
	RecordProviderPerformance(taskID string, provider *Provider, actualCost, actualLatency, qualityScore float64, success bool)
	GetProviders() []*Provider
	Close() error
}