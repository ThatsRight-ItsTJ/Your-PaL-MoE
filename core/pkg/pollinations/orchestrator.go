package pollinations

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

type Orchestrator struct {
	client *Client
}

type TaskDecomposition struct {
	Tasks       []Task `json:"tasks"`
	Reasoning   string `json:"reasoning"`
	Parallelism bool   `json:"parallelism"`
}

type Task struct {
	ID          string            `json:"id"`
	Type        string            `json:"type"` // text_generation, image_generation, etc.
	Description string            `json:"description"`
	Prompt      string            `json:"prompt"`
	Provider    string            `json:"provider"`
	Tier        string            `json:"tier"`
	Parameters  map[string]interface{} `json:"parameters"`
	Dependencies []string         `json:"dependencies"`
}

type ProviderRecommendation struct {
	Provider     string  `json:"provider"`
	Tier         string  `json:"tier"`
	Confidence   float64 `json:"confidence"`
	Reasoning    string  `json:"reasoning"`
	EstimatedCost float64 `json:"estimated_cost"`
}

func NewOrchestrator() *Orchestrator {
	return &Orchestrator{
		client: NewPollinationsClient(),
	}
}

// DecomposeRequest analyzes a user request and breaks it down into optimized tasks
func (o *Orchestrator) DecomposeRequest(ctx context.Context, userRequest string) (*TaskDecomposition, error) {
	prompt := fmt.Sprintf(`
Analyze this user request and decompose it into optimal AI tasks:

User Request: "%s"

Return JSON with task decomposition:
{
  "tasks": [
    {
      "id": "task_1",
      "type": "text_generation|image_generation|audio_generation|embeddings",
      "description": "Human readable description",
      "prompt": "Specific prompt for this task",
      "provider": "suggested_provider",
      "tier": "official|community|unofficial",
      "parameters": {},
      "dependencies": []
    }
  ],
  "reasoning": "Why this decomposition is optimal",
  "parallelism": true
}

Guidelines:
- Break complex requests into smaller, parallel tasks when possible
- Prefer free/community providers for simple tasks
- Use official providers only for complex/critical tasks
- Consider task dependencies
- Optimize for cost while maintaining quality
`, userRequest)

	response, err := o.client.GenerateText(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to decompose request: %w", err)
	}

	var decomposition TaskDecomposition
	if err := json.Unmarshal([]byte(response), &decomposition); err != nil {
		// Create simple fallback decomposition
		decomposition = TaskDecomposition{
			Tasks: []Task{
				{
					ID:          "task_1",
					Type:        "text_generation",
					Description: "Process user request",
					Prompt:      userRequest,
					Provider:    "auto",
					Tier:        "community",
					Parameters:  map[string]interface{}{},
					Dependencies: []string{},
				},
			},
			Reasoning:   "Simple single-task processing",
			Parallelism: false,
		}
	}

	return &decomposition, nil
}

// RecommendProvider suggests the best provider for a given task
func (o *Orchestrator) RecommendProvider(ctx context.Context, task Task, availableProviders []string) (*ProviderRecommendation, error) {
	prompt := fmt.Sprintf(`
Recommend the best AI provider for this task:

Task Type: %s
Task Description: %s
Available Providers: %v

Consider:
- Cost efficiency (prefer free/community over paid when quality is sufficient)
- Task complexity and quality requirements
- Provider reliability and capabilities

Return JSON:
{
  "provider": "provider_name",
  "tier": "official|community|unofficial",
  "confidence": 0.8,
  "reasoning": "Why this provider is optimal",
  "estimated_cost": 0.001
}
`, task.Type, task.Description, availableProviders)

	response, err := o.client.GenerateText(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider recommendation: %w", err)
	}

	var recommendation ProviderRecommendation
	if err := json.Unmarshal([]byte(response), &recommendation); err != nil {
		// Create fallback recommendation
		recommendation = ProviderRecommendation{
			Provider:      selectFallbackProvider(task.Type, availableProviders),
			Tier:          "community",
			Confidence:    0.6,
			Reasoning:     "Fallback selection based on task type",
			EstimatedCost: 0.0,
		}
	}

	return &recommendation, nil
}

func selectFallbackProvider(taskType string, providers []string) string {
	// Simple fallback logic
	for _, provider := range providers {
		providerLower := strings.ToLower(provider)
		switch taskType {
		case "image_generation":
			if strings.Contains(providerLower, "dalle") || strings.Contains(providerLower, "pollinations") {
				return provider
			}
		case "text_generation":
			if strings.Contains(providerLower, "gpt") || strings.Contains(providerLower, "claude") || strings.Contains(providerLower, "pollinations") {
				return provider
			}
		}
	}

	// Return first available provider
	if len(providers) > 0 {
		return providers[0]
	}
	return "unknown"
}

// OptimizeWorkflow analyzes a set of tasks and optimizes their execution
func (o *Orchestrator) OptimizeWorkflow(ctx context.Context, tasks []Task) ([]Task, error) {
	prompt := fmt.Sprintf(`
Optimize this AI workflow for cost and performance:

Tasks: %s

Return optimized JSON array of tasks with:
- Improved provider selection for cost optimization
- Optimal task ordering considering dependencies
- Parallel execution opportunities
- Cost vs quality trade-offs

Focus on maximizing savings while maintaining acceptable quality.
`, formatTasksForPrompt(tasks))

	response, err := o.client.GenerateText(ctx, prompt)
	if err != nil {
		return tasks, nil // Return original tasks if optimization fails
	}

	var optimizedTasks []Task
	if err := json.Unmarshal([]byte(response), &optimizedTasks); err != nil {
		return tasks, nil // Return original tasks if parsing fails
	}

	return optimizedTasks, nil
}

func formatTasksForPrompt(tasks []Task) string {
	data, _ := json.MarshalIndent(tasks, "", "  ")
	return string(data)
}