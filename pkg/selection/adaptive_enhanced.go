package selection

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/ThatsRight-ItsTJ/Your-PaL-MoE/pkg/analysis"
	"github.com/ThatsRight-ItsTJ/Your-PaL-MoE/pkg/config"
	"github.com/ThatsRight-ItsTJ/Your-PaL-MoE/pkg/providers"
)

// EnhancedAdaptiveSelector implements intelligent provider selection with capability filtering
type EnhancedAdaptiveSelector struct {
	csvProviders       []config.CSVProvider
	providerCapabilities map[string]ProviderCapabilities
	enhancedConfigs    map[string]*config.ProviderConfig
	performanceData    map[string]*ProviderMetrics
	weights            SelectionWeights
	mutex              sync.RWMutex
	yamlBuilder        *config.YAMLBuilder
	capabilityDetector *CapabilityDetector
	csvParser          *providers.CSVParser
}

// NewEnhancedAdaptiveSelector creates a new enhanced adaptive provider selector
func NewEnhancedAdaptiveSelector(csvPath string) (*EnhancedAdaptiveSelector, error) {
	yamlBuilder := config.NewYAMLBuilder()
	yamlBuilder.SetCSVPath(csvPath)
	yamlBuilder.SetConfigDir("./configs")
	
	csvParser := providers.NewCSVParser(csvPath)
	capabilityDetector := NewCapabilityDetector()
	
	selector := &EnhancedAdaptiveSelector{
		enhancedConfigs:      make(map[string]*config.ProviderConfig),
		providerCapabilities: make(map[string]ProviderCapabilities),
		performanceData:      make(map[string]*ProviderMetrics),
		yamlBuilder:          yamlBuilder,
		csvParser:            csvParser,
		capabilityDetector:   capabilityDetector,
		weights: SelectionWeights{
			Cost:        0.25,
			Quality:     0.40,
			Latency:     0.20,
			Reliability: 0.15,
		},
	}

	// Load CSV providers
	providers, err := yamlBuilder.ReadCSV()
	if err != nil {
		return nil, fmt.Errorf("failed to load CSV providers: %w", err)
	}
	selector.csvProviders = providers

	// Load and analyze provider capabilities
	if err := selector.loadProviderCapabilities(); err != nil {
		return nil, fmt.Errorf("failed to load provider capabilities: %w", err)
	}

	// Try to load enhanced configurations
	if err := yamlBuilder.BuildFromCSV(); err == nil {
		selector.loadEnhancedConfigs("./configs")
	}

	return selector, nil
}

// loadProviderCapabilities analyzes each provider's models to determine capabilities
func (eas *EnhancedAdaptiveSelector) loadProviderCapabilities() error {
	providerConfigs, err := eas.csvParser.LoadProviders()
	if err != nil {
		return fmt.Errorf("failed to load provider configs: %w", err)
	}

	for name, config := range providerConfigs {
		var models []string
		
		// Extract models based on source type
		switch config.ModelsSource.Type {
		case "list":
			if modelList, ok := config.ModelsSource.Value.([]string); ok {
				models = modelList
			}
		case "endpoint":
			// For endpoint-based models, we'll use the provider name as a hint
			models = []string{name}
		case "script":
			// For script-based models, we'll use the provider name as a hint
			models = []string{name}
		}

		// Detect capabilities from models
		capabilities := eas.capabilityDetector.DetectCapabilities(models)
		providerID := eas.generateProviderID(name)
		eas.providerCapabilities[providerID] = capabilities
	}

	return nil
}

// SelectProvider selects the best provider for a given task complexity with capability filtering
func (eas *EnhancedAdaptiveSelector) SelectProvider(complexity analysis.TaskComplexity, constraints map[string]interface{}) (ProviderScore, error) {
	eas.mutex.RLock()
	defer eas.mutex.RUnlock()

	// Detect task type from complexity context or constraints
	taskType := eas.detectTaskTypeFromContext(complexity, constraints)

	// Filter providers by capability first
	compatibleProviders := eas.filterCompatibleProviders(taskType)
	if len(compatibleProviders) == 0 {
		return ProviderScore{}, fmt.Errorf("no providers found with capability for task type: %s", taskType)
	}

	// Try enhanced selection first if available
	if len(eas.enhancedConfigs) > 0 {
		return eas.selectFromEnhancedFiltered(complexity, constraints, compatibleProviders, taskType)
	}

	// Fallback to CSV-based selection with filtering
	return eas.selectFromCSVFiltered(complexity, constraints, compatibleProviders, taskType)
}

// filterCompatibleProviders returns only providers that can handle the task type
func (eas *EnhancedAdaptiveSelector) filterCompatibleProviders(taskType TaskType) []string {
	var compatible []string
	
	for providerID, capabilities := range eas.providerCapabilities {
		if eas.capabilityDetector.IsProviderCompatible(capabilities, taskType) {
			compatible = append(compatible, providerID)
		}
	}
	
	return compatible
}

// detectTaskTypeFromContext determines task type from complexity and constraints
func (eas *EnhancedAdaptiveSelector) detectTaskTypeFromContext(complexity analysis.TaskComplexity, constraints map[string]interface{}) TaskType {
	// Check constraints for explicit task type
	if taskTypeStr, exists := constraints["task_type"]; exists {
		if taskType, ok := taskTypeStr.(string); ok {
			switch strings.ToLower(taskType) {
			case "image":
				return TaskTypeImage
			case "code":
				return TaskTypeCode
			case "audio":
				return TaskTypeAudio
			case "video":
				return TaskTypeVideo
			case "multimodal":
				return TaskTypeMultimodal
			}
		}
	}

	// Check for content in constraints to detect task type
	if content, exists := constraints["content"]; exists {
		if contentStr, ok := content.(string); ok {
			return eas.capabilityDetector.DetectTaskType(contentStr)
		}
	}

	// Default to text for general tasks
	return TaskTypeText
}

// selectFromEnhancedFiltered performs selection using enhanced configurations with filtering
func (eas *EnhancedAdaptiveSelector) selectFromEnhancedFiltered(complexity analysis.TaskComplexity, constraints map[string]interface{}, compatibleProviders []string, taskType TaskType) (ProviderScore, error) {
	var candidates []ProviderScore

	for _, provider := range eas.enhancedConfigs {
		providerID := eas.generateProviderID(provider.Name)
		
		// Only consider compatible providers
		if !eas.contains(compatibleProviders, providerID) {
			continue
		}
		
		score := eas.calculateProviderScore(*provider, complexity, constraints)
		
		// Boost score based on capability match
		if capabilities, exists := eas.providerCapabilities[providerID]; exists {
			score = eas.adjustScoreForCapabilities(score, capabilities, taskType)
		}
		
		candidates = append(candidates, score)
	}

	if len(candidates) == 0 {
		return ProviderScore{}, fmt.Errorf("no suitable enhanced providers found for task type: %s", taskType)
	}

	// Sort by total score (descending)
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].TotalScore > candidates[j].TotalScore
	})

	best := candidates[0]
	best.Reasoning = eas.generateEnhancedSelectionReasoning(best, complexity, constraints, taskType)

	return best, nil
}

// selectFromCSVFiltered performs fallback selection using only CSV data with filtering
func (eas *EnhancedAdaptiveSelector) selectFromCSVFiltered(complexity analysis.TaskComplexity, constraints map[string]interface{}, compatibleProviders []string, taskType TaskType) (ProviderScore, error) {
	var candidates []ProviderScore

	for _, provider := range eas.csvProviders {
		providerID := eas.generateProviderID(provider.Name)
		
		// Only consider compatible providers
		if !eas.contains(compatibleProviders, providerID) {
			continue
		}
		
		score := eas.calculateCSVProviderScore(provider, complexity, constraints)
		
		// Boost score based on capability match
		if capabilities, exists := eas.providerCapabilities[providerID]; exists {
			score = eas.adjustScoreForCapabilities(score, capabilities, taskType)
		}
		
		candidates = append(candidates, score)
	}

	if len(candidates) == 0 {
		return ProviderScore{}, fmt.Errorf("no suitable CSV providers found for task type: %s", taskType)
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].TotalScore > candidates[j].TotalScore
	})

	best := candidates[0]
	best.Reasoning = eas.generateCSVSelectionReasoningWithCapabilities(best, complexity, taskType)

	return best, nil
}

// adjustScoreForCapabilities boosts scores based on capability-task alignment
func (eas *EnhancedAdaptiveSelector) adjustScoreForCapabilities(score ProviderScore, capabilities ProviderCapabilities, taskType TaskType) ProviderScore {
	boost := 1.0
	
	// Apply capability-specific boosts
	switch taskType {
	case TaskTypeText:
		if capabilities.Text {
			boost = 1.1 // 10% boost for text capability
		}
	case TaskTypeImage:
		if capabilities.Image {
			boost = 1.2 // 20% boost for image capability (more specialized)
		}
	case TaskTypeCode:
		if capabilities.Code {
			boost = 1.15 // 15% boost for code capability
		} else if capabilities.Text {
			boost = 1.05 // Small boost if text model can handle code
		}
	case TaskTypeAudio:
		if capabilities.Audio {
			boost = 1.25 // 25% boost for audio capability (highly specialized)
		}
	case TaskTypeVideo:
		if capabilities.Video {
			boost = 1.3 // 30% boost for video capability (very specialized)
		}
	case TaskTypeMultimodal:
		if capabilities.Multimodal {
			boost = 1.2 // 20% boost for multimodal capability
		} else if capabilities.Text && capabilities.Image {
			boost = 1.1 // Smaller boost if has both text and image
		}
	}
	
	// Apply the boost
	score.TotalScore *= boost
	score.QualityScore *= boost
	
	return score
}

// Helper methods from original AdaptiveSelector (reused with modifications)

func (eas *EnhancedAdaptiveSelector) calculateProviderScore(provider config.ProviderConfig, complexity analysis.TaskComplexity, constraints map[string]interface{}) ProviderScore {
	qualityScore := eas.calculateQualityScore(provider, complexity)
	costScore := eas.calculateCostScore(provider, constraints)
	latencyScore := eas.calculateLatencyScore(provider.ID)
	reliabilityScore := eas.calculateReliabilityScore(provider.ID)

	totalScore := (qualityScore * eas.weights.Quality) +
		(costScore * eas.weights.Cost) +
		(latencyScore * eas.weights.Latency) +
		(reliabilityScore * eas.weights.Reliability)

	return ProviderScore{
		ProviderID:       provider.ID,
		TotalScore:       totalScore,
		CostScore:        costScore,
		QualityScore:     qualityScore,
		LatencyScore:     latencyScore,
		ReliabilityScore: reliabilityScore,
	}
}

func (eas *EnhancedAdaptiveSelector) calculateCSVProviderScore(provider config.CSVProvider, complexity analysis.TaskComplexity, constraints map[string]interface{}) ProviderScore {
	// Simple tier-based quality scoring
	qualityScore := eas.getTierQualityScore(provider.Tier)

	// Adjust based on complexity requirements
	if complexity.Score > 0.7 && provider.Tier != "official" {
		qualityScore *= 0.7 // Penalty for non-official providers on complex tasks
	}

	// Simple cost scoring based on tier
	costScore := eas.getTierCostScore(provider.Tier)

	// Default latency and reliability scores
	latencyScore := 0.7
	reliabilityScore := 0.8

	// Apply historical data if available
	providerID := eas.generateProviderID(provider.Name)
	if metrics, exists := eas.performanceData[providerID]; exists {
		latencyScore = eas.calculateLatencyScoreFromMetrics(metrics)
		reliabilityScore = metrics.SuccessRate
	}

	totalScore := (qualityScore * eas.weights.Quality) +
		(costScore * eas.weights.Cost) +
		(latencyScore * eas.weights.Latency) +
		(reliabilityScore * eas.weights.Reliability)

	return ProviderScore{
		ProviderID:       providerID,
		TotalScore:       totalScore,
		CostScore:        costScore,
		QualityScore:     qualityScore,
		LatencyScore:     latencyScore,
		ReliabilityScore: reliabilityScore,
	}
}

// Reuse helper methods from original AdaptiveSelector
func (eas *EnhancedAdaptiveSelector) calculateQualityScore(provider config.ProviderConfig, complexity analysis.TaskComplexity) float64 {
	// Use capability-based quality scoring if available
	providerID := eas.generateProviderID(provider.Name)
	if capabilities, exists := eas.providerCapabilities[providerID]; exists {
		return eas.calculateCapabilityBasedQuality(capabilities, complexity)
	}
	
	// Fallback to default scoring
	return 0.7
}

func (eas *EnhancedAdaptiveSelector) calculateCapabilityBasedQuality(capabilities ProviderCapabilities, complexity analysis.TaskComplexity) float64 {
	// Match provider capabilities to task requirements
	reasoningMatch := math.Min(float64(capabilities.Reasoning)/10.0, complexity.Reasoning/3.0)
	knowledgeMatch := math.Min(float64(capabilities.Knowledge)/10.0, complexity.Knowledge/3.0)
	computationMatch := math.Min(float64(capabilities.Computation)/10.0, complexity.Computation/3.0)

	score := (reasoningMatch + knowledgeMatch + computationMatch) / 3.0
	return math.Min(score, 1.0)
}

func (eas *EnhancedAdaptiveSelector) calculateCostScore(provider config.ProviderConfig, constraints map[string]interface{}) float64 {
	// Simple cost scoring - would be enhanced with real cost data
	return 0.7
}

func (eas *EnhancedAdaptiveSelector) calculateLatencyScore(providerID string) float64 {
	metrics, exists := eas.performanceData[providerID]
	if !exists {
		return 0.7 // Default score for unknown providers
	}
	return eas.calculateLatencyScoreFromMetrics(metrics)
}

func (eas *EnhancedAdaptiveSelector) calculateLatencyScoreFromMetrics(metrics *ProviderMetrics) float64 {
	targetLatency := 500 * time.Millisecond
	if metrics.AverageLatency <= targetLatency {
		return 1.0
	}

	maxLatency := 5 * time.Second
	if metrics.AverageLatency >= maxLatency {
		return 0.1
	}

	ratio := float64(metrics.AverageLatency-targetLatency) / float64(maxLatency-targetLatency)
	return 1.0 - (ratio * 0.9)
}

func (eas *EnhancedAdaptiveSelector) calculateReliabilityScore(providerID string) float64 {
	metrics, exists := eas.performanceData[providerID]
	if !exists {
		return 0.8 // Default high reliability for untested providers
	}
	return metrics.SuccessRate
}

func (eas *EnhancedAdaptiveSelector) getTierQualityScore(tier string) float64 {
	switch tier {
	case "official":
		return 0.9
	case "community":
		return 0.7
	case "unofficial":
		return 0.5
	default:
		return 0.6
	}
}

func (eas *EnhancedAdaptiveSelector) getTierCostScore(tier string) float64 {
	switch tier {
	case "official":
		return 0.3 // Higher cost
	case "community":
		return 0.7 // Medium cost
	case "unofficial":
		return 1.0 // Lower/free cost
	default:
		return 0.5
	}
}

func (eas *EnhancedAdaptiveSelector) generateEnhancedSelectionReasoning(score ProviderScore, complexity analysis.TaskComplexity, constraints map[string]interface{}, taskType TaskType) string {
	reasons := []string{}

	if score.QualityScore > 0.8 {
		reasons = append(reasons, "high capability match for task complexity")
	}
	if score.CostScore > 0.7 {
		reasons = append(reasons, "cost-effective option")
	}
	if score.LatencyScore > 0.8 {
		reasons = append(reasons, "fast response times")
	}
	if score.ReliabilityScore > 0.9 {
		reasons = append(reasons, "excellent reliability")
	}

	// Add capability-specific reasoning
	if capabilities, exists := eas.providerCapabilities[score.ProviderID]; exists {
		switch taskType {
		case TaskTypeImage:
			if capabilities.Image {
				reasons = append(reasons, "specialized image generation capability")
			}
		case TaskTypeCode:
			if capabilities.Code {
				reasons = append(reasons, "specialized code generation capability")
			}
		case TaskTypeAudio:
			if capabilities.Audio {
				reasons = append(reasons, "specialized audio processing capability")
			}
		}
	}

	complexityDesc := complexity.GetComplexityDescription()
	reasoning := fmt.Sprintf("Selected for %s %s task. Key factors: %s",
		complexityDesc, taskType, eas.joinReasons(reasons))

	return reasoning
}

func (eas *EnhancedAdaptiveSelector) generateCSVSelectionReasoningWithCapabilities(score ProviderScore, complexity analysis.TaskComplexity, taskType TaskType) string {
	complexityDesc := complexity.GetComplexityDescription()
	return fmt.Sprintf("Selected based on %s %s task requirements and provider capability filtering (total score: %.2f)",
		complexityDesc, taskType, score.TotalScore)
}

// UpdateProviderMetrics updates performance metrics for a provider
func (eas *EnhancedAdaptiveSelector) UpdateProviderMetrics(providerID string, latency time.Duration, success bool, quality float64) {
	eas.mutex.Lock()
	defer eas.mutex.Unlock()

	metrics, exists := eas.performanceData[providerID]
	if !exists {
		metrics = &ProviderMetrics{
			AverageLatency:     latency,
			SuccessRate:        1.0,
			QualityScore:       quality,
			TotalRequests:      0,
			SuccessfulRequests: 0,
			LastUpdated:        time.Now(),
		}
		eas.performanceData[providerID] = metrics
	}

	// Update metrics using exponential moving average
	alpha := 0.1 // Learning rate
	metrics.AverageLatency = time.Duration(float64(metrics.AverageLatency)*(1-alpha) + float64(latency)*alpha)

	metrics.TotalRequests++
	if success {
		metrics.SuccessfulRequests++
	}

	metrics.SuccessRate = float64(metrics.SuccessfulRequests) / float64(metrics.TotalRequests)
	metrics.QualityScore = metrics.QualityScore*(1-alpha) + quality*alpha
	metrics.LastUpdated = time.Now()
}

// GetProviderMetrics returns current metrics for all providers
func (eas *EnhancedAdaptiveSelector) GetProviderMetrics() map[string]*ProviderMetrics {
	eas.mutex.RLock()
	defer eas.mutex.RUnlock()

	result := make(map[string]*ProviderMetrics)
	for k, v := range eas.performanceData {
		result[k] = v
	}
	return result
}

// GetProviderCapabilities returns capabilities for all providers
func (eas *EnhancedAdaptiveSelector) GetProviderCapabilities() map[string]ProviderCapabilities {
	eas.mutex.RLock()
	defer eas.mutex.RUnlock()

	result := make(map[string]ProviderCapabilities)
	for k, v := range eas.providerCapabilities {
		result[k] = v
	}
	return result
}

// SetWeights updates the selection weights
func (eas *EnhancedAdaptiveSelector) SetWeights(weights SelectionWeights) {
	eas.mutex.Lock()
	defer eas.mutex.Unlock()
	eas.weights = weights
}

// Utility methods
func (eas *EnhancedAdaptiveSelector) contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func (eas *EnhancedAdaptiveSelector) generateProviderID(name string) string {
	return strings.ToLower(strings.ReplaceAll(name, " ", "_"))
}

func (eas *EnhancedAdaptiveSelector) joinReasons(reasons []string) string {
	if len(reasons) == 0 {
		return "general suitability"
	}
	if len(reasons) == 1 {
		return reasons[0]
	}
	if len(reasons) == 2 {
		return reasons[0] + " and " + reasons[1]
	}

	last := reasons[len(reasons)-1]
	others := strings.Join(reasons[:len(reasons)-1], ", ")
	return others + ", and " + last
}

func (eas *EnhancedAdaptiveSelector) loadEnhancedConfigs(configDir string) error {
	// Implementation would load YAML files from configDir
	return nil
}