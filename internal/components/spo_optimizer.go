package components

import (
	"fmt"
	"strings"
	"time"
)

// ComplexityLevel represents the complexity level of a task
type ComplexityLevel int

const (
	Low ComplexityLevel = iota
	Medium
	High
	VeryHigh
)

// String returns string representation of complexity level
func (cl ComplexityLevel) String() string {
	switch cl {
	case Low:
		return "low"
	case Medium:
		return "medium"
	case High:
		return "high"
	case VeryHigh:
		return "very_high"
	default:
		return "unknown"
	}
}

// TaskComplexity represents the complexity analysis of a task
type TaskComplexity struct {
	Overall              ComplexityLevel     `json:"overall"`
	Reasoning            ComplexityLevel     `json:"reasoning"`
	Mathematical         ComplexityLevel     `json:"mathematical"`
	Creative             ComplexityLevel     `json:"creative"`
	Factual              ComplexityLevel     `json:"factual"`
	TokenEstimate        int64               `json:"token_estimate"`
	RequiredCapabilities []string            `json:"required_capabilities"`
	Metadata             map[string]interface{} `json:"metadata"`
}

// SPOOptimizer represents a self-prompt optimizer
type SPOOptimizer struct {
	templates         map[string]string
	optimizationRules map[ComplexityLevel][]string
	cache            map[string]CachedOptimization
	maxCacheSize     int
	cacheHitRate     float64
	totalOptimizations int64
}

// CachedOptimization represents a cached optimization result
type CachedOptimization struct {
	OptimizedPrompt string
	Complexity      ComplexityLevel
	Timestamp       time.Time
	HitCount        int
}

// NewSPOOptimizer creates a new SPO optimizer
func NewSPOOptimizer() *SPOOptimizer {
	return &SPOOptimizer{
		templates: map[string]string{
			"reasoning": "Please analyze the following problem step by step: {prompt}",
			"creative":  "Generate creative and original content for: {prompt}",
			"factual":   "Provide accurate and factual information about: {prompt}",
			"mathematical": "Solve the following mathematical problem: {prompt}",
		},
		optimizationRules: map[ComplexityLevel][]string{
			Low:      {"simplify", "clarify"},
			Medium:   {"structure", "examples", "context"},
			High:     {"breakdown", "methodology", "constraints"},
			VeryHigh: {"systematic", "comprehensive", "multi-step"},
		},
		cache:        make(map[string]CachedOptimization),
		maxCacheSize: 1000,
	}
}

// OptimizePrompt optimizes a prompt based on complexity
func (spo *SPOOptimizer) OptimizePrompt(prompt string, complexity TaskComplexity) (string, error) {
	if strings.TrimSpace(prompt) == "" {
		return "", fmt.Errorf("prompt cannot be empty")
	}

	// Check cache first
	cacheKey := fmt.Sprintf("%s_%s", prompt, complexity.Overall.String())
	if cached, exists := spo.cache[cacheKey]; exists {
		cached.HitCount++
		spo.cache[cacheKey] = cached
		return cached.OptimizedPrompt, nil
	}

	// Apply optimization rules based on complexity
	optimizedPrompt := spo.applyOptimizationRules(prompt, complexity)

	// Cache the result
	if len(spo.cache) < spo.maxCacheSize {
		spo.cache[cacheKey] = CachedOptimization{
			OptimizedPrompt: optimizedPrompt,
			Complexity:      complexity.Overall,
			Timestamp:       time.Now(),
			HitCount:        1,
		}
	}

	spo.totalOptimizations++
	return optimizedPrompt, nil
}

// applyOptimizationRules applies optimization rules based on complexity
func (spo *SPOOptimizer) applyOptimizationRules(prompt string, complexity TaskComplexity) string {
	optimized := prompt

	// Get rules for the complexity level
	if rules, exists := spo.optimizationRules[complexity.Overall]; exists {
		for _, rule := range rules {
			optimized = spo.applyRule(optimized, rule, complexity)
		}
	}

	return optimized
}

// applyRule applies a specific optimization rule
func (spo *SPOOptimizer) applyRule(prompt string, rule string, complexity TaskComplexity) string {
	switch rule {
	case "simplify":
		return fmt.Sprintf("Please provide a simple and clear answer to: %s", prompt)
	case "clarify":
		return fmt.Sprintf("Please clarify and explain: %s", prompt)
	case "structure":
		return fmt.Sprintf("Please provide a structured response to: %s\n\nPlease organize your answer with clear sections.", prompt)
	case "examples":
		return fmt.Sprintf("%s\n\nPlease include relevant examples in your response.", prompt)
	case "context":
		return fmt.Sprintf("Considering the context and background, please address: %s", prompt)
	case "breakdown":
		return fmt.Sprintf("Please break down the following into manageable parts: %s\n\nProvide a step-by-step analysis.", prompt)
	case "methodology":
		return fmt.Sprintf("%s\n\nPlease explain your methodology and reasoning process.", prompt)
	case "constraints":
		return fmt.Sprintf("Considering all constraints and limitations, please address: %s", prompt)
	case "systematic":
		return fmt.Sprintf("Please provide a systematic and comprehensive analysis of: %s\n\nUse a structured approach with clear reasoning.", prompt)
	case "comprehensive":
		return fmt.Sprintf("%s\n\nPlease provide a comprehensive response covering all relevant aspects.", prompt)
	case "multi-step":
		return fmt.Sprintf("Please solve the following using a multi-step approach: %s\n\n1. First, analyze the problem\n2. Then, develop a solution strategy\n3. Finally, implement and verify the solution", prompt)
	default:
		return prompt
	}
}

// GetOptimizationStats returns optimization statistics
func (spo *SPOOptimizer) GetOptimizationStats() map[string]interface{} {
	cacheHits := 0
	for _, cached := range spo.cache {
		if cached.HitCount > 1 {
			cacheHits += cached.HitCount - 1
		}
	}

	hitRate := 0.0
	if spo.totalOptimizations > 0 {
		hitRate = float64(cacheHits) / float64(spo.totalOptimizations)
	}

	return map[string]interface{}{
		"total_optimizations": spo.totalOptimizations,
		"cache_size":         len(spo.cache),
		"cache_hit_rate":     hitRate,
		"max_cache_size":     spo.maxCacheSize,
	}
}

// ClearCache clears the optimization cache
func (spo *SPOOptimizer) ClearCache() {
	spo.cache = make(map[string]CachedOptimization)
}

// SetMaxCacheSize sets the maximum cache size
func (spo *SPOOptimizer) SetMaxCacheSize(size int) {
	spo.maxCacheSize = size
	
	// Trim cache if necessary
	if len(spo.cache) > size {
		// Simple implementation: clear all cache when over limit
		spo.ClearCache()
	}
}