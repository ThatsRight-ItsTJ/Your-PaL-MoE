package enhanced

import (
	"context"
	"fmt"
	"strings"
)

// AnalyzeComplexity analyzes the complexity of a given task
func (tr *TaskReasoner) AnalyzeComplexity(ctx context.Context, input RequestInput) (*TaskComplexity, error) {
	if input.Query == "" {
		return nil, fmt.Errorf("empty query provided")
	}

	// Analyze different aspects of complexity
	reasoning := tr.analyzeReasoningComplexity(input.Query)
	knowledge := tr.analyzeKnowledgeComplexity(input.Query)
	computation := tr.analyzeComputationComplexity(input.Query)
	coordination := tr.analyzeCoordinationComplexity(input.Query)

	// Calculate overall complexity
	overall := (reasoning + knowledge + computation + coordination) / 4.0

	complexity := &TaskComplexity{
		Reasoning:     reasoning,
		Knowledge:     knowledge,
		Computation:   computation,
		Coordination:  coordination,
		Overall:       overall,
		Score:         overall,
	}

	fmt.Printf("Task complexity analysis: Overall=%.2f, Reasoning=%.2f, Knowledge=%.2f, Computation=%.2f, Coordination=%.2f\n",
		overall, reasoning, knowledge, computation, coordination)

	return complexity, nil
}

// analyzeReasoningComplexity analyzes the reasoning complexity of a task
func (tr *TaskReasoner) analyzeReasoningComplexity(query string) float64 {
	query = strings.ToLower(query)
	
	// Check for complex reasoning patterns
	complexPatterns := []string{
		"analyze", "compare", "evaluate", "synthesize", "deduce", "infer",
		"reasoning", "logic", "proof", "argument", "conclusion", "premise",
		"because", "therefore", "consequently", "implies", "follows",
	}
	
	score := 1.0 // Base score
	for _, pattern := range complexPatterns {
		if strings.Contains(query, pattern) {
			score += 0.5
		}
	}
	
	// Cap at 4.0 (VeryHigh level)
	if score > 4.0 {
		score = 4.0
	}
	
	return score
}

// analyzeKnowledgeComplexity analyzes the knowledge complexity of a task
func (tr *TaskReasoner) analyzeKnowledgeComplexity(query string) float64 {
	query = strings.ToLower(query)
	
	// Check for knowledge-intensive patterns
	knowledgePatterns := []string{
		"explain", "describe", "define", "what is", "how does", "why",
		"history", "theory", "concept", "principle", "law", "rule",
		"research", "study", "academic", "scientific", "technical",
	}
	
	score := 1.0 // Base score
	for _, pattern := range knowledgePatterns {
		if strings.Contains(query, pattern) {
			score += 0.4
		}
	}
	
	// Check for domain-specific knowledge
	domains := []string{
		"medical", "legal", "financial", "engineering", "physics",
		"chemistry", "biology", "mathematics", "computer science",
	}
	
	for _, domain := range domains {
		if strings.Contains(query, domain) {
			score += 0.6
		}
	}
	
	if score > 4.0 {
		score = 4.0
	}
	
	return score
}

// analyzeComputationComplexity analyzes the computational complexity of a task
func (tr *TaskReasoner) analyzeComputationComplexity(query string) float64 {
	query = strings.ToLower(query)
	
	// Check for computation-intensive patterns
	computationPatterns := []string{
		"calculate", "compute", "solve", "optimize", "algorithm",
		"formula", "equation", "mathematical", "statistical", "numerical",
		"process", "transform", "convert", "generate", "create",
	}
	
	score := 1.0 // Base score
	for _, pattern := range computationPatterns {
		if strings.Contains(query, pattern) {
			score += 0.5
		}
	}
	
	// Check for complex computational tasks
	complexTasks := []string{
		"machine learning", "data analysis", "simulation", "modeling",
		"optimization", "algorithm design", "code generation",
	}
	
	for _, task := range complexTasks {
		if strings.Contains(query, task) {
			score += 0.8
		}
	}
	
	if score > 4.0 {
		score = 4.0
	}
	
	return score
}

// analyzeCoordinationComplexity analyzes the coordination complexity of a task
func (tr *TaskReasoner) analyzeCoordinationComplexity(query string) float64 {
	query = strings.ToLower(query)
	
	// Check for coordination-intensive patterns
	coordinationPatterns := []string{
		"plan", "organize", "coordinate", "manage", "schedule",
		"workflow", "process", "steps", "sequence", "order",
		"multiple", "various", "different", "several", "many",
	}
	
	score := 1.0 // Base score
	for _, pattern := range coordinationPatterns {
		if strings.Contains(query, pattern) {
			score += 0.4
		}
	}
	
	// Check for multi-step or multi-component tasks
	multiStepPatterns := []string{
		"first", "then", "next", "finally", "step by step",
		"phase", "stage", "component", "part", "section",
	}
	
	for _, pattern := range multiStepPatterns {
		if strings.Contains(query, pattern) {
			score += 0.6
		}
	}
	
	if score > 4.0 {
		score = 4.0
	}
	
	return score
}

// GetComplexityLevel converts a numeric score to a complexity level
func (tr *TaskReasoner) GetComplexityLevel(score float64) ComplexityLevel {
	return FloatToComplexityLevel(score)
}

// AnalyzeConstraints analyzes any constraints in the input
func (tr *TaskReasoner) AnalyzeConstraints(input RequestInput) map[string]interface{} {
	constraints := make(map[string]interface{})
	
	// Check if constraints are provided in input
	if input.Constraints != nil {
		for key, value := range input.Constraints {
			constraints[key] = value
		}
	}
	
	// Analyze query for implicit constraints
	query := strings.ToLower(input.Query)
	
	// Time constraints
	timeConstraints := []string{"urgent", "asap", "quickly", "immediate", "fast"}
	for _, constraint := range timeConstraints {
		if strings.Contains(query, constraint) {
			constraints["time_sensitive"] = true
			break
		}
	}
	
	// Quality constraints
	qualityConstraints := []string{"detailed", "comprehensive", "thorough", "accurate", "precise"}
	for _, constraint := range qualityConstraints {
		if strings.Contains(query, constraint) {
			constraints["high_quality"] = true
			break
		}
	}
	
	// Length constraints
	if strings.Contains(query, "brief") || strings.Contains(query, "short") || strings.Contains(query, "summary") {
		constraints["length"] = "short"
	} else if strings.Contains(query, "detailed") || strings.Contains(query, "comprehensive") {
		constraints["length"] = "long"
	}
	
	return constraints
}