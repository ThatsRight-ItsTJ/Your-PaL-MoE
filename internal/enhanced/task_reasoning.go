package enhanced

import (
	"fmt"
	"regexp"
	"strings"
)

// Implementation moved to constructors.go to avoid duplicate method declarations
// This file now only contains helper functions and constants

// complexityKeywords maps keywords to complexity indicators
var complexityKeywords = map[string]ComplexityLevel{
	// Low complexity indicators
	"simple":     Low,
	"basic":      Low,
	"easy":       Low,
	"straightforward": Low,
	
	// Medium complexity indicators
	"moderate":   Medium,
	"compare":    Medium,
	"analyze":    Medium,
	"explain":    Medium,
	
	// High complexity indicators
	"complex":    High,
	"difficult":  High,
	"advanced":   High,
	"intricate":  High,
	
	// Very high complexity indicators
	"extremely":  VeryHigh,
	"highly":     VeryHigh,
	"very":       VeryHigh,
	"multi-step": VeryHigh,
}

// mathematicalPatterns identifies mathematical content
var mathematicalPatterns = []*regexp.Regexp{
	regexp.MustCompile(`\d+\s*[\+\-\*\/\^]\s*\d+`),
	regexp.MustCompile(`\b(equation|formula|calculate|solve|derivative|integral)\b`),
	regexp.MustCompile(`\b(algebra|geometry|calculus|statistics|probability)\b`),
	regexp.MustCompile(`[∑∏∫∂∇∆]`),
}

// creativePatterns identifies creative content
var creativePatterns = []*regexp.Regexp{
	regexp.MustCompile(`\b(write|create|generate|compose|design)\b`),
	regexp.MustCompile(`\b(story|poem|article|essay|creative)\b`),
	regexp.MustCompile(`\b(imagine|brainstorm|invent|original)\b`),
}

// reasoningPatterns identifies reasoning content
var reasoningPatterns = []*regexp.Regexp{
	regexp.MustCompile(`\b(because|therefore|however|although|since)\b`),
	regexp.MustCompile(`\b(analyze|evaluate|compare|contrast|argue)\b`),
	regexp.MustCompile(`\b(logic|reasoning|conclusion|premise|inference)\b`),
}

// estimateTokensFromText provides a simple token estimation
func estimateTokensFromText(text string) int64 {
	words := strings.Fields(text)
	// Rough approximation: 1 token per 0.75 words
	return int64(float64(len(words)) * 1.33)
}

// detectMathematicalComplexity analyzes mathematical complexity
func detectMathematicalComplexity(content string) ComplexityLevel {
	content = strings.ToLower(content)
	
	mathScore := 0
	for _, pattern := range mathematicalPatterns {
		if pattern.MatchString(content) {
			mathScore++
		}
	}
	
	// Advanced math keywords
	advancedMath := []string{"calculus", "differential", "integral", "matrix", "vector", "theorem"}
	for _, keyword := range advancedMath {
		if strings.Contains(content, keyword) {
			mathScore += 2
		}
	}
	
	switch {
	case mathScore >= 4:
		return VeryHigh
	case mathScore >= 2:
		return High
	case mathScore >= 1:
		return Medium
	default:
		return Low
	}
}

// detectCreativeComplexity analyzes creative complexity
func detectCreativeComplexity(content string) ComplexityLevel {
	content = strings.ToLower(content)
	
	creativeScore := 0
	for _, pattern := range creativePatterns {
		if pattern.MatchString(content) {
			creativeScore++
		}
	}
	
	// Check for creative requirements
	creativeKeywords := []string{"original", "unique", "innovative", "artistic", "imaginative"}
	for _, keyword := range creativeKeywords {
		if strings.Contains(content, keyword) {
			creativeScore++
		}
	}
	
	switch {
	case creativeScore >= 3:
		return High
	case creativeScore >= 2:
		return Medium
	case creativeScore >= 1:
		return Medium
	default:
		return Low
	}
}

// detectReasoningComplexity analyzes reasoning complexity
func detectReasoningComplexity(content string) ComplexityLevel {
	content = strings.ToLower(content)
	
	reasoningScore := 0
	for _, pattern := range reasoningPatterns {
		matches := pattern.FindAllString(content, -1)
		reasoningScore += len(matches)
	}
	
	// Check for complex reasoning indicators
	complexReasoning := []string{"multi-step", "chain of thought", "logical", "deductive", "inductive"}
	for _, keyword := range complexReasoning {
		if strings.Contains(content, keyword) {
			reasoningScore += 2
		}
	}
	
	switch {
	case reasoningScore >= 5:
		return VeryHigh
	case reasoningScore >= 3:
		return High
	case reasoningScore >= 1:
		return Medium
	default:
		return Low
	}
}

// determineOverallComplexity calculates overall complexity from components
func determineOverallComplexity(reasoning, mathematical, creative, factual ComplexityLevel) ComplexityLevel {
	// Calculate weighted average
	total := int(reasoning)*2 + int(mathematical)*2 + int(creative) + int(factual)
	average := total / 6
	
	// Ensure we don't exceed bounds
	if average > int(VeryHigh) {
		return VeryHigh
	}
	if average < int(Low) {
		return Low
	}
	
	return ComplexityLevel(average)
}