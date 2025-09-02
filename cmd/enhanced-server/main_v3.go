package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/ThatsRight-ItsTJ/Your-PaL-MoE/internal/enhanced"
	"github.com/ThatsRight-ItsTJ/Your-PaL-MoE/pkg/selection"
)

var enhancedSystemV3 *enhanced.EnhancedSystemV3

func main() {
	log.Println("🚀 Starting Enhanced PaL-MoE Server v3.0.0 with Dynamic Model Discovery...")

	// Get CSV path from command line or use default
	csvPath := "providers.csv"
	if len(os.Args) > 1 {
		csvPath = os.Args[1]
	}

	// YAML directory (optional)
	yamlDir := "configs/yaml" // Can be empty if not using YAML configs

	// Initialize enhanced system v3 with dynamic model discovery
	var err error
	enhancedSystemV3, err = enhanced.NewEnhancedSystemV3(csvPath, yamlDir)
	if err != nil {
		log.Fatalf("❌ Failed to initialize enhanced system v3: %v", err)
	}

	providers := enhancedSystemV3.GetProviders()
	log.Printf("✅ Loaded %d providers with dynamic model discovery", len(providers))
	
	// Show provider information with model counts
	for _, provider := range providers {
		log.Printf("  📋 %s: %d models (%v)", provider.Name, len(provider.Models), provider.Models)
	}

	// Validate providers and show any issues
	if issues := enhancedSystemV3.ValidateProviders(); len(issues) > 0 {
		log.Println("⚠️  Provider validation issues found:")
		for providerName, providerIssues := range issues {
			for _, issue := range providerIssues {
				log.Printf("  ❗ %s: %s", providerName, issue)
			}
		}
	} else {
		log.Println("✅ All providers validated successfully")
	}

	// Show system information
	systemInfo := enhancedSystemV3.GetSystemInfo()
	if infoJSON, err := json.MarshalIndent(systemInfo, "", "  "); err == nil {
		log.Printf("📊 System Information:\n%s", string(infoJSON))
	}

	// Show provider capabilities
	capabilities := enhancedSystemV3.GetProviderCapabilities()
	log.Println("🎯 Provider Capabilities:")
	for name, cap := range capabilities {
		log.Printf("  📋 %s: Text=%t, Image=%t, Code=%t, Audio=%t, Video=%t, Multimodal=%t (R:%d, K:%d, C:%d)", 
			name, cap.Text, cap.Image, cap.Code, cap.Audio, cap.Video, cap.Multimodal,
			cap.Reasoning, cap.Knowledge, cap.Computation)
	}

	// Setup HTTP router
	router := mux.NewRouter()

	// API routes
	api := router.PathPrefix("/api/v1").Subrouter()
	
	// Enhanced processing endpoint
	api.HandleFunc("/process", handleProcessRequest).Methods("POST")
	api.HandleFunc("/analyze", handleAnalyzeRequest).Methods("POST")
	
	// Provider information endpoints
	api.HandleFunc("/providers", handleGetProviders).Methods("GET")
	api.HandleFunc("/providers/capabilities", handleGetProviderCapabilities).Methods("GET")
	api.HandleFunc("/providers/models", handleGetDetailedModelCapabilities).Methods("GET")
	
	// System endpoints
	api.HandleFunc("/system/info", handleGetSystemInfo).Methods("GET")
	api.HandleFunc("/system/refresh", handleRefreshAllModels).Methods("POST")
	api.HandleFunc("/system/weights", handleSetSelectionWeights).Methods("POST")
	
	// Dynamic model endpoints
	api.HandleFunc("/models/refresh", handleRefreshAllModels).Methods("POST")
	api.HandleFunc("/models/stats", handleGetDynamicModelStats).Methods("GET")

	// Health check
	api.HandleFunc("/health", handleHealth).Methods("GET")

	// Setup CORS
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"*"},
	})

	handler := c.Handler(router)

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🌐 Server starting on port %s", port)
	log.Printf("📡 API endpoints available at http://localhost:%s/api/v1/", port)
	log.Printf("🔍 Try: curl -X POST http://localhost:%s/api/v1/process -H 'Content-Type: application/json' -d '{\"request\":\"Write a Python function\"}'", port)
	log.Printf("🔄 Refresh models: curl -X POST http://localhost:%s/api/v1/models/refresh", port)

	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatalf("❌ Server failed to start: %v", err)
	}
}

// handleProcessRequest processes a request using enhanced provider selection
func handleProcessRequest(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Request string `json:"request"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Request == "" {
		http.Error(w, "Request field is required", http.StatusBadRequest)
		return
	}

	log.Printf("🔄 Processing request: %s", req.Request)

	response, err := enhancedSystemV3.ProcessRequest(req.Request)
	if err != nil {
		log.Printf("❌ Processing failed: %v", err)
		http.Error(w, fmt.Sprintf("Processing failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleAnalyzeRequest provides detailed analysis of provider selection
func handleAnalyzeRequest(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Request string `json:"request"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Request == "" {
		http.Error(w, "Request field is required", http.StatusBbadRequest)
		return
	}

	analysis := enhancedSystemV3.AnalyzeRequest(req.Request)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(analysis)
}

// handleGetProviders returns basic provider information
func handleGetProviders(w http.ResponseWriter, r *http.Request) {
	providers := make([]map[string]interface{}, 0)
	
	for _, provider := range enhancedSystemV3.GetProviders() {
		providers = append(providers, map[string]interface{}{
			"name":         provider.Name,
			"base_url":     provider.BaseURL,
			"models":       provider.Models,
			"model_count":  len(provider.Models),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"providers": providers,
		"count":     len(providers),
	})
}

// handleGetProviderCapabilities returns provider capabilities
func handleGetProviderCapabilities(w http.ResponseWriter, r *http.Request) {
	capabilities := enhancedSystemV3.GetProviderCapabilities()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(capabilities)
}

// handleGetDetailedModelCapabilities returns per-model capabilities
func handleGetDetailedModelCapabilities(w http.ResponseWriter, r *http.Request) {
	capabilities := enhancedSystemV3.GetDetailedModelCapabilities()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(capabilities)
}

// handleGetSystemInfo returns comprehensive system information
func handleGetSystemInfo(w http.ResponseWriter, r *http.Request) {
	info := enhancedSystemV3.GetSystemInfo()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

// handleRefreshAllModels refreshes models for all providers
func handleRefreshAllModels(w http.ResponseWriter, r *http.Request) {
	log.Println("🔄 Refreshing all provider models...")
	
	err := enhancedSystemV3.RefreshAllModels()
	if err != nil {
		log.Printf("❌ Failed to refresh models: %v", err)
		http.Error(w, fmt.Sprintf("Failed to refresh models: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "All provider models refreshed successfully",
		"info":    enhancedSystemV3.GetSystemInfo(),
	})
}

// handleGetDynamicModelStats returns dynamic model loading statistics
func handleGetDynamicModelStats(w http.ResponseWriter, r *http.Request) {
	analysis := enhancedSystemV3.AnalyzeRequest("stats") // Dummy request to get stats
	stats := analysis["dynamic_loading_stats"]

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// handleSetSelectionWeights updates provider selection weights
func handleSetSelectionWeights(w http.ResponseWriter, r *http.Request) {
	var weights selection.SelectionWeights

	if err := json.NewDecoder(r.Body).Decode(&weights); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	enhancedSystemV3.SetSelectionWeights(weights)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Selection weights updated",
		"weights": weights,
	})
}

// handleHealth returns server health status
func handleHealth(w http.ResponseWriter, r *http.Request) {
	systemInfo := enhancedSystemV3.GetSystemInfo()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":     "healthy",
		"version":    "3.0.0-dynamic-models",
		"providers":  len(enhancedSystemV3.GetProviders()),
		"system":     systemInfo,
		"timestamp":  "2024-01-01T00:00:00Z", // Placeholder
	})
}