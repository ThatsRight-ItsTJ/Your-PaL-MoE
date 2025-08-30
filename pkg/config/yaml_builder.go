package config

import (
	"fmt"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// YAMLBuilder provides functionality to build YAML configurations
type YAMLBuilder struct {
	config map[string]interface{}
}

// NewYAMLBuilder creates a new YAML builder
func NewYAMLBuilder() *YAMLBuilder {
	return &YAMLBuilder{
		config: make(map[string]interface{}),
	}
}

// SetField sets a field in the YAML configuration
func (y *YAMLBuilder) SetField(key string, value interface{}) *YAMLBuilder {
	y.config[key] = value
	return y
}

// SetNested sets a nested field using dot notation
func (y *YAMLBuilder) SetNested(path string, value interface{}) *YAMLBuilder {
	keys := strings.Split(path, ".")
	current := y.config
	
	for i, key := range keys[:len(keys)-1] {
		if _, exists := current[key]; !exists {
			current[key] = make(map[string]interface{})
		}
		if nested, ok := current[key].(map[string]interface{}); ok {
			current = nested
		} else {
			// Create new nested map if type doesn't match
			current[key] = make(map[string]interface{})
			current = current[key].(map[string]interface{})
		}
	}
	
	current[keys[len(keys)-1]] = value
	return y
}

// Build generates the YAML string
func (y *YAMLBuilder) Build() (string, error) {
	data, err := yaml.Marshal(y.config)
	if err != nil {
		return "", fmt.Errorf("failed to marshal YAML: %w", err)
	}
	return string(data), nil
}

// Reset clears the configuration
func (y *YAMLBuilder) Reset() *YAMLBuilder {
	y.config = make(map[string]interface{})
	return y
}

// GetConfig returns the current configuration map
func (y *YAMLBuilder) GetConfig() map[string]interface{} {
	return y.config
}

// AddTimestamp adds a timestamp field
func (y *YAMLBuilder) AddTimestamp() *YAMLBuilder {
	y.config["timestamp"] = time.Now().Format(time.RFC3339)
	return y
}

// AddMetadata adds metadata section
func (y *YAMLBuilder) AddMetadata(metadata map[string]string) *YAMLBuilder {
	if len(metadata) > 0 {
		y.config["metadata"] = metadata
	}
	return y
}