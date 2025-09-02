package enhanced

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"
)

// OptimizePrompt optimizes a prompt for better performance and cost efficiency
func (spo *SPOOptimizer) OptimizePrompt(ctx context.Context, input RequestInput, complexity *TaskComplexity) (*OptimizedPrompt, error) {
	if input.Query == "" {
		return nil, fmt.Errorf("empty query provided")
	}

	originalPrompt := input.Query
	optimizedText := spo.performOptimization(originalPrompt, complexity)
	
	// Calculate token reduction
	originalTokens := spo.estimateTokens(originalPrompt)
	optimizedTokens := spo.estimateTokens(optimizedText)
	tokenReduction := originalTokens - optimizedTokens
	
	// Calculate cost savings (rough estimate)
	costSavings := spo.calculateCostSavings(tokenReduction)
	
	reasoning := spo.generateOptimizationReasoning(originalPrompt, optimizedText, complexity)
	
	result := &OptimizedPrompt{
		OriginalPrompt: originalPrompt,
		OptimizedText:  optimizedText,
		Reasoning:      reasoning,
		TokenReduction: tokenReduction,
		CostSavings:    costSavings,
	}
	
	log.Printf("Prompt optimization: %d tokens reduced, $%.4f cost savings", tokenReduction, costSavings)
	
	return result, nil
}

// performOptimization performs the actual prompt optimization
func (spo *SPOOptimizer) performOptimization(prompt string, complexity *TaskComplexity) string {
	optimized := prompt
	
	// Remove redundant words and phrases
	optimized = spo.removeRedundancy(optimized)
	
	// Simplify based on complexity
	if complexity.Overall < 2.5 {
		optimized = spo.simplifyForLowComplexity(optimized)
	} else if complexity.Overall > 3.5 {
		optimized = spo.enhanceForHighComplexity(optimized)
	}
	
	// Apply general optimizations
	optimized = spo.applyGeneralOptimizations(optimized)
	
	return optimized
}

// removeRedundancy removes redundant words and phrases
func (spo *SPOOptimizer) removeRedundancy(text string) string {
	// Remove common redundant phrases
	redundantPhrases := []string{
		"please ", "could you ", "would you ", "can you ",
		"I would like ", "I want ", "I need ",
	}
	
	result := text
	for _, phrase := range redundantPhrases {
		result = strings.ReplaceAll(result, phrase, "")
	}
	
	return strings.TrimSpace(result)
}

// simplifyForLowComplexity simplifies prompts for low complexity tasks
func (spo *SPOOptimizer) simplifyForLowComplexity(text string) string {
	// For low complexity tasks, we can be more direct
	simplified := text
	
	// Remove unnecessary qualifiers
	qualifiers := []string{
		"very ", "quite ", "rather ", "somewhat ", "fairly ",
	}
	
	for _, qualifier := range qualifiers {
		simplified = strings.ReplaceAll(simplified, qualifier, "")
	}
	
	return simplified
}

// enhanceForHighComplexity enhances prompts for high complexity tasks
func (spo *SPOOptimizer) enhanceForHighComplexity(text string) string {
	// For high complexity tasks, we might need to add structure
	enhanced := text
	
	// Add thinking prompts for complex reasoning
	if strings.Contains(strings.ToLower(text), "analyze") || strings.Contains(strings.ToLower(text), "compare") {
		enhanced = "Think step by step. " + enhanced
	}
	
	return enhanced
}

// applyGeneralOptimizations applies general optimization rules
func (spo *SPOOptimizer) applyGeneralOptimizations(text string) string {
	optimized := text
	
	// Remove extra whitespace
	optimized = strings.Join(strings.Fields(optimized), " ")
	
	// Ensure proper capitalization
	if len(optimized) > 0 {
		optimized = strings.ToUpper(string(optimized[0])) + optimized[1:]
	}
	
	return optimized
}

// estimateTokens estimates the number of tokens in text
func (spo *SPOOptimizer) estimateTokens(text string) int {
	// Rough estimation: 1 token per 4 characters on average
	return len(text) / 4
}

// calculateCostSavings calculates estimated cost savings from token reduction
func (spo *SPOOptimizer) calculateCostSavings(tokenReduction int) float64 {
	// Assume average cost per token (rough estimate)
	avgCostPerToken := 0.00002 // $0.00002 per token
	return float64(tokenReduction) * avgCostPerToken
}

// generateOptimizationReasoning generates reasoning for the optimization
func (spo *SPOOptimizer) generateOptimizationReasoning(original, optimized string, complexity *TaskComplexity) string {
	reasoning := "Prompt optimization applied: "
	
	if len(optimized) < len(original) {
		reduction := len(original) - len(optimized)
		reasoning += fmt.Sprintf("Reduced length by %d characters. ", reduction)
	}
	
	complexityLevel := FloatToComplexityLevel(complexity.Overall)
	switch complexityLevel {
	case VeryHigh:
		reasoning += "Enhanced for high complexity task with structured thinking prompts."
	case High:
		reasoning += "Optimized for high complexity while maintaining clarity."
	case Medium:
		reasoning += "Balanced optimization for medium complexity task."
	case Low:
		reasoning += "Simplified for straightforward low complexity task."
	}
	
	return reasoning
}

// AnalyzePromptComplexity analyzes the complexity of a prompt
func (spo *SPOOptimizer) AnalyzePromptComplexity(prompt string) *TaskComplexity {
	// This is a simplified version - in practice, this would use more sophisticated analysis
	score := 1.0
	
	// Check for complexity indicators
	complexityIndicators := []string{
		"analyze", "compare", "evaluate", "synthesize", "explain",
		"detailed", "comprehensive", "step by step", "reasoning",
	}
	
	for _, indicator := range complexityIndicators {
		if strings.Contains(strings.ToLower(prompt), indicator) {
			score += 0.5
		}
	}
	
	if score > 4.0 {
		score = 4.0
	}
	
	return &TaskComplexity{
		Overall: score,
		Score:   score,
		Reasoning: score,
		Knowledge: score,
		Computation: score,
		Coordination: score,
	}
}

// OptimizeForProvider optimizes a prompt for a specific provider
func (spo *SPOOptimizer) OptimizeForProvider(prompt string, providerID string) string {
	// Provider-specific optimizations could be added here
	// For now, just return the original prompt
	return prompt
}

// BatchOptimize optimizes multiple prompts at once
func (spo *SPOOptimizer) BatchOptimize(ctx context.Context, prompts []string) ([]*OptimizedPrompt, error) {
	var results []*OptimizedPrompt
	
	for _, prompt := range prompts {
		input := RequestInput{Query: prompt}
		complexity := spo.AnalyzePromptComplexity(prompt)
		
		optimized, err := spo.OptimizePrompt(ctx, input, complexity)
		if err != nil {
			log.Printf("Failed to optimize prompt: %v", err)
			continue
		}
		
		results = append(results, optimized)
	}
	
	return results, nil
}