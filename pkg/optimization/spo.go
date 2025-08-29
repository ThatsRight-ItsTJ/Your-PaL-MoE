package optimization

import (
	"crypto/sha256"
	"fmt"
	"math"
	"strings"
	"sync"
)

// OptimizationResult represents the result of prompt optimization
type OptimizationResult struct {
	Original     string   `json:"original"`
	Optimized    string   `json:"optimized"`
	Improvements []string `json:"improvements"`
	Confidence   float64  `json:"confidence"`
	CostSavings  float64  `json:"cost_savings"`
}

// SPOOptimizer implements Self-Supervised Prompt Optimization
type SPOOptimizer struct {
	cache map[string]*OptimizationResult
	mutex sync.RWMutex
}

// NewSPOOptimizer creates a new SPO optimizer with caching
func NewSPOOptimizer() *SPOOptimizer {
	return &SPOOptimizer{
		cache: make(map[string]*OptimizationResult),
	}
}

// OptimizePrompt applies SPO strategies to enhance prompts
func (spo *SPOOptimizer) OptimizePrompt(original string, context map[string]interface{}) OptimizationResult {
	// Check cache first
	cacheKey := spo.generateCacheKey(original, context)
	
	spo.mutex.RLock()
	if cached, exists := spo.cache[cacheKey]; exists {
		spo.mutex.RUnlock()
		return *cached
	}
	spo.mutex.RUnlock()

	optimized := original
	var improvements []string

	// Strategy 1: Clarification Enhancement
	if len(original) < 50 {
		optimized += "\n\nPlease be specific and detailed in your response."
		improvements = append(improvements, "Added detailed guidance for short prompts")
	}

	// Strategy 2: Structure Addition
	if spo.needsStructure(original) {
		structure := spo.generateStructure(original, context)
		if structure != "" {
			optimized += "\n\n" + structure
			improvements = append(improvements, "Added analytical structure")
		}
	}

	// Strategy 3: Context Enhancement
	if domain, exists := context["domain"]; exists {
		contextPrompt := fmt.Sprintf("Context: This is a %s-related task.\n\n%s", domain, optimized)
		optimized = contextPrompt
		improvements = append(improvements, "Added domain context")
	}

	// Strategy 4: Constraint Addition
	constraints := spo.addConstraints(original, context)
	if constraints != "" {
		optimized += "\n\n" + constraints
		improvements = append(improvements, "Added helpful constraints")
	}

	// Strategy 5: Example Requests
	examples := spo.addExamples(original)
	if examples != "" {
		optimized += "\n\n" + examples
		improvements = append(improvements, "Requested concrete examples")
	}

	// Strategy 6: Output Format Specification
	if outputFormat, exists := context["output_format"]; exists {
		formatPrompt := fmt.Sprintf("\n\nPlease format your response as %s.", outputFormat)
		optimized += formatPrompt
		improvements = append(improvements, "Specified output format")
	}

	// Strategy 7: Quality Enhancement
	if spo.needsQualityGuidance(original) {
		qualityPrompt := "\n\nEnsure your response is accurate, well-reasoned, and comprehensive."
		optimized += qualityPrompt
		improvements = append(improvements, "Added quality guidance")
	}

	confidence := spo.calculateConfidence(original, optimized, improvements)
	costSavings := spo.estimateCostSavings(original, optimized, improvements)

	result := OptimizationResult{
		Original:     original,
		Optimized:    optimized,
		Improvements: improvements,
		Confidence:   confidence,
		CostSavings:  costSavings,
	}

	// Cache the result
	spo.mutex.Lock()
	spo.cache[cacheKey] = &result
	spo.mutex.Unlock()

	return result
}

// generateCacheKey creates a unique cache key for the optimization
func (spo *SPOOptimizer) generateCacheKey(original string, context map[string]interface{}) string {
	contextStr := ""
	for k, v := range context {
		contextStr += fmt.Sprintf("%s:%v;", k, v)
	}
	
	hash := sha256.Sum256([]byte(original + contextStr))
	return fmt.Sprintf("%x", hash)[:16] // Use first 16 chars of hash
}

// needsStructure determines if prompt needs structural enhancement
func (spo *SPOOptimizer) needsStructure(prompt string) bool {
	lowerPrompt := strings.ToLower(prompt)
	structureKeywords := []string{
		"analyze", "compare", "evaluate", "assess", "review",
		"explain", "describe", "outline", "summarize",
	}
	
	for _, keyword := range structureKeywords {
		if strings.Contains(lowerPrompt, keyword) {
			return true
		}
	}
	return false
}

// generateStructure creates appropriate structure based on prompt type
func (spo *SPOOptimizer) generateStructure(prompt string, context map[string]interface{}) string {
	lowerPrompt := strings.ToLower(prompt)
	
	if strings.Contains(lowerPrompt, "analyze") {
		return "Structure your analysis as follows:\n1. Overview\n2. Key findings\n3. Implications\n4. Recommendations\n5. Conclusion"
	}
	
	if strings.Contains(lowerPrompt, "compare") {
		return "Structure your comparison as follows:\n1. Introduction\n2. Similarities\n3. Differences\n4. Pros and cons\n5. Recommendation"
	}
	
	if strings.Contains(lowerPrompt, "explain") {
		return "Structure your explanation as follows:\n1. Definition/Overview\n2. Key concepts\n3. Examples\n4. Applications\n5. Summary"
	}
	
	if strings.Contains(lowerPrompt, "code") || strings.Contains(lowerPrompt, "program") {
		return "Structure your response as follows:\n1. Approach explanation\n2. Code implementation\n3. Usage examples\n4. Testing considerations\n5. Potential improvements"
	}
	
	return ""
}

// addConstraints adds helpful constraints based on context
func (spo *SPOOptimizer) addConstraints(prompt string, context map[string]interface{}) string {
	var constraints []string
	
	if maxTokens, exists := context["max_tokens"]; exists {
		constraints = append(constraints, fmt.Sprintf("Keep your response under %v tokens", maxTokens))
	}
	
	if strings.Contains(strings.ToLower(prompt), "technical") {
		constraints = append(constraints, "Use precise technical terminology")
	}
	
	if strings.Contains(strings.ToLower(prompt), "beginner") || strings.Contains(strings.ToLower(prompt), "simple") {
		constraints = append(constraints, "Explain concepts in simple, accessible language")
	}
	
	if len(constraints) > 0 {
		return "Constraints:\n- " + strings.Join(constraints, "\n- ")
	}
	
	return ""
}

// addExamples requests examples based on prompt type
func (spo *SPOOptimizer) addExamples(prompt string) string {
	lowerPrompt := strings.ToLower(prompt)
	
	if strings.Contains(lowerPrompt, "code") || strings.Contains(lowerPrompt, "program") {
		return "Provide concrete code examples where applicable."
	}
	
	if strings.Contains(lowerPrompt, "explain") || strings.Contains(lowerPrompt, "describe") {
		return "Include specific examples to illustrate your points."
	}
	
	if strings.Contains(lowerPrompt, "strategy") || strings.Contains(lowerPrompt, "approach") {
		return "Provide practical examples of implementation."
	}
	
	return ""
}

// needsQualityGuidance determines if prompt needs quality enhancement
func (spo *SPOOptimizer) needsQualityGuidance(prompt string) bool {
	// Add quality guidance for complex or important tasks
	lowerPrompt := strings.ToLower(prompt)
	qualityKeywords := []string{
		"important", "critical", "decision", "analysis", "research",
		"professional", "business", "academic", "scientific",
	}
	
	for _, keyword := range qualityKeywords {
		if strings.Contains(lowerPrompt, keyword) {
			return true
		}
	}
	
	// Also add for longer prompts (likely more complex)
	return len(prompt) > 200
}

// calculateConfidence estimates optimization confidence
func (spo *SPOOptimizer) calculateConfidence(original, optimized string, improvements []string) float64 {
	baseConfidence := 0.5
	
	// Improvement count bonus
	improvementBonus := float64(len(improvements)) * 0.1
	
	// Length improvement factor
	lengthRatio := float64(len(optimized)) / float64(len(original))
	lengthFactor := math.Min(lengthRatio-1.0, 1.0) * 0.2
	
	// Structural improvement bonus
	structuralBonus := 0.0
	for _, improvement := range improvements {
		if strings.Contains(improvement, "structure") || strings.Contains(improvement, "format") {
			structuralBonus += 0.15
		}
	}
	
	confidence := baseConfidence + improvementBonus + lengthFactor + structuralBonus
	return math.Min(confidence, 1.0)
}

// estimateCostSavings estimates potential cost savings from optimization
func (spo *SPOOptimizer) estimateCostSavings(original, optimized string, improvements []string) float64 {
	baseSavings := 0.05 // 5% baseline from clearer prompts
	
	// Structural improvements reduce need for follow-up questions
	structuralSavings := 0.0
	for _, improvement := range improvements {
		if strings.Contains(improvement, "structure") {
			structuralSavings += 0.10
		}
		if strings.Contains(improvement, "constraint") {
			structuralSavings += 0.05
		}
		if strings.Contains(improvement, "example") {
			structuralSavings += 0.05
		}
	}
	
	// Length optimization (more specific prompts often get better results faster)
	lengthRatio := float64(len(optimized)) / float64(len(original))
	if lengthRatio > 1.2 && lengthRatio < 2.0 {
		baseSavings += 0.10 // 10% savings from reduced iterations
	}
	
	totalSavings := baseSavings + structuralSavings
	return math.Min(totalSavings, 0.30) // Cap at 30% savings
}

// ClearCache clears the optimization cache
func (spo *SPOOptimizer) ClearCache() {
	spo.mutex.Lock()
	defer spo.mutex.Unlock()
	spo.cache = make(map[string]*OptimizationResult)
}

// GetCacheSize returns the current cache size
func (spo *SPOOptimizer) GetCacheSize() int {
	spo.mutex.RLock()
	defer spo.mutex.RUnlock()
	return len(spo.cache)
}