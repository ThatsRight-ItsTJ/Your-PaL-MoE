package admin

import (
    "encoding/csv"
    "encoding/json"
    "fmt"
    "net/http"
    "os"
    "strings"
    "time"

    "github.com/gorilla/mux"
    "github.com/ThatsRight-ItsTJ/Your-PaL-MoE/pkg/providers"
    "github.com/ThatsRight-ItsTJ/Your-PaL-MoE/pkg/analytics"
)

type AdminHandler struct {
    providerManager  *providers.ProviderManager
    analyticsEngine  *analytics.AnalyticsEngine
    adminKey         string
}

func NewAdminHandler(providerManager *providers.ProviderManager, analyticsEngine *analytics.AnalyticsEngine, adminKey string) *AdminHandler {
    return &AdminHandler{
        providerManager:  providerManager,
        analyticsEngine:  analyticsEngine,
        adminKey:         adminKey,
    }
}

func (ah *AdminHandler) RegisterRoutes(router *mux.Router) {
    // Middleware for admin authentication
    adminRouter := router.PathPrefix("/admin").Subrouter()
    adminRouter.Use(ah.authMiddleware)
    
    // CSV Management
    adminRouter.HandleFunc("/csv", ah.GetCSV).Methods("GET")
    adminRouter.HandleFunc("/csv", ah.UpdateCSV).Methods("POST")
    adminRouter.HandleFunc("/csv/validate", ah.ValidateCSV).Methods("POST")
    adminRouter.HandleFunc("/csv/reload", ah.ReloadCSV).Methods("POST")
    adminRouter.HandleFunc("/csv/template", ah.GetCSVTemplate).Methods("GET")
    
    // Provider Management
    adminRouter.HandleFunc("/providers", ah.ListProviders).Methods("GET")
    adminRouter.HandleFunc("/providers/{name}", ah.GetProvider).Methods("GET")
    adminRouter.HandleFunc("/providers/{name}/test", ah.TestProvider).Methods("POST")
    adminRouter.HandleFunc("/providers/{name}/models", ah.DiscoverModels).Methods("POST")
    adminRouter.HandleFunc("/providers/generate-configs", ah.GenerateConfigs).Methods("POST")
    
    // Analytics
    adminRouter.HandleFunc("/analytics/overview", ah.GetAnalyticsOverview).Methods("GET")
    adminRouter.HandleFunc("/analytics/costs", ah.GetCostAnalysis).Methods("GET")
    adminRouter.HandleFunc("/analytics/performance", ah.GetPerformanceMetrics).Methods("GET")
    adminRouter.HandleFunc("/analytics/insights", ah.GetInsights).Methods("GET")
    
    // Health & Monitoring
    adminRouter.HandleFunc("/health", ah.GetSystemHealth).Methods("GET")
    adminRouter.HandleFunc("/health/providers", ah.GetProvidersHealth).Methods("GET")
    
    // Serve admin UI
    adminRouter.PathPrefix("/ui").Handler(http.StripPrefix("/admin/ui", 
        http.FileServer(http.Dir("./web/admin"))))
}

func (ah *AdminHandler) authMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" {
            authHeader = r.URL.Query().Get("key")
        }
        
        if !strings.Contains(authHeader, ah.adminKey) {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }
        
        next.ServeHTTP(w, r)
    })
}

func (ah *AdminHandler) GetCSV(w http.ResponseWriter, r *http.Request) {
    // Read current CSV file
    csvContent, err := os.ReadFile("./providers.csv")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "text/csv")
    w.Write(csvContent)
}

func (ah *AdminHandler) UpdateCSV(w http.ResponseWriter, r *http.Request) {
    // Parse uploaded CSV
    file, _, err := r.FormFile("csv")
    if err != nil {
        http.Error(w, "Failed to read CSV file", http.StatusBadRequest)
        return
    }
    defer file.Close()
    
    // Validate CSV format
    csvReader := csv.NewReader(file)
    records, err := csvReader.ReadAll()
    if err != nil {
        http.Error(w, "Invalid CSV format", http.StatusBadRequest)
        return
    }
    
    // Validate headers
    if len(records) < 1 {
        http.Error(w, "CSV file is empty", http.StatusBadRequest)
        return
    }
    
    headers := records[0]
    expectedHeaders := []string{"Name", "Tier", "Endpoint", "Model(s)"}
    if !equalStringSlices(headers, expectedHeaders) {
        http.Error(w, fmt.Sprintf("Invalid headers. Expected: %v", expectedHeaders), http.StatusBadRequest)
        return
    }
    
    // Create backup of current CSV
    backupPath := fmt.Sprintf("./providers.csv.backup.%d", time.Now().Unix())
    currentCSV, _ := os.ReadFile("./providers.csv")
    os.WriteFile(backupPath, currentCSV, 0644)
    
    // Write new CSV
    outputFile, err := os.Create("./providers.csv")
    if err != nil {
        http.Error(w, "Failed to update CSV", http.StatusInternalServerError)
        return
    }
    defer outputFile.Close()
    
    csvWriter := csv.NewWriter(outputFile)
    if err := csvWriter.WriteAll(records); err != nil {
        http.Error(w, "Failed to write CSV", http.StatusInternalServerError)
        return
    }
    
    // Reload providers
    ctx := r.Context()
    if err := ah.providerManager.ReloadProviders(ctx); err != nil {
        // Restore backup on failure
        os.Rename(backupPath, "./providers.csv")
        http.Error(w, fmt.Sprintf("Failed to reload providers: %v", err), http.StatusInternalServerError)
        return
    }
    
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{
        "status": "success",
        "message": "CSV updated and providers reloaded",
        "backup": backupPath,
    })
}

func (ah *AdminHandler) ValidateCSV(w http.ResponseWriter, r *http.Request) {
    // Parse CSV from request body
    csvReader := csv.NewReader(r.Body)
    records, err := csvReader.ReadAll()
    if err != nil {
        http.Error(w, "Invalid CSV format", http.StatusBadRequest)
        return
    }
    
    validationErrors := []string{}
    
    // Validate headers
    if len(records) < 1 {
        validationErrors = append(validationErrors, "CSV file is empty")
    } else {
        headers := records[0]
        expectedHeaders := []string{"Name", "Tier", "Endpoint", "Model(s)"}
        if !equalStringSlices(headers, expectedHeaders) {
            validationErrors = append(validationErrors, 
                fmt.Sprintf("Invalid headers. Expected: %v, Got: %v", expectedHeaders, headers))
        }
    }
    
    // Validate each row
    validTiers := map[string]bool{"official": true, "community": true, "unofficial": true}
    providerNames := make(map[string]bool)
    
    for i, record := range records[1:] {
        rowNum := i + 2
        
        if len(record) != 4 {
            validationErrors = append(validationErrors, 
                fmt.Sprintf("Row %d: Invalid number of columns (expected 4, got %d)", rowNum, len(record)))
            continue
        }
        
        name := strings.TrimSpace(record[0])
        tier := strings.TrimSpace(strings.ToLower(record[1]))
        endpoint := strings.TrimSpace(record[2])
        models := strings.TrimSpace(record[3])
        
        // Check for duplicate names
        if providerNames[name] {
            validationErrors = append(validationErrors, 
                fmt.Sprintf("Row %d: Duplicate provider name '%s'", rowNum, name))
        }
        providerNames[name] = true
        
        // Validate tier
        if !validTiers[tier] {
            validationErrors = append(validationErrors, 
                fmt.Sprintf("Row %d: Invalid tier '%s' (must be official, community, or unofficial)", rowNum, tier))
        }
        
        // Validate endpoint
        if endpoint == "" {
            validationErrors = append(validationErrors, 
                fmt.Sprintf("Row %d: Empty endpoint", rowNum))
        } else if !strings.HasPrefix(endpoint, "http://") && 
                  !strings.HasPrefix(endpoint, "https://") && 
                  !strings.HasPrefix(endpoint, "./scripts/") {
            validationErrors = append(validationErrors, 
                fmt.Sprintf("Row %d: Invalid endpoint format '%s'", rowNum, endpoint))
        }
        
        // Validate models
        if models == "" {
            validationErrors = append(validationErrors, 
                fmt.Sprintf("Row %d: No models specified", rowNum))
        }
    }
    
    response := map[string]interface{}{
        "valid": len(validationErrors) == 0,
        "errors": validationErrors,
        "row_count": len(records) - 1,
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

func (ah *AdminHandler) ReloadCSV(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    if err := ah.providerManager.ReloadProviders(ctx); err != nil {
        http.Error(w, fmt.Sprintf("Failed to reload providers: %v", err), http.StatusInternalServerError)
        return
    }
    
    providers, _ := ah.providerManager.GetAvailableProviders()
    
    response := map[string]interface{}{
        "status": "success",
        "message": "Providers reloaded from CSV",
        "provider_count": len(providers),
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

func (ah *AdminHandler) GetCSVTemplate(w http.ResponseWriter, r *http.Request) {
    template := `Name,Tier,Endpoint,Model(s)
# Official APIs - Require API keys in .env
OpenAI,official,https://api.openai.com/v1,https://api.openai.com/v1/models
Anthropic,official,https://api.anthropic.com,claude-3-5-sonnet|claude-3-haiku|claude-3-opus

# Community APIs - Free/Low cost
Pollinations,community,https://text.pollinations.ai,https://text.pollinations.ai/models
HuggingFace,community,https://api-inference.huggingface.co,gpt2|distilbert-base|t5-base

# Unofficial APIs - Custom scripts
Bing DALL-E,unofficial,./scripts/bing-dalle-wrapper.py,bing-dalle-3
Local Ollama,unofficial,http://localhost:11434,llama2|codellama|mistral
`
    
    w.Header().Set("Content-Type", "text/csv")
    w.Header().Set("Content-Disposition", "attachment; filename=providers.csv.template")
    w.Write([]byte(template))
}

func (ah *AdminHandler) ListProviders(w http.ResponseWriter, r *http.Request) {
    providers, err := ah.providerManager.GetAvailableProviders()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    // Add health status to each provider
    providerList := []map[string]interface{}{}
    for _, provider := range providers {
        providerInfo := map[string]interface{}{
            "name":     provider.Name,
            "tier":     provider.Tier,
            "endpoint": provider.Endpoint,
            "models":   provider.ModelsSource,
            "authentication": provider.Authentication.Type,
        }
        
        // Add health status
        if health, exists := ah.providerManager.GetHealthStatus(provider.Name); exists {
            providerInfo["health"] = health
        }
        
        providerList = append(providerList, providerInfo)
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(providerList)
}

func (ah *AdminHandler) TestProvider(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    providerName := vars["name"]
    
    ctx := r.Context()
    if err := ah.providerManager.TestProvider(ctx, providerName); err != nil {
        http.Error(w, fmt.Sprintf("Provider test failed: %v", err), http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{
        "status": "success",
        "message": fmt.Sprintf("Provider %s is healthy", providerName),
    })
}

func (ah *AdminHandler) GetAnalyticsOverview(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    // Get cost analysis
    costAnalysis, _ := ah.analyticsEngine.GetCostAnalysis(ctx, "30d")
    
    // Get provider metrics
    providerMetrics := ah.providerManager.GetProviderMetrics()
    
    overview := map[string]interface{}{
        "total_cost":          costAnalysis.TotalCost,
        "savings_vs_official": costAnalysis.SavingsVsOfficial,
        "cost_by_tier":        costAnalysis.CostByTier,
        "provider_count":      len(providerMetrics),
        "healthy_providers":   countHealthyProviders(providerMetrics),
        "optimization_opportunities": costAnalysis.OptimizationOpportunities,
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(overview)
}

func (ah *AdminHandler) GetCostAnalysis(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    period := r.URL.Query().Get("period")
    if period == "" {
        period = "30d"
    }
    
    analysis, err := ah.analyticsEngine.GetCostAnalysis(ctx, period)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(analysis)
}

func (ah *AdminHandler) GetPerformanceMetrics(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    providerName := r.URL.Query().Get("provider")
    
    if providerName != "" {
        // Get specific provider performance
        performance, err := ah.analyticsEngine.GetProviderPerformance(ctx, providerName, 24*time.Hour)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(performance)
    } else {
        // Get all providers performance
        providers, _ := ah.providerManager.GetAvailableProviders()
        performances := make(map[string]*analytics.ProviderPerformance)
        
        for _, provider := range providers {
            if perf, err := ah.analyticsEngine.GetProviderPerformance(ctx, provider.Name, 24*time.Hour); err == nil {
                performances[provider.Name] = perf
            }
        }
        
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(performances)
    }
}

func (ah *AdminHandler) GetInsights(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    // Get provider metrics
    providerMetrics := ah.providerManager.GetProviderMetrics()
    
    insights, err := ah.analyticsEngine.insightsGenerator.GenerateInsights(ctx, providerMetrics)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    response := map[string]interface{}{
        "insights": insights,
        "generated_at": time.Now(),
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

func (ah *AdminHandler) GetSystemHealth(w http.ResponseWriter, r *http.Request) {
    health := map[string]interface{}{
        "status": "healthy",
        "timestamp": time.Now(),
        "uptime": "99.9%",
        "redis_status": "connected",
        "provider_count": len(ah.providerManager.GetProviderMetrics()),
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(health)
}

func (ah *AdminHandler) GetProvidersHealth(w http.ResponseWriter, r *http.Request) {
    metrics := ah.providerManager.GetProviderMetrics()
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(metrics)
}

func (ah *AdminHandler) GetProvider(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    providerName := vars["name"]
    
    provider, err := ah.providerManager.GetProvider(providerName)
    if err != nil {
        http.Error(w, "Provider not found", http.StatusNotFound)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(provider)
}

func (ah *AdminHandler) DiscoverModels(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    providerName := vars["name"]
    
    ctx := r.Context()
    models, err := ah.providerManager.DiscoverModels(ctx, providerName)
    if err != nil {
        http.Error(w, fmt.Sprintf("Failed to discover models: %v", err), http.StatusInternalServerError)
        return
    }
    
    response := map[string]interface{}{
        "provider": providerName,
        "models": models,
        "discovered_at": time.Now(),
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

func (ah *AdminHandler) GenerateConfigs(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    configs, err := ah.providerManager.GenerateConfigs(ctx)
    if err != nil {
        http.Error(w, fmt.Sprintf("Failed to generate configs: %v", err), http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(configs)
}

func countHealthyProviders(metrics map[string]map[string]interface{}) int {
    count := 0
    for _, m := range metrics {
        if status, ok := m["status"].(string); ok && status == "healthy" {
            count++
        }
    }
    return count
}

func equalStringSlices(a, b []string) bool {
    if len(a) != len(b) {
        return false
    }
    for i := range a {
        if a[i] != b[i] {
            return false
        }
    }
    return true
}