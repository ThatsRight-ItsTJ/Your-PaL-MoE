package selection

import (
	"strings"
)

// EnhancedCapabilityDetector uses model database for accurate capability detection
type EnhancedCapabilityDetector struct {
	modelDB *ModelDatabase
	// Keep original patterns as fallback
	*CapabilityDetector
}

// NewEnhancedCapabilityDetector creates a new enhanced capability detector with model database
func NewEnhancedCapabilityDetector() *EnhancedCapabilityDetector {
	return &EnhancedCapabilityDetector{
		modelDB:            NewModelDatabase(),
		CapabilityDetector: NewCapabilityDetector(),
	}
}

// DetectCapabilities analyzes models using the enhanced model database approach
func (ecd *EnhancedCapabilityDetector) DetectCapabilities(models []string, providerName string) ProviderCapabilities {
	// Use model database for accurate per-model analysis
	return ecd.modelDB.GetProviderCapabilities(models, providerName)
}

// DetectModelCapabilities gets capabilities for a specific model
func (ecd *EnhancedCapabilityDetector) DetectModelCapabilities(modelName, providerName string) ModelCapabilities {
	return ecd.modelDB.GetModelCapabilities(modelName, providerName)
}

// IsProviderCompatible checks if a provider can handle a specific task type (enhanced version)
func (ecd *EnhancedCapabilityDetector) IsProviderCompatible(capabilities ProviderCapabilities, taskType TaskType) bool {
	switch taskType {
	case TaskTypeText:
		return capabilities.Text
	case TaskTypeImage:
		return capabilities.Image
	case TaskTypeCode:
		// Enhanced: Code models are preferred, but high-quality text models can handle code too
		if capabilities.Code {
			return true
		}
		// Allow high-reasoning text models to handle code
		return capabilities.Text && capabilities.Reasoning >= 7
	case TaskTypeAudio:
		return capabilities.Audio
	case TaskTypeVideo:
		return capabilities.Video
	case TaskTypeMultimodal:
		// Enhanced: More flexible multimodal detection
		if capabilities.Multimodal {
			return true
		}
		// Allow combination of text and image capabilities
		return capabilities.Text && capabilities.Image
	default:
		return capabilities.Text // Default to text compatibility
	}
}

// DetectTaskType analyzes request content to determine task type (enhanced version)
func (ecd *EnhancedCapabilityDetector) DetectTaskType(content string) TaskType {
	contentLower := strings.ToLower(content)
	
	// Enhanced keyword detection with more sophisticated patterns
	
	// Image generation keywords (high priority)
	imageGenKeywords := []string{
		"generate image", "create image", "draw", "paint", "sketch", "illustrate",
		"dall-e", "midjourney", "stable diffusion", "image generation",
		"picture of", "photo of", "artwork", "visual representation",
	}
	for _, keyword := range imageGenKeywords {
		if strings.Contains(contentLower, keyword) {
			return TaskTypeImage
		}
	}
	
	// Image analysis keywords (multimodal)
	imageAnalysisKeywords := []string{
		"analyze image", "describe image", "what's in this image", 
		"image analysis", "vision", "visual analysis", "image description",
		"identify in image", "recognize in image",
	}
	for _, keyword := range imageAnalysisKeywords {
		if strings.Contains(contentLower, keyword) {
			return TaskTypeMultimodal
		}
	}
	
	// Code-related keywords (enhanced)
	codeKeywords := []string{
		"write code", "program", "function", "algorithm", "debug", "implement",
		"programming", "python", "javascript", "java", "cpp", "rust", "go",
		"code review", "refactor", "optimize code", "fix bug", "api",
		"class", "method", "variable", "loop", "condition", "syntax",
		"repository", "git", "commit", "pull request",
	}
	for _, keyword := range codeKeywords {
		if strings.Contains(contentLower, keyword) {
			return TaskTypeCode
		}
	}
	
	// Audio-related keywords
	audioKeywords := []string{
		"transcribe", "speech to text", "audio", "voice", "sound",
		"whisper", "tts", "text to speech", "speech recognition",
		"audio analysis", "music", "podcast", "recording",
	}
	for _, keyword := range audioKeywords {
		if strings.Contains(contentLower, keyword) {
			return TaskTypeAudio
		}
	}
	
	// Video-related keywords
	videoKeywords := []string{
		"video", "movie", "clip", "animation", "film", "footage",
		"video analysis", "video generation", "motion", "frames",
		"video editing", "cinematography",
	}
	for _, keyword := range videoKeywords {
		if strings.Contains(contentLower, keyword) {
			return TaskTypeVideo
		}
	}
	
	// Multimodal keywords (explicit)
	multimodalKeywords := []string{
		"multimodal", "cross-modal", "vision and language",
		"image and text", "visual question answering", "vqa",
	}
	for _, keyword := range multimodalKeywords {
		if strings.Contains(contentLower, keyword) {
			return TaskTypeMultimodal
		}
	}
	
	// Default to text for everything else
	return TaskTypeText
}

// GetModelDatabase returns the underlying model database for advanced operations
func (ecd *EnhancedCapabilityDetector) GetModelDatabase() *ModelDatabase {
	return ecd.modelDB
}

// GetDetailedCapabilities returns detailed capabilities for all models in a provider
func (ecd *EnhancedCapabilityDetector) GetDetailedCapabilities(models []string, providerName string) map[string]ModelCapabilities {
	result := make(map[string]ModelCapabilities)
	
	for _, model := range models {
		result[model] = ecd.modelDB.GetModelCapabilities(model, providerName)
	}
	
	return result
}

// ValidateCapabilities checks if detected capabilities make sense
func (ecd *EnhancedCapabilityDetector) ValidateCapabilities(capabilities ProviderCapabilities) (bool, []string) {
	var warnings []string
	valid := true
	
	// Check for conflicting capabilities
	if capabilities.Image && capabilities.Audio && capabilities.Video {
		warnings = append(warnings, "Provider claims to support image, audio, and video - unusual combination")
	}
	
	// Check for missing text capability in code models
	if capabilities.Code && !capabilities.Text {
		warnings = append(warnings, "Code capability without text capability is unusual")
		valid = false
	}
	
	// Check for multimodal without base capabilities
	if capabilities.Multimodal && !capabilities.Text && !capabilities.Image {
		warnings = append(warnings, "Multimodal capability without text or image capabilities")
		valid = false
	}
	
	// Check reasoning levels
	if capabilities.Reasoning > 10 || capabilities.Reasoning < 1 {
		warnings = append(warnings, "Reasoning score out of valid range (1-10)")
		valid = false
	}
	
	return valid, warnings
}

// RefreshModelCache clears the model database cache to force fresh lookups
func (ecd *EnhancedCapabilityDetector) RefreshModelCache() {
	ecd.modelDB.ClearCache()
}

// GetCacheStats returns statistics about the model database cache
func (ecd *EnhancedCapabilityDetector) GetCacheStats() map[string]interface{} {
	return ecd.modelDB.GetCacheStats()
}