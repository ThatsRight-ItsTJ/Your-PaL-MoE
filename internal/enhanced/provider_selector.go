package enhanced

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// ProviderTier represents the tier of a provider
type ProviderTier string

const (
	OfficialTier   ProviderTier = "official"
	CommunityTier  ProviderTier = "community"
	UnofficialTier ProviderTier = "unofficial"
)

// Provider represents an AI provider with its capabilities
type Provider struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Tier         ProviderTier           `json:"tier"`
	Endpoint     string                 `json:"endpoint"`
	APIKey       string                 `json:"api_key"`
	Models       []string               `json:"models"`        // Support multiple models
	ModelsSource string                 `json:"models_source"` // Either "list" or API endpoint URL
	CostPerToken float64                `json:"cost_per_token"`
	MaxTokens    int                    `json:"max_tokens"`
	Capabilities []string               `json:"capabilities"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	Metrics      ProviderMetrics        `json:"metrics"`
}

// ProviderMetrics tracks performance metrics for a provider
type ProviderMetrics struct {
	SuccessRate      float64   `json:"success_rate"`
	AverageLatency   float64   `json:"average_latency"`
	QualityScore     float64   `json:"quality_score"`
	CostEfficiency   float64   `json:"cost_efficiency"`
	LastUpdated      time.Time `json:"last_updated"`
	RequestCount     int64     `json:"request_count"`
	ErrorCount       int64     `json:"error_count"`
	AverageCost      float64   `json:"average_cost"`
	ReliabilityScore float64   `json:"reliability_score"`
}

// ProviderAssignment represents the assignment of tasks to providers
type ProviderAssignment struct {
	TaskID        string                 `json:"task_id"`
	ProviderID    string                 `json:"provider_id"`
	ProviderTier  ProviderTier           `json:"provider_tier"`
	Confidence    float64                `json:"confidence"`
	EstimatedCost float64                `json:"estimated_cost"`
	EstimatedTime int                    `json:"estimated_time"`
	Reasoning     string                 `json:"reasoning"`
	Alternatives  []AlternativeProvider  `json:"alternatives,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// AlternativeProvider represents an alternative provider option
type AlternativeProvider struct {
	ProviderID    string  `json:"provider_id"`
	Confidence    float64 `json:"confidence"`
	EstimatedCost float64 `json:"estimated_cost"`
	Reasoning     string  `json:"reasoning"`
}

// AdaptiveProviderSelector implements intelligent provider selection
type AdaptiveProviderSelector struct {
	logger *logrus.Logger
	
	// Provider management
	providers     []*Provider
	providersMutex sync.RWMutex
	providersFile  string
	
	// Selection configuration
	costWeight        float64
	performanceWeight float64
	latencyWeight     float64
	reliabilityWeight float64
	adaptationRate    float64
	
	// Performance tracking
	selectionHistory map[string][]SelectionRecord
	historyMutex     sync.RWMutex
}

// SelectionRecord tracks provider selection decisions
type SelectionRecord struct {
	TaskComplexity   TaskComplexity `json:"task_complexity"`
	SelectedProvider string         `json:"selected_provider"`
	ActualCost       float64        `json:"actual_cost"`
	ActualLatency    float64        `json:"actual_latency"`
	QualityScore     float64        `json:"quality_score"`
	Success          bool           `json:"success"`
	Timestamp        time.Time      `json:"timestamp"`
}

// NewAdaptiveProviderSelector creates a new adaptive provider selector
func NewAdaptiveProviderSelector(logger *logrus.Logger, providersFile string) (*AdaptiveProviderSelector, error) {
	selector := &AdaptiveProviderSelector{
		logger:            logger,
		providersFile:     providersFile,
		costWeight:        0.4,
		performanceWeight: 0.3,
		latencyWeight:     0.2,
		reliabilityWeight: 0.1,
		adaptationRate:    0.05,
		selectionHistory:  make(map[string][]SelectionRecord),
	}
	
	// Load providers from CSV file
	if err := selector.loadProviders(); err != nil {
		return nil, fmt.Errorf("failed to load providers: %w", err)
	}
	
	return selector, nil
}

// loadProviders loads providers from CSV file
func (a *AdaptiveProviderSelector) loadProviders() error {
	file, err := os.Open(a.providersFile)
	if err != nil {
		return fmt.Errorf("failed to open providers file: %w", err)
	}
	defer file.Close()
	
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read CSV: %w", err)
	}
	
	if len(records) == 0 {
		return fmt.Errorf("empty providers file")
	}
	
	// Skip header row
	for i := 1; i < len(records); i++ {
		provider, err := a.parseProviderRecord(records[i])
		if err != nil {
			a.logger.Warnf("Failed to parse provider record %d: %v", i, err)
			continue
		}
		a.providers = append(a.providers, provider)
	}
	
	a.logger.Infof("Loaded %d providers from %s", len(a.providers), a.providersFile)
	return nil
}

// parseProviderRecord parses a CSV record into a Provider struct
func (a *AdaptiveProviderSelector) parseProviderRecord(record []string) (*Provider, error) {
	if len(record) < 8 {
		return nil, fmt.Errorf("insufficient columns in provider record")
	}
	
	costPerToken, err := strconv.ParseFloat(record[6], 64)
	if err != nil {
		costPerToken = 0.0
	}
	
	maxTokens, err := strconv.Atoi(record[7])
	if err != nil {
		maxTokens = 4096
	}
	
	// Parse models (column 5) - can be either model list or models endpoint URL
	var models []string
	var modelsSource string
	
	modelField := strings.TrimSpace(record[5])
	if modelField != "" {
		// Check if it's a URL (models endpoint)
		if strings.HasPrefix(modelField, "http://") || strings.HasPrefix(modelField, "https://") {
			modelsSource = modelField
			// Models will be fetched from the endpoint later
			models = []string{} // Empty initially, to be populated by API call
		} else {
			// It's a delimited list of models
			modelsSource = "list"
			// Split by | or , delimiter
			if strings.Contains(modelField, "|") {
				models = strings.Split(modelField, "|")
			} else if strings.Contains(modelField, ",") {
				models = strings.Split(modelField, ",")
			} else if strings.Contains(modelField, ";") {
				models = strings.Split(modelField, ";")
			} else {
				// Single model
				models = []string{modelField}
			}
			
			// Clean up model names
			for i, model := range models {
				models[i] = strings.TrimSpace(model)
			}
		}
	}
	
	// Parse capabilities (column 8)
	capabilities := []string{}
	if len(record) > 8 && record[8] != "" {
		capabilities = strings.Split(record[8], ";")
		for i, cap := range capabilities {
			capabilities[i] = strings.TrimSpace(cap)
		}
	}
	
	// Parse additional info (column 9) - new fifth column
	metadata := make(map[string]interface{})
	if len(record) > 9 && record[9] != "" {
		metadata["additional_info"] = record[9]
		// Parse additional info for structured data like rate limits, costs, etc.
		parts := strings.Split(record[9], ",")
		for _, part := range parts {
			if strings.Contains(part, ":") {
				kv := strings.SplitN(part, ":", 2)
				if len(kv) == 2 {
					key := strings.TrimSpace(kv[0])
					value := strings.TrimSpace(kv[1])
					metadata[key] = value
				}
			}
		}
	}
	
	provider := &Provider{
		ID:           record[0],
		Name:         record[1],
		Tier:         ProviderTier(strings.ToLower(record[2])),
		Endpoint:     record[3],
		APIKey:       record[4],
		Models:       models,
		ModelsSource: modelsSource,
		CostPerToken: costPerToken,
		MaxTokens:    maxTokens,
		Capabilities: capabilities,
		Metadata:     metadata,
		Metrics: ProviderMetrics{
			SuccessRate:      0.9, // Default values
			AverageLatency:   1000.0,
			QualityScore:     0.8,
			CostEfficiency:   0.7,
			LastUpdated:      time.Now(),
			RequestCount:     0,
			ErrorCount:       0,
			AverageCost:      costPerToken,
			ReliabilityScore: 0.8,
		},
	}
	
	return provider, nil
}</to_replace>
</Editor.edit_file_by_replace>

<Editor.edit_file_by_replace>
<file_name>/workspace/Your-PaL-MoE/internal/enhanced/provider_selector.go</file_name>
<to_replace>	return providers
}

// startWatcher starts the file system watcher for CSV changes</to_replace>
<new_content>	return providers
}

// FetchModelsFromEndpoint fetches available models from a provider's models endpoint
func (a *AdaptiveProviderSelector) FetchModelsFromEndpoint(ctx context.Context, provider *Provider) error {
	if provider.ModelsSource == "list" || provider.ModelsSource == "" {
		return nil // No need to fetch, models are already in the list
	}
	
	a.logger.Debugf("Fetching models for provider %s from %s", provider.ID, provider.ModelsSource)
	
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequestWithContext(ctx, "GET", provider.ModelsSource, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	// Add authentication if API key is provided
	if provider.APIKey != "" && provider.APIKey != "none" {
		req.Header.Set("Authorization", "Bearer "+provider.APIKey)
	}
	
	resp, err := client.Do(req)
	if err != nil {
		a.logger.Warnf("Failed to fetch models for %s: %v", provider.ID, err)
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		a.logger.Warnf("Models endpoint returned status %d for provider %s", resp.StatusCode, provider.ID)
		return fmt.Errorf("API returned status %d", resp.StatusCode)
	}
	
	// Try to parse OpenAI-format response
	var modelsResponse struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&modelsResponse); err != nil {
		a.logger.Warnf("Failed to parse models response for %s: %v", provider.ID, err)
		return err
	}
	
	// Extract model IDs
	models := make([]string, len(modelsResponse.Data))
	for i, model := range modelsResponse.Data {
		models[i] = model.ID
	}
	
	// Update provider with fetched models
	provider.Models = models
	a.logger.Infof("Fetched %d models for provider %s: %v", len(models), provider.ID, models)
	
	return nil
}

// GetProviderByModel returns a provider that supports the specified model
func (a *AdaptiveProviderSelector) GetProviderByModel(modelName string) *Provider {
	a.providersMutex.RLock()
	defer a.providersMutex.RUnlock()
	
	for _, provider := range a.providers {
		for _, model := range provider.Models {
			if model == modelName {
				return provider
			}
		}
	}
	return nil
}

// ListAllModels returns all available models across all providers
func (a *AdaptiveProviderSelector) ListAllModels() map[string][]string {
	a.providersMutex.RLock()
	defer a.providersMutex.RUnlock()
	
	result := make(map[string][]string)
	for _, provider := range a.providers {
		result[provider.ID] = provider.Models
	}
	return result
}

// startWatcher starts the file system watcher for CSV changes</to_replace>
</Editor.edit_file_by_replace>

<Editor.edit_file_by_replace>
<file_name>/workspace/Your-PaL-MoE/internal/enhanced/provider_selector.go</file_name>
<to_replace>import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
)</to_replace>
<new_content>import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
)</to_replace>
</Editor.edit_file_by_replace>

<Editor.write>
<path>/workspace/Your-PaL-MoE/internal/enhanced/yaml_generator.go</path>
<content>package enhanced

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// YAMLGenerator handles CSV-to-YAML conversion using AI
type YAMLGenerator struct {
	logger      *logrus.Logger
	pollinationsURL string
	httpClient  *http.Client
}

// NewYAMLGenerator creates a new YAML generator
func NewYAMLGenerator(logger *logrus.Logger) *YAMLGenerator {
	return &YAMLGenerator{
		logger:          logger,
		pollinationsURL: "https://text.pollinations.ai/openai",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GenerateYAMLFromProvider converts a Provider to YAML using AI
func (y *YAMLGenerator) GenerateYAMLFromProvider(ctx context.Context, provider *Provider) (string, error) {
	y.logger.Infof("Generating YAML for provider: %s", provider.ID)
	
	prompt := y.buildPrompt(provider)
	
	response, err := y.callPollinationsAPI(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to call Pollinations API: %w", err)
	}
	
	// Extract YAML from response
	yaml := y.extractYAML(response)
	if yaml == "" {
		return "", fmt.Errorf("no valid YAML generated from AI response")
	}
	
	y.logger.Infof("Successfully generated YAML for provider %s", provider.ID)
	return yaml, nil
}

// buildPrompt creates the AI prompt for YAML generation
func (y *YAMLGenerator) buildPrompt(provider *Provider) string {
	var sb strings.Builder
	
	sb.WriteString("Generate a YAML configuration file for an AI provider with the following details:\n\n")
	sb.WriteString(fmt.Sprintf("Provider ID: %s\n", provider.ID))
	sb.WriteString(fmt.Sprintf("Provider Name: %s\n", provider.Name))
	sb.WriteString(fmt.Sprintf("Tier: %s\n", provider.Tier))
	sb.WriteString(fmt.Sprintf("Endpoint: %s\n", provider.Endpoint))
	sb.WriteString(fmt.Sprintf("Model: %s\n", provider.Model))
	sb.WriteString(fmt.Sprintf("Cost Per Token: %.6f\n", provider.CostPerToken))
	sb.WriteString(fmt.Sprintf("Max Tokens: %d\n", provider.MaxTokens))
	
	if len(provider.Capabilities) > 0 {
		sb.WriteString(fmt.Sprintf("Capabilities: %s\n", strings.Join(provider.Capabilities, ", ")))
	}
	
	// Add metadata/additional info
	if additionalInfo, exists := provider.Metadata["additional_info"]; exists {
		sb.WriteString(fmt.Sprintf("Additional Info: %s\n", additionalInfo))
	}
	
	// Add rate limits and other metadata
	for key, value := range provider.Metadata {
		if key != "additional_info" {
			sb.WriteString(fmt.Sprintf("%s: %v\n", strings.Title(strings.ReplaceAll(key, "_", " ")), value))
		}
	}
	
	sb.WriteString("\nGenerate a complete YAML configuration file that includes:\n")
	sb.WriteString("1. Provider configuration section\n")
	sb.WriteString("2. API settings (endpoint, authentication)\n")
	sb.WriteString("3. Model parameters (temperature, max_tokens, etc.)\n")
	sb.WriteString("4. Rate limiting configuration if applicable\n")
	sb.WriteString("5. Cost tracking settings\n")
	sb.WriteString("6. Retry and timeout configurations\n")
	sb.WriteString("7. Any provider-specific settings based on the additional info\n\n")
	sb.WriteString("Format the output as valid YAML. Include comments explaining each section.\n")
	sb.WriteString("Only return the YAML content, no additional text or explanation.")
	
	return sb.String()
}

// callPollinationsAPI calls the Pollinations OpenAI-compatible API
func (y *YAMLGenerator) callPollinationsAPI(ctx context.Context, prompt string) (string, error) {
	requestBody := map[string]interface{}{
		"model": "openai",
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "You are a YAML configuration expert. Generate clean, well-structured YAML configurations for AI provider integrations.",
			},
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"max_tokens":   2000,
		"temperature":  0.3,
		"stream":       false,
	}
	
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", y.pollinationsURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := y.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returned status %d", resp.StatusCode)
	}
	
	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}
	
	// Extract content from OpenAI-format response
	if choices, ok := response["choices"].([]interface{}); ok && len(choices) > 0 {
		if choice, ok := choices[0].(map[string]interface{}); ok {
			if message, ok := choice["message"].(map[string]interface{}); ok {
				if content, ok := message["content"].(string); ok {
					return content, nil
				}
			}
		}
	}
	
	return "", fmt.Errorf("invalid response format from API")
}

// extractYAML extracts YAML content from the AI response
func (y *YAMLGenerator) extractYAML(response string) string {
	// Remove any markdown code blocks
	response = strings.TrimSpace(response)
	
	// Look for YAML content between code blocks
	if strings.Contains(response, "```yaml") {
		parts := strings.Split(response, "```yaml")
		if len(parts) > 1 {
			yamlPart := strings.Split(parts[1], "```")[0]
			return strings.TrimSpace(yamlPart)
		}
	} else if strings.Contains(response, "```") {
		parts := strings.Split(response, "```")
		if len(parts) >= 2 {
			return strings.TrimSpace(parts[1])
		}
	}
	
	// If no code blocks, assume entire response is YAML
	return response
}

// GenerateYAMLBatch generates YAML for multiple providers
func (y *YAMLGenerator) GenerateYAMLBatch(ctx context.Context, providers []*Provider) (map[string]string, error) {
	results := make(map[string]string)
	
	for _, provider := range providers {
		yaml, err := y.GenerateYAMLFromProvider(ctx, provider)
		if err != nil {
			y.logger.Errorf("Failed to generate YAML for provider %s: %v", provider.ID, err)
			continue
		}
		results[provider.ID] = yaml
	}
	
	y.logger.Infof("Generated YAML for %d/%d providers", len(results), len(providers))
	return results, nil
}

// SaveYAMLToFile saves generated YAML to a file
func (y *YAMLGenerator) SaveYAMLToFile(providerID, yaml string) error {
	filename := fmt.Sprintf("configs/%s.yaml", providerID)
	
	// Create configs directory if it doesn't exist
	if err := ensureDir("configs"); err != nil {
		return fmt.Errorf("failed to create configs directory: %w", err)
	}
	
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filename, err)
	}
	defer file.Close()
	
	_, err = file.WriteString(yaml)
	if err != nil {
		return fmt.Errorf("failed to write YAML to file: %w", err)
	}
	
	y.logger.Infof("Saved YAML configuration to %s", filename)
	return nil
}

// Helper function to ensure directory exists
func ensureDir(dirName string) error {
	if _, err := os.Stat(dirName); os.IsNotExist(err) {
		return os.MkdirAll(dirName, 0755)
	}
	return nil
}
</content>
</Editor.write>

<Editor.edit_file_by_replace>
<file_name>/workspace/Your-PaL-MoE/internal/enhanced/enhanced_system.go</file_name>
<to_replace>	// Core components
	reasoning        *TaskReasoningEngine
	spoOptimizer     *SPOOptimizer
	providerSelector *AdaptiveProviderSelector</to_replace>
<new_content>	// Core components
	reasoning        *TaskReasoningEngine
	spoOptimizer     *SPOOptimizer
	providerSelector *AdaptiveProviderSelector
	yamlGenerator    *YAMLGenerator
		Metrics: ProviderMetrics{
			SuccessRate:      0.9, // Default values
			AverageLatency:   1000.0,
			QualityScore:     0.8,
			CostEfficiency:   0.7,
			LastUpdated:      time.Now(),
			RequestCount:     0,
			ErrorCount:       0,
			AverageCost:      costPerToken,
			ReliabilityScore: 0.8,
		},
	}
	
	return provider, nil
}

// SelectOptimalProvider selects the optimal provider for a task
func (a *AdaptiveProviderSelector) SelectOptimalProvider(ctx context.Context, taskID string, complexity TaskComplexity, requirements map[string]interface{}) (ProviderAssignment, error) {
	a.logger.Infof("Selecting optimal provider for task %s (complexity: %.2f)", taskID, complexity.Score)
	
	a.providersMutex.RLock()
	defer a.providersMutex.RUnlock()
	
	if len(a.providers) == 0 {
		return ProviderAssignment{}, fmt.Errorf("no providers available")
	}
	
	// Calculate scores for all providers
	scores := make([]ProviderScore, 0, len(a.providers))
	for _, provider := range a.providers {
		score := a.calculateProviderScore(provider, complexity, requirements)
		scores = append(scores, ProviderScore{
			Provider: provider,
			Score:    score,
		})
	}
	
	// Sort by score (highest first)
	for i := 0; i < len(scores)-1; i++ {
		for j := i + 1; j < len(scores); j++ {
			if scores[i].Score < scores[j].Score {
				scores[i], scores[j] = scores[j], scores[i]
			}
		}
	}
	
	// Select the best provider
	bestProvider := scores[0].Provider
	
	// Create alternatives list
	alternatives := make([]AlternativeProvider, 0, min(3, len(scores)-1))
	for i := 1; i < min(4, len(scores)); i++ {
		alternatives = append(alternatives, AlternativeProvider{
			ProviderID:    scores[i].Provider.ID,
			Confidence:    scores[i].Score,
			EstimatedCost: a.estimateTaskCost(scores[i].Provider, complexity),
			Reasoning:     fmt.Sprintf("Alternative option with score %.2f", scores[i].Score),
		})
	}
	
	assignment := ProviderAssignment{
		TaskID:        taskID,
		ProviderID:    bestProvider.ID,
		ProviderTier:  bestProvider.Tier,
		Confidence:    scores[0].Score,
		EstimatedCost: a.estimateTaskCost(bestProvider, complexity),
		EstimatedTime: a.estimateTaskTime(bestProvider, complexity),
		Reasoning:     a.generateSelectionReasoning(bestProvider, complexity, scores[0].Score),
		Alternatives:  alternatives,
		Metadata:      make(map[string]interface{}),
	}
	
	a.logger.Infof("Selected provider %s for task %s with confidence %.2f", 
		bestProvider.ID, taskID, scores[0].Score)
	
	return assignment, nil
}

// ProviderScore represents a provider with its calculated score
type ProviderScore struct {
	Provider *Provider
	Score    float64
}

// calculateProviderScore calculates a score for a provider based on task requirements
func (a *AdaptiveProviderSelector) calculateProviderScore(provider *Provider, complexity TaskComplexity, requirements map[string]interface{}) float64 {
	// Base score from provider tier
	tierScore := a.getTierScore(provider.Tier)
	
	// Cost efficiency score (lower cost is better)
	costScore := 1.0 - (provider.CostPerToken / 0.001) // Normalize against typical max cost
	if costScore < 0 {
		costScore = 0
	}
	
	// Performance score from metrics
	performanceScore := provider.Metrics.QualityScore
	
	// Latency score (lower latency is better)
	latencyScore := 1.0 - (provider.Metrics.AverageLatency / 5000.0) // Normalize against 5s max
	if latencyScore < 0 {
		latencyScore = 0
	}
	
	// Reliability score
	reliabilityScore := provider.Metrics.ReliabilityScore
	
	// Complexity alignment score
	complexityScore := a.getComplexityAlignmentScore(provider, complexity)
	
	// Weighted final score
	finalScore := a.costWeight*costScore +
		a.performanceWeight*performanceScore +
		a.latencyWeight*latencyScore +
		a.reliabilityWeight*reliabilityScore +
		0.2*tierScore +
		0.1*complexityScore
	
	return finalScore
}

// getTierScore returns a score based on provider tier
func (a *AdaptiveProviderSelector) getTierScore(tier ProviderTier) float64 {
	switch tier {
	case OfficialTier:
		return 1.0
	case CommunityTier:
		return 0.7
	case UnofficialTier:
		return 0.4
	default:
		return 0.5
	}
}

// getComplexityAlignmentScore returns a score based on how well the provider aligns with task complexity
func (a *AdaptiveProviderSelector) getComplexityAlignmentScore(provider *Provider, complexity TaskComplexity) float64 {
	// Higher complexity tasks should prefer higher-tier providers
	if complexity.Overall >= High && provider.Tier == OfficialTier {
		return 1.0
	} else if complexity.Overall >= Medium && provider.Tier == CommunityTier {
		return 0.8
	} else if complexity.Overall <= Medium && provider.Tier == UnofficialTier {
		return 0.6
	}
	return 0.5
}

// estimateTaskCost estimates the cost for a task with a given provider
func (a *AdaptiveProviderSelector) estimateTaskCost(provider *Provider, complexity TaskComplexity) float64 {
	// Estimate tokens based on complexity
	estimatedTokens := 100.0 // Base tokens
	
	switch complexity.Overall {
	case VeryHigh:
		estimatedTokens = 2000.0
	case High:
		estimatedTokens = 1000.0
	case Medium:
		estimatedTokens = 500.0
	case Low:
		estimatedTokens = 200.0
	}
	
	return provider.CostPerToken * estimatedTokens
}

// estimateTaskTime estimates the time for a task with a given provider
func (a *AdaptiveProviderSelector) estimateTaskTime(provider *Provider, complexity TaskComplexity) int {
	baseTime := int(provider.Metrics.AverageLatency) // Base latency in ms
	
	// Adjust based on complexity
	complexityMultiplier := 1.0
	switch complexity.Overall {
	case VeryHigh:
		complexityMultiplier = 2.0
	case High:
		complexityMultiplier = 1.5
	case Medium:
		complexityMultiplier = 1.2
	case Low:
		complexityMultiplier = 1.0
	}
	
	return int(float64(baseTime) * complexityMultiplier)
}

// generateSelectionReasoning generates reasoning for provider selection
func (a *AdaptiveProviderSelector) generateSelectionReasoning(provider *Provider, complexity TaskComplexity, score float64) string {
	return fmt.Sprintf("Selected %s (tier: %s) based on optimal balance of cost (%.4f per token), "+
		"quality score (%.2f), and complexity alignment for %s complexity task. Overall score: %.2f",
		provider.Name, provider.Tier, provider.CostPerToken, 
		provider.Metrics.QualityScore, complexity.Overall, score)
}

// UpdateProviderMetrics updates provider metrics based on execution results
func (a *AdaptiveProviderSelector) UpdateProviderMetrics(ctx context.Context, providerID string, success bool, cost float64, latency float64, qualityScore float64) error {
	a.providersMutex.Lock()
	defer a.providersMutex.Unlock()
	
	for _, provider := range a.providers {
		if provider.ID == providerID {
			// Update metrics using exponential moving average
			alpha := a.adaptationRate
			
			provider.Metrics.RequestCount++
			if !success {
				provider.Metrics.ErrorCount++
			}
			
			// Update success rate
			currentSuccessRate := float64(provider.Metrics.RequestCount-provider.Metrics.ErrorCount) / float64(provider.Metrics.RequestCount)
			provider.Metrics.SuccessRate = alpha*currentSuccessRate + (1-alpha)*provider.Metrics.SuccessRate
			
			// Update average latency
			provider.Metrics.AverageLatency = alpha*latency + (1-alpha)*provider.Metrics.AverageLatency
			
			// Update quality score
			if qualityScore > 0 {
				provider.Metrics.QualityScore = alpha*qualityScore + (1-alpha)*provider.Metrics.QualityScore
			}
			
			// Update average cost
			provider.Metrics.AverageCost = alpha*cost + (1-alpha)*provider.Metrics.AverageCost
			
			// Update cost efficiency
			provider.Metrics.CostEfficiency = provider.Metrics.QualityScore / (provider.Metrics.AverageCost + 0.001)
			
			// Update reliability score
			provider.Metrics.ReliabilityScore = provider.Metrics.SuccessRate * 0.7 + (1.0-provider.Metrics.AverageLatency/5000.0)*0.3
			
			provider.Metrics.LastUpdated = time.Now()
			
			a.logger.Debugf("Updated metrics for provider %s: success_rate=%.2f, quality=%.2f, cost_efficiency=%.2f",
				providerID, provider.Metrics.SuccessRate, provider.Metrics.QualityScore, provider.Metrics.CostEfficiency)
			
			return nil
		}
	}
	
	return fmt.Errorf("provider not found: %s", providerID)
}

// GetProviders returns all available providers
func (a *AdaptiveProviderSelector) GetProviders() []*Provider {
	a.providersMutex.RLock()
	defer a.providersMutex.RUnlock()
	
	// Return a copy to prevent external modification
	providers := make([]*Provider, len(a.providers))
	copy(providers, a.providers)
	return providers
}

// Helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}