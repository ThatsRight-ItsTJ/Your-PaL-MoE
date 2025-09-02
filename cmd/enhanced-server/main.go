package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ThatsRight-ItsTJ/Your-PaL-MoE/internal/enhanced"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

func main() {
	// Initialize logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Create some default providers for demonstration
	providers := []*enhanced.Provider{
		{
			Name:         "OpenAI",
			BaseURL:      "https://api.openai.com/v1",
			Models:       []string{"gpt-4", "gpt-3.5-turbo"},
			Tier:         enhanced.OfficialTier,
			MaxTokens:    4096,
			CostPerToken: 0.00003,
			Capabilities: []string{"reasoning", "creative", "mathematical"},
			RateLimits:   map[string]int64{"requests_per_minute": 60},
			Metadata:     make(map[string]interface{}),
			LastUpdated:  time.Now(),
		},
		{
			Name:         "Anthropic",
			BaseURL:      "https://api.anthropic.com/v1",
			Models:       []string{"claude-3-opus", "claude-3-sonnet"},
			Tier:         enhanced.OfficialTier,
			MaxTokens:    8192,
			CostPerToken: 0.000015,
			Capabilities: []string{"reasoning", "creative", "factual"},
			RateLimits:   map[string]int64{"requests_per_minute": 50},
			Metadata:     make(map[string]interface{}),
			LastUpdated:  time.Now(),
		},
	}

	// Initialize enhanced system
	system := enhanced.NewEnhancedSystem(providers)
	logger.Info("Enhanced system initialized successfully")

	// Create HTTP server
	server := &HTTPServer{
		system: system,
		logger: logger,
	}

	// Setup routes
	router := mux.NewRouter()
	router.HandleFunc("/health", server.healthHandler).Methods("GET")
	router.HandleFunc("/api/v1/process", server.processHandler).Methods("POST")
	router.HandleFunc("/api/v1/requests/{id}", server.getRequestHandler).Methods("GET")
	router.HandleFunc("/api/v1/providers", server.getProvidersHandler).Methods("GET")
	router.HandleFunc("/api/v1/providers/{id}/yaml", server.generateProviderYAMLHandler).Methods("GET")
	router.HandleFunc("/api/v1/providers/yaml/generate-all", server.generateAllYAMLsHandler).Methods("POST")
	router.HandleFunc("/api/v1/metrics", server.getMetricsHandler).Methods("GET")

	// Get port from environment variable, default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port

	// Start server
	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	go func() {
		logger.Infof("Starting Enhanced Your PaL MoE server on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatalf("Server forced to shutdown: %v", err)
	}

	logger.Info("Server exited")
}

// HTTPServer handles HTTP requests
type HTTPServer struct {
	system *enhanced.EnhancedSystem
	logger *logrus.Logger
}

func (h *HTTPServer) healthHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"version":   "enhanced-2.0.0",
		"features":  []string{"complexity-analysis", "provider-selection", "prompt-optimization"},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *HTTPServer) processHandler(w http.ResponseWriter, r *http.Request) {
	var input enhanced.RequestInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	h.logger.Infof("Processing request: %s", input.Content)

	// Process request with enhanced system
	result, err := h.system.ProcessRequest(r.Context(), input)
	if err != nil {
		h.logger.Errorf("Failed to process request: %v", err)
		http.Error(w, fmt.Sprintf("Processing failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (h *HTTPServer) getRequestHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	requestID := vars["id"]

	// Create a dummy RequestInput for demonstration
	input := enhanced.RequestInput{
		Content: fmt.Sprintf("Request ID: %s", requestID),
	}

	request, err := h.system.GetRequest(r.Context(), input)
	if err != nil {
		http.Error(w, fmt.Sprintf("Request processing failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(request)
}

func (h *HTTPServer) getProvidersHandler(w http.ResponseWriter, r *http.Request) {
	providers := h.system.GetProviders()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(providers)
}

func (h *HTTPServer) getMetricsHandler(w http.ResponseWriter, r *http.Request) {
	// Return dummy metrics for now
	metrics := map[string]interface{}{
		"total_requests":     100,
		"successful_requests": 95,
		"failed_requests":    5,
		"average_latency":    "150ms",
		"providers_active":   len(h.system.GetProviders()),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

func (h *HTTPServer) generateProviderYAMLHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	providerID := vars["id"]

	yaml, err := h.system.GenerateProviderYAML(providerID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to generate YAML: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/x-yaml")
	w.Write([]byte(yaml))
}

func (h *HTTPServer) generateAllYAMLsHandler(w http.ResponseWriter, r *http.Request) {
	yaml, err := h.system.GenerateAllProviderYAMLs()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to generate YAMLs: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"yaml":      yaml,
		"timestamp": time.Now().Unix(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}