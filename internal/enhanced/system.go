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

// RequestInput represents an incoming request
type RequestInput struct {
	ID          string                 `json:"id"`
	Content     string                 `json:"content"`
	Context     map[string]interface{} `json:"context"`
	Constraints map[string]interface{} `json:"constraints"`
	Mode        string                 `json:"mode,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
}

// RequestResult represents the result of processing a request
type RequestResult struct {
	ID               string                        `json:"id"`
	Status           string                        `json:"status"`
	Complexity       analysis.TaskComplexity       `json:"complexity"`
	OptimizedPrompt  optimization.OptimizationResult `json:"optimized_prompt"`
	Assignment       selection.ProviderScore       `json:"assignment"`
	Response         map[string]interface{}        `json:"response,omitempty"`
	TotalCost        float64                       `json:"total_cost"`
	TotalDuration    string                        `json:"total_duration"`
	Error            string                        `json:"error,omitempty"`
	CreatedAt        time.Time                     `json:"created_at"`
	CompletedAt      *time.Time                    `json:"completed_at,omitempty"`
}

// EnhancedSystem integrates all components of the Kilocode-inspired improvements
type EnhancedSystem struct {
	logger        *logrus.Logger
	analyzer      *analysis.ComplexityAnalyzer
	optimizer     *optimization.SPOOptimizer
	selector      *selection.AdaptiveSelector
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

// SystemMetrics tracks system-wide performance
type SystemMetrics struct {
	TotalRequests      int64   `json:"total_requests"`
	SuccessfulRequests int64   `json:"successful_requests"`
	FailedRequests     int64   `json:"failed_requests"`
	AverageResponseTime time.Duration `json:"average_response_time"`
	AverageCostSavings float64 `json:"average_cost_savings"`
	SystemUptime       time.Time `json:"system_uptime"`
	ActiveRequests     int64   `json:"active_requests"`
}

// NewEnhancedSystem creates a new enhanced Your-PaL-MoE system
func NewEnhancedSystem(logger *logrus.Logger, providersFile string) (*EnhancedSystem, error) {
	// Initialize components
	analyzer := analysis.NewComplexityAnalyzer()
	optimizer := optimization.NewSPOOptimizer()
	
	selector, err := selection.NewAdaptiveSelector(providersFile)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize selector: %w", err)
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
	yamlBuilder := config.NewYAMLBuilder(providersFile, "./configs")
	
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
	
	logger.Info("Enhanced Your-PaL-MoE system initialized successfully")
	return system, nil
}

// ProcessRequest processes an incoming request through the enhanced pipeline
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
	result.Complexity = complexity
	
	// Step 2: Self-Supervised Prompt Optimization
	es.logger.Infof("Optimizing prompt for request %s", input.ID)
	optimized := es.optimizer.OptimizePrompt(input.Content, input.Context)
	result.OptimizedPrompt = optimized
	
	// Step 3: Adaptive Provider Selection
	es.logger.Infof("Selecting provider for request %s", input.ID)
	providerScore, err := es.selector.SelectProvider(complexity, input.Constraints)
	if err != nil {
		result.Status = "failed"
		result.Error = fmt.Sprintf("Provider selection failed: %v", err)
		es.updateMetrics(func(m *SystemMetrics) { m.FailedRequests++ })
		return result, err
	}
	result.Assignment = providerScore
	
	// Step 4: Task Execution (simulated for now)
	es.logger.Infof("Executing task for request %s with provider %s", input.ID, providerScore.ProviderID)
	
	// Create orchestration task
	taskInput := map[string]interface{}{
		"original_prompt": input.Content,
		"optimized_prompt": optimized.Optimized,
		"complexity": complexity,
		"provider": providerScore,
		"context": input.Context,
		"constraints": input.Constraints,
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
					m.AverageResponseTime = time.Duration((int64(m.AverageResponseTime) + int64(duration)) / 2)
					m.AverageCostSavings = (m.AverageCostSavings + optimized.CostSavings) / 2
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

// GetProviders returns information about all providers
func (es *EnhancedSystem) GetProviders() map[string]interface{} {
	metrics := es.selector.GetProviderMetrics()
	modes := es.modeManager.ListModes()
	
	return map[string]interface{}{
		"provider_metrics": metrics,
		"available_modes":  modes,
		"total_providers":  len(metrics),
	}
}

// GetMetrics returns system performance metrics
func (es *EnhancedSystem) GetMetrics() SystemMetrics {
	es.metricsMutex.RLock()
	defer es.metricsMutex.RUnlock()
	
	metrics := es.metrics
	metrics.SystemUptime = time.Since(metrics.SystemUptime)
	return metrics
}

// GenerateProviderYAML generates YAML configuration for a specific provider
func (es *EnhancedSystem) GenerateProviderYAML(ctx context.Context, providerID string) (string, error) {
	// This would generate YAML for a specific provider
	// For now, return a placeholder
	return fmt.Sprintf("# YAML configuration for provider: %s\n# Generated at: %s\n", providerID, time.Now().Format(time.RFC3339)), nil
}

// GenerateAllProviderYAMLs generates YAML configurations for all providers
func (es *EnhancedSystem) GenerateAllProviderYAMLs(ctx context.Context) (map[string]string, error) {
	if err := es.yamlBuilder.BuildFromCSV(); err != nil {
		return nil, fmt.Errorf("failed to build YAML configurations: %w", err)
	}
	
	// Return placeholder for now
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
	
	// Apply provider cost factor (this would come from actual provider config)
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