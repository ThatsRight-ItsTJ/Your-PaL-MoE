package taskmaster

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// TaskRequest represents a user request that needs to be processed
type TaskRequest struct {
	ID         string                 `json:"id"`
	UserPrompt string                 `json:"user_prompt"`
	Options    TaskOptions            `json:"options"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// TaskOptions contains configuration for task processing
type TaskOptions struct {
	CostOptimization string        `json:"cost_optimization"` // "aggressive", "balanced", "quality"
	Timeout          time.Duration `json:"timeout"`
	ParallelExecution bool         `json:"parallel_execution"`
	ProviderPreference []string    `json:"provider_preference"`
	MaxCost          float64       `json:"max_cost"`
}

// TaskPlan represents a decomposed task execution plan
type TaskPlan struct {
	ID       string    `json:"id"`
	Tasks    []Task    `json:"tasks"`
	Metadata map[string]interface{} `json:"metadata"`
	CreatedAt time.Time `json:"created_at"`
}

// Task represents an individual task in the execution plan
type Task struct {
	ID       string      `json:"id"`
	Type     string      `json:"type"` // "chat_completion", "image_generation", "analysis", etc.
	Provider string      `json:"provider"`
	Request  interface{} `json:"request"`
	Priority int         `json:"priority"`
	Dependencies []string `json:"dependencies"`
}

// TaskResult represents the result of executing a task plan
type TaskResult struct {
	Success    bool                   `json:"success"`
	Data       interface{}            `json:"data"`
	Error      string                 `json:"error,omitempty"`
	Cost       float64                `json:"cost"`
	Duration   time.Duration          `json:"duration"`
	TaskResults map[string]interface{} `json:"task_results"`
}

// TaskEngine is the main engine for task decomposition and execution
type TaskEngine struct {
	config    *EngineConfig
	providers map[string]interface{}
	mutex     sync.RWMutex
}

// EngineConfig contains configuration for the task engine
type EngineConfig struct {
	MaxParallelTasks    int           `json:"max_parallel_tasks"`
	DefaultTimeout      time.Duration `json:"default_timeout"`
	CostOptimization    bool          `json:"cost_optimization"`
	ProviderHealthCheck bool          `json:"provider_health_check"`
}

// NewTaskEngine creates a new task engine instance
func NewTaskEngine(config *EngineConfig) *TaskEngine {
	if config == nil {
		config = &EngineConfig{
			MaxParallelTasks:    5,
			DefaultTimeout:      60 * time.Second,
			CostOptimization:    true,
			ProviderHealthCheck: true,
		}
	}

	return &TaskEngine{
		config:    config,
		providers: make(map[string]interface{}),
	}
}

// CreateTaskPlan analyzes a request and creates an optimized execution plan
func (e *TaskEngine) CreateTaskPlan(ctx context.Context, request TaskRequest) (*TaskPlan, error) {
	// Analyze the user prompt to determine task types
	taskTypes := e.analyzePrompt(request.UserPrompt)
	
	// Create tasks based on analysis
	tasks := make([]Task, 0)
	for i, taskType := range taskTypes {
		task := Task{
			ID:       fmt.Sprintf("%s_task_%d", request.ID, i+1),
			Type:     taskType,
			Request:  e.createTaskRequest(taskType, request.UserPrompt, request.Options),
			Priority: e.calculatePriority(taskType, request.Options),
		}
		
		// Assign optimal provider based on cost optimization
		provider, err := e.selectOptimalProvider(taskType, request.Options)
		if err != nil {
			return nil, fmt.Errorf("failed to select provider for task %s: %w", task.ID, err)
		}
		task.Provider = provider
		
		tasks = append(tasks, task)
	}
	
	// Create execution plan
	plan := &TaskPlan{
		ID:        request.ID,
		Tasks:     tasks,
		CreatedAt: time.Now(),
		Metadata: map[string]interface{}{
			"cost_optimization": request.Options.CostOptimization,
			"parallel_execution": request.Options.ParallelExecution,
			"original_prompt": request.UserPrompt,
		},
	}
	
	return plan, nil
}

// ExecuteTaskPlan executes a task plan and returns the aggregated result
func (e *TaskEngine) ExecuteTaskPlan(ctx context.Context, plan *TaskPlan) (*TaskResult, error) {
	start := time.Now()
	
	result := &TaskResult{
		TaskResults: make(map[string]interface{}),
	}
	
	// Execute tasks based on plan configuration
	if plan.Metadata["parallel_execution"] == true && len(plan.Tasks) > 1 {
		return e.executeTasksParallel(ctx, plan)
	} else {
		return e.executeTasksSequential(ctx, plan)
	}
}

// executeTasksParallel executes tasks in parallel
func (e *TaskEngine) executeTasksParallel(ctx context.Context, plan *TaskPlan) (*TaskResult, error) {
	start := time.Now()
	
	var wg sync.WaitGroup
	var mutex sync.Mutex
	
	results := make(map[string]interface{})
	errors := make([]error, 0)
	totalCost := 0.0
	
	// Limit concurrent tasks
	semaphore := make(chan struct{}, e.config.MaxParallelTasks)
	
	for _, task := range plan.Tasks {
		wg.Add(1)
		go func(t Task) {
			defer wg.Done()
			
			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			
			// Execute individual task
			taskResult, err := e.executeIndividualTask(ctx, t)
			
			mutex.Lock()
			defer mutex.Unlock()
			
			if err != nil {
				errors = append(errors, err)
			} else {
				results[t.ID] = taskResult
				if cost, ok := taskResult.(map[string]interface{})["cost"].(float64); ok {
					totalCost += cost
				}
			}
		}(task)
	}
	
	wg.Wait()
	
	// Aggregate results
	result := &TaskResult{
		Success:     len(errors) == 0,
		TaskResults: results,
		Cost:        totalCost,
		Duration:    time.Since(start),
	}
	
	if len(errors) > 0 {
		result.Error = fmt.Sprintf("Failed to execute %d tasks", len(errors))
		return result, nil
	}
	
	// Aggregate task outputs into final response
	result.Data = e.aggregateTaskResults(results)
	
	return result, nil
}

// executeTasksSequential executes tasks one by one
func (e *TaskEngine) executeTasksSequential(ctx context.Context, plan *TaskPlan) (*TaskResult, error) {
	start := time.Now()
	
	results := make(map[string]interface{})
	totalCost := 0.0
	
	for _, task := range plan.Tasks {
		taskResult, err := e.executeIndividualTask(ctx, task)
		if err != nil {
			return &TaskResult{
				Success:  false,
				Error:    fmt.Sprintf("Task %s failed: %v", task.ID, err),
				Duration: time.Since(start),
			}, nil
		}
		
		results[task.ID] = taskResult
		if cost, ok := taskResult.(map[string]interface{})["cost"].(float64); ok {
			totalCost += cost
		}
	}
	
	result := &TaskResult{
		Success:     true,
		TaskResults: results,
		Cost:        totalCost,
		Duration:    time.Since(start),
	}
	
	result.Data = e.aggregateTaskResults(results)
	
	return result, nil
}

// executeIndividualTask executes a single task
func (e *TaskEngine) executeIndividualTask(ctx context.Context, task Task) (interface{}, error) {
	// Mock implementation - in real system this would call actual providers
	time.Sleep(10 * time.Millisecond) // Simulate processing time
	
	return map[string]interface{}{
		"success": true,
		"data":    fmt.Sprintf("Mock result for task %s of type %s", task.ID, task.Type),
		"cost":    e.calculateTaskCost(task),
		"provider": task.Provider,
	}, nil
}

// analyzePrompt analyzes the user prompt to determine required task types
func (e *TaskEngine) analyzePrompt(prompt string) []string {
	// Simple keyword-based analysis - in production this would use NLP
	taskTypes := []string{}
	
	// Check for different task indicators
	if containsAny(prompt, []string{"generate", "create", "image", "picture", "draw"}) {
		taskTypes = append(taskTypes, "image_generation")
	}
	
	if containsAny(prompt, []string{"analyze", "explain", "summarize", "describe"}) {
		taskTypes = append(taskTypes, "text_analysis")
	}
	
	if containsAny(prompt, []string{"chat", "conversation", "respond", "answer"}) {
		taskTypes = append(taskTypes, "chat_completion")
	}
	
	// Default to chat completion if no specific task detected
	if len(taskTypes) == 0 {
		taskTypes = append(taskTypes, "chat_completion")
	}
	
	return taskTypes
}

// selectOptimalProvider selects the best provider for a task based on cost optimization
func (e *TaskEngine) selectOptimalProvider(taskType string, options TaskOptions) (string, error) {
	// Cost optimization strategy
	switch options.CostOptimization {
	case "aggressive":
		// Prefer free/unofficial providers
		return e.selectByTierPreference([]string{"unofficial", "community", "official"})
	case "balanced":
		// Balance cost and quality
		return e.selectByTierPreference([]string{"community", "unofficial", "official"})
	case "quality":
		// Prefer premium providers
		return e.selectByTierPreference([]string{"official", "community", "unofficial"})
	default:
		return e.selectByTierPreference([]string{"community", "official", "unofficial"})
	}
}

// selectByTierPreference selects a provider based on tier preference
func (e *TaskEngine) selectByTierPreference(tierPreference []string) (string, error) {
	// Mock provider selection - in real system this would check available providers
	providersByTier := map[string][]string{
		"unofficial": {"pollinations", "local_script"},
		"community":  {"huggingface", "replicate"},
		"official":   {"openai", "anthropic", "google"},
	}
	
	for _, tier := range tierPreference {
		if providers, exists := providersByTier[tier]; exists && len(providers) > 0 {
			return providers[0], nil // Return first available provider in tier
		}
	}
	
	return "", fmt.Errorf("no providers available")
}

// Helper functions
func containsAny(text string, keywords []string) bool {
	text = strings.ToLower(text)
	for _, keyword := range keywords {
		if strings.Contains(text, strings.ToLower(keyword)) {
			return true
		}
	}
	return false
}

func (e *TaskEngine) createTaskRequest(taskType, prompt string, options TaskOptions) interface{} {
	return map[string]interface{}{
		"prompt":    prompt,
		"task_type": taskType,
		"options":   options,
	}
}

func (e *TaskEngine) calculatePriority(taskType string, options TaskOptions) int {
	// Simple priority calculation
	priorities := map[string]int{
		"chat_completion":  1,
		"text_analysis":    2,
		"image_generation": 3,
	}
	
	if priority, exists := priorities[taskType]; exists {
		return priority
	}
	return 5
}

func (e *TaskEngine) calculateTaskCost(task Task) float64 {
	// Mock cost calculation based on provider tier
	costs := map[string]float64{
		"pollinations": 0.0,
		"local_script": 0.0,
		"huggingface":  0.001,
		"replicate":    0.002,
		"openai":       0.01,
		"anthropic":    0.008,
		"google":       0.006,
	}
	
	if cost, exists := costs[task.Provider]; exists {
		return cost
	}
	return 0.005 // Default cost
}

func (e *TaskEngine) aggregateTaskResults(results map[string]interface{}) interface{} {
	// Simple aggregation - combine all task outputs
	if len(results) == 1 {
		for _, result := range results {
			if data, ok := result.(map[string]interface{})["data"]; ok {
				return data
			}
		}
	}
	
	// Multiple tasks - combine outputs
	combined := make([]interface{}, 0)
	for _, result := range results {
		if data, ok := result.(map[string]interface{})["data"]; ok {
			combined = append(combined, data)
		}
	}
	
	return map[string]interface{}{
		"combined_results": combined,
		"task_count":      len(results),
	}
}