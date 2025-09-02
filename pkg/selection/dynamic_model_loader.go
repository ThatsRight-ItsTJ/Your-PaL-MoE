package selection

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"log"
)

// DynamicModelLoader handles fetching models from provider URLs
type DynamicModelLoader struct {
	httpClient *http.Client
	cache      map[string]CachedModels
}

// CachedModels stores models with expiration
type CachedModels struct {
	Models    []string
	FetchedAt time.Time
	TTL       time.Duration
}

// NewDynamicModelLoader creates a new dynamic model loader
func NewDynamicModelLoader() *DynamicModelLoader {
	return &DynamicModelLoader{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		cache: make(map[string]CachedModels),
	}
}

// LoadModelsFromSource loads models from either URL or pipe-delimited string
func (dml *DynamicModelLoader) LoadModelsFromSource(source string) ([]string, error) {
	// Check if source is a URL
	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		return dml.fetchModelsFromURL(source)
	}
	
	// Otherwise treat as pipe-delimited string
	return dml.parseStaticModels(source), nil
}

// fetchModelsFromURL fetches models from a URL endpoint
func (dml *DynamicModelLoader) fetchModelsFromURL(url string) ([]string, error) {
	// Check cache first
	if cached, exists := dml.cache[url]; exists {
		if time.Since(cached.FetchedAt) < cached.TTL {
			log.Printf("🔄 Using cached models for %s (%d models)", url, len(cached.Models))
			return cached.Models, nil
		}
	}
	
	log.Printf("🌐 Fetching models from URL: %s", url)
	
	resp, err := dml.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch models from %s: %w", url, err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d when fetching models from %s", resp.StatusCode, url)
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response from %s: %w", url, err)
	}
	
	models, err := dml.parseModelResponse(body, url)
	if err != nil {
		return nil, fmt.Errorf("failed to parse models from %s: %w", url, err)
	}
	
	// Cache the results
	dml.cache[url] = CachedModels{
		Models:    models,
		FetchedAt: time.Now(),
		TTL:       5 * time.Minute, // Cache for 5 minutes
	}
	
	log.Printf("✅ Fetched %d models from %s", len(models), url)
	return models, nil
}

// parseModelResponse parses different response formats from model endpoints
func (dml *DynamicModelLoader) parseModelResponse(body []byte, url string) ([]string, error) {
	var models []string
	
	// Try parsing as JSON array of strings
	var jsonArray []string
	if err := json.Unmarshal(body, &jsonArray); err == nil {
		return jsonArray, nil
	}
	
	// Try parsing as JSON object with models field
	var jsonObj map[string]interface{}
	if err := json.Unmarshal(body, &jsonObj); err == nil {
		// Common field names for model lists
		modelFields := []string{"models", "data", "model_list", "available_models"}
		
		for _, field := range modelFields {
			if modelsData, exists := jsonObj[field]; exists {
				switch v := modelsData.(type) {
				case []interface{}:
					for _, item := range v {
						if modelName, ok := item.(string); ok {
							models = append(models, modelName)
						} else if modelObj, ok := item.(map[string]interface{}); ok {
							// Try to extract model name from object
							if name, exists := modelObj["id"]; exists {
								if nameStr, ok := name.(string); ok {
									models = append(models, nameStr)
								}
							} else if name, exists := modelObj["name"]; exists {
								if nameStr, ok := name.(string); ok {
									models = append(models, nameStr)
								}
							} else if name, exists := modelObj["model"]; exists {
								if nameStr, ok := name.(string); ok {
									models = append(models, nameStr)
								}
							}
						}
					}
					if len(models) > 0 {
						return models, nil
					}
				case []string:
					return v, nil
				}
			}
		}
	}
	
	// Try parsing as plain text (newline or comma separated)
	text := strings.TrimSpace(string(body))
	if text != "" {
		// Try newline separated
		if strings.Contains(text, "\n") {
			lines := strings.Split(text, "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line != "" && !strings.HasPrefix(line, "#") {
					models = append(models, line)
				}
			}
			if len(models) > 0 {
				return models, nil
			}
		}
		
		// Try comma separated
		if strings.Contains(text, ",") {
			parts := strings.Split(text, ",")
			for _, part := range parts {
				part = strings.TrimSpace(part)
				if part != "" {
					models = append(models, part)
				}
			}
			if len(models) > 0 {
				return models, nil
			}
		}
		
		// Single model
		models = append(models, text)
		return models, nil
	}
	
	return nil, fmt.Errorf("unable to parse model response from %s", url)
}

// parseStaticModels parses pipe-delimited model string
func (dml *DynamicModelLoader) parseStaticModels(source string) []string {
	if source == "" {
		return []string{}
	}
	
	// Split by pipe and clean up
	parts := strings.Split(source, "|")
	var models []string
	
	for _, part := range parts {
		model := strings.TrimSpace(part)
		if model != "" {
			models = append(models, model)
		}
	}
	
	return models
}

// ClearCache clears the model cache
func (dml *DynamicModelLoader) ClearCache() {
	dml.cache = make(map[string]CachedModels)
	log.Println("🔄 Dynamic model loader cache cleared")
}

// GetCacheStats returns cache statistics
func (dml *DynamicModelLoader) GetCacheStats() map[string]interface{} {
	stats := map[string]interface{}{
		"cached_urls": len(dml.cache),
		"cache_entries": make([]map[string]interface{}, 0),
	}
	
	for url, cached := range dml.cache {
		entry := map[string]interface{}{
			"url":         url,
			"model_count": len(cached.Models),
			"fetched_at":  cached.FetchedAt.Format(time.RFC3339),
			"ttl_seconds": int(cached.TTL.Seconds()),
			"expired":     time.Since(cached.FetchedAt) > cached.TTL,
		}
		stats["cache_entries"] = append(stats["cache_entries"].([]map[string]interface{}), entry)
	}
	
	return stats
}