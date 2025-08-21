package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ScriptRequest represents a request to execute a script
type ScriptRequest struct {
	Prompt    string                 `json:"prompt"`
	Model     string                 `json:"model"`
	MaxTokens int                    `json:"max_tokens"`
	Options   map[string]interface{} `json:"options"`
}

// ScriptResponse represents the response from script execution
type ScriptResponse struct {
	Success  bool                   `json:"success"`
	Data     interface{}            `json:"data"`
	Error    string                 `json:"error,omitempty"`
	Cost     float64                `json:"cost"`
	Duration time.Duration          `json:"duration"`
	Provider string                 `json:"provider"`
}

// ScriptExecutor manages execution of unofficial API scripts
type ScriptExecutor struct {
	scriptsDir string
	config     interface{}
	timeout    time.Duration
}

// NewScriptExecutor creates a new script executor instance
func NewScriptExecutor(scriptsDir string, config interface{}) *ScriptExecutor {
	return &ScriptExecutor{
		scriptsDir: scriptsDir,
		config:     config,
		timeout:    30 * time.Second,
	}
}

// ExecuteScript runs a single script with the given request
func (s *ScriptExecutor) ExecuteScript(ctx context.Context, scriptPath string, request ScriptRequest) (*ScriptResponse, error) {
	start := time.Now()
	
	// Validate script exists
	if !filepath.IsAbs(scriptPath) {
		scriptPath = filepath.Join(s.scriptsDir, scriptPath)
	}
	
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return &ScriptResponse{
			Success:  false,
			Error:    fmt.Sprintf("script not found: %s", scriptPath),
			Duration: time.Since(start),
		}, nil
	}

	// Prepare request JSON
	requestJSON, err := json.Marshal(request)
	if err != nil {
		return &ScriptResponse{
			Success:  false,
			Error:    fmt.Sprintf("failed to marshal request: %v", err),
			Duration: time.Since(start),
		}, nil
	}

	// Create context with timeout
	ctxWithTimeout, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	// Execute script
	cmd := exec.CommandContext(ctxWithTimeout, s.getScriptInterpreter(scriptPath), scriptPath)
	cmd.Stdin = strings.NewReader(string(requestJSON))
	
	output, err := cmd.Output()
	duration := time.Since(start)

	if err != nil {
		return &ScriptResponse{
			Success:  false,
			Error:    fmt.Sprintf("script execution failed: %v", err),
			Duration: duration,
		}, nil
	}

	// Parse script output
	var response ScriptResponse
	if err := json.Unmarshal(output, &response); err != nil {
		return &ScriptResponse{
			Success:  false,
			Error:    fmt.Sprintf("failed to parse script output: %v", err),
			Duration: duration,
		}, nil
	}

	response.Duration = duration
	return &response, nil
}

// BatchExecuteScripts executes multiple scripts in parallel
func (s *ScriptExecutor) BatchExecuteScripts(ctx context.Context, requests []struct {
	ScriptPath string
	Request    ScriptRequest
}) ([]ScriptResponse, error) {
	responses := make([]ScriptResponse, len(requests))
	var wg sync.WaitGroup
	var mutex sync.Mutex

	for i, req := range requests {
		wg.Add(1)
		go func(index int, scriptPath string, request ScriptRequest) {
			defer wg.Done()

			response, err := s.ExecuteScript(ctx, scriptPath, request)
			
			mutex.Lock()
			defer mutex.Unlock()
			
			if err != nil {
				responses[index] = ScriptResponse{
					Success: false,
					Error:   err.Error(),
				}
			} else {
				responses[index] = *response
			}
		}(i, req.ScriptPath, req.Request)
	}

	wg.Wait()
	return responses, nil
}

// getScriptInterpreter determines the appropriate interpreter for a script
func (s *ScriptExecutor) getScriptInterpreter(scriptPath string) string {
	ext := strings.ToLower(filepath.Ext(scriptPath))
	
	switch ext {
	case ".py":
		return "python3"
	case ".js":
		return "node"
	case ".sh":
		return "bash"
	default:
		// Try to detect shebang
		if shebang := s.getShebang(scriptPath); shebang != "" {
			return shebang
		}
		return "python3" // Default to Python
	}
}

// getShebang reads the shebang line from a script
func (s *ScriptExecutor) getShebang(scriptPath string) string {
	file, err := os.Open(scriptPath)
	if err != nil {
		return ""
	}
	defer file.Close()

	buffer := make([]byte, 256)
	n, err := file.Read(buffer)
	if err != nil || n < 2 {
		return ""
	}

	content := string(buffer[:n])
	lines := strings.Split(content, "\n")
	if len(lines) == 0 || !strings.HasPrefix(lines[0], "#!") {
		return ""
	}

	shebang := strings.TrimPrefix(lines[0], "#!")
	shebang = strings.TrimSpace(shebang)
	
	// Extract just the interpreter name
	parts := strings.Fields(shebang)
	if len(parts) > 0 {
		interpreter := filepath.Base(parts[0])
		return interpreter
	}

	return ""
}

// ValidateScript checks if a script is valid and executable
func (s *ScriptExecutor) ValidateScript(scriptPath string) error {
	if !filepath.IsAbs(scriptPath) {
		scriptPath = filepath.Join(s.scriptsDir, scriptPath)
	}

	// Check if file exists
	info, err := os.Stat(scriptPath)
	if os.IsNotExist(err) {
		return fmt.Errorf("script does not exist: %s", scriptPath)
	}
	if err != nil {
		return fmt.Errorf("failed to stat script: %w", err)
	}

	// Check if file is executable
	if info.Mode()&0111 == 0 {
		return fmt.Errorf("script is not executable: %s", scriptPath)
	}

	// Try to determine interpreter
	interpreter := s.getScriptInterpreter(scriptPath)
	if interpreter == "" {
		return fmt.Errorf("cannot determine script interpreter for: %s", scriptPath)
	}

	// Check if interpreter is available
	if _, err := exec.LookPath(interpreter); err != nil {
		return fmt.Errorf("interpreter '%s' not found in PATH", interpreter)
	}

	return nil
}

// ListScripts returns a list of all available scripts
func (s *ScriptExecutor) ListScripts() ([]string, error) {
	var scripts []string

	err := filepath.Walk(s.scriptsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Check if it's a script file
		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".py" || ext == ".js" || ext == ".sh" || s.getShebang(path) != "" {
			relPath, err := filepath.Rel(s.scriptsDir, path)
			if err != nil {
				return err
			}
			scripts = append(scripts, relPath)
		}

		return nil
	})

	return scripts, err
}

// CreateScriptFromTemplate creates a new script from a template
func (s *ScriptExecutor) CreateScriptFromTemplate(scriptPath string, provider *ProviderConfig) error {
	if !filepath.IsAbs(scriptPath) {
		scriptPath = filepath.Join(s.scriptsDir, scriptPath)
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(scriptPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create script directory: %w", err)
	}

	// Generate script content using auto-configurator
	configurator := NewAutoConfigurator("", "")
	scriptContent := configurator.generateScriptTemplate(provider)

	// Write script file
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
		return fmt.Errorf("failed to write script file: %w", err)
	}

	return nil
}