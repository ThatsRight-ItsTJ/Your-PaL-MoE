package components

import (
	"time"
	"github.com/tj/Your-PaL-MoE/internal/enhanced"
)

// Type aliases to bridge the gap between components and enhanced packages
type TaskComplexity = enhanced.TaskComplexity
type ComplexityLevel = enhanced.ComplexityLevel
type RequestInput = enhanced.RequestInput
type OptimizedPrompt = enhanced.OptimizedPrompt
type ExecutionResult = enhanced.ExecutionResult
type Config = enhanced.Config

// Constants from enhanced package
const (
	VeryHigh = enhanced.VeryHigh
	High     = enhanced.High
	Medium   = enhanced.Medium
	Low      = enhanced.Low
)

// Additional types that might be needed
type ProviderMetrics struct {
	RequestCount    int64         `json:"request_count"`
	SuccessRate     float64       `json:"success_rate"`
	AverageLatency  time.Duration `json:"average_latency"`
	ErrorRate       float64       `json:"error_rate"`
	TotalCost       float64       `json:"total_cost"`
	LastUpdated     time.Time     `json:"last_updated"`
}

type SystemMetrics struct {
	TotalRequests     int64         `json:"total_requests"`
	SuccessfulRequests int64        `json:"successful_requests"`
	FailedRequests    int64         `json:"failed_requests"`
	AverageLatency    time.Duration `json:"average_latency"`
	TotalCost         float64       `json:"total_cost"`
	LastUpdated       time.Time     `json:"last_updated"`
}

type RequestResult struct {
	Response    string                 `json:"response"`
	Provider    string                 `json:"provider"`
	Model       string                 `json:"model"`
	Complexity  TaskComplexity         `json:"complexity"`
	Cost        float64                `json:"cost"`
	Duration    time.Duration          `json:"duration"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}