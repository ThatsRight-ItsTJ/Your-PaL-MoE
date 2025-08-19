package taskmaster

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type TaskExecutor struct {
	maxConcurrency int
	timeout        time.Duration
}

func NewTaskExecutor() *TaskExecutor {
	return &TaskExecutor{
		maxConcurrency: 5,
		timeout:        60 * time.Second,
	}
}

func (te *TaskExecutor) ExecuteSequential(ctx context.Context, tasks []*Task) error {
	for _, task := range tasks {
		if err := te.executeTask(ctx, task); err != nil {
			return fmt.Errorf("task %s failed: %w", task.ID, err)
		}
	}
	return nil
}

func (te *TaskExecutor) ExecuteParallel(ctx context.Context, tasks []*Task) error {
	// Build dependency graph
	dependencyMap := make(map[string][]string)
	taskMap := make(map[string]*Task)

	for _, task := range tasks {
		taskMap[task.ID] = task
		dependencyMap[task.ID] = task.Dependencies
	}

	// Execute tasks in dependency order with parallelism
	return te.executeWithDependencies(ctx, taskMap, dependencyMap)
}

func (te *TaskExecutor) executeWithDependencies(ctx context.Context, taskMap map[string]*Task, dependencies map[string][]string) error {
	completed := make(map[string]bool)
	running := make(map[string]bool)
	var wg sync.WaitGroup
	errorChan := make(chan error, len(taskMap))
	semaphore := make(chan struct{}, te.maxConcurrency)

	for len(completed) < len(taskMap) {
		// Find tasks that can be executed (dependencies completed)
		var readyTasks []*Task
		for taskID, task := range taskMap {
			if completed[taskID] || running[taskID] {
				continue
			}

			canExecute := true
			for _, depID := range dependencies[taskID] {
				if !completed[depID] {
					canExecute = false
					break
				}
			}

			if canExecute {
				readyTasks = append(readyTasks, task)
			}
		}

		if len(readyTasks) == 0 && len(running) == 0 {
			return fmt.Errorf("circular dependency detected or all remaining tasks failed")
		}

		// Execute ready tasks
		for _, task := range readyTasks {
			wg.Add(1)
			running[task.ID] = true

			go func(t *Task) {
				defer wg.Done()

				// Acquire semaphore
				semaphore <- struct{}{}
				defer func() { <-semaphore }()

				if err := te.executeTask(ctx, t); err != nil {
					errorChan <- fmt.Errorf("task %s failed: %w", t.ID, err)
					return
				}

				completed[t.ID] = true
				delete(running, t.ID)
			}(task)
		}

		// Wait for at least one task to complete
		if len(readyTasks) > 0 {
			// Wait a bit to allow tasks to start
			time.Sleep(100 * time.Millisecond)
		}

		// Check for errors
		select {
		case err := <-errorChan:
			return err
		default:
		}
	}

	wg.Wait()

	// Check for any remaining errors
	close(errorChan)
	for err := range errorChan {
		if err != nil {
			return err
		}
	}

	return nil
}

func (te *TaskExecutor) executeTask(ctx context.Context, task *Task) error {
	task.Status = "running"
	task.StartTime = time.Now()

	// Create task-specific context with timeout
	taskCtx, cancel := context.WithTimeout(ctx, te.timeout)
	defer cancel()

	// Execute based on task type
	var result interface{}
	var err error

	switch task.Type {
	case "text_generation":
		result, err = te.executeTextGeneration(taskCtx, task)
	case "image_generation":
		result, err = te.executeImageGeneration(taskCtx, task)
	case "audio_generation":
		result, err = te.executeAudioGeneration(taskCtx, task)
	case "embeddings":
		result, err = te.executeEmbeddings(taskCtx, task)
	default:
		err = fmt.Errorf("unsupported task type: %s", task.Type)
	}

	task.EndTime = time.Now()

	if err != nil {
		task.Status = "failed"
		task.Error = err.Error()
		return err
	}

	task.Status = "completed"
	task.Result = result
	return nil
}

func (te *TaskExecutor) executeTextGeneration(ctx context.Context, task *Task) (interface{}, error) {
	// This is a placeholder implementation
	// In the real system, this would call the provider's API

	// Simulate processing time based on tier
	var delay time.Duration
	switch task.Tier {
	case "official":
		delay = 2 * time.Second // Premium APIs are typically faster
	case "community":
		delay = 5 * time.Second // Community APIs might be slower
	case "unofficial":
		delay = 8 * time.Second // Unofficial APIs might be slowest
	default:
		delay = 3 * time.Second
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(delay):
	}

	// Generate mock response based on provider
	response := fmt.Sprintf("Response from %s (%s tier) for prompt: %s", 
		task.Provider, task.Tier, truncateString(task.Prompt, 50))

	return response, nil
}

func (te *TaskExecutor) executeImageGeneration(ctx context.Context, task *Task) (interface{}, error) {
	// Simulate image generation
	delay := 10 * time.Second // Image generation typically takes longer

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(delay):
	}

	response := map[string]interface{}{
		"image_url": fmt.Sprintf("/images/generatedimage.jpg", task.ID),
		"provider":  task.Provider,
		"prompt":    task.Prompt,
	}

	return response, nil
}

func (te *TaskExecutor) executeAudioGeneration(ctx context.Context, task *Task) (interface{}, error) {
	// Simulate audio generation
	delay := 15 * time.Second

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(delay):
	}

	response := map[string]interface{}{
		"audio_url": fmt.Sprintf("https://generated-audio.example.com/%s.mp3", task.ID),
		"provider":  task.Provider,
		"prompt":    task.Prompt,
	}

	return response, nil
}

func (te *TaskExecutor) executeEmbeddings(ctx context.Context, task *Task) (interface{}, error) {
	// Simulate embeddings generation
	delay := 1 * time.Second // Embeddings are typically fast

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(delay):
	}

	// Generate mock embedding vector
	embedding := make([]float64, 384) // Common embedding dimension
	for i := range embedding {
		embedding[i] = 0.1 // Mock values
	}

	response := map[string]interface{}{
		"embedding": embedding,
		"provider":  task.Provider,
		"input":     task.Prompt,
	}

	return response, nil
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func (te *TaskExecutor) SetMaxConcurrency(max int) {
	te.maxConcurrency = max
}

func (te *TaskExecutor) SetTimeout(timeout time.Duration) {
	te.timeout = timeout
}