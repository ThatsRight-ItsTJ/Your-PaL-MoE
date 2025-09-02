package enhanced

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/Your-PaL-MoE/pkg/analysis"
	"github.com/Your-PaL-MoE/pkg/config"
	"github.com/Your-PaL-MoE/pkg/modes"
	"github.com/Your-PaL-MoE/pkg/optimization"
	"github.com/Your-PaL-MoE/pkg/orchestration"
	"github.com/Your-PaL-MoE/pkg/selection"
	"github.com/sirupsen/logrus"
)

// EnhancedSystem integrates all components with capability-aware provider selection
type EnhancedSystem struct {
	logger        *logrus.Logger
	analyzer      *analysis.ComplexityAnalyzer
	optimizer     *optimization.SPOOptimizer
	selector      *selection.EnhancedAdaptiveSelector // Use enhanced selector
	modeManager   *modes.ModeManager
	orchestrator  *orchestration.Orchestrator
	yamlBuilder   *config.YAMLBuilder
	
	// Request tracking
	requests      map[string]*RequestResult
	requestsMutex sync.RWMutex
	
	// System metrics
	metrics       SystemMetrics
	metricsMutex  sync.RWMutex
	
	// Configuration
	providersFile string
}

// NewEnhancedSystem creates a new enhanced Your-PaL-MoE system with capability filtering
func NewEnhancedSystem(logger *logrus.Logger, providersFile string) (*EnhancedSystem, error) {
	// Initialize components
	analyzer := analysis.NewComplexityAnalyzer()
	optimizer := optimization.NewSPOOptimizer()
	
	// Use the enhanced selector with capability filtering
	selector, err := selection.NewEnhancedAdaptiveSelector(providersFile)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize enhanced selector: %w", err)
	}
	
	// Create agents.csv if it doesn't exist (with default agents)
	agentsFile := "agents.csv"
	if err := createDefaultAgentsCSV(agentsFile); err != nil {
		logger.Warnf("Could not create default agents.csv: %v", err)
	}
	
	modeManager, err := modes.NewModeManager(agentsFile, "./modes_cache")
	if err != nil {
		return nil, fmt.Errorf("failed to initialize mode manager: %w", err)
	}
	
	orchestrator := orchestration.NewOrchestrator(modeManager, selector)
	yamlBuilder := config.NewYAMLBuilder()
	yamlBuilder.SetCSVPath(providersFile)
	yamlBuilder.SetConfigDir("./configs")
	
	system := &EnhancedSystem{
		logger:        logger,
		analyzer:      analyzer,
		optimizer:     optimizer,
		selector:      selector,
		modeManager:   modeManager,
		orchestrator:  orchestrator,
		yamlBuilder:   yamlBuilder,
		requests:      make(map[string]*RequestResult),
		providersFile: providersFile,
		metrics: SystemMetrics{
			SystemUptime: time.Now(),
		},
	}
	
	// Generate enhanced configurations
	if err := yamlBuilder.BuildFromCSV(); err != nil {
		logger.Warnf("Could not generate enhanced YAML configs: %v", err)
	}
	
	logger.Info("Enhanced Your-PaL-MoE system with capability filtering initialized successfully")
	return system, nil
}

// ProcessRequest processes an incoming request through the enhanced pipeline with capability filtering
func (es *EnhancedSystem) ProcessRequest(ctx context.Context, input RequestInput) (*RequestResult, error) {
	startTime := time.Now()
	
	// Create request result
	result := &RequestResult{
		ID:        input.ID,
		Status:    "processing",
		CreatedAt: startTime,
	}
	
	// Store request
	es.requestsMutex.Lock()
	es.requests[input.ID] = result
	es.requestsMutex.Unlock()
	
	// Update metrics
	es.updateMetrics(func(m *SystemMetrics) {
		m.TotalRequests++
		m.ActiveRequests++
	})
	
	defer func() {
		es.updateMetrics(func(m *SystemMetrics) {
			m.ActiveRequests--
		})
	}()
	
	// Step 1: Task Complexity Analysis
	es.logger.Infof("Analyzing complexity for request %s", input.ID)
	complexity := es.analyzer.AnalyzeTask(input.Content, input.Context)
	result.Complexity = TaskComplexity{
		Overall: ComplexityLevel(complexity.Overall),
		Score:   complexity.Score,
	}
	
	// Step 2: Self-Supervised Prompt Optimization
	es.logger.Infof("Optimizing prompt for request %s", input.ID)
	optimized := es.optimizer.OptimizePrompt(input.Content, input.Context)
	result.OptimizedPrompt = OptimizedPrompt{
		Original:    optimized.Original,
		Optimized:   optimized.Optimized,
		CostSavings: optimized.CostSavings,
		Confidence:  optimized.Confidence,
	}
	
	// Step 3: Enhanced Adaptive Provider Selection with Capability Filtering
	es.logger.Infof("Selecting provider with capability filtering for request %s", input.ID)
	
	// Add content to constraints for task type detection
	enhancedConstraints := make(map[string]interface{})
	if input.Constraints != nil {
		for k, v := range input.Constraints {
			enhancedConstraints[k] = v
		}
	}
	enhancedConstraints["content"] = input.Content
	
	providerScore, err := es.selector.SelectProvider(complexity, enhancedConstraints)
	if err != nil {
		result.Status = "failed"
		result.Error = fmt.Sprintf("Provider selection failed: %v", err)
		es.updateMetrics(func(m *SystemMetrics) { m.FailedRequests++ })
		return result, err
	}
	
	result.Assignment = ProviderAssignment{
		ProviderID:    providerScore.ProviderID,
		Confidence:    providerScore.TotalScore,
		EstimatedCost: providerScore.CostScore,
	}
	
	es.logger.Infof("Selected provider %s with %.2f%% confidence for request %s. Reasoning: %s", 
		providerScore.ProviderID, providerScore.TotalScore*100, input.ID, providerScore.Reasoning)
	
	// Step 4: Task Execution (simulated for now)
	es.logger.Infof("Executing task for request %s with provider %s", input.ID, providerScore.ProviderID)
	
	// Create orchestration task
	taskInput := map[string]interface{}{
		"original_prompt": input.Content,
		"optimized_prompt": optimized.Optimized,
		"complexity": complexity,
		"provider": providerScore,
		"context": input.Context,
		"constraints": enhancedConstraints,
	}
	
	mode := input.Mode
	if mode == "" {
		mode = "router" // Default mode
	}
	
	task, err := es.orchestrator.CreateTask("process_request", optimized.Optimized, mode, taskInput, nil)
	if err != nil {
		result.Status = "failed"
		result.Error = fmt.Sprintf("Task creation failed: %v", err)
		es.updateMetrics(func(m *SystemMetrics) { m.FailedRequests++ })
		return result, err
	}
	
	// Wait for task completion (with timeout)
	timeout := time.After(30 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	
	for {
		select {
		case <-timeout:
			result.Status = "timeout"
			result.Error = "Request timed out"
			es.updateMetrics(func(m *SystemMetrics) { m.FailedRequests++ })
			return result, fmt.Errorf("request timeout")
			
		case <-ticker.C:
			currentTask, err := es.orchestrator.GetTask(task.ID)
			if err != nil {
				continue
			}
			
			if currentTask.Status == orchestration.TaskCompleted {
				result.Response = currentTask.Output
				result.Status = "completed"
				
				// Calculate metrics
				duration := time.Since(startTime)
				result.TotalDuration = duration.String()
				result.TotalCost = es.estimateCost(optimized, providerScore)
				
				completedAt := time.Now()
				result.CompletedAt = &completedAt
				
				// Update provider metrics
				es.selector.UpdateProviderMetrics(providerScore.ProviderID, duration, true, 0.8)
				
				// Update system metrics
				es.updateMetrics(func(m *SystemMetrics) {
					m.SuccessfulRequests++
					m.AverageResponseTime = (m.AverageResponseTime + float64(duration.Milliseconds())) / 2
				})
				
				es.logger.Infof("Request %s completed successfully in %v", input.ID, duration)
				return result, nil
				
			} else if currentTask.Status == orchestration.TaskFailed {
				result.Status = "failed"
				result.Error = currentTask.Error
				
				// Update provider metrics for failure
				es.selector.UpdateProviderMetrics(providerScore.ProviderID, time.Since(startTime), false, 0.0)
				
				es.updateMetrics(func(m *SystemMetrics) { m.FailedRequests++ })
				return result, fmt.Errorf("task failed: %s", currentTask.Error)
			}
		}
	}
}

// GetRequest retrieves a request result by ID
func (es *EnhancedSystem) GetRequest(requestID string) (*RequestResult, error) {
	es.requestsMutex.RLock()
	defer es.requestsMutex.RUnlock()
	
	result, exists := es.requests[requestID]
	if !exists {
		return nil, fmt.Errorf("request not found: %s", requestID)
	}
	
	return result, nil
}

// GetProviders returns information about all providers including capabilities
func (es *EnhancedSystem) GetProviders() map[string]interface{} {
	metrics := es.selector.GetProviderMetrics()
	capabilities := es.selector.GetProviderCapabilities()
	modes := es.modeManager.ListModes()
	
	return map[string]interface{}{
		"provider_metrics":     metrics,
		"provider_capabilities": capabilities,
		"available_modes":      modes,
		"total_providers":      len(metrics),
	}
}

// GetMetrics returns system performance metrics
func (es *EnhancedSystem) GetMetrics() SystemMetrics {
	es.metricsMutex.RLock()
	defer es.metricsMutex.RUnlock()
	
	metrics := es.metrics
	metrics.LastUpdated = time.Now()
	return metrics
}

// GenerateProviderYAML generates YAML configuration for a specific provider
func (es *EnhancedSystem) GenerateProviderYAML(ctx context.Context, providerID string) (string, error) {
	return fmt.Sprintf("# YAML configuration for provider: %s\n# Generated at: %s\n", providerID, time.Now().Format(time.RFC3339)), nil
}

// GenerateAllProviderYAMLs generates YAML configurations for all providers
func (es *EnhancedSystem) GenerateAllProviderYAMLs(ctx context.Context) (map[string]string, error) {
	if err := es.yamlBuilder.BuildFromCSV(); err != nil {
		return nil, fmt.Errorf("failed to build YAML configurations: %w", err)
	}
	
	yamls := map[string]string{
		"generated_at": time.Now().Format(time.RFC3339),
		"status": "completed",
	}
	
	return yamls, nil
}

// Shutdown gracefully shuts down the enhanced system
func (es *EnhancedSystem) Shutdown() {
	es.logger.Info("Shutting down Enhanced Your-PaL-MoE system...")
	es.orchestrator.Shutdown()
	es.optimizer.ClearCache()
	es.logger.Info("Enhanced system shutdown completed")
}

// Helper methods

func (es *EnhancedSystem) updateMetrics(updateFunc func(*SystemMetrics)) {
	es.metricsMutex.Lock()
	defer es.metricsMutex.Unlock()
	updateFunc(&es.metrics)
}

func (es *EnhancedSystem) estimateCost(optimized optimization.OptimizationResult, provider selection.ProviderScore) float64 {
	// Simple cost estimation based on prompt length and provider
	baseTokens := float64(len(optimized.Optimized)) / 4.0 // Rough token estimation
	baseCost := baseTokens * 0.00002 // $0.02 per 1k tokens baseline
	
	// Apply provider cost factor
	providerFactor := 1.0
	if provider.CostScore > 0.8 {
		providerFactor = 0.5 // Cheaper provider
	} else if provider.CostScore < 0.3 {
		providerFactor = 2.0 // More expensive provider
	}
	
	return baseCost * providerFactor * (1.0 - optimized.CostSavings)
}

// createDefaultAgentsCSV creates a default agents.csv file if it doesn't exist
func createDefaultAgentsCSV(filename string) error {
	// Check if file already exists
	if _, err := os.Stat(filename); err == nil {
		return nil // File already exists
	}
	
	defaultAgents := `Name,Type,Providers,Specialization,Other
Router,analyzer,all,task-complexity,Default entry point for request analysis
Optimizer,enhancer,gpt-4|claude,prompt-optimization,SPO techniques for cost savings
CodeExpert,specialist,gpt-4|local-codellama,programming,Syntax validation and code generation
DataAnalyst,specialist,gpt-4|claude,data-analysis,Visualization and statistical analysis
Debugger,specialist,all,troubleshooting,Systematic problem diagnosis
Researcher,specialist,gpt-4|claude,research,Information gathering and synthesis
Creative,specialist,gpt-4|claude,creative,Content generation and ideation`
	
	return os.WriteFile(filename, []byte(defaultAgents), 0644)
}