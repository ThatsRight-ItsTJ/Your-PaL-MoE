package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"
)

// Simple test to verify Pollinations connectivity
func main() {
	fmt.Println("Testing Pollinations API connectivity...")
	
	// Test 1: Simple connectivity test
	testConnectivity()
	
	// Test 2: Simple text generation
	testSimpleGeneration()
	
	// Test 3: YAML generation test
	testYAMLGeneration()
}

func testConnectivity() {
	fmt.Println("\n=== Test 1: Connectivity ===")
	
	client := &http.Client{Timeout: 10 * time.Second}
	
	// Simple test prompt
	prompt := "Hello"
	encodedPrompt := url.QueryEscape(prompt)
	testURL := fmt.Sprintf("https://text.pollinations.ai/%s", encodedPrompt)
	
	fmt.Printf("Testing URL: %s\n", testURL)
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	req, err := http.NewRequestWithContext(ctx, "GET", testURL, nil)
	if err != nil {
		log.Printf("❌ Failed to create request: %v", err)
		return
	}
	
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("❌ Failed to make request: %v", err)
		return
	}
	defer resp.Body.Close()
	
	fmt.Printf("Status Code: %d\n", resp.StatusCode)
	fmt.Printf("Content-Type: %s\n", resp.Header.Get("Content-Type"))
	
	if resp.StatusCode == 200 {
		fmt.Println("✅ Pollinations API is reachable")
	} else {
		fmt.Printf("❌ Unexpected status code: %d\n", resp.StatusCode)
	}
}

func testSimpleGeneration() {
	fmt.Println("\n=== Test 2: Simple Generation ===")
	
	client := &http.Client{Timeout: 15 * time.Second}
	
	// Test simple text generation
	prompt := "Generate a simple greeting"
	encodedPrompt := url.QueryEscape(prompt)
	testURL := fmt.Sprintf("https://text.pollinations.ai/%s", encodedPrompt)
	
	fmt.Printf("Testing prompt: %s\n", prompt)
	
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	
	req, err := http.NewRequestWithContext(ctx, "GET", testURL, nil)
	if err != nil {
		log.Printf("❌ Failed to create request: %v", err)
		return
	}
	
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("❌ Failed to make request: %v", err)
		return
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == 200 {
		body := make([]byte, 1024)
		n, _ := resp.Body.Read(body)
		response := string(body[:n])
		
		fmt.Println("✅ Simple generation successful")
		fmt.Printf("Response: %s\n", response[:min(len(response), 200)])
	} else {
		fmt.Printf("❌ Generation failed with status: %d\n", resp.StatusCode)
	}
}

func testYAMLGeneration() {
	fmt.Println("\n=== Test 3: YAML Generation ===")
	
	client := &http.Client{Timeout: 30 * time.Second}
	
	// Test YAML generation with provider data
	prompt := `Generate a YAML configuration file for an AI provider with the following details:

Provider ID: test_provider
Provider Name: Test Provider
Tier: community
Endpoint: https://example.com/api
Model: test-model
Cost Per Token: 0.001000
Max Tokens: 2048
Capabilities: chat
Rate Limit: 100/min

Generate a complete YAML configuration file that includes:
1. Provider configuration section
2. API settings (endpoint, authentication)
3. Model parameters (temperature, max_tokens, etc.)
4. Rate limiting configuration if applicable
5. Cost tracking settings
6. Retry and timeout configurations

Format the output as valid YAML. Include comments explaining each section.
Only return the YAML content, no additional text or explanation.`

	encodedPrompt := url.QueryEscape(prompt)
	testURL := fmt.Sprintf("https://text.pollinations.ai/%s", encodedPrompt)
	
	fmt.Println("Testing YAML generation...")
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	req, err := http.NewRequestWithContext(ctx, "GET", testURL, nil)
	if err != nil {
		log.Printf("❌ Failed to create request: %v", err)
		return
	}
	
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("❌ Failed to make request: %v", err)
		return
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == 200 {
		body := make([]byte, 4096)
		n, _ := resp.Body.Read(body)
		response := string(body[:n])
		
		fmt.Println("✅ YAML generation successful")
		fmt.Println("Generated YAML:")
		fmt.Println("================")
		fmt.Printf("%s\n", response)
		fmt.Println("================")
		
		// Save to file for inspection
		if err := saveToFile("test_output.yaml", response); err != nil {
			log.Printf("Warning: Could not save to file: %v", err)
		} else {
			fmt.Println("✅ YAML saved to test_output.yaml")
		}
	} else {
		fmt.Printf("❌ YAML generation failed with status: %d\n", resp.StatusCode)
	}
}

func saveToFile(filename, content string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	
	_, err = file.WriteString(content)
	return err
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}