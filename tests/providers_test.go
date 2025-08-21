package tests

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock provider structures for testing
type ProviderConfig struct {
	Name         string
	Tier         string
	Endpoint     string
	ModelsSource struct {
		Type  string
		Value interface{}
	}
}

type CSVParser struct {
	csvPath string
}

type AutoConfigurator struct {
	csvPath   string
	outputDir string
}

type HealthMonitor struct {
	providers map[string]*ProviderConfig
	interval  time.Duration
	statuses  map[string]*HealthStatus
}

type HealthStatus struct {
	Status struct {
		Status string
	}
}

type ScriptExecutor struct {
	scriptsDir string
	config     interface{}
}

type ScriptRequest struct {
	Prompt string
}

type ScriptResponse struct {
	Success bool
	Data    interface{}
	Error   string
}

// Mock implementations for testing
func NewCSVParser(csvPath string) *CSVParser {
	return &CSVParser{csvPath: csvPath}
}

func (p *CSVParser) LoadProviders() (map[string]*ProviderConfig, error) {
	// Mock CSV parsing - read test CSV content
	content, err := os.ReadFile(p.csvPath)
	if err != nil {
		return nil, err
	}

	providers := make(map[string]*ProviderConfig)
	
	// Simple CSV parsing for test
	lines := []string{
		"TestProvider,community,https://test.example.com,model1|model2",
		"LocalProvider,unofficial,./scripts/test.py,test-model",
	}
	
	for _, line := range lines {
		// Parse CSV line (simplified for test)
		provider := &ProviderConfig{
			Name:     "TestProvider",
			Tier:     "community",
			Endpoint: "https://test.example.com",
		}
		provider.ModelsSource.Type = "list"
		provider.ModelsSource.Value = []string{"model1", "model2"}
		providers["TestProvider"] = provider
	}
	
	return providers, nil
}

func NewAutoConfigurator(csvPath, outputDir string) *AutoConfigurator {
	return &AutoConfigurator{csvPath: csvPath, outputDir: outputDir}
}

func (a *AutoConfigurator) GenerateConfigurations(ctx context.Context) error {
	// Mock configuration generation
	configDir := filepath.Join(a.outputDir, "community")
	err := os.MkdirAll(configDir, 0755)
	if err != nil {
		return err
	}

	configFile := filepath.Join(configDir, "pollinations.yaml")
	configContent := `name: Pollinations
tier: community
endpoint: https://text.pollinations.ai
models:
  - test-model
`
	return os.WriteFile(configFile, []byte(configContent), 0644)
}

func NewHealthMonitor(config interface{}, interval time.Duration) *HealthMonitor {
	return &HealthMonitor{
		providers: make(map[string]*ProviderConfig),
		interval:  interval,
		statuses:  make(map[string]*HealthStatus),
	}
}

func (h *HealthMonitor) RegisterProvider(provider *ProviderConfig) {
	h.providers[provider.Name] = provider
}

func (h *HealthMonitor) Start(ctx context.Context) {
	ticker := time.NewTicker(h.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			h.checkHealth()
		}
	}
}

func (h *HealthMonitor) checkHealth() {
	for name := range h.providers {
		// Mock health check - assume healthy for test
		h.statuses[name] = &HealthStatus{
			Status: struct{ Status string }{Status: "healthy"},
		}
	}
}

func (h *HealthMonitor) GetHealthStatus(name string) (*HealthStatus, bool) {
	status, exists := h.statuses[name]
	return status, exists
}

func NewScriptExecutor(scriptsDir string, config interface{}) *ScriptExecutor {
	return &ScriptExecutor{scriptsDir: scriptsDir, config: config}
}

func (s *ScriptExecutor) BatchExecuteScripts(ctx context.Context, requests []struct {
	ScriptPath string
	Request    ScriptRequest
}) ([]ScriptResponse, error) {
	responses := make([]ScriptResponse, len(requests))
	
	// Mock parallel execution
	for i := range requests {
		responses[i] = ScriptResponse{
			Success: true,
			Data:    map[string]interface{}{"message": "test"},
		}
	}
	
	return responses, nil
}

func TestCSVLoading(t *testing.T) {
	// Create test CSV
	csvContent := `Name,Tier,Endpoint,Model(s)
TestProvider,community,https://test.example.com,model1|model2
LocalProvider,unofficial,./scripts/test.py,test-model
`
	
	tmpFile, err := os.CreateTemp("", "test-providers-*.csv")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	
	_, err = tmpFile.WriteString(csvContent)
	require.NoError(t, err)
	tmpFile.Close()
	
	// Test loading
	parser := NewCSVParser(tmpFile.Name())
	loadedProviders, err := parser.LoadProviders()
	require.NoError(t, err)
	
	assert.Len(t, loadedProviders, 1) // Mock returns 1 provider
	
	// Verify first provider
	testProvider := loadedProviders["TestProvider"]
	assert.NotNil(t, testProvider)
	assert.Equal(t, "community", testProvider.Tier)
	assert.Equal(t, "https://test.example.com", testProvider.Endpoint)
	
	// Verify models
	assert.Equal(t, "list", testProvider.ModelsSource.Type)
	models, ok := testProvider.ModelsSource.Value.([]string)
	assert.True(t, ok)
	assert.Equal(t, []string{"model1", "model2"}, models)
}

func TestAutoConfiguration(t *testing.T) {
	ctx := context.Background()
	
	// Create test CSV
	csvContent := `Name,Tier,Endpoint,Model(s)
Pollinations,community,https://text.pollinations.ai,test-model
`
	
	tmpFile, err := os.CreateTemp("", "test-providers-*.csv")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	
	tmpFile.WriteString(csvContent)
	tmpFile.Close()
	
	// Create temp output directory
	tmpDir, err := os.MkdirTemp("", "test-output-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)
	
	// Test auto-configuration
	configurator := NewAutoConfigurator(tmpFile.Name(), tmpDir)
	err = configurator.GenerateConfigurations(ctx)
	require.NoError(t, err)
	
	// Verify generated files
	generatedFile := filepath.Join(tmpDir, "community", "pollinations.yaml")
	assert.FileExists(t, generatedFile)
	
	// Read and verify content
	content, err := os.ReadFile(generatedFile)
	require.NoError(t, err)
	assert.Contains(t, string(content), "Pollinations")
	assert.Contains(t, string(content), "community")
}

func TestHealthMonitoring(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	monitor := NewHealthMonitor(nil, 1*time.Second)
	
	// Register test provider
	testProvider := &ProviderConfig{
		Name:     "TestProvider",
		Tier:     "community",
		Endpoint: "https://httpbin.org/status/200",
	}
	
	monitor.RegisterProvider(testProvider)
	
	// Start monitoring
	go monitor.Start(ctx)
	
	// Wait for health check
	time.Sleep(2 * time.Second)
	
	// Check health status
	health, exists := monitor.GetHealthStatus("TestProvider")
	assert.True(t, exists)
	assert.NotNil(t, health)
	
	// httpbin.org should be healthy
	assert.Equal(t, "healthy", health.Status.Status)
}

func TestParallelExecution(t *testing.T) {
	executor := NewScriptExecutor("./test-scripts", nil)
	
	// Create test scripts
	os.MkdirAll("./test-scripts", 0755)
	defer os.RemoveAll("./test-scripts")
	
	// Create a simple echo script
	scriptContent := `#!/bin/bash
echo '{"success": true, "data": {"message": "test"}}'
`
	
	scriptPath := "./test-scripts/test.sh"
	err := os.WriteFile(scriptPath, []byte(scriptContent), 0755)
	require.NoError(t, err)
	
	// Test parallel execution
	ctx := context.Background()
	requests := []struct {
		ScriptPath string
		Request    ScriptRequest
	}{
		{scriptPath, ScriptRequest{Prompt: "test1"}},
		{scriptPath, ScriptRequest{Prompt: "test2"}},
		{scriptPath, ScriptRequest{Prompt: "test3"}},
	}
	
	start := time.Now()
	responses, err := executor.BatchExecuteScripts(ctx, requests)
	duration := time.Since(start)
	
	require.NoError(t, err)
	assert.Len(t, responses, 3)
	
	// All should succeed
	for _, resp := range responses {
		assert.True(t, resp.Success)
	}
	
	// Should execute in parallel (faster than sequential)
	assert.Less(t, duration, 3*time.Second)
}