package components

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/ThatsRight-ItsTJ/Your-PaL-MoE/internal/enhanced"
)

// SPOOptimizer implements Self-Prompt Optimization
type SPOOptimizer struct {
	templates          map[string]string
	optimizationRules  map[enhanced.ComplexityLevel][]string
	cache             map[string]CachedOptimization
	maxCacheSize      int
	cacheHitRate      float64
	totalOptimizations int64
}

// CachedOptimization represents a cached optimization result
type CachedOptimization struct {
	OptimizedPrompt string
	Complexity      enhanced.ComplexityLevel
	Timestamp       time.Time
	HitCount        int
}

// OptimizedPrompt represents an optimized prompt with metadata
type OptimizedPrompt struct {
	OriginalPrompt   string                 `json:"original_prompt"`
	OptimizedPrompt  string                 `json:"optimized_prompt"`
	Complexity       enhanced.ComplexityLevel `json:"complexity"`
	OptimizationRules []string              `json:"optimization_rules"`
	Metadata         map[string]interface{} `json:"metadata"`
	ProcessingTime   time.Duration          `json:"processing_time"`
}

// Config represents configuration for the SPO optimizer
type Config struct {
	MaxCacheSize      int                                                `json:"max_cache_size"`
	OptimizationRules map[enhanced.ComplexityLevel][]string             `json:"optimization_rules"`
	Templates         map[string]string                                  `json:"templates"`
}

// NewSPOOptimizer creates a new SPO optimizer with default configuration
func NewSPOOptimizer(cfg *Config) *SPOOptimizer {
	if cfg == nil {
		cfg = &Config{
			MaxCacheSize: 1000,
			OptimizationRules: map[enhanced.ComplexityLevel][]string{
				enhanced.Low:      {"Keep responses concise", "Use simple language"},
				enhanced.Medium:   {"Provide context", "Use examples"},
				enhanced.High:     {"Break down problems", "Provide detailed reasoning"},
				enhanced.VeryHigh: {"Use systematic analysis", "Employ chain-of-thought"},
			},
			Templates: make(map[string]string),
		}
	}

	return &SPOOptimizer{
		templates:         cfg.Templates,
		optimizationRules: cfg.OptimizationRules,
		cache:            make(map[string]CachedOptimization),
		maxCacheSize:     cfg.MaxCacheSize,
		cacheHitRate:     0.0,
	}
}

// OptimizePrompt optimizes a prompt based on task complexity
func (spo *SPOOptimizer) OptimizePrompt(ctx context.Context, prompt string, complexity enhanced.TaskComplexity) (*OptimizedPrompt, error) {
	startTime := time.Now()
	spo.totalOptimizations++

	// Check cache first
	cacheKey := fmt.Sprintf("%s_%s", prompt, complexity.Overall.String())
	if cached, exists := spo.cache[cacheKey]; exists {
		cached.HitCount++
		spo.cache[cacheKey] = cached
		spo.updateCacheHitRate(true)
		
		return &OptimizedPrompt{
			OriginalPrompt:   prompt,
			OptimizedPrompt:  cached.OptimizedPrompt,
			Complexity:       complexity.Overall,
			OptimizationRules: spo.optimizationRules[complexity.Overall],
			Metadata:         map[string]interface{}{"cache_hit": true},
			ProcessingTime:   time.Since(startTime),
		}, nil
	}

	spo.updateCacheHitRate(false)

	// Apply optimization rules based on complexity
	optimized := spo.applyOptimizationRules(prompt, complexity.Overall)
	
	// Apply specific optimizations based on task type
	optimized = spo.applyTaskSpecificOptimizations(optimized, complexity)
	
	// Cache the result
	spo.cacheOptimization(cacheKey, optimized, complexity.Overall)

	result := &OptimizedPrompt{
		OriginalPrompt:   prompt,
		OptimizedPrompt:  optimized,
		Complexity:       complexity.Overall,
		OptimizationRules: spo.optimizationRules[complexity.Overall],
		Metadata:         map[string]interface{}{"cache_hit": false},
		ProcessingTime:   time.Since(startTime),
	}

	return result, nil
}

// OptimizePromptBatch optimizes multiple prompts in batch
func (spo *SPOOptimizer) OptimizePromptBatch(ctx context.Context, prompts []string, complexity enhanced.TaskComplexity) ([]*OptimizedPrompt, error) {
	results := make([]*OptimizedPrompt, len(prompts))
	
	for i, prompt := range prompts {
		optimized, err := spo.OptimizePrompt(ctx, prompt, complexity)
		if err != nil {
			return nil, fmt.Errorf("failed to optimize prompt %d: %w", i, err)
		}
		results[i] = optimized
	}
	
	return results, nil
}

// GetOptimizationTemplate returns an optimization template for a given complexity
func (spo *SPOOptimizer) GetOptimizationTemplate(complexity enhanced.TaskComplexity) string {
	switch complexity.Overall {
	case enhanced.Low:
		return "Please provide a clear and concise response to: %s"
	case enhanced.Medium:
		return "Please analyze and provide a detailed response with examples for: %s"
	case enhanced.High:
		return "Please break down this complex problem step by step and provide detailed reasoning: %s"
	case enhanced.VeryHigh:
		return "Please use systematic analysis and chain-of-thought reasoning to thoroughly address: %s"
	default:
		return "Please respond to: %s"
	}
}

// applyOptimizationRules applies complexity-specific optimization rules
func (spo *SPOOptimizer) applyOptimizationRules(prompt string, complexity enhanced.ComplexityLevel) string {
	rules, exists := spo.optimizationRules[complexity]
	if !exists {
		return prompt
	}

	optimized := prompt
	
	// Apply each rule
	for _, rule := range rules {
		switch rule {
		case "Keep responses concise":
			optimized = "Be concise. " + optimized
		case "Use simple language":
			optimized = "Use simple, clear language. " + optimized
		case "Provide context":
			optimized = "Provide relevant context and background. " + optimized
		case "Use examples":
			optimized = optimized + " Include relevant examples."
		case "Break down problems":
			optimized = "Break this down step by step. " + optimized
		case "Provide detailed reasoning":
			optimized = optimized + " Show your reasoning process."
		case "Use systematic analysis":
			optimized = "Use systematic analysis. " + optimized
		case "Employ chain-of-thought":
			optimized = "Think through this step by step using chain-of-thought reasoning. " + optimized
		}
	}
	
	return optimized
}

// applyTaskSpecificOptimizations applies optimizations based on specific task aspects
func (spo *SPOOptimizer) applyTaskSpecificOptimizations(prompt string, complexity enhanced.TaskComplexity) string {
	optimized := prompt
	
	// Mathematical tasks
	if complexity.Mathematical >= enhanced.Medium {
		optimized = "For mathematical problems, show all work and calculations. " + optimized
	}
	
	// Creative tasks
	if complexity.Creative >= enhanced.Medium {
		optimized = optimized + " Be creative and think outside the box."
	}
	
	// Reasoning tasks
	if complexity.Reasoning >= enhanced.High {
		optimized = "Use logical reasoning and provide clear justification for your conclusions. " + optimized
	}
	
	return optimized
}

// cacheOptimization stores an optimization result in cache
func (spo *SPOOptimizer) cacheOptimization(key, optimized string, complexity enhanced.ComplexityLevel) {
	// Remove oldest entries if cache is full
	if len(spo.cache) >= spo.maxCacheSize {
		spo.evictOldestCacheEntry()
	}
	
	spo.cache[key] = CachedOptimization{
		OptimizedPrompt: optimized,
		Complexity:      complexity,
		Timestamp:       time.Now(),
		HitCount:        0,
	}
}

// evictOldestCacheEntry removes the oldest cache entry
func (spo *SPOOptimizer) evictOldestCacheEntry() {
	var oldestKey string
	var oldestTime time.Time = time.Now()
	
	for key, cached := range spo.cache {
		if cached.Timestamp.Before(oldestTime) {
			oldestTime = cached.Timestamp
			oldestKey = key
		}
	}
	
	if oldestKey != "" {
		delete(spo.cache, oldestKey)
	}
}

// updateCacheHitRate updates the cache hit rate statistics
func (spo *SPOOptimizer) updateCacheHitRate(hit bool) {
	if spo.totalOptimizations == 0 {
		return
	}
	
	hits := int64(spo.cacheHitRate * float64(spo.totalOptimizations-1))
	if hit {
		hits++
	}
	
	spo.cacheHitRate = float64(hits) / float64(spo.totalOptimizations)
}

// GetStats returns optimization statistics
func (spo *SPOOptimizer) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"total_optimizations": spo.totalOptimizations,
		"cache_hit_rate":      spo.cacheHitRate,
		"cache_size":          len(spo.cache),
		"max_cache_size":      spo.maxCacheSize,
	}
}

// ClearCache clears the optimization cache
func (spo *SPOOptimizer) ClearCache() {
	spo.cache = make(map[string]CachedOptimization)
	spo.cacheHitRate = 0.0
	log.Println("SPO optimizer cache cleared")
}

// ValidatePrompt validates that a prompt meets basic requirements
func (spo *SPOOptimizer) ValidatePrompt(prompt string) error {
	if strings.TrimSpace(prompt) == "" {
		return fmt.Errorf("prompt cannot be empty")
	}
	
	if len(prompt) > 10000 {
		return fmt.Errorf("prompt too long (max 10000 characters)")
	}
	
	return nil
}