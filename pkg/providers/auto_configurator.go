package providers

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// AutoConfigurator generates provider configurations automatically
type AutoConfigurator struct {
	csvPath   string
	outputDir string
	parser    *CSVParser
}

// NewAutoConfigurator creates a new auto-configurator instance
func NewAutoConfigurator(csvPath, outputDir string) *AutoConfigurator {
	return &AutoConfigurator{
		csvPath:   csvPath,
		outputDir: outputDir,
		parser:    NewCSVParser(csvPath),
	}
}

// GenerateConfigurations creates YAML configurations for all providers
func (a *AutoConfigurator) GenerateConfigurations(ctx context.Context) error {
	providers, err := a.parser.LoadProviders()
	if err != nil {
		return fmt.Errorf("failed to load providers: %w", err)
	}

	for _, provider := range providers {
		if err := a.generateProviderConfig(ctx, provider); err != nil {
			return fmt.Errorf("failed to generate config for %s: %w", provider.Name, err)
		}
	}

	return nil
}

// generateProviderConfig creates a YAML configuration file for a single provider
func (a *AutoConfigurator) generateProviderConfig(ctx context.Context, provider *ProviderConfig) error {
	// Create tier directory
	tierDir := filepath.Join(a.outputDir, provider.Tier)
	if err := os.MkdirAll(tierDir, 0755); err != nil {
		return fmt.Errorf("failed to create tier directory: %w", err)
	}

	// Generate filename
	filename := strings.ToLower(strings.ReplaceAll(provider.Name, " ", "_")) + ".yaml"
	configPath := filepath.Join(tierDir, filename)

	// Generate configuration content
	configContent, err := a.generateConfigContent(provider)
	if err != nil {
		return fmt.Errorf("failed to generate config content: %w", err)
	}

	// Write configuration file
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// generateConfigContent creates the YAML content for a provider
func (a *AutoConfigurator) generateConfigContent(provider *ProviderConfig) (string, error) {
	tmpl := `name: {{ .Name }}
tier: {{ .Tier }}
endpoint: {{ .Endpoint }}
enabled: {{ .Enabled }}
{{- if .ApiKey }}
api_key: {{ .ApiKey }}
{{- end }}

{{- if .Headers }}
headers:
{{- range $key, $value := .Headers }}
  {{ $key }}: {{ $value }}
{{- end }}
{{- end }}

models_source:
  type: {{ .ModelsSource.Type }}
{{- if eq .ModelsSource.Type "list" }}
  models:
{{- range .ModelsSource.Value }}
    - {{ . }}
{{- end }}
{{- else if eq .ModelsSource.Type "endpoint" }}
  endpoint: {{ .ModelsSource.Value }}
{{- else if eq .ModelsSource.Type "script" }}
  script: {{ .ModelsSource.Value }}
{{- end }}

# Provider-specific configuration
{{- if eq .Tier "official" }}
authentication:
  type: "api_key"
  header: "Authorization"
  prefix: "Bearer "
rate_limits:
  requests_per_minute: 60
  requests_per_day: 1000
{{- else if eq .Tier "community" }}
rate_limits:
  requests_per_minute: 30
  requests_per_day: 500
{{- else if eq .Tier "unofficial" }}
session_management:
  enabled: true
  rotation_interval: "1h"
  pool_size: 5
rate_limits:
  requests_per_minute: 20
  requests_per_day: 200
{{- end }}

health_check:
  enabled: true
  interval: "5m"
  timeout: "30s"
  endpoint: "{{ .Endpoint }}/health"
  
cost_optimization:
  priority: {{ if eq .Tier "unofficial" }}1{{ else if eq .Tier "community" }}2{{ else }}3{{ end }}
  cost_per_request: {{ if eq .Tier "unofficial" }}0.0{{ else if eq .Tier "community" }}0.001{{ else }}0.01{{ end }}
`

	t, err := template.New("provider").Parse(tmpl)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf strings.Builder
	if err := t.Execute(&buf, provider); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// GenerateScriptTemplate creates a template script for unofficial providers
func (a *AutoConfigurator) GenerateScriptTemplate(provider *ProviderConfig, scriptPath string) error {
	if provider.Tier != "unofficial" {
		return nil // Only generate scripts for unofficial providers
	}

	// Create scripts directory if it doesn't exist
	scriptDir := filepath.Dir(scriptPath)
	if err := os.MkdirAll(scriptDir, 0755); err != nil {
		return fmt.Errorf("failed to create script directory: %w", err)
	}

	// Generate script template
	scriptContent := a.generateScriptTemplate(provider)

	// Write script file
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
		return fmt.Errorf("failed to write script file: %w", err)
	}

	return nil
}

// generateScriptTemplate creates a Python script template for unofficial APIs
func (a *AutoConfigurator) generateScriptTemplate(provider *ProviderConfig) string {
	return fmt.Sprintf(`#!/usr/bin/env python3
"""
Auto-generated script template for %s
Edit this file to implement the actual API integration
"""

import sys
import json
import requests
from typing import Dict, Any

def main():
    """Main entry point for the script"""
    try:
        # Read request data from stdin
        request_data = json.loads(sys.stdin.read())
        
        # Extract common parameters
        prompt = request_data.get('prompt', '')
        model = request_data.get('model', 'default')
        max_tokens = request_data.get('max_tokens', 100)
        
        # TODO: Implement actual API call to %s
        # Replace this mock implementation with real API integration
        
        response = call_api(prompt, model, max_tokens)
        
        # Return response in standard format
        result = {
            'success': True,
            'data': response,
            'cost': 0.0,  # Update with actual cost if available
            'provider': '%s'
        }
        
        print(json.dumps(result))
        
    except Exception as e:
        error_result = {
            'success': False,
            'error': str(e),
            'provider': '%s'
        }
        print(json.dumps(error_result))
        sys.exit(1)

def call_api(prompt: str, model: str, max_tokens: int) -> Dict[str, Any]:
    """
    TODO: Implement actual API call
    
    Args:
        prompt: User prompt
        model: Model to use
        max_tokens: Maximum tokens to generate
        
    Returns:
        API response data
    """
    
    # Mock implementation - replace with actual API call
    headers = {
        'Content-Type': 'application/json',
        'User-Agent': 'Intelligent-AI-Gateway/1.0'
    }
    
    # TODO: Add authentication headers if needed
    # headers['Authorization'] = 'Bearer YOUR_TOKEN'
    
    payload = {
        'prompt': prompt,
        'model': model,
        'max_tokens': max_tokens
    }
    
    try:
        # TODO: Replace with actual endpoint
        response = requests.post('%s', headers=headers, json=payload, timeout=30)
        response.raise_for_status()
        
        # TODO: Parse actual response format
        return {
            'text': f'Mock response to: {prompt}',
            'model': model,
            'usage': {
                'prompt_tokens': len(prompt.split()),
                'completion_tokens': 10,
                'total_tokens': len(prompt.split()) + 10
            }
        }
        
    except requests.exceptions.RequestException as e:
        raise Exception(f'API request failed: {e}')

if __name__ == '__main__':
    main()
`, provider.Name, provider.Endpoint, provider.Name, provider.Name, provider.Endpoint)
}