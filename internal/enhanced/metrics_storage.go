package enhanced

import (
	"database/sql"
	"fmt"
	"time"
	
	_ "github.com/mattn/go-sqlite3"
)

// MetricsStorage handles persistent storage of cost-aware provider metrics
type MetricsStorage struct {
	db     *sql.DB
	dbPath string
}

// NewMetricsStorage creates a new cost-focused metrics storage instance
func NewMetricsStorage(dbPath string) (*MetricsStorage, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	
	storage := &MetricsStorage{
		db:     db,
		dbPath: dbPath,
	}
	
	if err := storage.initializeTables(); err != nil {
		return nil, fmt.Errorf("failed to initialize tables: %w", err)
	}
	
	return storage, nil
}

// initializeTables creates the necessary tables with cost tracking
func (m *MetricsStorage) initializeTables() error {
	schema := `
	CREATE TABLE IF NOT EXISTS provider_metrics (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		provider_name TEXT NOT NULL,
		model TEXT NOT NULL,
		timestamp DATETIME NOT NULL,
		request_count INTEGER NOT NULL DEFAULT 0,
		failure_count INTEGER NOT NULL DEFAULT 0,
		latency_ms REAL NOT NULL DEFAULT 0,
		tokens_used INTEGER NOT NULL DEFAULT 0,
		cost REAL NOT NULL DEFAULT 0,
		cost_per_token REAL NOT NULL DEFAULT 0,
		rate_limited BOOLEAN NOT NULL DEFAULT 0,
		success_rate REAL NOT NULL DEFAULT 0,
		quality_score REAL NOT NULL DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	
	CREATE TABLE IF NOT EXISTS rate_limit_status (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		provider_name TEXT NOT NULL,
		model TEXT NOT NULL,
		requests_per_minute INTEGER NOT NULL DEFAULT 0,
		requests_remaining INTEGER NOT NULL DEFAULT 0,
		tokens_per_minute INTEGER NOT NULL DEFAULT 0,
		tokens_remaining INTEGER NOT NULL DEFAULT 0,
		reset_time DATETIME NOT NULL,
		last_updated DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(provider_name, model)
	);
	
	CREATE TABLE IF NOT EXISTS cost_optimization_log (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME NOT NULL,
		original_provider TEXT NOT NULL,
		selected_provider TEXT NOT NULL,
		model TEXT NOT NULL,
		estimated_cost_original REAL NOT NULL,
		actual_cost_selected REAL NOT NULL,
		cost_savings REAL NOT NULL,
		tokens_used INTEGER NOT NULL,
		task_complexity TEXT NOT NULL,
		selection_reason TEXT NOT NULL
	);
	
	CREATE INDEX IF NOT EXISTS idx_provider_metrics_lookup 
	ON provider_metrics(provider_name, model, timestamp);
	
	CREATE INDEX IF NOT EXISTS idx_rate_limit_lookup 
	ON rate_limit_status(provider_name, model);
	`
	
	_, err := m.db.Exec(schema)
	return err
}

// RecordProviderMetrics stores cost-aware metrics entry
func (m *MetricsStorage) RecordProviderMetrics(
	providerName, model string,
	requestCount, failureCount, tokensUsed int64,
	latency, cost float64,
	rateLimited bool,
) error {
	successRate := 1.0
	if requestCount > 0 {
		successRate = float64(requestCount-failureCount) / float64(requestCount)
	}
	
	costPerToken := 0.0
	if tokensUsed > 0 {
		costPerToken = cost / float64(tokensUsed)
	}
	
	query := `
		INSERT INTO provider_metrics 
		(provider_name, model, timestamp, request_count, failure_count, 
		 latency_ms, tokens_used, cost, cost_per_token, rate_limited, success_rate)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	_, err := m.db.Exec(query, 
		providerName, model, time.Now(),
		requestCount, failureCount, latency, tokensUsed, cost, costPerToken, rateLimited, successRate)
	
	return err
}

// UpdateRateLimitStatus stores current rate limit information
func (m *MetricsStorage) UpdateRateLimitStatus(
	providerName, model string,
	requestsPerMin, requestsRemaining, tokensPerMin, tokensRemaining int64,
	resetTime time.Time,
) error {
	query := `
		INSERT OR REPLACE INTO rate_limit_status 
		(provider_name, model, requests_per_minute, requests_remaining,
		 tokens_per_minute, tokens_remaining, reset_time, last_updated)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	_, err := m.db.Exec(query,
		providerName, model, requestsPerMin, requestsRemaining,
		tokensPerMin, tokensRemaining, resetTime, time.Now())
	
	return err
}

// GetCostAnalysis provides cost efficiency analysis for provider selection
func (m *MetricsStorage) GetCostAnalysis(
	providerName, model string,
	window time.Duration,
) (*CostAnalysis, error) {
	cutoffTime := time.Now().Add(-window)
	
	query := `
		SELECT 
			AVG(cost_per_token) as avg_cost_per_token,
			SUM(cost) as total_cost,
			SUM(tokens_used) as total_tokens,
			AVG(success_rate) as avg_success_rate,
			COUNT(CASE WHEN rate_limited = 1 THEN 1 END) as rate_limit_hits,
			COUNT(*) as total_records
		FROM provider_metrics
		WHERE provider_name = ? AND model = ? AND timestamp >= ?
	`
	
	var analysis CostAnalysis
	row := m.db.QueryRow(query, providerName, model, cutoffTime)
	
	err := row.Scan(
		&analysis.AvgCostPerToken,
		&analysis.TotalCost,
		&analysis.TotalTokens,
		&analysis.AvgSuccessRate,
		&analysis.RateLimitHits,
		&analysis.TotalRecords,
	)
	
	if err != nil {
		return nil, err
	}
	
	analysis.ProviderName = providerName
	analysis.Model = model
	
	return &analysis, nil
}

// RecordCostOptimization logs cost optimization decisions
func (m *MetricsStorage) RecordCostOptimization(
	originalProvider, selectedProvider, model string,
	estimatedCostOriginal, actualCostSelected float64,
	tokensUsed int64,
	taskComplexity string,
	reason string,
) error {
	costSavings := estimatedCostOriginal - actualCostSelected
	
	query := `
		INSERT INTO cost_optimization_log
		(timestamp, original_provider, selected_provider, model,
		 estimated_cost_original, actual_cost_selected, cost_savings,
		 tokens_used, task_complexity, selection_reason)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	_, err := m.db.Exec(query,
		time.Now(), originalProvider, selectedProvider, model,
		estimatedCostOriginal, actualCostSelected, costSavings,
		tokensUsed, taskComplexity, reason)
	
	return err
}

// GetCostSavingsReport provides cost optimization analytics
func (m *MetricsStorage) GetCostSavingsReport(days int) (*CostSavingsReport, error) {
	cutoffDate := time.Now().AddDate(0, 0, -days)
	
	query := `
		SELECT 
			COUNT(*) as total_optimizations,
			SUM(cost_savings) as total_savings,
			AVG(cost_savings) as avg_savings_per_request,
			SUM(estimated_cost_original) as total_original_cost,
			SUM(actual_cost_selected) as total_actual_cost,
			(SUM(cost_savings) / SUM(estimated_cost_original)) * 100 as savings_percentage
		FROM cost_optimization_log
		WHERE timestamp >= ?
	`
	
	var report CostSavingsReport
	row := m.db.QueryRow(query, cutoffDate)
	
	err := row.Scan(
		&report.TotalOptimizations,
		&report.TotalSavings,
		&report.AvgSavingsPerRequest,
		&report.TotalOriginalCost,
		&report.TotalActualCost,
		&report.SavingsPercentage,
	)
	
	return &report, err
}

// Close closes the database connection
func (m *MetricsStorage) Close() error {
	if m.db != nil {
		return m.db.Close()
	}
	return nil
}

// CostAnalysis represents cost efficiency metrics
type CostAnalysis struct {
	ProviderName     string  `json:"provider_name"`
	Model            string  `json:"model"`
	AvgCostPerToken  float64 `json:"avg_cost_per_token"`
	TotalCost        float64 `json:"total_cost"`
	TotalTokens      int64   `json:"total_tokens"`
	AvgSuccessRate   float64 `json:"avg_success_rate"`
	RateLimitHits    int64   `json:"rate_limit_hits"`
	TotalRecords     int64   `json:"total_records"`
}

// CostSavingsReport represents cost optimization performance
type CostSavingsReport struct {
	TotalOptimizations    int64   `json:"total_optimizations"`
	TotalSavings         float64 `json:"total_savings"`
	AvgSavingsPerRequest float64 `json:"avg_savings_per_request"`
	TotalOriginalCost    float64 `json:"total_original_cost"`
	TotalActualCost      float64 `json:"total_actual_cost"`
	SavingsPercentage    float64 `json:"savings_percentage"`
}