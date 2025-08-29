package analysis

import (
	"math"
	"regexp"
	"strings"
)

// TaskComplexity represents multi-dimensional task analysis
type TaskComplexity struct {
	Reasoning    float64 `json:"reasoning"`
	Knowledge    float64 `json:"knowledge"`
	Computation  float64 `json:"computation"`
	Coordination float64 `json:"coordination"`
	Overall      float64 `json:"overall"`
	Score        float64 `json:"score"`
}

// ComplexityAnalyzer analyzes task complexity using pattern matching
type ComplexityAnalyzer struct {
	reasoningPatterns    []*regexp.Regexp
	knowledgePatterns    []*regexp.Regexp
	computationPatterns  []*regexp.Regexp
	coordinationPatterns []*regexp.Regexp
}

// NewComplexityAnalyzer creates a new complexity analyzer
func NewComplexityAnalyzer() *ComplexityAnalyzer {
	return &ComplexityAnalyzer{
		reasoningPatterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)\b(analyze|reasoning|logic|deduce|infer|conclude|evaluate)\b`),
			regexp.MustCompile(`(?i)\b(compare|contrast|pros and cons|advantages|disadvantages)\b`),
			regexp.MustCompile(`(?i)\b(strategy|approach|methodology|framework|systematic)\b`),
			regexp.MustCompile(`(?i)\b(critical thinking|decision|judgment|assessment|analysis)\b`),
		},
		knowledgePatterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)\b(research|facts|information|data|statistics|studies)\b`),
			regexp.MustCompile(`(?i)\b(history|background|context|domain|field|expertise)\b`),
			regexp.MustCompile(`(?i)\b(technical|scientific|academic|professional|specialized)\b`),
			regexp.MustCompile(`(?i)\b(explain|describe|define|clarify|elaborate)\b`),
		},
		computationPatterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)\b(calculate|compute|algorithm|formula|equation|math)\b`),
			regexp.MustCompile(`(?i)\b(process|transform|convert|parse|generate|create)\b`),
			regexp.MustCompile(`(?i)\b(code|program|script|function|implementation)\b`),
			regexp.MustCompile(`(?i)\b(optimization|performance|efficiency|automation)\b`),
		},
		coordinationPatterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)\b(coordinate|manage|organize|plan|schedule|workflow)\b`),
			regexp.MustCompile(`(?i)\b(multiple|several|various|different|diverse|complex)\b`),
			regexp.MustCompile(`(?i)\b(integrate|combine|merge|synthesize|consolidate)\b`),
			regexp.MustCompile(`(?i)\b(collaborate|teamwork|multi-step|sequential|parallel)\b`),
		},
	}
}

// AnalyzeTask performs multi-dimensional complexity analysis
func (ca *ComplexityAnalyzer) AnalyzeTask(prompt string, context map[string]interface{}) TaskComplexity {
	text := strings.ToLower(prompt)

	reasoning := ca.scorePatterns(text, ca.reasoningPatterns)
	knowledge := ca.scorePatterns(text, ca.knowledgePatterns)
	computation := ca.scorePatterns(text, ca.computationPatterns)
	coordination := ca.scorePatterns(text, ca.coordinationPatterns)

	// Apply context modifiers
	if domain, exists := context["domain"]; exists {
		switch domain {
		case "ai", "technical", "engineering":
			knowledge += 0.3
			reasoning += 0.2
		case "creative", "marketing":
			reasoning += 0.2
			coordination += 0.1
		case "data", "analytics":
			computation += 0.3
			knowledge += 0.2
		}
	}

	// Apply length modifiers (longer prompts often indicate complexity)
	lengthFactor := math.Min(float64(len(prompt))/500.0, 1.0) * 0.2
	reasoning += lengthFactor
	knowledge += lengthFactor

	// Apply question complexity modifiers
	questionCount := strings.Count(text, "?")
	if questionCount > 1 {
		coordination += float64(questionCount) * 0.1
	}

	// Normalize scores to 0-3 range
	reasoning = math.Min(reasoning, 3.0)
	knowledge = math.Min(knowledge, 3.0)
	computation = math.Min(computation, 3.0)
	coordination = math.Min(coordination, 3.0)

	overall := (reasoning + knowledge + computation + coordination) / 4.0
	score := overall / 3.0 // Normalize to 0-1 range

	return TaskComplexity{
		Reasoning:    reasoning,
		Knowledge:    knowledge,
		Computation:  computation,
		Coordination: coordination,
		Overall:      overall,
		Score:        score,
	}
}

// scorePatterns calculates score based on pattern matches
func (ca *ComplexityAnalyzer) scorePatterns(text string, patterns []*regexp.Regexp) float64 {
	score := 0.0
	for _, pattern := range patterns {
		matches := pattern.FindAllString(text, -1)
		// Each match adds 0.1, with diminishing returns
		matchScore := float64(len(matches)) * 0.1
		if matchScore > 0.5 {
			matchScore = 0.5 + (matchScore-0.5)*0.5 // Diminishing returns after 0.5
		}
		score += matchScore
	}
	return score
}

// GetComplexityDescription returns human-readable complexity description
func (tc TaskComplexity) GetComplexityDescription() string {
	if tc.Score < 0.3 {
		return "Low complexity - suitable for basic models"
	} else if tc.Score < 0.6 {
		return "Medium complexity - requires capable models"
	} else if tc.Score < 0.8 {
		return "High complexity - needs advanced models"
	} else {
		return "Very high complexity - requires top-tier models"
	}
}

// GetDominantDimension returns the dimension with highest score
func (tc TaskComplexity) GetDominantDimension() string {
	max := tc.Reasoning
	dimension := "reasoning"

	if tc.Knowledge > max {
		max = tc.Knowledge
		dimension = "knowledge"
	}
	if tc.Computation > max {
		max = tc.Computation
		dimension = "computation"
	}
	if tc.Coordination > max {
		dimension = "coordination"
	}

	return dimension
}