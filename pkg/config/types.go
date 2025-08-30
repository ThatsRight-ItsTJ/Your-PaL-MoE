package config

import "time"

// CSVProvider represents a provider loaded from CSV
type CSVProvider struct {
	Name         string `csv:"name"`
	Tier         string `csv:"tier"`
	Endpoint     string `csv:"endpoint"`
	ModelsSource string `csv:"models_source"`
	APIKey       string `csv:"api_key"`
}

// ProviderConfig represents enhanced provider configuration
type ProviderConfig struct {
	ID           string       `yaml:"id"`
	Name         string       `yaml:"name"`
	Tier         string       `yaml:"tier"`
	Endpoint     string       `yaml:"endpoint"`
	Capabilities Capabilities `yaml:"capabilities"`
	CostTracking CostTracking `yaml:"cost_tracking"`
	Metadata     map[string]string `yaml:"metadata"`
}

// Capabilities represents provider capabilities
type Capabilities struct {
	Reasoning     int `yaml:"reasoning"`
	Knowledge     int `yaml:"knowledge"`
	Computation   int `yaml:"computation"`
	Coordination  int `yaml:"coordination"`
}

// CostTracking represents cost tracking information
type CostTracking struct {
	CostPerToken   float64 `yaml:"cost_per_token"`
	CostPerRequest float64 `yaml:"cost_per_request"`
	LastUpdated    time.Time `yaml:"last_updated"`
}