package enhanced

import (
	"time"
)

// ProcessResponse represents the response from processing a request
type ProcessResponse struct {
	Content    string                 `json:"content"`
	Provider   string                 `json:"provider"`
	Models     []string               `json:"models"`
	Confidence float64                `json:"confidence"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// Provider represents an AI provider with its configuration
type Provider struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Tier         ProviderTier           `json:"tier"`
	BaseURL      string                 `json:"base_url"`
	APIKey       string                 `json:"api_key,omitempty"`
	Models       []string               `json:"models"`
	Capabilities []string               `json:"capabilities"`
	CostPerToken float64                `json:"cost_per_token"`
	MaxTokens    int                    `json:"max_tokens"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// ProviderTier represents the tier of a provider
type ProviderTier string

const (
	OfficialTier   ProviderTier = "official"
	CommunityTier  ProviderTier = "community"
	UnofficialTier ProviderTier = "unofficial"
)

// RequestInput represents input for processing
type RequestInput struct {
	ID        string                 `json:"id"`
	Content   string                 `json:"content"`
	Type      string                 `json:"type,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// TaskType represents different types of tasks
type TaskType string

const (
	TaskTypeText       TaskType = "text"
	TaskTypeCode       TaskType = "code"
	TaskTypeImage      TaskType = "image"
	TaskTypeAudio      TaskType = "audio"
	TaskTypeVideo      TaskType = "video"
	TaskTypeMultimodal TaskType = "multimodal"
)

// ProviderCapabilities represents capabilities of a provider
type ProviderCapabilities struct {
	Text        bool `json:"text"`
	Image       bool `json:"image"`
	Code        bool `json:"code"`
	Audio       bool `json:"audio"`
	Video       bool `json:"video"`
	Multimodal  bool `json:"multimodal"`
	Reasoning   int  `json:"reasoning"`   // 0-10 scale
	Knowledge   int  `json:"knowledge"`   // 0-10 scale
	Computation int  `json:"computation"` // 0-10 scale
}

// ModelCapabilities represents capabilities of a specific model
type ModelCapabilities struct {
	Name         string   `json:"name"`
	Type         string   `json:"type"`
	MaxTokens    int      `json:"max_tokens"`
	CostPerToken float64  `json:"cost_per_token"`
	Capabilities []string `json:"capabilities"`
	Performance  int      `json:"performance"` // 0-10 scale
}

// SelectionWeights represents weights for provider selection
type SelectionWeights struct {
	Performance float64 `json:"performance"`
	Cost        float64 `json:"cost"`
	Reliability float64 `json:"reliability"`
	Speed       float64 `json:"speed"`
}

// TaskComplexity represents the complexity analysis of a task
type TaskComplexity struct {
	Overall     float64 `json:"overall"`
	Technical   float64 `json:"technical"`
	Creative    float64 `json:"creative"`
	Reasoning   float64 `json:"reasoning"`
	Computation float64 `json:"computation"`
}

// EnhancedAdaptiveSelector represents an enhanced provider selector
type EnhancedAdaptiveSelector struct {
	providers    []Provider
	weights      SelectionWeights
	capabilities map[string]ProviderCapabilities
	models       map[string]map[string]ModelCapabilities
}

// ProviderPricing represents pricing information for a provider
type ProviderPricing struct {
	ProviderID      string  `json:"provider_id"`
	ModelID         string  `json:"model_id"`
	InputCost       float64 `json:"input_cost"`
	OutputCost      float64 `json:"output_cost"`
	Currency        string  `json:"currency"`
	BillingUnit     string  `json:"billing_unit"`
	MinimumCharge   float64 `json:"minimum_charge"`
	FreeTokens      int     `json:"free_tokens"`
	SubscriptionFee float64 `json:"subscription_fee"`
}

// ProviderAssignment represents an assignment of a provider to a task
type ProviderAssignment struct {
	TaskID       string                 `json:"task_id"`
	ProviderID   string                 `json:"provider_id"`
	ModelID      string                 `json:"model_id"`
	Confidence   float64                `json:"confidence"`
	Reasoning    string                 `json:"reasoning"`
	EstimatedCost float64               `json:"estimated_cost"`
	Priority     int                    `json:"priority"`
	Metadata     map[string]interface{} `json:"metadata"`
	AssignedAt   time.Time              `json:"assigned_at"`
}

// AlternativeProvider represents an alternative provider option
type AlternativeProvider struct {
	Provider   Provider `json:"provider"`
	Confidence float64  `json:"confidence"`
	Reason     string   `json:"reason"`
	Cost       float64  `json:"cost"`
	Ranking    int      `json:"ranking"`
}

// OptimizedPrompt represents an optimized prompt for better results
type OptimizedPrompt struct {
	Original     string                 `json:"original"`
	Optimized    string                 `json:"optimized"`
	Improvements []string               `json:"improvements"`
	Confidence   float64                `json:"confidence"`
	Strategy     string                 `json:"strategy"`
	Metadata     map[string]interface{} `json:"metadata"`
	CreatedAt    time.Time              `json:"created_at"`
}

// ComplexityLevel represents the complexity level of a task
type ComplexityLevel string

const (
	ComplexityLow      ComplexityLevel = "low"
	ComplexityMedium   ComplexityLevel = "medium"
	ComplexityHigh     ComplexityLevel = "high"
	ComplexityCritical ComplexityLevel = "critical"
)

// NewEnhancedAdaptiveSelector creates a new enhanced adaptive selector
func NewEnhancedAdaptiveSelector(providers []Provider) (*EnhancedAdaptiveSelector, error) {
	selector := &EnhancedAdaptiveSelector{
		providers:    providers,
		capabilities: make(map[string]ProviderCapabilities),
		models:       make(map[string]map[string]ModelCapabilities),
		weights: SelectionWeights{
			Performance: 0.4,
			Cost:        0.3,
			Reliability: 0.2,
			Speed:       0.1,
		},
	}
	
	// Initialize capabilities for each provider
	for _, provider := range providers {
		capabilities := ProviderCapabilities{
			Text:        true, // default capability
			Reasoning:   5,    // default medium reasoning
			Knowledge:   5,    // default medium knowledge
			Computation: 5,    // default medium computation
		}
		
		// Set capabilities based on provider capabilities
		for _, cap := range provider.Capabilities {
			switch cap {
			case "image":
				capabilities.Image = true
			case "code":
				capabilities.Code = true
			case "audio":
				capabilities.Audio = true
			case "video":
				capabilities.Video = true
			case "multimodal":
				capabilities.Multimodal = true
			}
		}
		
		selector.capabilities[provider.ID] = capabilities
		
		// Initialize model capabilities
		modelCaps := make(map[string]ModelCapabilities)
		for _, model := range provider.Models {
			modelCaps[model] = ModelCapabilities{
				Name:         model,
				Type:         "text",
				MaxTokens:    provider.MaxTokens,
				CostPerToken: provider.CostPerToken,
				Capabilities: provider.Capabilities,
				Performance:  5, // default medium performance
			}
		}
		selector.models[provider.ID] = modelCaps
	}
	
	return selector, nil
}

// SelectProvider selects the best provider for a given task
func (s *EnhancedAdaptiveSelector) SelectProvider(complexity TaskComplexity, metadata map[string]interface{}) (*Provider, float64, error) {
	if len(s.providers) == 0 {
		return nil, 0, fmt.Errorf("no providers available")
	}
	
	bestProvider := &s.providers[0]
	bestScore := 0.0
	
	for i := range s.providers {
		provider := &s.providers[i]
		score := s.calculateProviderScore(provider, complexity, metadata)
		
		if score > bestScore {
			bestScore = score
			bestProvider = provider
		}
	}
	
	confidence := bestScore / 10.0 // normalize to 0-1
	if confidence > 1.0 {
		confidence = 1.0
	}
	
	return bestProvider, confidence, nil
}

// calculateProviderScore calculates a score for a provider based on task requirements
func (s *EnhancedAdaptiveSelector) calculateProviderScore(provider *Provider, complexity TaskComplexity, metadata map[string]interface{}) float64 {
	capabilities, exists := s.capabilities[provider.ID]
	if !exists {
		return 1.0 // default low score
	}
	
	score := 0.0
	
	// Performance scoring
	performanceScore := float64(capabilities.Reasoning+capabilities.Knowledge+capabilities.Computation) / 3.0
	score += performanceScore * s.weights.Performance
	
	// Cost scoring (lower cost = higher score)
	costScore := 10.0 - (provider.CostPerToken * 1000000) // normalize cost
	if costScore < 0 {
		costScore = 0
	}
	score += costScore * s.weights.Cost
	
	// Reliability scoring (based on tier)
	reliabilityScore := 5.0
	switch provider.Tier {
	case OfficialTier:
		reliabilityScore = 9.0
	case CommunityTier:
		reliabilityScore = 6.0
	case UnofficialTier:
		reliabilityScore = 3.0
	}
	score += reliabilityScore * s.weights.Reliability
	
	// Speed scoring (assume higher max tokens = faster)
	speedScore := float64(provider.MaxTokens) / 1000.0
	if speedScore > 10.0 {
		speedScore = 10.0
	}
	score += speedScore * s.weights.Speed
	
	return score
}

// AnalyzeRequest analyzes a request and returns detailed information
func (s *EnhancedAdaptiveSelector) AnalyzeRequest(request string) map[string]interface{} {
	return map[string]interface{}{
		"request_length":   len(request),
		"provider_count":   len(s.providers),
		"selection_method": "enhanced_adaptive",
		"weights":          s.weights,
	}
}

// GetProviderCapabilities returns capabilities for all providers
func (s *EnhancedAdaptiveSelector) GetProviderCapabilities() map[string]ProviderCapabilities {
	return s.capabilities
}

// GetDetailedModelCapabilities returns detailed model capabilities
func (s *EnhancedAdaptiveSelector) GetDetailedModelCapabilities() map[string]map[string]ModelCapabilities {
	return s.models
}

// SetSelectionWeights sets the selection weights
func (s *EnhancedAdaptiveSelector) SetSelectionWeights(weights SelectionWeights) {
	s.weights = weights
}