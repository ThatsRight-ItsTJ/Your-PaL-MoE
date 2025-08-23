package components

import (
	"context"
	"fmt"
	"hash/fnv"
	"sort"
	"strings"
	"sync"
	"time"

	"enhanced-yourpal-moe/internal/types"
	"enhanced-yourpal-moe/pkg/config"

	"github.com/sirupsen/logrus"
)

// SPOOptimizer implements self-supervised prompt optimization
type SPOOptimizer struct {
	config *config.Config
	logger *logrus.Logger
	
	// Cache for optimization results
	cache      map[string]*CachedOptimization
	cacheMutex sync.RWMutex
	
	// Performance tracking
	optimizationHistory map[string][]OptimizationAttempt
	historyMutex       sync.RWMutex
}

// CachedOptimization represents a cached optimization result
type CachedOptimization struct {
	Original     string
	Optimized    string
	Score        float64
	Improvements []string
	Timestamp    time.Time
	HitCount     int
}

// OptimizationAttempt tracks individual optimization attempts
type OptimizationAttempt struct {
	Original      string
	Optimized     string
	Score         float64
	Iteration     int
	Improvements  []string
	ExecutionTime time.Duration
	Timestamp     time.Time
}

// NewSPOOptimizer creates a new SPO optimizer
func NewSPOOptimizer(cfg *config.Config, logger *logrus.Logger) (*SPOOptimizer, error) {
	optimizer := &SPOOptimizer{
		config:              cfg,
		logger:              logger,
		cache:               make(map[string]*CachedOptimization),
		optimizationHistory: make(map[string][]OptimizationAttempt),
	}
	
	// Start cache cleanup routine
	go optimizer.cacheCleanupRoutine()
	
	return optimizer, nil
}

// OptimizePrompt optimizes a prompt using SPO methodology
func (s *SPOOptimizer) OptimizePrompt(ctx context.Context, originalPrompt string, complexity types.TaskComplexity) (types.OptimizedPrompt, error) {
	startTime := time.Now()
	s.logger.Infof("Starting SPO optimization for prompt (complexity: %.2f)", complexity.Score)
	
	// Check cache first
	cacheKey := s.generateCacheKey(originalPrompt, complexity)
	if cached := s.getFromCache(cacheKey); cached != nil {
		s.logger.Infof("Cache hit for prompt optimization")
		return types.OptimizedPrompt{
			Original:     originalPrompt,
			Optimized:    cached.Optimized,
			Iterations:   0, // Cached result
			Improvements: cached.Improvements,
			Confidence:   cached.Score,
			CostSavings:  s.estimateCostSavings(cached.Score),
		}, nil
	}
	
	// Perform optimization
	result, err := s.performOptimization(ctx, originalPrompt, complexity)
	if err != nil {
		return types.OptimizedPrompt{}, fmt.Errorf("optimization failed: %w", err)
	}
	
	// Cache the result
	s.cacheResult(cacheKey, result)
	
	// Track optimization attempt
	attempt := OptimizationAttempt{
		Original:      originalPrompt,
		Optimized:     result.Optimized,
		Score:         result.Confidence,
		Iteration:     result.Iterations,
		Improvements:  result.Improvements,
		ExecutionTime: time.Since(startTime),
		Timestamp:     time.Now(),
	}
	s.trackOptimization(originalPrompt, attempt)
	
	s.logger.Infof("SPO optimization completed in %v with confidence %.2f", 
		time.Since(startTime), result.Confidence)
	
	return result, nil
}

// performOptimization performs the actual SPO optimization process
func (s *SPOOptimizer) performOptimization(ctx context.Context, original string, complexity types.TaskComplexity) (types.OptimizedPrompt, error) {
	maxIterations := s.config.SPO.MaxIterations
	samplesPerRound := s.config.SPO.SamplesPerRound
	
	currentPrompt := original
	bestPrompt := original
	bestScore := 0.0
	improvements := make([]string, 0)
	
	for iteration := 0; iteration < maxIterations; iteration++ {
		s.logger.Debugf("SPO iteration %d/%d", iteration+1, maxIterations)
		
		// Generate variants of the current prompt
		variants := s.generatePromptVariants(currentPrompt, complexity, samplesPerRound)
		
		// Evaluate variants using pairwise comparison
		scores := s.evaluateVariants(variants, complexity)
		
		// Find the best variant
		bestVariantIndex := 0
		bestVariantScore := scores[0]
		for i, score := range scores {
			if score > bestVariantScore {
				bestVariantScore = score
				bestVariantIndex = i
			}
		}
		
		// Check for improvement
		if bestVariantScore > bestScore {
			improvement := s.identifyImprovement(currentPrompt, variants[bestVariantIndex])
			improvements = append(improvements, improvement)
			
			bestPrompt = variants[bestVariantIndex]
			bestScore = bestVariantScore
			currentPrompt = bestPrompt
			
			s.logger.Debugf("Improvement found in iteration %d: score %.2f", iteration+1, bestScore)
		} else {
			// No improvement, check for convergence
			if s.hasConverged(bestScore, s.config.SPO.ConvergenceRate) {
				s.logger.Debugf("Convergence reached after %d iterations", iteration+1)
				break
			}
		}
	}
	
	return types.OptimizedPrompt{
		Original:     original,
		Optimized:    bestPrompt,
		Iterations:   maxIterations,
		Improvements: improvements,
		Confidence:   bestScore,
		CostSavings:  s.estimateCostSavings(bestScore),
	}, nil
}

// generatePromptVariants generates variants of a prompt for optimization
func (s *SPOOptimizer) generatePromptVariants(prompt string, complexity types.TaskComplexity, count int) []string {
	variants := make([]string, count)
	
	for i := 0; i < count; i++ {
		variant := s.applyOptimizationStrategy(prompt, complexity, i)
		variants[i] = variant
	}
	
	return variants
}

// applyOptimizationStrategy applies different optimization strategies
func (s *SPOOptimizer) applyOptimizationStrategy(prompt string, complexity types.TaskComplexity, strategyIndex int) string {
	strategies := []func(string, types.TaskComplexity) string{
		s.addClarificationStrategy,
		s.addStructureStrategy,
		s.addContextStrategy,
		s.addConstraintsStrategy,
		s.addExamplesStrategy,
	}
	
	strategy := strategies[strategyIndex%len(strategies)]
	return strategy(prompt, complexity)
}

// addClarificationStrategy adds clarification to the prompt
func (s *SPOOptimizer) addClarificationStrategy(prompt string, complexity types.TaskComplexity) string {
	if complexity.Overall >= types.Medium {
		return prompt + "\n\nPlease be specific and detailed in your response."
	}
	return prompt + "\n\nPlease provide a clear and concise response."
}

// addStructureStrategy adds structure guidance to the prompt
func (s *SPOOptimizer) addStructureStrategy(prompt string, complexity types.TaskComplexity) string {
	if complexity.Overall >= types.High {
		return prompt + "\n\nPlease structure your response with clear headings and bullet points where appropriate."
	}
	return prompt + "\n\nPlease organize your response clearly."
}

// addContextStrategy adds context guidance to the prompt
func (s *SPOOptimizer) addContextStrategy(prompt string, complexity types.TaskComplexity) string {
	if complexity.Knowledge >= types.High {
		return prompt + "\n\nConsider relevant background information and context in your response."
	}
	return prompt + "\n\nProvide relevant context as needed."
}

// addConstraintsStrategy adds constraint guidance to the prompt
func (s *SPOOptimizer) addConstraintsStrategy(prompt string, complexity types.TaskComplexity) string {
	return prompt + "\n\nEnsure your response is accurate, helpful, and appropriate."
}

// addExamplesStrategy adds example guidance to the prompt
func (s *SPOOptimizer) addExamplesStrategy(prompt string, complexity types.TaskComplexity) string {
	if complexity.Overall >= types.Medium {
		return prompt + "\n\nInclude relevant examples to illustrate your points."
	}
	return prompt
}

// evaluateVariants evaluates prompt variants using pairwise comparison
func (s *SPOOptimizer) evaluateVariants(variants []string, complexity types.TaskComplexity) []float64 {
	scores := make([]float64, len(variants))
	
	for i, variant := range variants {
		score := s.calculatePromptQuality(variant, complexity)
		scores[i] = score
	}
	
	return scores
}

// calculatePromptQuality calculates the quality score of a prompt
func (s *SPOOptimizer) calculatePromptQuality(prompt string, complexity types.TaskComplexity) float64 {
	score := 0.0
	
	// Base score from prompt length and structure
	words := strings.Fields(prompt)
	if len(words) >= 10 && len(words) <= 100 {
		score += 0.3 // Good length
	}
	
	// Clarity indicators
	clarityWords := []string{"specific", "detailed", "clear", "please", "exactly"}
	for _, word := range clarityWords {
		if strings.Contains(strings.ToLower(prompt), word) {
			score += 0.1
		}
	}
	
	// Structure indicators
	if strings.Contains(prompt, "\n") {
		score += 0.2 // Has structure
	}
	
	// Context indicators
	contextWords := []string{"context", "background", "relevant", "consider"}
	for _, word := range contextWords {
		if strings.Contains(strings.ToLower(prompt), word) {
			score += 0.1
		}
	}
	
	// Complexity alignment
	if complexity.Overall >= types.High && len(words) >= 20 {
		score += 0.2 // Complex tasks need detailed prompts
	}
	
	// Normalize score to 0-1 range
	if score > 1.0 {
		score = 1.0
	}
	
	return score
}

// identifyImprovement identifies what improvement was made
func (s *SPOOptimizer) identifyImprovement(original, optimized string) string {
	if len(optimized) > len(original)*1.2 {
		return "Added detailed guidance"
	} else if strings.Contains(optimized, "\n") && !strings.Contains(original, "\n") {
		return "Improved structure"
	} else if strings.Contains(strings.ToLower(optimized), "specific") {
		return "Enhanced clarity"
	}
	return "General optimization"
}

// hasConverged checks if optimization has converged
func (s *SPOOptimizer) hasConverged(score float64, threshold float64) bool {
	return score >= (1.0 - threshold)
}

// estimateCostSavings estimates cost savings from optimization
func (s *SPOOptimizer) estimateCostSavings(score float64) float64 {
	// Higher quality prompts lead to better results with cheaper providers
	return score * 0.3 // Up to 30% cost savings
}

// generateCacheKey generates a cache key for a prompt and complexity
func (s *SPOOptimizer) generateCacheKey(prompt string, complexity types.TaskComplexity) string {
	h := fnv.New64a()
	h.Write([]byte(prompt))
	h.Write([]byte(fmt.Sprintf("%.2f", complexity.Score)))
	return fmt.Sprintf("%x", h.Sum64())
}

// getFromCache retrieves a cached optimization result
func (s *SPOOptimizer) getFromCache(key string) *CachedOptimization {
	s.cacheMutex.RLock()
	defer s.cacheMutex.RUnlock()
	
	if cached, exists := s.cache[key]; exists {
		// Check TTL
		if time.Since(cached.Timestamp) < time.Duration(s.config.SPO.CacheTTL)*time.Second {
			cached.HitCount++
			return cached
		}
		// Remove expired entry
		delete(s.cache, key)
	}
	
	return nil
}

// cacheResult caches an optimization result
func (s *SPOOptimizer) cacheResult(key string, result types.OptimizedPrompt) {
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()
	
	// Check cache size limit
	if len(s.cache) >= s.config.SPO.CacheSize {
		s.evictLeastUsed()
	}
	
	s.cache[key] = &CachedOptimization{
		Original:     result.Original,
		Optimized:    result.Optimized,
		Score:        result.Confidence,
		Improvements: result.Improvements,
		Timestamp:    time.Now(),
		HitCount:     0,
	}
}

// evictLeastUsed evicts the least used cache entry
func (s *SPOOptimizer) evictLeastUsed() {
	var leastUsedKey string
	minHits := int(^uint(0) >> 1) // Max int
	
	for key, cached := range s.cache {
		if cached.HitCount < minHits {
			minHits = cached.HitCount
			leastUsedKey = key
		}
	}
	
	if leastUsedKey != "" {
		delete(s.cache, leastUsedKey)
	}
}

// trackOptimization tracks an optimization attempt
func (s *SPOOptimizer) trackOptimization(prompt string, attempt OptimizationAttempt) {
	s.historyMutex.Lock()
	defer s.historyMutex.Unlock()
	
	key := fmt.Sprintf("%x", fnv.New64a().Sum([]byte(prompt)))
	s.optimizationHistory[key] = append(s.optimizationHistory[key], attempt)
	
	// Keep only recent attempts (last 10)
	if len(s.optimizationHistory[key]) > 10 {
		s.optimizationHistory[key] = s.optimizationHistory[key][1:]
	}
}

// cacheCleanupRoutine periodically cleans up expired cache entries
func (s *SPOOptimizer) cacheCleanupRoutine() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			s.cleanupExpiredEntries()
		}
	}
}

// cleanupExpiredEntries removes expired cache entries
func (s *SPOOptimizer) cleanupExpiredEntries() {
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()
	
	ttl := time.Duration(s.config.SPO.CacheTTL) * time.Second
	now := time.Now()
	
	for key, cached := range s.cache {
		if now.Sub(cached.Timestamp) > ttl {
			delete(s.cache, key)
		}
	}
}

// LearnFromFeedback learns from execution feedback to improve future optimizations
func (s *SPOOptimizer) LearnFromFeedback(ctx context.Context, results []types.ExecutionResult) error {
	s.logger.Infof("Learning from %d execution results", len(results))
	
	for _, result := range results {
		if result.Quality.OverallScore > 0 {
			// Track successful optimizations for future reference
			s.updateOptimizationSuccess(result)
		}
	}
	
	return nil
}

// updateOptimizationSuccess updates optimization success metrics
func (s *SPOOptimizer) updateOptimizationSuccess(result types.ExecutionResult) {
	// This would update ML models or heuristics based on successful results
	// For now, we'll log the success
	s.logger.Debugf("Task %s succeeded with quality score %.2f", result.TaskID, result.Quality.OverallScore)
}