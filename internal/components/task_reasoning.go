package components

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/ThatsRight-ItsTJ/Your-PaL-MoE/internal/enhanced"
)

// TaskReasoner analyzes task complexity and requirements
type TaskReasoner struct {
	config *TaskReasonerConfig
}

// TaskReasonerConfig represents configuration for the task reasoner
type TaskReasonerConfig struct {
	ComplexityWeights map[string]float64                   `json:"complexity_weights"`
	TokenMultipliers  map[enhanced.ComplexityLevel]float64 `json:"token_multipliers"`
}

// NewTaskReasoner creates a new task reasoner with default configuration
func NewTaskReasoner() *TaskReasoner {
	return &TaskReasoner{
		config: &TaskReasonerConfig{
			ComplexityWeights: map[string]float64{
				"reasoning":    0.3,
				"mathematical": 0.25,
				"creative":     0.2,
				"factual":      0.25,
			},
			TokenMultipliers: map[enhanced.ComplexityLevel]float64{
				enhanced.Low:      1.0,
				enhanced.Medium:   1.2,
				enhanced.High:     1.5,
				enhanced.VeryHigh: 2.0,
			},
		},
	}
}

// AnalyzeComplexity analyzes the complexity of a task
func (tr *TaskReasoner) AnalyzeComplexity(content string) (*enhanced.TaskComplexity, error) {
	if strings.TrimSpace(content) == "" {
		return nil, fmt.Errorf("content cannot be empty")
	}

	// Analyze different aspects of complexity
	reasoning := tr.detectReasoningComplexity(content)
	mathematical := tr.detectMathematicalComplexity(content)
	creative := tr.detectCreativeComplexity(content)
	factual := tr.detectFactualComplexity(content)

	// Calculate overall complexity
	overall := tr.determineOverallComplexity(reasoning, mathematical, creative, factual)

	// Estimate tokens
	tokenEstimate := tr.estimateTokensFromText(content, overall)

	// Determine required capabilities
	capabilities := tr.determineRequiredCapabilities(reasoning, mathematical, creative, factual)

	return &enhanced.TaskComplexity{
		Overall:              overall,
		Reasoning:            reasoning,
		Mathematical:         mathematical,
		Creative:             creative,
		Factual:              factual,
		TokenEstimate:        tokenEstimate,
		RequiredCapabilities: capabilities,
		Metadata:             make(map[string]interface{}),
	}, nil
}

// detectReasoningComplexity analyzes reasoning complexity
func (tr *TaskReasoner) detectReasoningComplexity(content string) enhanced.ComplexityLevel {
	content = strings.ToLower(content)
	
	reasoningPatterns := []*regexp.Regexp{
		regexp.MustCompile(`\b(because|therefore|however|although|since)\b`),
		regexp.MustCompile(`\b(analyze|evaluate|compare|contrast|argue)\b`),
		regexp.MustCompile(`\b(logic|reasoning|conclusion|premise|inference)\b`),
	}
	
	score := 0
	for _, pattern := range reasoningPatterns {
		matches := pattern.FindAllString(content, -1)
		score += len(matches)
	}
	
	switch {
	case score >= 5:
		return enhanced.VeryHigh
	case score >= 3:
		return enhanced.High
	case score >= 1:
		return enhanced.Medium
	default:
		return enhanced.Low
	}
}

// detectMathematicalComplexity analyzes mathematical complexity
func (tr *TaskReasoner) detectMathematicalComplexity(content string) enhanced.ComplexityLevel {
	content = strings.ToLower(content)
	
	mathPatterns := []*regexp.Regexp{
		regexp.MustCompile(`\d+\s*[\+\-\*\/\^]\s*\d+`),
		regexp.MustCompile(`\b(equation|formula|calculate|solve|derivative|integral)\b`),
		regexp.MustCompile(`\b(algebra|geometry|calculus|statistics|probability)\b`),
	}
	
	score := 0
	for _, pattern := range mathPatterns {
		if pattern.MatchString(content) {
			score++
		}
	}
	
	// Advanced math keywords
	advancedMath := []string{"calculus", "differential", "integral", "matrix", "vector", "theorem"}
	for _, keyword := range advancedMath {
		if strings.Contains(content, keyword) {
			score += 2
		}
	}
	
	switch {
	case score >= 4:
		return enhanced.VeryHigh
	case score >= 2:
		return enhanced.High
	case score >= 1:
		return enhanced.Medium
	default:
		return enhanced.Low
	}
}

// detectCreativeComplexity analyzes creative complexity
func (tr *TaskReasoner) detectCreativeComplexity(content string) enhanced.ComplexityLevel {
	content = strings.ToLower(content)
	
	creativePatterns := []*regexp.Regexp{
		regexp.MustCompile(`\b(write|create|generate|compose|design)\b`),
		regexp.MustCompile(`\b(story|poem|article|essay|creative)\b`),
		regexp.MustCompile(`\b(imagine|brainstorm|invent|original)\b`),
	}
	
	score := 0
	for _, pattern := range creativePatterns {
		if pattern.MatchString(content) {
			score++
		}
	}
	
	// Check for creative requirements
	creativeKeywords := []string{"original", "unique", "innovative", "artistic", "imaginative"}
	for _, keyword := range creativeKeywords {
		if strings.Contains(content, keyword) {
			score++
		}
	}
	
	switch {
	case score >= 3:
		return enhanced.High
	case score >= 2:
		return enhanced.Medium
	case score >= 1:
		return enhanced.Medium
	default:
		return enhanced.Low
	}
}

// detectFactualComplexity analyzes factual complexity
func (tr *TaskReasoner) detectFactualComplexity(content string) enhanced.ComplexityLevel {
	content = strings.ToLower(content)
	
	factualPatterns := []*regexp.Regexp{
		regexp.MustCompile(`\b(fact|data|information|research|study)\b`),
		regexp.MustCompile(`\b(when|where|who|what|how)\b`),
		regexp.MustCompile(`\b(define|explain|describe|list)\b`),
	}
	
	score := 0
	for _, pattern := range factualPatterns {
		matches := pattern.FindAllString(content, -1)
		score += len(matches)
	}
	
	switch {
	case score >= 5:
		return enhanced.VeryHigh
	case score >= 3:
		return enhanced.High
	case score >= 1:
		return enhanced.Medium
	default:
		return enhanced.Low
	}
}

// determineOverallComplexity calculates overall complexity from components
func (tr *TaskReasoner) determineOverallComplexity(reasoning, mathematical, creative, factual enhanced.ComplexityLevel) enhanced.ComplexityLevel {
	// Calculate weighted average
	total := int(reasoning)*2 + int(mathematical)*2 + int(creative) + int(factual)
	average := total / 6
	
	// Ensure we don't exceed bounds
	if average > int(enhanced.VeryHigh) {
		return enhanced.VeryHigh
	}
	if average < int(enhanced.Low) {
		return enhanced.Low
	}
	
	return enhanced.ComplexityLevel(average)
}

// estimateTokensFromText provides a simple token estimation
func (tr *TaskReasoner) estimateTokensFromText(text string, complexity enhanced.ComplexityLevel) int64 {
	words := strings.Fields(text)
	baseTokens := float64(len(words)) * 1.33 // Rough approximation: 1 token per 0.75 words
	
	multiplier, exists := tr.config.TokenMultipliers[complexity]
	if !exists {
		multiplier = 1.0
	}
	
	return int64(baseTokens * multiplier)
}

// determineRequiredCapabilities determines what capabilities are needed
func (tr *TaskReasoner) determineRequiredCapabilities(reasoning, mathematical, creative, factual enhanced.ComplexityLevel) []string {
	var capabilities []string
	
	if reasoning >= enhanced.Medium {
		capabilities = append(capabilities, "reasoning")
	}
	if mathematical >= enhanced.Medium {
		capabilities = append(capabilities, "mathematical")
	}
	if creative >= enhanced.Medium {
		capabilities = append(capabilities, "creative")
	}
	if factual >= enhanced.Medium {
		capabilities = append(capabilities, "factual")
	}
	
	return capabilities
}