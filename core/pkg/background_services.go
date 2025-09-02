package pkg

import (
	"context"
	"time"

	"github.com/labring/aiproxy/core/model"
	log "github.com/sirupsen/logrus"
)

// StartKeyRotationService - Key rotation service
func StartKeyRotationService(ctx context.Context) {
	ticker := time.NewTicker(time.Hour * 24) // Check daily
	defer ticker.Stop()

	log.Info("Key rotation service started")

	for {
		select {
		case <-ctx.Done():
			log.Info("Key rotation service stopped")
			return
		case <-ticker.C:
			rotateExpiredKeys()
		}
	}
}

func rotateExpiredKeys() {
	var tokens []model.TokenEnhanced
	
	// Find tokens that need rotation
	err := model.DB.Where("auto_rotate = ? AND rotation_interval_days IS NOT NULL", true).
		Where("last_rotated_at IS NULL OR last_rotated_at < ?", 
			time.Now().AddDate(0, 0, -30)). // Default 30 days if no interval specified
		Find(&tokens).Error
	
	if err != nil {
		log.Errorf("Failed to find tokens for rotation: %v", err)
		return
	}

	for _, token := range tokens {
		shouldRotate := false
		
		if token.LastRotatedAt == nil {
			// Never rotated, check creation date
			if token.RotationIntervalDays != nil {
				daysSinceCreation := int(time.Since(token.CreatedAt).Hours() / 24)
				if daysSinceCreation >= *token.RotationIntervalDays {
					shouldRotate = true
				}
			}
		} else {
			// Check last rotation date
			if token.RotationIntervalDays != nil {
				daysSinceRotation := int(time.Since(*token.LastRotatedAt).Hours() / 24)
				if daysSinceRotation >= *token.RotationIntervalDays {
					shouldRotate = true
				}
			}
		}

		if shouldRotate {
			_, err := RotateAPIKey(context.Background(), token.ID)
			if err != nil {
				log.Errorf("Failed to auto-rotate key %d: %v", token.ID, err)
			} else {
				log.Infof("Auto-rotated key %d", token.ID)
			}
		}
	}
}

// StartUsageResetService - Usage reset service (daily/monthly quotas)
func StartUsageResetService(ctx context.Context) {
	ticker := time.NewTicker(time.Hour) // Check hourly
	defer ticker.Stop()

	log.Info("Usage reset service started")

	for {
		select {
		case <-ctx.Done():
			log.Info("Usage reset service stopped")
			return
		case <-ticker.C:
			resetDailyUsage()
			resetMonthlyUsage()
			resetCostLimits()
		}
	}
}

func resetDailyUsage() {
	now := time.Now()
	
	// Reset daily usage counters at midnight
	if now.Hour() == 0 && now.Minute() < 5 { // 5-minute window
		err := model.DB.Model(&model.TokenEnhanced{}).
			Where("quota_requests_daily IS NOT NULL").
			Update("requests_used_daily", 0).Error
		
		if err != nil {
			log.Errorf("Failed to reset daily usage: %v", err)
		} else {
			log.Info("Daily usage counters reset")
		}
	}
}

func resetMonthlyUsage() {
	now := time.Now()
	
	// Reset monthly usage counters at start of month
	if now.Day() == 1 && now.Hour() == 0 && now.Minute() < 5 { // 5-minute window
		err := model.DB.Model(&model.TokenEnhanced{}).
			Where("quota_requests_monthly IS NOT NULL").
			Update("requests_used_monthly", 0).Error
		
		if err != nil {
			log.Errorf("Failed to reset monthly usage: %v", err)
		} else {
			log.Info("Monthly usage counters reset")
		}
	}
}

func resetCostLimits() {
	now := time.Now()
	
	// Reset cost limits based on reset date
	var tokens []model.TokenEnhanced
	err := model.DB.Where("cost_reset_date IS NOT NULL AND cost_reset_date <= ?", now).
		Find(&tokens).Error
	
	if err != nil {
		log.Errorf("Failed to find tokens for cost reset: %v", err)
		return
	}

	for _, token := range tokens {
		// Reset cost usage and update reset date (assume monthly reset)
		nextResetDate := token.CostResetDate.AddDate(0, 1, 0)
		
		err := model.DB.Model(&token).Updates(map[string]interface{}{
			"cost_used_usd":   0,
			"cost_reset_date": nextResetDate,
		}).Error
		
		if err != nil {
			log.Errorf("Failed to reset cost for token %d: %v", token.ID, err)
		} else {
			log.Infof("Cost limit reset for token %d", token.ID)
		}
	}
}

// StartKeyExpirationService - Key expiration service
func StartKeyExpirationService(ctx context.Context) {
	ticker := time.NewTicker(time.Hour * 6) // Check every 6 hours
	defer ticker.Stop()

	log.Info("Key expiration service started")

	for {
		select {
		case <-ctx.Done():
			log.Info("Key expiration service stopped")
			return
		case <-ticker.C:
			expireOldKeys()
		}
	}
}

func expireOldKeys() {
	now := time.Now()
	
	// Find and expire keys that have passed their expiration date
	var expiredTokens []model.TokenEnhanced
	err := model.DB.Where("expires_at IS NOT NULL AND expires_at <= ? AND status = ?", 
		now, model.TokenStatusEnabled).
		Find(&expiredTokens).Error
	
	if err != nil {
		log.Errorf("Failed to find expired tokens: %v", err)
		return
	}

	for _, token := range expiredTokens {
		err := ExpireAPIKey(context.Background(), token.ID)
		if err != nil {
			log.Errorf("Failed to expire token %d: %v", token.ID, err)
		} else {
			log.Infof("Expired token %d", token.ID)
		}
	}

	if len(expiredTokens) > 0 {
		log.Infof("Expired %d tokens", len(expiredTokens))
	}
}

// StartUsageAnalyticsService - Usage analytics aggregation
func StartUsageAnalyticsService(ctx context.Context) {
	ticker := time.NewTicker(time.Hour) // Run hourly
	defer ticker.Stop()

	log.Info("Usage analytics service started")

	for {
		select {
		case <-ctx.Done():
			log.Info("Usage analytics service stopped")
			return
		case <-ticker.C:
			aggregateUsageAnalytics()
		}
	}
}

func aggregateUsageAnalytics() {
	// This would typically aggregate usage data into summary tables
	// for better performance when generating reports
	
	now := time.Now()
	hourAgo := now.Add(-time.Hour)
	
	// Aggregate audit logs for the last hour
	var stats struct {
		TotalRequests      int64
		SuccessfulRequests int64
		FailedRequests     int64
		UniqueTokens       int64
	}

	err := model.DB.Model(&model.AuditLog{}).
		Where("created_at BETWEEN ? AND ? AND action = ?", 
			hourAgo, now, model.AuditActionAPIRequest).
		Select(`
			COUNT(*) as total_requests,
			SUM(CASE WHEN success = true THEN 1 ELSE 0 END) as successful_requests,
			SUM(CASE WHEN success = false THEN 1 ELSE 0 END) as failed_requests,
			COUNT(DISTINCT token_id) as unique_tokens
		`).
		Scan(&stats).Error

	if err != nil {
		log.Errorf("Failed to aggregate usage analytics: %v", err)
		return
	}

	log.Debugf("Usage analytics for last hour: %+v", stats)
}

// StartAllBackgroundServices starts all key management background services
func StartAllBackgroundServices(ctx context.Context) {
	go StartKeyRotationService(ctx)
	go StartUsageResetService(ctx)
	go StartKeyExpirationService(ctx)
	go StartUsageAnalyticsService(ctx)
	
	log.Info("All key management background services started")
}