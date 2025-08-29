package modes

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// ModeConfig represents an agent mode configuration
type ModeConfig struct {
	Slug               string   `yaml:"slug"`
	Name               string   `yaml:"name"`
	RoleDefinition     string   `yaml:"role_definition"`
	WhenToUse          string   `yaml:"when_to_use"`
	Description        string   `yaml:"description"`
	Capabilities       []string `yaml:"capabilities"`
	Tools              []string `yaml:"tools"`
	CustomInstructions string   `yaml:"custom_instructions"`
	Providers          []string `yaml:"providers"`
}

// CachedMode represents a cached mode with usage statistics
type CachedMode struct {
	Config     ModeConfig
	CreatedAt  time.Time
	UsageCount int
}

// AgentCSV represents an agent definition from CSV
type AgentCSV struct {
	Name           string
	Type           string
	Providers      string
	Specialization string
	Other          string
}

// ModeManager manages AI agent modes
type ModeManager struct {
	modes     map[string]ModeConfig
	cache     map[string]*CachedMode
	csvPath   string
	yamlCache string
	mutex     sync.RWMutex
}

// NewModeManager creates a new mode manager
func NewModeManager(csvPath, yamlCache string) (*ModeManager, error) {
	manager := &ModeManager{
		modes:     make(map[string]ModeConfig),
		cache:     make(map[string]*CachedMode),
		csvPath:   csvPath,
		yamlCache: yamlCache,
	}

	// Load from CSV and generate modes
	if err := manager.loadFromCSV(); err != nil {
		return nil, fmt.Errorf("failed to load modes from CSV: %w", err)
	}

	return manager, nil
}

// loadFromCSV loads agent definitions from CSV and builds mode configurations
func (mm *ModeManager) loadFromCSV() error {
	agents, err := mm.readAgentCSV()
	if err != nil {
		return err
	}

	for _, agent := range agents {
		mode := mm.buildModeConfig(agent)
		mm.modes[mode.Slug] = mode
	}

	// Save as YAML cache for performance
	return mm.saveYAMLCache()
}

// buildModeConfig creates a mode configuration from agent CSV data
func (mm *ModeManager) buildModeConfig(agent AgentCSV) ModeConfig {
	slug := strings.ToLower(strings.ReplaceAll(agent.Name, " ", "-"))

	return ModeConfig{
		Slug:               slug,
		Name:               agent.Name,
		RoleDefinition:     mm.generateRoleDefinition(agent),
		WhenToUse:          mm.generateUsageGuide(agent),
		Description:        agent.Specialization,
		Capabilities:       mm.parseCapabilities(agent.Type, agent.Specialization),
		Tools:              mm.determineTools(agent.Type),
		CustomInstructions: mm.parseCustomInstructions(agent.Other),
		Providers:          mm.parseProviders(agent.Providers),
	}
}

// generateRoleDefinition creates role definition based on agent type and specialization
func (mm *ModeManager) generateRoleDefinition(agent AgentCSV) string {
	templates := map[string]string{
		"analyzer": "You are the Your-PaL-MoE %s, an intelligent task analyzer specializing in %s. Your goal is to analyze incoming requests and make optimal decisions based on task requirements.",
		"enhancer": "You are the Your-PaL-MoE %s, specializing in %s. Your goal is to enhance and optimize user inputs for better results while maintaining efficiency.",
		"specialist": "You are the Your-PaL-MoE %s, a domain expert in %s. Your goal is to provide specialized expertise and high-quality outputs in your area of specialization.",
		"coordinator": "You are the Your-PaL-MoE %s, specializing in %s. Your goal is to coordinate multiple tasks and ensure seamless workflow execution.",
	}

	template, exists := templates[agent.Type]
	if !exists {
		template = "You are the Your-PaL-MoE %s, specializing in %s. Your goal is to assist users effectively in your area of expertise."
	}

	return fmt.Sprintf(template, agent.Name, agent.Specialization)
}

// generateUsageGuide creates usage guidance based on specialization
func (mm *ModeManager) generateUsageGuide(agent AgentCSV) string {
	guides := map[string]string{
		"task-complexity": "Use this mode for initial request analysis and provider selection. Perfect for cost-optimization and intelligent routing decisions.",
		"prompt-optimization": "Use when prompts need enhancement for better performance or cost savings. Ideal for complex tasks that benefit from structured prompts.",
		"programming": "Use for code generation, review, debugging, and technical implementation tasks.",
		"data-analysis": "Use for analyzing datasets, creating visualizations, and generating data-driven insights.",
		"troubleshooting": "Use when debugging issues, investigating errors, or diagnosing problems systematically.",
		"research": "Use for comprehensive research tasks, information gathering, and knowledge synthesis.",
		"creative": "Use for creative writing, content generation, and innovative problem-solving tasks.",
	}

	guide, exists := guides[agent.Specialization]
	if !exists {
		guide = fmt.Sprintf("Use this mode for tasks related to %s.", agent.Specialization)
	}

	return guide
}

// parseCapabilities extracts capabilities from type and specialization
func (mm *ModeManager) parseCapabilities(agentType, specialization string) []string {
	capabilities := []string{}

	// Base capabilities by type
	switch agentType {
	case "analyzer":
		capabilities = append(capabilities, "task_analysis", "complexity_assessment", "provider_selection")
	case "enhancer":
		capabilities = append(capabilities, "prompt_optimization", "quality_improvement", "efficiency_enhancement")
	case "specialist":
		capabilities = append(capabilities, "domain_expertise", "specialized_knowledge", "technical_implementation")
	case "coordinator":
		capabilities = append(capabilities, "workflow_management", "task_coordination", "resource_allocation")
	}

	// Additional capabilities by specialization
	switch specialization {
	case "programming":
		capabilities = append(capabilities, "code_generation", "debugging", "syntax_validation")
	case "data-analysis":
		capabilities = append(capabilities, "data_processing", "visualization", "statistical_analysis")
	case "research":
		capabilities = append(capabilities, "information_gathering", "source_validation", "synthesis")
	case "creative":
		capabilities = append(capabilities, "creative_writing", "ideation", "content_generation")
	}

	return capabilities
}

// determineTools determines available tools based on agent type
func (mm *ModeManager) determineTools(agentType string) []string {
	tools := map[string][]string{
		"analyzer": {"complexity_analyzer", "pattern_matcher", "decision_engine"},
		"enhancer": {"spo_optimizer", "prompt_enhancer", "quality_assessor"},
		"specialist": {"domain_knowledge_base", "technical_validator", "implementation_guide"},
		"coordinator": {"task_scheduler", "resource_manager", "workflow_engine"},
	}

	if toolList, exists := tools[agentType]; exists {
		return toolList
	}

	return []string{"general_processor"}
}

// parseCustomInstructions extracts custom instructions from Other field
func (mm *ModeManager) parseCustomInstructions(other string) string {
	if other == "" {
		return ""
	}

	// Look for instruction patterns
	if strings.Contains(strings.ToLower(other), "instruction") {
		return other
	}

	// Convert other information to instructions
	return fmt.Sprintf("Additional context: %s", other)
}

// parseProviders parses provider list from string
func (mm *ModeManager) parseProviders(providers string) []string {
	if providers == "" || providers == "all" {
		return []string{"all"}
	}

	// Split by pipe or comma
	var providerList []string
	if strings.Contains(providers, "|") {
		providerList = strings.Split(providers, "|")
	} else {
		providerList = strings.Split(providers, ",")
	}

	// Clean up provider names
	for i, provider := range providerList {
		providerList[i] = strings.TrimSpace(provider)
	}

	return providerList
}

// GetMode retrieves a mode configuration and updates usage statistics
func (mm *ModeManager) GetMode(slug string) (ModeConfig, error) {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	mode, exists := mm.modes[slug]
	if !exists {
		return ModeConfig{}, fmt.Errorf("mode not found: %s", slug)
	}

	// Update cache
	if cached, exists := mm.cache[slug]; exists {
		cached.UsageCount++
	} else {
		mm.cache[slug] = &CachedMode{
			Config:     mode,
			CreatedAt:  time.Now(),
			UsageCount: 1,
		}
	}

	return mode, nil
}

// ListModes returns all available modes
func (mm *ModeManager) ListModes() []ModeConfig {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()

	modes := make([]ModeConfig, 0, len(mm.modes))
	for _, mode := range mm.modes {
		modes = append(modes, mode)
	}

	return modes
}

// GetModeUsageStats returns usage statistics for all modes
func (mm *ModeManager) GetModeUsageStats() map[string]*CachedMode {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()

	stats := make(map[string]*CachedMode)
	for k, v := range mm.cache {
		stats[k] = v
	}

	return stats
}

// readAgentCSV reads agent definitions from CSV file
func (mm *ModeManager) readAgentCSV() ([]AgentCSV, error) {
	file, err := os.Open(mm.csvPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var agents []AgentCSV
	for i, record := range records {
		if i == 0 { // Skip header
			continue
		}

		if len(record) >= 5 {
			agents = append(agents, AgentCSV{
				Name:           record[0],
				Type:           record[1],
				Providers:      record[2],
				Specialization: record[3],
				Other:          record[4],
			})
		}
	}

	return agents, nil
}

// saveYAMLCache saves mode configurations as YAML cache
func (mm *ModeManager) saveYAMLCache() error {
	if mm.yamlCache == "" {
		return nil // No cache path specified
	}

	// Create cache directory if it doesn't exist
	if err := os.MkdirAll(mm.yamlCache, 0755); err != nil {
		return err
	}

	// Save each mode as a separate YAML file
	for slug, mode := range mm.modes {
		data, err := yaml.Marshal(mode)
		if err != nil {
			continue // Skip failed marshaling
		}

		filename := fmt.Sprintf("%s/%s.yaml", mm.yamlCache, slug)
		if err := os.WriteFile(filename, data, 0644); err != nil {
			continue // Skip failed writes
		}
	}

	return nil
}

// ReloadModes reloads modes from CSV
func (mm *ModeManager) ReloadModes() error {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	// Clear existing modes
	mm.modes = make(map[string]ModeConfig)

	// Reload from CSV
	return mm.loadFromCSV()
}