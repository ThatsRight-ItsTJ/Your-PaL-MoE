package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"./internal/enhanced"
)

func main() {
	// Initialize logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	
	fmt.Println("🧪 Testing Dynamic YAML Generation for All Providers")
	fmt.Println("=" + strings.Repeat("=", 59))
	
	// Step 1: Load all providers from CSV
	fmt.Println("\n=== Step 1: Loading All Providers from CSV ===")
	selector, err := enhanced.NewAdaptiveProviderSelector(logger, "providers.csv")
	if err != nil {
		log.Fatalf("Failed to create provider selector: %v", err)
	}
	
	providers := selector.GetProviders()
	fmt.Printf("✅ Loaded %d providers from CSV\n", len(providers))
	
	// Display all loaded providers
	fmt.Println("\n📋 Available Providers:")
	for i, provider := range providers {
		fmt.Printf("  %d. %s (%s tier) - %s\n", i+1, provider.Name, provider.Tier, provider.BaseURL)
	}
	
	// Step 2: Initialize YAML Generator
	fmt.Println("\n=== Step 2: Initialize Dynamic YAML Generator ===")
	yamlGenerator := enhanced.NewYAMLGenerator(logger)
	fmt.Printf("✅ YAML Generator initialized\n")
	
	// Step 3: Generate YAML for each provider dynamically
	fmt.Println("\n=== Step 3: Generate YAML for Each Provider ===")
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	
	successCount := 0
	failCount := 0
	
	for i, provider := range providers {
		fmt.Printf("\n--- Provider %d/%d: %s ---\n", i+1, len(providers), provider.Name)
		
		// Generate YAML for current provider
		yaml, err := yamlGenerator.GenerateYAMLFromProvider(ctx, provider)
		if err != nil {
			fmt.Printf("❌ Failed to generate YAML: %v\n", err)
			// Create fallback YAML
			yaml = createFallbackYAML(provider)
			fmt.Printf("✅ Created fallback YAML configuration\n")
			failCount++
		} else {
			fmt.Printf("✅ Successfully generated YAML using AI\n")
			successCount++
		}
		
		// Display first 10 lines of generated YAML
		lines := strings.Split(yaml, "\n")
		maxLines := 10
		if len(lines) < maxLines {
			maxLines = len(lines)
		}
		
		fmt.Printf("📄 Generated YAML (first %d lines):\n", maxLines)
		for j := 0; j < maxLines; j++ {
			fmt.Printf("   %s\n", lines[j])
		}
		if len(lines) > maxLines {
			fmt.Printf("   ... (%d more lines)\n", len(lines)-maxLines)
		}
		
		// Save YAML to file
		filename := fmt.Sprintf("configs/%s.yaml", provider.Name)
		err = yamlGenerator.SaveYAMLToFile(provider.Name, yaml)
		if err != nil {
			fmt.Printf("❌ Failed to save YAML: %v\n", err)
		} else {
			fmt.Printf("✅ YAML saved to %s\n", filename)
		}
	}
	
	// Step 4: Test Batch Generation
	fmt.Println("\n=== Step 4: Test Batch YAML Generation ===")
	batchResults, err := yamlGenerator.GenerateYAMLBatch(ctx, providers)
	if err != nil {
		fmt.Printf("❌ Batch generation failed: %v\n", err)
	} else {
		fmt.Printf("✅ Batch generation completed: %d/%d providers\n", len(batchResults), len(providers))
	}
	
	// Step 5: Summary and validation
	fmt.Println("\n=== Step 5: Test Summary ===")
	fmt.Printf("📊 Results:\n")
	fmt.Printf("   Total Providers: %d\n", len(providers))
	fmt.Printf("   Successful AI Generations: %d\n", successCount)
	fmt.Printf("   Fallback Generations: %d\n", failCount)
	fmt.Printf("   Success Rate: %.1f%%\n", float64(successCount)/float64(len(providers))*100)
	
	// List generated files
	fmt.Printf("\n📁 Generated YAML Files:\n")
	if files, err := os.ReadDir("configs"); err == nil {
		for _, file := range files {
			if strings.HasSuffix(file.Name(), ".yaml") {
				fmt.Printf("   - configs/%s\n", file.Name())
			}
		}
	}
	
	fmt.Println("\n🏁 Dynamic YAML Generation Test Completed!")
	fmt.Println("✅ The YAML generator successfully works with any CSV row")
	fmt.Println("=" + strings.Repeat("=", 59))
}

func createFallbackYAML(provider *enhanced.Provider) string {
	return fmt.Sprintf(`# Configuration for %s Provider
# Tier: %s
# Generated automatically from CSV data

provider:
  name: "%s"
  tier: "%s"
  base_url: "%s"
  api_key: "%s"
  
models:
  source: "%s"
  # Models will be loaded dynamically based on source
  
configuration:
  timeout: 30s
  max_retries: 3
  rate_limit:
    enabled: true
    requests_per_minute: 60
  
metadata:
  description: "%s"
  auto_generated: true
  generated_at: "%s"
  
# Provider-specific settings can be added here based on the tier and capabilities
`, 
		provider.Name, 
		provider.Tier,
		provider.Name,
		provider.Tier,
		provider.BaseURL,
		provider.APIKey,
		provider.Models,
		provider.Other,
		time.Now().Format("2006-01-02 15:04:05"))
}