package selection

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// ModelCapabilities represents detailed capabilities of a specific model
type ModelCapabilities struct {
	ModelName    string   `json:"model_name"`
	Text         bool     `json:"text"`
	Image        bool     `json:"image"`
	Code         bool     `json:"code"`
	Audio        bool     `json:"audio"`
	Video        bool     `json:"video"`
	Multimodal   bool     `json:"multimodal"`
	PipelineTag  string   `json:"pipeline_tag"`
	Tags         []string `json:"tags"`
	Reasoning    int      `json:"reasoning"`
	Knowledge    int      `json:"knowledge"`
	Computation  int      `json:"computation"`
	Confidence   float64  `json:"confidence"` // How confident we are in this assessment
	Source       string   `json:"source"`     // "huggingface", "manual", "provider_hint"
	LastUpdated  time.Time `json:"last_updated"`
}

// HuggingFaceModelInfo represents the response from HF API
type HuggingFaceModelInfo struct {
	ID          string   `json:"id"`
	PipelineTag string   `json:"pipeline_tag"`
	Tags        []string `json:"tags"`
	Downloads   int      `json:"downloads"`
	Likes       int      `json:"likes"`
	Library     []string `json:"library_name"`
}

// ModelDatabase manages model capability information from multiple sources
type ModelDatabase struct {
	modelCache    map[string]ModelCapabilities
	providerHints map[string][]string
	mutex         sync.RWMutex
	httpClient    *http.Client
	
	// Known model patterns for fallback
	knownModels   map[string]ModelCapabilities
}

// NewModelDatabase creates a new model database with enhanced capability detection
func NewModelDatabase() *ModelDatabase {
	db := &ModelDatabase{
		modelCache: make(map[string]ModelCapabilities),
		httpClient: &http.Client{Timeout: 10 * time.Second},
		knownModels: make(map[string]ModelCapabilities),
		providerHints: map[string][]string{
			"openai":           {"text", "code", "multimodal"},
			"anthropic":        {"text", "code"},
			"pollinations":     {"image", "text"},
			"together":         {"text", "code"},
			"local_ollama":     {"text", "code"},
			"huggingface":      {"text", "image", "code", "audio"},
			"stability":        {"image"},
			"midjourney":       {"image"},
			"elevenlabs":       {"audio"},
			"runwayml":         {"video"},
		},
	}
	
	// Initialize known models database
	db.initializeKnownModels()
	
	return db
}

// initializeKnownModels populates the database with known model capabilities
func (md *ModelDatabase) initializeKnownModels() {
	knownModels := map[string]ModelCapabilities{
		// OpenAI Models
		"gpt-4": {
			ModelName: "gpt-4", Text: true, Code: true, Multimodal: true,
			PipelineTag: "text-generation", Reasoning: 9, Knowledge: 9, Computation: 8,
			Confidence: 1.0, Source: "manual",
		},
		"gpt-4-turbo": {
			ModelName: "gpt-4-turbo", Text: true, Code: true, Multimodal: true,
			PipelineTag: "text-generation", Reasoning: 9, Knowledge: 9, Computation: 8,
			Confidence: 1.0, Source: "manual",
		},
		"gpt-3.5-turbo": {
			ModelName: "gpt-3.5-turbo", Text: true, Code: true,
			PipelineTag: "text-generation", Reasoning: 7, Knowledge: 8, Computation: 7,
			Confidence: 1.0, Source: "manual",
		},
		"dall-e-3": {
			ModelName: "dall-e-3", Image: true,
			PipelineTag: "text-to-image", Reasoning: 6, Knowledge: 7, Computation: 8,
			Confidence: 1.0, Source: "manual",
		},
		
		// Anthropic Models
		"claude-3-5-sonnet": {
			ModelName: "claude-3-5-sonnet", Text: true, Code: true,
			PipelineTag: "text-generation", Reasoning: 9, Knowledge: 9, Computation: 8,
			Confidence: 1.0, Source: "manual",
		},
		"claude-3-haiku": {
			ModelName: "claude-3-haiku", Text: true, Code: true,
			PipelineTag: "text-generation", Reasoning: 7, Knowledge: 8, Computation: 7,
			Confidence: 1.0, Source: "manual",
		},
		"claude-3-opus": {
			ModelName: "claude-3-opus", Text: true, Code: true,
			PipelineTag: "text-generation", Reasoning: 9, Knowledge: 9, Computation: 8,
			Confidence: 1.0, Source: "manual",
		},
		
		// Code Models
		"qwen-coder": {
			ModelName: "qwen-coder", Text: true, Code: true,
			PipelineTag: "text-generation", Reasoning: 7, Knowledge: 7, Computation: 9,
			Confidence: 1.0, Source: "manual",
		},
		"codellama": {
			ModelName: "codellama", Text: true, Code: true,
			PipelineTag: "text-generation", Reasoning: 6, Knowledge: 6, Computation: 9,
			Confidence: 1.0, Source: "manual",
		},
		
		// Image Models
		"cogview-3-flash": {
			ModelName: "cogview-3-flash", Image: true,
			PipelineTag: "text-to-image", Reasoning: 5, Knowledge: 6, Computation: 7,
			Confidence: 1.0, Source: "manual",
		},
		"stable-diffusion": {
			ModelName: "stable-diffusion", Image: true,
			PipelineTag: "text-to-image", Reasoning: 5, Knowledge: 6, Computation: 7,
			Confidence: 1.0, Source: "manual",
		},
		"midjourney": {
			ModelName: "midjourney", Image: true,
			PipelineTag: "text-to-image", Reasoning: 6, Knowledge: 7, Computation: 7,
			Confidence: 1.0, Source: "manual",
		},
		
		// Multimodal Models
		"gpt-4v": {
			ModelName: "gpt-4v", Text: true, Image: true, Multimodal: true,
			PipelineTag: "image-to-text", Reasoning: 9, Knowledge: 9, Computation: 8,
			Confidence: 1.0, Source: "manual",
		},
		"gemini-pro-vision": {
			ModelName: "gemini-pro-vision", Text: true, Image: true, Multimodal: true,
			PipelineTag: "image-to-text", Reasoning: 8, Knowledge: 8, Computation: 7,
			Confidence: 1.0, Source: "manual",
		},
		
		// Audio Models
		"whisper": {
			ModelName: "whisper", Audio: true,
			PipelineTag: "automatic-speech-recognition", Reasoning: 5, Knowledge: 6, Computation: 6,
			Confidence: 1.0, Source: "manual",
		},
		
		// Open Source Models
		"llama-2-70b": {
			ModelName: "llama-2-70b", Text: true, Code: true,
			PipelineTag: "text-generation", Reasoning: 8, Knowledge: 8, Computation: 7,
			Confidence: 1.0, Source: "manual",
		},
		"llama-2-13b": {
			ModelName: "llama-2-13b", Text: true, Code: true,
			PipelineTag: "text-generation", Reasoning: 7, Knowledge: 7, Computation: 6,
			Confidence: 1.0, Source: "manual",
		},
		"mistral-7b": {
			ModelName: "mistral-7b", Text: true, Code: true,
			PipelineTag: "text-generation", Reasoning: 6, Knowledge: 7, Computation: 6,
			Confidence: 1.0, Source: "manual",
		},
	}
	
	md.mutex.Lock()
	defer md.mutex.Unlock()
	
	for name, capabilities := range knownModels {
		capabilities.LastUpdated = time.Now()
		md.knownModels[name] = capabilities
		md.modelCache[name] = capabilities
	}
}

// GetModelCapabilities retrieves capabilities for a specific model using multiple sources
func (md *ModelDatabase) GetModelCapabilities(modelName, providerName string) ModelCapabilities {
	// Normalize model name
	normalizedName := strings.ToLower(strings.TrimSpace(modelName))
	
	// 1. Check cache first
	md.mutex.RLock()
	if cached, exists := md.modelCache[normalizedName]; exists {
		// Check if cache is still fresh (24 hours)
		if time.Since(cached.LastUpdated) < 24*time.Hour {
			md.mutex.RUnlock()
			return cached
		}
	}
	md.mutex.RUnlock()
	
	// 2. Check known models database
	if known, exists := md.knownModels[normalizedName]; exists {
		md.cacheModel(normalizedName, known)
		return known
	}
	
	// 3. Try Hugging Face API
	if hfCapabilities := md.queryHuggingFace(normalizedName); hfCapabilities.Confidence > 0 {
		md.cacheModel(normalizedName, hfCapabilities)
		return hfCapabilities
	}
	
	// 4. Use enhanced pattern matching
	if patternCapabilities := md.detectFromPatterns(normalizedName); patternCapabilities.Confidence > 0 {
		md.cacheModel(normalizedName, patternCapabilities)
		return patternCapabilities
	}
	
	// 5. Fallback to provider hints
	providerCapabilities := md.inferFromProvider(normalizedName, providerName)
	md.cacheModel(normalizedName, providerCapabilities)
	
	return providerCapabilities
}

// queryHuggingFace queries the Hugging Face Hub API for model information
func (md *ModelDatabase) queryHuggingFace(modelName string) ModelCapabilities {
	url := fmt.Sprintf("https://huggingface.co/api/models/%s", modelName)
	
	resp, err := md.httpClient.Get(url)
	if err != nil || resp.StatusCode != 200 {
		return ModelCapabilities{ModelName: modelName, Confidence: 0}
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ModelCapabilities{ModelName: modelName, Confidence: 0}
	}
	
	var hfModel HuggingFaceModelInfo
	if err := json.Unmarshal(body, &hfModel); err != nil {
		return ModelCapabilities{ModelName: modelName, Confidence: 0}
	}
	
	return md.parseHuggingFaceModel(hfModel)
}

// parseHuggingFaceModel converts HF model info to our capability format
func (md *ModelDatabase) parseHuggingFaceModel(hfModel HuggingFaceModelInfo) ModelCapabilities {
	capabilities := ModelCapabilities{
		ModelName:   hfModel.ID,
		PipelineTag: hfModel.PipelineTag,
		Tags:        hfModel.Tags,
		Confidence:  0.8, // High confidence from HF API
		Source:      "huggingface",
		LastUpdated: time.Now(),
	}
	
	// Map pipeline tags to capabilities
	switch hfModel.PipelineTag {
	case "text-generation", "text2text-generation":
		capabilities.Text = true
		capabilities.Reasoning = 7
		capabilities.Knowledge = 7
		capabilities.Computation = 6
	case "text-to-image", "image-to-image":
		capabilities.Image = true
		capabilities.Reasoning = 5
		capabilities.Knowledge = 6
		capabilities.Computation = 7
	case "image-to-text", "image-classification":
		capabilities.Image = true
		capabilities.Multimodal = true
		capabilities.Reasoning = 6
		capabilities.Knowledge = 7
		capabilities.Computation = 6
	case "automatic-speech-recognition", "text-to-speech":
		capabilities.Audio = true
		capabilities.Reasoning = 5
		capabilities.Knowledge = 6
		capabilities.Computation = 6
	case "video-classification":
		capabilities.Video = true
		capabilities.Reasoning = 5
		capabilities.Knowledge = 6
		capabilities.Computation = 7
	default:
		// Check tags for additional hints
		tagStr := strings.ToLower(strings.Join(hfModel.Tags, " "))
		if strings.Contains(tagStr, "code") || strings.Contains(tagStr, "programming") {
			capabilities.Code = true
			capabilities.Text = true
			capabilities.Computation = 8
		} else {
			capabilities.Text = true // Default fallback
		}
	}
	
	// Enhance based on model tags
	tagStr := strings.ToLower(strings.Join(hfModel.Tags, " "))
	if strings.Contains(tagStr, "multimodal") {
		capabilities.Multimodal = true
	}
	if strings.Contains(tagStr, "reasoning") {
		capabilities.Reasoning = min(capabilities.Reasoning+2, 10)
	}
	
	return capabilities
}

// detectFromPatterns uses enhanced pattern matching for model detection
func (md *ModelDatabase) detectFromPatterns(modelName string) ModelCapabilities {
	modelLower := strings.ToLower(modelName)
	capabilities := ModelCapabilities{
		ModelName:   modelName,
		Source:      "pattern_matching",
		LastUpdated: time.Now(),
		Confidence:  0.6,
	}
	
	// Enhanced pattern matching with specific model families
	patterns := map[string]func(*ModelCapabilities){
		// Text/Chat models
		"gpt":     func(c *ModelCapabilities) { c.Text, c.Code = true, true; c.Reasoning, c.Knowledge = 8, 8 },
		"claude":  func(c *ModelCapabilities) { c.Text, c.Code = true, true; c.Reasoning, c.Knowledge = 9, 9 },
		"llama":   func(c *ModelCapabilities) { c.Text, c.Code = true, true; c.Reasoning, c.Knowledge = 7, 7 },
		"mistral": func(c *ModelCapabilities) { c.Text, c.Code = true, true; c.Reasoning, c.Knowledge = 7, 7 },
		"gemini":  func(c *ModelCapabilities) { c.Text, c.Code = true, true; c.Reasoning, c.Knowledge = 8, 8 },
		"qwen":    func(c *ModelCapabilities) { c.Text, c.Code = true, true; c.Reasoning, c.Knowledge = 7, 7 },
		
		// Code-specific models
		"coder":     func(c *ModelCapabilities) { c.Text, c.Code = true, true; c.Computation = 9 },
		"codellama": func(c *ModelCapabilities) { c.Text, c.Code = true, true; c.Computation = 9 },
		"starcoder": func(c *ModelCapabilities) { c.Text, c.Code = true, true; c.Computation = 9 },
		"copilot":   func(c *ModelCapabilities) { c.Text, c.Code = true, true; c.Computation = 9 },
		
		// Image models
		"dall-e":           func(c *ModelCapabilities) { c.Image = true; c.Reasoning = 6 },
		"dalle":            func(c *ModelCapabilities) { c.Image = true; c.Reasoning = 6 },
		"stable-diffusion": func(c *ModelCapabilities) { c.Image = true; c.Reasoning = 5 },
		"midjourney":       func(c *ModelCapabilities) { c.Image = true; c.Reasoning = 6 },
		"cogview":          func(c *ModelCapabilities) { c.Image = true; c.Reasoning = 5 },
		"imagen":           func(c *ModelCapabilities) { c.Image = true; c.Reasoning = 6 },
		"firefly":          func(c *ModelCapabilities) { c.Image = true; c.Reasoning = 5 },
		
		// Audio models
		"whisper": func(c *ModelCapabilities) { c.Audio = true; c.Reasoning = 5 },
		"speech":  func(c *ModelCapabilities) { c.Audio = true; c.Reasoning = 5 },
		"voice":   func(c *ModelCapabilities) { c.Audio = true; c.Reasoning = 5 },
		
		// Video models
		"video": func(c *ModelCapabilities) { c.Video = true; c.Reasoning = 5 },
		"movie": func(c *ModelCapabilities) { c.Video = true; c.Reasoning = 5 },
		
		// Multimodal indicators
		"vision":     func(c *ModelCapabilities) { c.Multimodal, c.Text, c.Image = true, true, true },
		"multimodal": func(c *ModelCapabilities) { c.Multimodal, c.Text = true, true },
	}
	
	matched := false
	for pattern, applyFunc := range patterns {
		if strings.Contains(modelLower, pattern) {
			applyFunc(&capabilities)
			matched = true
			capabilities.Confidence = 0.7 // Higher confidence for pattern match
		}
	}
	
	if !matched {
		// Default fallback
		capabilities.Text = true
		capabilities.Reasoning = 5
		capabilities.Knowledge = 6
		capabilities.Computation = 5
		capabilities.Confidence = 0.3
	}
	
	return capabilities
}

// inferFromProvider uses provider context to infer model capabilities
func (md *ModelDatabase) inferFromProvider(modelName, providerName string) ModelCapabilities {
	capabilities := ModelCapabilities{
		ModelName:   modelName,
		Source:      "provider_hint",
		LastUpdated: time.Now(),
		Confidence:  0.4,
		Reasoning:   5,
		Knowledge:   6,
		Computation: 5,
	}
	
	providerLower := strings.ToLower(providerName)
	
	// Provider-specific capability inference
	switch {
	case strings.Contains(providerLower, "openai"):
		capabilities.Text, capabilities.Code = true, true
		capabilities.Reasoning, capabilities.Knowledge = 8, 8
	case strings.Contains(providerLower, "anthropic"):
		capabilities.Text, capabilities.Code = true, true
		capabilities.Reasoning, capabilities.Knowledge = 9, 9
	case strings.Contains(providerLower, "pollinations"):
		if strings.Contains(providerLower, "text") {
			capabilities.Text = true
		} else {
			capabilities.Image = true
		}
	case strings.Contains(providerLower, "stability"):
		capabilities.Image = true
	case strings.Contains(providerLower, "midjourney"):
		capabilities.Image = true
	case strings.Contains(providerLower, "elevenlabs"):
		capabilities.Audio = true
	case strings.Contains(providerLower, "runway"):
		capabilities.Video = true
	default:
		// Generic provider - assume text capability
		capabilities.Text = true
	}
	
	return capabilities
}

// cacheModel stores model capabilities in cache
func (md *ModelDatabase) cacheModel(modelName string, capabilities ModelCapabilities) {
	md.mutex.Lock()
	defer md.mutex.Unlock()
	
	capabilities.LastUpdated = time.Now()
	md.modelCache[modelName] = capabilities
}

// GetProviderCapabilities analyzes all models for a provider to determine overall capabilities
func (md *ModelDatabase) GetProviderCapabilities(models []string, providerName string) ProviderCapabilities {
	aggregated := ProviderCapabilities{
		Models: models,
	}
	
	totalReasoning := 0
	totalKnowledge := 0
	totalComputation := 0
	modelCount := len(models)
	
	if modelCount == 0 {
		// No models specified, infer from provider
		return md.inferProviderCapabilities(providerName)
	}
	
	// Analyze each model
	for _, model := range models {
		modelCap := md.GetModelCapabilities(model, providerName)
		
		// Aggregate boolean capabilities (OR logic)
		aggregated.Text = aggregated.Text || modelCap.Text
		aggregated.Image = aggregated.Image || modelCap.Image
		aggregated.Code = aggregated.Code || modelCap.Code
		aggregated.Audio = aggregated.Audio || modelCap.Audio
		aggregated.Video = aggregated.Video || modelCap.Video
		aggregated.Multimodal = aggregated.Multimodal || modelCap.Multimodal
		
		// Aggregate numeric capabilities (average)
		totalReasoning += modelCap.Reasoning
		totalKnowledge += modelCap.Knowledge
		totalComputation += modelCap.Computation
	}
	
	// Calculate averages
	aggregated.Reasoning = totalReasoning / modelCount
	aggregated.Knowledge = totalKnowledge / modelCount
	aggregated.Computation = totalComputation / modelCount
	
	return aggregated
}

// inferProviderCapabilities provides fallback capabilities based on provider name
func (md *ModelDatabase) inferProviderCapabilities(providerName string) ProviderCapabilities {
	capabilities := ProviderCapabilities{
		Reasoning:   5,
		Knowledge:   6,
		Computation: 5,
	}
	
	providerLower := strings.ToLower(providerName)
	
	// Provider-specific defaults
	if hints, exists := md.providerHints[providerLower]; exists {
		for _, hint := range hints {
			switch hint {
			case "text":
				capabilities.Text = true
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
	} else {
		// Default to text if unknown
		capabilities.Text = true
	}
	
	return capabilities
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ClearCache clears the model cache
func (md *ModelDatabase) ClearCache() {
	md.mutex.Lock()
	defer md.mutex.Unlock()
	
	md.modelCache = make(map[string]ModelCapabilities)
}

// GetCacheStats returns cache statistics
func (md *ModelDatabase) GetCacheStats() map[string]interface{} {
	md.mutex.RLock()
	defer md.mutex.RUnlock()
	
	return map[string]interface{}{
		"cached_models": len(md.modelCache),
		"known_models":  len(md.knownModels),
		"cache_size":    len(md.modelCache),
	}
}