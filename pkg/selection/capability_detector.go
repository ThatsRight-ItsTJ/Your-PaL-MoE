package selection

import (
	"strings"
	"regexp"
)

// ProviderCapabilities represents what a provider can handle
type ProviderCapabilities struct {
	Text        bool     `json:"text"`
	Image       bool     `json:"image"`
	Code        bool     `json:"code"`
	Audio       bool     `json:"audio"`
	Video       bool     `json:"video"`
	Multimodal  bool     `json:"multimodal"`
	Models      []string `json:"models"`
	Reasoning   int      `json:"reasoning"`    // 1-10 scale
	Knowledge   int      `json:"knowledge"`   // 1-10 scale
	Computation int      `json:"computation"` // 1-10 scale
}

// CapabilityDetector analyzes provider models to determine capabilities
type CapabilityDetector struct {
	// Model patterns for different capabilities
	textPatterns     []*regexp.Regexp
	imagePatterns    []*regexp.Regexp
	codePatterns     []*regexp.Regexp
	audioPatterns    []*regexp.Regexp
	videoPatterns    []*regexp.Regexp
	multimodalPatterns []*regexp.Regexp
	
	// Quality indicators
	highQualityPatterns []*regexp.Regexp
	mediumQualityPatterns []*regexp.Regexp
}

// NewCapabilityDetector creates a new capability detector
func NewCapabilityDetector() *CapabilityDetector {
	return &CapabilityDetector{
		textPatterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)\b(gpt|claude|llama|mistral|gemini|palm|text|chat|instruct)\b`),
			regexp.MustCompile(`(?i)\b(turbo|davinci|curie|babbage)\b`),
		},
		imagePatterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)\b(dall-e|dalle|midjourney|stable-diffusion|imagen|firefly)\b`),
			regexp.MustCompile(`(?i)\b(image|vision|visual|draw|paint|generate)\b`),
		},
		codePatterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)\b(code|codex|copilot|starcoder|codellama|programming)\b`),
			regexp.MustCompile(`(?i)\b(python|javascript|java|cpp|rust|go)\b`),
		},
		audioPatterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)\b(whisper|speech|audio|voice|tts|stt)\b`),
		},
		videoPatterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)\b(video|movie|clip|animation)\b`),
		},
		multimodalPatterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)\b(multimodal|vision|gpt-4v|claude-3|gemini-pro-vision)\b`),
		},
		highQualityPatterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)\b(gpt-4|claude-3|gemini-pro|opus|sonnet)\b`),
			regexp.MustCompile(`(?i)\b(70b|65b|175b)\b`), // Large parameter models
		},
		mediumQualityPatterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)\b(gpt-3\.5|claude-2|llama-2|mistral-7b|13b|7b)\b`),
		},
	}
}

// DetectCapabilities analyzes models to determine provider capabilities
func (cd *CapabilityDetector) DetectCapabilities(models []string) ProviderCapabilities {
	capabilities := ProviderCapabilities{
		Models: models,
	}
	
	allModelsText := strings.ToLower(strings.Join(models, " "))
	
	// Detect basic capabilities
	capabilities.Text = cd.matchesAnyPattern(allModelsText, cd.textPatterns)
	capabilities.Image = cd.matchesAnyPattern(allModelsText, cd.imagePatterns)
	capabilities.Code = cd.matchesAnyPattern(allModelsText, cd.codePatterns)
	capabilities.Audio = cd.matchesAnyPattern(allModelsText, cd.audioPatterns)
	capabilities.Video = cd.matchesAnyPattern(allModelsText, cd.videoPatterns)
	capabilities.Multimodal = cd.matchesAnyPattern(allModelsText, cd.multimodalPatterns)
	
	// If no specific capability detected but has text patterns, assume text capability
	if !capabilities.Text && !capabilities.Image && !capabilities.Code && !capabilities.Audio && !capabilities.Video {
		capabilities.Text = true // Default to text if unclear
	}
	
	// Detect quality levels
	capabilities.Reasoning = cd.detectReasoningLevel(allModelsText)
	capabilities.Knowledge = cd.detectKnowledgeLevel(allModelsText)
	capabilities.Computation = cd.detectComputationLevel(allModelsText)
	
	return capabilities
}

// matchesAnyPattern checks if text matches any of the given patterns
func (cd *CapabilityDetector) matchesAnyPattern(text string, patterns []*regexp.Regexp) bool {
	for _, pattern := range patterns {
		if pattern.MatchString(text) {
			return true
		}
	}
	return false
}

// detectReasoningLevel determines reasoning capability level (1-10)
func (cd *CapabilityDetector) detectReasoningLevel(modelsText string) int {
	if cd.matchesAnyPattern(modelsText, cd.highQualityPatterns) {
		return 9 // High reasoning for advanced models
	}
	if cd.matchesAnyPattern(modelsText, cd.mediumQualityPatterns) {
		return 7 // Medium reasoning
	}
	if strings.Contains(modelsText, "instruct") || strings.Contains(modelsText, "chat") {
		return 6 // Instruction-tuned models have decent reasoning
	}
	return 5 // Default reasoning level
}

// detectKnowledgeLevel determines knowledge capability level (1-10)
func (cd *CapabilityDetector) detectKnowledgeLevel(modelsText string) int {
	if cd.matchesAnyPattern(modelsText, cd.highQualityPatterns) {
		return 9 // High knowledge for advanced models
	}
	if cd.matchesAnyPattern(modelsText, cd.mediumQualityPatterns) {
		return 7 // Medium knowledge
	}
	if strings.Contains(modelsText, "70b") || strings.Contains(modelsText, "65b") {
		return 8 // Large models have more knowledge
	}
	return 6 // Default knowledge level
}

// detectComputationLevel determines computation capability level (1-10)
func (cd *CapabilityDetector) detectComputationLevel(modelsText string) int {
	if cd.matchesAnyPattern(modelsText, cd.codePatterns) {
		return 8 // Code models are good at computation
	}
	if cd.matchesAnyPattern(modelsText, cd.highQualityPatterns) {
		return 7 // Advanced models handle computation well
	}
	if strings.Contains(modelsText, "math") || strings.Contains(modelsText, "calculator") {
		return 9 // Math-specific models
	}
	return 5 // Default computation level
}

// TaskType represents different types of tasks
type TaskType string

const (
	TaskTypeText       TaskType = "text"
	TaskTypeImage      TaskType = "image"
	TaskTypeCode       TaskType = "code"
	TaskTypeAudio      TaskType = "audio"
	TaskTypeVideo      TaskType = "video"
	TaskTypeMultimodal TaskType = "multimodal"
)

// DetectTaskType analyzes request content to determine task type
func (cd *CapabilityDetector) DetectTaskType(content string) TaskType {
	contentLower := strings.ToLower(content)
	
	// Check for image-related keywords
	imageKeywords := []string{"image", "picture", "photo", "draw", "paint", "generate image", "create image", "visual", "dall-e", "midjourney"}
	for _, keyword := range imageKeywords {
		if strings.Contains(contentLower, keyword) {
			return TaskTypeImage
		}
	}
	
	// Check for code-related keywords
	codeKeywords := []string{"code", "program", "function", "algorithm", "debug", "implement", "programming", "python", "javascript", "java", "cpp", "rust", "go"}
	for _, keyword := range codeKeywords {
		if strings.Contains(contentLower, keyword) {
			return TaskTypeCode
		}
	}
	
	// Check for audio-related keywords
	audioKeywords := []string{"audio", "speech", "voice", "sound", "transcribe", "whisper", "tts", "text to speech"}
	for _, keyword := range audioKeywords {
		if strings.Contains(contentLower, keyword) {
			return TaskTypeAudio
		}
	}
	
	// Check for video-related keywords
	videoKeywords := []string{"video", "movie", "clip", "animation", "film"}
	for _, keyword := range videoKeywords {
		if strings.Contains(contentLower, keyword) {
			return TaskTypeVideo
		}
	}
	
	// Check for multimodal keywords
	multimodalKeywords := []string{"analyze image", "describe image", "image and text", "vision", "multimodal"}
	for _, keyword := range multimodalKeywords {
		if strings.Contains(contentLower, keyword) {
			return TaskTypeMultimodal
		}
	}
	
	// Default to text
	return TaskTypeText
}

// IsProviderCompatible checks if a provider can handle a specific task type
func (cd *CapabilityDetector) IsProviderCompatible(capabilities ProviderCapabilities, taskType TaskType) bool {
	switch taskType {
	case TaskTypeText:
		return capabilities.Text
	case TaskTypeImage:
		return capabilities.Image
	case TaskTypeCode:
		return capabilities.Code || capabilities.Text // Text models can often handle code
	case TaskTypeAudio:
		return capabilities.Audio
	case TaskTypeVideo:
		return capabilities.Video
	case TaskTypeMultimodal:
		return capabilities.Multimodal || (capabilities.Text && capabilities.Image)
	default:
		return capabilities.Text // Default to text compatibility
	}
}