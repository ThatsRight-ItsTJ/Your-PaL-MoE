package pkg

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/labring/aiproxy/core/model"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Data structures for enhanced API key management
type CreateKeyParams struct {
	UserID              int                    `json:"user_id"`
	Name                string                 `json:"name"`
	Description         string                 `json:"description"`
	Tags                []string               `json:"tags"`
	Environment         string                 `json:"environment"`
	
	// Limits
	CostLimitUSD        *float64              `json:"cost_limit_usd"`
	RateLimitRPM        *int                  `json:"rate_limit_rpm"`
	RateLimitRPH        *int                  `json:"rate_limit_rph"`
	RateLimitRPD        *int                  `json:"rate_limit_rpd"`
	QuotaDaily          *int                  `json:"quota_daily"`
	QuotaMonthly        *int                  `json:"quota_monthly"`
	
	// Access Control
	AllowedModels       []string              `json:"allowed_models"`
	BlockedModels       []string              `json:"blocked_models"`
	AllowedEndpoints    []string              `json:"allowed_endpoints"`
	
	// Lifecycle
	ExpiresAt           *time.Time            `json:"expires_at"`
	AutoRotate          bool                  `json:"auto_rotate"`
	RotationIntervalDays *int                 `json:"rotation_interval_days"`
}

type KeyValidationResult struct {
	Valid               bool                  `json:"valid"`
	Key                 *model.TokenEnhanced  `json:"key,omitempty"`
	RateLimitStatus     RateLimitStatus       `json:"rate_limit_status"`
	CostLimitStatus     CostLimitStatus       `json:"cost_limit_status"`
	ErrorMessage        string                `json:"error_message,omitempty"`
}

type UsageMetrics struct {
	RequestCount        int                   `json:"request_count"`
	TokensUsed          int                   `json:"tokens_used"`
	CostUSD             float64               `json:"cost_usd"`
	ModelUsed           string                `json:"model_used"`
	Endpoint            string                `json:"endpoint"`
}

type RateLimitStatus struct {
	RPMRemaining        int                   `json:"rpm_remaining"`
	RPHRemaining        int                   `json:"rph_remaining"`
	RPDRemaining        int                   `json:"rpd_remaining"`
	ResetTime           time.Time             `json:"reset_time"`
}

type CostLimitStatus struct {
	LimitUSD            float64               `json:"limit_usd"`
	UsedUSD             float64               `json:"used_usd"`
	RemainingUSD        float64               `json:"remaining_usd"`
	ResetDate           time.Time             `json:"reset_date"`
}

type APIKey struct {
	*model.TokenEnhanced
	PlainKey            string                `json:"plain_key,omitempty"`
}

// Key creation with enhanced features
func CreateAPIKey(ctx context.Context, params CreateKeyParams) (*APIKey, error) {
	// Convert tags to JSON
	tagsJSON, err := json.Marshal(params.Tags)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal tags: %w", err)
	}

	// Convert allowed/blocked models to JSON
	allowedModelsJSON, err := json.Marshal(params.AllowedModels)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal allowed models: %w", err)
	}

	blockedModelsJSON, err := json.Marshal(params.BlockedModels)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal blocked models: %w", err)
	}

	// Convert allowed endpoints to JSON
	allowedEndpointsJSON, err := json.Marshal(params.AllowedEndpoints)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal allowed endpoints: %w", err)
	}

	// Create enhanced token
	token := &model.TokenEnhanced{
		Token: model.Token{
			Name:    model.EmptyNullString(params.Name),
			GroupID: fmt.Sprintf("user_%d", params.UserID), // Map user to group
			Status:  model.TokenStatusEnabled,
		},
		CostLimitUSD:         params.CostLimitUSD,
		RateLimitRPM:         params.RateLimitRPM,
		RateLimitRPH:         params.RateLimitRPH,
		RateLimitRPD:         params.RateLimitRPD,
		QuotaRequestsDaily:   params.QuotaDaily,
		QuotaRequestsMonthly: params.QuotaMonthly,
		Tags:                 tagsJSON,
		Description:          params.Description,
		Environment:          params.Environment,
		ExpiresAt:            params.ExpiresAt,
		AutoRotate:           params.AutoRotate,
		RotationIntervalDays: params.RotationIntervalDays,
		AllowedModels:        allowedModelsJSON,
		BlockedModels:        blockedModelsJSON,
		AllowedEndpoints:     allowedEndpointsJSON,
		CreatedBy:            &params.UserID,
		TokenStatus:          "active",
	}

	// Set default environment if not provided
	if token.Environment == "" {
		token.Environment = "production"
	}

	err = model.DB.Create(token).Error
	if err != nil {
		return nil, fmt.Errorf("failed to create token: %w", err)
	}

	// Log audit event
	metadata, _ := json.Marshal(map[string]interface{}{
		"environment": token.Environment,
		"tags":        params.Tags,
	})
	
	err = model.LogAuditEvent(
		&params.UserID, 
		&token.ID, 
		model.AuditActionKeyCreated, 
		model.ResourceTypeToken, 
		fmt.Sprintf("%d", token.ID),
		nil, "", "", "", 
		nil, 
		tagsJSON, 
		metadata, 
		true, 
		"",
	)
	if err != nil {
		log.Errorf("Failed to log audit event: %v", err)
	}

	return &APIKey{
		TokenEnhanced: token,
		PlainKey:      token.Key,
	}, nil
}

// Key validation with usage tracking
func ValidateAPIKey(ctx context.Context, keyHash string) (*KeyValidationResult, error) {
	if keyHash == "" {
		return &KeyValidationResult{
			Valid:        false,
			ErrorMessage: "no token provided",
		}, nil
	}

	// Get token from database
	var token model.TokenEnhanced
	err := model.DB.Where("key = ?", keyHash).First(&token).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &KeyValidationResult{
				Valid:        false,
				ErrorMessage: "invalid token",
			}, nil
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Check if token is disabled
	if token.Status == model.TokenStatusDisabled {
		return &KeyValidationResult{
			Valid:        false,
			ErrorMessage: fmt.Sprintf("token (%s[%d]) is disabled", token.Name, token.ID),
		}, nil
	}

	// Check if token is expired
	if token.ExpiresAt != nil && time.Now().After(*token.ExpiresAt) {
		return &KeyValidationResult{
			Valid:        false,
			ErrorMessage: "token has expired",
		}, nil
	}

	// Check quota limits
	if token.Quota > 0 && token.UsedAmount >= token.Quota {
		return &KeyValidationResult{
			Valid:        false,
			ErrorMessage: fmt.Sprintf("token (%s[%d]) quota is exhausted", token.Name, token.ID),
		}, nil
	}

	// Check cost limits
	costStatus := CostLimitStatus{}
	if token.CostLimitUSD != nil {
		costStatus.LimitUSD = *token.CostLimitUSD
		costStatus.UsedUSD = token.CostUsedUSD
		costStatus.RemainingUSD = *token.CostLimitUSD - token.CostUsedUSD
		
		if token.CostResetDate != nil {
			costStatus.ResetDate = *token.CostResetDate
		}

		if token.CostUsedUSD >= *token.CostLimitUSD {
			return &KeyValidationResult{
				Valid:           false,
				ErrorMessage:    "cost limit exceeded",
				CostLimitStatus: costStatus,
			}, nil
		}
	}

	// Check rate limits
	rateLimitStatus := checkRateLimits(&token)
	if !rateLimitStatus.WithinLimits {
		return &KeyValidationResult{
			Valid:           false,
			ErrorMessage:    "rate limit exceeded",
			RateLimitStatus: rateLimitStatus.Status,
		}, nil
	}

	// Update last used time
	now := time.Now()
	token.LastUsedAt = &now
	model.DB.Model(&token).Where("id = ?", token.ID).Update("last_used_at", now)

	return &KeyValidationResult{
		Valid:           true,
		Key:             &token,
		RateLimitStatus: rateLimitStatus.Status,
		CostLimitStatus: costStatus,
	}, nil
}

type rateLimitCheckResult struct {
	WithinLimits bool
	Status       RateLimitStatus
}

func checkRateLimits(token *model.TokenEnhanced) rateLimitCheckResult {
	now := time.Now()
	result := rateLimitCheckResult{
		WithinLimits: true,
		Status: RateLimitStatus{
			ResetTime: now.Add(time.Minute), // Default to next minute
		},
	}

	// For simplification, we'll implement basic rate limiting
	// In production, this should use Redis or similar for distributed rate limiting

	// Check RPM (requests per minute)
	if token.RateLimitRPM != nil {
		// Implementation would check requests in the last minute
		result.Status.RPMRemaining = *token.RateLimitRPM
	}

	// Check RPH (requests per hour)
	if token.RateLimitRPH != nil {
		// Implementation would check requests in the last hour
		result.Status.RPHRemaining = *token.RateLimitRPH
		result.Status.ResetTime = now.Add(time.Hour)
	}

	// Check RPD (requests per day)
	if token.RateLimitRPD != nil {
		// Implementation would check requests in the last day
		result.Status.RPDRemaining = *token.RateLimitRPD
		result.Status.ResetTime = now.Add(24 * time.Hour)
	}

	return result
}

// Usage tracking and limits enforcement
func TrackKeyUsage(ctx context.Context, keyID int, usage UsageMetrics) error {
	// Update usage statistics
	updates := map[string]interface{}{
		"request_count": gorm.Expr("request_count + ?", usage.RequestCount),
		"usage_count":   gorm.Expr("usage_count + ?", usage.RequestCount),
		"last_used_at":  time.Now(),
	}

	if usage.TokensUsed > 0 {
		// Implementation would track token usage
		log.Debugf("Token usage for key %d: %d tokens", keyID, usage.TokensUsed)
	}

	result := model.DB.Model(&model.TokenEnhanced{}).
		Where("id = ?", keyID).
		Updates(updates)

	if result.Error != nil {
		return fmt.Errorf("failed to update key usage: %w", result.Error)
	}

	// Log usage for analytics
	err := model.LogAuditEvent(
		nil,
		&keyID,
		model.AuditActionAPIRequest,
		model.ResourceTypeAPI,
		usage.Endpoint,
		nil, "", usage.Endpoint, "POST",
		nil, nil,
		json.RawMessage(fmt.Sprintf(`{"model": "%s", "tokens": %d, "cost": %f}`, 
			usage.ModelUsed, usage.TokensUsed, usage.CostUSD)),
		true, "",
	)
	if err != nil {
		log.Errorf("Failed to log usage audit event: %v", err)
	}

	return nil
}

// Cost tracking and enforcement
func TrackKeyCost(ctx context.Context, keyID int, costUSD float64) error {
	if costUSD <= 0 {
		return nil
	}

	updates := map[string]interface{}{
		"cost_used_usd": gorm.Expr("cost_used_usd + ?", costUSD),
		"used_amount":   gorm.Expr("used_amount + ?", costUSD),
		"last_used_at":  time.Now(),
	}

	result := model.DB.Model(&model.TokenEnhanced{}).
		Where("id = ?", keyID).
		Updates(updates)

	if result.Error != nil {
		return fmt.Errorf("failed to update key cost: %w", result.Error)
	}

	return nil
}

// Key lifecycle management
func RotateAPIKey(ctx context.Context, keyID int) (*APIKey, error) {
	var token model.TokenEnhanced
	err := model.DB.Where("id = ?", keyID).First(&token).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find token: %w", err)
	}

	// Generate new key
	oldKey := token.Key
	token.Key = "" // This will trigger key generation in BeforeCreate
	
	// Update rotation timestamp
	now := time.Now()
	token.LastRotatedAt = &now

	err = model.DB.Model(&token).Where("id = ?", keyID).Updates(map[string]interface{}{
		"key":              token.Key,
		"last_rotated_at": now,
	}).Error
	if err != nil {
		return nil, fmt.Errorf("failed to rotate key: %w", err)
	}

	// Log audit event
	metadata, _ := json.Marshal(map[string]interface{}{
		"old_key_hash": oldKey[:8] + "...", // Only log partial for security
		"rotation_type": "manual",
	})

	err = model.LogAuditEvent(
		nil, &keyID,
		model.AuditActionKeyRotated,
		model.ResourceTypeToken,
		fmt.Sprintf("%d", keyID),
		nil, "", "", "",
		nil, nil, metadata,
		true, "",
	)
	if err != nil {
		log.Errorf("Failed to log key rotation audit event: %v", err)
	}

	return &APIKey{
		TokenEnhanced: &token,
		PlainKey:      token.Key,
	}, nil
}

func ExpireAPIKey(ctx context.Context, keyID int) error {
	updates := map[string]interface{}{
		"status":      model.TokenStatusDisabled,
		"expires_at":  time.Now(),
		"token_status": "expired",
	}

	result := model.DB.Model(&model.TokenEnhanced{}).
		Where("id = ?", keyID).
		Updates(updates)

	if result.Error != nil {
		return fmt.Errorf("failed to expire key: %w", result.Error)
	}

	// Log audit event
	err := model.LogAuditEvent(
		nil, &keyID,
		model.AuditActionKeyUpdated,
		model.ResourceTypeToken,
		fmt.Sprintf("%d", keyID),
		nil, "", "", "",
		nil, 
		json.RawMessage(`{"status": "expired"}`),
		nil,
		true, "",
	)
	if err != nil {
		log.Errorf("Failed to log key expiration audit event: %v", err)
	}

	return nil
}

func SetKeyExpiration(ctx context.Context, keyID int, expiresAt time.Time) error {
	updates := map[string]interface{}{
		"expires_at": expiresAt,
	}

	result := model.DB.Model(&model.TokenEnhanced{}).
		Where("id = ?", keyID).
		Updates(updates)

	if result.Error != nil {
		return fmt.Errorf("failed to set key expiration: %w", result.Error)
	}

	return nil
}