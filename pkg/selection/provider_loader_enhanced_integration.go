package selection

import (
	"fmt"
	"log"
	"strings"
)

// LoadProvidersFromCSVWithDynamicModels enhances the existing CSV loader with dynamic model discovery
func LoadProvidersFromCSVWithDynamicModels(filename string) ([]Provider, error) {
	log.Printf("🚀 Loading providers from CSV with dynamic model discovery: %s", filename)
	
	// Load providers using existing CSV logic
	providers, err := LoadProvidersFromCSV(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to load providers from CSV: %w", err)
	}
	
	// Create dynamic model loader
	modelLoader := NewDynamicModelLoader()
	
	// Process each provider to check for dynamic model sources
	for i, provider := range providers {
		originalModels := provider.Models
		log.Printf("🔍 Processing provider: %s (original models: %v)", provider.Name, originalModels)
		
		// Check if any model entry looks like a URL endpoint
		var dynamicModels []string
		var staticModels []string
		
		for _, model := range originalModels {
			if strings.HasPrefix(model, "/") || strings.HasPrefix(model, "http://") || strings.HasPrefix(model, "https://") {
				// This looks like an endpoint - try to fetch models dynamically
				var modelURL string
				if strings.HasPrefix(model, "/") {
					// Relative path - combine with base URL
					baseURL := strings.TrimSuffix(provider.BaseURL, "/")
					modelURL = baseURL + model
				} else {
					// Absolute URL
					modelURL = model
				}
				
				log.Printf("🌐 Attempting dynamic model discovery for %s from: %s", provider.Name, modelURL)
				
				fetchedModels, err := modelLoader.LoadModelsFromSource(modelURL)
				if err != nil {
					log.Printf("⚠️  Failed to fetch models from %s: %v", modelURL, err)
					// Keep the original model entry as fallback
					staticModels = append(staticModels, model)
				} else {
					log.Printf("✅ Successfully fetched %d models from %s", len(fetchedModels), modelURL)
					dynamicModels = append(dynamicModels, fetchedModels...)
				}
			} else {
				// Static model name
				staticModels = append(staticModels, model)
			}
		}
		
		// Combine dynamic and static models
		var finalModels []string
		finalModels = append(finalModels, dynamicModels...)
		finalModels = append(finalModels, staticModels...)
		
		// Remove duplicates
		finalModels = removeDuplicates(finalModels)
		
		if len(finalModels) == 0 {
			log.Printf("⚠️  No models found for %s, using original list", provider.Name)
			finalModels = originalModels
		}
		
		// Update provider with final model list
		providers[i].Models = finalModels
		
		if len(finalModels) != len(originalModels) {
			log.Printf("🔄 Updated %s: %d → %d models", provider.Name, len(originalModels), len(finalModels))
		}
	}
	
	log.Printf("✅ Successfully loaded %d providers with dynamic model discovery", len(providers))
	return providers, nil
}

// removeDuplicates removes duplicate strings from a slice
func removeDuplicates(slice []string) []string {
	seen := make(map[string]bool)
	var result []string
	
	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}
	
	return result
}

// RefreshProviderModelsFromCSV refreshes models for CSV-loaded providers
func RefreshProviderModelsFromCSV(providers []Provider) error {
	log.Println("🔄 Refreshing models for CSV-loaded providers...")
	
	modelLoader := NewDynamicModelLoader()
	modelLoader.ClearCache() // Clear cache to force fresh fetches
	
	updated := 0
	for i, provider := range providers {
		hasEndpoints := false
		var newModels []string
		
		for _, model := range provider.Models {
			if strings.HasPrefix(model, "/") || strings.HasPrefix(model, "http://") || strings.HasPrefix(model, "https://") {
				hasEndpoints = true
				
				var modelURL string
				if strings.HasPrefix(model, "/") {
					baseURL := strings.TrimSuffix(provider.BaseURL, "/")
					modelURL = baseURL + model
				} else {
					modelURL = model
				}
				
				fetchedModels, err := modelLoader.LoadModelsFromSource(modelURL)
				if err != nil {
					log.Printf("⚠️  Failed to refresh models for %s from %s: %v", provider.Name, modelURL, err)
					newModels = append(newModels, model) // Keep original as fallback
				} else {
					newModels = append(newModels, fetchedModels...)
				}
			} else {
				newModels = append(newModels, model)
			}
		}
		
		if hasEndpoints {
			oldCount := len(provider.Models)
			providers[i].Models = removeDuplicates(newModels)
			newCount := len(providers[i].Models)
			
			if oldCount != newCount {
				log.Printf("🔄 Refreshed %s: %d → %d models", provider.Name, oldCount, newCount)
				updated++
			}
		}
	}
	
	log.Printf("✅ Refreshed models for %d providers", updated)
	return nil
}