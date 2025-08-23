package enhanced

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// ExecutionStatus represents the status of task execution
type ExecutionStatus string

const (
	StatusPending   ExecutionStatus = "pending"
	StatusRunning   ExecutionStatus = "running"
	StatusCompleted ExecutionStatus = "completed"
	StatusFailed    ExecutionStatus = "failed"
	StatusCancelled ExecutionStatus = "cancelled"
)

// ProcessingRequest represents a request being processed through the system
type ProcessingRequest struct {
	ID              string               `json:"id"`
	Input           RequestInput         `json:"input"`
	Complexity      TaskComplexity       `json:"complexity"`
	OptimizedPrompt OptimizedPrompt      `json:"optimized_prompt"`
	Assignment      ProviderAssignment   `json:"assignment"`
	Status          ExecutionStatus      `json:"status"`
	Result          string               `json:"result"`
	TotalCost       float64              `json:"total_cost"`
	TotalDuration   time.Duration        `json:"total_duration"`
	CreatedAt       time.Time            `json:"created_at"`
	CompletedAt     *time.Time           `json:"completed_at,omitempty"`
	Error           string               `json:"error,omitempty"`
}

// SystemMetrics represents overall system performance metrics
type SystemMetrics struct {
	TotalRequests        int64              `json:"total_requests"`
	SuccessfulRequests   int64              `json:"successful_requests"`
	FailedRequests       int64              `json:"failed_requests"`
	AverageResponseTime  float64            `json:"average_response_time"`
	TotalCost            float64            `json:"total_cost"`
	CostSavings          float64            `json:"cost_savings"`
	ActiveRequests       int                `json:"active_requests"`
	ProviderHealthScores map[string]float64 `json:"provider_health_scores"`
	LastUpdated          time.Time          `json:"last_updated"`
}

// EnhancedSystem represents the main enhanced Your PaL MoE system
type EnhancedSystem struct {
	logger *logrus.Logger
	
	// Core components
	reasoning        *TaskReasoningEngine
	spoOptimizer     *SPOOptimizer
	providerSelector *AdaptiveProviderSelector
	
	// Configuration
	maxParallelTasks int
	taskTimeout      time.Duration
	
	// State management
	activeRequests map[string]*ProcessingRequest
	mutex          sync.RWMutex
	metrics        *SystemMetrics
	
	// Context for cancellation
	ctx    context.Context
	cancel context.CancelFunc
}

// NewEnhancedSystem creates a new enhanced Your PaL MoE system
func NewEnhancedSystem(logger *logrus.Logger, providersFile string) (*EnhancedSystem, error) {
	ctx, cancel := context.WithCancel(context.Background())
	
	system := &EnhancedSystem{
		logger:           logger,
		maxParallelTasks: 10,
		taskTimeout:      5 * time.Minute,
		activeRequests:   make(map[string]*ProcessingRequest),
		ctx:              ctx,
		cancel:           cancel,
		metrics: &SystemMetrics{
			ProviderHealthScores: make(map[string]float64),
			LastUpdated:          time.Now(),
		},
	}
	
	// Initialize components
	if err := system.initializeComponents(providersFile); err != nil {
		return nil, fmt.Errorf("failed to initialize components: %w", err)
	}
	
	// Start background tasks
	go system.metricsCollector()
	
	return system, nil
}

// initializeComponents initializes all system components
func (s *EnhancedSystem) initializeComponents(providersFile string) error {
	var err error
	
	// Task Reasoning Engine
	s.reasoning, err = NewTaskReasoningEngine(s.logger)
	if err != nil {
		return fmt.Errorf("failed to create reasoning engine: %w", err)
	}
	
	// SPO Optimizer
	s.spoOptimizer, err = NewSPOOptimizer(s.logger)
	if err != nil {
		return fmt.Errorf("failed to create SPO optimizer: %w", err)
	}
	
	// Provider Selector
	s.providerSelector, err = NewAdaptiveProviderSelector(s.logger, providersFile)
	if err != nil {
		return fmt.Errorf("failed to create provider selector: %w", err)
	}
	
	return nil
}

// ProcessRequest processes a request through the enhanced pipeline
func (s *EnhancedSystem) ProcessRequest(ctx context.Context, input RequestInput) (*ProcessingRequest, error) {
	s.logger.Infof("Processing request: %s", input.ID)
	
	// Create processing request
	request := &ProcessingRequest{
		ID:        input.ID,
		Input:     input,
		Status:    StatusPending,
		CreatedAt: time.Now(),
	}
	
	// Store active request
	s.mutex.Lock()
	s.activeRequests[input.ID] = request
	s.mutex.Unlock()
	
	// Process through enhanced pipeline
	if err := s.processPipeline(ctx, request); err != nil {
		request.Status = StatusFailed
		request.Error = err.Error()
		s.logger.Errorf("Request %s failed: %v", input.ID, err)
		s.updateMetrics(request, false)
		return request, err
	}
	
	request.Status = StatusCompleted
	completedAt := time.Now()
	request.CompletedAt = &completedAt
	request.TotalDuration = completedAt.Sub(request.CreatedAt)
	
	// Update metrics
	s.updateMetrics(request, true)
	
	s.logger.Infof("Request %s completed in %v", input.ID, request.TotalDuration)
	return request, nil
}

// processPipeline processes a request through the enhanced pipeline
func (s *EnhancedSystem) processPipeline(ctx context.Context, request *ProcessingRequest) error {
	startTime := time.Now()
	
	// Step 1: Task Reasoning and Complexity Analysis
	s.logger.Infof("Step 1: Analyzing task complexity for request %s", request.ID)
	complexity, err := s.reasoning.AnalyzeComplexity(ctx, request.Input)
	if err != nil {
		return fmt.Errorf("complexity analysis failed: %w", err)
	}
	request.Complexity = complexity
	
	// Step 2: SPO Analysis and Prompt Optimization
	s.logger.Infof("Step 2: Optimizing prompt for request %s", request.ID)
	optimizedPrompt, err := s.spoOptimizer.OptimizePrompt(ctx, request.Input.Content, complexity)
	if err != nil {
		return fmt.Errorf("prompt optimization failed: %w", err)
	}
	request.OptimizedPrompt = optimizedPrompt
	
	// Step 3: Extract requirements for provider selection
	requirements, err := s.reasoning.ExtractRequirements(ctx, request.Input)
	if err != nil {
		return fmt.Errorf("requirement extraction failed: %w", err)
	}
	
	// Step 4: Adaptive Provider Selection
	s.logger.Infof("Step 4: Selecting provider for request %s", request.ID)
	assignment, err := s.providerSelector.SelectOptimalProvider(ctx, request.ID, complexity, requirements)
	if err != nil {
		return fmt.Errorf("provider selection failed: %w", err)
	}
	request.Assignment = assignment
	
	// Step 5: Execute task (simulated for now)
	s.logger.Infof("Step 5: Executing task for request %s with provider %s", request.ID, assignment.ProviderID)
	request.Status = StatusRunning
	
	// Simulate task execution
	result, cost, err := s.simulateTaskExecution(ctx, request.OptimizedPrompt.Optimized, assignment)
	if err != nil {
		// Update provider metrics for failure
		s.providerSelector.UpdateProviderMetrics(ctx, assignment.ProviderID, false, 0, 
			float64(time.Since(startTime).Milliseconds()), 0)
		return fmt.Errorf("task execution failed: %w", err)
	}
	
	request.Result = result
	request.TotalCost = cost
	
	// Update provider metrics for success
	qualityScore := s.calculateQualityScore(result, complexity)
	s.providerSelector.UpdateProviderMetrics(ctx, assignment.ProviderID, true, cost, 
		float64(time.Since(startTime).Milliseconds()), qualityScore)
	
	return nil
}

// simulateTaskExecution simulates task execution (replace with actual API calls)
func (s *EnhancedSystem) simulateTaskExecution(ctx context.Context, prompt string, assignment ProviderAssignment) (string, float64, error) {
	// Simulate processing time based on complexity
	processingTime := time.Duration(assignment.EstimatedTime) * time.Millisecond
	
	select {
	case <-ctx.Done():
		return "", 0, ctx.Err()
	case <-time.After(processingTime):
		// Simulate successful execution
		result := fmt.Sprintf("Processed prompt using provider %s: %s", 
			assignment.ProviderID, prompt[:min(len(prompt), 100)])
		return result, assignment.EstimatedCost, nil
	}
}

// calculateQualityScore calculates a quality score for the result
func (s *EnhancedSystem) calculateQualityScore(result string, complexity TaskComplexity) float64 {
	// Simple heuristic based on result length and complexity
	baseScore := 0.7
	
	if len(result) > 50 {
		baseScore += 0.1
	}
	if len(result) > 200 {
		baseScore += 0.1
	}
	
	// Adjust for complexity
	if complexity.Overall >= High {
		baseScore += 0.1
	}
	
	if baseScore > 1.0 {
		baseScore = 1.0
	}
	
	return baseScore
}

// GetRequest returns a processing request by ID
func (s *EnhancedSystem) GetRequest(requestID string) (*ProcessingRequest, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	request, exists := s.activeRequests[requestID]
	if !exists {
		return nil, fmt.Errorf("request not found: %s", requestID)
	}
	
	return request, nil
}

// GetProviders returns all available providers
func (s *EnhancedSystem) GetProviders() []*Provider {
	return s.providerSelector.GetProviders()
}

// GetMetrics returns system metrics
func (s *EnhancedSystem) GetMetrics() *SystemMetrics {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	// Update active requests count
	s.metrics.ActiveRequests = len(s.activeRequests)
	s.metrics.LastUpdated = time.Now()
	
	// Update provider health scores
	providers := s.providerSelector.GetProviders()
	for _, provider := range providers {
		s.metrics.ProviderHealthScores[provider.ID] = s.calculateHealthScore(provider)
	}
	
	return s.metrics
}

// updateMetrics updates system metrics
func (s *EnhancedSystem) updateMetrics(request *ProcessingRequest, success bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	s.metrics.TotalRequests++
	if success {
		s.metrics.SuccessfulRequests++
	} else {
		s.metrics.FailedRequests++
	}
	
	// Update average response time
	if request.CompletedAt != nil {
		duration := request.CompletedAt.Sub(request.CreatedAt).Seconds()
		s.metrics.AverageResponseTime = (s.metrics.AverageResponseTime*float64(s.metrics.TotalRequests-1) + duration) / float64(s.metrics.TotalRequests)
	}
	
	// Update total cost
	s.metrics.TotalCost += request.TotalCost
	
	// Estimate cost savings from optimization
	if request.OptimizedPrompt.CostSavings > 0 {
		s.metrics.CostSavings += request.OptimizedPrompt.CostSavings * request.TotalCost
	}
}

// calculateHealthScore calculates a health score for a provider
func (s *EnhancedSystem) calculateHealthScore(provider *Provider) float64 {
	metrics := provider.Metrics
	
	// Calculate weighted health score
	successWeight := 0.4
	latencyWeight := 0.2
	qualityWeight := 0.3
	reliabilityWeight := 0.1
	
	// Normalize latency score (lower is better)
	latencyScore := 1.0
	if metrics.AverageLatency > 0 {
		latencyScore = 1.0 / (1.0 + metrics.AverageLatency/1000.0)
	}
	
	healthScore := successWeight*metrics.SuccessRate +
		latencyWeight*latencyScore +
		qualityWeight*metrics.QualityScore +
		reliabilityWeight*metrics.ReliabilityScore
	
	return healthScore
}

// metricsCollector runs background metrics collection
func (s *EnhancedSystem) metricsCollector() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.collectMetrics()
		}
	}
}

// collectMetrics collects system metrics
func (s *EnhancedSystem) collectMetrics() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	// Clean up completed requests older than 1 hour
	cutoff := time.Now().Add(-1 * time.Hour)
	for id, request := range s.activeRequests {
		if request.CompletedAt != nil && request.CompletedAt.Before(cutoff) {
			delete(s.activeRequests, id)
		}
	}
	
	// Update provider health scores
	providers := s.providerSelector.GetProviders()
	for _, provider := range providers {
		s.metrics.ProviderHealthScores[provider.ID] = s.calculateHealthScore(provider)
	}
}

// Shutdown gracefully shuts down the system
func (s *EnhancedSystem) Shutdown() error {
	s.logger.Info("Shutting down enhanced system...")
	s.cancel()
	return nil
}

// Helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}