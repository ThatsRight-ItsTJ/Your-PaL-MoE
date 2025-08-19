package taskmaster

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/labring/aiproxy/core/pkg/pollinations"
	"github.com/labring/aiproxy/core/pkg/providers"
)

type TaskMaster struct {
	orchestrator    *pollinations.Orchestrator
	providerManager *providers.ProviderManager
	executor        *TaskExecutor
	costOptimizer   *CostOptimizer
	mu              sync.RWMutex
}

type Request struct {
	ID               string                 `json:"id"`
	UserPrompt       string                 `json:"user_prompt"`
	Model            string                 `json:"model,omitempty"`
	Parameters       map[string]interface{} `json:"parameters,omitempty"`
	CostOptimization bool                   `json:"cost_optimization"`
	ParallelExecution bool                  `json:"parallel_execution"`
	TierPreference   []string               `json:"tier_preference,omitempty"`
	CostLimit        float64                `json:"cost_limit,omitempty"`
	QualityMin       int                    `json:"quality_min,omitempty"`
}

type Task struct {
	ID           string                 `json:"id"`
	Type         string                 `json:"type"`
	Description  string                 `json:"description"`
	Prompt       string                 `json:"prompt"`
	Provider     string                 `json:"provider"`
	Tier         string                 `json:"tier"`
	Model        string                 `json:"model"`
	Parameters   map[string]interface{} `json:"parameters"`
	Dependencies []string               `json:"dependencies"`
	Status       string                 `json:"status"` // pending, running, completed, failed
	Result       interface{}            `json:"result,omitempty"`
	Error        string                 `json:"error,omitempty"`
	StartTime    time.Time              `json:"start_time"`
	EndTime      time.Time              `json:"end_time"`
	Cost         float64                `json:"cost"`
}

type WorkflowExecution struct {
	RequestID     string            `json:"request_id"`
	Tasks         []*Task           `json:"tasks"`
	Status        string            `json:"status"` // pending, running, completed, failed
	TotalCost     float64           `json:"total_cost"`
	StartTime     time.Time         `json:"start_time"`
	EndTime       time.Time         `json:"end_time"`
	Result        interface{}       `json:"result,omitempty"`
	Optimization  OptimizationStats `json:"optimization"`
}

type OptimizationStats struct {
	OriginalCost   float64 `json:"original_cost"`
	OptimizedCost  float64 `json:"optimized_cost"`
	Savings        float64 `json:"savings"`
	SavingsPercent float64 `json:"savings_percent"`
	Reasoning      string  `json:"reasoning"`
}

func NewTaskMaster(providerManager *providers.ProviderManager) *TaskMaster {
	return &TaskMaster{
		orchestrator:    pollinations.NewOrchestrator(),
		providerManager: providerManager,
		executor:        NewTaskExecutor(),
		costOptimizer:   NewCostOptimizer(),
	}
}

func (tm *TaskMaster) ProcessRequest(ctx context.Context, request Request) (*WorkflowExecution, error) {
	execution := &WorkflowExecution{
		RequestID: request.ID,
		Status:    "pending",
		StartTime: time.Now(),
	}

	// Step 1: Decompose request into tasks
	decomposition, err := tm.orchestrator.DecomposeRequest(ctx, request.UserPrompt)
	if err != nil {
		return nil, fmt.Errorf("failed to decompose request: %w", err)
	}

	// Step 2: Convert to internal task format
	tasks := tm.convertToTasks(decomposition.Tasks, request)
	execution.Tasks = tasks

	// Step 3: Optimize for cost if requested
	if request.CostOptimization {
		optimizedTasks, stats, err := tm.costOptimizer.OptimizeTasks(ctx, tasks, tm.providerManager)
		if err != nil {
			fmt.Printf("Warning: Cost optimization failed: %v\n", err)
		} else {
			execution.Tasks = optimizedTasks
			execution.Optimization = stats
		}
	}

	// Step 4: Route tasks to optimal providers
	for _, task := range execution.Tasks {
		routerRequest := providers.RouterRequest{
			TaskType:       task.Type,
			Prompt:         task.Prompt,
			Model:          task.Model,
			Parameters:     task.Parameters,
			CostLimit:      request.CostLimit,
			QualityMin:     request.QualityMin,
			TierPreference: request.TierPreference,
		}

		routerResponse, err := tm.providerManager.RouteRequest(ctx, routerRequest)
		if err != nil {
			task.Status = "failed"
			task.Error = fmt.Sprintf("Failed to route task: %v", err)
			continue
		}

		task.Provider = routerResponse.Provider
		task.Tier = routerResponse.Tier
		task.Model = routerResponse.Model
		task.Cost = routerResponse.EstimatedCost
	}

	// Step 5: Execute tasks
	execution.Status = "running"
	if request.ParallelExecution && decomposition.Parallelism {
		err = tm.executor.ExecuteParallel(ctx, execution.Tasks)
	} else {
		err = tm.executor.ExecuteSequential(ctx, execution.Tasks)
	}

	if err != nil {
		execution.Status = "failed"
		return execution, fmt.Errorf("execution failed: %w", err)
	}

	// Step 6: Aggregate results
	execution.Result = tm.aggregateResults(execution.Tasks)
	execution.TotalCost = tm.calculateTotalCost(execution.Tasks)
	execution.Status = "completed"
	execution.EndTime = time.Now()

	return execution, nil
}

func (tm *TaskMaster) convertToTasks(orchestratorTasks []pollinations.Task, request Request) []*Task {
	var tasks []*Task

	for i, oTask := range orchestratorTasks {
		task := &Task{
			ID:           fmt.Sprintf("%s_task_%d", request.ID, i+1),
			Type:         oTask.Type,
			Description:  oTask.Description,
			Prompt:       oTask.Prompt,
			Model:        request.Model,
			Parameters:   oTask.Parameters,
			Dependencies: oTask.Dependencies,
			Status:       "pending",
		}

		// Override model if specified in orchestrator task
		if oTask.Provider != "" && oTask.Provider != "auto" {
			task.Provider = oTask.Provider
		}

		tasks = append(tasks, task)
	}

	return tasks
}

func (tm *TaskMaster) aggregateResults(tasks []*Task) interface{} {
	// Simple aggregation - for text tasks, concatenate results
	// For more complex aggregation, this would use AI to intelligently combine results

	if len(tasks) == 1 {
		return tasks[0].Result
	}

	var results []interface{}
	for _, task := range tasks {
		if task.Status == "completed" && task.Result != nil {
			results = append(results, task.Result)
		}
	}

	// If all tasks are text generation, concatenate them
	if tm.allTextGeneration(tasks) {
		var combined string
		for _, result := range results {
			if text, ok := result.(string); ok {
				if combined != "" {
					combined += "\n\n"
				}
				combined += text
			}
		}
		return combined
	}

	return map[string]interface{}{
		"tasks":   results,
		"summary": fmt.Sprintf("Completed %d of %d tasks", len(results), len(tasks)),
	}
}

func (tm *TaskMaster) allTextGeneration(tasks []*Task) bool {
	for _, task := range tasks {
		if task.Type != "text_generation" {
			return false
		}
	}
	return true
}

func (tm *TaskMaster) calculateTotalCost(tasks []*Task) float64 {
	var total float64
	for _, task := range tasks {
		total += task.Cost
	}
	return total
}

func (tm *TaskMaster) GetExecution(executionID string) (*WorkflowExecution, error) {
	// This would typically retrieve from a database
	// For now, return a placeholder
	return nil, fmt.Errorf("execution %s not found", executionID)
}

func (tm *TaskMaster) ListExecutions() ([]*WorkflowExecution, error) {
	// This would typically retrieve from a database
	return []*WorkflowExecution{}, nil
}

func (tm *TaskMaster) EstimateCost(ctx context.Context, request Request) (*CostEstimate, error) {
	// Decompose request
	decomposition, err := tm.orchestrator.DecomposeRequest(ctx, request.UserPrompt)
	if err != nil {
		return nil, fmt.Errorf("failed to decompose request for cost estimation: %w", err)
	}

	tasks := tm.convertToTasks(decomposition.Tasks, request)

	// Get cost estimates for each task
	var totalCost float64
	var estimates []TaskCostEstimate

	for _, task := range tasks {
		routerRequest := providers.RouterRequest{
			TaskType:       task.Type,
			Prompt:         task.Prompt,
			Model:          task.Model,
			TierPreference: request.TierPreference,
		}

		routerResponse, err := tm.providerManager.RouteRequest(ctx, routerRequest)
		if err != nil {
			continue
		}

		estimate := TaskCostEstimate{
			TaskID:        task.ID,
			Provider:      routerResponse.Provider,
			Tier:          routerResponse.Tier,
			EstimatedCost: routerResponse.EstimatedCost,
		}

		estimates = append(estimates, estimate)
		totalCost += routerResponse.EstimatedCost
	}

	return &CostEstimate{
		TotalCost:     totalCost,
		TaskEstimates: estimates,
		Currency:      "USD",
	}, nil
}

type CostEstimate struct {
	TotalCost     float64             `json:"total_cost"`
	TaskEstimates []TaskCostEstimate  `json:"task_estimates"`
	Currency      string              `json:"currency"`
}

type TaskCostEstimate struct {
	TaskID        string  `json:"task_id"`
	Provider      string  `json:"provider"`
	Tier          string  `json:"tier"`
	EstimatedCost float64 `json:"estimated_cost"`
}