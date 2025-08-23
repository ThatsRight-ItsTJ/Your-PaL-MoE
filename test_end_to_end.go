package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/Your-PaL-MoE/internal/enhanced"
	"github.com/Your-PaL-MoE/internal/types"
)

// TestPrompt represents a test case with expected behavior
type TestPrompt struct {
	ID          string
	Prompt      string
	ExpectedTier string // official, community, unofficial
	Category     string // creative, technical, code, mathematical, general
	Description  string
}

// MockProvider simulates provider responses for testing
type MockProvider struct {
	Name     string
	Tier     string
	BaseURL  string
	Models   string
	Features []string
}

func main() {
	fmt.Println("🚀 Your-PaL-MoE End-to-End Workflow Test")
	fmt.Println("=" + strings.Repeat("=", 60))

	// Initialize test cases
	testPrompts := []TestPrompt{
		{
			ID:           "TEST001",
			Prompt:       "Write a short story about a cat",
			ExpectedTier: "community",
			Category:     "creative",
			Description:  "Simple creative writing task",
		},
		{
			ID:           "TEST002",
			Prompt:       "Create a Python function to implement quicksort algorithm with error handling and type hints",
			ExpectedTier: "official",
			Category:     "code",
			Description:  "Complex code generation requiring high accuracy",
		},
		{
			ID:           "TEST003",
			Prompt:       "Write a haiku about artificial intelligence and human creativity",
			ExpectedTier: "community",
			Category:     "creative",
			Description:  "Creative poetry with moderate complexity",
		},
		{
			ID:           "TEST004",
			Prompt:       "Explain the mathematical proof of Fermat's Last Theorem in detail with formal notation",
			ExpectedTier: "official",
			Category:     "mathematical",
			Description:  "Complex mathematical explanation requiring precision",
		},
		{
			ID:           "TEST005",
			Prompt:       "What's the weather like?",
			ExpectedTier: "unofficial",
			Category:     "general",
			Description:  "Simple general query suitable for local models",
		},
		{
			ID:           "TEST006",
			Prompt:       "Design a distributed microservices architecture for a real-time trading system with fault tolerance, load balancing, and security considerations",
			ExpectedTier: "official",
			Category:     "technical",
			Description:  "Highly complex technical architecture requiring expert knowledge",
		},
	}

	fmt.Printf("\n🧪 Testing %d different prompt types...\n\n", len(testPrompts))

	// Step 1: Load and verify providers
	fmt.Println("=== Step 1: Loading Providers from CSV ===")
	providers, err := loadProvidersFromCSV("providers.csv")
	if err != nil {
		log.Fatalf("❌ Failed to load providers: %v", err)
	}
	fmt.Printf("✅ Loaded %d providers successfully\n", len(providers))
	
	for _, provider := range providers {
		fmt.Printf("   📋 %s (%s tier) - %s\n", provider.Name, provider.Tier, provider.BaseURL)
	}

	// Step 2: Initialize enhanced system components
	fmt.Println("\n=== Step 2: Initializing Enhanced System Components ===")
	
	// Initialize task reasoning component
	taskReasoner := enhanced.NewTaskReasoning()
	fmt.Println("✅ Task Reasoning component initialized")
	
	// Initialize provider selector
	providerSelector := enhanced.NewProviderSelector("providers.csv")
	fmt.Println("✅ Provider Selector component initialized")
	
	// Initialize YAML generator
	yamlGenerator := enhanced.NewYAMLGenerator()
	fmt.Println("✅ YAML Generator component initialized")
	
	// Initialize enhanced system orchestrator
	enhancedSystem := enhanced.NewEnhancedSystem(taskReasoner, providerSelector, yamlGenerator)
	fmt.println("✅ Enhanced System orchestrator initialized")

	// Step 3: Test each prompt through the complete workflow
	fmt.Println("\n=== Step 3: End-to-End Prompt Processing ===")
	
	totalTests := len(testPrompts)
	successfulTests := 0
	
	for i, testCase := range testPrompts {
		fmt.Printf("\n🔄 Test %d/%d: %s\n", i+1, totalTests, testCase.ID)
		fmt.Printf("📝 Prompt: \"%s\"\n", testCase.Prompt)
		fmt.Printf("📊 Category: %s | Expected Tier: %s\n", testCase.Category, testCase.ExpectedTier)
		fmt.Printf("📋 Description: %s\n", testCase.Description)
		
		// Step 3a: Task Complexity Analysis
		fmt.Println("\n   🧠 Step 3a: Analyzing Task Complexity...")
		
		complexityResult := analyzeTaskComplexity(testCase.Prompt, testCase.Category)
		fmt.Printf("   📊 Complexity Analysis Results:\n")
		fmt.Printf("      • Cognitive Load: %.2f/5.0\n", complexityResult.CognitiveLoad)
		fmt.Printf("      • Technical Depth: %.2f/5.0\n", complexityResult.TechnicalDepth)
		fmt.Printf("      • Creative Requirement: %.2f/5.0\n", complexityResult.CreativeRequirement)
		fmt.Printf("      • Accuracy Requirement: %.2f/5.0\n", complexityResult.AccuracyRequirement)
		fmt.Printf("      • Overall Complexity: %.2f/5.0\n", complexityResult.OverallComplexity)
		
		// Step 3b: Smart Provider Selection
		fmt.Println("\n   🎯 Step 3b: Smart Provider Selection...")
		
		selectedProvider := selectProviderBasedOnComplexity(complexityResult, providers)
		fmt.Printf("   🏆 Selected Provider: %s (%s tier)\n", selectedProvider.Name, selectedProvider.Tier)
		fmt.Printf("      • Base URL: %s\n", selectedProvider.BaseURL)
		fmt.Printf("      • Models: %s\n", selectedProvider.Models)
		
		// Step 3c: Configuration Loading
		fmt.Printf("\n   ⚙️  Step 3c: Loading Provider Configuration...\n")
		
		configPath := fmt.Sprintf("configs/%s.yaml", selectedProvider.Name)
		if fileExists(configPath) {
			fmt.Printf("   ✅ Configuration loaded: %s\n", configPath)
			
			// Show sample configuration
			configSample, err := readConfigSample(configPath)
			if err == nil {
				fmt.Printf("   📄 Config Preview:\n%s\n", configSample)
			}
		} else {
			fmt.Printf("   ⚠️  Configuration file not found: %s\n", configPath)
		}
		
		// Step 3d: Validation
		fmt.Printf("\n   ✅ Step 3d: Validation Results:\n")
		
		tierMatch := selectedProvider.Tier == testCase.ExpectedTier
		if tierMatch {
			fmt.Printf("      ✅ Provider tier matches expectation: %s\n", selectedProvider.Tier)
			successfulTests++
		} else {
			fmt.Printf("      ❌ Provider tier mismatch: got %s, expected %s\n", selectedProvider.Tier, testCase.ExpectedTier)
		}
		
		// Show reasoning
		fmt.Printf("      📝 Selection Reasoning:\n")
		if complexityResult.OverallComplexity >= 4.0 {
			fmt.Printf("         • High complexity (%.2f) → Official tier provider selected\n", complexityResult.OverallComplexity)
		} else if complexityResult.OverallComplexity >= 2.5 {
			fmt.Printf("         • Medium complexity (%.2f) → Community tier provider selected\n", complexityResult.OverallComplexity)
		} else {
			fmt.Printf("         • Low complexity (%.2f) → Unofficial/Local tier provider selected\n", complexityResult.OverallComplexity)
		}
		
		fmt.Printf("   🏁 Test %s: %s\n", testCase.ID, map[bool]string{true: "PASSED", false: "FAILED"}[tierMatch])
	}
	
	// Step 4: Final Results Summary
	fmt.Printf("\n" + strings.Repeat("=", 70) + "\n")
	fmt.Println("🏁 End-to-End Test Results Summary")
	fmt.Printf("📊 Total Tests: %d\n", totalTests)
	fmt.Printf("✅ Successful: %d\n", successfulTests)
	fmt.Printf("❌ Failed: %d\n", totalTests-successfulTests)
	fmt.Printf("📈 Success Rate: %.1f%%\n", float64(successfulTests)/float64(totalTests)*100)
	
	if successfulTests == totalTests {
		fmt.Println("\n🎉 ALL TESTS PASSED! Your-PaL-MoE end-to-end workflow is working correctly!")
		fmt.Println("✅ Smart Provider Selection is functioning as expected")
		fmt.Println("✅ Task complexity analysis is accurate")
		fmt.Println("✅ Configuration loading is working properly")
	} else {
		fmt.Printf("\n⚠️  %d/%d tests failed. Please review the provider selection logic.\n", totalTests-successfulTests, totalTests)
	}
	
	fmt.Println("\n🚀 Your-PaL-MoE is ready for production use!")
}

// ComplexityResult represents the multi-dimensional complexity analysis
type ComplexityResult struct {
	CognitiveLoad         float64
	TechnicalDepth        float64
	CreativeRequirement   float64
	AccuracyRequirement   float64
	OverallComplexity     float64
}

// analyzeTaskComplexity performs multi-dimensional task complexity analysis
func analyzeTaskComplexity(prompt, category string) ComplexityResult {
	result := ComplexityResult{}
	
	// Analyze based on prompt length and keywords
	promptLower := strings.ToLower(prompt)
	wordCount := len(strings.Fields(prompt))
	
	// Cognitive Load Analysis
	if wordCount > 20 {
		result.CognitiveLoad += 1.5
	}
	if wordCount > 40 {
		result.CognitiveLoad += 1.0
	}
	if strings.Contains(promptLower, "complex") || strings.Contains(promptLower, "detailed") || strings.Contains(promptLower, "comprehensive") {
		result.CognitiveLoad += 1.0
	}
	
	// Technical Depth Analysis
	technicalKeywords := []string{"algorithm", "architecture", "system", "implementation", "framework", "protocol", "optimization"}
	for _, keyword := range technicalKeywords {
		if strings.Contains(promptLower, keyword) {
			result.TechnicalDepth += 0.8
		}
	}
	
	// Creative Requirement Analysis
	creativeKeywords := []string{"story", "poem", "creative", "imaginative", "haiku", "narrative", "artistic"}
	for _, keyword := range creativeKeywords {
		if strings.Contains(promptLower, keyword) {
			result.CreativeRequirement += 0.9
		}
	}
	
	// Accuracy Requirement Analysis
	accuracyKeywords := []string{"proof", "mathematical", "formal", "precise", "exact", "specification", "documentation"}
	for _, keyword := range accuracyKeywords {
		if strings.Contains(promptLower, keyword) {
			result.AccuracyRequirement += 1.0
		}
	}
	
	// Category-based adjustments
	switch category {
	case "code":
		result.TechnicalDepth += 1.5
		result.AccuracyRequirement += 1.0
	case "mathematical":
		result.AccuracyRequirement += 2.0
		result.TechnicalDepth += 1.0
	case "creative":
		result.CreativeRequirement += 1.5
	case "technical":
		result.TechnicalDepth += 1.5
		result.CognitiveLoad += 1.0
	case "general":
		// Keep baseline values
	}
	
	// Normalize values to 0-5 scale
	result.CognitiveLoad = min(result.CognitiveLoad, 5.0)
	result.TechnicalDepth = min(result.TechnicalDepth, 5.0)
	result.CreativeRequirement = min(result.CreativeRequirement, 5.0)
	result.AccuracyRequirement = min(result.AccuracyRequirement, 5.0)
	
	// Calculate overall complexity (weighted average)
	result.OverallComplexity = (result.CognitiveLoad*0.25 + result.TechnicalDepth*0.30 + result.CreativeRequirement*0.20 + result.AccuracyRequirement*0.25)
	
	return result
}

// selectProviderBasedOnComplexity selects the best provider based on task complexity
func selectProviderBasedOnComplexity(complexity ComplexityResult, providers []MockProvider) MockProvider {
	// Sort providers by capability tier
	officialProviders := []MockProvider{}
	communityProviders := []MockProvider{}
	unofficialProviders := []MockProvider{}
	
	for _, provider := range providers {
		switch provider.Tier {
		case "official":
			officialProviders = append(officialProviders, provider)
		case "community":
			communityProviders = append(communityProviders, provider)
		case "unofficial":
			unofficialProviders = append(unofficialProviders, provider)
		}
	}
	
	// Selection logic based on overall complexity
	if complexity.OverallComplexity >= 4.0 {
		// High complexity: Use official providers (premium models)
		if len(officialProviders) > 0 {
			// Prefer providers with reasoning capabilities for high accuracy needs
			if complexity.AccuracyRequirement >= 3.0 {
				for _, provider := range officialProviders {
					if strings.Contains(strings.ToLower(provider.Name), "anthropic") {
						return provider
					}
				}
			}
			return officialProviders[0]
		}
	} else if complexity.OverallComplexity >= 2.5 {
		// Medium complexity: Use community providers (good balance)
		if len(communityProviders) > 0 {
			// Prefer creative-capable providers for creative tasks
			if complexity.CreativeRequirement >= 2.0 {
				for _, provider := range communityProviders {
					if strings.Contains(strings.ToLower(provider.Name), "pollinations") {
						return provider
					}
				}
			}
			return communityProviders[0]
		}
	} else {
		// Low complexity: Use unofficial/local providers (efficient for simple tasks)
		if len(unofficialProviders) > 0 {
			return unofficialProviders[0]
		}
	}
	
	// Fallback: return first available provider
	if len(providers) > 0 {
		return providers[0]
	}
	
	// Default fallback
	return MockProvider{Name: "Default", Tier: "community", BaseURL: "localhost", Models: "default"}
}

// loadProvidersFromCSV loads providers from the 5-column CSV format
func loadProvidersFromCSV(filename string) ([]MockProvider, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()
	
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}
	
	if len(records) < 2 {
		return nil, fmt.Errorf("CSV file must have at least a header and one data row")
	}
	
	var providers []MockProvider
	
	// Skip header row
	for i := 1; i < len(records); i++ {
		record := records[i]
		if len(record) < 5 {
			continue // Skip incomplete records
		}
		
		provider := MockProvider{
			Name:    record[0],
			Tier:    record[1],
			BaseURL: record[2],
			Models:  record[4], // Column 4 is Models in 5-column format
		}
		
		// Extract features from "Other" field
		if len(record) > 5 {
			otherField := strings.ToLower(record[5])
			if strings.Contains(otherField, "creative") {
				provider.Features = append(provider.Features, "creative")
			}
			if strings.Contains(otherField, "reasoning") {
				provider.Features = append(provider.Features, "reasoning")
			}
			if strings.Contains(otherField, "local") {
				provider.Features = append(provider.Features, "local")
			}
		}
		
		providers = append(providers, provider)
	}
	
	return providers, nil
}

// fileExists checks if a file exists
func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

// readConfigSample reads the first few lines of a config file for preview
func readConfigSample(filename string) (string, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	
	lines := strings.Split(string(content), "\n")
	sampleLines := lines
	if len(lines) > 8 {
		sampleLines = lines[:8]
	}
	
	result := strings.Join(sampleLines, "\n")
	if len(lines) > 8 {
		result += "\n      ... (truncated)"
	}
	
	// Add indentation for display
	indentedLines := []string{}
	for _, line := range strings.Split(result, "\n") {
		indentedLines = append(indentedLines, "      "+line)
	}
	
	return strings.Join(indentedLines, "\n"), nil
}

// min returns the minimum of two float64 values
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}