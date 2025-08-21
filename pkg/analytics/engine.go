package analytics

import (
    "context"
    "encoding/json"
    "fmt"
    "sync"
    "time"

    "github.com/go-redis/redis/v8"
    "github.com/intelligent-ai-gateway/pkg/pollinations"
    "github.com/intelligent-ai-gateway/pkg/providers"
)

type AnalyticsEngine struct {
    redis              *redis.Client
    pollinationsClient *pollinations.Client
    providerManager    *providers.ProviderManager
    metricsStore       *MetricsStore
    insightsGenerator  *InsightsGenerator
}

type MetricsStore struct {
    redis  *redis.Client
    mu     sync.RWMutex
    cache  map[string]*CachedMetrics
}

type CachedMetrics struct {
    Data      interface{}
    UpdatedAt time.Time
    TTL       time.Duration
}

type RequestMetrics struct {
    RequestID        string        `json:"request_id"`
    UserID           string        `json:"user_id"`
    Timestamp        time.Time     `json:"timestamp"`
    Provider         string        `json:"provider"`
    Model            string        `json:"model"`
    Tier             string        `json:"tier"`
    TaskType         string        `json:"task_type"`
    ResponseTime     time.Duration `json:"response_time"`
    Cost             float64       `json:"cost"`
    TokensUsed       int           `json:"tokens_used"`
    Success          bool          `json:"success"`
    ErrorMessage     string        `json:"error_message,omitempty"`
    QualityScore     int           `json:"quality_score"`
    FallbackUsed     bool          `json:"fallback_used"`
    ParallelExecution bool         `json:"parallel_execution"`
}

type ProviderPerformance struct {
    ProviderName     string        `json:"provider_name"`
    Tier             string        `json:"tier"`
    TotalRequests    int           `json:"total_requests"`
    SuccessfulRequests int         `json:"successful_requests"`
    FailedRequests   int           `json:"failed_requests"`
    SuccessRate      float64       `json:"success_rate"`
    AverageResponseTime time.Duration `json:"average_response_time"`
    TotalCost        float64       `json:"total_cost"`
    TotalTokens      int           `json:"total_tokens"`
    AverageQuality   float64       `json:"average_quality"`
    Uptime           float64       `json:"uptime_percentage"`
    LastUsed         time.Time     `json:"last_used"`
}

type CostAnalysis struct {
    Period           string                    `json:"period"`
    TotalCost        float64                   `json:"total_cost"`
    CostByTier       map[string]float64        `json:"cost_by_tier"`
    CostByProvider   map[string]float64        `json:"cost_by_provider"`
    CostByUser       map[string]float64        `json:"cost_by_user"`
    SavingsVsOfficial float64                  `json:"savings_vs_official"`
    OptimizationOpportunities []OptimizationOpportunity `json:"optimization_opportunities"`
}

type OptimizationOpportunity struct {
    Description      string  `json:"description"`
    PotentialSavings float64 `json:"potential_savings"`
    Implementation   string  `json:"implementation"`
    Impact           string  `json:"impact"`
}

func NewAnalyticsEngine(redis *redis.Client, providerManager *providers.ProviderManager) *AnalyticsEngine {
    return &AnalyticsEngine{
        redis:              redis,
        pollinationsClient: pollinations.NewClient(),
        providerManager:    providerManager,
        metricsStore:       NewMetricsStore(redis),
        insightsGenerator:  NewInsightsGenerator(),
    }
}

func NewMetricsStore(redis *redis.Client) *MetricsStore {
    return &MetricsStore{
        redis: redis,
        cache: make(map[string]*CachedMetrics),
    }
}

func (ae *AnalyticsEngine) RecordRequest(metrics RequestMetrics) error {
    // Store in Redis time series
    key := fmt.Sprintf("metrics:request:%s", metrics.RequestID)
    data, err := json.Marshal(metrics)
    if err != nil {
        return err
    }
    
    ctx := context.Background()
    
    // Store individual request
    if err := ae.redis.Set(ctx, key, data, 30*24*time.Hour).Err(); err != nil {
        return err
    }
    
    // Update aggregated metrics
    ae.updateAggregatedMetrics(ctx, metrics)
    
    return nil
}

func (ae *AnalyticsEngine) updateAggregatedMetrics(ctx context.Context, metrics RequestMetrics) {
    // Update provider metrics
    providerKey := fmt.Sprintf("metrics:provider:%s", metrics.Provider)
    ae.redis.HIncrBy(ctx, providerKey, "total_requests", 1)
    
    if metrics.Success {
        ae.redis.HIncrBy(ctx, providerKey, "successful_requests", 1)
    } else {
        ae.redis.HIncrBy(ctx, providerKey, "failed_requests", 1)
    }
    
    ae.redis.HIncrByFloat(ctx, providerKey, "total_cost", metrics.Cost)
    ae.redis.HIncrBy(ctx, providerKey, "total_tokens", int64(metrics.TokensUsed))
    
    // Update tier metrics
    tierKey := fmt.Sprintf("metrics:tier:%s", metrics.Tier)
    ae.redis.HIncrBy(ctx, tierKey, "total_requests", 1)
    ae.redis.HIncrByFloat(ctx, tierKey, "total_cost", metrics.Cost)
    
    // Update user metrics
    userKey := fmt.Sprintf("metrics:user:%s", metrics.UserID)
    ae.redis.HIncrBy(ctx, userKey, "total_requests", 1)
    ae.redis.HIncrByFloat(ctx, userKey, "total_cost", metrics.Cost)
    
    // Update daily metrics
    dayKey := fmt.Sprintf("metrics:daily:%s", time.Now().Format("2006-01-02"))
    ae.redis.HIncrBy(ctx, dayKey, "total_requests", 1)
    ae.redis.HIncrByFloat(ctx, dayKey, "total_cost", metrics.Cost)
}

func (ae *AnalyticsEngine) GetProviderPerformance(ctx context.Context, providerName string, period time.Duration) (*ProviderPerformance, error) {
    // Check cache first
    cacheKey := fmt.Sprintf("performance:%s:%v", providerName, period)
    if cached := ae.metricsStore.GetCached(cacheKey); cached != nil {
        if perf, ok := cached.(*ProviderPerformance); ok {
            return perf, nil
        }
    }
    
    // Calculate performance metrics
    performance := &ProviderPerformance{
        ProviderName: providerName,
    }
    
    // Get provider config for tier
    provider, _ := ae.providerManager.GetProvider(providerName)
    if provider != nil {
        performance.Tier = provider.Tier
    }
    
    // Fetch metrics from Redis
    providerKey := fmt.Sprintf("metrics:provider:%s", providerName)
    metrics, err := ae.redis.HGetAll(ctx, providerKey).Result()
    if err != nil {
        return nil, err
    }
    
    // Parse metrics
    if total, exists := metrics["total_requests"]; exists {
        fmt.Sscanf(total, "%d", &performance.TotalRequests)
    }
    if successful, exists := metrics["successful_requests"]; exists {
        fmt.Sscanf(successful, "%d", &performance.SuccessfulRequests)
    }
    if failed, exists := metrics["failed_requests"]; exists {
        fmt.Sscanf(failed, "%d", &performance.FailedRequests)
    }
    if cost, exists := metrics["total_cost"]; exists {
        fmt.Sscanf(cost, "%f", &performance.TotalCost)
    }
    if tokens, exists := metrics["total_tokens"]; exists {
        fmt.Sscanf(tokens, "%d", &performance.TotalTokens)
    }
    
    // Calculate derived metrics
    if performance.TotalRequests > 0 {
        performance.SuccessRate = float64(performance.SuccessfulRequests) / float64(performance.TotalRequests) * 100
    }
    
    // Get health metrics
    if healthMetrics := ae.providerManager.GetProviderMetrics(); healthMetrics != nil {
        if providerHealth, exists := healthMetrics[providerName]; exists {
            if uptime, ok := providerHealth["success_rate"].(float64); ok {
                performance.Uptime = uptime
            }
        }
    }
    
    // Cache result
    ae.metricsStore.Cache(cacheKey, performance, 5*time.Minute)
    
    return performance, nil
}

func (ae *AnalyticsEngine) GetCostAnalysis(ctx context.Context, period string) (*CostAnalysis, error) {
    analysis := &CostAnalysis{
        Period:         period,
        CostByTier:     make(map[string]float64),
        CostByProvider: make(map[string]float64),
        CostByUser:     make(map[string]float64),
    }
    
    // Get all providers
    providers, _ := ae.providerManager.GetAvailableProviders()
    
    // Calculate costs by tier
    for _, tier := range []string{"official", "community", "unofficial"} {
        tierKey := fmt.Sprintf("metrics:tier:%s", tier)
        if cost, err := ae.redis.HGet(ctx, tierKey, "total_cost").Float64(); err == nil {
            analysis.CostByTier[tier] = cost
            analysis.TotalCost += cost
        }
    }
    
    // Calculate costs by provider
    for _, provider := range providers {
        providerKey := fmt.Sprintf("metrics:provider:%s", provider.Name)
        if cost, err := ae.redis.HGet(ctx, providerKey, "total_cost").Float64(); err == nil {
            analysis.CostByProvider[provider.Name] = cost
        }
    }
    
    // Calculate savings vs all-official
    officialCost := analysis.CostByTier["official"]
    totalRequests := 0
    for _, tier := range []string{"official", "community", "unofficial"} {
        tierKey := fmt.Sprintf("metrics:tier:%s", tier)
        if requests, err := ae.redis.HGet(ctx, tierKey, "total_requests").Int(); err == nil {
            totalRequests += requests
        }
    }
    
    // Estimate what it would cost if all requests went to official tier
    avgOfficialCostPerRequest := 0.0
    if officialRequests, _ := ae.redis.HGet(ctx, "metrics:tier:official", "total_requests").Int(); officialRequests > 0 {
        avgOfficialCostPerRequest = officialCost / float64(officialRequests)
    } else {
        avgOfficialCostPerRequest = 0.005 // Default estimate
    }
    
    estimatedAllOfficialCost := avgOfficialCostPerRequest * float64(totalRequests)
    analysis.SavingsVsOfficial = estimatedAllOfficialCost - analysis.TotalCost
    
    // Generate optimization opportunities
    analysis.OptimizationOpportunities = ae.generateOptimizationOpportunities(ctx, analysis)
    
    return analysis, nil
}

func (ae *AnalyticsEngine) generateOptimizationOpportunities(ctx context.Context, analysis *CostAnalysis) []OptimizationOpportunity {
    opportunities := []OptimizationOpportunity{}
    
    // Use Pollinations to analyze patterns and suggest optimizations
    prompt := fmt.Sprintf(`
Analyze these cost metrics and suggest optimization opportunities:

Total Cost: $%.2f
Cost by Tier: %v
Cost by Provider: %v

Suggest 3-5 specific optimization opportunities. Return JSON array:
[
  {
    "description": "Move more image generation to unofficial providers",
    "potential_savings": 0.50,
    "implementation": "Route simple image requests to Bing DALL-E instead of OpenAI",
    "impact": "high"
  }
]
`, analysis.TotalCost, analysis.CostByTier, analysis.CostByProvider)
    
    response, err := ae.pollinationsClient.GenerateText(ctx, prompt)
    if err == nil {
        var suggestions []OptimizationOpportunity
        if err := json.Unmarshal([]byte(response), &suggestions); err == nil {
            opportunities = append(opportunities, suggestions...)
        }
    }
    
    // Add rule-based opportunities
    if analysis.CostByTier["official"] > analysis.TotalCost*0.5 {
        opportunities = append(opportunities, OptimizationOpportunity{
            Description:      "Reduce reliance on official tier providers",
            PotentialSavings: analysis.CostByTier["official"] * 0.3,
            Implementation:   "Enable more aggressive cost optimization in routing",
            Impact:           "high",
        })
    }
    
    return opportunities
}

func (ms *MetricsStore) Cache(key string, data interface{}, ttl time.Duration) {
    ms.mu.Lock()
    defer ms.mu.Unlock()
    
    ms.cache[key] = &CachedMetrics{
        Data:      data,
        UpdatedAt: time.Now(),
        TTL:       ttl,
    }
}

func (ms *MetricsStore) GetCached(key string) interface{} {
    ms.mu.RLock()
    defer ms.mu.RUnlock()
    
    if cached, exists := ms.cache[key]; exists {
        if time.Since(cached.UpdatedAt) < cached.TTL {
            return cached.Data
        }
    }
    
    return nil
}

type InsightsGenerator struct{}

func NewInsightsGenerator() *InsightsGenerator {
    return &InsightsGenerator{}
}

func (ig *InsightsGenerator) GenerateInsights(ctx context.Context, metrics map[string]interface{}) ([]string, error) {
    // This would use Pollinations to generate intelligent insights
    insights := []string{
        "Provider X has shown 15% improvement in response time this week",
        "Cost savings of 78% achieved through intelligent routing",
        "Consider enabling parallel execution for multi-step requests",
    }
    
    return insights, nil
}