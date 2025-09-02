package enhanced

import (
	"fmt"
	"strings"
)

// Implementation moved to constructors.go to avoid duplicate method declarations
// This file now only contains helper functions and constants

// optimizationTemplates provides prompt optimization templates
var optimizationTemplates = map[ComplexityLevel][]string{
	Low: {
		"Please provide a clear and concise answer to: {prompt}",
		"Simply explain: {prompt}",
		"Answer directly: {prompt}",
	},
	Medium: {
		"Please analyze and explain: {prompt}",
		"Provide a detailed response to: {prompt}",
		"Consider the following and respond thoughtfully: {prompt}",
	},
	High: {
		"Please provide a comprehensive analysis of: {prompt}",
		"Think step by step and address: {prompt}",
		"Carefully consider all aspects and respond to: {prompt}",
	},
	VeryHigh: {
		"Please provide a thorough, multi-faceted analysis of: {prompt}",
		"Use chain-of-thought reasoning to address: {prompt}",
		"Systematically break down and comprehensively address: {prompt}",
	},
}

// complexityRules defines optimization rules for different complexity levels
var complexityRules = map[ComplexityLevel][]string{
	Low: {
		"Keep responses concise and direct",
		"Use simple language and structure",
		"Focus on the core question",
	},
	Medium: {
		"Provide context and explanation",
		"Use examples where helpful",
		"Structure the response clearly",
	},
	High: {
		"Break down complex problems into steps",
		"Provide detailed reasoning",
		"Consider multiple perspectives",
	},
	VeryHigh: {
		"Use systematic analysis approach",
		"Employ chain-of-thought reasoning",
		"Address edge cases and nuances",
	},
}

// applyOptimizationRules applies complexity-specific optimization rules
func applyOptimizationRules(prompt string, complexity ComplexityLevel) string {
	rules, exists := complexityRules[complexity]
	if !exists {
		return prompt
	}
	
	optimized := prompt
	
	// Apply basic formatting based on complexity
	switch complexity {
	case Low:
		optimized = fmt.Sprintf("Please provide a brief, direct answer: %s", prompt)
	case Medium:
		optimized = fmt.Sprintf("Please explain clearly: %s", prompt)
	case High:
		optimized = fmt.Sprintf("Please analyze step by step: %s", prompt)
	case VeryHigh:
		optimized = fmt.Sprintf("Please provide a comprehensive, systematic analysis: %s", prompt)
	}
	
	return optimized
}

// selectOptimizationTemplate selects the best template for the complexity level
func selectOptimizationTemplate(complexity ComplexityLevel, promptLength int) string {
	templates, exists := optimizationTemplates[complexity]
	if !exists || len(templates) == 0 {
		return "{prompt}" // Fallback template
	}
	
	// Select template based on prompt length
	if promptLength < 50 {
		return templates[0] // Simple template for short prompts
	} else if promptLength < 200 {
		if len(templates) > 1 {
			return templates[1] // Medium template
		}
		return templates[0]
	} else {
		if len(templates) > 2 {
			return templates[2] // Complex template for long prompts
		}
		return templates[len(templates)-1]
	}
}

// enhancePromptForComplexity adds complexity-specific enhancements
func enhancePromptForComplexity(prompt string, complexity TaskComplexity) string {
	enhanced := prompt
	
	// Add reasoning instructions for high complexity
	if complexity.Reasoning >= High {
		enhanced = "Think step by step. " + enhanced
	}
	
	// Add mathematical precision for math tasks
	if complexity.Mathematical >= Medium {
		enhanced += " Please show your work and calculations."
	}
	
	// Add creativity instructions for creative tasks
	if complexity.Creative >= Medium {
		enhanced += " Be creative and original in your response."
	}
	
	// Add factual accuracy instructions
	if complexity.Factual >= Medium {
		enhanced += " Please ensure accuracy and cite sources where appropriate."
	}
	
	return enhanced
}

// validateOptimizedPrompt ensures the optimized prompt meets quality standards
func validateOptimizedPrompt(original, optimized string) bool {
	// Basic validation checks
	if len(optimized) == 0 {
		return false
	}
	
	// Ensure we haven't lost the core content
	originalWords := strings.Fields(strings.ToLower(original))
	optimizedWords := strings.Fields(strings.ToLower(optimized))
	
	// Check if key words from original are preserved
	keyWordsPreserved := 0
	for _, word := range originalWords {
		if len(word) > 3 { // Only check meaningful words
			for _, optWord := range optimizedWords {
				if word == optWord {
					keyWordsPreserved++
					break
				}
			}
		}
	}
	
	// At least 70% of meaningful words should be preserved
	meaningfulWords := 0
	for _, word := range originalWords {
		if len(word) > 3 {
			meaningfulWords++
		}
	}
	
	if meaningfulWords > 0 {
		preservationRatio := float64(keyWordsPreserved) / float64(meaningfulWords)
		return preservationRatio >= 0.7
	}
	
	return true // If no meaningful words to check, assume valid
}