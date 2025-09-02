package enhanced

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

// EnhancedSystemV3 integrates dynamic model discovery with the enhanced system
type EnhancedSystemV3 struct {
	providers   []Provider
	selector    *EnhancedAdaptiveSelector
	modelLoader *DynamicModelLoader
	yamlLoader  *YAMLProviderLoader
}

// DynamicModelLoader handles dynamic model loading
type DynamicModelLoader struct {
	cache map[string][]string
}

// YAMLProviderLoader handles YAML provider configurations
type YAMLProviderLoader struct {
	providers map[string]Provider
}

// NewDynamicModelLoader creates a new dynamic model loader
func NewDynamicModelLoader() *DynamicModelLoader {
	return &DynamicModelLoader{
		cache: make(map[string][]string),
	}
}

// NewYAMLProviderLoader creates a new YAML provider loader
func NewYAMLProviderLoader() *YAMLProviderLoader {
	return &YAMLProviderLoader{
		providers: make(map[string]Provider),
	}
}

// LoadModelsFromSource loads models from a source URL
func (dml *DynamicModelLoader) LoadModelsFromSource(sourceURL string) ([]string, error) {
	// Mock implementation - in reality, this would fetch from the URL
	if cached, exists := dml.cache[sourceURL]; exists {
		return cached, nil
	}
	
	// Simulate loading models
	models := []string{"model-1", "model-2", "model-3"}
	dml.cache[sourceURL] = models
	return models, nil
}

// ClearCache clears the model cache
func (dml *DynamicModelLoader) ClearCache() {
	dml.cache = make(map[string][]string)
}

// GetCacheStats returns cache statistics
func (dml *DynamicModelLoader) GetCacheStats() map[string]interface{} {
	return map[string]interface{}{
		"cached_sources": len(dml.cache),
		"total_models":   0, // Would calculate total across all cached sources
	}
}

// RefreshAllModels refreshes all models in the YAML loader
func (ypl *YAMLProviderLoader) RefreshAllModels() {
	// Mock implementation
	log.Println("Refreshing YAML provider models...")
}

// LoadProvidersFromCSVWithDynamicModels loads providers from CSV with dynamic model discovery
func LoadProvidersFromCSVWithDynamicModels(csvPath string) ([]Provider, error) {
	// Mock implementation - in reality, this would parse the CSV
	providers := []Provider{
		{
			ID:           "pollinations",
			Name:         "Pollinations",
			Tier:         CommunityTier,
			BaseURL:      "https://text.pollinations.ai",
			Models:       []string{"openai", "creative"},
			Capabilities: []string{"text", "creative"},
			CostPerToken: 0.000001,
			MaxTokens:    2048,
		},
		{
			ID:           "openai",
			Name:         "OpenAI",
			Tier:         OfficialTier,
			BaseURL:      "https://api.openai.com/v1",
			Models:       []string{"gpt-4", "gpt-3.5-turbo"},
			Capabilities: []string{"text", "code", "reasoning"},
			CostPerToken: 0.03,
			MaxTokens:    4096,
		},
	}
	
	return providers, nil
}

// RefreshProviderModelsFromCSV refreshes models for providers from CSV
func RefreshProviderModelsFromCSV(providers []Provider) error {
	// Mock implementation
	log.Println("Refreshing provider models from CSV...")
	return nil
}

// NewEnhancedSystemV3 creates a new enhanced system with full dynamic model support
func NewEnhancedSystemV3(csvPath string, yamlDir string) (*EnhancedSystemV3, error) {
	log.Println("🚀 Initializing Enhanced System v3.0 with dynamic model discovery...")
	
	// Load providers with dynamic model discovery
	providers, err := LoadProvidersFromCSVWithDynamicModels(csvPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load providers: %w", err)
	}
	
	// Create enhanced selector
	selector, err := NewEnhancedAdaptiveSelector(providers)
	if err != nil {
		return nil, fmt.Errorf("failed to create selector: %w", err)
	}
	
	// Create enhanced system
	system := &EnhancedSystemV3{
		providers:   providers,
		selector:    selector,
		modelLoader: NewDynamicModelLoader(),
		yamlLoader:  NewYAMLProviderLoader(),
	}
	
	log.Printf("✅ Enhanced System v3.0 initialized with %d providers", len(providers))
	return system, nil
}

// ProcessRequest handles a request using enhanced provider selection with dynamic models
func (es *EnhancedSystemV3) ProcessRequest(request string) (*ProcessResponse, error) {
	log.Printf("🚀 Processing request with Enhanced System v3.0: %.100s...", request)

	// Analyze task complexity
	complexity := es.analyzeTaskComplexity(request)
	
	// Create metadata for selection
	metadata := map[string]interface{}{
		"request_length": len(request),
		"timestamp":      "2024-01-01T00:00:00Z", // placeholder
	}

	// Select the best provider using enhanced selection
	provider, confidence, err := es.selector.SelectProvider(complexity, metadata)
	if err != nil {
		return nil, fmt.Errorf("provider selection failed: %w", err)
	}

	log.Printf("✅ Enhanced selection: %s (confidence: %.2f%%) - Models: %v", 
		provider.Name, confidence*100, provider.Models)

	// Get detailed analysis for logging
	analysis := es.selector.AnalyzeRequest(request)
	if analysisJSON, err := json.MarshalIndent(analysis, "", "  "); err == nil {
		log.Printf("📊 Detailed analysis:\n%s", string(analysisJSON))
	}

	// Generate response
	response := es.generateResponse(provider, request, confidence)

	return response, nil
}

// analyzeTaskComplexity analyzes the complexity of a task
func (es *EnhancedSystemV3) analyzeTaskComplexity(request string) TaskComplexity {
	// Simple complexity analysis based on request content
	requestLower := strings.ToLower(request)
	wordCount := len(strings.Fields(request))
	
	complexity := TaskComplexity{
		Overall:     2.5, // default medium complexity
		Technical:   1.0,
		Creative:    1.0,
		Reasoning:   1.0,
		Computation: 1.0,
	}
	
	// Adjust based on content
	if strings.Contains(requestLower, "code") || strings.Contains(requestLower, "program") {
		complexity.Technical += 2.0
		complexity.Overall += 1.0
	}
	
	if strings.Contains(requestLower, "creative") || strings.Contains(requestLower, "story") {
		complexity.Creative += 2.0
		complexity.Overall += 0.5
	}
	
	if strings.Contains(requestLower, "complex") || wordCount > 50 {
		complexity.Overall += 1.5
		complexity.Reasoning += 1.0
	}
	
	// Normalize to 0-5 scale
	if complexity.Overall > 5.0 {
		complexity.Overall = 5.0
	}
	
	return complexity
}

// generateResponse creates a mock response from the selected provider
func (es *EnhancedSystemV3) generateResponse(provider *Provider, request string, confidence float64) *ProcessResponse {
	taskType := es.detectTaskType(request)
	
	var responseContent string
	var responseType string

	switch taskType {
	case TaskTypeImage:
		responseContent = fmt.Sprintf("🎨 [MOCK IMAGE GENERATION] Generated image using %s\nAvailable models: %v\nSelected based on image generation capabilities", 
			provider.Name, provider.Models)
		responseType = "image_generation"
		
	case TaskTypeCode:
		responseContent = fmt.Sprintf("💻 [MOCK CODE GENERATION] Generated code using %s\nAvailable models: %v\n\n```python\n# Example code generated by %s\ndef hello_world():\n    print('Hello from %s with dynamic models!')\n```", 
			provider.Name, provider.Models, provider.Name, provider.Name)
		responseType = "code_generation"
		
	case TaskTypeAudio:
		responseContent = fmt.Sprintf("🎵 [MOCK AUDIO PROCESSING] Processed audio using %s\nAvailable models: %v", 
			provider.Name, provider.Models)
		responseType = "audio_processing"
		
	case TaskTypeVideo:
		responseContent = fmt.Sprintf("🎬 [MOCK VIDEO PROCESSING] Processed video using %s\nAvailable models: %v", 
			provider.Name, provider.Models)
		responseType = "video_processing"
		
	case TaskTypeMultimodal:
		responseContent = fmt.Sprintf("🔄 [MOCK MULTIMODAL] Processed multimodal request using %s\nAvailable models: %v", 
			provider.Name, provider.Models)
		responseType = "multimodal"
		
	default: // TaskTypeText
		responseContent = fmt.Sprintf("📝 [MOCK TEXT GENERATION] Generated text response using %s\nAvailable models: %v\n\nThis is a mock response to: %s\n\nNote: Models were loaded dynamically from provider endpoints where available.", 
			provider.Name, provider.Models, request)
		responseType = "text_generation"
	}

	return &ProcessResponse{
		Content:    responseContent,
		Provider:   provider.Name,
		Models:     provider.Models,
		Confidence: confidence,
		Metadata: map[string]interface{}{
			"task_type":        string(taskType),
			"response_type":    responseType,
			"selection_method": "enhanced_v3_dynamic_models",
			"provider_url":     provider.BaseURL,
			"model_count":      len(provider.Models),
			"dynamic_loading":  es.hasDynamicModels(provider),
		},
	}
}

// detectTaskType detects the type of task from the request
func (es *EnhancedSystemV3) detectTaskType(request string) TaskType {
	requestLower := strings.ToLower(request)
	
	if strings.Contains(requestLower, "image") || strings.Contains(requestLower, "picture") || strings.Contains(requestLower, "photo") {
		return TaskTypeImage
	}
	if strings.Contains(requestLower, "code") || strings.Contains(requestLower, "program") || strings.Contains(requestLower, "function") {
		return TaskTypeCode
	}
	if strings.Contains(requestLower, "audio") || strings.Contains(requestLower, "sound") || strings.Contains(requestLower, "music") {
		return TaskTypeAudio
	}
	if strings.Contains(requestLower, "video") || strings.Contains(requestLower, "movie") {
		return TaskTypeVideo
	}
	if strings.Contains(requestLower, "multimodal") || strings.Contains(requestLower, "multi-modal") {
		return TaskTypeMultimodal
	}
	
	return TaskTypeText
}

// hasDynamicModels checks if a provider uses dynamic model loading
func (es *EnhancedSystemV3) hasDynamicModels(provider *Provider) bool {
	for _, model := range provider.Models {
		if strings.HasPrefix(model, "/") || strings.HasPrefix(model, "http://") || strings.HasPrefix(model, "https://") {
			return true
		}
	}
	return false
}

// RefreshAllModels refreshes models for all providers
func (es *EnhancedSystemV3) RefreshAllModels() error {
	log.Println("🔄 Refreshing all provider models...")
	
	// Clear caches
	es.modelLoader.ClearCache()
	es.yamlLoader.RefreshAllModels()
	
	// Refresh CSV-based providers
	err := RefreshProviderModelsFromCSV(es.providers)
	if err != nil {
		return fmt.Errorf("failed to refresh CSV provider models: %w", err)
	}
	
	// Update the selector with refreshed providers
	selector, err := NewEnhancedAdaptiveSelector(es.providers)
	if err != nil {
		return fmt.Errorf("failed to recreate selector: %w", err)
	}
	es.selector = selector
	
	log.Println("✅ All provider models refreshed successfully")
	return nil
}

// GetProviders returns the list of providers
func (es *EnhancedSystemV3) GetProviders() []Provider {
	return es.providers
}

// GetProviderCapabilities returns detailed capabilities for all providers
func (es *EnhancedSystemV3) GetProviderCapabilities() map[string]ProviderCapabilities {
	return es.selector.GetProviderCapabilities()
}

// GetDetailedModelCapabilities returns per-model capabilities
func (es *EnhancedSystemV3) GetDetailedModelCapabilities() map[string]map[string]ModelCapabilities {
	return es.selector.GetDetailedModelCapabilities()
}

// AnalyzeRequest provides detailed analysis of provider selection
func (es *EnhancedSystemV3) AnalyzeRequest(request string) map[string]interface{} {
	analysis := es.selector.AnalyzeRequest(request)
	
	// Add dynamic loading information
	analysis["dynamic_loading_stats"] = es.getDynamicLoadingStats()
	
	return analysis
}

// getDynamicLoadingStats returns statistics about dynamic model loading
func (es *EnhancedSystemV3) getDynamicLoadingStats() map[string]interface{} {
	stats := map[string]interface{}{
		"providers_with_dynamic_models": 0,
		"total_dynamic_endpoints":       0,
		"cache_stats":                   es.modelLoader.GetCacheStats(),
	}
	
	dynamicProviders := 0
	totalEndpoints := 0
	
	for _, provider := range es.providers {
		hasDynamic := false
		endpoints := 0
		
		for _, model := range provider.Models {
			if strings.HasPrefix(model, "/") || strings.HasPrefix(model, "http://") || strings.HasPrefix(model, "https://") {
				hasDynamic = true
				endpoints++
			}
		}
		
		if hasDynamic {
			dynamicProviders++
		}
		totalEndpoints += endpoints
	}
	
	stats["providers_with_dynamic_models"] = dynamicProviders
	stats["total_dynamic_endpoints"] = totalEndpoints
	
	return stats
}

// GetSystemInfo returns comprehensive system information
func (es *EnhancedSystemV3) GetSystemInfo() map[string]interface{} {
	providerCount := len(es.providers)
	totalModels := 0
	dynamicProviders := 0
	
	for _, provider := range es.providers {
		totalModels += len(provider.Models)
		if es.hasDynamicModels(&provider) {
			dynamicProviders++
		}
	}
	
	return map[string]interface{}{
		"version":              "3.0.0-dynamic-models",
		"provider_count":       providerCount,
		"dynamic_providers":    dynamicProviders,
		"total_models":         totalModels,
		"capabilities":         es.GetProviderCapabilities(),
		"dynamic_loading_stats": es.getDynamicLoadingStats(),
		"selection_method":     "enhanced_adaptive_with_dynamic_models",
		"features": []string{
			"dynamic_model_discovery",
			"csv_endpoint_integration", 
			"yaml_configuration_support",
			"model_database_integration",
			"huggingface_api_support",
			"per_model_capability_analysis",
			"enhanced_task_detection",
			"weighted_provider_scoring",
			"capability_validation",
			"real_time_model_refresh",
		},
	}
}

// ValidateProviders checks if all providers have valid configurations
func (es *EnhancedSystemV3) ValidateProviders() map[string][]string {
	issues := make(map[string][]string)
	
	for _, provider := range es.providers {
		var providerIssues []string
		
		// Check basic provider configuration
		if provider.Name == "" {
			providerIssues = append(providerIssues, "Provider name is empty")
		}
		
		if provider.BaseURL == "" {
			providerIssues = append(providerIssues, "Provider BaseURL is empty")
		}
		
		if len(provider.Models) == 0 {
			providerIssues = append(providerIssues, "Provider has no models configured")
		}
		
		// Check for dynamic model endpoints
		for _, model := range provider.Models {
			if strings.HasPrefix(model, "/") || strings.HasPrefix(model, "http://") || strings.HasPrefix(model, "https://") {
				// Try to validate the endpoint
				var modelURL string
				if strings.HasPrefix(model, "/") {
					baseURL := strings.TrimSuffix(provider.BaseURL, "/")
					modelURL = baseURL + model
				} else {
					modelURL = model
				}
				
				// Test if we can fetch from this endpoint (non-blocking)
				_, err := es.modelLoader.LoadModelsFromSource(modelURL)
				if err != nil {
					providerIssues = append(providerIssues, 
						fmt.Sprintf("Dynamic model endpoint unreachable: %s (%v)", modelURL, err))
				}
			}
		}
		
		if len(providerIssues) > 0 {
			issues[provider.Name] = providerIssues
		}
	}
	
	return issues
}

// SetSelectionWeights allows customizing provider selection criteria
func (es *EnhancedSystemV3) SetSelectionWeights(weights SelectionWeights) {
	es.selector.SetSelectionWeights(weights)
	log.Printf("⚖️ Selection weights updated: %+v", weights)
}