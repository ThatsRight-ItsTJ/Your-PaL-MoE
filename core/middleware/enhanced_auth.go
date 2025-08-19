package middleware

import (
	"encoding/json"
	"net"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/labring/aiproxy/core/model"
	"github.com/labring/aiproxy/core/pkg"
	log "github.com/sirupsen/logrus"
)

// EnhancedAuthMiddleware provides enhanced API key validation with rate limiting and cost controls
func EnhancedAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip authentication for certain endpoints
		if shouldSkipAuth(c.Request.URL.Path) {
			c.Next()
			return
		}

		// Extract API key from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header required",
			})
			c.Abort()
			return
		}

		// Parse Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization header format",
			})
			c.Abort()
			return
		}

		apiKey := parts[1]

		// Validate API key using enhanced validation
		result, err := pkg.ValidateAPIKey(c.Request.Context(), apiKey)
		if err != nil {
			log.Errorf("Failed to validate API key: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Authentication service error",
			})
			c.Abort()
			return
		}

		if !result.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": result.ErrorMessage,
			})
			c.Abort()
			return
		}

		// Check model access control
		requestedModel := c.GetHeader("X-Model") // or extract from request body
		if requestedModel != "" && !isModelAllowed(result.Key, requestedModel) {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Access denied for this model",
			})
			c.Abort()
			return
		}

		// Check endpoint access control
		if !isEndpointAllowed(result.Key, c.Request.URL.Path) {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Access denied for this endpoint",
			})
			c.Abort()
			return
		}

		// Set rate limit headers
		if result.RateLimitStatus.RPMRemaining >= 0 {
			c.Header("X-RateLimit-Remaining-RPM", string(rune(result.RateLimitStatus.RPMRemaining)))
		}
		if result.RateLimitStatus.RPHRemaining >= 0 {
			c.Header("X-RateLimit-Remaining-RPH", string(rune(result.RateLimitStatus.RPHRemaining)))
		}
		if result.RateLimitStatus.RPDRemaining >= 0 {
			c.Header("X-RateLimit-Remaining-RPD", string(rune(result.RateLimitStatus.RPDRemaining)))
		}

		// Set cost limit headers
		if result.CostLimitStatus.LimitUSD > 0 {
			c.Header("X-Cost-Limit", string(rune(int(result.CostLimitStatus.LimitUSD))))
			c.Header("X-Cost-Remaining", string(rune(int(result.CostLimitStatus.RemainingUSD))))
		}

		// Store token info in context for later use
		c.Set("token", result.Key)
		c.Set("token_id", result.Key.ID)
		if result.Key.CreatedBy != nil {
			c.Set("user_id", *result.Key.CreatedBy)
		}

		// Log API request for audit
		go logAPIRequest(c, result.Key)

		c.Next()

		// Track usage after request completion
		go trackUsageAfterRequest(c, result.Key)
	}
}

func shouldSkipAuth(path string) bool {
	skipPaths := []string{
		"/health",
		"/swagger",
		"/api/v1/keys/validate", // Public validation endpoint
		"/metrics",
	}

	for _, skipPath := range skipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}
	return false
}

func isModelAllowed(token *model.TokenEnhanced, requestedModel string) bool {
	// If no restrictions, allow all
	if len(token.AllowedModels) == 0 && len(token.BlockedModels) == 0 {
		return true
	}

	// Check blocked models first
	if len(token.BlockedModels) > 0 {
		var blockedModels []string
		if err := json.Unmarshal(token.BlockedModels, &blockedModels); err == nil {
			for _, blocked := range blockedModels {
				if blocked == requestedModel {
					return false
				}
			}
		}
	}

	// Check allowed models
	if len(token.AllowedModels) > 0 {
		var allowedModels []string
		if err := json.Unmarshal(token.AllowedModels, &allowedModels); err == nil {
			for _, allowed := range allowedModels {
				if allowed == requestedModel {
					return true
				}
			}
			return false // Not in allowed list
		}
	}

	return true
}

func isEndpointAllowed(token *model.TokenEnhanced, endpoint string) bool {
	// If no endpoint restrictions, allow all
	if len(token.AllowedEndpoints) == 0 {
		return true
	}

	var allowedEndpoints []string
	if err := json.Unmarshal(token.AllowedEndpoints, &allowedEndpoints); err != nil {
		return true // If can't parse, allow
	}

	for _, allowed := range allowedEndpoints {
		if strings.HasPrefix(endpoint, allowed) {
			return true
		}
	}

	return false
}

func logAPIRequest(c *gin.Context, token *model.TokenEnhanced) {
	clientIP, _, _ := net.SplitHostPort(c.Request.RemoteAddr)
	ipAddr := net.ParseIP(clientIP)

	var userID *int
	if token.CreatedBy != nil {
		userID = token.CreatedBy
	}

	tokenID := token.ID

	err := model.LogAuditEvent(
		userID,
		&tokenID,
		model.AuditActionAPIRequest,
		model.ResourceTypeAPI,
		c.Request.URL.Path,
		&ipAddr,
		c.GetHeader("User-Agent"),
		c.Request.URL.Path,
		c.Request.Method,
		nil, nil, nil,
		true, "", // Will be updated if request fails
	)
	if err != nil {
		log.Errorf("Failed to log API request: %v", err)
	}
}

func trackUsageAfterRequest(c *gin.Context, token *model.TokenEnhanced) {
	// Extract usage metrics from response
	// This would typically be done in the actual API handlers
	usage := pkg.UsageMetrics{
		RequestCount: 1,
		TokensUsed:   0, // Would be extracted from actual response
		CostUSD:      0, // Would be calculated based on usage
		ModelUsed:    c.GetHeader("X-Model"),
		Endpoint:     c.Request.URL.Path,
	}

	err := pkg.TrackKeyUsage(c.Request.Context(), token.ID, usage)
	if err != nil {
		log.Errorf("Failed to track key usage: %v", err)
	}
}