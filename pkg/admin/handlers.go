package admin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/ThatsRight-ItsTJ/Your-PaL-MoE/pkg/analytics"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// AdminHandlers provides HTTP handlers for admin functionality
type AdminHandlers struct {
	logger          *logrus.Logger
	analyticsEngine *analytics.AnalyticsEngine
}

// NewAdminHandlers creates a new AdminHandlers instance
func NewAdminHandlers(logger *logrus.Logger, analyticsEngine *analytics.AnalyticsEngine) *AdminHandlers {
	return &AdminHandlers{
		logger:          logger,
		analyticsEngine: analyticsEngine,
	}
}

// GetSystemMetrics returns system performance metrics
func (ah *AdminHandlers) GetSystemMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := ah.analyticsEngine.GetSystemMetrics()
	
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(metrics); err != nil {
		ah.logger.Errorf("Failed to encode metrics: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// GetProviderMetrics returns provider-specific metrics
func (ah *AdminHandlers) GetProviderMetrics(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	providerID := vars["id"]
	
	if providerID == "" {
		http.Error(w, "Provider ID is required", http.StatusBadRequest)
		return
	}

	metrics := ah.analyticsEngine.GetProviderMetrics(providerID)
	
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(metrics); err != nil {
		ah.logger.Errorf("Failed to encode provider metrics: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// GetCostAnalysis returns cost analysis data
func (ah *AdminHandlers) GetCostAnalysis(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters for time range
	query := r.URL.Query()
	hoursStr := query.Get("hours")
	hours := 24 // default to 24 hours
	
	if hoursStr != "" {
		if h, err := strconv.Atoi(hoursStr); err == nil && h > 0 {
			hours = h
		}
	}

	since := time.Now().Add(-time.Duration(hours) * time.Hour)
	costAnalysis := ah.analyticsEngine.GetCostAnalysis(since)
	
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(costAnalysis); err != nil {
		ah.logger.Errorf("Failed to encode cost analysis: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// GetProviderPerformance returns provider performance analysis
func (ah *AdminHandlers) GetProviderPerformance(w http.ResponseWriter, r *http.Request) {
	performance := ah.analyticsEngine.GetProviderPerformance()
	
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(performance); err != nil {
		ah.logger.Errorf("Failed to encode provider performance: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// GetOptimizationInsights returns optimization recommendations
func (ah *AdminHandlers) GetOptimizationInsights(w http.ResponseWriter, r *http.Request) {
	insights, err := ah.analyticsEngine.GenerateInsights()
	if err != nil {
		ah.logger.Errorf("Failed to generate insights: %v", err)
		http.Error(w, fmt.Sprintf("Failed to generate insights: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(insights); err != nil {
		ah.logger.Errorf("Failed to encode insights: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// GetHealthStatus returns overall system health
func (ah *AdminHandlers) GetHealthStatus(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"version":   "1.0.0",
		"uptime":    time.Since(time.Now().Add(-24 * time.Hour)).String(), // placeholder
	}
	
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(health); err != nil {
		ah.logger.Errorf("Failed to encode health status: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// RegisterRoutes registers all admin routes
func (ah *AdminHandlers) RegisterRoutes(router *mux.Router) {
	adminRouter := router.PathPrefix("/admin").Subrouter()
	
	adminRouter.HandleFunc("/metrics/system", ah.GetSystemMetrics).Methods("GET")
	adminRouter.HandleFunc("/metrics/provider/{id}", ah.GetProviderMetrics).Methods("GET")
	adminRouter.HandleFunc("/analytics/cost", ah.GetCostAnalysis).Methods("GET")
	adminRouter.HandleFunc("/analytics/performance", ah.GetProviderPerformance).Methods("GET")
	adminRouter.HandleFunc("/insights", ah.GetOptimizationInsights).Methods("GET")
	adminRouter.HandleFunc("/health", ah.GetHealthStatus).Methods("GET")
}