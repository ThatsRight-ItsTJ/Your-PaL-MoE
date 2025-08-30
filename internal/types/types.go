package types

import "time"

// RequestInput represents input for processing requests
type RequestInput struct {
	ID        string                 `json:"id"`
	Content   string                 `json:"content"`
	Context   map[string]interface{} `json:"context"`
	Timestamp time.Time              `json:"timestamp"`
}

// RequestResult represents the result of processing a request
type RequestResult struct {
	ID           string                 `json:"id"`
	Status       string                 `json:"status"`
	Response     string                 `json:"response"`
	Provider     string                 `json:"provider"`
	Model        string                 `json:"model"`
	Cost         float64                `json:"cost"`
	ResponseTime time.Duration          `json:"response_time"`
	Metadata     map[string]interface{} `json:"metadata"`
	Timestamp    time.Time              `json:"timestamp"`
}

// Provider represents an AI provider configuration
type Provider struct {
	Name           string            `json:"name"`
	Tier           string            `json:"tier"`
	Endpoint       string            `json:"endpoint"`
	ModelsSource   string            `json:"models_source"`
	Authentication AuthConfig        `json:"authentication"`
	Metadata       map[string]string `json:"metadata"`
}

// AuthConfig represents authentication configuration
type AuthConfig struct {
	Type   string `json:"type"`
	APIKey string `json:"api_key,omitempty"`
}