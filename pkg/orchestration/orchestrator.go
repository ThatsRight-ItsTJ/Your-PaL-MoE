package orchestration

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Your-PaL-MoE/pkg/analysis"
	"github.com/Your-PaL-MoE/pkg/modes"
	"github.com/Your-PaL-MoE/pkg/optimization"
	"github.com/Your-PaL-MoE/pkg/selection"
)

// Task represents a unit of work in the orchestration system
type Task struct {
	ID           string                 `json:"id"`
	Type         string                 `json:"type"`
	Description  string                 `json:"description"`
	Status       TaskStatus             `json:"status"`
	Mode         string                 `json:"mode"`
	Input        map[string]interface{} `json:"input"`
	Output       map[string]interface{} `json:"output"`
	CreatedAt    time.Time              `json:"created_at"`
	CompletedAt  *time.Time             `json:"completed_at,omitempty"`
	Dependencies []string               `json:"dependencies"`
	Error        string                 `json:"error,omitempty"`
}

// TaskStatus represents the current state of a task
type TaskStatus string

const (
	TaskPending   TaskStatus = "pending"
	TaskRunning   TaskStatus = "running"
	TaskCompleted TaskStatus = "completed"
	TaskFailed    TaskStatus = "failed"
)

// Orchestrator manages multi-agent task coordination
type Orchestrator struct {
	tasks       map[string]*Task
	taskQueue   chan *Task
	modeManager *modes.ModeManager
	selector    *selection.AdaptiveSelector
	analyzer    *analysis.ComplexityAnalyzer
	optimizer   *optimization.SPOOptimizer
	workers     []*Worker
	mutex       sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
}

// Worker processes tasks from the orchestrator
type Worker struct {
	ID           string
	orchestrator *Orchestrator
	running      bool
}

// NewOrchestrator creates a new task orchestrator
func NewOrchestrator(modeManager *modes.ModeManager, selector *selection.AdaptiveSelector) *Orchestrator {
	ctx, cancel := context.WithCancel(context.Background())
	
	o := &Orchestrator{
		tasks:       make(map[string]*Task),
		taskQueue:   make(chan *Task, 100),
		modeManager: modeManager,
		selector:    selector,
		analyzer:    analysis.NewComplexityAnalyzer(),
		optimizer:   optimization.NewSPOOptimizer(),
		workers:     make([]*Worker, 0),
		ctx:         ctx,
		cancel:      cancel,
	}

	// Start workers
	for i := 0; i < 5; i++ {
		worker := &Worker{
			ID:           fmt.Sprintf("worker-%d", i),
			orchestrator: o,
		}
		o.workers = append(o.workers, worker)
		go worker.run()
	}

	return o
}

// CreateTask creates a new task and adds it to the orchestration queue
func (o *Orchestrator) CreateTask(taskType, description, mode string, input map[string]interface{}, dependencies []string) (*Task, error) {
	task := &Task{
		ID:           generateTaskID(),
		Type:         taskType,
		Description:  description,
		Status:       TaskPending,
		Mode:         mode,
		Input:        input,
		CreatedAt:    time.Now(),
		Dependencies: dependencies,
	}

	o.mutex.Lock()
	o.tasks[task.ID] = task
	o.mutex.Unlock()

	// Check if dependencies are met
	if o.dependenciesMet(task) {
		select {
		case o.taskQueue <- task:
			task.Status = TaskRunning
		default:
			// Queue is full, task remains pending
		}
	}

	return task, nil
}

// GetTask retrieves a task by ID
func (o *Orchestrator) GetTask(taskID string) (*Task, error) {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	task, exists := o.tasks[taskID]
	if !exists {
		return nil, fmt.Errorf("task not found: %s", taskID)
	}

	return task, nil
}

// ListTasks returns all tasks with optional status filter
func (o *Orchestrator) ListTasks(status TaskStatus) []*Task {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	var tasks []*Task
	for _, task := range o.tasks {
		if status == "" || task.Status == status {
			tasks = append(tasks, task)
		}
	}

	return tasks
}

// dependenciesMet checks if all task dependencies are completed
func (o *Orchestrator) dependenciesMet(task *Task) bool {
	for _, depID := range task.Dependencies {
		o.mutex.RLock()
		depTask, exists := o.tasks[depID]
		o.mutex.RUnlock()

		if !exists || depTask.Status != TaskCompleted {
			return false
		}
	}
	return true
}

// Shutdown gracefully shuts down the orchestrator
func (o *Orchestrator) Shutdown() {
	o.cancel()
	
	// Wait for workers to finish
	for _, worker := range o.workers {
		worker.running = false
	}
	
	close(o.taskQueue)
}

// Worker implementation

// run starts the worker's main processing loop
func (w *Worker) run() {
	w.running = true
	for w.running {
		select {
		case task := <-w.orchestrator.taskQueue:
			if task != nil {
				w.processTask(task)
			}
		case <-w.orchestrator.ctx.Done():
			w.running = false
			return
		case <-time.After(time.Second):
			// Check for tasks with met dependencies
			w.checkPendingTasks()
		}
	}
}

// processTask processes a single task
func (w *Worker) processTask(task *Task) {
	defer func() {
		if r := recover(); r != nil {
			task.Status = TaskFailed
			task.Error = fmt.Sprintf("Task panicked: %v", r)
		}
	}()

	// Get mode configuration
	mode, err := w.orchestrator.modeManager.GetMode(task.Mode)
	if err != nil {
		task.Status = TaskFailed
		task.Error = fmt.Sprintf("Mode not found: %v", err)
		return
	}

	// Analyze task complexity
	complexity := w.orchestrator.analyzer.AnalyzeTask(task.Description, task.Input)

	// Select provider
	providerScore, err := w.orchestrator.selector.SelectProvider(complexity, task.Input)
	if err != nil {
		task.Status = TaskFailed
		task.Error = fmt.Sprintf("Provider selection failed: %v", err)
		return
	}

	// Optimize prompt if needed
	optimized := w.orchestrator.optimizer.OptimizePrompt(task.Description, task.Input)

	// Execute task with selected provider
	result, err := w.executeWithProvider(providerScore, mode, optimized, complexity)
	if err != nil {
		task.Status = TaskFailed
		task.Error = fmt.Sprintf("Execution failed: %v", err)
		return
	}

	// Store results
	task.Output = result
	task.Status = TaskCompleted
	now := time.Now()
	task.CompletedAt = &now

	// Trigger dependent tasks
	w.triggerDependents(task.ID)
}

// executeWithProvider executes the task using the selected provider
func (w *Worker) executeWithProvider(provider selection.ProviderScore, mode modes.ModeConfig, prompt optimization.OptimizationResult, complexity analysis.TaskComplexity) (map[string]interface{}, error) {
	// This is where the actual provider execution would happen
	// For now, we'll simulate the execution
	
	result := map[string]interface{}{
		"provider_id":     provider.ProviderID,
		"mode":           mode.Name,
		"complexity":     complexity,
		"optimization":   prompt,
		"reasoning":      provider.Reasoning,
		"execution_time": time.Now().Format(time.RFC3339),
		"status":         "completed",
	}

	// Simulate processing time based on complexity
	processingTime := time.Duration(complexity.Score * 1000) * time.Millisecond
	time.Sleep(processingTime)

	return result, nil
}

// checkPendingTasks checks for pending tasks whose dependencies are now met
func (w *Worker) checkPendingTasks() {
	w.orchestrator.mutex.RLock()
	allTasks := make([]*Task, 0)
	for _, task := range w.orchestrator.tasks {
		allTasks = append(allTasks, task)
	}
	w.orchestrator.mutex.RUnlock()

	for _, task := range allTasks {
		if task.Status == TaskPending && w.orchestrator.dependenciesMet(task) {
			select {
			case w.orchestrator.taskQueue <- task:
				task.Status = TaskRunning
			default:
				// Queue is full, task remains pending
			}
		}
	}
}

// triggerDependents triggers tasks that depend on the completed task
func (w *Worker) triggerDependents(taskID string) {
	w.orchestrator.mutex.RLock()
	allTasks := make([]*Task, 0)
	for _, task := range w.orchestrator.tasks {
		allTasks = append(allTasks, task)
	}
	w.orchestrator.mutex.RUnlock()

	for _, task := range allTasks {
		if task.Status == TaskPending {
			for _, dep := range task.Dependencies {
				if dep == taskID && w.orchestrator.dependenciesMet(task) {
					select {
					case w.orchestrator.taskQueue <- task:
						task.Status = TaskRunning
					default:
						// Queue is full, task remains pending
					}
					break
				}
			}
		}
	}
}

// generateTaskID generates a unique task ID
func generateTaskID() string {
	return fmt.Sprintf("task_%d", time.Now().UnixNano())
}

// GetStats returns orchestrator statistics
func (o *Orchestrator) GetStats() map[string]interface{} {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	stats := map[string]interface{}{
		"total_tasks":     len(o.tasks),
		"pending_tasks":   0,
		"running_tasks":   0,
		"completed_tasks": 0,
		"failed_tasks":    0,
		"active_workers":  len(o.workers),
		"queue_size":      len(o.taskQueue),
	}

	for _, task := range o.tasks {
		switch task.Status {
		case TaskPending:
			stats["pending_tasks"] = stats["pending_tasks"].(int) + 1
		case TaskRunning:
			stats["running_tasks"] = stats["running_tasks"].(int) + 1
		case TaskCompleted:
			stats["completed_tasks"] = stats["completed_tasks"].(int) + 1
		case TaskFailed:
			stats["failed_tasks"] = stats["failed_tasks"].(int) + 1
		}
	}

	return stats
}