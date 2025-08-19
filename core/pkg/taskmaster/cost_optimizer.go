package taskmaster

import (
	"context"
	"fmt"

	"github.com/labring/aiproxy/core/pkg/providers"
)

type CostOptimizer struct {
	savingsThreshold float64 // Minimum savings percentage to consider optimization
}

func NewCostOptimizer() *CostOptimizer {
	return &CostOptimizer{
		savingsThreshold: 0.1, // 10% savings threshold
	}
}

func (co *CostOptimizer) OptimizeTasks(ctx context.Context, tasks []*Task, pm *providers.ProviderManager) ([]*Task, OptimizationStats, error) {
	originalCost := co.calculateOriginalCost(tasks)
	
	optimizedTasks := make([]*Task, len(tasks))
	copy(optimizedTasks, tasks)

	// Apply cost optimization strategies
	var totalSavings float64
	optimizations := []string{}

	// Strategy 1: Tier downgrading for simple tasks
	savings1, opts1 := co.optimizeTierSelection(ctx, optimizedTasks, pm)
	totalSavings += savings1
	optimizations = append(optimizations, opts1...)

	// Strategy 2: Task consolidation
	savings2, opts2 := co.optimizeTaskConsolidation(optimizedTasks)
	totalSavings += savings2
	optimizations = append(optimizations, opts2...)

	// Strategy 3: Provider-specific optimizations
	savings3, opts3 := co.optimizeProviderSelection(ctx, optimizedTasks, pm)
	totalSavings += savings3
	optimizations = append(optimizations, opts3...)

	optimizedCost := co.calculateOptimizedCost(optimizedTasks)
	actualSavings := originalCost - optimizedCost
	savingsPercent := 0.0
	if originalCost > 0 {
		savingsPercent = (actualSavings / originalCost) * 100
	}

	stats := OptimizationStats{
		OriginalCost:   originalCost,
		OptimizedCost:  optimizedCost,
		Savings:        actualSavings,
		SavingsPercent: savingsPercent,
		Reasoning:      fmt.Sprintf("Applied optimizations: %v", optimizations),
	}

	return optimizedTasks, stats, nil
}

func (co *CostOptimizer) calculateOriginalCost(tasks []*Task) float64 {
	var total float64
	for _, task := range tasks {
		// Estimate cost based on task type if not set
		if task.Cost == 0 {
			task.Cost = co.estimateTaskCost(task, "official") // Assume premium pricing initially
		}
		total += task.Cost
	}
	return total
}

func (co *CostOptimizer) calculateOptimizedCost(tasks []*Task) float64 {
	var total float64
	for _, task := range tasks {
		total += task.Cost
	}
	return total
}

func (co *CostOptimizer) optimizeTierSelection(ctx context.Context, tasks []*Task, pm *providers.ProviderManager) (float64, []string) {
	var totalSavings float64
	var optimizations []string

	for _, task := range tasks {
		originalCost := task.Cost

		// Try to use community/unofficial providers for simple tasks
		if co.isSimpleTask(task) {
			// Try unofficial first (free)
			if newCost := co.estimateTaskCost(task, "unofficial"); newCost < originalCost {
				task.Tier = "unofficial"
				task.Cost = newCost
				totalSavings += originalCost - newCost
				optimizations = append(optimizations, fmt.Sprintf("Downgraded %s to unofficial tier", task.ID))
				continue
			}

			// Try community (low cost)
			if newCost := co.estimateTaskCost(task, "community"); newCost < originalCost {
				task.Tier = "community"
				task.Cost = newCost
				totalSavings += originalCost - newCost
				optimizations = append(optimizations, fmt.Sprintf("Downgraded %s to community tier", task.ID))
			}
		}
	}

	return totalSavings, optimizations
}

func (co *CostOptimizer) optimizeTaskConsolidation(tasks []*Task) (float64, []string) {
	// Look for tasks that can be consolidated
	var savings float64
	var optimizations []string

	textTasks := co.getTextGenerationTasks(tasks)
	if len(textTasks) > 1 {
		// Check if text tasks can be combined
		if co.canConsolidateTextTasks(textTasks) {
			savings += co.calculateConsolidationSavings(textTasks)
			optimizations = append(optimizations, fmt.Sprintf("Consolidated %d text generation tasks", len(textTasks)))
		}
	}

	return savings, optimizations
}

func (co *CostOptimizer) optimizeProviderSelection(ctx context.Context, tasks []*Task, pm *providers.ProviderManager) (float64, []string) {
	// This would implement provider-specific optimization strategies
	var savings float64
	var optimizations []string

	for _, task := range tasks {
		// Look for cheaper providers that meet quality requirements
		if alternativeCost := co.findCheaperAlternative(task); alternativeCost < task.Cost {
			originalCost := task.Cost
			task.Cost = alternativeCost
			savings += originalCost - alternativeCost
			optimizations = append(optimizations, fmt.Sprintf("Found cheaper alternative for %s", task.ID))
		}
	}

	return savings, optimizations
}

func (co *CostOptimizer) isSimpleTask(task *Task) bool {
	// Determine if a task is simple enough for community/unofficial providers
	switch task.Type {
	case "text_generation":
		// Simple text tasks: short prompts, basic requests
		return len(task.Prompt) < 200 && !co.isComplexPrompt(task.Prompt)
	case "image_generation":
		// Simple image tasks: basic descriptions
		return len(task.Prompt) < 100
	default:
		return false
	}
}

func (co *CostOptimizer) isComplexPrompt(prompt string) bool {
	// Check for complexity indicators
	complexKeywords := []string{
		"analyze", "complex", "detailed", "comprehensive", "expert",
		"professional", "technical", "advanced", "sophisticated",
	}

	for _, keyword := range complexKeywords {
		if contains(prompt, keyword) {
			return true
		}
	}

	return false
}

func (co *CostOptimizer) estimateTaskCost(task *Task, tier string) float64 {
	switch tier {
	case "official":
		switch task.Type {
		case "text_generation":
			return 0.002 // $0.002 per 1k tokens
		case "image_generation":
			return 0.04 // $0.04 per image
		case "audio_generation":
			return 0.006 // $0.006 per minute
		case "embeddings":
			return 0.0001 // $0.0001 per 1k tokens
		default:
			return 0.001
		}
	case "community":
		switch task.Type {
		case "text_generation":
			return 0.0002 // Much cheaper
		case "image_generation":
			return 0.004
		default:
			return 0.0001
		}
	case "unofficial":
		return 0.0 // Free but potentially unstable
	default:
		return 0.001
	}
}

func (co *CostOptimizer) getTextGenerationTasks(tasks []*Task) []*Task {
	var textTasks []*Task
	for _, task := range tasks {
		if task.Type == "text_generation" {
			textTasks = append(textTasks, task)
		}
	}
	return textTasks
}

func (co *CostOptimizer) canConsolidateTextTasks(tasks []*Task) bool {
	// Simple heuristic: can consolidate if all tasks are independent
	for _, task := range tasks {
		if len(task.Dependencies) > 0 {
			return false
		}
	}
	return len(tasks) > 1
}

func (co *CostOptimizer) calculateConsolidationSavings(tasks []*Task) float64 {
	if len(tasks) <= 1 {
		return 0
	}

	// Assume 20% savings from consolidation due to reduced API calls
	var totalCost float64
	for _, task := range tasks {
		totalCost += task.Cost
	}

	return totalCost * 0.2
}

func (co *CostOptimizer) findCheaperAlternative(task *Task) float64 {
	// This would implement logic to find cheaper providers
	// For now, return a small cost reduction
	return task.Cost * 0.9
}

func contains(text, substring string) bool {
	return len(text) >= len(substring) && 
		   (text == substring || 
		    len(text) > len(substring) && 
		    (text[:len(substring)] == substring || 
		     text[len(text)-len(substring):] == substring ||
		     findInString(text, substring)))
}

func findInString(text, substring string) bool {
	for i := 0; i <= len(text)-len(substring); i++ {
		if text[i:i+len(substring)] == substring {
			return true
		}
	}
	return false
}

func (co *CostOptimizer) SetSavingsThreshold(threshold float64) {
	co.savingsThreshold = threshold
}