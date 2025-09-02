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