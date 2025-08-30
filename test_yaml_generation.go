package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/Your-PaL-MoE/internal/enhanced"
)

func main() {
	fmt.Println("🔧 YAML Configuration Generator Test")
	fmt.Println("====================================")

	// Check if providers.csv exists
	if _, err := os.Stat("providers.csv"); os.IsNotExist(err) {
		log.Fatal("❌ providers.csv file not found. Please ensure it exists in the current directory.")
	}

	// Initialize enhanced system
	fmt.Println("🚀 Initializing Enhanced System...")
	system, err := enhanced.NewEnhancedSystem(nil, "providers.csv")
	if err != nil {
		log.Fatalf("❌ Failed to initialize enhanced system: %v", err)
	}
	defer system.Shutdown()

	fmt.Println("✅ Enhanced System initialized successfully")

	// Test generating YAML for all providers
	fmt.Println("\n📝 Generating YAML configurations for all providers...")
	
	ctx := context.Background()
	yamls, err := system.GenerateAllProviderYAMLs(ctx)
	if err != nil {
		log.Fatalf("❌ Failed to generate YAML configurations: %v", err)
	}

	fmt.Printf("✅ Successfully generated %d YAML configurations\n\n", len(yamls))

	// Display generated YAMLs
	for providerName, yamlContent := range yamls {
		fmt.Printf("📋 Provider: %s\n", providerName)
		fmt.Printf("📄 YAML Configuration:\n")
		fmt.Printf("─────────────────────────────────────\n")
		fmt.Printf("%s\n", yamlContent)
		fmt.Printf("─────────────────────────────────────\n\n")
	}

	// Test individual provider YAML generation
	fmt.Println("🔍 Testing individual provider YAML generation...")
	
	providers := system.GetProviders()
	if len(providers) > 0 {
		firstProvider := providers[0]
		fmt.Printf("📋 Testing YAML generation for provider: %s\n", firstProvider.Name)
		
		yaml, err := system.GenerateProviderYAML(ctx, firstProvider.Name)
		if err != nil {
			log.Printf("⚠️  Failed to generate YAML for %s: %v", firstProvider.Name, err)
		} else {
			fmt.Printf("✅ Individual YAML generation successful for %s\n", firstProvider.Name)
			fmt.Printf("📄 Generated YAML:\n%s\n", yaml)
		}
	}

	fmt.Println("🎉 YAML Generation Test Completed Successfully!")
}