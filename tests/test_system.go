package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/ThatsRight-ItsTJ/Your-PaL-MoE/internal/enhanced"
)

// Test script to verify the enhanced system functionality
func main() {
	// Setup logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	
	fmt.Println("🚀 Testing Enhanced Your-PaL-MoE System")
	fmt.Println("=====================================")
	
	// Test 1: CSV Loading and Provider Detection
	testCSVLoading(logger)
	
	// Test 2: YAML Generation
	testYAMLGeneration(logger)
	
	// Test 3: CSV Hot Reload
	testCSVHotReload(logger)
	
	fmt.Println("\n🏁 All tests completed!")
}

func testCSVLoading(logger *logrus.Logger) {
	fmt.Println("\n=== Test 1: CSV Loading ===")
	
	// Create test CSV
	createTestCSV()
	
	// Initialize provider selector
	selector, err := enhanced.NewAdaptiveProviderSelector(logger, "test_providers.csv")
	if err != nil {
		log.Fatalf("❌ Failed to create provider selector: %v", err)
	}
	defer selector.Close()
	
	// Get providers
	providers := selector.GetProviders()
	
	if len(providers) > 0 {
		fmt.Printf("✅ Successfully loaded %d providers\n", len(providers))
		for _, provider := range providers {
			fmt.Printf("  - %s (%s, %s)\n", provider.Name, provider.ID, provider.Tier)
		}
	} else {
		fmt.Println("❌ No providers loaded")
	}
}

func testYAMLGeneration(logger *logrus.Logger) {
	fmt.Println("\n=== Test 2: YAML Generation ===")
	
	// Initialize YAML generator
	generator := enhanced.NewYAMLGenerator(logger)
	
	// Create test provider
	provider := &enhanced.Provider{
		ID:           "test_provider",
		Name:         "Test Provider", 
		Tier:         enhanced.CommunityTier,
		Endpoint:     "https://example.com/api",
		APIKey:       "",
		Model:        "test-model",
		CostPerToken: 0.001,
		MaxTokens:    2048,
		Capabilities: []string{"chat"},
		Metadata: map[string]interface{}{
			"additional_info": "rate_limit:100/min",
			"rate_limit":      "100/min",
		},
	}
	
	fmt.Println("Generating YAML for test provider...")
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	yaml, err := generator.GenerateYAMLFromProvider(ctx, provider)
	if err != nil {
		fmt.Printf("❌ YAML generation failed: %v\n", err)
		return
	}
	
	fmt.Println("✅ YAML generation successful")
	fmt.Println("Generated YAML:")
	fmt.Println("================")
	fmt.Printf("%s\n", yaml)
	fmt.Println("================")
	
	// Save to file
	if err := saveToFile("generated_config.yaml", yaml); err != nil {
		log.Printf("Warning: Could not save YAML: %v", err)
	} else {
		fmt.Println("✅ YAML saved to generated_config.yaml")
	}
}

func testCSVHotReload(logger *logrus.Logger) {
	fmt.Println("\n=== Test 3: CSV Hot Reload ===")
	
	// Create initial CSV
	createTestCSV()
	
	// Initialize provider selector
	selector, err := enhanced.NewAdaptiveProviderSelector(logger, "test_providers.csv")
	if err != nil {
		log.Fatalf("❌ Failed to create provider selector: %v", err)
	}
	defer selector.Close()
	
	// Check initial providers
	initialProviders := selector.GetProviders()
	fmt.Printf("Initial providers count: %d\n", len(initialProviders))
	
	// Modify CSV file
	fmt.Println("Modifying CSV file...")
	createModifiedCSV()
	
	// Wait for file watcher to detect changes
	fmt.Println("Waiting for file watcher to detect changes...")
	time.Sleep(2 * time.Second)
	
	// Check if providers were reloaded
	updatedProviders := selector.GetProviders()
	fmt.Printf("Updated providers count: %d\n", len(updatedProviders))
	
	if len(updatedProviders) > len(initialProviders) {
		fmt.Println("✅ CSV hot reload successful")
		fmt.Println("New providers:")
		for i := len(initialProviders); i < len(updatedProviders); i++ {
			fmt.Printf("  + %s (%s)\n", updatedProviders[i].Name, updatedProviders[i].ID)
		}
	} else {
		fmt.Println("❌ CSV hot reload may not have worked (check logs)")
	}
	
	// Test manual reload
	fmt.Println("\nTesting manual reload...")
	if err := selector.ReloadProviders(); err != nil {
		fmt.Printf("❌ Manual reload failed: %v\n", err)
	} else {
		fmt.Println("✅ Manual reload successful")
	}
}

func createTestCSV() {
	content := `ID,Name,Tier,Endpoint,APIKey,Model,CostPerToken,MaxTokens,Capabilities,AdditionalInfo
pollinations,Pollinations,community,https://text.pollinations.ai,,openai,0.000001,2048,chat;creative,no_auth:true,free_tier:true
test_provider,Test Provider,community,https://example.com/api,,test-model,0.001,2048,chat,rate_limit:100/min`

	if err := os.WriteFile("test_providers.csv", []byte(content), 0644); err != nil {
		log.Fatalf("Failed to create test CSV: %v", err)
	}
}

func createModifiedCSV() {
	content := `ID,Name,Tier,Endpoint,APIKey,Model,CostPerToken,MaxTokens,Capabilities,AdditionalInfo
pollinations,Pollinations,community,https://text.pollinations.ai,,openai,0.000001,2048,chat;creative,no_auth:true,free_tier:true
test_provider,Test Provider,community,https://example.com/api,,test-model,0.001,2048,chat,rate_limit:100/min
new_provider,New Provider,unofficial,http://localhost:8080,,llama-2,0,4096,chat,local:true,gpu_required:false`

	if err := os.WriteFile("test_providers.csv", []byte(content), 0644); err != nil {
		log.Fatalf("Failed to create modified CSV: %v", err)
	}
}

func saveToFile(filename, content string) error {
	return os.WriteFile(filename, []byte(content), 0644)
}