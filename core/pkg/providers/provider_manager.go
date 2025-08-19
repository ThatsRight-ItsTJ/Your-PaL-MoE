package providers

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/labring/aiproxy/core/pkg/pollinations"
)

type ProviderManager struct {
	providers    map[string]*ProviderConfig
	healthMonitor *HealthMonitor
	orchestrator *pollinations.Orchestrator
	mu           sync.RWMutex
	csvPath      string
	configDir    string
}

type ProviderStatus struct {
	Name         string        `json:"name"`
	Tier         string        `json:"tier"`
	Health       HealthStatus  `json:"health"`
	Capabilities []string      `json:"capabilities"`
	Models       []string      `json:"models"`
	LastCheck    time.Time     `json:"last_check"`
}

type RouterRequest struct {
	TaskType     string            `json:"task_type"`
	Prompt       string            `json:"prompt"`
	Model        string            `json:"model,omitempty"`
	Parameters   map[string]interface{} `json:"parameters,omitempty"`
	CostLimit    float64           `json:"cost_limit,omitempty"`
	QualityMin   int               `json:"quality_min,omitempty"`
	TierPreference []string        `json:"tier_preference,omitempty"`
}

type RouterResponse struct {
	Provider     string            `json:"provider"`
	Tier         string            `json:"tier"`
	Endpoint     string            `json:"endpoint"`
	Model        string            `json:"model"`
	EstimatedCost float64          `json:"estimated_cost"`
	QualityScore int               `json:"quality_score"`
	Reasoning    string            `json:"reasoning"`
	Auth         AuthConfig        `json:"auth"`
}

func NewProviderManager(csvPath, configDir string) *ProviderManager {
	return &ProviderManager{
		providers:     make(map[string]*ProviderConfig),
		healthMonitor: NewHealthMonitor(),
		orchestrator:  pollinations.NewOrchestrator(),
		csvPath:       csvPath,
		configDir:     configDir,
	}
}

func (pm *ProviderManager) Initialize(ctx context.Context) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Load providers from CSV
	parser := NewCSVParser(pm.csvPath)
	providers, err := parser.LoadProviders()
	if err != nil {
		return fmt.Errorf("failed to load providers: %w", err)
	}

	pm.providers = providers

	// Generate configurations if they don't exist
	configurator := NewAutoConfigurator(pm.csvPath, filepath.Join(pm.configDir, "generated", "providers"))
	if err := configurator.GenerateConfigurations(ctx); err != nil {
		return fmt.Errorf("failed to generate configurations: %w", err)
	}

	// Start health monitoring
	pm.healthMonitor.StartMonitoring(ctx, providers)

	return nil
}

func (pm *ProviderManager) RouteRequest(ctx context.Context, request RouterRequest) (*RouterResponse, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	// Get healthy providers
	candidates := pm.getHealthyProviders(request.TaskType)
	if len(candidates) == 0 {
		return nil, fmt.Errorf("no healthy providers available for task type: %s", request.TaskType)
	}

	// Apply tier preference filtering
	if len(request.TierPreference) > 0 {
		candidates = pm.filterByTierPreference(candidates, request.TierPreference)
	}

	// Apply cost and quality constraints
	candidates = pm.filterByCostAndQuality(candidates, request.CostLimit, request.QualityMin)

	if len(candidates) == 0 {
		return nil, fmt.Errorf("no providers match the specified constraints")
	}

	// Use orchestrator to select best provider
	bestProvider, err := pm.selectBestProvider(ctx, request, candidates)
	if err != nil {
		return nil, fmt.Errorf("failed to select best provider: %w", err)
	}

	// Build response
	response := &RouterResponse{
		Provider:      bestProvider.Name,
		Tier:          bestProvider.Tier,
		Endpoint:      bestProvider.Endpoint,
		Model:         request.Model,
		EstimatedCost: pm.estimateRequestCost(bestProvider, request),
		QualityScore:  pm.getProviderQualityScore(bestProvider),
		Auth:          bestProvider.Authentication,
		Reasoning:     fmt.Sprintf("Selected %s tier provider for optimal cost/quality balance", bestProvider.Tier),
	}

	return response, nil
}

func (pm *ProviderManager) getHealthyProviders(taskType string) []*ProviderConfig {
	var candidates []*ProviderConfig

	for _, provider := range pm.providers {
		// Check if provider supports the task type
		if pm.supportsTaskType(provider, taskType) && pm.isHealthy(provider) {
			candidates = append(candidates, provider)
		}
	}

	return candidates
}

func (pm *ProviderManager) supportsTaskType(provider *ProviderConfig, taskType string) bool {
	// This would normally check the provider's capabilities
	// For now, assume all providers support text generation
	switch taskType {
	case "text_generation":
		return true
	case "image_generation":
		// Check if provider name suggests image generation capability
		return pm.hasImageCapability(provider)
	default:
		return false
	}
}

func (pm *ProviderManager) hasImageCapability(provider *ProviderConfig) bool {
	name := provider.Name
	return name == "Bing DALL-E" || name == "Pollinations" || 
		   name == "OpenAI" // OpenAI supports DALL-E
}

func (pm *ProviderManager) isHealthy(provider *ProviderConfig) bool {
	return provider.Health.Status == "healthy" || provider.Health.Status == "unknown"
}

func (pm *ProviderManager) filterByTierPreference(providers []*ProviderConfig, tierPreference []string) []*ProviderConfig {
	var filtered []*ProviderConfig

	// Try each tier in order of preference
	for _, preferredTier := range tierPreference {
		for _, provider := range providers {
			if provider.Tier == preferredTier {
				filtered = append(filtered, provider)
			}
		}
		// If we found providers in this tier, return them
		if len(filtered) > 0 {
			return filtered
		}
	}

	// If no preferred tier found, return all providers
	return providers
}

func (pm *ProviderManager) filterByCostAndQuality(providers []*ProviderConfig, costLimit float64, qualityMin int) []*ProviderConfig {
	var filtered []*ProviderConfig

	for _, provider := range providers {
		estimatedCost := pm.getProviderBaseCost(provider)
		qualityScore := pm.getProviderQualityScore(provider)

		// Apply cost filter
		if costLimit > 0 && estimatedCost > costLimit {
			continue
		}

		// Apply quality filter
		if qualityMin > 0 && qualityScore < qualityMin {
			continue
		}

		filtered = append(filtered, provider)
	}

	return filtered
}

func (pm *ProviderManager) selectBestProvider(ctx context.Context, request RouterRequest, candidates []*ProviderConfig) (*ProviderConfig, error) {
	// Simple cost-optimized selection
	// Prefer free providers (community/unofficial) over paid (official)
	
	// Try unofficial first (free but potentially unstable)
	for _, provider := range candidates {
		if provider.Tier == "unofficial" {
			return provider, nil
		}
	}

	// Then community (free/low-cost, more stable)
	for _, provider := range candidates {
		if provider.Tier == "community" {
			return provider, nil
		}
	}

	// Finally official (paid but premium quality)
	for _, provider := range candidates {
		if provider.Tier == "official" {
			return provider, nil
		}
	}

	// Fallback to first available
	if len(candidates) > 0 {
		return candidates[0], nil
	}

	return nil, fmt.Errorf("no suitable provider found")
}

func (pm *ProviderManager) estimateRequestCost(provider *ProviderConfig, request RouterRequest) float64 {
	// Estimate cost based on tier and request type
	switch provider.Tier {
	case "official":
		switch request.TaskType {
		case "text_generation":
			return 0.002 // $0.002 per 1k tokens
		case "image_generation":
			return 0.04 // $0.04 per image
		default:
			return 0.001
		}
	case "community":
		return 0.0001 // Very low cost
	case "unofficial":
		return 0.0 // Free
	default:
		return 0.001
	}
}

func (pm *ProviderManager) getProviderBaseCost(provider *ProviderConfig) float64 {
	switch provider.Tier {
	case "official":
		return 0.002
	case "community":
		return 0.0001
	case "unofficial":
		return 0.0
	default:
		return 0.001
	}
}

func (pm *ProviderManager) getProviderQualityScore(provider *ProviderConfig) int {
	switch provider.Tier {
	case "official":
		return 9
	case "community":
		return 7
	case "unofficial":
		return 6
	default:
		return 5
	}
}

func (pm *ProviderManager) GetProviderStatus() []ProviderStatus {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	var statuses []ProviderStatus
	for _, provider := range pm.providers {
		status := ProviderStatus{
			Name:         provider.Name,
			Tier:         provider.Tier,
			Health:       provider.Health,
			Capabilities: provider.Capabilities,
			Models:       pm.getProviderModels(provider),
			LastCheck:    provider.Health.LastCheck,
		}
		statuses = append(statuses, status)
	}

	return statuses
}

func (pm *ProviderManager) getProviderModels(provider *ProviderConfig) []string {
	if models, ok := provider.ModelsSource.Value.([]string); ok {
		return models
	}
	return []string{}
}

func (pm *ProviderManager) RefreshProviders(ctx context.Context) error {
	return pm.Initialize(ctx)
}

func (pm *ProviderManager) GetProvider(name string) (*ProviderConfig, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	provider, exists := pm.providers[name]
	return provider, exists
}

func (pm *ProviderManager) UpdateProviderHealth(name string, health HealthStatus) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if provider, exists := pm.providers[name]; exists {
		provider.Health = health
		provider.LastUpdated = time.Now()
	}
}