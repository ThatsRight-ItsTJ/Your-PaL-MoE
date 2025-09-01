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

	"github.com/Your-PaL-MoE/internal/enhanced"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

func main() {
	// Initialize logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Initialize enhanced system
	providersFile := "providers.csv"
	if len(os.Args) > 1 {
		providersFile = os.Args[1]
	}

	system, err := enhanced.NewEnhancedSystem(logger, providersFile)
	if err != nil {
		logger.Fatalf("Failed to initialize enhanced system: %v", err)
	}

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

	// Start server
	srv := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	go func() {
		logger.Info("Starting Enhanced Your PaL MoE server on :8080")
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

	system.Shutdown()
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
		"version":   "enhanced-1.0.0",
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

	// Generate ID if not provided
	if input.ID == "" {
		input.ID = fmt.Sprintf("req_%d", time.Now().UnixNano())
	}

	// Set timestamp
	input.Timestamp = time.Now()

	h.logger.Infof("Processing request: %s", input.ID)

	// Process request
	result, err := h.system.ProcessRequest(r.Context(), input)
	if err != nil {
		h.logger.Errorf("Failed to process request %s: %v", input.ID, err)
		http.Error(w, fmt.Sprintf("Processing failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (h *HTTPServer) getRequestHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	requestID := vars["id"]

	request, err := h.system.GetRequest(requestID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Request not found: %v", err), http.StatusNotFound)
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
	metrics := h.system.GetMetrics()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}
func (h *HTTPServer) generateProviderYAMLHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	providerID := vars["id"]

	yaml, err := h.system.GenerateProviderYAML(r.Context(), providerID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to generate YAML: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/x-yaml")
	w.Write([]byte(yaml))
}

func (h *HTTPServer) generateAllYAMLsHandler(w http.ResponseWriter, r *http.Request) {
	yamls, err := h.system.GenerateAllProviderYAMLs(r.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to generate YAMLs: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"generated_count": len(yamls),
		"providers":       yamls,
		"timestamp":       time.Now().Unix(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
