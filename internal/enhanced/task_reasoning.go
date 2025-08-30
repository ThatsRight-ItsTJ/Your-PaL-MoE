package enhanced

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// TaskReasoningEngine analyzes task complexity and requirements
type TaskReasoningEngine struct {
	logger *logrus.Logger
	
	// Analysis patterns
	complexityPatterns map[string]*regexp.Regexp
	domainPatterns     map[string]*regexp.Regexp
	intentPatterns     map[string]*regexp.Regexp
}

// NewTaskReasoningEngine creates a new task reasoning engine
func NewTaskReasoningEngine(logger *logrus.Logger) (*TaskReasoningEngine, error) {
	engine := &TaskReasoningEngine{
		logger: logger,
	}
	
	engine.initializePatterns()
	return engine, nil
}

// initializePatterns initializes regex patterns for analysis
func (t *TaskReasoningEngine) initializePatterns() {
	// Complexity indicators
	t.complexityPatterns = map[string]*regexp.Regexp{
		"high_complexity":   regexp.MustCompile(`(?i)\b(complex|sophisticated|advanced|intricate|comprehensive|detailed analysis|multi-step|research)\b`),
		"medium_complexity": regexp.MustCompile(`(?i)\b(analyze|compare|evaluate|design|create|develop|explain)\b`),
		"low_complexity":    regexp.MustCompile(`(?i)\b(simple|basic|quick|summarize|list|define|what is)\b`),
	}
	
	// Domain patterns
	t.domainPatterns = map[string]*regexp.Regexp{
		"code":     regexp.MustCompile(`(?i)\b(code|programming|function|algorithm|debug|implement|software)\b`),
		"math":     regexp.MustCompile(`(?i)\b(calculate|solve|equation|formula|mathematics|statistics)\b`),
		"analysis": regexp.MustCompile(`(?i)\b(analyze|research|study|investigate|examine|report)\b`),
		"creative": regexp.MustCompile(`(?i)\b(write|create|design|story|creative|artistic|brainstorm)\b`),
		"business": regexp.MustCompile(`(?i)\b(business|strategy|market|finance|proposal|plan)\b`),
	}
	
	// Intent patterns
	t.intentPatterns = map[string]*regexp.Regexp{
		"generation":         regexp.MustCompile(`(?i)\b(generate|create|write|produce|make|build)\b`),
		"analysis":          regexp.MustCompile(`(?i)\b(analyze|examine|study|investigate|review)\b`),
		"transformation":    regexp.MustCompile(`(?i)\b(transform|convert|translate|modify|change)\b`),
		"question_answering": regexp.MustCompile(`(?i)\b(what|how|why|when|where|explain|describe)\b`),
		"problem_solving":   regexp.MustCompile(`(?i)\b(solve|fix|debug|resolve|troubleshoot)\b`),
	}
}

// AnalyzeComplexity analyzes the complexity of a task request
func (t *TaskReasoningEngine) AnalyzeComplexity(ctx context.Context, input RequestInput) (TaskComplexity, error) {
	t.logger.Infof("Analyzing complexity for request: %s", input.ID)
	
	content := strings.ToLower(input.Content)
	
	// Analyze different dimensions of complexity
	reasoning := t.analyzeReasoningComplexity(content)
	knowledge := t.analyzeKnowledgeComplexity(content)
	computation := t.analyzeComputationComplexity(content)
	coordination := t.analyzeCoordinationComplexity(content)
	
	// Calculate overall complexity
	overall := t.calculateOverallComplexity(reasoning, knowledge, computation, coordination)
	
	// Calculate numerical score
	score := t.calculateComplexityScore(reasoning, knowledge, computation, coordination)
	
	complexity := TaskComplexity{
		Reasoning:    reasoning,
		Knowledge:    knowledge,
		Computation:  computation,
		Coordination: coordination,
		Overall:      overall,
		Score:        score,
	}
	
	t.logger.Infof("Complexity analysis for %s: overall=%s, score=%.2f", input.ID, overall.String(), score)
	return complexity, nil
}

// analyzeReasoningComplexity analyzes the reasoning complexity required
func (t *TaskReasoningEngine) analyzeReasoningComplexity(content string) ComplexityLevel {
	if t.complexityPatterns["high_complexity"].MatchString(content) {
		return VeryHigh
	}
	if t.complexityPatterns["medium_complexity"].MatchString(content) {
		return High
	}
	if t.complexityPatterns["low_complexity"].MatchString(content) {
		return Low
	}
	return Medium
}

// analyzeKnowledgeComplexity analyzes the domain knowledge required
func (t *TaskReasoningEngine) analyzeKnowledgeComplexity(content string) ComplexityLevel {
	domainMatches := 0
	specializedDomains := 0
	
	for domain, pattern := range t.domainPatterns {
		if pattern.MatchString(content) {
			domainMatches++
			if domain == "code" || domain == "math" {
				specializedDomains++
			}
		}
	}
	
	if specializedDomains >= 2 {
		return VeryHigh
	}
	if specializedDomains >= 1 {
		return High
	}
	if domainMatches >= 2 {
		return Medium
	}
	return Low
}

// analyzeComputationComplexity analyzes the computational complexity
func (t *TaskReasoningEngine) analyzeComputationComplexity(content string) ComplexityLevel {
	computationWords := []string{
		"calculate", "compute", "process", "algorithm", "optimize",
		"iterate", "recursive", "complex analysis", "big data",
	}
	
	matches := 0
	for _, word := range computationWords {
		if strings.Contains(content, word) {
			matches++
		}
	}
	
	if matches >= 3 {
		return VeryHigh
	} else if matches >= 2 {
		return High
	} else if matches >= 1 {
		return Medium
	}
	return Low
}

// analyzeCoordinationComplexity analyzes the coordination complexity
func (t *TaskReasoningEngine) analyzeCoordinationComplexity(content string) ComplexityLevel {
	coordinationWords := []string{
		"multiple", "various", "several", "combine", "integrate",
		"coordinate", "orchestrate", "parallel", "simultaneous",
	}
	
	matches := 0
	for _, word := range coordinationWords {
		if strings.Contains(content, word) {
			matches++
		}
	}
	
	if matches >= 3 {
		return VeryHigh
	} else if matches >= 2 {
		return High
	} else if matches >= 1 {
		return Medium
	}
	return Low
}

// calculateOverallComplexity calculates the overall complexity level
func (t *TaskReasoningEngine) calculateOverallComplexity(reasoning, knowledge, computation, coordination ComplexityLevel) ComplexityLevel {
	total := int(reasoning) + int(knowledge) + int(computation) + int(coordination)
	average := float64(total) / 4.0
	
	if average >= 3.0 {
		return VeryHigh
	} else if average >= 2.0 {
		return High
	} else if average >= 1.0 {
		return Medium
	}
	return Low
}

// calculateComplexityScore calculates a numerical complexity score (0.0 to 1.0)
func (t *TaskReasoningEngine) calculateComplexityScore(reasoning, knowledge, computation, coordination ComplexityLevel) float64 {
	weights := map[string]float64{
		"reasoning":    0.3,
		"knowledge":    0.25,
		"computation":  0.25,
		"coordination": 0.2,
	}
	
	score := weights["reasoning"]*float64(reasoning)/3.0 +
		weights["knowledge"]*float64(knowledge)/3.0 +
		weights["computation"]*float64(computation)/3.0 +
		weights["coordination"]*float64(coordination)/3.0
	
	return score
}

// ExtractRequirements extracts requirements from the task input
func (t *TaskReasoningEngine) ExtractRequirements(ctx context.Context, input RequestInput) (map[string]interface{}, error) {
	requirements := make(map[string]interface{})
	content := strings.ToLower(input.Content)
	
	// Extract domain
	for domain, pattern := range t.domainPatterns {
		if pattern.MatchString(content) {
			requirements["domain"] = domain
			break
		}
	}
	
	// Extract intent
	for intent, pattern := range t.intentPatterns {
		if pattern.MatchString(content) {
			requirements["intent"] = intent
			break
		}
	}
	
	// Extract quality requirements
	if strings.Contains(content, "high quality") || strings.Contains(content, "detailed") {
		requirements["quality_level"] = "high"
	} else if strings.Contains(content, "quick") || strings.Contains(content, "fast") {
		requirements["quality_level"] = "medium"
	} else {
		requirements["quality_level"] = "standard"
	}
	
	// Extract output format
	if strings.Contains(content, "json") {
		requirements["output_format"] = "json"
	} else if strings.Contains(content, "markdown") {
		requirements["output_format"] = "markdown"
	} else if strings.Contains(content, "code") {
		requirements["output_format"] = "code"
	} else {
		requirements["output_format"] = "text"
	}
	
	// Extract constraints from context
	if input.Constraints != nil {
		for key, value := range input.Constraints {
			requirements[key] = value
		}
	}
	
	return requirements, nil
}