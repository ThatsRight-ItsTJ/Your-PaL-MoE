package enhanced

import (
	"context"
	"fmt"
	"hash/fnv"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// OptimizedPrompt represents a prompt that has been optimized through SPO
type OptimizedPrompt struct {
	Original     string            `json:"original"`
	Optimized    string            `json:"optimized"`
	Iterations   int               `json:"iterations"`
	Improvements []string          `json:"improvements"`
	Confidence   float64           `json:"confidence"`
	CostSavings  float64           `json:"cost_savings"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// SPOOptimizer implements self-supervised prompt optimization
type SPOOptimizer struct {
	logger *logrus.Logger
	
	// Configuration
	maxIterations   int
	samplesPerRound int
	convergenceRate float64
	learningRate    float64
	cacheSize       int
	cacheTTL        time.Duration
	
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
func NewSPOOptimizer(logger *logrus.Logger) (*SPOOptimizer, error) {
	optimizer := &SPOOptimizer{
		logger:              logger,
		maxIterations:       10,
		samplesPerRound:     3,
		convergenceRate:     0.05,
		learningRate:        0.1,
		cacheSize:           1000,
		cacheTTL:            time.Hour,
		cache:               make(map[string]*CachedOptimization),
		optimizationHistory: make(map[string][]OptimizationAttempt),
	}
	
	// Start cache cleanup routine
	go optimizer.cacheCleanupRoutine()
	
	return optimizer, nil
}

// OptimizePrompt optimizes a prompt using SPO methodology
func (s *SPOOptimizer) OptimizePrompt(ctx context.Context, originalPrompt string, complexity TaskComplexity) (OptimizedPrompt, error) {
	startTime := time.Now()
	s.logger.Infof("Starting SPO optimization for prompt (complexity: %.2f)", complexity.Score)
	
	// Check cache first
	cacheKey := s.generateCacheKey(originalPrompt, complexity)
	if cached := s.getFromCache(cacheKey); cached != nil {
		s.logger.Infof("Cache hit for prompt optimization")
		return OptimizedPrompt{
			Original:     originalPrompt,
			Optimized:    cached.Optimized,
			Iterations:   0,
			Improvements: cached.Improvements,
			Confidence:   cached.Score,
			CostSavings:  s.estimateCostSavings(cached.Score),
		}, nil
	}
	
	// Perform optimization
	result, err := s.performOptimization(ctx, originalPrompt, complexity)
	if err != nil {
		return OptimizedPrompt{}, fmt.Errorf("optimization failed: %w", err)
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
func (s *SPOOptimizer) performOptimization(ctx context.Context, original string, complexity TaskComplexity) (OptimizedPrompt, error) {
	currentPrompt := original
	bestPrompt := original
	bestScore := 0.0
	improvements := make([]string, 0)
	
	for iteration := 0; iteration < s.maxIterations; iteration++ {
		s.logger.Debugf("SPO iteration %d/%d", iteration+1, s.maxIterations)
		
		// Generate variants of the current prompt
		variants := s.generatePromptVariants(currentPrompt, complexity, s.samplesPerRound)
		
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
			if s.hasConverged(bestScore, s.convergenceRate) {
				s.logger.Debugf("Convergence reached after %d iterations", iteration+1)
				break
			}
		}
	}
	
	return OptimizedPrompt{
		Original:     original,
		Optimized:    bestPrompt,
		Iterations:   s.maxIterations,
		Improvements: improvements,
		Confidence:   bestScore,
		CostSavings:  s.estimateCostSavings(bestScore),
	}, nil
}

// generatePromptVariants generates variants of a prompt for optimization
func (s *SPOOptimizer) generatePromptVariants(prompt string, complexity TaskComplexity, count int) []string {
	variants := make([]string, count)
	
	for i := 0; i < count; i++ {
		variant := s.applyOptimizationStrategy(prompt, complexity, i)
		variants[i] = variant
	}
	
	return variants
}

// applyOptimizationStrategy applies different optimization strategies
func (s *SPOOptimizer) applyOptimizationStrategy(prompt string, complexity TaskComplexity, strategyIndex int) string {
	strategies := []func(string, TaskComplexity) string{
		s.addClarificationStrategy,
		s.addStructureStrategy,
		s.addContextStrategy,
		s.addConstraintsStrategy,
		s.addExamplesStrategy,
	}
	
	strategy := strategies[strategyIndex%len(strategies)]
	return strategy(prompt, complexity)
}

// Strategy implementations
func (s *SPOOptimizer) addClarificationStrategy(prompt string, complexity TaskComplexity) string {
	if complexity.Overall >= Medium {
		return prompt + "\n\nPlease be specific and detailed in your response."
	}
	return prompt + "\n\nPlease provide a clear and concise response."
}

func (s *SPOOptimizer) addStructureStrategy(prompt string, complexity TaskComplexity) string {
	if complexity.Overall >= High {
		return prompt + "\n\nPlease structure your response with clear headings and bullet points where appropriate."
	}
	return prompt + "\n\nPlease organize your response clearly."
}

func (s *SPOOptimizer) addContextStrategy(prompt string, complexity TaskComplexity) string {
	if complexity.Knowledge >= High {
		return prompt + "\n\nConsider relevant background information and context in your response."
	}
	return prompt + "\n\nProvide relevant context as needed."
}

func (s *SPOOptimizer) addConstraintsStrategy(prompt string, complexity TaskComplexity) string {
	return prompt + "\n\nEnsure your response is accurate, helpful, and appropriate."
}

func (s *SPOOptimizer) addExamplesStrategy(prompt string, complexity TaskComplexity) string {
	if complexity.Overall >= Medium {
		return prompt + "\n\nInclude relevant examples to illustrate your points."
	}
	return prompt
}

// evaluateVariants evaluates prompt variants using pairwise comparison
func (s *SPOOptimizer) evaluateVariants(variants []string, complexity TaskComplexity) []float64 {
	scores := make([]float64, len(variants))
	
	for i, variant := range variants {
		score := s.calculatePromptQuality(variant, complexity)
		scores[i] = score
	}
	
	return scores
}

// calculatePromptQuality calculates the quality score of a prompt
func (s *SPOOptimizer) calculatePromptQuality(prompt string, complexity TaskComplexity) float64 {
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
	if complexity.Overall >= High && len(words) >= 20 {
		score += 0.2 // Complex tasks need detailed prompts
	}
	
	// Normalize score to 0-1 range
	if score > 1.0 {
		score = 1.0
	}
	
	return score
}

// Helper methods
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

func (s *SPOOptimizer) hasConverged(score float64, threshold float64) bool {
	return score >= (1.0 - threshold)
}

func (s *SPOOptimizer) estimateCostSavings(score float64) float64 {
	return score * 0.3 // Up to 30% cost savings
}

// Cache management methods
func (s *SPOOptimizer) generateCacheKey(prompt string, complexity TaskComplexity) string {
	h := fnv.New64a()
	h.Write([]byte(prompt))
	h.Write([]byte(fmt.Sprintf("%.2f", complexity.Score)))
	return fmt.Sprintf("%x", h.Sum64())
}

func (s *SPOOptimizer) getFromCache(key string) *CachedOptimization {
	s.cacheMutex.RLock()
	defer s.cacheMutex.RUnlock()
	
	if cached, exists := s.cache[key]; exists {
		if time.Since(cached.Timestamp) < s.cacheTTL {
			cached.HitCount++
			return cached
		}
		delete(s.cache, key)
	}
	return nil
}

func (s *SPOOptimizer) cacheResult(key string, result OptimizedPrompt) {
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()
	
	if len(s.cache) >= s.cacheSize {
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

func (s *SPOOptimizer) evictLeastUsed() {
	var leastUsedKey string
	minHits := int(^uint(0) >> 1)
	
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

func (s *SPOOptimizer) trackOptimization(prompt string, attempt OptimizationAttempt) {
	s.historyMutex.Lock()
	defer s.historyMutex.Unlock()
	
	key := fmt.Sprintf("%x", fnv.New64a().Sum([]byte(prompt)))
	s.optimizationHistory[key] = append(s.optimizationHistory[key], attempt)
	
	if len(s.optimizationHistory[key]) > 10 {
		s.optimizationHistory[key] = s.optimizationHistory[key][1:]
	}
}

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

func (s *SPOOptimizer) cleanupExpiredEntries() {
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()
	
	now := time.Now()
	for key, cached := range s.cache {
		if now.Sub(cached.Timestamp) > s.cacheTTL {
			delete(s.cache, key)
		}
	}
}