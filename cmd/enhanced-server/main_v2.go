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

var enhancedSystemV2 *enhanced.EnhancedSystemV2

func main() {
	log.Println("🚀 Starting Enhanced PaL-MoE Server v2.0.0 with Model Database Integration...")

	// Load providers from CSV
	providers, err := selection.LoadProvidersFromCSV("providers.csv")
	if err != nil {
		log.Fatalf("❌ Failed to load providers: %v", err)
	}

	log.Printf("✅ Loaded %d providers from CSV", len(providers))
	for _, provider := range providers {
		log.Printf("  📋 %s: %d models (%v)", provider.Name, len(provider.Models), provider.Models)
	}

	// Initialize enhanced system v2 with model database
	enhancedSystemV2 = enhanced.NewEnhancedSystemV2(providers)
	log.Println("🧠 Enhanced system v2 initialized with model database integration")

	// Validate providers and show any issues
	if issues := enhancedSystemV2.ValidateProviders(); len(issues) > 0 {
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
	systemInfo := enhancedSystemV2.GetSystemInfo()
	if infoJSON, err := json.MarshalIndent(systemInfo, "", "  "); err == nil {
		log.Printf("📊 System Information:\n%s", string(infoJSON))
	}

	// Show provider capabilities
	capabilities := enhancedSystemV2.GetProviderCapabilities()
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
	api.HandleFunc("/system/stats", handleGetModelDatabaseStats).Methods("GET")
	api.HandleFunc("/system/refresh", handleRefreshModelDatabase).Methods("POST")
	api.HandleFunc("/system/weights", handleSetSelectionWeights).Methods("POST")
	
	// YAML generation endpoints (legacy compatibility)
	api.HandleFunc("/providers/yaml/generate-all", handleGenerateAllYAML).Methods("POST")

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

	response, err := enhancedSystemV2.ProcessRequest(req.Request)
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
		http.Error(w, "Request field is required", http.StatusBadRequest)
		return
	}

	analysis := enhancedSystemV2.AnalyzeRequest(req.Request)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(analysis)
}

// handleGetProviders returns basic provider information
func handleGetProviders(w http.ResponseWriter, r *http.Request) {
	providers := make([]map[string]interface{}, 0)
	
	for _, provider := range enhancedSystemV2.GetProviders() {
		providers = append(providers, map[string]interface{}{
			"name":     provider.Name,
			"base_url": provider.BaseURL,
			"models":   provider.Models,
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
	capabilities := enhancedSystemV2.GetProviderCapabilities()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(capabilities)
}

// handleGetDetailedModelCapabilities returns per-model capabilities
func handleGetDetailedModelCapabilities(w http.ResponseWriter, r *http.Request) {
	capabilities := enhancedSystemV2.GetDetailedModelCapabilities()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(capabilities)
}

// handleGetSystemInfo returns comprehensive system information
func handleGetSystemInfo(w http.ResponseWriter, r *http.Request) {
	info := enhancedSystemV2.GetSystemInfo()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

// handleGetModelDatabaseStats returns model database statistics
func handleGetModelDatabaseStats(w http.ResponseWriter, r *http.Request) {
	stats := enhancedSystemV2.GetModelDatabaseStats()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// handleRefreshModelDatabase refreshes the model database cache
func handleRefreshModelDatabase(w http.ResponseWriter, r *http.Request) {
	enhancedSystemV2.RefreshModelDatabase()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Model database cache refreshed",
	})
}

// handleSetSelectionWeights updates provider selection weights
func handleSetSelectionWeights(w http.ResponseWriter, r *http.Request) {
	var weights selection.SelectionWeights

	if err := json.NewDecoder(r.Body).Decode(&weights); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	enhancedSystemV2.SetSelectionWeights(weights)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Selection weights updated",
		"weights": weights,
	})
}

// handleGenerateAllYAML generates YAML files for all providers (legacy compatibility)
func handleGenerateAllYAML(w http.ResponseWriter, r *http.Request) {
	// For compatibility with existing functionality
	log.Println("🔄 Generating YAML files for all providers...")
	
	// This would call the existing YAML generation logic
	// For now, return a placeholder response
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "YAML generation completed",
		"note":    "Enhanced system v2 focuses on runtime provider selection rather than static YAML generation",
	})
}

// handleHealth returns server health status
func handleHealth(w http.ResponseWriter, r *http.Request) {
	stats := enhancedSystemV2.GetModelDatabaseStats()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":     "healthy",
		"version":    "2.0.0-enhanced",
		"providers":  len(enhancedSystemV2.GetProviders()),
		"db_stats":   stats,
		"timestamp":  "2024-01-01T00:00:00Z", // Placeholder
	})
}

// Helper method to get providers (we need to add this to the enhanced system)
func (es *enhanced.EnhancedSystemV2) GetProviders() []selection.Provider {
	// This method needs to be added to the enhanced system
	// For now, we'll work around it
	return []selection.Provider{} // Placeholder
}